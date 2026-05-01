package enterprisebff

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type currentUser struct {
	ID          int64   `json:"id"`
	Email       string  `json:"email"`
	Username    string  `json:"username"`
	Role        string  `json:"role"`
	Balance     float64 `json:"balance"`
	Concurrency int     `json:"concurrency"`
	Status      string  `json:"status"`
	RunMode     string  `json:"run_mode,omitempty"`
}

type envelope[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type enterpriseLoginRequest struct {
	CompanyName    string `json:"company_name" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required"`
	TurnstileToken string `json:"turnstile_token"`
}

type coreLoginRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	TurnstileToken string `json:"turnstile_token,omitempty"`
}

type enterpriseLogin2FARequest struct {
	CompanyName string `json:"company_name" binding:"required"`
	TempToken   string `json:"temp_token" binding:"required"`
	TotpCode    string `json:"totp_code" binding:"required"`
}

type coreLogin2FARequest struct {
	TempToken string `json:"temp_token"`
	TotpCode  string `json:"totp_code"`
}

func (s *Server) currentUserFromCore(ctx context.Context, headers http.Header) (*currentUser, *upstreamResponse, error) {
	resp, err := s.doUpstreamRequest(ctx, http.MethodGet, "/auth/me", "", headers, nil)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp, nil
	}

	var payload envelope[currentUser]
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		return nil, resp, err
	}
	if payload.Code != 0 {
		return nil, resp, nil
	}
	return &payload.Data, resp, nil
}

func (s *Server) requireCurrentUser(c *gin.Context) (*currentUser, bool) {
	user, resp, err := s.currentUserFromCore(c.Request.Context(), c.Request.Header)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    http.StatusBadGateway,
			"message": "Failed to verify current user",
		})
		return nil, false
	}
	if user == nil {
		if resp != nil {
			writeUpstreamResponse(c, resp)
			return nil, false
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "Unauthorized",
		})
		return nil, false
	}
	if s.enterpriseStore != nil {
		profile, profileErr := s.enterpriseStore.GetByUserID(c.Request.Context(), user.ID)
		if profileErr != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"code":    http.StatusBadGateway,
				"message": "Failed to resolve enterprise profile",
			})
			return nil, false
		}
		if profile == nil {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "Enterprise access required",
			})
			return nil, false
		}
		c.Set("enterprise_profile", profile)
	}
	return user, true
}

func (s *Server) requireAdmin(c *gin.Context) (*currentUser, bool) {
	user, ok := s.requireCurrentUser(c)
	if !ok {
		return nil, false
	}
	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    http.StatusForbidden,
			"message": "Admin access required",
		})
		return nil, false
	}
	return user, true
}

func (s *Server) handleEnterpriseLogin(c *gin.Context) {
	var req enterpriseLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Invalid request payload",
		})
		return
	}

	profile, err := s.enterpriseStore.MatchUserByEmailAndCompany(c.Request.Context(), req.Email, req.CompanyName)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    http.StatusBadGateway,
			"message": "Failed to resolve enterprise profile",
		})
		return
	}
	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "Invalid company credentials",
		})
		return
	}

	body, err := json.Marshal(coreLoginRequest{
		Email:          req.Email,
		Password:       req.Password,
		TurnstileToken: req.TurnstileToken,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to build upstream login request",
		})
		return
	}

	resp, err := s.doUpstreamRequest(c.Request.Context(), http.MethodPost, "/auth/login", "", c.Request.Header, body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    http.StatusBadGateway,
			"message": "Failed to reach upstream core",
		})
		return
	}
	if resp.StatusCode == http.StatusOK && isJSONResponse(resp.Header) && len(resp.Body) > 0 {
		if transformed, transformErr := injectEnterpriseIntoAuthResponse(resp.Body, profile); transformErr == nil {
			resp.Body = transformed
		}
	}
	writeUpstreamResponse(c, resp)
}

func (s *Server) handleEnterpriseLogin2FA(c *gin.Context) {
	var req enterpriseLogin2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Invalid request payload",
		})
		return
	}

	body, err := json.Marshal(coreLogin2FARequest{
		TempToken: req.TempToken,
		TotpCode:  req.TotpCode,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to build upstream 2FA request",
		})
		return
	}

	resp, err := s.doUpstreamRequest(c.Request.Context(), http.MethodPost, "/auth/login/2fa", "", c.Request.Header, body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    http.StatusBadGateway,
			"message": "Failed to reach upstream core",
		})
		return
	}

	if resp.StatusCode == http.StatusOK && isJSONResponse(resp.Header) && len(resp.Body) > 0 {
		userID, userErr := extractUserIDFromEnvelope(resp.Body)
		if userErr == nil {
			profile, profileErr := s.enterpriseStore.GetByUserID(c.Request.Context(), userID)
			if profileErr != nil {
				c.JSON(http.StatusBadGateway, gin.H{
					"code":    http.StatusBadGateway,
					"message": "Failed to resolve enterprise profile",
				})
				return
			}
			if profile == nil || normalizeCompanyName(profile.Name) != normalizeCompanyName(req.CompanyName) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "Invalid company credentials",
				})
				return
			}
			if transformed, transformErr := injectEnterpriseIntoAuthResponse(resp.Body, profile); transformErr == nil {
				resp.Body = transformed
			}
		}
	}

	writeUpstreamResponse(c, resp)
}

func (s *Server) handleEnterpriseMe(c *gin.Context) {
	user, resp, err := s.currentUserFromCore(c.Request.Context(), c.Request.Header)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    http.StatusBadGateway,
			"message": "Failed to verify current user",
		})
		return
	}
	if user == nil || resp == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "Unauthorized",
		})
		return
	}

	profile, profileErr := s.enterpriseStore.GetByUserID(c.Request.Context(), user.ID)
	if profileErr != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    http.StatusBadGateway,
			"message": "Failed to resolve enterprise profile",
		})
		return
	}
	if profile == nil {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    http.StatusForbidden,
			"message": "Enterprise access required",
		})
		return
	}

	if transformed, transformErr := injectEnterpriseIntoCurrentUser(resp.Body, profile); transformErr == nil {
		resp.Body = transformed
	}
	writeUpstreamResponse(c, resp)
}

func (s *Server) handleEnterprisePublicSettings(c *gin.Context) {
	if strings.TrimSpace(c.GetHeader("Authorization")) == "" {
		s.proxy(c, "/settings/public", nil)
		return
	}

	user, _, err := s.currentUserFromCore(c.Request.Context(), c.Request.Header)
	if err != nil || user == nil {
		s.proxy(c, "/settings/public", nil)
		return
	}

	profile, profileErr := s.enterpriseStore.GetByUserID(c.Request.Context(), user.ID)
	if profileErr != nil || profile == nil {
		s.proxy(c, "/settings/public", nil)
		return
	}

	resp, err := s.doUpstreamRequest(c.Request.Context(), http.MethodGet, "/settings/public", c.Request.URL.RawQuery, c.Request.Header, nil)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    http.StatusBadGateway,
			"message": "Failed to reach upstream core",
		})
		return
	}
	if resp.StatusCode == http.StatusOK && isJSONResponse(resp.Header) && len(resp.Body) > 0 {
		if transformed, transformErr := injectEnterpriseIntoPublicSettings(resp.Body, profile); transformErr == nil {
			resp.Body = transformed
		}
	}
	writeUpstreamResponse(c, resp)
}
