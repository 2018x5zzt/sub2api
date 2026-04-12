package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func TestRedeemHandler_RedeemCodeStillWorks(t *testing.T) {
	env := newRedeemHandlerTestEnv(t)
	env.userRepo.users[1] = &service.User{
		ID:      1,
		Email:   "redeem@test.local",
		Role:    service.RoleUser,
		Status:  service.StatusActive,
		Balance: 0,
	}
	env.redeemRepo.addCode(&service.RedeemCode{
		ID:     101,
		Code:   "BALANCE-ONLY",
		Type:   service.RedeemTypeBalance,
		Value:  12.5,
		Status: service.StatusUnused,
	})

	rec := performRedeemRequest(t, env.handler, 1, "BALANCE-ONLY")
	require.Equal(t, http.StatusOK, rec.Code)

	var envelope testRedeemEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, 0, envelope.Code)

	var payload RedeemResponse
	require.NoError(t, json.Unmarshal(envelope.Data, &payload))
	require.Equal(t, "Code redeemed successfully", payload.Message)
	require.Equal(t, service.RedeemTypeBalance, payload.Type)
	require.Equal(t, 12.5, payload.Value)

	user, err := env.userRepo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, 12.5, user.Balance)

	code, err := env.redeemRepo.GetByCode(context.Background(), "BALANCE-ONLY")
	require.NoError(t, err)
	require.Equal(t, service.StatusUsed, code.Status)
	require.NotNil(t, code.UsedBy)
	require.EqualValues(t, 1, *code.UsedBy)
}

func TestRedeemHandler_PromoFallbackRedeemsOncePerUser(t *testing.T) {
	env := newRedeemHandlerTestEnv(t)
	env.userRepo.users[1] = &service.User{
		ID:      1,
		Email:   "promo@test.local",
		Role:    service.RoleUser,
		Status:  service.StatusActive,
		Balance: 0,
	}
	env.promoRepo.addCode(&service.PromoCode{
		ID:          201,
		Code:        "HELLO",
		BonusAmount: 8.5,
		Status:      service.PromoCodeStatusActive,
	})

	first := performRedeemRequest(t, env.handler, 1, "hello")
	require.Equal(t, http.StatusOK, first.Code)

	var firstEnvelope testRedeemEnvelope
	require.NoError(t, json.Unmarshal(first.Body.Bytes(), &firstEnvelope))
	require.Equal(t, 0, firstEnvelope.Code)

	var firstPayload RedeemResponse
	require.NoError(t, json.Unmarshal(firstEnvelope.Data, &firstPayload))
	require.Equal(t, "Promo code redeemed successfully", firstPayload.Message)
	require.Equal(t, service.RedeemTypeBalance, firstPayload.Type)
	require.Equal(t, 8.5, firstPayload.Value)

	user, err := env.userRepo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, 8.5, user.Balance)

	second := performRedeemRequest(t, env.handler, 1, "HELLO")
	require.Equal(t, http.StatusConflict, second.Code)

	var secondEnvelope testRedeemEnvelope
	require.NoError(t, json.Unmarshal(second.Body.Bytes(), &secondEnvelope))
	require.Equal(t, http.StatusConflict, secondEnvelope.Code)
	require.Equal(t, "PROMO_CODE_ALREADY_USED", secondEnvelope.Reason)

	user, err = env.userRepo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, 8.5, user.Balance)
}

func TestRedeemHandler_PromoDisabledReturnsRedeemNotFound(t *testing.T) {
	env := newRedeemHandlerTestEnv(t)
	env.userRepo.users[1] = &service.User{
		ID:      1,
		Email:   "disabled@test.local",
		Role:    service.RoleUser,
		Status:  service.StatusActive,
		Balance: 0,
	}
	require.NoError(t, env.settingRepo.Set(context.Background(), service.SettingKeyPromoCodeEnabled, "false"))
	env.promoRepo.addCode(&service.PromoCode{
		ID:          301,
		Code:        "HELLO",
		BonusAmount: 5,
		Status:      service.PromoCodeStatusActive,
	})

	rec := performRedeemRequest(t, env.handler, 1, "HELLO")
	require.Equal(t, http.StatusNotFound, rec.Code)

	var envelope testRedeemEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, http.StatusNotFound, envelope.Code)
	require.Equal(t, "REDEEM_CODE_NOT_FOUND", envelope.Reason)

	user, err := env.userRepo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, 0.0, user.Balance)
}

