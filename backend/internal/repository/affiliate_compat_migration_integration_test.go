//go:build integration

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

func TestAffiliateCompatMigration_MigratesLegacyInviteDataWithoutDoubleCrediting(t *testing.T) {
	ctx := context.Background()
	tx := testTx(t)

	migrationSQL, err := fs.ReadFile(migrations.FS, "134_xlabapi_invite_to_affiliate_compat.sql")
	require.NoError(t, err, "compat migration must exist")

	schema := "affiliate_compat_" + strings.NewReplacer("/", "_", " ", "_", "-", "_").Replace(strings.ToLower(t.Name()))
	schema = regexp.MustCompile(`[^a-z0-9_]+`).ReplaceAllString(schema, "_")
	if len(schema) > 55 {
		schema = schema[:55]
	}
	schema = fmt.Sprintf("%s_%d", schema, time.Now().UnixNano()%100000)

	_, err = tx.ExecContext(ctx, `CREATE SCHEMA `+quoteIdent(schema))
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `SET LOCAL search_path TO `+quoteIdent(schema)+`, public`)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, affiliateCompatLegacyFixtureSQL)
	require.NoError(t, err)

	_, err = tx.ExecContext(ctx, string(migrationSQL))
	require.NoError(t, err)

	var userCount int
	require.NoError(t, tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_affiliates`).Scan(&userCount))
	require.Equal(t, 4, userCount)

	var inviter struct {
		Code         string
		Count        int
		Quota        float64
		FrozenQuota  float64
		HistoryQuota float64
	}
	require.NoError(t, tx.QueryRowContext(ctx, `
SELECT aff_code, aff_count, aff_quota::double precision, aff_frozen_quota::double precision, aff_history_quota::double precision
FROM user_affiliates
WHERE user_id = 1`).Scan(
		&inviter.Code,
		&inviter.Count,
		&inviter.Quota,
		&inviter.FrozenQuota,
		&inviter.HistoryQuota,
	))
	require.Equal(t, "ABCDEFGH", inviter.Code, "legacy invite codes must be normalized for upstream affiliate lookup")
	require.Equal(t, 2, inviter.Count)
	require.Zero(t, inviter.Quota, "old invite rewards were already credited to user balance and must not become transferable again")
	require.Zero(t, inviter.FrozenQuota)
	require.InDelta(t, 3.0, inviter.HistoryQuota, 0.000001)

	var inviteeInviterID int64
	require.NoError(t, tx.QueryRowContext(ctx, `SELECT inviter_id FROM user_affiliates WHERE user_id = 2`).Scan(&inviteeInviterID))
	require.Equal(t, int64(1), inviteeInviterID)

	var duplicateNormalizedCode string
	require.NoError(t, tx.QueryRowContext(ctx, `SELECT aff_code FROM user_affiliates WHERE user_id = 2`).Scan(&duplicateNormalizedCode))
	require.NotEqual(t, "ABCDEFGH", duplicateNormalizedCode, "case-folded duplicate legacy codes need a deterministic fallback")
	require.Regexp(t, `^[A-Z0-9_-]{4,32}$`, duplicateNormalizedCode)

	var invalidFallbackCode string
	require.NoError(t, tx.QueryRowContext(ctx, `SELECT aff_code FROM user_affiliates WHERE user_id = 3`).Scan(&invalidFallbackCode))
	require.Regexp(t, `^[A-Z0-9_-]{4,32}$`, invalidFallbackCode)

	var ledger struct {
		Action       string
		Amount       float64
		SourceUserID int64
	}
	require.NoError(t, tx.QueryRowContext(ctx, `
SELECT action, amount::double precision, source_user_id
FROM user_affiliate_ledger
WHERE user_id = 1`).Scan(&ledger.Action, &ledger.Amount, &ledger.SourceUserID))
	require.Equal(t, "accrue", ledger.Action)
	require.InDelta(t, 3.0, ledger.Amount, 0.000001)
	require.Equal(t, int64(2), ledger.SourceUserID)

	var inviterLedgerRows int
	require.NoError(t, tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_affiliate_ledger WHERE user_id = 1`).Scan(&inviterLedgerRows))
	require.Equal(t, 1, inviterLedgerRows, "only applied inviter rewards should be carried into affiliate history")

	requireSettingValue(t, tx, "affiliate_enabled", "true")
	requireSettingValue(t, tx, "affiliate_rebate_rate", "3")
	requireSettingValue(t, tx, "available_channels_enabled", "true")

	_, err = tx.ExecContext(ctx, string(migrationSQL))
	require.NoError(t, err, "compat migration must be idempotent")
}

