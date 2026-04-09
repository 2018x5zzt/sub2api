package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"math/big"
	"sort"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrPromoCodeNotFound               = infraerrors.NotFound("PROMO_CODE_NOT_FOUND", "promo code not found")
	ErrPromoCodeExpired                = infraerrors.BadRequest("PROMO_CODE_EXPIRED", "promo code has expired")
	ErrPromoCodeDisabled               = infraerrors.BadRequest("PROMO_CODE_DISABLED", "promo code is disabled")
	ErrPromoCodeMaxUsed                = infraerrors.BadRequest("PROMO_CODE_MAX_USED", "promo code has reached maximum uses")
	ErrPromoCodeAlreadyUsed            = infraerrors.Conflict("PROMO_CODE_ALREADY_USED", "you have already used this promo code")
	ErrPromoCodeInvalid                = infraerrors.BadRequest("PROMO_CODE_INVALID", "invalid promo code")
	ErrPromoCodeInvalidRandomConfig    = infraerrors.BadRequest("PROMO_CODE_INVALID_RANDOM_CONFIG", "invalid benefit red packet configuration")
	ErrPromoCodeLeaderboardUnavailable = infraerrors.BadRequest("PROMO_CODE_LEADERBOARD_UNAVAILABLE", "leaderboard is not available for this benefit code")
	ErrPromoCodeLeaderboardForbidden   = infraerrors.Forbidden("PROMO_CODE_LEADERBOARD_FORBIDDEN", "redeem this benefit code before viewing the leaderboard")
	ErrPromoCodeUsernameRequired       = infraerrors.BadRequest("PROMO_CODE_USERNAME_REQUIRED", "set a username before redeeming this red packet")
)

// PromoService 优惠码服务
type PromoService struct {
	promoRepo            PromoCodeRepository
	userRepo             UserRepository
	billingCacheService  *BillingCacheService
	entClient            *dbent.Client
	authCacheInvalidator APIKeyAuthCacheInvalidator
	randomReader         io.Reader
}

// NewPromoService 创建优惠码服务实例
func NewPromoService(
	promoRepo PromoCodeRepository,
	userRepo UserRepository,
	billingCacheService *BillingCacheService,
	entClient *dbent.Client,
	authCacheInvalidator APIKeyAuthCacheInvalidator,
) *PromoService {
	return &PromoService{
		promoRepo:            promoRepo,
		userRepo:             userRepo,
		billingCacheService:  billingCacheService,
		entClient:            entClient,
		authCacheInvalidator: authCacheInvalidator,
		randomReader:         rand.Reader,
	}
}

func normalizePromoCodeScene(scene string) string {
	switch strings.ToLower(strings.TrimSpace(scene)) {
	case "", PromoCodeSceneRegister:
		return PromoCodeSceneRegister
	case PromoCodeSceneBenefit:
		return PromoCodeSceneBenefit
	default:
		return ""
	}
}

func (s *PromoService) ensurePromoCodeScene(promoCode *PromoCode, scene string) error {
	if promoCode == nil {
		return nil
	}
	if normalizePromoCodeScene(promoCode.Scene) != scene {
		return ErrPromoCodeNotFound
	}
	return nil
}

// ValidatePromoCode 验证优惠码（注册前调用）
// 返回 nil, nil 表示空码（不报错）
func (s *PromoService) ValidatePromoCode(ctx context.Context, code string) (*PromoCode, error) {
	return s.validatePromoCodeInScene(ctx, code, PromoCodeSceneRegister)
}

// ValidateBenefitCode 验证福利码（用户兑换入口调用）
func (s *PromoService) ValidateBenefitCode(ctx context.Context, code string) (*PromoCode, error) {
	return s.validatePromoCodeInScene(ctx, code, PromoCodeSceneBenefit)
}

func (s *PromoService) validatePromoCodeInScene(ctx context.Context, code string, scene string) (*PromoCode, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, nil // 空码不报错，直接返回
	}

	promoCode, err := s.promoRepo.GetByCode(ctx, code)
	if err != nil {
		// 保留原始错误类型，不要统一映射为 NotFound
		return nil, err
	}
	if err := s.ensurePromoCodeScene(promoCode, scene); err != nil {
		return nil, err
	}

	if err := s.validatePromoCodeStatus(promoCode); err != nil {
		return nil, err
	}

	return promoCode, nil
}