func TestRedeemHandler_BenefitFallbackWorksWhenPromoSettingDisabled(t *testing.T) {
	env := newRedeemHandlerTestEnv(t)
	env.userRepo.users[1] = &service.User{
		ID:       1,
		Email:    "benefit@test.local",
		Username: "benefit_user",
		Role:     service.RoleUser,
		Status:   service.StatusActive,
		Balance:  2,
	}
	require.NoError(t, env.settingRepo.Set(context.Background(), service.SettingKeyPromoCodeEnabled, "false"))
	env.promoRepo.addCode(&service.PromoCode{
		ID:             401,
		Code:           "HELLO",
		Scene:          service.PromoCodeSceneBenefit,
		BonusAmount:    10,
		Status:         service.PromoCodeStatusActive,
		SuccessMessage: "祝你今天也有好心情。",
	})

	rec := performRedeemRequest(t, env.handler, 1, "HELLO")
	require.Equal(t, http.StatusOK, rec.Code)

	var envelope testRedeemEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, 0, envelope.Code)

	var payload RedeemResponse
	require.NoError(t, json.Unmarshal(envelope.Data, &payload))
	require.Equal(t, service.PromoCodeSceneBenefit, payload.Scene)
	require.Equal(t, "祝你今天也有好心情。", payload.SuccessMessage)
	require.Equal(t, 10.0, payload.Value)

	user, err := env.userRepo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, 12.0, user.Balance)
}

func TestRedeemHandler_BenefitRedPacketReturnsBreakdownAndLeaderboardFlag(t *testing.T) {
	env := newRedeemHandlerTestEnv(t)
	env.userRepo.users[1] = &service.User{
		ID:       1,
		Email:    "lucky@test.local",
		Username: "lucky_user",
		Role:     service.RoleUser,
		Status:   service.StatusActive,
	}
	env.promoRepo.addCode(&service.PromoCode{
		ID:                    402,
		Code:                  "LUCKY100",
		Scene:                 service.PromoCodeSceneBenefit,
		BonusAmount:           20,
		RandomBonusPoolAmount: 1000,
		RandomBonusRemaining:  1000,
		MaxUses:               1,
		LeaderboardEnabled:    true,
		Status:                service.PromoCodeStatusActive,
		SuccessMessage:        "手气不错。",
	})

	rec := performRedeemRequest(t, env.handler, 1, "LUCKY100")
	require.Equal(t, http.StatusOK, rec.Code)

	var envelope testRedeemEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, 0, envelope.Code)

	var payload RedeemResponse
	require.NoError(t, json.Unmarshal(envelope.Data, &payload))
	require.Equal(t, 1020.0, payload.Value)
	require.Equal(t, 20.0, payload.FixedValue)
	require.Equal(t, 1000.0, payload.RandomValue)
	require.Equal(t, 1020.0, payload.TotalValue)
	require.True(t, payload.LeaderboardEnabled)

	user, err := env.userRepo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, 1020.0, user.Balance)
}

func TestRedeemHandler_BenefitRedPacketRequiresUsername(t *testing.T) {
	env := newRedeemHandlerTestEnv(t)
	env.userRepo.users[1] = &service.User{
		ID:     1,
		Email:  "nousername@test.local",
		Role:   service.RoleUser,
		Status: service.StatusActive,
	}
	env.promoRepo.addCode(&service.PromoCode{
		ID:                    403,
		Code:                  "LUCKYNAME",
		Scene:                 service.PromoCodeSceneBenefit,
		BonusAmount:           20,
		RandomBonusPoolAmount: 100,
		RandomBonusRemaining:  100,
		MaxUses:               10,
		LeaderboardEnabled:    true,
		Status:                service.PromoCodeStatusActive,
	})

	rec := performRedeemRequest(t, env.handler, 1, "LUCKYNAME")
	require.Equal(t, http.StatusBadRequest, rec.Code)

	var envelope testRedeemEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, "PROMO_CODE_USERNAME_REQUIRED", envelope.Reason)
}

