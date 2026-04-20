package enterprisebff

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/ent/userattributedefinition"
	"github.com/Wei-Shaw/sub2api/ent/userattributevalue"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type EnterpriseProfile struct {
	UserID      int64
	Name        string `json:"enterprise_name"`
	DisplayName string `json:"enterprise_display_name,omitempty"`
	SupportInfo string `json:"enterprise_support_contact,omitempty"`
}

type EnterpriseVisibleGroup struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Platform string `json:"platform"`
}

func (p *EnterpriseProfile) DisplayLabel() string {
	if p == nil {
		return ""
	}
	if strings.TrimSpace(p.DisplayName) != "" {
		return p.DisplayName
	}
	return p.Name
}

type EnterpriseStore interface {
	MatchUserByEmailAndCompany(ctx context.Context, email, companyName string) (*EnterpriseProfile, error)
	GetByUserID(ctx context.Context, userID int64) (*EnterpriseProfile, error)
	ListVisibleGroups(ctx context.Context, userID int64) ([]EnterpriseVisibleGroup, error)
	SameEnterprise(ctx context.Context, actorUserID, targetUserID int64) (bool, error)
}

type entEnterpriseStore struct {
	client                      *dbent.Client
	settingRepo                 service.SettingRepository
	keys                        []string
	visibleGroupIDsByEnterprise map[string]map[int64]struct{}
}

func newEntEnterpriseStore(client *dbent.Client, settingRepo service.SettingRepository, cfg *Config) EnterpriseStore {
	return &entEnterpriseStore{
		client:      client,
		settingRepo: settingRepo,
		keys: []string{
			cfg.EnterpriseAttributeKey,
			cfg.EnterpriseDisplayNameAttributeKey,
			cfg.EnterpriseSupportContactAttribute,
		},
		visibleGroupIDsByEnterprise: normalizeEnterpriseVisibleGroupIDsByEnterprise(cfg.EnterpriseVisibleGroupIDsByEnterprise),
	}
}