// validatePromoCodeStatus 验证优惠码状态
func (s *PromoService) validatePromoCodeStatus(promoCode *PromoCode) error {
	if !promoCode.CanUse() {
		if promoCode.IsExpired() {
			return ErrPromoCodeExpired
		}
		if promoCode.Status == PromoCodeStatusDisabled {
			return ErrPromoCodeDisabled
		}
		if promoCode.MaxUses > 0 && promoCode.UsedCount >= promoCode.MaxUses {
			return ErrPromoCodeMaxUsed
		}
		return ErrPromoCodeInvalid
	}
	return nil
}

func isBenefitRedPacket(promoCode *PromoCode) bool {
	return promoCode != nil &&
		normalizePromoCodeScene(promoCode.Scene) == PromoCodeSceneBenefit &&
		promoCode.RandomBonusPoolAmount > 0
}

func validatePromoCodeConfig(scene string, fixedBonusAmount, randomBonusPoolAmount float64, maxUses int, leaderboardEnabled bool) error {
	if fixedBonusAmount < 0 || randomBonusPoolAmount < 0 || maxUses < 0 {
		return ErrPromoCodeInvalidRandomConfig
	}

	switch scene {
	case PromoCodeSceneRegister:
		if randomBonusPoolAmount > 0 || leaderboardEnabled {
			return ErrPromoCodeInvalidRandomConfig
		}
	case PromoCodeSceneBenefit:
		if randomBonusPoolAmount > 0 && maxUses <= 0 {
			return ErrPromoCodeInvalidRandomConfig
		}
		if leaderboardEnabled && randomBonusPoolAmount <= 0 {
			return ErrPromoCodeInvalidRandomConfig
		}
	default:
		return ErrPromoCodeInvalid
	}

	return nil
}

// ApplyPromoCode 应用优惠码（注册成功后调用）
// 使用事务和行锁确保并发安全
func (s *PromoService) ApplyPromoCode(ctx context.Context, userID int64, code string) error {
	_, err := s.applyPromoCodeInScene(ctx, userID, code, PromoCodeSceneRegister)
	return err
}

// ApplyBenefitCode 应用福利码（已注册用户兑换入口调用）
func (s *PromoService) ApplyBenefitCode(ctx context.Context, userID int64, code string) error {
	_, err := s.applyPromoCodeInScene(ctx, userID, code, PromoCodeSceneBenefit)
	return err
}

func (s *PromoService) RedeemBenefitCode(ctx context.Context, userID int64, code string) (*BenefitCodeRedeemResult, error) {
	return s.applyPromoCodeInScene(ctx, userID, code, PromoCodeSceneBenefit)
}

