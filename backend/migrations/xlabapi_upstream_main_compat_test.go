package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXlabapiCompatibilitySchemaContainsSharedSubscriptionProducts(t *testing.T) {
	content, err := FS.ReadFile("138_xlabapi_upstream_main_subscription_compat.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS subscription_products")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS subscription_product_groups")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS user_product_subscriptions")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS product_subscription_migration_sources")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS product_id")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS product_subscription_id")
	require.Contains(t, sql, "CREATE UNIQUE INDEX IF NOT EXISTS subscription_products_code_unique_active")
}

func TestXlabapiCompatibilitySchemaContainsDailyCarryoverFields(t *testing.T) {
	content, err := FS.ReadFile("138_xlabapi_upstream_main_subscription_compat.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "ALTER TABLE user_subscriptions")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS daily_carryover_in_usd")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS daily_carryover_remaining_usd")
	require.Contains(t, sql, "legacy_daily_carryover_in_usd")
	require.Contains(t, sql, "legacy_daily_carryover_remaining_usd")
}

func TestXlabapiCompatibilitySchemaContainsChannelPricingPlatform(t *testing.T) {
	content, err := FS.ReadFile("139_xlabapi_upstream_main_channel_affiliate_compat.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "ALTER TABLE channel_model_pricing")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS platform")
	require.Contains(t, sql, "COMMENT ON COLUMN channel_model_pricing.platform")
}

func TestXlabapiCompatibilitySchemaContainsAffiliateCompatibilityFields(t *testing.T) {
	content, err := FS.ReadFile("139_xlabapi_upstream_main_channel_affiliate_compat.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS user_affiliates")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS user_affiliate_ledger")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS legacy_invite_reward_record_id")
	require.Contains(t, sql, "available_channels_enabled")
	require.Contains(t, sql, "affiliate_rebate_freeze_hours")
}
