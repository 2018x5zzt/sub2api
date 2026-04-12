//go:build integration

package repository

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var mutatingStatementRE = regexp.MustCompile(`\b(INSERT|UPDATE|DELETE|CREATE|ALTER|DROP|TRUNCATE|MERGE|COPY|GRANT|REVOKE|LOCK)\b`)

type sqlQueryResult struct {
	Columns []string
	Rows    []map[string]*string
}

func TestInviteRolloutVerificationSQL_OverviewMetricsAndReadOnlyStatements(t *testing.T) {
	tx := testTx(t)
	resetInviteVerificationTables(t, tx)
	seedInviteVerificationOverviewFixture(t, tx)

	statements := executableVerificationStatements(t, readInviteRolloutVerificationSQL(t))
	require.NotEmpty(t, statements)

	results := executeVerificationStatements(t, tx, statements[:1])
	require.Len(t, results[0].Rows, 8)
	metrics := metricMap(t, results[0])

	require.Equal(t, "1", metrics["bound_users_total"])
	require.Equal(t, "1", metrics["register_bind_event_rows_total"])
	require.Equal(t, "1", metrics["register_bind_distinct_invitees_total"])
	require.Equal(t, "1", metrics["base_invite_reward_rows_total"])
	require.Equal(t, "5.00000000", metrics["base_invite_reward_amount_total"])
	require.Equal(t, "1", metrics["admin_rebind_event_rows_total"])
	require.Equal(t, "1", metrics["manual_invite_grant_rows_total"])
	require.Equal(t, "1", metrics["recompute_delta_rows_total"])
}

func TestInviteRolloutVerificationSQL_BindingChecksExposeDriftSamples(t *testing.T) {
	tx := testTx(t)
	resetInviteVerificationTables(t, tx)
	seedInviteVerificationBindingFixture(t, tx)

	statements := executableVerificationStatements(t, readInviteRolloutVerificationSQL(t))
	require.GreaterOrEqual(t, len(statements), 5, "expected statements 1-5 to exist")

	results := executeVerificationStatements(t, tx, statements[:5])

	bindingMetrics := metricMap(t, results[1])
	require.Equal(t, "1", bindingMetrics["bound_users_missing_register_bind_total"])
	require.Equal(t, "1", bindingMetrics["register_bind_without_bound_user_total"])
	require.Equal(t, "1", bindingMetrics["register_bind_duplicate_invitee_total"])
	require.Equal(t, "1", bindingMetrics["register_bind_inviter_mismatch_total"])
	require.Equal(t, "1", bindingMetrics["register_bind_effective_at_mismatch_total"])

	missingRegisterBindSamples := results[2]
	requireColumns(t, missingRegisterBindSamples, "invitee_user_id", "current_inviter_user_id", "expected_effective_at")
	require.Equal(t, []string{"1105"}, columnValues(t, missingRegisterBindSamples, "invitee_user_id"))

	duplicateRegisterBindSamples := results[3]
	requireColumns(t, duplicateRegisterBindSamples, "invitee_user_id", "register_bind_count", "first_effective_at", "last_effective_at")
	require.Equal(t, []string{"1107"}, columnValues(t, duplicateRegisterBindSamples, "invitee_user_id"))

	mismatchSamples := results[4]
	requireColumns(t, mismatchSamples, "invitee_user_id", "current_inviter_user_id", "event_inviter_user_id", "expected_effective_at", "event_effective_at")
	require.Equal(t, []string{"1108", "1109"}, columnValues(t, mismatchSamples, "invitee_user_id"))

	mismatchByInvitee := make(map[string]map[string]*string, len(mismatchSamples.Rows))
	for _, row := range mismatchSamples.Rows {
		invitee := row["invitee_user_id"]
		require.NotNil(t, invitee, "invitee_user_id must not be null in mismatch samples")
		mismatchByInvitee[*invitee] = row
	}

	row1108, ok := mismatchByInvitee["1108"]
	require.True(t, ok, "expected mismatch sample for invitee 1108")
	require.NotNil(t, row1108["current_inviter_user_id"])
	require.NotNil(t, row1108["event_inviter_user_id"])
	require.Equal(t, "1101", *row1108["current_inviter_user_id"])
	require.Equal(t, "1102", *row1108["event_inviter_user_id"])

	row1109, ok := mismatchByInvitee["1109"]
	require.True(t, ok, "expected mismatch sample for invitee 1109")
	require.NotNil(t, row1109["current_inviter_user_id"])
	require.NotNil(t, row1109["event_inviter_user_id"])
	require.NotNil(t, row1109["expected_effective_at"])
	require.NotNil(t, row1109["event_effective_at"])
	require.Equal(t, "1101", *row1109["current_inviter_user_id"])
	require.Equal(t, "1101", *row1109["event_inviter_user_id"])
	require.NotEqual(t, *row1109["expected_effective_at"], *row1109["event_effective_at"])
}

