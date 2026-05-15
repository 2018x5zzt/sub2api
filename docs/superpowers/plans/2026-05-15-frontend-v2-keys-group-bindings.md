# frontend-v2 Keys Group Binding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore API key group selection/editing in the frontend-v2 `/keys` page, and require an explicit budget multiplier whenever a key is bound to a dynamic pricing group.

**Architecture:** Keep the existing React page and the current backend API contract. Add a reusable group selector section to the key create/edit modal, source options from the existing user-visible groups API, and compute the dynamic-budget requirement from the selected group metadata. Preserve the app's current card/modal/input/button styling so the new controls feel native to frontend-v2.

**Tech Stack:** React 18, TypeScript, TanStack Query, existing frontend-v2 UI primitives, existing `/api/keys` and `/groups/available` contracts.

---

### Task 1: Add a failing Keys page test for group binding and dynamic budget validation

**Files:**
- Create: `frontend-v2/src/pages/user/__tests__/Keys.spec.tsx`
- Modify: `frontend-v2/src/api/keys.ts` (only if test imports need explicit exports)

- [ ] **Step 1: Write the failing test**

```tsx
it('allows selecting a group when creating a key and requires budget when the target group is dynamic', async () => {
  // render page with one fixed group and one dynamic group
  // open create modal
  // select the dynamic group
  // expect a budget input to appear
  // submit without entering budget -> expect validation error and no create call
  // enter a budget and submit -> expect createKey called with group_id and budget_multiplier
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npm test -- frontend-v2/src/pages/user/__tests__/Keys.spec.tsx -t "allows selecting a group"`
Expected: FAIL because the current page has no group selector / no dynamic-budget validation.

- [ ] **Step 3: Do not implement production code yet**

Keep the failure as the proof that the feature is missing.

### Task 2: Implement the group selector and dynamic budget flow in the Keys page

**Files:**
- Modify: `frontend-v2/src/pages/user/Keys.tsx`
- Modify: `frontend-v2/src/api/keys.ts`
- Modify: `frontend-v2/src/api/models.ts` if needed to expose group fetching from the page
- Modify: `frontend-v2/src/types/index.ts` only if the page needs a narrow local type update for group metadata

- [ ] **Step 1: Write minimal implementation**

```tsx
// Use modelsAPI.getUserGroups() (or the existing user-visible groups API) to load groups.
// Add a reusable group field inside the modal that works for both create and edit.
// When a fixed group is selected, clear budget_multiplier.
// When a dynamic group is selected, show a budget multiplier input prefilled from the group's default_budget_multiplier (fallback 8) and keep it required.
// On submit, block save if the selected group is dynamic and budget_multiplier is missing.
// Send both group_id and budget_multiplier to create/update only when the group is dynamic.
```

- [ ] **Step 2: Run the focused page test**

Run: `npm test -- frontend-v2/src/pages/user/__tests__/Keys.spec.tsx`
Expected: PASS.

- [ ] **Step 3: Verify visuals stay consistent**

Check the modal uses the existing `card`, `input`, `btn`, and native `select` styling patterns already used in frontend-v2 admin pages.

### Task 3: Add regression coverage for editing an existing key's group

**Files:**
- Modify: `frontend-v2/src/pages/user/__tests__/Keys.spec.tsx`

- [ ] **Step 1: Write the failing test**

```tsx
it('updates an existing key when switching its group', async () => {
  // render with one existing key bound to a fixed group
  // open edit modal
  // switch the group to another available group
  // save
  // expect updateKey called with the new group_id
})
```

- [ ] **Step 2: Run the test to verify it fails before code changes**

Run: `npm test -- frontend-v2/src/pages/user/__tests__/Keys.spec.tsx -t "updates an existing key"`
Expected: FAIL until edit-mode group handling is in place.

- [ ] **Step 3: Run the whole relevant frontend-v2 test file and typecheck**

Run: `npm test -- frontend-v2/src/pages/user/__tests__/Keys.spec.tsx && npm run typecheck`
Expected: PASS with no type errors.

### Task 4: Build and smoke test the frontend-v2 bundle

**Files:**
- None expected unless the build exposes a missing import/type

- [ ] **Step 1: Build frontend-v2**

Run: `npm run build`
Expected: PASS.

- [ ] **Step 2: Report any follow-up fixes if the build reveals hidden regressions**

Only fix what the build/test run proves is broken.