func TestRedeemHandler_GetBenefitLeaderboardAfterRedeem(t *testing.T) {
	env := newRedeemHandlerTestEnv(t)
	env.userRepo.users[1] = &service.User{
		ID:       1,
		Email:    "rank@test.local",
		Username: "rank_user",
		Role:     service.RoleUser,
		Status:   service.StatusActive,
	}
	env.promoRepo.addCode(&service.PromoCode{
		ID:                    404,
		Code:                  "LUCKYRANK",
		Scene:                 service.PromoCodeSceneBenefit,
		BonusAmount:           20,
		RandomBonusPoolAmount: 100,
		RandomBonusRemaining:  100,
		MaxUses:               1,
		LeaderboardEnabled:    true,
		Status:                service.PromoCodeStatusActive,
	})

	redeemRec := performRedeemRequest(t, env.handler, 1, "LUCKYRANK")
	require.Equal(t, http.StatusOK, redeemRec.Code)

	rec := performBenefitLeaderboardRequest(t, env.handler, 1, "LUCKYRANK")
	require.Equal(t, http.StatusOK, rec.Code)

	var envelope testRedeemEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, 0, envelope.Code)

	var payload BenefitLeaderboardResponse
	require.NoError(t, json.Unmarshal(envelope.Data, &payload))
	require.Len(t, payload.Entries, 1)
	require.Equal(t, "rank_user", payload.Entries[0].DisplayName)
	require.Equal(t, 1, payload.Entries[0].Rank)
	require.NotNil(t, payload.CurrentUserRank)
	require.Equal(t, 1, *payload.CurrentUserRank)
}

func TestRedeemHandler_GetBenefitLeaderboardReturnsAllEntries(t *testing.T) {
	env := newRedeemHandlerTestEnv(t)
	env.promoRepo.addCode(&service.PromoCode{
		ID:                    405,
		Code:                  "LUCKYALL",
		Scene:                 service.PromoCodeSceneBenefit,
		BonusAmount:           1,
		RandomBonusPoolAmount: 100,
		RandomBonusRemaining:  100,
		MaxUses:               25,
		LeaderboardEnabled:    true,
		Status:                service.PromoCodeStatusActive,
	})

	for i := 1; i <= 21; i++ {
		userID := int64(i)
		env.userRepo.users[userID] = &service.User{
			ID:       userID,
			Email:    fmt.Sprintf("rank%d@test.local", i),
			Username: fmt.Sprintf("rank_user_%02d", i),
			Role:     service.RoleUser,
			Status:   service.StatusActive,
		}
		require.NoError(t, env.promoRepo.CreateUsage(context.Background(), &service.PromoCodeUsage{
			PromoCodeID:       405,
			UserID:            userID,
			BonusAmount:       float64(22 - i),
			FixedBonusAmount:  1,
			RandomBonusAmount: float64(21 - i),
			UsedAt:            time.Unix(int64(i), 0).UTC(),
		}))
	}

	rec := performBenefitLeaderboardRequest(t, env.handler, 1, "LUCKYALL")
	require.Equal(t, http.StatusOK, rec.Code)

	var envelope testRedeemEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, 0, envelope.Code)

	var payload BenefitLeaderboardResponse
	require.NoError(t, json.Unmarshal(envelope.Data, &payload))
	require.Len(t, payload.Entries, 21)
	require.Equal(t, "rank_user_01", payload.Entries[0].DisplayName)
	require.Equal(t, 1, payload.Entries[0].Rank)
	require.Equal(t, "rank_user_21", payload.Entries[20].DisplayName)
	require.Equal(t, 21, payload.Entries[20].Rank)
	require.NotNil(t, payload.CurrentUserRank)
	require.Equal(t, 1, *payload.CurrentUserRank)
}