func TestInviteRolloutVerificationSQL_RewardChecksSeparateAdminCorrections(t *testing.T) {
	tx := testTx(t)
	resetInviteVerificationTables(t, tx)
	seedInviteVerificationRewardFixture(t, tx)

	statements := executableVerificationStatements(t, readInviteRolloutVerificationSQL(t))
	require.GreaterOrEqual(t, len(statements), 9, "expected statements 1-9 to exist")

	results := executeVerificationStatements(t, tx, statements[:9])

	rewardSummary := results[5]
	requireColumns(t, rewardSummary,
		"reward_type",
		"status",
		"rows_total",
		"reward_amount_total",
		"distinct_invitees_total",
		"distinct_reward_targets_total",
		"rows_with_admin_action_total",
		"rows_with_null_trigger_code_total",
	)

	summaryByType := make(map[string]map[string]*string, len(rewardSummary.Rows))
	for _, row := range rewardSummary.Rows {
		rewardType := row["reward_type"]
		status := row["status"]
		require.NotNil(t, rewardType, "reward_type must not be null")
		require.NotNil(t, status, "status must not be null")
		summaryByType[*rewardType+"|"+*status] = row
	}

	baseApplied, ok := summaryByType["base_invite_reward|applied"]
	require.True(t, ok, "expected summary row for base_invite_reward applied")
	require.Equal(t, "3", requireRowValue(t, baseApplied, "rows_total"))
	require.Equal(t, "15.00000000", requireRowValue(t, baseApplied, "reward_amount_total"))
	require.Equal(t, "1", requireRowValue(t, baseApplied, "rows_with_null_trigger_code_total"))
	require.Equal(t, "1", requireRowValue(t, baseApplied, "rows_with_admin_action_total"))

	manualApplied, ok := summaryByType["manual_invite_grant|applied"]
	require.True(t, ok, "expected summary row for manual_invite_grant applied")
	require.Equal(t, "2", requireRowValue(t, manualApplied, "rows_total"))
	require.Equal(t, "27.50000000", requireRowValue(t, manualApplied, "reward_amount_total"))
	require.Equal(t, "1", requireRowValue(t, manualApplied, "rows_with_admin_action_total"))

	recomputeApplied, ok := summaryByType["recompute_delta|applied"]
	require.True(t, ok, "expected summary row for recompute_delta applied")
	require.Equal(t, "2", requireRowValue(t, recomputeApplied, "rows_total"))
	require.Equal(t, "3.00000000", requireRowValue(t, recomputeApplied, "reward_amount_total"))
	require.Equal(t, "1", requireRowValue(t, recomputeApplied, "rows_with_admin_action_total"))

	rewardAnomalyMetrics := metricMap(t, results[6])
	require.Equal(t, "1", rewardAnomalyMetrics["base_reward_with_admin_action_total"])
	require.Equal(t, "1", rewardAnomalyMetrics["manual_grant_without_admin_action_total"])
	require.Equal(t, "1", rewardAnomalyMetrics["recompute_delta_without_admin_action_total"])
	require.Equal(t, "1", rewardAnomalyMetrics["base_reward_without_trigger_code_total"])

	rewardAnomalySamples := results[7]
	requireColumns(t, rewardAnomalySamples,
		"id",
		"reward_type",
		"reward_role",
		"reward_amount",
		"admin_action_id",
		"trigger_redeem_code_id",
		"created_at",
	)
	require.Equal(t, []string{"5202", "5205", "5207"}, columnValues(t, rewardAnomalySamples, "id"))
	anomalyByID := make(map[string]map[string]*string, len(rewardAnomalySamples.Rows))
	for _, row := range rewardAnomalySamples.Rows {
		id := row["id"]
		require.NotNil(t, id, "id must not be null in reward anomaly samples")
		anomalyByID[*id] = row
	}

	row5202, ok := anomalyByID["5202"]
	require.True(t, ok, "expected anomaly sample for id 5202")
	require.NotNil(t, row5202["admin_action_id"])
	require.NotNil(t, row5202["trigger_redeem_code_id"])

	row5205, ok := anomalyByID["5205"]
	require.True(t, ok, "expected anomaly sample for id 5205")
	require.Nil(t, row5205["admin_action_id"])

	row5207, ok := anomalyByID["5207"]
	require.True(t, ok, "expected anomaly sample for id 5207")
	require.Nil(t, row5207["admin_action_id"])

	baseRewardObservationSamples := results[8]
	requireColumns(t, baseRewardObservationSamples,
		"id",
		"inviter_user_id",
		"invitee_user_id",
		"reward_target_user_id",
		"reward_role",
		"reward_amount",
		"trigger_redeem_code_id",
		"created_at",
	)
	require.Equal(t, []string{"5201", "5202", "5203"}, columnValues(t, baseRewardObservationSamples, "id"))
}

