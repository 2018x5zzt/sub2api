package enterprisebff

import (
	"encoding/json"
)

type responseTransformer func([]byte) ([]byte, error)

func transformKeysEnvelope(body []byte) ([]byte, error) {
	return transformEnvelope(body, normalizeKeyContainer)
}

func transformUsageEnvelope(body []byte) ([]byte, error) {
	return transformEnvelope(body, normalizeUsageContainer)
}

func transformEnvelope(body []byte, mutate func(any) any) ([]byte, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return body, nil
	}
	data, ok := payload["data"]
	if !ok {
		return body, nil
	}
	payload["data"] = mutate(data)
	normalized, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return normalized, nil
}

func normalizeKeyContainer(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		if items, ok := typed["items"].([]any); ok {
			for _, item := range items {
				normalizeKeyItem(item)
			}
			return typed
		}
		normalizeKeyItem(typed)
		return typed
	case []any:
		for _, item := range typed {
			normalizeKeyItem(item)
		}
	}
	return value
}

func normalizeKeyItem(value any) {
	item, ok := value.(map[string]any)
	if !ok {
		return
	}
	if quotaUsed, ok := item["quota_used"]; ok {
		if _, exists := item["used_quota"]; !exists {
			item["used_quota"] = quotaUsed
		}
	}
	if _, exists := item["group_id"]; !exists {
		if group, ok := item["group"].(map[string]any); ok {
			if groupID, ok := group["id"]; ok {
				item["group_id"] = groupID
			}
		}
	}
}

func normalizeUsageContainer(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		if items, ok := typed["items"].([]any); ok {
			for _, item := range items {
				normalizeUsageItem(item)
			}
			return typed
		}
		normalizeUsageItem(typed)
		return typed
	case []any:
		for _, item := range typed {
			normalizeUsageItem(item)
		}
	}
	return value
}

func normalizeUsageItem(value any) {
	item, ok := value.(map[string]any)
	if !ok {
		return
	}
	if totalCost, ok := item["total_cost"]; ok {
		if _, exists := item["billable_cost"]; !exists {
			item["billable_cost"] = totalCost
		}
		return
	}
	if billableCost, ok := item["billable_cost"]; ok {
		item["total_cost"] = billableCost
	}
}
