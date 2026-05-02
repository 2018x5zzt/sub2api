-- Disable upstream subscription_plans that were generated only as a
-- compatibility projection of xlabapi product subscriptions.
--
-- Product subscriptions are not equivalent to legacy group subscriptions:
-- one product can be bound to multiple runtime groups, and one runtime group
-- can be shared by multiple products. The projected subscription_plans do not
-- carry product_id, so leaving them for sale lets checkout fulfill by group_id
-- and can assign the first bound product (for example gpt_daily_45) instead of
-- the product shown in the plan name/features.

UPDATE subscription_plans
SET
    for_sale = FALSE,
    updated_at = NOW()
WHERE for_sale = TRUE
  AND price = 0
  AND features LIKE '%Migrated from xlabapi subscription product%';
