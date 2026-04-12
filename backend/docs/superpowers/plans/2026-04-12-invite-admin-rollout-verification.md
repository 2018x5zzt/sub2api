# Invite Admin Rollout Verification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a read-only SQL rollout verification checklist for the invite admin rollout, plus integration coverage that proves the SQL stays read-only, executable, and semantically aligned with the spec.

**Architecture:** Store the operator-facing checklist as plain SQL under `resources/sql/ops`, not as a migration. Validate it with a repository integration test that seeds minimal invite data inside a single `sql.Tx`, executes each SQL statement sequentially, and asserts the expected summary metrics, anomaly metrics, sample rows, and final statement contract.

**Tech Stack:** Go, `database/sql`, `github.com/lib/pq`, PostgreSQL, Testcontainers-backed repository integration tests, plain SQL

---

## File Structure

- `backend/resources/sql/ops/invite_admin_rollout_verification.sql`
  - The operator-facing read-only SQL checklist.
  - Final contract: 9 executable statements in this order:
    1. fixed metric overview
    2. binding alignment metrics
    3. missing `register_bind` samples
    4. duplicate `register_bind` samples
    5. inviter/effective-at mismatch samples
    6. reward summary by `reward_type` and `status`
    7. reward anomaly metrics
    8. reward attribution anomaly samples
    9. base reward observation samples
- `backend/internal/repository/invite_rollout_verification_sql_integration_test.go`
  - Integration-only test file for the SQL contract.
  - Owns fixture inserts, SQL file loading, statement splitting, dynamic row scanning, metric mapping, and result-column assertions.

## Task 1: Scaffold SQL Execution Coverage and the Overview Block

**Files:**
- Create: `backend/internal/repository/invite_rollout_verification_sql_integration_test.go`
- Create: `backend/resources/sql/ops/invite_admin_rollout_verification.sql`
- Test: `backend/internal/repository/invite_rollout_verification_sql_integration_test.go`

- [ ] **Step 1: Write the failing test**

Add the new integration test file with the first contract test plus the reusable helpers that later tasks will extend:

```go
//go:build integration

package repository

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type sqlQueryResult struct {
	Columns []string
	Rows    []map[string]string
}

func TestInviteRolloutVerificationSQL_OverviewMetricsAndReadOnlyStatements(t *testing.T) {
	tx := testTx(t)
	seedInviteVerificationOverviewFixture(t, tx)

	statements := executableVerificationStatements(t, readInviteRolloutVerificationSQL(t))
	require.NotEmpty(t, statements)

	results := executeVerificationStatements(t, tx, statements[:1])
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
		out = append(out, stmt)
	}
	return out
}

func executeVerificationStatements(t *testing.T, tx *sql.Tx, statements []string) []sqlQueryResult {
	t.Helper()

	ctx := context.Background()
	results := make([]sqlQueryResult, 0, len(statements))
	for _, stmt := range statements {
		rows, err := tx.QueryContext(ctx, stmt)
		require.NoError(t, err)

		cols, err := rows.Columns()
		require.NoError(t, err)

		result := sqlQueryResult{Columns: cols}
		raw := make([]sql.RawBytes, len(cols))
		args := make([]any, len(cols))
		for i := range raw {
			args[i] = &raw[i]
		}

		for rows.Next() {
			require.NoError(t, rows.Scan(args...))

			row := make(map[string]string, len(cols))
			for i, col := range cols {
				row[col] = string(raw[i])
			}
			result.Rows = append(result.Rows, row)
		}

		require.NoError(t, rows.Err())
		require.NoError(t, rows.Close())
		results = append(results, result)
	}
	return results
}

func metricMap(t *testing.T, result sqlQueryResult) map[string]string {
	t.Helper()
	require.Equal(t, []string{"metric_name", "metric_value"}, result.Columns)

	out := make(map[string]string, len(result.Rows))
	for _, row := range result.Rows {
		out[row["metric_name"]] = row["metric_value"]
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
		values = append(values, row[column])
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
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test -tags=integration ./internal/repository -run 'TestInviteRolloutVerificationSQL_OverviewMetricsAndReadOnlyStatements' -v
```

Expected: FAIL because `backend/resources/sql/ops/invite_admin_rollout_verification.sql` does not exist yet.

- [ ] **Step 3: Write minimal implementation**

Create the SQL file with the first statement only:

```sql
-- Pre-deploy and post-deploy invite admin rollout verification checklist.
-- This file is read-only and must not be treated as a migration.

WITH metrics AS (
  SELECT 1 AS ord, 'bound_users_total' AS metric_name, COUNT(*)::text AS metric_value
  FROM users
  WHERE invited_by_user_id IS NOT NULL

  UNION ALL

  SELECT 2, 'register_bind_event_rows_total', COUNT(*)::text
  FROM invite_relationship_events
  WHERE event_type = 'register_bind'

  UNION ALL

  SELECT 3, 'register_bind_distinct_invitees_total', COUNT(DISTINCT invitee_user_id)::text
  FROM invite_relationship_events
  WHERE event_type = 'register_bind'

  UNION ALL

  SELECT 4, 'base_invite_reward_rows_total', COUNT(*)::text
  FROM invite_reward_records
  WHERE reward_type = 'base_invite_reward'

  UNION ALL

  SELECT 5, 'base_invite_reward_amount_total', COALESCE(SUM(reward_amount), 0)::text
  FROM invite_reward_records
  WHERE reward_type = 'base_invite_reward'

  UNION ALL

  SELECT 6, 'admin_rebind_event_rows_total', COUNT(*)::text
  FROM invite_relationship_events
  WHERE event_type = 'admin_rebind'

  UNION ALL

  SELECT 7, 'manual_invite_grant_rows_total', COUNT(*)::text
  FROM invite_reward_records
  WHERE reward_type = 'manual_invite_grant'

  UNION ALL

  SELECT 8, 'recompute_delta_rows_total', COUNT(*)::text
  FROM invite_reward_records
  WHERE reward_type = 'recompute_delta'
)
SELECT metric_name, metric_value
FROM metrics
ORDER BY ord;
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test -tags=integration ./internal/repository -run 'TestInviteRolloutVerificationSQL_OverviewMetricsAndReadOnlyStatements' -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/repository/invite_rollout_verification_sql_integration_test.go backend/resources/sql/ops/invite_admin_rollout_verification.sql
git commit -m "test: scaffold invite rollout verification sql"
```

## Task 2: Add Binding Alignment Metrics and Drift Samples

**Files:**
- Modify: `backend/internal/repository/invite_rollout_verification_sql_integration_test.go`
- Modify: `backend/resources/sql/ops/invite_admin_rollout_verification.sql`
- Test: `backend/internal/repository/invite_rollout_verification_sql_integration_test.go`

- [ ] **Step 1: Write the failing test**

Extend the same test file with a binding-focused fixture and assertions for statements 2-5:

```go
func TestInviteRolloutVerificationSQL_BindingChecksExposeDriftSamples(t *testing.T) {
	tx := testTx(t)
	seedInviteVerificationBindingFixture(t, tx)

	statements := executableVerificationStatements(t, readInviteRolloutVerificationSQL(t))
	require.GreaterOrEqual(t, len(statements), 5)

	results := executeVerificationStatements(t, tx, statements[:5])
	metrics := metricMap(t, results[1])

	require.Equal(t, "1", metrics["bound_users_missing_register_bind_total"])
	require.Equal(t, "1", metrics["register_bind_without_bound_user_total"])
	require.Equal(t, "1", metrics["register_bind_duplicate_invitee_total"])
	require.Equal(t, "1", metrics["register_bind_inviter_mismatch_total"])
	require.Equal(t, "1", metrics["register_bind_effective_at_mismatch_total"])

	requireColumns(t, results[2], "invitee_user_id", "current_inviter_user_id", "expected_effective_at")
	require.Equal(t, []string{"1105"}, columnValues(results[2], "invitee_user_id"))

	requireColumns(t, results[3], "invitee_user_id", "register_bind_count", "first_effective_at", "last_effective_at")
	require.Equal(t, []string{"1107"}, columnValues(results[3], "invitee_user_id"))

	requireColumns(t, results[4], "invitee_user_id", "current_inviter_user_id", "event_inviter_user_id", "expected_effective_at", "event_effective_at")
	require.ElementsMatch(t, []string{"1108", "1109"}, columnValues(results[4], "invitee_user_id"))
}

func seedInviteVerificationBindingFixture(t *testing.T, tx *sql.Tx) {
	t.Helper()

	base := time.Date(2026, 4, 2, 9, 0, 0, 0, time.UTC)
	insertVerificationUser(t, tx, 1101, "binding-inviter-a@example.com", "IV-BD-1101", nil, nil, base)
	insertVerificationUser(t, tx, 1102, "binding-inviter-b@example.com", "IV-BD-1102", nil, nil, base)

	insertVerificationUser(t, tx, 1104, "binding-match@example.com", "IV-BD-1104", int64Ptr(1101), timePtr(base.Add(time.Hour)), base)
	insertVerificationRelationshipEvent(t, tx, 4101, 1104, nil, int64Ptr(1101), "register_bind", base.Add(time.Hour), nil, "")

	insertVerificationUser(t, tx, 1105, "binding-missing-event@example.com", "IV-BD-1105", int64Ptr(1101), timePtr(base.Add(2*time.Hour)), base)

	insertVerificationUser(t, tx, 1106, "binding-orphan-event@example.com", "IV-BD-1106", nil, nil, base)
	insertVerificationRelationshipEvent(t, tx, 4102, 1106, nil, int64Ptr(1101), "register_bind", base.Add(3*time.Hour), nil, "")

	insertVerificationUser(t, tx, 1107, "binding-duplicate@example.com", "IV-BD-1107", int64Ptr(1101), timePtr(base.Add(4*time.Hour)), base)
	insertVerificationRelationshipEvent(t, tx, 4103, 1107, nil, int64Ptr(1101), "register_bind", base.Add(4*time.Hour), nil, "")
	insertVerificationRelationshipEvent(t, tx, 4104, 1107, nil, int64Ptr(1101), "register_bind", base.Add(5*time.Hour), nil, "")

	insertVerificationUser(t, tx, 1108, "binding-inviter-mismatch@example.com", "IV-BD-1108", int64Ptr(1101), timePtr(base.Add(6*time.Hour)), base)
	insertVerificationRelationshipEvent(t, tx, 4105, 1108, nil, int64Ptr(1102), "register_bind", base.Add(6*time.Hour), nil, "")

	insertVerificationUser(t, tx, 1109, "binding-time-mismatch@example.com", "IV-BD-1109", int64Ptr(1101), timePtr(base.Add(7*time.Hour)), base)
	insertVerificationRelationshipEvent(t, tx, 4106, 1109, nil, int64Ptr(1101), "register_bind", base.Add(8*time.Hour), nil, "")
}

func timePtr(v time.Time) *time.Time {
	return &v
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test -tags=integration ./internal/repository -run 'TestInviteRolloutVerificationSQL_BindingChecksExposeDriftSamples' -v
```

Expected: FAIL because the SQL file still exposes only the overview statement.

- [ ] **Step 3: Write minimal implementation**

Append four more read-only statements to the SQL file immediately after the overview block:

```sql
WITH bound_users AS (
  SELECT
    id AS invitee_user_id,
    invited_by_user_id AS current_inviter_user_id,
    COALESCE(invite_bound_at, created_at) AS expected_effective_at
  FROM users
  WHERE invited_by_user_id IS NOT NULL
),
register_bind_ranked AS (
  SELECT
    invitee_user_id,
    new_inviter_user_id AS event_inviter_user_id,
    effective_at,
    created_at,
    COUNT(*) OVER (PARTITION BY invitee_user_id) AS register_bind_count,
    ROW_NUMBER() OVER (
      PARTITION BY invitee_user_id
      ORDER BY effective_at DESC, id DESC
    ) AS rn
  FROM invite_relationship_events
  WHERE event_type = 'register_bind'
),
register_bind_primary AS (
  SELECT invitee_user_id, event_inviter_user_id, effective_at, created_at, register_bind_count
  FROM register_bind_ranked
  WHERE rn = 1
),
metrics AS (
  SELECT 1 AS ord, 'bound_users_missing_register_bind_total' AS metric_name, COUNT(*)::text AS metric_value
  FROM bound_users bu
  LEFT JOIN register_bind_primary rbp USING (invitee_user_id)
  WHERE rbp.invitee_user_id IS NULL

  UNION ALL

  SELECT 2, 'register_bind_without_bound_user_total', COUNT(*)::text
  FROM register_bind_primary rbp
  LEFT JOIN bound_users bu USING (invitee_user_id)
  WHERE bu.invitee_user_id IS NULL

  UNION ALL

  SELECT 3, 'register_bind_duplicate_invitee_total', COUNT(*)::text
  FROM register_bind_primary
  WHERE register_bind_count > 1

  UNION ALL

  SELECT 4, 'register_bind_inviter_mismatch_total', COUNT(*)::text
  FROM bound_users bu
  JOIN register_bind_primary rbp USING (invitee_user_id)
  WHERE bu.current_inviter_user_id IS DISTINCT FROM rbp.event_inviter_user_id

  UNION ALL

  SELECT 5, 'register_bind_effective_at_mismatch_total', COUNT(*)::text
  FROM bound_users bu
  JOIN register_bind_primary rbp USING (invitee_user_id)
  WHERE bu.expected_effective_at IS DISTINCT FROM rbp.effective_at
)
SELECT metric_name, metric_value
FROM metrics
ORDER BY ord;

WITH bound_users AS (
  SELECT
    id AS invitee_user_id,
    invited_by_user_id AS current_inviter_user_id,
    COALESCE(invite_bound_at, created_at) AS expected_effective_at
  FROM users
  WHERE invited_by_user_id IS NOT NULL
),
register_bind_primary AS (
  SELECT invitee_user_id
  FROM (
    SELECT
      invitee_user_id,
      ROW_NUMBER() OVER (PARTITION BY invitee_user_id ORDER BY effective_at DESC, id DESC) AS rn
    FROM invite_relationship_events
    WHERE event_type = 'register_bind'
  ) ranked
  WHERE rn = 1
)
SELECT
  bu.invitee_user_id::text AS invitee_user_id,
  bu.current_inviter_user_id::text AS current_inviter_user_id,
  bu.expected_effective_at::text AS expected_effective_at
FROM bound_users bu
LEFT JOIN register_bind_primary rbp USING (invitee_user_id)
WHERE rbp.invitee_user_id IS NULL
ORDER BY bu.invitee_user_id
LIMIT 50;

WITH duplicate_events AS (
  SELECT
    invitee_user_id,
    COUNT(*) AS register_bind_count,
    MIN(effective_at) AS first_effective_at,
    MAX(effective_at) AS last_effective_at
  FROM invite_relationship_events
  WHERE event_type = 'register_bind'
  GROUP BY invitee_user_id
  HAVING COUNT(*) > 1
)
SELECT
  invitee_user_id::text AS invitee_user_id,
  register_bind_count::text AS register_bind_count,
  first_effective_at::text AS first_effective_at,
  last_effective_at::text AS last_effective_at
FROM duplicate_events
ORDER BY invitee_user_id
LIMIT 50;

WITH bound_users AS (
  SELECT
    id AS invitee_user_id,
    invited_by_user_id AS current_inviter_user_id,
    COALESCE(invite_bound_at, created_at) AS expected_effective_at
  FROM users
  WHERE invited_by_user_id IS NOT NULL
),
register_bind_ranked AS (
  SELECT
    invitee_user_id,
    new_inviter_user_id AS event_inviter_user_id,
    effective_at AS event_effective_at,
    ROW_NUMBER() OVER (PARTITION BY invitee_user_id ORDER BY effective_at DESC, id DESC) AS rn
  FROM invite_relationship_events
  WHERE event_type = 'register_bind'
)
SELECT
  bu.invitee_user_id::text AS invitee_user_id,
  bu.current_inviter_user_id::text AS current_inviter_user_id,
  rb.event_inviter_user_id::text AS event_inviter_user_id,
  bu.expected_effective_at::text AS expected_effective_at,
  rb.event_effective_at::text AS event_effective_at
FROM bound_users bu
JOIN register_bind_ranked rb
  ON bu.invitee_user_id = rb.invitee_user_id
 AND rb.rn = 1
WHERE bu.current_inviter_user_id IS DISTINCT FROM rb.event_inviter_user_id
   OR bu.expected_effective_at IS DISTINCT FROM rb.event_effective_at
ORDER BY bu.invitee_user_id
LIMIT 50;
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test -tags=integration ./internal/repository -run 'TestInviteRolloutVerificationSQL_BindingChecksExposeDriftSamples' -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/repository/invite_rollout_verification_sql_integration_test.go backend/resources/sql/ops/invite_admin_rollout_verification.sql
git commit -m "feat: add invite binding rollout verification queries"
```

## Task 3: Add Reward Attribution Metrics and Reward Anomaly Samples