func TestInviteRolloutVerificationSQL_FullContractAndOperatorNotes(t *testing.T) {
	tx := testTx(t)
	resetInviteVerificationTables(t, tx)

	content := readInviteRolloutVerificationSQL(t)

	require.Contains(t, content, "Pre-deploy and post-deploy")
	require.Contains(t, content, "not be treated as a migration")
	require.Contains(t, content, "rebind_inviter")
	require.Contains(t, content, "manual_reward_grant")
	require.Contains(t, content, "recompute_rewards")

	statements := executableVerificationStatements(t, content)
	require.Len(t, statements, 9, "executable verification statement count is part of the rollout contract")

	results := executeVerificationStatements(t, tx, statements)
	require.Len(t, results, 9, "expected one query result per executable statement")

	requireColumns(t, results[0], "metric_name", "metric_value")
	requireColumns(t, results[1], "metric_name", "metric_value")
	requireColumns(t, results[2], "invitee_user_id", "current_inviter_user_id", "expected_effective_at")
	requireColumns(t, results[3], "invitee_user_id", "register_bind_count", "first_effective_at", "last_effective_at")
	requireColumns(t, results[4], "invitee_user_id", "current_inviter_user_id", "event_inviter_user_id", "expected_effective_at", "event_effective_at")
	requireColumns(t, results[5],
		"reward_type",
		"status",
		"rows_total",
		"reward_amount_total",
		"distinct_invitees_total",
		"distinct_reward_targets_total",
		"rows_with_admin_action_total",
		"rows_with_null_trigger_code_total",
	)
	requireColumns(t, results[6], "metric_name", "metric_value")
	requireColumns(t, results[7], "id", "reward_type", "reward_role", "reward_amount", "admin_action_id", "trigger_redeem_code_id", "created_at")
	requireColumns(t, results[8], "id", "inviter_user_id", "invitee_user_id", "reward_target_user_id", "reward_role", "reward_amount", "trigger_redeem_code_id", "created_at")
}

func readInviteRolloutVerificationSQL(t *testing.T) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("..", "..", "resources", "sql", "ops", "invite_admin_rollout_verification.sql"))
	require.NoError(t, err)
	return string(content)
}

func executableVerificationStatements(t *testing.T, content string) []string {
	t.Helper()

	raw := splitSQLStatements(content)
	out := make([]string, 0, len(raw))
	for _, stmt := range raw {
		trimmed := stripSQLLineComment(strings.TrimSpace(stmt))
		if trimmed == "" {
			continue
		}

		upper := strings.ToUpper(trimmed)
		sanitized := strings.ToUpper(stripSingleQuotedLiterals(trimmed))
		require.True(t,
			strings.HasPrefix(upper, "SELECT") || strings.HasPrefix(upper, "WITH"),
			"verification SQL must stay read-only, got: %s",
			trimmed,
		)
		require.False(t, mutatingStatementRE.MatchString(sanitized),
			"verification SQL must stay read-only; detected mutation keyword in: %s",
			trimmed,
		)
		out = append(out, stmt)
	}
	return out
}

