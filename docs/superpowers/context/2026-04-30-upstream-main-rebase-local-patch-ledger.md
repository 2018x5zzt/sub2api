# Upstream Main Rebase Local Patch Ledger

**Date:** 2026-04-30

## Classification Legend

- replay: required local behavior must be restored on the upstream-main branch
- superseded: upstream already includes equivalent or stronger behavior
- fuse: upstream and xlabapi both contain partial behavior; implement the union
- preserve migration only: SQL/history must remain compatible, but runtime code may be represented by upstream
- obsolete: no runtime or migration behavior should be carried forward
- documentation: keep only if it documents current integration behavior

## Required Replay Domains

| Domain | Classification | Notes |
| --- | --- | --- |
| OpenAI/Codex/Claude compatibility | replay/fuse | preserve local default instructions, reasoning relay, gpt-5.4/gpt-5.5, image/Sora, Claude/Sonnet behavior |
| shared subscription products | replay/fuse | preserve product schema, billing authorization, usage settlement, frontend views |
| subscription daily carryover | replay | preserve behavior and hide migrated legacy carryover as already patched |
| subscription pro models and image group | replay | preserve seed and group/model behavior |
| available channels and affiliate transition | fuse | channel-first upstream surface plus xlabapi compatibility decisions |
| model plaza compatibility | replay as compatibility | old route remains redirect/alias; primary product becomes Available Channels |
| dynamic group budget multiplier | replay if not upstream-equivalent | preserve local billing behavior |
| enterprise visible groups | replay if still required by product | preserve if current production depends on it |
| token refresh reused-token terminal behavior | replay if not upstream-equivalent | preserve operational safety |
| local migration checksum compatibility | preserve migration only | do not break existing production migrations |

## Required Upstream Domains

| Domain | Upstream source | Integration action |
| --- | --- | --- |
| account bulk edit | 65c27d2c..53b24bc2 | child plan |
| Vertex service accounts | 6d11f9ed, 489a4d93, 93d91e20 | child plan |
| WebSearch and notifications | upstream WebSearch/notify chain | child plan |
| runtime stability | v0.1.119..v0.1.121 stability fixes | child plan |

## Commit Classification Table

Populate this table before replaying code:

| Commit | Subject | Domain | Classification | Destination child plan |
| --- | --- | --- | --- | --- |
| d208624d | fix: add default responses instructions in openai compat | gateway | replay | gateway compatibility |
| bd0837aa | fix: enable subscription pro models and image group | subscription | replay | schema and channels |
| 7426265b | fix(openai): bump codex cli version | gateway | fuse | gateway compatibility |
| 24a570ac | fix: hide migrated legacy subscription carryover | subscription | replay | schema and channels |
| cca30da8 | fix: support sonnet 4.6 thinking model | gateway | replay | gateway compatibility |
