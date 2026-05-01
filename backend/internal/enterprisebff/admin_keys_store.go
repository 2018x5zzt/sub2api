package enterprisebff

import (
	"context"
	"database/sql"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/apikey"
	"github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/ent/user"
	coreconfig "github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	repo "github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type AdminKeyListFilters struct {
	Search  string
	Status  string
	UserID  *int64
	GroupID *int64
}

type AdminKeyStore interface {
	List(context.Context, pagination.PaginationParams, AdminKeyListFilters) ([]service.APIKey, int64, error)
	Get(context.Context, int64) (*service.APIKey, error)
	Create(context.Context, int64, service.CreateAPIKeyRequest) (*service.APIKey, error)
	Update(context.Context, int64, service.UpdateAPIKeyRequest) (*service.APIKey, error)
	Delete(context.Context, int64) error
}

type adminKeyStore struct {
	client        *dbent.Client
	apiKeyRepo    service.APIKeyRepository
	apiKeyService *service.APIKeyService
}

func NewAdminKeyStore(client *dbent.Client, sqlDB *sql.DB, cfg *coreconfig.Config) AdminKeyStore {
	apiKeyRepo := repo.NewAPIKeyRepository(client, sqlDB)
	userRepo := repo.NewUserRepository(client, sqlDB)
	groupRepo := repo.NewGroupRepository(client, sqlDB)
	userSubRepo := repo.NewUserSubscriptionRepository(client)
	userGroupRateRepo := repo.NewUserGroupRateRepository(sqlDB)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, userRepo, groupRepo, userSubRepo, userGroupRateRepo, nil, cfg)

	return &adminKeyStore{
		client:        client,
		apiKeyRepo:    apiKeyRepo,
		apiKeyService: apiKeyService,
	}
}

func (s *adminKeyStore) List(ctx context.Context, params pagination.PaginationParams, filters AdminKeyListFilters) ([]service.APIKey, int64, error) {
	q := s.client.APIKey.Query().
		Where(apikey.DeletedAtIsNil())

	if filters.Search != "" {
		search := strings.TrimSpace(filters.Search)
		q = q.Where(apikey.Or(
			apikey.NameContainsFold(search),
			apikey.KeyContainsFold(search),
			apikey.HasUserWith(
				user.Or(
					user.EmailContainsFold(search),
					user.UsernameContainsFold(search),
				),
			),
			apikey.HasGroupWith(group.NameContainsFold(search)),
		))
	}
	if filters.Status != "" {
		q = q.Where(apikey.StatusEQ(filters.Status))
	}
	if filters.UserID != nil {
		q = q.Where(apikey.UserIDEQ(*filters.UserID))
	}
	if filters.GroupID != nil {
		if *filters.GroupID == 0 {
			q = q.Where(apikey.GroupIDIsNil())
		} else {
			q = q.Where(apikey.GroupIDEQ(*filters.GroupID))
		}
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := q.
		WithUser().
		WithGroup().
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(apikey.FieldID)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	out := make([]service.APIKey, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapAPIKeyEntity(row))
	}
	return out, int64(total), nil
}

func (s *adminKeyStore) Get(ctx context.Context, id int64) (*service.APIKey, error) {
	return s.apiKeyRepo.GetByID(ctx, id)
}

func (s *adminKeyStore) Create(ctx context.Context, ownerID int64, req service.CreateAPIKeyRequest) (*service.APIKey, error) {
	return s.apiKeyService.Create(ctx, ownerID, req)
}

func (s *adminKeyStore) Update(ctx context.Context, id int64, req service.UpdateAPIKeyRequest) (*service.APIKey, error) {
	key, err := s.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.apiKeyService.Update(ctx, id, key.UserID, req)
}

func (s *adminKeyStore) Delete(ctx context.Context, id int64) error {
	key, err := s.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.apiKeyService.Delete(ctx, id, key.UserID)
}

