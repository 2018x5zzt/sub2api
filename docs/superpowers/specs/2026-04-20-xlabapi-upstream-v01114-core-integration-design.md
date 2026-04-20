# Xlabapi Upstream v0.1.114 Core Integration Design

**Date:** 2026-04-20

**Goal:** Produce one clean integration branch that keeps `xlabapi` behavior intact while absorbing the already-prepared `merge/upstream-v01114-core` runtime fixes on top of the latest `origin/xlabapi`.

## Summary

This design does not attempt to fast-forward `xlabapi` all the way to the current `upstream/main`.

Instead, it narrows the integration target to the smallest meaningful next step that already has repository support:

- start from the latest fetched `origin/xlabapi`
- preserve the two newer remote `xlabapi` commits for `gpt-5.4 xhigh relay`
- absorb the prepared `merge/upstream-v01114-core` branch
- resolve conflicts without dropping `xlabapi` custom behavior
- verify the integrated branch as a stable handoff point for later upstream work

This creates a controlled checkpoint between the historical `v0.1.106`-based `xlabapi` lineage and the much newer `upstream/main` lineage.

## Problem

`xlabapi` is not currently aligned with the real upstream tip.

After refreshing remotes on 2026-04-20:

- local `xlabapi` is behind `origin/xlabapi` by 2 commits
- `origin/xlabapi` is ahead of `upstream/main` by 153 commits and behind by 405 commits
- `origin/main` is also far behind `upstream/main`
- the common ancestor between `origin/xlabapi` and `upstream/main` remains `v0.1.106`

Trying to absorb all 405 upstream-only commits in one step would mix three different concerns:

1. bring local `xlabapi` up to the current remote `xlabapi`
2. validate the repository's existing `merge/upstream-v01114-core` integration work
3. continue the much larger rebase or merge path from `v0.1.114` toward the current upstream tip

This design separates those concerns so the first integration checkpoint can be validated cleanly.

## Goals

- Use the latest fetched `origin/xlabapi` as the business baseline.
- Absorb the prepared `merge/upstream-v01114-core` branch into a new clean integration branch.
- Preserve the two `origin/xlabapi` commits that landed after local `xlabapi`:
  - `7cdaf9a0 fix(openai): normalize gpt-5.4 reasoning relay aliases`
  - `04424606 merge: gpt54 xhigh relay fix into xlabapi`
- Preserve `xlabapi`-specific behavior unless the incoming runtime fix intentionally supersedes it.
- Keep the user's current dirty working tree untouched.
- Produce a verified integration point that can later be merged back to `xlabapi` or used as the base for the next upstream chase.

## Non-Goals

- Do not absorb the remaining 405 upstream-only commits in this phase.
- Do not redesign `xlabapi` product behavior while resolving merge conflicts.
- Do not clean up unrelated local uncommitted changes in the current working tree.
- Do not silently rewrite branch history on `xlabapi`, `origin/xlabapi`, or `merge/upstream-v01114-core`.
- Do not deploy anything in this phase.

## Current Branch Topology

The relevant references are:

- `origin/xlabapi`
  - latest remote `xlabapi` business baseline
  - includes the `gpt-5.4 xhigh relay` merge that local `xlabapi` has not yet fast-forwarded to
- `merge/upstream-v01114-core`
  - existing repository branch intended to port upstream `v0.1.114` core runtime fixes into the `xlabapi` line
  - relative to `origin/xlabapi`, it introduces two branch-only commits:
    - `a2c36d14 fix(frontend): hoist edit modal TLS profile loader`
    - `e34f8f3a merge: port upstream v0.1.114 core runtime fixes`
- current working tree on local `xlabapi`
  - contains four uncommitted modifications
  - these changes must remain untouched during the integration work

This means the practical integration target is not abstract. It is:

`origin/xlabapi` + branch-only work from `merge/upstream-v01114-core`

with explicit protection for the remote `xlabapi` relay fixes.

## Integration Boundary

This phase is complete when the integration branch includes:

- all commits already reachable from `origin/xlabapi`
- the runtime-fix content currently reachable only from `merge/upstream-v01114-core`
- conflict resolutions needed to make those histories coexist

This phase is not responsible for anything that exists only in newer `upstream/main` history after `v0.1.114`.

That boundary is intentional. It keeps the proof obligation small enough to reason about.

## Isolation Strategy

Integration work must happen in an isolated git worktree.

