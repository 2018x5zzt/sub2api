package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type EnterpriseVisibleGroupSetting struct {
	EnterpriseName  string  `json:"enterprise_name"`
	VisibleGroupIDs []int64 `json:"visible_group_ids"`
}

func normalizeEnterpriseVisibleGroupEnterpriseName(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func ParseEnterpriseVisibleGroupIDsByEnterprise(raw string) (map[string][]int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parsed := make(map[string][]int64)
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}

	return normalizeEnterpriseVisibleGroupIDMap(parsed)
}

func BuildEnterpriseVisibleGroupRules(raw map[string][]int64) []EnterpriseVisibleGroupSetting {
	normalized, err := normalizeEnterpriseVisibleGroupIDMap(raw)
	if err != nil || len(normalized) == 0 {
		return []EnterpriseVisibleGroupSetting{}
	}

	enterpriseNames := make([]string, 0, len(normalized))
	for enterpriseName := range normalized {
		enterpriseNames = append(enterpriseNames, enterpriseName)
	}
	sort.Strings(enterpriseNames)

	out := make([]EnterpriseVisibleGroupSetting, 0, len(enterpriseNames))
	for _, enterpriseName := range enterpriseNames {
		groupIDs := append([]int64(nil), normalized[enterpriseName]...)
		out = append(out, EnterpriseVisibleGroupSetting{
			EnterpriseName:  enterpriseName,
			VisibleGroupIDs: groupIDs,
		})
	}

	return out
}

func normalizeEnterpriseVisibleGroupIDMap(raw map[string][]int64) (map[string][]int64, error) {
	if len(raw) == 0 {
		return map[string][]int64{}, nil
	}

	out := make(map[string][]int64, len(raw))
	for enterpriseName, ids := range raw {
		normalizedName := normalizeEnterpriseVisibleGroupEnterpriseName(enterpriseName)
		if normalizedName == "" {
			continue
		}

		deduped, err := normalizeEnterpriseVisibleGroupIDs(ids)
		if err != nil {
			return nil, fmt.Errorf("enterprise %q: %w", enterpriseName, err)
		}
		if len(deduped) == 0 {
			continue
		}

		out[normalizedName] = mergeEnterpriseVisibleGroupIDLists(out[normalizedName], deduped)
	}

	return out, nil
}

func normalizeEnterpriseVisibleGroups(
	ctx context.Context,
	rules []EnterpriseVisibleGroupSetting,
	groupReader DefaultSubscriptionGroupReader,
) (map[string][]int64, error) {
	out := make(map[string][]int64)
	if len(rules) == 0 {
		return out, nil
	}

	for _, rule := range rules {
		enterpriseName := normalizeEnterpriseVisibleGroupEnterpriseName(rule.EnterpriseName)
		if enterpriseName == "" {
			continue
		}

		seen := make(map[int64]struct{}, len(rule.VisibleGroupIDs))
		filtered := make([]int64, 0, len(rule.VisibleGroupIDs))
		for _, groupID := range rule.VisibleGroupIDs {
			if groupID <= 0 {
				continue
			}
			if _, ok := seen[groupID]; ok {
				continue
			}
			seen[groupID] = struct{}{}

			if groupReader == nil {
				filtered = append(filtered, groupID)
				continue
			}

			group, err := groupReader.GetByID(ctx, groupID)
			if err != nil {
				if errors.Is(err, ErrGroupNotFound) {
					continue
				}
				return nil, fmt.Errorf("get enterprise visible group %d: %w", groupID, err)
			}
			if !group.IsActive() {
				continue
			}

			filtered = append(filtered, groupID)
		}

		if len(filtered) == 0 {
			continue
		}

		out[enterpriseName] = mergeEnterpriseVisibleGroupIDLists(out[enterpriseName], filtered)
	}

	return out, nil
}

func normalizeEnterpriseVisibleGroupIDs(ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	seen := make(map[int64]struct{}, len(ids))
	deduped := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return nil, fmt.Errorf("invalid group id %d", id)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		deduped = append(deduped, id)
	}

	sort.Slice(deduped, func(i, j int) bool {
		return deduped[i] < deduped[j]
	})
	return deduped, nil
}

func mergeEnterpriseVisibleGroupIDLists(existing []int64, incoming []int64) []int64 {
	if len(existing) == 0 {
		return append([]int64(nil), incoming...)
	}
	if len(incoming) == 0 {
		return append([]int64(nil), existing...)
	}

	merged := make([]int64, 0, len(existing)+len(incoming))
	seen := make(map[int64]struct{}, len(existing)+len(incoming))
	for _, id := range existing {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		merged = append(merged, id)
	}
	for _, id := range incoming {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		merged = append(merged, id)
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i] < merged[j]
	})
	return merged
}