func (s *PromoService) applyPromoCodeInScene(ctx context.Context, userID int64, code string, scene string) (*BenefitCodeRedeemResult, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, nil
	}

	// 开启事务
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)

	// 在事务中获取并锁定优惠码记录（FOR UPDATE）
	promoCode, err := s.promoRepo.GetByCodeForUpdate(txCtx, code)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePromoCodeScene(promoCode, scene); err != nil {
		return nil, err
	}

	// 在事务中验证优惠码状态
	if err := s.validatePromoCodeStatus(promoCode); err != nil {
		return nil, err
	}

	// 在事务中检查用户是否已使用过此优惠码
	existing, err := s.promoRepo.GetUsageByPromoCodeAndUser(txCtx, promoCode.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("check existing usage: %w", err)
	}
	if existing != nil {
		return nil, ErrPromoCodeAlreadyUsed
	}

	if isBenefitRedPacket(promoCode) {
		user, getUserErr := s.userRepo.GetByID(txCtx, userID)
		if getUserErr != nil {
			return nil, fmt.Errorf("get user: %w", getUserErr)
		}
		if strings.TrimSpace(user.Username) == "" {
			return nil, ErrPromoCodeUsernameRequired
		}
	}

	fixedBonusAmount := promoCode.BonusAmount
	randomBonusAmount := 0.0
	if isBenefitRedPacket(promoCode) {
		randomBonusAmount, err = s.allocateBenefitRandomBonus(promoCode)
		if err != nil {
			return nil, err
		}
	}
	totalBonusAmount := fixedBonusAmount + randomBonusAmount

	// 增加用户余额
	if totalBonusAmount != 0 {
		if err := s.userRepo.UpdateBalance(txCtx, userID, totalBonusAmount); err != nil {
			return nil, fmt.Errorf("update user balance: %w", err)
		}
	}

	// 创建使用记录
	usage := &PromoCodeUsage{
		PromoCodeID:       promoCode.ID,
		UserID:            userID,
		BonusAmount:       totalBonusAmount,
		FixedBonusAmount:  fixedBonusAmount,
		RandomBonusAmount: randomBonusAmount,
		UsedAt:            time.Now(),
	}
	if err := s.promoRepo.CreateUsage(txCtx, usage); err != nil {
		return nil, fmt.Errorf("create usage record: %w", err)
	}

	promoCode.UsedCount++
	if isBenefitRedPacket(promoCode) {
		remainingCents := promoAmountToCents(promoCode.RandomBonusRemaining) - promoAmountToCents(randomBonusAmount)
		if remainingCents < 0 {
			remainingCents = 0
		}
		promoCode.RandomBonusRemaining = promoCentsToAmount(remainingCents)
	}

	if err := s.promoRepo.Update(txCtx, promoCode); err != nil {
		return nil, fmt.Errorf("update promo code after redeem: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	s.invalidatePromoCaches(ctx, userID, totalBonusAmount)

	// 失效余额缓存
	if s.billingCacheService != nil {
		go func() {
			cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.billingCacheService.InvalidateUserBalance(cacheCtx, userID)
		}()
	}

	return &BenefitCodeRedeemResult{
		PromoCode:         promoCode,
		Usage:             usage,
		FixedBonusAmount:  fixedBonusAmount,
		RandomBonusAmount: randomBonusAmount,
		TotalBonusAmount:  totalBonusAmount,
	}, nil
}

func (s *PromoService) invalidatePromoCaches(ctx context.Context, userID int64, bonusAmount float64) {
	if bonusAmount == 0 || s.authCacheInvalidator == nil {
		return
	}
	s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
}

// GenerateRandomCode 生成随机优惠码
func (s *PromoService) GenerateRandomCode() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return strings.ToUpper(hex.EncodeToString(bytes)), nil
}

