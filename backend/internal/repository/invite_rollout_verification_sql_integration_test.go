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
		require.True(t,
			strings.HasPrefix(upper, "SELECT") || strings.HasPrefix(upper, "WITH"),
			"verification SQL must stay read-only, got: %s",
			trimmed,
		)
		require.False(t, mutatingStatementRE.MatchString(upper),
			"verification SQL must stay read-only; detected mutation keyword in: %s",
			trimmed,
		)
		out = append(out, stmt)
	}
	return out
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

func columnValues(result sqlQueryResult, column string) []string {
	values := make([]string, 0, len(result.Rows))
	for _, row := range result.Rows {
		if v := row[column]; v != nil {
			values = append(values, *v)
		} else {
			values = append(values, "")
		}
	}
	sort.Strings(values)
	return values
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
