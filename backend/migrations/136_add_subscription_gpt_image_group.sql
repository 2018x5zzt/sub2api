-- Add a product-settled OpenAI image subscription group and retire legacy daily group key entry points.

DROP INDEX IF EXISTS subscription_product_groups_group_unique_active;

WITH source_group AS (
    SELECT *
    FROM groups
    WHERE id = 30
      AND deleted_at IS NULL
),
inserted_group AS (
    INSERT INTO groups (
        name,
        description,
        rate_multiplier,
        is_exclusive,
        status,
        platform,
        subscription_type,
        daily_limit_usd,
        weekly_limit_usd,
        monthly_limit_usd,
        default_validity_days,
        image_price_1k,
        image_price_2k,
        image_price_4k,
        claude_code_only,
        fallback_group_id,
        model_routing,
        model_routing_enabled,
        fallback_group_id_on_invalid_request,
        mcp_xml_inject,
        supported_model_scopes,
        sora_image_price_360,
        sora_image_price_540,
        sora_video_price_per_request,
        sora_video_price_per_request_hd,
        sort_order,
        sora_storage_quota_bytes,
        allow_messages_dispatch,
        default_mapped_model,
        pricing_mode,
        default_budget_multiplier,
        created_at,
        updated_at
    )
    SELECT
        '【订阅】gpt-image',
        description,
        rate_multiplier,
        is_exclusive,
        'active',
        platform,
        'subscription',
        daily_limit_usd,
        weekly_limit_usd,
        monthly_limit_usd,
        default_validity_days,
        image_price_1k,
        image_price_2k,
        image_price_4k,
        claude_code_only,
        fallback_group_id,
        model_routing,
        model_routing_enabled,
        fallback_group_id_on_invalid_request,
        mcp_xml_inject,
        supported_model_scopes,
        sora_image_price_360,
        sora_image_price_540,
        sora_video_price_per_request,
        sora_video_price_per_request_hd,
        sort_order + 1,
        sora_storage_quota_bytes,
        allow_messages_dispatch,
        default_mapped_model,
        pricing_mode,
        default_budget_multiplier,
        NOW(),
        NOW()
    FROM source_group
    WHERE NOT EXISTS (
        SELECT 1
        FROM groups
        WHERE name = '【订阅】gpt-image'
          AND deleted_at IS NULL
    )
    RETURNING id
)
UPDATE groups target
SET
    description = source_group.description,
    rate_multiplier = source_group.rate_multiplier,
    is_exclusive = source_group.is_exclusive,
    status = 'active',
    platform = source_group.platform,
    subscription_type = 'subscription',
    daily_limit_usd = source_group.daily_limit_usd,
    weekly_limit_usd = source_group.weekly_limit_usd,
    monthly_limit_usd = source_group.monthly_limit_usd,
    default_validity_days = source_group.default_validity_days,
    image_price_1k = source_group.image_price_1k,
    image_price_2k = source_group.image_price_2k,
    image_price_4k = source_group.image_price_4k,
    claude_code_only = source_group.claude_code_only,
    fallback_group_id = source_group.fallback_group_id,
    model_routing = source_group.model_routing,
    model_routing_enabled = source_group.model_routing_enabled,
    fallback_group_id_on_invalid_request = source_group.fallback_group_id_on_invalid_request,
    mcp_xml_inject = source_group.mcp_xml_inject,
    supported_model_scopes = source_group.supported_model_scopes,
    sora_image_price_360 = source_group.sora_image_price_360,
    sora_image_price_540 = source_group.sora_image_price_540,
    sora_video_price_per_request = source_group.sora_video_price_per_request,
    sora_video_price_per_request_hd = source_group.sora_video_price_per_request_hd,
    sora_storage_quota_bytes = source_group.sora_storage_quota_bytes,
    allow_messages_dispatch = source_group.allow_messages_dispatch,
    default_mapped_model = source_group.default_mapped_model,
    pricing_mode = source_group.pricing_mode,
    default_budget_multiplier = source_group.default_budget_multiplier,
    updated_at = NOW()
FROM source_group
WHERE target.name = '【订阅】gpt-image'
  AND target.deleted_at IS NULL;

INSERT INTO account_groups (
    account_id,
    group_id,
    priority,
    billing_multiplier,
    created_at
)
SELECT
    ag.account_id,
    target.id,
    ag.priority,
    ag.billing_multiplier,
    NOW()
FROM account_groups ag
JOIN groups target ON target.name = '【订阅】gpt-image' AND target.deleted_at IS NULL
WHERE ag.group_id = 30
ON CONFLICT (account_id, group_id) DO UPDATE
SET
    priority = EXCLUDED.priority,
    billing_multiplier = EXCLUDED.billing_multiplier;

INSERT INTO subscription_product_groups (
    product_id,
    group_id,
    debit_multiplier,
    status,
    sort_order,
    created_at,
    updated_at
)
SELECT
    sp.id,
    target.id,
    0.5,
    'active',
    30,
    NOW(),
    NOW()
FROM subscription_products sp
JOIN groups target ON target.name = '【订阅】gpt-image' AND target.deleted_at IS NULL
WHERE sp.code IN ('gpt_daily_45', 'gpt_daily_150', 'gpt_daily_225', 'gpt_daily_450')
  AND sp.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM subscription_product_groups existing
      WHERE existing.product_id = sp.id
        AND existing.group_id = target.id
        AND existing.deleted_at IS NULL
  );

UPDATE subscription_product_groups spg
SET
    debit_multiplier = 0.5,
    status = 'active',
    sort_order = 30,
    updated_at = NOW()
FROM subscription_products sp, groups target
WHERE spg.product_id = sp.id
  AND spg.group_id = target.id
  AND sp.code IN ('gpt_daily_45', 'gpt_daily_150', 'gpt_daily_225', 'gpt_daily_450')
  AND sp.deleted_at IS NULL
  AND target.name = '【订阅】gpt-image'
  AND target.deleted_at IS NULL
  AND spg.deleted_at IS NULL;

UPDATE api_keys
SET
    group_id = 21,
    updated_at = NOW()
WHERE group_id IN (22, 23, 32)
  AND status = 'active'
  AND deleted_at IS NULL;