Reasons:

- the current `/root/sub2api-src` working tree is dirty
- the user explicitly wants those uncommitted changes preserved
- branch integration and verification should start from a clean filesystem state

The isolated worktree should:

- be created from the same repository
- start from a clean branch based on `origin/xlabapi`
- perform the merge or cherry-pick work there
- leave the current working tree unchanged

## Merge Strategy

The recommended integration sequence is:

1. create a clean branch from `origin/xlabapi`
2. bring in `merge/upstream-v01114-core`
3. resolve conflicts with minimal behavior-preserving edits
4. inspect the resulting diff against both parents
5. run targeted verification before declaring the branch usable

The preferred integration shape is a normal merge commit if that preserves history cleanly.

If the existing merge branch contains noisy or misleading merge ancestry that makes conflict handling harder than necessary, cherry-picking the two branch-only commits is acceptable, but only if:

- the resulting code matches the intended `merge/upstream-v01114-core` content
- the final branch state is easier to audit than a forced merge
- the decision is documented during execution

## Conflict Resolution Policy

Conflict resolution must follow these rules:

1. preserve remote `xlabapi` behavior by default
2. keep the intent of upstream runtime fixes where they are clearly additive or bug-fixing
3. avoid opportunistic refactors
4. avoid mixing in the current uncommitted working-tree edits
5. prefer the smallest coherent diff that restores both sides' intended behavior

In practical terms, the most sensitive surfaces are:

- OpenAI gateway request and response normalization
- usage accounting and scheduler snapshot logic
- upstream response size handling
- frontend account modal behavior

These areas deserve explicit post-merge inspection even if the merge applies cleanly.

## Verification Design

Verification in this phase must prove branch coherence, not full product completeness.

The verification bar has three layers.

### 1. Baseline integrity

Confirm the worktree starts clean and reflects the intended integration base:

- branch starts from `origin/xlabapi`
- no carry-over from the user's dirty local changes

### 2. Integration intent

Confirm the merged branch actually contains the expected `merge/upstream-v01114-core` deltas:

- scheduler snapshot and usage-accounting changes are present
- upstream response limit support is present
- account modal TLS profile loader fix is present

### 3. Regression protection

Confirm the integrated branch does not lose known `xlabapi` protections:

- `gpt-5.4 xhigh relay` normalization stays intact
- existing `xlabapi` gateway behavior does not regress in the touched code paths
- the merged code builds and its targeted tests pass

The exact commands belong in the implementation plan, but verification must be fresh and executable from the clean integration worktree.

## Deliverable

The deliverable for this phase is one clean, verified integration branch.

That branch must be suitable for one of two next steps:

1. merge back into `xlabapi` after review
2. continue as the base for the next upstream catch-up phase

The choice between those two paths is intentionally deferred until the branch is verified.

## Risks

### Risk: the existing `merge/upstream-v01114-core` branch partially overlaps with newer `origin/xlabapi` relay fixes

Mitigation:

- base the work on `origin/xlabapi`, not local `xlabapi`
- inspect merge results specifically around OpenAI gateway reasoning and relay handling

### Risk: branch history suggests a clean merge, but the final code silently drops an `xlabapi` customization

Mitigation:

- inspect the post-merge diff in touched hotspot files
- run targeted tests for the gateway and usage paths

### Risk: current dirty local changes leak into the integration work

Mitigation:

- use a separate worktree from the start
- do not perform the merge in `/root/sub2api-src`

### Risk: this checkpoint creates false confidence that upstream synchronization is complete

Mitigation:

- explicitly document that this phase stops at the prepared `v0.1.114` integration checkpoint
- keep the later 405-upstream-commit chase as a separate phase

## Success Criteria

This design is complete only when all of the following are true:

- a clean integration branch exists from `origin/xlabapi`
- the branch absorbs the branch-only content from `merge/upstream-v01114-core`
- the branch still contains the `origin/xlabapi` `gpt-5.4 xhigh relay` fixes
- the user's current dirty working tree remains unchanged
- targeted verification for the touched integration surfaces passes
- the branch is ready either for merge-back review or for the next upstream catch-up phase

## Next Phase

If this phase succeeds, the next design and plan can choose between:

1. landing this verified integration checkpoint back onto `xlabapi`
2. using the verified checkpoint as the new base to continue absorbing newer `upstream/main` history

That decision should be made only after reviewing the actual integrated diff and fresh verification results.