func promoAmountToCents(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

func promoCentsToAmount(cents int64) float64 {
	return float64(cents) / 100
}

func (s *PromoService) randomInt64Inclusive(min, max int64) (int64, error) {
	if max < min {
		return 0, ErrPromoCodeInvalidRandomConfig
	}
	if max == min {
		return min, nil
	}
	reader := s.randomReader
	if reader == nil {
		reader = rand.Reader
	}
	n, err := rand.Int(reader, big.NewInt(max-min+1))
	if err != nil {
		return 0, fmt.Errorf("draw benefit random bonus: %w", err)
	}
	return min + n.Int64(), nil
}

func (s *PromoService) allocateBenefitRandomBonus(promoCode *PromoCode) (float64, error) {
	if promoCode == nil || promoCode.RandomBonusPoolAmount <= 0 {
		return 0, nil
	}

	remainingPoolCents := promoAmountToCents(promoCode.RandomBonusRemaining)
	if remainingPoolCents <= 0 {
		return 0, nil
	}

	remainingClaimants := int64(1)
	if promoCode.MaxUses > 0 {
		remainingClaimants = int64(promoCode.MaxUses - promoCode.UsedCount)
		if remainingClaimants < 1 {
			remainingClaimants = 1
		}
	}

	if remainingClaimants <= 1 {
		return promoCentsToAmount(remainingPoolCents), nil
	}

	minCurrentCents := int64(0)
	reservedForOthersCents := int64(0)
	if remainingPoolCents >= remainingClaimants {
		minCurrentCents = 1
		reservedForOthersCents = remainingClaimants - 1
	}

	maxCurrentCents := remainingPoolCents - reservedForOthersCents
	avgCapCents := (remainingPoolCents * 2) / remainingClaimants
	if avgCapCents > 0 && avgCapCents < maxCurrentCents {
		maxCurrentCents = avgCapCents
	}
	if maxCurrentCents < minCurrentCents {
		maxCurrentCents = minCurrentCents
	}

	drawnCents, err := s.randomInt64Inclusive(minCurrentCents, maxCurrentCents)
	if err != nil {
		return 0, err
	}
	return promoCentsToAmount(drawnCents), nil
}

func benefitLeaderboardDisplayName(user *User, fallbackUserID int64) string {
	if user != nil {
		if username := strings.TrimSpace(user.Username); username != "" {
			return username
		}
		if email := strings.TrimSpace(user.Email); email != "" {
			return MaskEmail(email)
		}
	}
	return fmt.Sprintf("User #%d", fallbackUserID)
}

func rankBenefitUsages(usages []PromoCodeUsage, currentUserID int64, limit int) *BenefitLeaderboardResult {
	sorted := append([]PromoCodeUsage(nil), usages...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].RandomBonusAmount == sorted[j].RandomBonusAmount {
			return sorted[i].UsedAt.Before(sorted[j].UsedAt)
		}
		return sorted[i].RandomBonusAmount > sorted[j].RandomBonusAmount
	})

	if limit <= 0 {
		limit = len(sorted)
	}

	result := &BenefitLeaderboardResult{
		Entries: make([]BenefitLeaderboardEntry, 0, minInt(limit, len(sorted))),
	}

	for idx, usage := range sorted {
		rank := idx + 1
		entry := BenefitLeaderboardEntry{
			Rank:              rank,
			DisplayName:       benefitLeaderboardDisplayName(usage.User, usage.UserID),
			FixedBonusAmount:  usage.FixedBonusAmount,
			RandomBonusAmount: usage.RandomBonusAmount,
			TotalBonusAmount:  usage.BonusAmount,
			UsedAt:            usage.UsedAt,
			IsCurrentUser:     usage.UserID == currentUserID,
		}
		if rank <= limit {
			result.Entries = append(result.Entries, entry)
		}
		if usage.UserID == currentUserID {
			rankCopy := rank
			randomCopy := usage.RandomBonusAmount
			totalCopy := usage.BonusAmount
			result.CurrentUserRank = &rankCopy
			result.CurrentUserRandomAmount = &randomCopy
			result.CurrentUserTotalAmount = &totalCopy
		}
	}

	return result
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Create 创建优惠码
func (s *PromoService) Create(ctx context.Context, input *CreatePromoCodeInput) (*PromoCode, error) {
	code := strings.TrimSpace(input.Code)
	scene := normalizePromoCodeScene(input.Scene)
	if scene == "" {
		return nil, ErrPromoCodeInvalid
	}
	if err := validatePromoCodeConfig(scene, input.BonusAmount, input.RandomBonusPoolAmount, input.MaxUses, input.LeaderboardEnabled); err != nil {
		return nil, err
	}
	if code == "" {
		// 自动生成
		var err error
		code, err = s.GenerateRandomCode()
		if err != nil {
			return nil, err
		}
	}

	promoCode := &PromoCode{
		Code:                  strings.ToUpper(code),
		Scene:                 scene,
		BonusAmount:           input.BonusAmount,
		RandomBonusPoolAmount: input.RandomBonusPoolAmount,
		RandomBonusRemaining:  input.RandomBonusPoolAmount,
		MaxUses:               input.MaxUses,
		UsedCount:             0,
		LeaderboardEnabled:    input.LeaderboardEnabled,
		Status:                PromoCodeStatusActive,
		ExpiresAt:             input.ExpiresAt,
		SuccessMessage:        strings.TrimSpace(input.SuccessMessage),
		Notes:                 input.Notes,
	}

	if err := s.promoRepo.Create(ctx, promoCode); err != nil {
		return nil, fmt.Errorf("create promo code: %w", err)
	}

	return promoCode, nil
}

// GetByID 根据ID获取优惠码
func (s *PromoService) GetByID(ctx context.Context, id int64) (*PromoCode, error) {
	code, err := s.promoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return code, nil
}