func mapAPIKeyEntity(m *dbent.APIKey) service.APIKey {
	out := service.APIKey{
		ID:            m.ID,
		UserID:        m.UserID,
		Key:           m.Key,
		Name:          m.Name,
		GroupID:       m.GroupID,
		Status:        m.Status,
		IPWhitelist:   m.IPWhitelist,
		IPBlacklist:   m.IPBlacklist,
		LastUsedAt:    m.LastUsedAt,
		Quota:         m.Quota,
		QuotaUsed:     m.QuotaUsed,
		ExpiresAt:     m.ExpiresAt,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
		RateLimit5h:   m.RateLimit5h,
		RateLimit1d:   m.RateLimit1d,
		RateLimit7d:   m.RateLimit7d,
		Usage5h:       m.Usage5h,
		Usage1d:       m.Usage1d,
		Usage7d:       m.Usage7d,
		Window5hStart: m.Window5hStart,
		Window1dStart: m.Window1dStart,
		Window7dStart: m.Window7dStart,
	}
	if m.Edges.User != nil {
		out.User = &service.User{
			ID:                  m.Edges.User.ID,
			Email:               m.Edges.User.Email,
			Username:            m.Edges.User.Username,
			Notes:               m.Edges.User.Notes,
			PasswordHash:        m.Edges.User.PasswordHash,
			Role:                m.Edges.User.Role,
			Balance:             m.Edges.User.Balance,
			Concurrency:         m.Edges.User.Concurrency,
			Status:              m.Edges.User.Status,
			TotpSecretEncrypted: m.Edges.User.TotpSecretEncrypted,
			TotpEnabled:         m.Edges.User.TotpEnabled,
			TotpEnabledAt:       m.Edges.User.TotpEnabledAt,
			CreatedAt:           m.Edges.User.CreatedAt,
			UpdatedAt:           m.Edges.User.UpdatedAt,
		}
	}
	if m.Edges.Group != nil {
		out.Group = &service.Group{
			ID:                              m.Edges.Group.ID,
			Name:                            m.Edges.Group.Name,
			Description:                     derefString(m.Edges.Group.Description),
			Platform:                        m.Edges.Group.Platform,
			RateMultiplier:                  m.Edges.Group.RateMultiplier,
			IsExclusive:                     m.Edges.Group.IsExclusive,
			Status:                          m.Edges.Group.Status,
			Hydrated:                        true,
			SubscriptionType:                m.Edges.Group.SubscriptionType,
			DailyLimitUSD:                   m.Edges.Group.DailyLimitUsd,
			WeeklyLimitUSD:                  m.Edges.Group.WeeklyLimitUsd,
			MonthlyLimitUSD:                 m.Edges.Group.MonthlyLimitUsd,
			ImagePrice1K:                    m.Edges.Group.ImagePrice1k,
			ImagePrice2K:                    m.Edges.Group.ImagePrice2k,
			ImagePrice4K:                    m.Edges.Group.ImagePrice4k,
			DefaultValidityDays:             m.Edges.Group.DefaultValidityDays,
			ClaudeCodeOnly:                  m.Edges.Group.ClaudeCodeOnly,
			FallbackGroupID:                 m.Edges.Group.FallbackGroupID,
			FallbackGroupIDOnInvalidRequest: m.Edges.Group.FallbackGroupIDOnInvalidRequest,
			ModelRouting:                    m.Edges.Group.ModelRouting,
			ModelRoutingEnabled:             m.Edges.Group.ModelRoutingEnabled,
			MCPXMLInject:                    m.Edges.Group.McpXMLInject,
			SupportedModelScopes:            m.Edges.Group.SupportedModelScopes,
			SortOrder:                       m.Edges.Group.SortOrder,
			AllowMessagesDispatch:           m.Edges.Group.AllowMessagesDispatch,
			DefaultMappedModel:              m.Edges.Group.DefaultMappedModel,
			CreatedAt:                       m.Edges.Group.CreatedAt,
			UpdatedAt:                       m.Edges.Group.UpdatedAt,
		}
	}
	return out
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func normalizeCompatibleKeyMap(item map[string]any) map[string]any {
	if item == nil {
		return nil
	}
	if quotaUsed, ok := item["quota_used"]; ok {
		item["used_quota"] = quotaUsed
	}
	return item
}

func normalizeCompatibleUsageMap(item map[string]any) map[string]any {
	if item == nil {
		return nil
	}
	if totalCost, ok := item["total_cost"]; ok {
		item["billable_cost"] = totalCost
	}
	return item
}

func copyTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
