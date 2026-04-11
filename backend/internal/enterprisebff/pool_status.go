package enterprisebff

import (
	"encoding/json"
	"math"
	"net/http"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type enterpriseTrendPoint struct {
	BucketTime    string `json:"bucket_time"`
	HealthPercent int    `json:"health_percent"`
	HealthState   string `json:"health_state"`
}

type enterprisePoolStatusGroup struct {
	GroupID       int64                  `json:"group_id"`
	GroupName     string                 `json:"group_name"`
	HealthPercent int                    `json:"health_percent"`
	HealthState   string                 `json:"health_state"`
	Trend         []enterpriseTrendPoint `json:"trend,omitempty"`
	UpdatedAt     string                 `json:"updated_at"`
}

type enterprisePoolStatusResponse struct {
	VisibleGroupCount    int                         `json:"visible_group_count"`
	OverallHealthPercent *int                        `json:"overall_health_percent"`
	UpdatedAt            string                      `json:"updated_at"`
	Groups               []enterprisePoolStatusGroup `json:"groups"`
}

func (s *Server) handleEnterpriseVisibleGroups(c *gin.Context) {
	user, ok := s.requireCurrentUser(c)
	if !ok {
		return
	}

	visibleGroups, err := s.enterpriseStore.ListVisibleGroups(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    http.StatusBadGateway,
			"message": "Failed to resolve enterprise groups",
		})
		return
	}

	response.Success(c, visibleGroups)
}

func (s *Server) handleEnterprisePoolStatus(c *gin.Context) {
	user, ok := s.requireCurrentUser(c)
	if !ok {
		return
	}

	visibleGroups, err := s.enterpriseStore.ListVisibleGroups(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    http.StatusBadGateway,
			"message": "Failed to resolve enterprise groups",
		})
		return
	}
	if len(visibleGroups) == 0 {
		response.Success(c, enterprisePoolStatusResponse{
			VisibleGroupCount: 0,
			Groups:            []enterprisePoolStatusGroup{},
		})
		return
	}

	coreResp, err := s.doUpstreamRequest(c.Request.Context(), http.MethodGet, "/groups/pool-status", c.Request.URL.RawQuery, c.Request.Header, nil)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    http.StatusBadGateway,
			"message": "Failed to reach upstream core",
		})
		return
	}
	if coreResp.StatusCode != http.StatusOK {
		writeUpstreamResponse(c, coreResp)
		return
	}

	currentGroups, checkedAt, err := decodeUserPoolStatus(coreResp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    http.StatusBadGateway,
			"message": "Failed to decode upstream pool status",
		})
		return
	}

	visibleIDs := collectVisibleGroupIDs(visibleGroups)
	history := map[int64][]service.GroupHealthSnapshot{}
	if s.healthRepo != nil {
		history, err = s.healthRepo.ListRecentByGroupIDs(c.Request.Context(), visibleIDs, time.Now().UTC().Add(-24*time.Hour))
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"code":    http.StatusBadGateway,
				"message": "Failed to load pool trend history",
			})
			return
		}
	}

	response.Success(c, buildEnterprisePoolStatusResponse(visibleGroups, currentGroups, history, checkedAt))
}

func collectVisibleGroupIDs(groups []EnterpriseVisibleGroup) []int64 {
	out := make([]int64, 0, len(groups))
	for _, group := range groups {
		out = append(out, group.ID)
	}
	return out
}

func decodeUserPoolStatus(body []byte) (map[int64]handler.UserVisibleGroupPoolStatus, time.Time, error) {
	var envelope struct {
		Data struct {
			CheckedAt string                               `json:"checked_at"`
			Groups    []handler.UserVisibleGroupPoolStatus `json:"groups"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, time.Time{}, err
	}

	checkedAt, err := time.Parse(time.RFC3339, envelope.Data.CheckedAt)
	if err != nil {
		return nil, time.Time{}, err
	}

	out := make(map[int64]handler.UserVisibleGroupPoolStatus, len(envelope.Data.Groups))
	for _, group := range envelope.Data.Groups {
		out[group.GroupID] = group
	}
	return out, checkedAt, nil
}

func buildEnterprisePoolStatusResponse(
	visibleGroups []EnterpriseVisibleGroup,
	currentGroups map[int64]handler.UserVisibleGroupPoolStatus,
	history map[int64][]service.GroupHealthSnapshot,
	checkedAt time.Time,
) enterprisePoolStatusResponse {
	out := make([]enterprisePoolStatusGroup, 0, len(visibleGroups))
	totalPercent := 0

	for _, visible := range visibleGroups {
		current, ok := currentGroups[visible.ID]
		if !ok {
			continue
		}

		healthPercent := int(math.Round(current.AvailabilityRatio * 100))
		totalPercent += healthPercent

		trend := make([]enterpriseTrendPoint, 0, len(history[visible.ID]))
		for _, point := range history[visible.ID] {
			trend = append(trend, enterpriseTrendPoint{
				BucketTime:    point.BucketTime.UTC().Format(time.RFC3339),
				HealthPercent: point.HealthPercent,
				HealthState:   point.HealthState,
			})
		}

		out = append(out, enterprisePoolStatusGroup{
			GroupID:       visible.ID,
			GroupName:     visible.Name,
			HealthPercent: healthPercent,
			HealthState:   current.Status,
			Trend:         trend,
			UpdatedAt:     checkedAt.UTC().Format(time.RFC3339),
		})
	}

	var overall *int
	if len(out) > 0 {
		value := totalPercent / len(out)
		overall = &value
	}

	return enterprisePoolStatusResponse{
		VisibleGroupCount:    len(visibleGroups),
		OverallHealthPercent: overall,
		UpdatedAt:            checkedAt.UTC().Format(time.RFC3339),
		Groups:               out,
	}
}
