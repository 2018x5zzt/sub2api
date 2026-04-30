# Upstream Main Rebase Baseline

**Date:** 2026-04-30
**Integration branch:** integrate/xlabapi-upstream-main-rebase-20260430
**Integration root:** upstream/main

## Source Refs

| Ref | Commit | Role |
| --- | --- | --- |
| upstream/main | 48912014a16e2dd1cfca8b7cad785d0e8e7bfeec | new baseline |
| xlabapi | 0ea4541d14d6f0d825331a3c9f34ae36e56b0091 | local behavior source |
| origin/xlabapi | cca30da8c7f7314469e58cef8d17ddcd38442684 | published xlabapi baseline before local docs |
| merge-base | 6a2cf09ee05ff4833c93592f6c68cf21415febde | common ancestor |

## Divergence

| Comparison | Local-side commits | Upstream-side commits |
| --- | ---: | ---: |
| xlabapi...upstream/main | 247 | 770 |

## Required Outcomes

- Use upstream/main as the integration root.
- Preserve xlabapi production migration compatibility.
- Preserve OpenAI/Codex/Claude compatibility.
- Migrate model-plaza behavior toward channel-first Available Channels with compatibility routes.
- Include account bulk edit, Vertex, WebSearch/notifications, and runtime stability.

## Verification Gate

This branch cannot replace xlabapi until all child plans complete their targeted tests and final verification passes.