// Update 更新优惠码
func (s *PromoService) Update(ctx context.Context, id int64, input *UpdatePromoCodeInput) (*PromoCode, error) {
	promoCode, err := s.promoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Code != nil {
		promoCode.Code = strings.ToUpper(strings.TrimSpace(*input.Code))
	}
	if normalizePromoCodeScene(promoCode.Scene) == "" {
		promoCode.Scene = PromoCodeSceneRegister
	}
	if input.BonusAmount != nil {
		promoCode.BonusAmount = *input.BonusAmount
	}
	if input.MaxUses != nil {
		promoCode.MaxUses = *input.MaxUses
	}
	if input.RandomBonusPoolAmount != nil {
		usages, listErr := s.promoRepo.ListAllUsagesByPromoCode(ctx, promoCode.ID)
		if listErr != nil {
			return nil, fmt.Errorf("list promo usages: %w", listErr)
		}
		var usedRandomCents int64
		for i := range usages {
			usedRandomCents += promoAmountToCents(usages[i].RandomBonusAmount)
		}
		newPoolCents := promoAmountToCents(*input.RandomBonusPoolAmount)
		if newPoolCents < usedRandomCents {
			return nil, ErrPromoCodeInvalidRandomConfig
		}
		promoCode.RandomBonusPoolAmount = *input.RandomBonusPoolAmount
		promoCode.RandomBonusRemaining = promoCentsToAmount(newPoolCents - usedRandomCents)
	}
	if input.LeaderboardEnabled != nil {
		promoCode.LeaderboardEnabled = *input.LeaderboardEnabled
	}
	if input.Status != nil {
		promoCode.Status = *input.Status
	}
	if input.ExpiresAt != nil {
		promoCode.ExpiresAt = input.ExpiresAt
	}
	if input.SuccessMessage != nil {
		promoCode.SuccessMessage = strings.TrimSpace(*input.SuccessMessage)
	}
	if input.Notes != nil {
		promoCode.Notes = *input.Notes
	}
	if promoCode.MaxUses > 0 && promoCode.UsedCount > promoCode.MaxUses {
		return nil, ErrPromoCodeInvalidRandomConfig
	}
	if err := validatePromoCodeConfig(
		normalizePromoCodeScene(promoCode.Scene),
		promoCode.BonusAmount,
		promoCode.RandomBonusPoolAmount,
		promoCode.MaxUses,
		promoCode.LeaderboardEnabled,
	); err != nil {
		return nil, err
	}

	if err := s.promoRepo.Update(ctx, promoCode); err != nil {
		return nil, fmt.Errorf("update promo code: %w", err)
	}

	return promoCode, nil
}

// Delete 删除优惠码
func (s *PromoService) Delete(ctx context.Context, id int64) error {
	if err := s.promoRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete promo code: %w", err)
	}
	return nil
}

// List 获取优惠码列表
func (s *PromoService) List(ctx context.Context, params pagination.PaginationParams, scene, status, search string) ([]PromoCode, *pagination.PaginationResult, error) {
	normalizedScene := normalizePromoCodeScene(scene)
	if normalizedScene == "" {
		normalizedScene = PromoCodeSceneRegister
	}
	return s.promoRepo.ListWithFilters(ctx, params, normalizedScene, status, search)
}

// ListUsages 获取使用记录
func (s *PromoService) ListUsages(ctx context.Context, promoCodeID int64, params pagination.PaginationParams) ([]PromoCodeUsage, *pagination.PaginationResult, error) {
	return s.promoRepo.ListUsagesByPromoCode(ctx, promoCodeID, params)
}

func (s *PromoService) GetBenefitLeaderboard(ctx context.Context, userID int64, code string, limit int) (*BenefitLeaderboardResult, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, ErrPromoCodeNotFound
	}

	promoCode, err := s.promoRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePromoCodeScene(promoCode, PromoCodeSceneBenefit); err != nil {
		return nil, err
	}
	if !promoCode.LeaderboardEnabled || promoCode.RandomBonusPoolAmount <= 0 {
		return nil, ErrPromoCodeLeaderboardUnavailable
	}

	usage, err := s.promoRepo.GetUsageByPromoCodeAndUser(ctx, promoCode.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("get promo usage by user: %w", err)
	}
	if usage == nil {
		return nil, ErrPromoCodeLeaderboardForbidden
	}

	usages, err := s.promoRepo.ListAllUsagesByPromoCode(ctx, promoCode.ID)
	if err != nil {
		return nil, fmt.Errorf("list promo usages for leaderboard: %w", err)
	}

	result := rankBenefitUsages(usages, userID, limit)
	result.PromoCode = promoCode
	return result, nil
}
