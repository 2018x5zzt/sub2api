package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newCreateAndRedeemHandler creates a RedeemHandler with a non-nil (but minimal)
// RedeemService so that CreateAndRedeem's nil guard passes and we can test the
// parameter-validation layer that runs before any service call.
func newCreateAndRedeemHandler() *RedeemHandler {
	return &RedeemHandler{
		adminService:  newStubAdminService(),
		redeemService: &service.RedeemService{}, // non-nil to pass nil guard
	}
}

func newGenerateHandler() (*RedeemHandler, *stubAdminService) {
	adminSvc := newStubAdminService()
	return &RedeemHandler{
		adminService:  adminSvc,
		redeemService: &service.RedeemService{},
	}, adminSvc
}

// postCreateAndRedeemValidation calls CreateAndRedeem and returns the response
// status code. For cases that pass validation and proceed into the service layer,
// a panic may occur (because RedeemService internals are nil); this is expected
// and treated as "validation passed" (returns 0 to indicate panic).
func postCreateAndRedeemValidation(t *testing.T, handler *RedeemHandler, body any) (code int) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	jsonBytes, err := json.Marshal(body)
	require.NoError(t, err)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/admin/redeem-codes/create-and-redeem", bytes.NewReader(jsonBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	defer func() {
		if r := recover(); r != nil {
			// Panic means we passed validation and entered service layer (expected for minimal stub).
			code = 0
		}
	}()
	handler.CreateAndRedeem(c)
	return w.Code
}

func postGenerate(t *testing.T, handler *RedeemHandler, body any) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	jsonBytes, err := json.Marshal(body)
	require.NoError(t, err)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/admin/redeem-codes/generate", bytes.NewReader(jsonBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Generate(c)
	return w
}

func TestCreateAndRedeem_TypeDefaultsToBalance(t *testing.T) {
	// 不传 type 字段时应默认 balance，不触发 subscription 校验。
	// 验证通过后进入 service 层会 panic（返回 0），说明默认值生效。
	h := newCreateAndRedeemHandler()
	code := postCreateAndRedeemValidation(t, h, map[string]any{
		"code":    "test-balance-default",
		"value":   10.0,
		"user_id": 1,
	})

	assert.NotEqual(t, http.StatusBadRequest, code,
		"omitting type should default to balance and pass validation")
}

func TestCreateAndRedeem_SubscriptionRequiresGroupID(t *testing.T) {
	h := newCreateAndRedeemHandler()
	code := postCreateAndRedeemValidation(t, h, map[string]any{
		"code":          "test-sub-no-group",
		"type":          "subscription",
		"value":         29.9,
		"user_id":       1,
		"validity_days": 30,
		// group_id 缺失
	})

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestCreateAndRedeem_SubscriptionRequiresPositiveValidityDays(t *testing.T) {
	groupID := int64(5)
	h := newCreateAndRedeemHandler()

	cases := []struct {
		name         string
		validityDays int
	}{
		{"zero", 0},
		{"negative", -1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code := postCreateAndRedeemValidation(t, h, map[string]any{
				"code":          "test-sub-bad-days-" + tc.name,
				"type":          "subscription",
				"value":         29.9,
				"user_id":       1,
				"group_id":      groupID,
				"validity_days": tc.validityDays,
			})

			assert.Equal(t, http.StatusBadRequest, code)
		})
	}
}

func TestCreateAndRedeem_SubscriptionValidParamsPassValidation(t *testing.T) {
	groupID := int64(5)
	h := newCreateAndRedeemHandler()
	code := postCreateAndRedeemValidation(t, h, map[string]any{
		"code":          "test-sub-valid",
		"type":          "subscription",
		"value":         29.9,
		"user_id":       1,
		"group_id":      groupID,
		"validity_days": 31,
	})

	assert.NotEqual(t, http.StatusBadRequest, code,
		"valid subscription params should pass validation")
}

func TestCreateAndRedeem_BalanceIgnoresSubscriptionFields(t *testing.T) {
	h := newCreateAndRedeemHandler()
	// balance 类型不传 group_id 和 validity_days，不应报 400
	code := postCreateAndRedeemValidation(t, h, map[string]any{
		"code":    "test-balance-no-extras",
		"type":    "balance",
		"value":   50.0,
		"user_id": 1,
	})

	assert.NotEqual(t, http.StatusBadRequest, code,
		"balance type should not require group_id or validity_days")
}

func TestCreateAndRedeem_RejectsLegacyInvitationType(t *testing.T) {
	h := newCreateAndRedeemHandler()
	code := postCreateAndRedeemValidation(t, h, map[string]any{
		"code":    "legacy-invitation",
		"type":    "invitation",
		"value":   0,
		"user_id": 1,
	})

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestGenerateRedeemCodes_DefaultsSourceTypeToSystemGrant(t *testing.T) {
	h, adminSvc := newGenerateHandler()
	w := postGenerate(t, h, map[string]any{
		"count": 1,
		"type":  "balance",
		"value": 10,
	})

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, adminSvc.lastGenerateRedeem)
	require.Equal(t, service.RedeemSourceSystemGrant, adminSvc.lastGenerateRedeem.SourceType)
}

func TestGenerateRedeemCodes_PassesExplicitCommercialSourceType(t *testing.T) {
	h, adminSvc := newGenerateHandler()
	w := postGenerate(t, h, map[string]any{
		"count":       1,
		"type":        "balance",
		"value":       10,
		"source_type": "commercial",
	})

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, adminSvc.lastGenerateRedeem)
	require.Equal(t, service.RedeemSourceCommercial, adminSvc.lastGenerateRedeem.SourceType)
}

func TestGenerateRedeemCodes_RejectsLegacyInvitationType(t *testing.T) {
	h, _ := newGenerateHandler()
	w := postGenerate(t, h, map[string]any{
		"count": 1,
		"type":  "invitation",
		"value": 0,
	})

	require.Equal(t, http.StatusBadRequest, w.Code)
}
