package enterprisebff

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

var errUnauthorizedEnterpriseGroup = errors.New("无权使用该号池")

type requestedGroupBinding struct {
	Present bool
	GroupID *int64
}

func parseRequestedGroupBinding(body []byte) (requestedGroupBinding, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return requestedGroupBinding{}, err
	}

	value, ok := raw["group_id"]
	if !ok {
		return requestedGroupBinding{Present: false}, nil
	}
	if string(value) == "null" {
		return requestedGroupBinding{Present: true, GroupID: nil}, nil
	}

	var id int64
	if err := json.Unmarshal(value, &id); err != nil {
		return requestedGroupBinding{}, err
	}

	return requestedGroupBinding{
		Present: true,
		GroupID: &id,
	}, nil
}

func (s *Server) authorizeRequestedGroup(
	ctx context.Context,
	actor *currentUser,
	targetUserID int64,
	binding requestedGroupBinding,
) error {
	if s == nil || s.enterpriseStore == nil || !binding.Present || binding.GroupID == nil {
		return nil
	}

	sameEnterprise, err := s.enterpriseStore.SameEnterprise(ctx, actor.ID, targetUserID)
	if err != nil {
		return err
	}
	if !sameEnterprise {
		return errUnauthorizedEnterpriseGroup
	}

	visibleGroups, err := s.enterpriseStore.ListVisibleGroups(ctx, targetUserID)
	if err != nil {
		return err
	}
	for _, group := range visibleGroups {
		if group.ID == *binding.GroupID {
			return nil
		}
	}

	return errUnauthorizedEnterpriseGroup
}

func (s *Server) proxyValidatedKeyMutation(
	c *gin.Context,
	actor *currentUser,
	targetUserID int64,
	upstreamPath string,
	transformer responseTransformer,
) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Invalid request body",
		})
		return
	}

	binding, err := parseRequestedGroupBinding(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}
	if err := s.authorizeRequestedGroup(c.Request.Context(), actor, targetUserID, binding); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    http.StatusForbidden,
			"message": err.Error(),
		})
		return
	}

	resp, err := s.doUpstreamRequest(c.Request.Context(), c.Request.Method, upstreamPath, c.Request.URL.RawQuery, c.Request.Header, body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    http.StatusBadGateway,
			"message": "Failed to reach upstream core",
		})
		return
	}
	if transformer != nil && isJSONResponse(resp.Header) && len(resp.Body) > 0 {
		if transformed, transformErr := transformer(resp.Body); transformErr == nil {
			resp.Body = transformed
		}
	}
	writeUpstreamResponse(c, resp)
}
