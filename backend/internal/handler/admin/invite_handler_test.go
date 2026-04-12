//go:build unit

package admin

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func performAdminJSONRequest(t *testing.T, method, path string, body any, handlerFn func(*gin.Context)) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1})
	c.Set(string(middleware.ContextKeyUserRole), service.RoleAdmin)

	handlerFn(c)
	return w
}

func TestInviteHandler_GetStats(t *testing.T) {
	h := NewInviteHandler(newStubAdminService())
	w := performAdminJSONRequest(t, http.MethodGet, "/api/v1/admin/invites/stats", nil, func(c *gin.Context) {
		h.GetStats(c)
	})

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"total_invited_users":3`)
}

func TestInviteHandler_RebindRequiresReason(t *testing.T) {
	h := NewInviteHandler(newStubAdminService())
	w := performAdminJSONRequest(t, http.MethodPost, "/api/v1/admin/invites/rebind", map[string]any{
		"invitee_user_id":     8,
		"new_inviter_user_id": 9,
		"reason":              "",
	}, func(c *gin.Context) {
		h.Rebind(c)
	})

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInviteHandler_CreateManualGrantRejectsNonPositiveRewardAmount(t *testing.T) {
	h := NewInviteHandler(newStubAdminService())
	w := performAdminJSONRequest(t, http.MethodPost, "/api/v1/admin/invites/manual-grants", map[string]any{
		"target_user_id": 9,
		"reason":         "manual correction",
		"lines": []map[string]any{
			{
				"inviter_user_id":       1,
				"invitee_user_id":       2,
				"reward_target_user_id": 1,
				"reward_role":           service.InviteRewardRoleInviter,
				"reward_amount":         -1,
			},
		},
	}, func(c *gin.Context) {
		h.CreateManualGrant(c)
	})

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInviteHandler_CreateManualGrantRejectsUnknownRewardRole(t *testing.T) {
	h := NewInviteHandler(newStubAdminService())
	w := performAdminJSONRequest(t, http.MethodPost, "/api/v1/admin/invites/manual-grants", map[string]any{
		"target_user_id": 9,
		"reason":         "manual correction",
		"lines": []map[string]any{
			{
				"inviter_user_id":       1,
				"invitee_user_id":       2,
				"reward_target_user_id": 1,
				"reward_role":           "unexpected",
				"reward_amount":         1,
			},
		},
	}, func(c *gin.Context) {
		h.CreateManualGrant(c)
	})

	require.Equal(t, http.StatusBadRequest, w.Code)
}
