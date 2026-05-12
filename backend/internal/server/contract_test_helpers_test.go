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
	VerifyCodes     map[string]string
	Registration    map[string]bool
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
			VerifyCodes:     map[string]string{},
			Registration:    map[string]bool{},
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

func (s *ContractTestServer) SeedUser(args ...any) int64 {
	panic("TODO: 测试者填充")
}

func (s *ContractTestServer) SeedAPIKey(userID int64) int64 {
	panic("TODO: 测试者填充")
}

func (s *ContractTestServer) SeedAPIKeys(userID int64, count int) []int64 {
	panic("TODO: 测试者填充")
}

func (s *ContractTestServer) SeedRefreshToken(userID int64) string {
	panic("TODO: 测试者填充")
}

func (s *ContractTestServer) Seed2FAChallenge(userID int64) string {
	panic("TODO: 测试者填充")
}

func (s *ContractTestServer) SeedUserProfile(userID int64) {
	panic("TODO: 测试者填充")
}

func (s *ContractTestServer) SeedAvailableGroups(userID int64) {
	panic("TODO: 测试者填充")
}

func (s *ContractTestServer) SeedAvailableGroup(userID, groupID int64) {
	panic("TODO: 测试者填充")
}

func (s *ContractTestServer) SeedGroupRates(userID int64) {
	panic("TODO: 测试者填充")
}

func (s *ContractTestServer) SeedUsageLogs(userID int64) {
	panic("TODO: 测试者填充")
}

func (s *ContractTestServer) SeedUsageStats(userID int64) {
	panic("TODO: 测试者填充")
}

func (s *ContractTestServer) AllowRegistration() {
	panic("TODO: 测试者填充")
}

func (s *ContractTestServer) AcceptVerifyCode(email, code string) {
	panic("TODO: 测试者填充")
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
		if isDynamicIDKey(key) {
			return "<id>"
		}
		return v
	default:
		if isDynamicIDKey(key) && isNumericKind(v) {
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