func normalizeCompanyName(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func (s *entEnterpriseStore) MatchUserByEmailAndCompany(ctx context.Context, email, companyName string) (*EnterpriseProfile, error) {
	row, err := s.client.User.Query().
		Where(
			user.EmailEqualFold(strings.TrimSpace(email)),
			user.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	profile, err := s.GetByUserID(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, nil
	}
	if normalizeCompanyName(profile.Name) != normalizeCompanyName(companyName) {
		return nil, nil
	}
	return profile, nil
}

func (s *entEnterpriseStore) GetByUserID(ctx context.Context, userID int64) (*EnterpriseProfile, error) {
	values, err := s.client.UserAttributeValue.Query().
		Where(userattributevalue.UserIDEQ(userID)).
		WithDefinition(func(query *dbent.UserAttributeDefinitionQuery) {
			query.Where(
				userattributedefinition.DeletedAtIsNil(),
				userattributedefinition.KeyIn(s.keys...),
			)
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	profile := &EnterpriseProfile{UserID: userID}
	for _, value := range values {
		definition := value.Edges.Definition
		if definition == nil {
			continue
		}
		switch definition.Key {
		case s.keys[0]:
			profile.Name = strings.TrimSpace(value.Value)
		case s.keys[1]:
			profile.DisplayName = strings.TrimSpace(value.Value)
		case s.keys[2]:
			profile.SupportInfo = strings.TrimSpace(value.Value)
		}
	}

	if profile.Name == "" {
		return nil, nil
	}
	if profile.DisplayName == "" {
		profile.DisplayName = profile.Name
	}
	return profile, nil
}

func (s *entEnterpriseStore) ListVisibleGroups(ctx context.Context, userID int64) ([]EnterpriseVisibleGroup, error) {
	profile, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	configuredGroupIDs := s.visibleGroupIDSetForEnterprise(ctx, profile)
	if len(configuredGroupIDs) == 0 {
		return []EnterpriseVisibleGroup{}, nil
	}

	rows, err := s.client.Group.Query().
		Where(
			group.StatusEQ(service.StatusActive),
		).
		Order(dbent.Asc(group.FieldSortOrder), dbent.Asc(group.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	allGroups := make([]service.Group, 0, len(rows))
	for _, row := range rows {
		allGroups = append(allGroups, service.Group{
			ID:               row.ID,
			Name:             row.Name,
			Platform:         row.Platform,
			Status:           row.Status,
			IsExclusive:      row.IsExclusive,
			SubscriptionType: row.SubscriptionType,
		})
	}
	return selectEnterpriseVisibleGroups(
		allGroups,
		configuredGroupIDs,
	), nil
}

func (s *entEnterpriseStore) SameEnterprise(ctx context.Context, actorUserID, targetUserID int64) (bool, error) {
	actor, err := s.GetByUserID(ctx, actorUserID)
	if err != nil || actor == nil {
		return false, err
	}

	target, err := s.GetByUserID(ctx, targetUserID)
	if err != nil || target == nil {
		return false, err
	}

	return normalizeCompanyName(actor.Name) == normalizeCompanyName(target.Name), nil
}

func (s *entEnterpriseStore) visibleGroupIDSetForEnterprise(ctx context.Context, profile *EnterpriseProfile) map[int64]struct{} {
	if profile == nil {
		return nil
	}
	enterpriseName := normalizeCompanyName(profile.Name)
	if enterpriseName == "" {
		return nil
	}
	if configuredGroupIDs, ok := s.dbVisibleGroupIDSetForEnterprise(ctx, enterpriseName); ok {
		return configuredGroupIDs
	}
	return s.visibleGroupIDsByEnterprise[enterpriseName]
}

func (s *entEnterpriseStore) dbVisibleGroupIDSetForEnterprise(ctx context.Context, enterpriseName string) (map[int64]struct{}, bool) {
	if s.settingRepo == nil {
		return nil, false
	}

	raw, err := s.settingRepo.GetValue(ctx, service.SettingKeyEnterpriseVisibleGroupIDsByEnterprise)
	if err != nil {
		if errors.Is(err, service.ErrSettingNotFound) {
			return nil, false
		}
		log.Printf("enterprise-bff: failed to load enterprise visible groups setting: %v", err)
		return nil, false
	}
	if strings.TrimSpace(raw) == "" {
		return nil, false
	}

	parsed, err := service.ParseEnterpriseVisibleGroupIDsByEnterprise(raw)
	if err != nil {
		log.Printf("enterprise-bff: failed to parse enterprise visible groups setting: %v", err)
		return nil, false
	}

	return normalizeEnterpriseVisibleGroupIDsByEnterprise(parsed)[enterpriseName], true
}

func normalizeEnterpriseVisibleGroupIDsByEnterprise(raw map[string][]int64) map[string]map[int64]struct{} {
	if len(raw) == 0 {
		return nil
	}

	out := make(map[string]map[int64]struct{}, len(raw))
	for enterpriseName, ids := range raw {
		normalizedName := normalizeCompanyName(enterpriseName)
		if normalizedName == "" {
			continue
		}
		if _, ok := out[normalizedName]; !ok {
			out[normalizedName] = make(map[int64]struct{}, len(ids))
		}
		for _, id := range ids {
			if id > 0 {
				out[normalizedName][id] = struct{}{}
			}
		}
		if len(out[normalizedName]) == 0 {
			delete(out, normalizedName)
		}
	}

	return out
}

func selectEnterpriseVisibleGroups(
	groups []service.Group,
	configuredGroupIDs map[int64]struct{},
) []EnterpriseVisibleGroup {
	if len(configuredGroupIDs) == 0 {
		return []EnterpriseVisibleGroup{}
	}

	out := make([]EnterpriseVisibleGroup, 0, len(groups))
	for _, candidate := range groups {
		if !candidate.IsActive() {
			continue
		}
		if _, ok := configuredGroupIDs[candidate.ID]; !ok {
			continue
		}

		out = append(out, EnterpriseVisibleGroup{
			ID:       candidate.ID,
			Name:     candidate.Name,
			Platform: candidate.Platform,
		})
	}
	return out
}

func injectEnterpriseIntoAuthResponse(body []byte, profile *EnterpriseProfile) ([]byte, error) {
	if profile == nil {
		return body, nil
	}
	return transformEnvelope(body, func(data any) any {
		raw, ok := data.(map[string]any)
		if !ok {
			return data
		}
		userValue, ok := raw["user"].(map[string]any)
		if !ok {
			return data
		}
		raw["user"] = mergeEnterpriseFields(userValue, profile)
		return raw
	})
}

func injectEnterpriseIntoCurrentUser(body []byte, profile *EnterpriseProfile) ([]byte, error) {
	if profile == nil {
		return body, nil
	}
	return transformEnvelope(body, func(data any) any {
		raw, ok := data.(map[string]any)
		if !ok {
			return data
		}
		return mergeEnterpriseFields(raw, profile)
	})
}

func injectEnterpriseIntoPublicSettings(body []byte, profile *EnterpriseProfile) ([]byte, error) {
	if profile == nil {
		return body, nil
	}
	return transformEnvelope(body, func(data any) any {
		raw, ok := data.(map[string]any)
		if !ok {
			return data
		}
		if label := profile.DisplayLabel(); label != "" {
			raw["site_name"] = label
			raw["enterprise_display_name"] = label
		}
		raw["enterprise_name"] = profile.Name
		if profile.SupportInfo != "" {
			raw["contact_info"] = profile.SupportInfo
			raw["enterprise_support_contact"] = profile.SupportInfo
		}
		return raw
	})
}

func mergeEnterpriseFields(raw map[string]any, profile *EnterpriseProfile) map[string]any {
	raw["enterprise_name"] = profile.Name
	if label := profile.DisplayLabel(); label != "" {
		raw["enterprise_display_name"] = label
	}
	if profile.SupportInfo != "" {
		raw["enterprise_support_contact"] = profile.SupportInfo
	}
	return raw
}

func extractUserIDFromEnvelope(body []byte) (int64, error) {
	var payload struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, err
	}
	if len(payload.Data) == 0 {
		return 0, errors.New("missing response data")
	}

	var data struct {
		ID   int64 `json:"id"`
		User struct {
			ID int64 `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(payload.Data, &data); err != nil {
		return 0, err
	}
	if data.User.ID > 0 {
		return data.User.ID, nil
	}
	if data.ID > 0 {
		return data.ID, nil
	}
	return 0, errors.New("user id not found in response")
}