**Files:**
- Modify: `backend/internal/repository/invite_rollout_verification_sql_integration_test.go`
- Modify: `backend/resources/sql/ops/invite_admin_rollout_verification.sql`
- Test: `backend/internal/repository/invite_rollout_verification_sql_integration_test.go`

- [ ] **Step 1: Write the failing test**

Add a reward-focused fixture and assertions for statements 6-9:

```go
func TestInviteRolloutVerificationSQL_RewardChecksSeparateAdminCorrections(t *testing.T) {
	tx := testTx(t)
	seedInviteVerificationRewardFixture(t, tx)

	statements := executableVerificationStatements(t, readInviteRolloutVerificationSQL(t))
	require.GreaterOrEqual(t, len(statements), 9)

	results := executeVerificationStatements(t, tx, statements)
	requireColumns(t, results[5], "reward_type", "status", "rows_total", "reward_amount_total", "distinct_invitees_total", "distinct_reward_targets_total", "rows_with_admin_action_total", "rows_with_null_trigger_code_total")

	baseSummary := rowByValue(t, results[5], "reward_type", "base_invite_reward")
	require.Equal(t, "applied", baseSummary["status"])
	require.Equal(t, "3", baseSummary["rows_total"])
	require.Equal(t, "15.00000000", baseSummary["reward_amount_total"])
	require.Equal(t, "1", baseSummary["rows_with_null_trigger_code_total"])
	require.Equal(t, "1", baseSummary["rows_with_admin_action_total"])

	manualSummary := rowByValue(t, results[5], "reward_type", "manual_invite_grant")
	require.Equal(t, "2", manualSummary["rows_total"])
	require.Equal(t, "27.50000000", manualSummary["reward_amount_total"])
	require.Equal(t, "1", manualSummary["rows_with_admin_action_total"])

	recomputeSummary := rowByValue(t, results[5], "reward_type", "recompute_delta")
	require.Equal(t, "2", recomputeSummary["rows_total"])
	require.Equal(t, "3.00000000", recomputeSummary["reward_amount_total"])
	require.Equal(t, "1", recomputeSummary["rows_with_admin_action_total"])

	metrics := metricMap(t, results[6])
	require.Equal(t, "1", metrics["base_reward_with_admin_action_total"])
	require.Equal(t, "1", metrics["manual_grant_without_admin_action_total"])
	require.Equal(t, "1", metrics["recompute_delta_without_admin_action_total"])
	require.Equal(t, "1", metrics["base_reward_without_trigger_code_total"])

	requireColumns(t, results[7], "id", "reward_type", "reward_role", "reward_amount", "admin_action_id", "trigger_redeem_code_id", "created_at")
	require.ElementsMatch(t, []string{"5202", "5205", "5207"}, columnValues(results[7], "id"))

	requireColumns(t, results[8], "id", "inviter_user_id", "invitee_user_id", "reward_target_user_id", "reward_role", "reward_amount", "trigger_redeem_code_id", "created_at")
	require.ElementsMatch(t, []string{"5201", "5202", "5203"}, columnValues(results[8], "id"))
}

func rowByValue(t *testing.T, result sqlQueryResult, keyColumn, keyValue string) map[string]string {
	t.Helper()
	for _, row := range result.Rows {
		if row[keyColumn] == keyValue {
			return row
		}
	}
	t.Fatalf("missing row where %s=%s", keyColumn, keyValue)
	return nil
}

func seedInviteVerificationRewardFixture(t *testing.T, tx *sql.Tx) {
	t.Helper()

	base := time.Date(2026, 4, 3, 9, 0, 0, 0, time.UTC)

	insertVerificationUser(t, tx, 1201, "reward-inviter@example.com", "IV-RW-1201", nil, nil, base)
	insertVerificationUser(t, tx, 1202, "reward-invitee-a@example.com", "IV-RW-1202", int64Ptr(1201), timePtr(base.Add(time.Hour)), base)
	insertVerificationUser(t, tx, 1203, "reward-invitee-b@example.com", "IV-RW-1203", int64Ptr(1201), timePtr(base.Add(2*time.Hour)), base)
	insertVerificationUser(t, tx, 1204, "reward-operator@example.com", "IV-RW-1204", nil, nil, base)

	insertVerificationRedeemCode(t, tx, 2201, "REWARD-CODE-1", "balance", 100, "unused", "commercial", base)
	insertVerificationRedeemCode(t, tx, 2202, "REWARD-CODE-2", "balance", 50, "unused", "commercial", base)

	insertVerificationAdminAction(t, tx, 3201, "manual_reward_grant", 1204, 1202, "manual fix", base)
	insertVerificationAdminAction(t, tx, 3202, "recompute_rewards", 1204, 1201, "recompute fix", base)
	insertVerificationAdminAction(t, tx, 3203, "manual_reward_grant", 1204, 1201, "bad base linkage sample", base)

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
		CreatedAt:           base.Add(time.Hour),
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
		CreatedAt:           base.Add(2 * time.Hour),
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
		CreatedAt:          base.Add(3 * time.Hour),
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
		CreatedAt:          base.Add(4 * time.Hour),
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
		CreatedAt:          base.Add(5 * time.Hour),
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
		CreatedAt:          base.Add(6 * time.Hour),
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
		CreatedAt:          base.Add(7 * time.Hour),
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test -tags=integration ./internal/repository -run 'TestInviteRolloutVerificationSQL_RewardChecksSeparateAdminCorrections' -v
```