func stripSingleQuotedLiterals(s string) string {
	var builder strings.Builder
	inside := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inside {
			if ch == '\'' {
				if i+1 < len(s) && s[i+1] == '\'' {
					builder.WriteByte(' ') // treat escaped quote as space placeholder
					i++
					continue
				}
				inside = false
				builder.WriteByte(' ')
				continue
			}
			builder.WriteByte(' ')
			continue
		}
		if ch == '\'' {
			inside = true
			builder.WriteByte(' ')
			continue
		}
		builder.WriteByte(ch)
	}
	return builder.String()
}

func executeVerificationStatements(t *testing.T, tx *sql.Tx, statements []string) []sqlQueryResult {
	t.Helper()

	results := make([]sqlQueryResult, 0, len(statements))
	for _, stmt := range statements {
		result := func() sqlQueryResult {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			rows, err := tx.QueryContext(ctx, stmt)
			require.NoError(t, err)
			cols, err := rows.Columns()
			require.NoError(t, err)
			result := sqlQueryResult{Columns: cols}
			raw := make([]sql.NullString, len(cols))
			args := make([]any, len(cols))
			for i := range raw {
				args[i] = &raw[i]
			}
			for rows.Next() {
				require.NoError(t, rows.Scan(args...))
				row := make(map[string]*string, len(cols))
				for i, col := range cols {
					if raw[i].Valid {
						value := raw[i].String
						row[col] = &value
					} else {
						row[col] = nil
					}
				}
				result.Rows = append(result.Rows, row)
			}
			require.NoError(t, rows.Err())
			require.NoError(t, rows.Close())
			return result
		}()
		results = append(results, result)
	}
	return results
}

func resetInviteVerificationTables(t *testing.T, tx *sql.Tx) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, `DELETE FROM invite_reward_records`)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `DELETE FROM invite_relationship_events`)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `DELETE FROM invite_admin_actions`)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `DELETE FROM redeem_codes`)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `DELETE FROM users`)
	require.NoError(t, err)
}

func metricMap(t *testing.T, result sqlQueryResult) map[string]string {
	t.Helper()
	require.Equal(t, []string{"metric_name", "metric_value"}, result.Columns)

	out := make(map[string]string, len(result.Rows))
	for _, row := range result.Rows {
		name := row["metric_name"]
		value := row["metric_value"]
		require.NotNil(t, name, "metric_name must not be null")
		require.NotNil(t, value, "metric_value must not be null")
		out[*name] = *value
	}
	return out
}

func requireColumns(t *testing.T, result sqlQueryResult, expected ...string) {
	t.Helper()
	require.Equal(t, expected, result.Columns)
}

func columnValues(t *testing.T, result sqlQueryResult, column string) []string {
	values := make([]string, 0, len(result.Rows))
	for _, row := range result.Rows {
		value := row[column]
		require.NotNilf(t, value, "column %s returned NULL", column)
		values = append(values, *value)
	}
	sort.Strings(values)
	return values
}

func requireRowValue(t *testing.T, row map[string]*string, column string) string {
	t.Helper()
	value := row[column]
	require.NotNilf(t, value, "column %s must not be null", column)
	return *value
}

