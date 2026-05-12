package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const contractGoldenRoot = "testdata/contracts/api_v1"

// ContractTestServer is the shared HTTP contract-test harness for /api/v1.
// The router and stub containers are intentionally minimal in this scaffold;
// C001-C014 should fill the route wiring and repo behavior without changing
// business code.
type ContractTestServer struct {
	Router *gin.Engine
	Stubs  *ContractStubRepos
}

// ContractStubRepos is a named placeholder for repo/service stubs that C001-C014
// will grow as each contract test is implemented. The generic maps give the
// next worker a stable home for seeded state without choosing concrete repo
// interfaces in this scaffold.
type ContractStubRepos struct {
	Users           map[int64]any
	APIKeys         map[int64]any
	RefreshTokens   map[string]any
	TwoFAChallenges map[string]any
	AvailableGroups map[int64]any
	GroupRates      map[int64]any
	UsageLogs       map[int64][]any
	UsageStats      map[int64]*usagestats.DashboardStats
	DashboardStats  *usagestats.DashboardStats
	VerifyCodes     map[string]string
	Registration    map[string]bool
	nextUserID      int64
	nextAPIKeyID    int64
}

func newContractTestServer(t *testing.T) *ContractTestServer {
	t.Helper()
	gin.SetMode(gin.TestMode)

	return &ContractTestServer{
		Router: gin.New(),
		Stubs: &ContractStubRepos{
			Users:           map[int64]any{},
			APIKeys:         map[int64]any{},
			RefreshTokens:   map[string]any{},
			TwoFAChallenges: map[string]any{},
			AvailableGroups: map[int64]any{},
			GroupRates:      map[int64]any{},
			UsageLogs:       map[int64][]any{},
			UsageStats:      map[int64]*usagestats.DashboardStats{},
			VerifyCodes:     map[string]string{},
			Registration:    map[string]bool{},
			nextUserID:      1,
			nextAPIKeyID:    1,
		},
	}
}

func doJSON(srv *ContractTestServer, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	if srv == nil || srv.Router == nil {
		panic("nil ContractTestServer or Router")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	srv.Router.ServeHTTP(rec, req)
	return rec
}

func assertGolden(t *testing.T, goldenName string, actual []byte) {
	t.Helper()

	goldenPath := filepath.Join(contractGoldenRoot, goldenName)
	expected, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "read golden %s", goldenPath)

	normalizedExpected := normalizeContractJSON(t, expected)
	normalizedActual := normalizeContractJSON(t, actual)
	if !bytes.Equal(normalizedExpected, normalizedActual) {
		t.Fatalf("golden mismatch for %s\n%s", goldenPath, unifiedLineDiff(string(normalizedExpected), string(normalizedActual)))
	}
}

func authHeader(userID int64) map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer contract-user-%d", userID),
	}
}

func adminHeader() map[string]string {
	return map[string]string{
		"Authorization": "Bearer contract-admin",
		"X-Admin":       "true",
	}
}

// SeedUser stores a deterministic in-memory user fixture and returns its ID.
// Optional args are intentionally loose to match the contract-test plan:
// args[0] email string, args[1] password string.
func (s *ContractTestServer) SeedUser(args ...any) int64 {
	if s == nil || s.Stubs == nil {
		panic("nil ContractTestServer or Stubs")
	}

	id := s.Stubs.nextUserID
	s.Stubs.nextUserID++

	email := fmt.Sprintf("contract-user-%d@example.com", id)
	if len(args) > 0 {
		if value, ok := args[0].(string); ok && value != "" {
			email = value
		}
	}
	password := "pass"
	if len(args) > 1 {
		if value, ok := args[1].(string); ok && value != "" {
			password = value
		}
	}

	s.Stubs.Users[id] = map[string]any{
		"id":       id,
		"email":    email,
		"password": password,
		"role":     "user",
		"status":   "active",
	}
	return id
}

// SeedAPIKey stores a deterministic in-memory API key fixture for userID.
func (s *ContractTestServer) SeedAPIKey(userID int64) int64 {
	if s == nil || s.Stubs == nil {
		panic("nil ContractTestServer or Stubs")
	}

	id := s.Stubs.nextAPIKeyID
	s.Stubs.nextAPIKeyID++
	s.Stubs.APIKeys[id] = map[string]any{
		"id":      id,
		"user_id": userID,
		"key":     fmt.Sprintf("sk-contract-%d", id),
		"status":  "active",
	}
	return id
}

