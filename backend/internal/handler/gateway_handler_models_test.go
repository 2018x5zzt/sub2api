package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type gatewayModelsAccountRepoStub struct {
	service.AccountRepository

	byGroup map[int64][]service.Account
	all     []service.Account
}

func (s *gatewayModelsAccountRepoStub) ListSchedulableByGroupID(_ context.Context, groupID int64) ([]service.Account, error) {
	accounts := s.byGroup[groupID]
	out := make([]service.Account, len(accounts))
	copy(out, accounts)
	return out, nil
}

func (s *gatewayModelsAccountRepoStub) ListSchedulable(_ context.Context) ([]service.Account, error) {
	out := make([]service.Account, len(s.all))
	copy(out, s.all)
	return out, nil
}

func newGatewayModelsTestContext(t *testing.T, group *service.Group) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)

	var groupID *int64
	if group != nil {
		id := group.ID
		groupID = &id
	}
	c.Set(string(servermiddleware.ContextKeyAPIKey), &service.APIKey{
		ID:      1,
		GroupID: groupID,
		Group:   group,
	})
	return c, recorder
}

func decodeOpenAIModelIDs(t *testing.T, body []byte) []string {
	t.Helper()

	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))

	ids := make([]string, 0, len(payload.Data))
	for _, model := range payload.Data {
		ids = append(ids, model.ID)
	}
	return ids
}

func TestGatewayHandlerModels_HidesImageModelsForNonGPTImageGroup_DefaultCatalog(t *testing.T) {
	groupID := int64(29)
	group := &service.Group{ID: groupID, Name: "pro号池", Platform: service.PlatformOpenAI}
	gatewayService := service.NewGatewayService(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				groupID: {
					{Platform: service.PlatformOpenAI},
				},
			},
		},
		nil, nil, nil, nil, nil, nil, nil, &config.Config{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	handler := &GatewayHandler{gatewayService: gatewayService}
	c, recorder := newGatewayModelsTestContext(t, group)

	handler.Models(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	modelIDs := decodeOpenAIModelIDs(t, recorder.Body.Bytes())
	require.Contains(t, modelIDs, "gpt-5")
	require.NotContains(t, modelIDs, "gpt-image-2")
}

func TestGatewayHandlerModels_HidesImageModelsForNonGPTImageGroup_MappedCatalog(t *testing.T) {
	groupID := int64(29)
	group := &service.Group{ID: groupID, Name: "pro号池", Platform: service.PlatformOpenAI}
	gatewayService := service.NewGatewayService(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				groupID: {
					{
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gpt-5":       "gpt-5",
								"gpt-image-2": "gpt-image-2",
							},
						},
					},
				},
			},
		},
		nil, nil, nil, nil, nil, nil, nil, &config.Config{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	handler := &GatewayHandler{gatewayService: gatewayService}
	c, recorder := newGatewayModelsTestContext(t, group)

	handler.Models(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	modelIDs := decodeOpenAIModelIDs(t, recorder.Body.Bytes())
	require.Contains(t, modelIDs, "gpt-5")
	require.NotContains(t, modelIDs, "gpt-image-2")
}

func TestGatewayHandlerModels_OpenAIPassthroughIgnoresStaleMappings(t *testing.T) {
	groupID := int64(29)
	group := &service.Group{ID: groupID, Name: "pro号池", Platform: service.PlatformOpenAI}
	gatewayService := service.NewGatewayService(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				groupID: {
					{
						Platform: service.PlatformOpenAI,
						Type:     service.AccountTypeOAuth,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"legacy-alias": "gpt-5.4",
							},
						},
						Extra: map[string]any{
							"openai_passthrough": true,
						},
					},
				},
			},
		},
		nil, nil, nil, nil, nil, nil, nil, &config.Config{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	handler := &GatewayHandler{gatewayService: gatewayService}
	c, recorder := newGatewayModelsTestContext(t, group)

	handler.Models(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	modelIDs := decodeOpenAIModelIDs(t, recorder.Body.Bytes())
	require.Contains(t, modelIDs, "gpt-5.5")
	require.NotContains(t, modelIDs, "legacy-alias")
	require.NotContains(t, modelIDs, "gpt-image-2")
}

func TestGatewayHandlerModels_KeepsImageModelsForGPTImageGroup(t *testing.T) {
	groupID := int64(30)
	group := &service.Group{ID: groupID, Name: "gpt-image", Platform: service.PlatformOpenAI}
	gatewayService := service.NewGatewayService(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				groupID: {
					{Platform: service.PlatformOpenAI},
				},
			},
		},
		nil, nil, nil, nil, nil, nil, nil, &config.Config{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	handler := &GatewayHandler{gatewayService: gatewayService}
	c, recorder := newGatewayModelsTestContext(t, group)

	handler.Models(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	modelIDs := decodeOpenAIModelIDs(t, recorder.Body.Bytes())
	require.Contains(t, modelIDs, "gpt-image-2")
	require.NotContains(t, modelIDs, "gpt-5")
	for _, modelID := range modelIDs {
		require.True(t, service.IsOpenAIImageGenerationModel(modelID), "unexpected non-image model %s", modelID)
	}
}
