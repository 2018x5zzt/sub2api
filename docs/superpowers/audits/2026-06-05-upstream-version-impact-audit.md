# Upstream Version Impact Audit for xlabapi

## Context

`xlabapi` is the production branch. It diverged from upstream around `48912014` and now contains production-only changes for product subscriptions, affiliate settlement, frontend-v2, enterprise BFF, OpenAI gateway compatibility, and deployment workflows.

The requested strategy is **B then A**:

1. First audit upstream by version ranges.
2. Then merge upstream version ranges one by one, with gates before each range.

This audit uses the refreshed `upstream/main` at:

- `f1aa5896 Merge pull request #2993 from ghostg00/fix/openai-5h-used-percent-direct`

The local `xlabapi` also has unpushed local commits through:

- `77a0f9cd docs(upstream): plan selective upstream ports`

## Important note about the paused selective worktree

There is an isolated worktree at:

`/root/.config/superpowers/worktrees/sub2api-src/upstream-selective-xlabapi-20260605`

It already contains several selected low-risk upstream ports:

- `4808d9a4 test(openai): share compact keepalive repo stub`
- `72c8ee6d fix(security): redact admin account credentials`
- `a62721e3 fix(security): enforce custom page visibility`
- `1a26f4a4 fix(upstream): port low-risk legacy UI and deploy fixes`
- `b182a824 fix(upstream): port totp and ccswitch legacy fixes`
- `eb41b5be fix(openai): support versioned compatible base urls`

That branch is currently clean and not merged into `xlabapi`. Under the new version-by-version strategy, treat it as a reference/salvage branch, not as the primary integration path.

## Version range metrics

Counts are approximate impact indicators based on changed paths in each range.

| Range | Commits | Schema/migration files | Payment/subscription/redeem/quota paths | Gateway paths | Frontend paths | Risk | Summary |
|---|---:|---:|---:|---:|---:|---|---|
| `48912014..v0.1.121` | 0 | 0 | 0 | 0 | 0 | None | Fork baseline / tag only. |
| `v0.1.121..v0.1.122` | 17 | 2 | 1 | 29 | 13 | Medium | OpenAI compatible upstream improvements, affiliate admin records, usage fixes. |
| `v0.1.122..v0.1.123` | 2 | 0 | 0 | 9 | 0 | Low-Medium | Unknown OpenAI model fallback removal and related billing/model behavior. |
| `v0.1.123..v0.1.124` | 25 | 8 | 3 | 41 | 33 | High | Image governance, risk control, GitHub/Google OAuth, markdown pages, redeem affiliate, Codex image bridge. |
| `v0.1.124..v0.1.125` | 6 | 0 | 0 | 1 | 14 | Medium | Terms/legal pages, moderation key UI, model whitelist, default redact-thinking beta fix. |
| `v0.1.125..v0.1.126` | 20 | 0 | 44 | 16 | 41 | High | Airwallex/multi-currency payment, OAuth import, CCSwitch model fix, payment display fixes, Antigravity UA. |
| `v0.1.126..v0.1.127` | 74 | 8 | 28 | 49 | 64 | Very High | Many gateway, payment, subscription, quota, OAuth, frontend and account changes. |
| `v0.1.127..v0.1.128` | 44 | 5 | 9 | 26 | 29 | Very High | Channel monitor, OpenAI force chat completions, payment/mobile QR, model sync, email templates. |
| `v0.1.128..v0.1.129` | 6 | 0 | 0 | 1 | 5 | Medium | API key usage daily detail and group/API key denial fixes. |
| `v0.1.129..v0.1.130` | 33 | 2 | 7 | 24 | 26 | High | Redeem batch update, Bedrock, OIDC, risk control, subscription expiry email toggle, account tests. |
| `v0.1.130..v0.1.131` | 10 | 3 | 10 | 33 | 19 | Very High | User-platform quota and HTTP/2 response header timeout; quota schema risk. |
| `v0.1.131..v0.1.132` | 24 | 2 | 0 | 34 | 18 | High | Ops classification, local business-limited reasons, WS failover, scheduler model cooldown. |
| `v0.1.132..v0.1.133` | 24 | 1 | 0 | 47 | 20 | High | apicompat usage/token fixes, codex_cli_only plugin allowance, endpoint capability gating, account auto-pause. |
| `v0.1.133..upstream/main` | 30 | 0 | 7 | 54 | 18 | Very High | User-platform quota flusher, large gateway refactor, Codex Responses bridge redesign, OpenAI WS/image bridge fixes. |