func seedInviteVerificationOverviewFixture(t *testing.T, tx *sql.Tx) {
	t.Helper()

	boundAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC)
	operatorCreatedAt := time.Date(2026, 4, 1, 8, 0, 0, 0, time.UTC)

	insertVerificationUser(t, tx, 1001, "overview-inviter@example.com", "IV-OV-1001", nil, nil, createdAt)
	insertVerificationUser(t, tx, 1002, "overview-invitee@example.com", "IV-OV-1002", int64Ptr(1001), &boundAt, createdAt)
	insertVerificationUser(t, tx, 1003, "overview-operator@example.com", "IV-OV-1003", nil, nil, operatorCreatedAt)
	insertVerificationUser(t, tx, 1004, "overview-rebind-target@example.com", "IV-OV-1004", nil, nil, createdAt)

	insertVerificationRedeemCode(t, tx, 2001, "OVERVIEW-CODE-1", "balance", 100, "unused", "commercial", createdAt)
	insertVerificationAdminAction(t, tx, 3001, "manual_reward_grant", 1003, 1002, "manual correction", createdAt)
	insertVerificationAdminAction(t, tx, 3002, "recompute_rewards", 1003, 1002, "recompute correction", createdAt)

	insertVerificationRelationshipEvent(t, tx, 4001, 1002, nil, int64Ptr(1001), "register_bind", boundAt, nil, "")
	insertVerificationRelationshipEvent(t, tx, 4002, 1004, nil, int64Ptr(1003), "admin_rebind", boundAt.Add(30*time.Minute), int64Ptr(1003), "overview rebind sample")

	insertVerificationRewardRecord(t, tx, verificationRewardRecord{
		ID:                  5001,
		InviterUserID:       1001,
		InviteeUserID:       1002,
		TriggerRedeemCodeID: int64Ptr(2001),
		TriggerValue:        100,
		RewardTargetUserID:  1001,
		RewardRole:          "inviter",
		RewardType:          "base_invite_reward",
		RewardAmount:        "5.00000000",
		Status:              "applied",
		CreatedAt:           createdAt.Add(time.Hour),
	})
	insertVerificationRewardRecord(t, tx, verificationRewardRecord{
		ID:                 5002,
		InviterUserID:      1001,
		InviteeUserID:      1002,
		RewardTargetUserID: 1002,
		RewardRole:         "invitee",
		RewardType:         "manual_invite_grant",
		RewardAmount:       "18.50000000",
		Status:             "applied",
		AdminActionID:      int64Ptr(3001),
		CreatedAt:          createdAt.Add(2 * time.Hour),
	})
	insertVerificationRewardRecord(t, tx, verificationRewardRecord{
		ID:                 5003,
		InviterUserID:      1001,
		InviteeUserID:      1002,
		RewardTargetUserID: 1001,
		RewardRole:         "inviter",
		RewardType:         "recompute_delta",
		RewardAmount:       "2.50000000",
		Status:             "applied",
		AdminActionID:      int64Ptr(3002),
		CreatedAt:          createdAt.Add(3 * time.Hour),
	})
}

func seedInviteVerificationBindingFixture(t *testing.T, tx *sql.Tx) {
	t.Helper()

	base := time.Date(2026, 4, 2, 9, 0, 0, 0, time.UTC)

	insertVerificationUser(t, tx, 1101, "binding-inviter-a@example.com", "IV-BD-1101", nil, nil, base)
	insertVerificationUser(t, tx, 1102, "binding-inviter-b@example.com", "IV-BD-1102", nil, nil, base)

	boundAt1104 := base.Add(1 * time.Hour)
	boundAt1105 := base.Add(2 * time.Hour)
	boundAt1107 := base.Add(4 * time.Hour)
	boundAt1108 := base.Add(5 * time.Hour)
	boundAt1109 := base.Add(6 * time.Hour)

	insertVerificationUser(t, tx, 1104, "binding-matched@example.com", "IV-BD-1104", int64Ptr(1101), timePtr(boundAt1104), base)
	insertVerificationRelationshipEvent(t, tx, 4101, 1104, nil, int64Ptr(1101), "register_bind", boundAt1104, nil, "")

	insertVerificationUser(t, tx, 1105, "binding-missing-event@example.com", "IV-BD-1105", int64Ptr(1101), timePtr(boundAt1105), base)

	insertVerificationUser(t, tx, 1106, "binding-orphan-event@example.com", "IV-BD-1106", nil, nil, base)
	insertVerificationRelationshipEvent(t, tx, 4102, 1106, nil, int64Ptr(1101), "register_bind", base.Add(3*time.Hour), nil, "")

	insertVerificationUser(t, tx, 1107, "binding-duplicate@example.com", "IV-BD-1107", int64Ptr(1101), timePtr(boundAt1107), base)
	insertVerificationRelationshipEvent(t, tx, 4103, 1107, nil, int64Ptr(1101), "register_bind", boundAt1107, nil, "")
	insertVerificationRelationshipEvent(t, tx, 4104, 1107, nil, int64Ptr(1101), "register_bind", boundAt1107.Add(10*time.Minute), nil, "")

	insertVerificationUser(t, tx, 1108, "binding-inviter-mismatch@example.com", "IV-BD-1108", int64Ptr(1101), timePtr(boundAt1108), base)
	insertVerificationRelationshipEvent(t, tx, 4105, 1108, nil, int64Ptr(1102), "register_bind", boundAt1108, nil, "")

	insertVerificationUser(t, tx, 1109, "binding-effective-at-mismatch@example.com", "IV-BD-1109", int64Ptr(1101), timePtr(boundAt1109), base)
	insertVerificationRelationshipEvent(t, tx, 4106, 1109, nil, int64Ptr(1101), "register_bind", boundAt1109.Add(15*time.Minute), nil, "")
}

