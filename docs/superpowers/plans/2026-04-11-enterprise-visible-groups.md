# Enterprise Visible Groups Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make enterprise users see only enterprise-configured groups inside `enterprise-bff`, without changing ordinary user visibility rules.

**Architecture:** `enterprise-bff` loads a JSON config mapping enterprise names to visible group IDs, normalizes the enterprise name, and returns only active groups present in that mapping. The core backend and shared user/group semantics remain unchanged.

**Tech Stack:** Go, Gin, ent, testify

---

### Task 1: Lock the new visibility contract with tests

**Files:**
- Modify: `backend/internal/enterprisebff/enterprise_visible_groups_test.go`
- Test: `backend/internal/enterprisebff/enterprise_visible_groups_test.go`

- [ ] **Step 1: Write the failing test**

Add tests that assert:

```go
func TestSelectEnterpriseVisibleGroups_ReturnsOnlyConfiguredGroups(t *testing.T) {}
func TestSelectEnterpriseVisibleGroups_ReturnsEmptyWhenEnterpriseHasNoConfiguredGroups(t *testing.T) {}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/enterprisebff -run 'TestSelectEnterpriseVisibleGroups_' -count=1`
Expected: FAIL because the current implementation still exposes public groups by default.

- [ ] **Step 3: Write minimal implementation**

Change enterprise visibility selection so it uses only enterprise-configured group IDs.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/enterprisebff -run 'TestSelectEnterpriseVisibleGroups_' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/enterprisebff/enterprise.go backend/internal/enterprisebff/enterprise_visible_groups_test.go
git commit -m "fix: restrict enterprise visible groups to configured mappings"
```

### Task 2: Load enterprise visibility mappings from config

**Files:**
- Modify: `backend/internal/enterprisebff/config.go`
- Test: `backend/internal/enterprisebff/config_test.go`

- [ ] **Step 1: Write the failing test**

Add config tests that assert valid JSON is parsed and enterprise names are normalized at lookup time.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/enterprisebff -run 'TestParseEnterpriseVisibleGroupIDsByEnterprise|TestLoadConfig' -count=1`
Expected: FAIL because the config field and parser do not exist yet.

- [ ] **Step 3: Write minimal implementation**

Add `ENTERPRISE_BFF_VISIBLE_GROUP_IDS_BY_ENTERPRISE` parsing to `Config`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/enterprisebff -run 'TestParseEnterpriseVisibleGroupIDsByEnterprise|TestLoadConfig' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/enterprisebff/config.go backend/internal/enterprisebff/config_test.go
git commit -m "feat: load enterprise visible group mappings from config"
```

### Task 3: Verify dependent enterprise handlers still honor the new contract

**Files:**
- Modify: `backend/internal/enterprisebff/pool_status_test.go`
- Modify: `backend/internal/enterprisebff/key_authorization_test.go`

- [ ] **Step 1: Write or adjust tests**

Ensure enterprise pool status and key authorization still filter using enterprise-visible groups only.

- [ ] **Step 2: Run test to verify current behavior**

Run: `go test ./internal/enterprisebff -run 'TestEnterprisePoolStatus|TestAuthorizeRequestedGroup' -count=1`
Expected: PASS after the visibility implementation is updated.

- [ ] **Step 3: Keep implementation minimal**

Only adjust tests if names or expectations changed with the default-deny contract.

- [ ] **Step 4: Run focused package verification**

Run: `go test ./internal/enterprisebff -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/enterprisebff/pool_status_test.go backend/internal/enterprisebff/key_authorization_test.go
git commit -m "test: verify enterprise handlers follow visible group mappings"
```