## Subscription/product/payment impact

The largest risk to existing xlabapi product subscribers starts at `v0.1.125..v0.1.126` and continues through later ranges.

Key indicators:

- `v0.1.125..v0.1.126`
  - `b23055af feat: add Airwallex payments and multi-currency support`
  - payment/order/provider/fulfillment changes across backend and frontends.
  - This can affect payment callbacks, order snapshots, currency/amount math, and subscription fulfillment.

- `v0.1.126..v0.1.127`
  - `61b62721 fix(payment): apply product affix to subscriptions`
  - `a4884b4e fix(subscription): 将日卡改为一次性每日配额`
  - `e4aaf0af feat(redeem): 兑换码支持设置使用有效期`
  - Frontend subscription quota display helpers.
  - These can affect existing product-subscription semantics and daily quota behavior.

- `v0.1.130..v0.1.131` and later
  - `6b39b344 feat(quota): 用户 × 平台 USD 配额`
  - later quota sentinel/flusher changes.
  - These add another quota layer that can change authorization behavior for existing users.

Because `xlabapi` has product subscriptions (`subscription_products`, `product_subscription_id`, shared subscription product migrations, and frontend-v2 product pages), these ranges need explicit product-subscription regression tests before production merge.

## Gateway impact

The gateway changes are valuable, but the highest-risk sections are not isolated bug fixes anymore.

Lower-risk gateway candidates already identified or partially ported:

- `679c0865 fix(openai): handle versioned compatible base URLs`
- `a6117429 fix(gateway): detach upstream context unconditionally for image generation`
- `33ac8eb2 fix openai http2 response header timeout`

Higher-risk gateway ranges:

- `v0.1.132..v0.1.133`
  - endpoint capability gating, OpenAI WS usage, codex_cli_only plugin behavior, account auto-pause.
- `v0.1.133..upstream/main`
  - request body refactor series, request view abstraction, Codex Responses bridge redesign, oversized WS bridge, dedup and usage semantics.

These should be merged only after product/subscription safety gates are green, because gateway usage accounting and quota behavior are coupled.

## Recommended B then A sequence

### Phase B: Audit and prepare gates

1. Keep current selective worktree paused.
2. Create a new versioned integration worktree for `v0.1.122`.
3. Before merging each version range, record:
   - schema/migration files touched
   - payment/subscription/redeem/quota files touched
   - gateway files touched
   - frontend-v2 files touched
   - rollback point
4. For subscription-sensitive ranges, add or run regression tests for:
   - existing active product subscriptions still authorize API keys
   - product subscription quota still decrements as xlabapi expects
   - payment fulfillment still writes expected product subscription and usage log fields
   - redeem code product subscriptions still map to product/group bindings

### Phase A: Version-by-version merge order

Recommended order:

1. `48912014..v0.1.121`
   - No-op / baseline tag. No merge needed.
2. `v0.1.121..v0.1.122`
   - First actual range. Medium risk. Merge in isolated worktree and resolve conflicts.
3. `v0.1.122..v0.1.123`
   - Small range. Good candidate after v0.1.122 tests pass.
4. `v0.1.123..v0.1.124`
   - High risk due schema/OAuth/risk-control/redeem. Requires explicit approval after reviewing v0.1.122/v0.1.123 result.
5. `v0.1.124..v0.1.125`
   - Medium risk but depends on whether v0.1.124 content is accepted.
6. `v0.1.125..v0.1.126`
   - High subscription/payment risk. Do not merge until product-subscription regression suite is ready.
7. `v0.1.126..v0.1.127` through latest
   - Very high risk. Only proceed after earlier ranges are green and subscription/payment behavior is confirmed.

## First recommended merge target

Start with `v0.1.121..v0.1.122` into a new worktree.

Required gates for that range:

- No `enterprisebff` deletions.
- No unintended frontend-v2 rollback.
- Backend tests for OpenAI gateway/usage and affiliate/admin records.
- Existing xlabapi product subscription tests still pass.
- `frontend-v2` typecheck/build still pass.

If `v0.1.122` produces large conflicts in OpenAI gateway or affiliate schema, stop and turn it into a selective sub-plan rather than forcing a full version merge.