type redeemHandlerTestEnv struct {
	handler     *RedeemHandler
	userRepo    *testRedeemUserRepo
	redeemRepo  *testRedeemCodeRepo
	promoRepo   *testPromoCodeRepo
	settingRepo *testSettingRepo
}

func newRedeemHandlerTestEnv(t *testing.T) *redeemHandlerTestEnv {
	t.Helper()

	client := newRedeemHandlerTestClient(t)
	userRepo := &testRedeemUserRepo{users: make(map[int64]*service.User)}
	redeemRepo := &testRedeemCodeRepo{
		byID:   make(map[int64]*service.RedeemCode),
		byCode: make(map[string]*service.RedeemCode),
	}
	promoRepo := &testPromoCodeRepo{
		byID:      make(map[int64]*service.PromoCode),
		byCode:    make(map[string]*service.PromoCode),
		usages:    make(map[int64]map[int64]*service.PromoCodeUsage),
		nextUseID: 1,
		userRepo:  userRepo,
	}
	settingRepo := &testSettingRepo{all: make(map[string]string)}

	redeemService := service.NewRedeemService(redeemRepo, userRepo, nil, nil, nil, nil, client, nil)
	promoService := service.NewPromoService(promoRepo, userRepo, nil, client, nil)
	settingService := service.NewSettingService(settingRepo, &config.Config{})

	return &redeemHandlerTestEnv{
		handler:     NewRedeemHandler(redeemService, promoService, settingService),
		userRepo:    userRepo,
		redeemRepo:  redeemRepo,
		promoRepo:   promoRepo,
		settingRepo: settingRepo,
	}
}