// SeedAPIKeys stores count deterministic in-memory API key fixtures for userID.
func (s *ContractTestServer) SeedAPIKeys(userID int64, count int) []int64 {
	ids := make([]int64, 0, count)
	for i := 0; i < count; i++ {
		ids = append(ids, s.SeedAPIKey(userID))
	}
	return ids
}

// SeedRefreshToken stores a deterministic in-memory refresh token fixture.
func (s *ContractTestServer) SeedRefreshToken(userID int64) string {
	token := fmt.Sprintf("rt_contract_user_%d", userID)
	s.Stubs.RefreshTokens[token] = map[string]any{
		"user_id": userID,
		"token":   token,
	}
	return token
}

// Seed2FAChallenge stores a deterministic in-memory two-factor challenge.
func (s *ContractTestServer) Seed2FAChallenge(userID int64) string {
	challengeID := fmt.Sprintf("2fa-contract-user-%d", userID)
	s.Stubs.TwoFAChallenges[challengeID] = map[string]any{
		"user_id": userID,
		"id":      challengeID,
	}
	return challengeID
}

// SeedUserProfile marks that profile fixtures exist for userID.
func (s *ContractTestServer) SeedUserProfile(userID int64) {
	s.Stubs.Users[userID] = mergeContractFixture(s.Stubs.Users[userID], map[string]any{
		"profile_seeded": true,
	})
}

// SeedAvailableGroups stores default available group fixtures for userID.
func (s *ContractTestServer) SeedAvailableGroups(userID int64) {
	s.SeedAvailableGroup(userID, 1)
	s.SeedAvailableGroup(userID, 2)
}

// SeedAvailableGroup stores one available group fixture for userID.
func (s *ContractTestServer) SeedAvailableGroup(userID, groupID int64) {
	s.Stubs.AvailableGroups[groupID] = map[string]any{
		"id":      groupID,
		"user_id": userID,
		"name":    fmt.Sprintf("contract-group-%d", groupID),
	}
}

// SeedGroupRates stores default group-rate fixtures for userID.
func (s *ContractTestServer) SeedGroupRates(userID int64) {
	s.Stubs.GroupRates[userID] = []map[string]any{
		{"group_id": int64(1), "rate": 1.0},
		{"group_id": int64(2), "rate": 1.5},
	}
}

// SeedUsageLogs stores deterministic usage-log fixtures for userID.
func (s *ContractTestServer) SeedUsageLogs(userID int64) {
	s.Stubs.UsageLogs[userID] = []any{
		map[string]any{
			"id":            int64(1),
			"user_id":       userID,
			"model":         "gpt-4.1-mini",
			"total_tokens":  int64(450),
			"actual_cost":   0.45,
			"duration_ms":   120,
			"created_at":    "2025-01-02T03:04:05Z",
			"request_count": int64(1),
		},
	}
}

// SeedUsageStats stores deterministic dashboard-stat fixtures for userID and
// exposes the same aggregate as the default admin dashboard stats fixture.
func (s *ContractTestServer) SeedUsageStats(userID int64) {
	stats := &usagestats.DashboardStats{
		TotalUsers:               3,
		TodayNewUsers:            1,
		ActiveUsers:              2,
		HourlyActiveUsers:        1,
		StatsUpdatedAt:           "2025-01-02T03:04:05Z",
		StatsStale:               false,
		TotalAPIKeys:             5,
		ActiveAPIKeys:            4,
		TotalAccounts:            6,
		NormalAccounts:           4,
		ErrorAccounts:            1,
		RateLimitAccounts:        1,
		OverloadAccounts:         0,
		TotalRequests:            120,
		TotalInputTokens:         1000,
		TotalOutputTokens:        2000,
		TotalCacheCreationTokens: 300,
		TotalCacheReadTokens:     400,
		TotalTokens:              3700,
		TotalCost:                12.34,
		TotalActualCost:          10.5,
		TodayRequests:            12,
		TodayInputTokens:         100,
		TodayOutputTokens:        200,
		TodayCacheCreationTokens: 30,
		TodayCacheReadTokens:     40,
		TodayTokens:              370,
		TodayCost:                1.23,
		TodayActualCost:          1.05,
		AverageDurationMs:        98.7,
		Rpm:                      7,
		Tpm:                      210,
	}
	s.Stubs.UsageStats[userID] = stats
	s.Stubs.DashboardStats = stats
}

// AllowRegistration marks registration as enabled for contract tests.
func (s *ContractTestServer) AllowRegistration() {
	s.Stubs.Registration["enabled"] = true
}

