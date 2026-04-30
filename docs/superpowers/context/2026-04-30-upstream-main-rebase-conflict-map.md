# Upstream Main Rebase Conflict Map

**Date:** 2026-04-30

## High-Risk Conflict Areas

| Area | Conflict type | Resolution owner |
| --- | --- | --- |
| Ent generated code | content conflicts across `backend/ent/*` | schema/migration phase; regenerate after schema decisions |
| Migrations | overlapping numbered migrations and local production history | schema/migration phase; preserve local high-water mark and add compatibility migrations |
| Account admin handlers | content conflicts and upstream bulk edit changes | account bulk edit phase |
| Channels and available channels | add/add and content conflicts | channels/model-surface phase |
| Affiliate and invite | add/add and content conflicts | schema/migration plus channel phase; preserve xlabapi affiliate-only invite retirement semantics |
| Settings DTO and public settings | content conflicts | WebSearch/notifications and channels phases |
| Gateway/OpenAI/Codex/Claude | content conflicts | gateway compatibility phase |
| Sora/image paths | upstream removals vs xlabapi modifications | gateway compatibility phase; preserve xlabapi compatibility where still used |
| Frontend account/settings/channel pages | content conflicts | child plan owned by related feature area |

## Raw Conflict Capture

The raw dry-merge output was generated with:

```bash
git -C /root/sub2api-src merge-tree --write-tree --name-only xlabapi upstream/main > /tmp/xlabapi-upstream-main-merge-tree.txt
```

The command exits non-zero when conflicts are found. The raw output is intentionally not committed because it is machine output and may change when either source ref advances.