Expected: FAIL because the SQL file does not yet expose reward summary, reward anomaly metrics, or reward sample queries.

- [ ] **Step 3: Write minimal implementation**

Append the final four statements to the SQL file:

```sql
SELECT
  reward_type,
  status,
  COUNT(*)::text AS rows_total,
  COALESCE(SUM(reward_amount), 0)::text AS reward_amount_total,
  COUNT(DISTINCT invitee_user_id)::text AS distinct_invitees_total,
  COUNT(DISTINCT reward_target_user_id)::text AS distinct_reward_targets_total,
  COUNT(*) FILTER (WHERE admin_action_id IS NOT NULL)::text AS rows_with_admin_action_total,
  COUNT(*) FILTER (WHERE trigger_redeem_code_id IS NULL)::text AS rows_with_null_trigger_code_total
FROM invite_reward_records
GROUP BY reward_type, status
ORDER BY
  CASE reward_type
    WHEN 'base_invite_reward' THEN 1
    WHEN 'manual_invite_grant' THEN 2
    WHEN 'recompute_delta' THEN 3
    ELSE 99
  END,
  status;

WITH metrics AS (
  SELECT 1 AS ord, 'base_reward_with_admin_action_total' AS metric_name, COUNT(*)::text AS metric_value
  FROM invite_reward_records
  WHERE reward_type = 'base_invite_reward'
    AND admin_action_id IS NOT NULL

  UNION ALL

  SELECT 2, 'manual_grant_without_admin_action_total', COUNT(*)::text
  FROM invite_reward_records
  WHERE reward_type = 'manual_invite_grant'
    AND admin_action_id IS NULL

  UNION ALL

  SELECT 3, 'recompute_delta_without_admin_action_total', COUNT(*)::text
  FROM invite_reward_records
  WHERE reward_type = 'recompute_delta'
    AND admin_action_id IS NULL

  UNION ALL

  SELECT 4, 'base_reward_without_trigger_code_total', COUNT(*)::text
  FROM invite_reward_records
  WHERE reward_type = 'base_invite_reward'
    AND trigger_redeem_code_id IS NULL
)
SELECT metric_name, metric_value
FROM metrics
ORDER BY ord;

SELECT
  id::text AS id,
  reward_type,
  reward_role,
  reward_amount::text AS reward_amount,
  COALESCE(admin_action_id::text, '') AS admin_action_id,
  COALESCE(trigger_redeem_code_id::text, '') AS trigger_redeem_code_id,
  created_at::text AS created_at
FROM invite_reward_records
WHERE (reward_type = 'base_invite_reward' AND admin_action_id IS NOT NULL)
   OR (reward_type = 'manual_invite_grant' AND admin_action_id IS NULL)
   OR (reward_type = 'recompute_delta' AND admin_action_id IS NULL)
ORDER BY id
LIMIT 50;

SELECT
  id::text AS id,
  inviter_user_id::text AS inviter_user_id,
  invitee_user_id::text AS invitee_user_id,
  reward_target_user_id::text AS reward_target_user_id,
  reward_role,
  reward_amount::text AS reward_amount,
  COALESCE(trigger_redeem_code_id::text, '') AS trigger_redeem_code_id,
  created_at::text AS created_at
FROM invite_reward_records
WHERE reward_type = 'base_invite_reward'
ORDER BY id
LIMIT 50;
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test -tags=integration ./internal/repository -run 'TestInviteRolloutVerificationSQL_RewardChecksSeparateAdminCorrections' -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/repository/invite_rollout_verification_sql_integration_test.go backend/resources/sql/ops/invite_admin_rollout_verification.sql
git commit -m "feat: add invite reward rollout verification queries"
```