// AcceptVerifyCode stores the accepted verification code for email.
func (s *ContractTestServer) AcceptVerifyCode(email, code string) {
	s.Stubs.VerifyCodes[email] = code
}

func mergeContractFixture(existing any, updates map[string]any) map[string]any {
	merged := map[string]any{}
	if current, ok := existing.(map[string]any); ok {
		for key, value := range current {
			merged[key] = value
		}
	}
	for key, value := range updates {
		merged[key] = value
	}
	return merged
}

func normalizeContractJSON(t *testing.T, raw []byte) []byte {
	t.Helper()

	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		trimmed := bytes.TrimSpace(raw)
		return append([]byte(nil), trimmed...)
	}

	normalized := normalizeContractValue(value, "")
	out, err := json.MarshalIndent(normalized, "", "  ")
	require.NoError(t, err)
	return append(out, '\n')
}

func normalizeContractValue(value any, key string) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for childKey, childValue := range v {
			out[childKey] = normalizeContractValue(childValue, childKey)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, childValue := range v {
			out[i] = normalizeContractValue(childValue, key)
		}
		return out
	case string:
		return normalizeContractString(key, v)
	case json.Number:
		if isDynamicIDKey(key) || isDynamicRuntimeMetricKey(key) {
			return "<id>"
		}
		return v
	default:
		if (isDynamicIDKey(key) || isDynamicRuntimeMetricKey(key)) && isNumericKind(v) {
			return "<id>"
		}
		return v
	}
}

func normalizeContractString(key, value string) any {
	if isDynamicTimeKey(key) || looksLikeTime(value) {
		return "<time>"
	}
	if isDynamicIDKey(key) && value != "" {
		return "<id>"
	}
	if isDynamicTokenKey(key) || looksLikeToken(value) {
		return normalizeTokenPrefix(value)
	}
	return value
}

func isDynamicTimeKey(key string) bool {
	k := strings.ToLower(key)
	return k == "created_at" || k == "updated_at" || k == "expires_at" || k == "paid_at" || k == "completed_at" ||
		strings.HasSuffix(k, "_at") || strings.HasSuffix(k, "_time") || strings.HasSuffix(k, "_date") || strings.HasSuffix(k, "_start") || strings.HasSuffix(k, "_end")
}

func looksLikeTime(value string) bool {
	if value == "" {
		return false
	}
	if _, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return true
	}
	if _, err := time.Parse("2006-01-02", value); err == nil {
		return true
	}
	return false
}

func isDynamicIDKey(key string) bool {
	k := strings.ToLower(key)
	return k == "id" || strings.HasSuffix(k, "_id") || strings.HasSuffix(k, "_ids")
}

func isDynamicRuntimeMetricKey(key string) bool {
	return strings.ToLower(key) == "uptime"
}

func isDynamicTokenKey(key string) bool {
	k := strings.ToLower(key)
	return strings.Contains(k, "token") || strings.Contains(k, "secret") || k == "key" || strings.HasSuffix(k, "_key")
}

func looksLikeToken(value string) bool {
	return strings.HasPrefix(value, "sk-") || strings.HasPrefix(value, "sk_") || strings.HasPrefix(value, "eyJ") || strings.HasPrefix(value, "tok_") || strings.HasPrefix(value, "rt_")
}

func normalizeTokenPrefix(value string) string {
	for _, sep := range []string{"_", "-", "."} {
		if idx := strings.Index(value, sep); idx > 0 {
			return value[:idx+1] + "<token>"
		}
	}
	return "<token>"
}

func isNumericKind(value any) bool {
	if value == nil {
		return false
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func unifiedLineDiff(expected, actual string) string {
	expectedLines := strings.Split(strings.TrimSuffix(expected, "\n"), "\n")
	actualLines := strings.Split(strings.TrimSuffix(actual, "\n"), "\n")
	maxLen := len(expectedLines)
	if len(actualLines) > maxLen {
		maxLen = len(actualLines)
	}

	var out []string
	out = append(out, "--- golden", "+++ actual")
	for i := 0; i < maxLen; i++ {
		var e, a string
		if i < len(expectedLines) {
			e = expectedLines[i]
		}
		if i < len(actualLines) {
			a = actualLines[i]
		}
		if e == a {
			continue
		}
		out = append(out, fmt.Sprintf("@@ line %d @@", i+1))
		if i < len(expectedLines) {
			out = append(out, "-"+e)
		}
		if i < len(actualLines) {
			out = append(out, "+"+a)
		}
	}
	sort.Strings(out[2:])
	return strings.Join(out, "\n")
}