func seedInviteVerificationRewardFixture(t *testing.T, tx *sql.Tx) {
	t.Helper()

	base := time.Date(2026, 4, 3, 9, 0, 0, 0, time.UTC)
	boundAt1202 := base.Add(1 * time.Hour)
	boundAt1203 := base.Add(2 * time.Hour)

	insertVerificationUser(t, tx, 1201, "reward-inviter@example.com", "IV-RW-1201", nil, nil, base)
	insertVerificationUser(t, tx, 1202, "reward-invitee-a@example.com", "IV-RW-1202", int64Ptr(1201), timePtr(boundAt1202), base)
	insertVerificationUser(t, tx, 1203, "reward-invitee-b@example.com", "IV-RW-1203", int64Ptr(1201), timePtr(boundAt1203), base)
	insertVerificationUser(t, tx, 1204, "reward-operator@example.com", "IV-RW-1204", nil, nil, base)

	insertVerificationRedeemCode(t, tx, 2201, "REWARD-CODE-2201", "balance", 100, "unused", "commercial", base)
	insertVerificationRedeemCode(t, tx, 2202, "REWARD-CODE-2202", "balance", 50, "unused", "commercial", base)

	insertVerificationAdminAction(t, tx, 3201, "manual_reward_grant", 1204, 1202, "manual grant sample", base.Add(3*time.Hour))
	insertVerificationAdminAction(t, tx, 3202, "recompute_rewards", 1204, 1201, "recompute sample", base.Add(4*time.Hour))
	insertVerificationAdminAction(t, tx, 3203, "manual_reward_grant", 1204, 1201, "bad base linkage sample", base.Add(5*time.Hour))

	insertVerificationRewardRecord(t, tx, verificationRewardRecord{
		ID:                  5201,
		InviterUserID:       1201,
		InviteeUserID:       1202,
		TriggerRedeemCodeID: int64Ptr(2201),
		TriggerValue:        100,
		RewardTargetUserID:  1201,
		RewardRole:          "inviter",
		RewardType:          "base_invite_reward",
		RewardAmount:        "5.00000000",
		Status:              "applied",
		CreatedAt:           base.Add(6 * time.Hour),
	})
	insertVerificationRewardRecord(t, tx, verificationRewardRecord{
		ID:                  5202,
		InviterUserID:       1201,
		InviteeUserID:       1202,
		TriggerRedeemCodeID: int64Ptr(2202),
		TriggerValue:        50,
		RewardTargetUserID:  1201,
		RewardRole:          "inviter",
		RewardType:          "base_invite_reward",
		RewardAmount:        "7.00000000",
		Status:              "applied",
		AdminActionID:       int64Ptr(3203),
		CreatedAt:           base.Add(7 * time.Hour),
	})
	insertVerificationRewardRecord(t, tx, verificationRewardRecord{
		ID:                 5203,
		InviterUserID:      1201,
		InviteeUserID:      1203,
		RewardTargetUserID: 1203,
		RewardRole:         "invitee",
		RewardType:         "base_invite_reward",
		RewardAmount:       "3.00000000",
		Status:             "applied",
		CreatedAt:          base.Add(8 * time.Hour),
	})
	insertVerificationRewardRecord(t, tx, verificationRewardRecord{
		ID:                 5204,
		InviterUserID:      1201,
		InviteeUserID:      1202,
		RewardTargetUserID: 1202,
		RewardRole:         "invitee",
		RewardType:         "manual_invite_grant",
		RewardAmount:       "18.50000000",
		Status:             "applied",
		AdminActionID:      int64Ptr(3201),
		CreatedAt:          base.Add(9 * time.Hour),
	})
	insertVerificationRewardRecord(t, tx, verificationRewardRecord{
		ID:                 5205,
		InviterUserID:      1201,
		InviteeUserID:      1203,
		RewardTargetUserID: 1203,
		RewardRole:         "invitee",
		RewardType:         "manual_invite_grant",
		RewardAmount:       "9.00000000",
		Status:             "applied",
		CreatedAt:          base.Add(10 * time.Hour),
	})
	insertVerificationRewardRecord(t, tx, verificationRewardRecord{
		ID:                 5206,
		InviterUserID:      1201,
		InviteeUserID:      1202,
		RewardTargetUserID: 1201,
		RewardRole:         "inviter",
		RewardType:         "recompute_delta",
		RewardAmount:       "1.00000000",
		Status:             "applied",
		AdminActionID:      int64Ptr(3202),
		CreatedAt:          base.Add(11 * time.Hour),
	})
	insertVerificationRewardRecord(t, tx, verificationRewardRecord{
		ID:                 5207,
		InviterUserID:      1201,
		InviteeUserID:      1203,
		RewardTargetUserID: 1203,
		RewardRole:         "invitee",
		RewardType:         "recompute_delta",
		RewardAmount:       "2.00000000",
		Status:             "applied",
		CreatedAt:          base.Add(12 * time.Hour),
	})
}