func newRedeemHandlerTestClient(t *testing.T) *dbent.Client {
	t.Helper()

	db, err := sql.Open("sqlite", "file:redeem_handler_test?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })

	return client
}

func performRedeemRequest(t *testing.T, h *RedeemHandler, userID int64, code string) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	body := []byte(`{"code":"` + code + `"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/redeem", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: userID, Concurrency: 1})

	h.Redeem(c)
	return rec
}

func performBenefitLeaderboardRequest(t *testing.T, h *RedeemHandler, userID int64, code string) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	body := []byte(`{"code":"` + code + `"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/redeem/benefit-leaderboard", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: userID, Concurrency: 1})

	h.GetBenefitLeaderboard(c)
	return rec
}

type testRedeemEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Reason  string          `json:"reason,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type testRedeemUserRepo struct {
	users map[int64]*service.User
}

func (r *testRedeemUserRepo) Create(ctx context.Context, user *service.User) error {
	return errors.New("not implemented")
}

func (r *testRedeemUserRepo) GetByID(ctx context.Context, id int64) (*service.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, service.ErrUserNotFound
	}
	clone := *user
	return &clone, nil
}

func (r *testRedeemUserRepo) GetByEmail(ctx context.Context, email string) (*service.User, error) {
	for _, user := range r.users {
		if user.Email == email {
			clone := *user
			return &clone, nil
		}
	}
	return nil, service.ErrUserNotFound
}

func (r *testRedeemUserRepo) GetByInviteCode(ctx context.Context, code string) (*service.User, error) {
	for _, user := range r.users {
		if user.InviteCode == code {
			clone := *user
			return &clone, nil
		}
	}
	return nil, service.ErrUserNotFound
}

func (r *testRedeemUserRepo) GetFirstAdmin(ctx context.Context) (*service.User, error) {
	return nil, service.ErrUserNotFound
}

func (r *testRedeemUserRepo) Update(ctx context.Context, user *service.User) error {
	return errors.New("not implemented")
}

func (r *testRedeemUserRepo) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (r *testRedeemUserRepo) List(ctx context.Context, params pagination.PaginationParams) ([]service.User, *pagination.PaginationResult, error) {
	return nil, nil, errors.New("not implemented")
}

func (r *testRedeemUserRepo) ListWithFilters(ctx context.Context, params pagination.PaginationParams, filters service.UserListFilters) ([]service.User, *pagination.PaginationResult, error) {
	return nil, nil, errors.New("not implemented")
}

func (r *testRedeemUserRepo) UpdateBalance(ctx context.Context, id int64, amount float64) error {
	user, ok := r.users[id]
	if !ok {
		return service.ErrUserNotFound
	}
	user.Balance += amount
	return nil
}

func (r *testRedeemUserRepo) DeductBalance(ctx context.Context, id int64, amount float64) error {
	user, ok := r.users[id]
	if !ok {
		return service.ErrUserNotFound
	}
	user.Balance -= amount
	return nil
}

func (r *testRedeemUserRepo) UpdateConcurrency(ctx context.Context, id int64, amount int) error {
	user, ok := r.users[id]
	if !ok {
		return service.ErrUserNotFound
	}
	user.Concurrency += amount
	return nil
}

func (r *testRedeemUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	for _, user := range r.users {
		if user.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func (r *testRedeemUserRepo) ExistsByInviteCode(ctx context.Context, code string) (bool, error) {
	for _, user := range r.users {
		if user.InviteCode == code {
			return true, nil
		}
	}
	return false, nil
}

func (r *testRedeemUserRepo) CountInviteesByInviter(ctx context.Context, inviterID int64) (int64, error) {
	var total int64
	for _, user := range r.users {
		if user.InvitedByUserID != nil && *user.InvitedByUserID == inviterID {
			total++
		}
	}
	return total, nil
}

func (r *testRedeemUserRepo) RemoveGroupFromAllowedGroups(ctx context.Context, groupID int64) (int64, error) {
	return 0, errors.New("not implemented")
}

func (r *testRedeemUserRepo) AddGroupToAllowedGroups(ctx context.Context, userID int64, groupID int64) error {
	return errors.New("not implemented")
}

func (r *testRedeemUserRepo) RemoveGroupFromUserAllowedGroups(ctx context.Context, userID int64, groupID int64) error {
	return errors.New("not implemented")
}

func (r *testRedeemUserRepo) UpdateTotpSecret(ctx context.Context, userID int64, encryptedSecret *string) error {
	return errors.New("not implemented")
}

func (r *testRedeemUserRepo) EnableTotp(ctx context.Context, userID int64) error {
	return errors.New("not implemented")
}

func (r *testRedeemUserRepo) DisableTotp(ctx context.Context, userID int64) error {
	return errors.New("not implemented")
}

type testRedeemCodeRepo struct {
	byID   map[int64]*service.RedeemCode
	byCode map[string]*service.RedeemCode
}

func (r *testRedeemCodeRepo) addCode(code *service.RedeemCode) {
	cloned := cloneRedeemCode(code)
	r.byID[cloned.ID] = cloned
	r.byCode[cloned.Code] = cloned
}

func (r *testRedeemCodeRepo) Create(ctx context.Context, code *service.RedeemCode) error {
	return errors.New("not implemented")
}

func (r *testRedeemCodeRepo) CreateBatch(ctx context.Context, codes []service.RedeemCode) error {
	return errors.New("not implemented")
}

func (r *testRedeemCodeRepo) GetByID(ctx context.Context, id int64) (*service.RedeemCode, error) {
	code, ok := r.byID[id]
	if !ok {
		return nil, service.ErrRedeemCodeNotFound
	}
	return cloneRedeemCode(code), nil
}

func (r *testRedeemCodeRepo) GetByCode(ctx context.Context, code string) (*service.RedeemCode, error) {
	redeemCode, ok := r.byCode[code]
	if !ok {
		return nil, service.ErrRedeemCodeNotFound
	}
	return cloneRedeemCode(redeemCode), nil
}

func (r *testRedeemCodeRepo) Update(ctx context.Context, code *service.RedeemCode) error {
	return errors.New("not implemented")
}

func (r *testRedeemCodeRepo) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (r *testRedeemCodeRepo) Use(ctx context.Context, id, userID int64) error {
	code, ok := r.byID[id]
	if !ok {
		return service.ErrRedeemCodeNotFound
	}
	if code.Status != service.StatusUnused {
		return service.ErrRedeemCodeUsed
	}

	now := time.Now().UTC()
	code.Status = service.StatusUsed
	code.UsedBy = &userID
	code.UsedAt = &now
	return nil
}

func (r *testRedeemCodeRepo) List(ctx context.Context, params pagination.PaginationParams) ([]service.RedeemCode, *pagination.PaginationResult, error) {
	return nil, nil, errors.New("not implemented")
}

func (r *testRedeemCodeRepo) ListWithFilters(ctx context.Context, params pagination.PaginationParams, codeType, status, search string) ([]service.RedeemCode, *pagination.PaginationResult, error) {
	return nil, nil, errors.New("not implemented")
}

func (r *testRedeemCodeRepo) ListByUser(ctx context.Context, userID int64, limit int) ([]service.RedeemCode, error) {
	return nil, nil
}

func (r *testRedeemCodeRepo) ListByUserPaginated(ctx context.Context, userID int64, params pagination.PaginationParams, codeType string) ([]service.RedeemCode, *pagination.PaginationResult, error) {
	return nil, nil, errors.New("not implemented")
}

func (r *testRedeemCodeRepo) SumPositiveBalanceByUser(ctx context.Context, userID int64) (float64, error) {
	return 0, errors.New("not implemented")
}

type testPromoCodeRepo struct {
	byID      map[int64]*service.PromoCode
	byCode    map[string]*service.PromoCode
	usages    map[int64]map[int64]*service.PromoCodeUsage
	nextUseID int64
	userRepo  *testRedeemUserRepo
}

func (r *testPromoCodeRepo) addCode(code *service.PromoCode) {
	cloned := clonePromoCode(code)
	normalized := strings.ToUpper(cloned.Code)
	cloned.Code = normalized
	r.byID[cloned.ID] = cloned
	r.byCode[normalized] = cloned
}

func (r *testPromoCodeRepo) Create(ctx context.Context, code *service.PromoCode) error {
	return errors.New("not implemented")
}

func (r *testPromoCodeRepo) GetByID(ctx context.Context, id int64) (*service.PromoCode, error) {
	code, ok := r.byID[id]
	if !ok {
		return nil, service.ErrPromoCodeNotFound
	}
	return clonePromoCode(code), nil
}

func (r *testPromoCodeRepo) GetByCode(ctx context.Context, code string) (*service.PromoCode, error) {
	promoCode, ok := r.byCode[strings.ToUpper(strings.TrimSpace(code))]
	if !ok {
		return nil, service.ErrPromoCodeNotFound
	}
	return clonePromoCode(promoCode), nil
}

func (r *testPromoCodeRepo) GetByCodeForUpdate(ctx context.Context, code string) (*service.PromoCode, error) {
	return r.GetByCode(ctx, code)
}

func (r *testPromoCodeRepo) Update(ctx context.Context, code *service.PromoCode) error {
	if _, ok := r.byID[code.ID]; !ok {
		return service.ErrPromoCodeNotFound
	}
	cloned := clonePromoCode(code)
	normalized := strings.ToUpper(strings.TrimSpace(cloned.Code))
	cloned.Code = normalized
	r.byID[cloned.ID] = cloned
	r.byCode[normalized] = cloned
	return nil
}

func (r *testPromoCodeRepo) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (r *testPromoCodeRepo) List(ctx context.Context, params pagination.PaginationParams) ([]service.PromoCode, *pagination.PaginationResult, error) {
	return nil, nil, errors.New("not implemented")
}

func (r *testPromoCodeRepo) ListWithFilters(ctx context.Context, params pagination.PaginationParams, scene, status, search string) ([]service.PromoCode, *pagination.PaginationResult, error) {
	return nil, nil, errors.New("not implemented")
}

func (r *testPromoCodeRepo) CreateUsage(ctx context.Context, usage *service.PromoCodeUsage) error {
	if r.usages[usage.PromoCodeID] == nil {
		r.usages[usage.PromoCodeID] = make(map[int64]*service.PromoCodeUsage)
	}
	cloned := *usage
	cloned.ID = r.nextUseID
	r.nextUseID++
	r.usages[usage.PromoCodeID][usage.UserID] = &cloned
	usage.ID = cloned.ID
	return nil
}

func (r *testPromoCodeRepo) GetUsageByPromoCodeAndUser(ctx context.Context, promoCodeID, userID int64) (*service.PromoCodeUsage, error) {
	byUser := r.usages[promoCodeID]
	if byUser == nil {
		return nil, nil
	}
	usage, ok := byUser[userID]
	if !ok {
		return nil, nil
	}
	cloned := *usage
	return &cloned, nil
}

func (r *testPromoCodeRepo) ListAllUsagesByPromoCode(ctx context.Context, promoCodeID int64) ([]service.PromoCodeUsage, error) {
	byUser := r.usages[promoCodeID]
	if byUser == nil {
		return nil, nil
	}
	out := make([]service.PromoCodeUsage, 0, len(byUser))
	for _, usage := range byUser {
		cloned := *usage
		if user, ok := r.findUser(usage.UserID); ok {
			cloned.User = user
		}
		out = append(out, cloned)
	}
	return out, nil
}

func (r *testPromoCodeRepo) ListUsagesByPromoCode(ctx context.Context, promoCodeID int64, params pagination.PaginationParams) ([]service.PromoCodeUsage, *pagination.PaginationResult, error) {
	return nil, nil, errors.New("not implemented")
}

func (r *testPromoCodeRepo) IncrementUsedCount(ctx context.Context, id int64) error {
	code, ok := r.byID[id]
	if !ok {
		return service.ErrPromoCodeNotFound
	}
	code.UsedCount++
	return nil
}

func (r *testPromoCodeRepo) findUser(userID int64) (*service.User, bool) {
	if r.userRepo == nil {
		return nil, false
	}
	user, err := r.userRepo.GetByID(context.Background(), userID)
	return user, err == nil
}

type testSettingRepo struct {
	all map[string]string
}

func (r *testSettingRepo) Get(ctx context.Context, key string) (*service.Setting, error) {
	value, ok := r.all[key]
	if !ok {
		return nil, service.ErrSettingNotFound
	}
	return &service.Setting{Key: key, Value: value}, nil
}

func (r *testSettingRepo) GetValue(ctx context.Context, key string) (string, error) {
	value, ok := r.all[key]
	if !ok {
		return "", service.ErrSettingNotFound
	}
	return value, nil
}

func (r *testSettingRepo) Set(ctx context.Context, key, value string) error {
	r.all[key] = value
	return nil
}

func (r *testSettingRepo) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		out[key] = r.all[key]
	}
	return out, nil
}

func (r *testSettingRepo) SetMultiple(ctx context.Context, settings map[string]string) error {
	for key, value := range settings {
		r.all[key] = value
	}
	return nil
}

func (r *testSettingRepo) GetAll(ctx context.Context) (map[string]string, error) {
	out := make(map[string]string, len(r.all))
	for key, value := range r.all {
		out[key] = value
	}
	return out, nil
}

func (r *testSettingRepo) Delete(ctx context.Context, key string) error {
	delete(r.all, key)
	return nil
}

func cloneRedeemCode(code *service.RedeemCode) *service.RedeemCode {
	if code == nil {
		return nil
	}

	cloned := *code
	if code.UsedBy != nil {
		usedBy := *code.UsedBy
		cloned.UsedBy = &usedBy
	}
	if code.UsedAt != nil {
		usedAt := *code.UsedAt
		cloned.UsedAt = &usedAt
	}
	if code.GroupID != nil {
		groupID := *code.GroupID
		cloned.GroupID = &groupID
	}
	if code.User != nil {
		user := *code.User
		cloned.User = &user
	}
	if code.Group != nil {
		group := *code.Group
		cloned.Group = &group
	}
	return &cloned
}

func clonePromoCode(code *service.PromoCode) *service.PromoCode {
	if code == nil {
		return nil
	}

	cloned := *code
	if code.ExpiresAt != nil {
		expiresAt := *code.ExpiresAt
		cloned.ExpiresAt = &expiresAt
	}
	if len(code.UsageRecords) > 0 {
		cloned.UsageRecords = append([]service.PromoCodeUsage(nil), code.UsageRecords...)
	}
	return &cloned
}
