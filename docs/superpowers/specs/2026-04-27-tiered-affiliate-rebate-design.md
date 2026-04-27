# Tiered Affiliate Rebate Design

Date: 2026-04-27
Status: Approved in conversation, written for implementation planning

## Summary

This design defines the final XlabAPI affiliate rebate rules for paid invite outcomes and registration-count concurrency rewards.

The product has two separate invitation benefits:

1. Affiliate rebates are based on invitees with recharge records. Internally, an invitee becomes effective after redeeming a commercial sales redeem code. User-facing copy calls this "invited users with recharge records".
2. Concurrency rewards are based on raw invited registrations. They do not require paid redemption.

Affiliate rebates are credited as balance through the existing affiliate quota flow. Rebates are not cash payouts and are not retroactively recalculated when an inviter reaches a higher tier.

## Product Language

Use user-facing language that avoids exposing internal redeem-code implementation details.

Preferred copy:

- "邀请返利"
- "当上包工头，Token 不用愁"
- "有充值记录的邀请人数"
- "最高 20% 返利，以余额形式到账"
- "不同商品类型返利系数有所不同：余额充值、日卡 100%，周卡 60%，月卡 30%；以实际到账金额为准。"
- "普通邀请也能提升并发：邀请 1 人注册并发提升到 5，邀请 5 人注册并发提升到 10。"
- "如果你没有正在使用 API 的朋友，也可以邀请小号注册，先解锁更高并发。"

Avoid user-facing wording such as:

- "商业售卖类型兑换码"
- "source_type=commercial"
- "SKU coefficient"

## Frontend Product Positioning

The invite center should make the affiliate system feel valuable and easy to understand, not like an internal rebate ledger.

Product naming:

- Product name and navigation label: "邀请返利"
- Ladder-system slogan: "当上包工头，Token 不用愁"

The first screen should communicate three ideas immediately:

1. Highest rebate: "最高 20%"
2. Rebate form: "以余额形式到账"
3. Two growth tracks:
   - invite users with recharge records to climb the rebate ladder
   - invite registered users to unlock higher concurrency

Suggested frontend structure:

- Keep the page title/navigation as "邀请返利".
- Add a compact tier-ladder hero/header with the slogan "当上包工头，Token 不用愁", showing current tier, current effective invitee count, and next tier progress.
- A clear tier table showing Bronze through Diamond ranges and nominal rates.
- A "返利怎么算" section with one simple formula:

```text
返利余额 = 商品总额度 × 等级比例 × 商品系数
```

- A product factor explainer:
  - 余额充值 / 日卡: 100%
  - 周卡: 60%
  - 月卡: 30%
- A separate "并发加速" section for registration invites:
  - 默认并发 3
  - 邀请 1 人注册，并发提升到 5
  - 邀请 5 人注册，并发提升到 10
- Friendly helper copy:
  - "如果你没有用 API 的朋友，也可以邀请小号注册，先解锁更高并发。"

The UI must not imply that small-account registration creates rebate eligibility. It only unlocks the registration-count concurrency benefit.

## Effective Invitee Definition

An invitee counts as one effective invitee for rebate tiering when all conditions are true:

- the invitee is bound to the inviter through the affiliate relationship
- the invitee redeems a qualifying commercial sales redeem code
- the redeem code source is commercial
- the redeem code type is either balance or subscription

Each invitee counts at most once toward the inviter's effective invitee count.

Benefit, compensation, system-grant, admin-adjustment, and registration-only events do not count as effective invitees for rebate tiers.

## Rebate Tiers

The rebate tier is determined by the inviter's effective invitee count at the time a qualifying redemption is settled.

| Tier | User-facing name | Effective invitees | Nominal rebate rate |
| --- | --- | ---: | ---: |
| Bronze | 青铜 | 1-2 | 5% |
| Silver | 白银 | 3-9 | 8% |
| Gold | 黄金 | 10-29 | 12% |
| Platinum | 铂金 | 30-49 | 15% |
| Diamond | 钻石 | 50+ | 20% |

There is no rebate tier for 0 effective invitees. If no tier matches, no tier rebate is applied unless an explicit per-user override or global fallback is configured for operational compatibility.

## Non-Retroactive Tiering

Tier upgrades are not retroactive.

When an inviter reaches a higher tier, only the current and future qualifying redemptions use the higher rate. Previously settled rebates keep their original rate and amount.

Example:

- 1st effective invitee redemption uses Bronze 5%.
- 2nd effective invitee redemption uses Bronze 5%.
- 3rd effective invitee redemption uses Silver 8%.
- The first two redemptions remain settled at 5%; the system does not补差价.

The settlement path should preserve this by calculating the effective invitee count after the current commercial redemption has been marked as used, inside the same database transaction.

## Rebate Formula

The final rebate formula is:

```text
rebate_balance = rebate_base_quota * tier_rebate_rate * sku_rebate_factor
```

Where:

- `rebate_base_quota` is the quota amount used as the rebate base.
- `tier_rebate_rate` is the inviter's nominal tier percentage as a decimal.
- `sku_rebate_factor` is the product-type coefficient.

The rebate is credited to the inviter's affiliate quota, then transferred to user balance through the existing affiliate quota transfer flow.

## Rebate Base

Balance and subscription products use different rebate bases.

| Product type | Rebate base |
| --- | --- |
| Balance recharge | Balance amount credited by the commercial balance redeem code |
| Daily card | Total quota included in the daily card |
| Weekly card | Total quota included in the weekly card |
| Monthly card | Total quota included in the monthly card |

