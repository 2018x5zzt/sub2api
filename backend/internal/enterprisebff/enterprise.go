package enterprisebff

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/ent/userattributedefinition"
	"github.com/Wei-Shaw/sub2api/ent/userattributevalue"
)

type EnterpriseProfile struct {
	UserID      int64
	Name        string `json:"enterprise_name"`
	DisplayName string `json:"enterprise_display_name,omitempty"`
	SupportInfo string `json:"enterprise_support_contact,omitempty"`
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
}

type entEnterpriseStore struct {
	client *dbent.Client
	keys   []string
}

func newEntEnterpriseStore(client *dbent.Client, cfg *Config) EnterpriseStore {
	return &entEnterpriseStore{
		client: client,
		keys: []string{
			cfg.EnterpriseAttributeKey,
			cfg.EnterpriseDisplayNameAttributeKey,
			cfg.EnterpriseSupportContactAttribute,
		},
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