type verificationRewardRecord struct {
	ID                  int64
	InviterUserID       int64
	InviteeUserID       int64
	TriggerRedeemCodeID *int64
	TriggerValue        float64
	RewardTargetUserID  int64
	RewardRole          string
	RewardType          string
	RewardAmount        string
	Status              string
	AdminActionID       *int64
	CreatedAt           time.Time
}

func insertVerificationUser(t *testing.T, tx *sql.Tx, id int64, email, inviteCode string, invitedBy *int64, boundAt *time.Time, createdAt time.Time) {
	t.Helper()
	_, err := tx.ExecContext(context.Background(), `
		INSERT INTO users (
			id, email, password_hash, role, status, invite_code, invited_by_user_id, invite_bound_at, created_at, updated_at
		) VALUES ($1, $2, 'hash', 'user', 'active', $3, $4, $5, $6, $6)
	`, id, email, inviteCode, invitedBy, boundAt, createdAt)
	require.NoError(t, err)
}

func insertVerificationRedeemCode(t *testing.T, tx *sql.Tx, id int64, code, redeemType string, value float64, status, sourceType string, createdAt time.Time) {
	t.Helper()
	_, err := tx.ExecContext(context.Background(), `
		INSERT INTO redeem_codes (
			id, code, type, value, status, source_type, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, code, redeemType, value, status, sourceType, createdAt)
	require.NoError(t, err)
}

func insertVerificationAdminAction(t *testing.T, tx *sql.Tx, id int64, actionType string, operatorUserID, targetUserID int64, reason string, createdAt time.Time) {
	t.Helper()
	_, err := tx.ExecContext(context.Background(), `
		INSERT INTO invite_admin_actions (
			id, action_type, operator_user_id, target_user_id, reason, request_snapshot_json, result_snapshot_json, created_at
		) VALUES ($1, $2, $3, $4, $5, '{}'::jsonb, '{}'::jsonb, $6)
	`, id, actionType, operatorUserID, targetUserID, reason, createdAt)
	require.NoError(t, err)
}

func insertVerificationRelationshipEvent(t *testing.T, tx *sql.Tx, id, inviteeUserID int64, previousInviterUserID, newInviterUserID *int64, eventType string, effectiveAt time.Time, operatorUserID *int64, reason string) {
	t.Helper()
	_, err := tx.ExecContext(context.Background(), `
		INSERT INTO invite_relationship_events (
			id, invitee_user_id, previous_inviter_user_id, new_inviter_user_id, event_type, effective_at, operator_user_id, reason, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), $6)
	`, id, inviteeUserID, previousInviterUserID, newInviterUserID, eventType, effectiveAt, operatorUserID, reason)
	require.NoError(t, err)
}

func insertVerificationRewardRecord(t *testing.T, tx *sql.Tx, row verificationRewardRecord) {
	t.Helper()
	_, err := tx.ExecContext(context.Background(), `
		INSERT INTO invite_reward_records (
			id, inviter_user_id, invitee_user_id, trigger_redeem_code_id, trigger_redeem_code_value,
			reward_target_user_id, reward_role, reward_type, reward_amount, status, admin_action_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::decimal(20,8), $10, $11, $12)
	`, row.ID, row.InviterUserID, row.InviteeUserID, row.TriggerRedeemCodeID, row.TriggerValue, row.RewardTargetUserID, row.RewardRole, row.RewardType, row.RewardAmount, row.Status, row.AdminActionID, row.CreatedAt)
	require.NoError(t, err)
}

func int64Ptr(v int64) *int64 {
	return &v
}

func timePtr(v time.Time) *time.Time {
	return &v
}