For subscription cards, the rebate base is not the sale price. It is the total card quota.

Example:

- Monthly card sale price: 225
- Monthly card total quota: 6750
- Inviter tier: Silver 8%
- Monthly factor: 0.3
- Rebate: `6750 * 0.08 * 0.3 = 162`

The same principle applies to daily and weekly cards. Daily cards use total quota with factor 1.0; weekly cards use total quota with factor 0.6.

## SKU Rebate Factors

Use the following final factors. Earlier draft factors such as weekly 0.7 and monthly 0.4 are intentionally not used.

| Product type | Factor |
| --- | ---: |
| Balance recharge | 1.0 |
| Daily card | 1.0 |
| Weekly card | 0.6 |
| Monthly card | 0.3 |

Equivalent actual rebate rates:

| Tier | Nominal rate | Balance / daily | Weekly | Monthly |
| --- | ---: | ---: | ---: | ---: |
| Bronze | 5% | 5.0% | 3.0% | 1.5% |
| Silver | 8% | 8.0% | 4.8% | 2.4% |
| Gold | 12% | 12.0% | 7.2% | 3.6% |
| Platinum | 15% | 15.0% | 9.0% | 4.5% |
| Diamond | 20% | 20.0% | 12.0% | 6.0% |

## Subscription Card Classification

Subscription redeem codes must resolve to a card cadence before rebate settlement:

- daily card
- weekly card
- monthly card

The implementation should classify product redeem codes from product metadata or validity configuration, using the existing subscription product model where possible.

Recommended fallback:

- validity of 1 day maps to daily
- validity of 7 days maps to weekly
- validity of 30 days maps to monthly

If a subscription redeem code cannot be classified, it should not receive a subscription rebate until classification is fixed. This avoids accidentally treating thin-margin cards as factor 1.0.

## Registration Concurrency Rewards

Registration-count rewards are separate from effective invitee rebates.

Rules:

| Invited registered users | User concurrency |
| ---: | ---: |
| default | 3 |
| 1+ | 5 |
| 5+ | 10 |

These rewards are based on invited user registration count, not recharge count. A user can increase concurrency through registrations even if invitees never recharge.

This rule can be shown directly in the invite center. The copy may explicitly say that users without API-using friends can invite small accounts to unlock higher concurrency, as long as the copy does not imply fake recharge or rebate eligibility.

The implementation should make this idempotent:

- applying the rule should never reduce a higher existing concurrency granted by admin or another product mechanism
- repeated calls should not keep adding concurrency
- cache invalidation should follow the existing concurrency update path

## Existing System Fit

Current code already has several useful foundations:

- `user_affiliates.inviter_id` stores inviter binding
- `aff_count` stores raw invited registration count
- `CountEffectiveInvitees` counts commercial balance/subscription redeemers
- `RedeemService.Redeem` marks the redeem code used before affiliate settlement in the same transaction
- affiliate quota accrual uses `user_affiliate_ledger`
- affiliate quota transfer credits user balance
- admin settings already support configurable tier JSON

Implementation should keep those boundaries and avoid introducing a second affiliate ledger.

## Implementation Direction

1. Set the default affiliate tier table to the approved thresholds and rates.
2. Preserve admin override support for per-user rebate rates.
3. Extend affiliate accrual input from a single numeric base amount to a structured rebate event containing product type, rebate base quota, and factor.
4. For balance redeem codes, use the code value as rebate base and factor 1.0.
5. For subscription product redeem codes, resolve total quota and cadence, then apply daily 1.0, weekly 0.6, or monthly 0.3.
6. Keep settlement transactional and non-retroactive.
7. Add registration-count concurrency reward handling after inviter binding or registration completion.
8. Update user-facing copy and admin copy to match the approved language.
9. Update the invite center UI so the product remains "邀请返利", while the tier ladder uses the slogan "当上包工头，Token 不用愁" and shows rebate ladder value, balance payout, SKU factors, and the separate concurrency unlock track.

## Testing Strategy

Backend tests should cover:

- tier thresholds: 1, 2, 3, 9, 10, 29, 30, 49, 50
- non-retroactive behavior by verifying the 3rd effective invitee receives Silver while earlier settled ledger rows remain unchanged
- balance rebate base uses balance amount with factor 1.0
- daily card uses total quota with factor 1.0
- weekly card uses total quota with factor 0.6
- monthly card uses total quota with factor 0.3
- monthly 225 / total quota 6750 / Silver 8% settles rebate 162
- benefit/system-grant redeem codes do not count and do not accrue affiliate rebate
- raw invited registration count upgrades concurrency to 5 at 1 invite and 10 at 5 invites
- concurrency reward is idempotent and never reduces a higher existing concurrency

Frontend tests should cover:

- tier table displays Bronze through Diamond with the approved ranges and rates
- invite center keeps the product name "邀请返利"
- tier ladder section presents the slogan "当上包工头，Token 不用愁"
- hero copy communicates "最高 20%" and "以余额形式到账"
- rule copy says "有充值记录的邀请人数"
- rule copy exposes product factors as balance/daily 100%, weekly 60%, monthly 30%
- concurrency copy clearly separates registration invites from recharge-record invitees
- small-account copy is visible but does not imply rebate eligibility
- no user-facing copy exposes internal source type or redeem-code implementation wording

## Rollout Notes

The change affects money-like balance accrual. Deploy with a rollback overlay and verify:

- public settings still load
- affiliate center loads for a normal user
- a commercial balance redemption accrues the expected amount
- a commercial monthly-card redemption uses total quota and the 0.3 factor
- recent logs have no affiliate settlement errors