func quoteIdent(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func requireSettingValue(t *testing.T, tx *sql.Tx, key, expected string) {
	t.Helper()

	var got string
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT value FROM settings WHERE key = $1`, key).Scan(&got))
	require.Equal(t, expected, got)
}

const affiliateCompatLegacyFixtureSQL = `
CREATE TABLE users (
  id BIGINT PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  username VARCHAR(100) NOT NULL DEFAULT '',
  password_hash VARCHAR(255) NOT NULL DEFAULT '',
  balance DECIMAL(20,8) NOT NULL DEFAULT 0,
  invite_code VARCHAR(32),
  invited_by_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
  invite_bound_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE settings (
  id BIGSERIAL PRIMARY KEY,
  key VARCHAR(100) NOT NULL UNIQUE,
  value TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE redeem_codes (
  id BIGINT PRIMARY KEY,
  code VARCHAR(32) NOT NULL UNIQUE,
  type VARCHAR(20) NOT NULL DEFAULT 'balance',
  value DECIMAL(20,8) NOT NULL,
  source_type VARCHAR(32) NOT NULL DEFAULT 'commercial',
  status VARCHAR(20) NOT NULL DEFAULT 'used',
  used_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE invite_reward_records (
  id BIGINT PRIMARY KEY,
  inviter_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  invitee_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  trigger_redeem_code_id BIGINT REFERENCES redeem_codes(id) ON DELETE RESTRICT,
  trigger_redeem_code_value DECIMAL(20,8) NOT NULL,
  reward_target_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  reward_role VARCHAR(32) NOT NULL,
  reward_type VARCHAR(64) NOT NULL,
  reward_rate DECIMAL(10,8),
  reward_amount DECIMAL(20,8) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'applied',
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO users (id, email, username, invite_code, invited_by_user_id, invite_bound_at, created_at) VALUES
  (1, 'inviter@example.com', 'inviter', 'abcdEFGH', NULL, NULL, NOW() - INTERVAL '10 days'),
  (2, 'invitee-one@example.com', 'invitee1', 'ABCDEFGH', 1, NOW() - INTERVAL '9 days', NOW() - INTERVAL '9 days'),
  (3, 'invitee-two@example.com', 'invitee2', 'bad code!', 1, NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),
  (4, 'plain@example.com', 'plain', NULL, NULL, NULL, NOW() - INTERVAL '7 days');

INSERT INTO redeem_codes (id, code, value, used_by) VALUES
  (10, 'COMMERCIAL-10', 100, 2),
  (11, 'COMMERCIAL-11', 50, 3),
  (12, 'COMMERCIAL-12', 25, 2);

INSERT INTO invite_reward_records (
  id,
  inviter_user_id,
  invitee_user_id,
  trigger_redeem_code_id,
  trigger_redeem_code_value,
  reward_target_user_id,
  reward_role,
  reward_type,
  reward_rate,
  reward_amount,
  status
) VALUES
  (100, 1, 2, 10, 100, 1, 'inviter', 'base_invite_reward', 0.03, 3, 'applied'),
  (101, 1, 2, 10, 100, 2, 'invitee', 'base_invite_reward', 0.03, 3, 'applied'),
  (102, 1, 3, 11, 50, 1, 'inviter', 'base_invite_reward', 0.03, 1.5, 'pending');
`
