-- Backfill legacy unused balance redeem codes to commercial source type.
-- This preserves historical used/redeemed codes and avoids touching non-balance code types.

UPDATE redeem_codes
SET source_type = 'commercial'
WHERE type = 'balance'
  AND status = 'unused'
  AND COALESCE(BTRIM(source_type), '') <> 'commercial';