## Task 4: Lock the Full Statement Contract and Operator Notes

**Files:**
- Modify: `backend/internal/repository/invite_rollout_verification_sql_integration_test.go`
- Modify: `backend/resources/sql/ops/invite_admin_rollout_verification.sql`
- Test: `backend/internal/repository/invite_rollout_verification_sql_integration_test.go`

- [ ] **Step 1: Write the failing test**

Add a final smoke test for file comments, statement count, read-only behavior, and result-column order:

```go
func TestInviteRolloutVerificationSQL_FullContractAndOperatorNotes(t *testing.T) {
	content := readInviteRolloutVerificationSQL(t)
	require.Contains(t, content, "Pre-deploy and post-deploy")
	require.Contains(t, content, "not be treated as a migration")
	require.Contains(t, content, "rebind_inviter")
	require.Contains(t, content, "manual_reward_grant")
	require.Contains(t, content, "recompute_rewards")

	statements := executableVerificationStatements(t, content)
	require.Len(t, statements, 9)

	tx := testTx(t)
	seedInviteVerificationOverviewFixture(t, tx)
	results := executeVerificationStatements(t, tx, statements)
	require.Len(t, results, 9)

	requireColumns(t, results[0], "metric_name", "metric_value")
	requireColumns(t, results[1], "metric_name", "metric_value")
	requireColumns(t, results[2], "invitee_user_id", "current_inviter_user_id", "expected_effective_at")
	requireColumns(t, results[3], "invitee_user_id", "register_bind_count", "first_effective_at", "last_effective_at")
	requireColumns(t, results[4], "invitee_user_id", "current_inviter_user_id", "event_inviter_user_id", "expected_effective_at", "event_effective_at")
	requireColumns(t, results[5], "reward_type", "status", "rows_total", "reward_amount_total", "distinct_invitees_total", "distinct_reward_targets_total", "rows_with_admin_action_total", "rows_with_null_trigger_code_total")
	requireColumns(t, results[6], "metric_name", "metric_value")
	requireColumns(t, results[7], "id", "reward_type", "reward_role", "reward_amount", "admin_action_id", "trigger_redeem_code_id", "created_at")
	requireColumns(t, results[8], "id", "inviter_user_id", "invitee_user_id", "reward_target_user_id", "reward_role", "reward_amount", "trigger_redeem_code_id", "created_at")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test -tags=integration ./internal/repository -run 'TestInviteRolloutVerificationSQL_FullContractAndOperatorNotes' -v
```

Expected: FAIL until the SQL file comments and final statement ordering exactly match the contract.

- [ ] **Step 3: Write minimal implementation**

Polish the SQL file header comments and keep the 9-statement order fixed:

```sql
-- Pre-deploy and post-deploy invite admin rollout verification checklist.
-- This file is read-only and must not be treated as a migration.
-- If operators ran rebind_inviter, manual_reward_grant, or recompute_rewards
-- during the comparison window, admin-related metric changes may be expected.
-- Without explicit admin actions, binding alignment metrics and
-- base_invite_reward totals should remain stable.
```

Do not change the executable statement order after this point.

- [ ] **Step 4: Run test to verify it passes**

Run the focused smoke test first:

```bash
go test -tags=integration ./internal/repository -run 'TestInviteRolloutVerificationSQL_FullContractAndOperatorNotes' -v
```

Expected: PASS

Then run the complete verification suite for this feature:

```bash
go test -tags=integration ./internal/repository -run 'TestInviteRolloutVerificationSQL_' -v
```

Expected: PASS

Then re-run the invite repository integration suite to ensure the new test file did not regress existing invite coverage:

```bash
go test -tags=integration ./internal/repository -run 'TestInvite(Admin|Growth)RepoSuite|TestInviteRolloutVerificationSQL_' -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/repository/invite_rollout_verification_sql_integration_test.go backend/resources/sql/ops/invite_admin_rollout_verification.sql
git commit -m "test: harden invite rollout verification sql contract"
```
