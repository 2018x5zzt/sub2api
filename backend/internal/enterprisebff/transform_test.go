package enterprisebff

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransformKeysEnvelopeAddsUsedQuotaAlias(t *testing.T) {
	body := []byte(`{"code":0,"message":"success","data":{"items":[{"id":1,"quota_used":12.5,"group":{"id":9}}]}}`)

	got, err := transformKeysEnvelope(body)
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(got, &payload))

	data := payload["data"].(map[string]any)
	item := data["items"].([]any)[0].(map[string]any)
	require.Equal(t, 12.5, item["used_quota"])
	require.Equal(t, float64(9), item["group_id"])
}

func TestTransformUsageEnvelopeAddsBillableCostAlias(t *testing.T) {
	body := []byte(`{"code":0,"message":"success","data":{"items":[{"id":1,"total_cost":0.12,"actual_cost":0.08,"rate_multiplier":2}]}}`)

	got, err := transformUsageEnvelope(body)
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(got, &payload))

	data := payload["data"].(map[string]any)
	item := data["items"].([]any)[0].(map[string]any)
	require.Equal(t, 0.12, item["billable_cost"])
	require.Equal(t, 0.12, item["total_cost"])
}
