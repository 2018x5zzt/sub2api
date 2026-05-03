# xlabapi Miku Iframe SSO Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let `https://xlabapi.com` open `https://ai.mikuapi.org/` as an iframe image-generation workspace with automatic xlabapi-based Miku login, while Miku keeps standalone email/password login and exposes xlabapi login as an optional alternate path.

**Architecture:** xlabapi issues short-lived one-time tickets for authenticated xlabapi users. Miku exchanges each ticket server-to-server, finds or creates a local Miku user, binds the xlabapi identity, and issues its own Miku JWT stored in the existing `localStorage.token` and `localStorage.user` keys. The iframe path uses URL tickets plus Miku-local storage and does not depend on third-party cookies.

**Tech Stack:** xlabapi backend: Go / Gin / Ent / PostgreSQL. xlabapi frontend: Vue 3 / Vite / Pinia / pnpm. Miku backend: Go / Gin / Ent / PostgreSQL / Redis. Miku frontend: Vue 3 / Vite / Pinia / npm. Runtime framing policy is enforced by Gin headers and nginx.

---

## File Structure

### xlabapi Repository: `/root/sub2api-src`

- Create `/root/sub2api-src/backend/ent/schema/miku_sso_ticket.go`: Ent schema for single-use Miku SSO tickets.
- Generate `/root/sub2api-src/backend/ent/mikussoticket*.go` and `/root/sub2api-src/backend/ent/mikussoticket/*`: Ent generated code for the new schema.
- Create `/root/sub2api-src/backend/migrations/145_miku_sso_tickets.sql`: idempotent SQL migration for `miku_sso_tickets`.
- Modify `/root/sub2api-src/backend/internal/config/config.go`: add `MikuSSOConfig` and defaults under `miku_sso`.
- Create `/root/sub2api-src/backend/internal/service/miku_sso_service.go`: ticket generation, hashing, redirect validation, and atomic consume logic.
- Create `/root/sub2api-src/backend/internal/service/miku_sso_service_test.go`: service-level tests for validation, expiry, and single-use behavior.
- Create `/root/sub2api-src/backend/internal/handler/miku_sso_handler.go`: `GET /api/v1/miku-sso/authorize` and `POST /api/v1/miku-sso/token`.
- Create `/root/sub2api-src/backend/internal/handler/miku_sso_handler_test.go`: handler tests for auth requirements and exchange responses.
- Modify `/root/sub2api-src/backend/internal/service/wire.go` and `/root/sub2api-src/backend/internal/handler/wire.go`: add `NewMikuSSOService` and expose it through `AuthHandler`.
- Modify `/root/sub2api-src/backend/internal/server/routes/auth.go`: register the Miku SSO routes.
- Modify `/root/sub2api-src/backend/internal/server/middleware/security_headers.go`: guarantee `frame-src https://ai.mikuapi.org` is allowed when Miku SSO is enabled.
- Create `/root/sub2api-src/frontend/src/api/mikuSso.ts`: API wrapper that requests a JSON Miku iframe URL from xlabapi authorize.
- Modify `/root/sub2api-src/frontend/src/api/index.ts`: export `mikuSsoAPI`.
- Create `/root/sub2api-src/frontend/src/views/user/ImageGenerationView.vue`: dedicated iframe container for Miku.
- Create `/root/sub2api-src/frontend/src/views/user/__tests__/ImageGenerationView.spec.ts`: iframe URL, loading, and fallback tests.
- Modify `/root/sub2api-src/frontend/src/router/index.ts`: add `/image-generation`.
- Modify `/root/sub2api-src/frontend/src/components/layout/AppSidebar.vue`: add the "Online Image Generation" item before `/models`.
- Modify `/root/sub2api-src/frontend/src/i18n/locales/en.ts` and `/root/sub2api-src/frontend/src/i18n/locales/zh.ts`: add navigation and page labels.
- Modify `/root/sub2api-src/deploy/.env.example`: document xlabapi-side Miku SSO env vars.

### Miku Repository: `/root/miku-ai-studio`

- Create `/root/miku-ai-studio/backend/internal/ent/schema/auth_identity.go`: external identity binding table.
- Create `/root/miku-ai-studio/backend/internal/ent/schema/sso_login_session.go`: short-lived frontend consume sessions.
- Modify `/root/miku-ai-studio/backend/internal/ent/schema/user.go`: add edges to auth identities and SSO sessions.
- Generate `/root/miku-ai-studio/backend/internal/ent/authidentity*.go`, `/root/miku-ai-studio/backend/internal/ent/authidentity/*`, `/root/miku-ai-studio/backend/internal/ent/ssologinsession*.go`, and `/root/miku-ai-studio/backend/internal/ent/ssologinsession/*`: Ent generated code.
- Modify `/root/miku-ai-studio/backend/internal/config/config.go`: add `XlabapiSSOConfig` loaded from environment variables.
- Create `/root/miku-ai-studio/backend/internal/service/xlabapi_sso.go`: server-to-server ticket exchange, identity binding, local user creation, username/email conflict handling, and SSO session creation.
- Create `/root/miku-ai-studio/backend/internal/service/xlabapi_sso_test.go`: unit tests for mapping, conflict behavior, binding rules, expiry, and single-use sessions.
- Create `/root/miku-ai-studio/backend/internal/handler/xlabapi_sso.go`: backend start, callback, bind, and consume endpoints.
- Create `/root/miku-ai-studio/backend/internal/handler/xlabapi_sso_test.go`: handler tests for start URL generation, callback, bind, consume, and replay rejection.
- Modify `/root/miku-ai-studio/backend/internal/handler/auth.go`: reuse the existing Miku JWT/user response shape for SSO-issued logins.
- Modify `/root/miku-ai-studio/backend/cmd/server/main.go`: construct SSO service and register routes.
- Create `/root/miku-ai-studio/backend/internal/middleware/frame_ancestors.go`: set `Content-Security-Policy: frame-ancestors https://xlabapi.com` for `/embed/xlabapi`.
- Create `/root/miku-ai-studio/backend/internal/middleware/frame_ancestors_test.go`: header tests for embed route policy.
- Modify `/root/miku-ai-studio/nginx.conf`: keep the backend CSP on embed routes and avoid global headers that block xlabapi framing of `/embed/xlabapi`.
- Modify `/root/miku-ai-studio/deploy/.env.example`: document Miku-side xlabapi SSO env vars.
- Modify `/root/miku-ai-studio/frontend/src/api/index.ts`: add `authAPI.xlabapiCallback` and `authAPI.consumeXlabapiSession`.
- Modify `/root/miku-ai-studio/frontend/src/stores/user.ts`: add `setSession(token, user)` using the existing localStorage keys.
- Modify `/root/miku-ai-studio/frontend/src/router/index.ts`: add `/sso/xlabapi/callback` and `/embed/xlabapi`.
- Modify `/root/miku-ai-studio/frontend/src/views/Login.vue`: add the optional xlabapi login button while keeping password login primary.
- Create `/root/miku-ai-studio/frontend/src/views/XlabapiCallback.vue`: standalone callback consumer.
- Create `/root/miku-ai-studio/frontend/src/views/XlabapiEmbed.vue`: iframe ticket consumer and fallback login shell.
- Create `/root/miku-ai-studio/frontend/src/__tests__/xlabapi-sso.test.ts`: frontend tests for login button, callback, embed success, and embed fallback.

---

### Task 1: xlabapi Backend Ticket Issuer

**Files:**
- Create: `/root/sub2api-src/backend/ent/schema/miku_sso_ticket.go`
- Create: `/root/sub2api-src/backend/migrations/145_miku_sso_tickets.sql`
- Create: `/root/sub2api-src/backend/internal/service/miku_sso_service.go`
- Create: `/root/sub2api-src/backend/internal/service/miku_sso_service_test.go`
- Create: `/root/sub2api-src/backend/internal/handler/miku_sso_handler.go`
- Create: `/root/sub2api-src/backend/internal/handler/miku_sso_handler_test.go`
- Modify: `/root/sub2api-src/backend/internal/config/config.go`
- Modify: `/root/sub2api-src/backend/internal/service/wire.go`
- Modify: `/root/sub2api-src/backend/internal/handler/wire.go`
- Modify: `/root/sub2api-src/backend/internal/server/routes/auth.go`
- Modify: `/root/sub2api-src/backend/internal/server/middleware/security_headers.go`

- [ ] **Step 1: Write the failing service tests**

Add tests in `/root/sub2api-src/backend/internal/service/miku_sso_service_test.go`:

```go
func TestMikuSSOValidateRedirectToAcceptsInternalPath(t *testing.T) {
	got, err := ValidateMikuRedirectTo("/workspace")
	require.NoError(t, err)
	require.Equal(t, "/workspace", got)
}

func TestMikuSSOValidateRedirectToRejectsExternalURL(t *testing.T) {
	_, err := ValidateMikuRedirectTo("https://evil.example/workspace")
	require.Error(t, err)
}

func TestMikuSSOValidateRedirectToRejectsProtocolRelativeURL(t *testing.T) {
	_, err := ValidateMikuRedirectTo("//evil.example/workspace")
	require.Error(t, err)
}

func TestMikuSSOTicketHashIsDeterministicAndDoesNotExposeTicket(t *testing.T) {
	hash1 := HashMikuSSOTicket("ticket-value")
	hash2 := HashMikuSSOTicket("ticket-value")
	require.Equal(t, hash1, hash2)
	require.NotContains(t, hash1, "ticket-value")
	require.Len(t, hash1, 64)
}
```

- [ ] **Step 2: Run the service tests to verify they fail**

Run:

```bash
cd /root/sub2api-src/backend
go test ./internal/service -run 'TestMikuSSO' -v
```

Expected: fail with undefined `ValidateMikuRedirectTo` and `HashMikuSSOTicket`.

- [ ] **Step 3: Add config, schema, migration, and service implementation**

Add `MikuSSOConfig` to `/root/sub2api-src/backend/internal/config/config.go`:

```go
type MikuSSOConfig struct {
	Enabled       bool          `mapstructure:"enabled"`
	ClientID      string        `mapstructure:"client_id"`
	ClientSecret  string        `mapstructure:"client_secret"`
	ProviderKey   string        `mapstructure:"provider_key"`
	FrontendURL   string        `mapstructure:"frontend_url"`
	TicketTTL      time.Duration `mapstructure:"ticket_ttl"`
	AllowedModes  []string      `mapstructure:"allowed_redirect_modes"`
}
```

Wire it into `Config` as:

```go
MikuSSO MikuSSOConfig `mapstructure:"miku_sso"`
```

Then set defaults:

```go
viper.SetDefault("miku_sso.enabled", false)
viper.SetDefault("miku_sso.client_id", "miku-prod")
viper.SetDefault("miku_sso.client_secret", "")
viper.SetDefault("miku_sso.provider_key", "xlabapi-prod")
viper.SetDefault("miku_sso.frontend_url", "https://ai.mikuapi.org")
viper.SetDefault("miku_sso.ticket_ttl", 2*time.Minute)
viper.SetDefault("miku_sso.allowed_redirect_modes", []string{"standalone", "embed"})
```

Create `/root/sub2api-src/backend/ent/schema/miku_sso_ticket.go`:

```go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type MikuSSOTicket struct {
	ent.Schema
}

func (MikuSSOTicket) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "miku_sso_tickets"}}
}

func (MikuSSOTicket) Fields() []ent.Field {
	return []ent.Field{
		field.String("ticket_hash").NotEmpty().Unique().SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Int64("user_id"),
		field.String("intent").MaxLen(20).NotEmpty(),
		field.String("redirect_mode").MaxLen(20).NotEmpty(),
		field.String("redirect_to").NotEmpty().SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("state").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.JSON("claims", map[string]any{}).Default(func() map[string]any { return map[string]any{} }).SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("expires_at").SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("consumed_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("created_at").Default(time.Now).Immutable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (MikuSSOTicket) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ticket_hash").Unique(),
		index.Fields("user_id"),
		index.Fields("expires_at"),
	}
}
```

Create `/root/sub2api-src/backend/migrations/145_miku_sso_tickets.sql`:

```sql
CREATE TABLE IF NOT EXISTS miku_sso_tickets (
  id BIGSERIAL PRIMARY KEY,
  ticket_hash TEXT NOT NULL UNIQUE,
  user_id BIGINT NOT NULL,
  intent VARCHAR(20) NOT NULL,
  redirect_mode VARCHAR(20) NOT NULL,
  redirect_to TEXT NOT NULL,
  state TEXT,
  claims JSONB NOT NULL DEFAULT '{}'::jsonb,
  expires_at TIMESTAMPTZ NOT NULL,
  consumed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_miku_sso_tickets_user_id ON miku_sso_tickets(user_id);
CREATE INDEX IF NOT EXISTS idx_miku_sso_tickets_expires_at ON miku_sso_tickets(expires_at);
```

Create `/root/sub2api-src/backend/internal/service/miku_sso_service.go` with:

```go
const (
	MikuSSOIntentLogin = "login"
	MikuSSOIntentBind  = "bind"
	MikuSSOModeEmbed   = "embed"
	MikuSSOModeStandalone = "standalone"
)

func HashMikuSSOTicket(ticket string) string {
	sum := sha256.Sum256([]byte(ticket))
	return hex.EncodeToString(sum[:])
}

func ValidateMikuRedirectTo(raw string) (string, error) {
	if raw == "" {
		return "/", nil
	}
	if !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") || strings.Contains(raw, "\\") {
		return "", infraerrors.BadRequest("INVALID_REDIRECT_TO", "redirect_to must be an internal Miku path")
	}
	return raw, nil
}
```

Implement service methods:

```go
func (s *MikuSSOService) Authorize(ctx context.Context, userID int64, input MikuSSOAuthorizeInput) (string, error)
func (s *MikuSSOService) Exchange(ctx context.Context, input MikuSSOExchangeInput) (*MikuSSOClaims, error)
```

`Authorize` must create a random 32-byte ticket, store only `HashMikuSSOTicket(ticket)`, expire it at `now + 2 minutes`, and compute:

```text
https://ai.mikuapi.org/embed/xlabapi?ticket=<ticket>&state=<state>
```

for `redirect_mode=embed`, or:

```text
https://ai.mikuapi.org/api/v1/auth/xlabapi/callback?ticket=<ticket>&state=<state>
```

for `redirect_mode=standalone`.

The handler must support both response modes:

```text
GET /api/v1/miku-sso/authorize?...                -> 302 redirect to the computed Miku URL
GET /api/v1/miku-sso/authorize?...&response_mode=json -> 200 {"redirect_url":"<computed Miku URL>"}
```

`Exchange` must authenticate Miku client credentials, load the ticket row with a row lock or an atomic `UPDATE ... WHERE consumed_at IS NULL AND expires_at > NOW()`, set `consumed_at`, and return only server-side claims loaded from the xlabapi user row.

- [ ] **Step 4: Write the failing handler tests**

Add tests in `/root/sub2api-src/backend/internal/handler/miku_sso_handler_test.go`:

```go
func TestMikuSSOAuthorizeRequiresXlabapiLogin(t *testing.T) {
	r := gin.New()
	r.GET("/api/v1/miku-sso/authorize", handler.AuthorizeMikuSSO)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/miku-sso/authorize?intent=login&redirect_mode=embed&redirect_to=/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMikuSSOTokenRejectsMissingClientAuth(t *testing.T) {
	r := gin.New()
	r.POST("/api/v1/miku-sso/token", handler.ExchangeMikuSSOToken)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/miku-sso/token", strings.NewReader(`{"ticket":"abc","redirect_mode":"embed"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}
```

- [ ] **Step 5: Register routes and pass the handler tests**

Register in `/root/sub2api-src/backend/internal/server/routes/auth.go`:

```go
mikuSSO := v1.Group("/miku-sso")
mikuSSO.GET("/authorize", gin.HandlerFunc(jwtAuth), h.Auth.AuthorizeMikuSSO)
mikuSSO.POST("/token", rateLimiter.LimitWithOptions("miku-sso-token", 60, time.Minute, middleware.RateLimitOptions{
	FailureMode: middleware.RateLimitFailClose,
}), h.Auth.ExchangeMikuSSOToken)
```

Run:

```bash
cd /root/sub2api-src/backend
go test ./internal/service ./internal/handler ./internal/server/middleware -run 'TestMikuSSO|TestSecurityHeaders' -v
```

Expected: pass.

- [ ] **Step 6: Generate Ent code and run focused backend tests**

Run:

```bash
cd /root/sub2api-src/backend
go generate ./ent
go test ./internal/service ./internal/handler ./internal/server/middleware -run 'TestMikuSSO|TestSecurityHeaders' -v
```

Expected: pass.

- [ ] **Step 7: Commit xlabapi backend**

Run:

```bash
cd /root/sub2api-src
git add backend/ent/schema/miku_sso_ticket.go backend/ent backend/migrations/145_miku_sso_tickets.sql backend/internal/config/config.go backend/internal/service/miku_sso_service.go backend/internal/service/miku_sso_service_test.go backend/internal/handler/miku_sso_handler.go backend/internal/handler/miku_sso_handler_test.go backend/internal/service/wire.go backend/internal/handler/wire.go backend/internal/server/routes/auth.go backend/internal/server/middleware/security_headers.go
git commit -m "feat: add xlabapi miku sso ticket issuer"
```

### Task 2: Miku Backend Identity Binding and Ticket Consumption

**Files:**
- Create: `/root/miku-ai-studio/backend/internal/ent/schema/auth_identity.go`
- Create: `/root/miku-ai-studio/backend/internal/ent/schema/sso_login_session.go`
- Modify: `/root/miku-ai-studio/backend/internal/ent/schema/user.go`
- Modify: `/root/miku-ai-studio/backend/internal/config/config.go`
- Create: `/root/miku-ai-studio/backend/internal/service/xlabapi_sso.go`
- Create: `/root/miku-ai-studio/backend/internal/service/xlabapi_sso_test.go`
- Create: `/root/miku-ai-studio/backend/internal/handler/xlabapi_sso.go`
- Create: `/root/miku-ai-studio/backend/internal/handler/xlabapi_sso_test.go`
- Modify: `/root/miku-ai-studio/backend/internal/handler/auth.go`
- Create: `/root/miku-ai-studio/backend/internal/middleware/frame_ancestors.go`
- Create: `/root/miku-ai-studio/backend/internal/middleware/frame_ancestors_test.go`
- Modify: `/root/miku-ai-studio/backend/cmd/server/main.go`

- [ ] **Step 1: Write the failing Miku SSO service tests**

Add tests in `/root/miku-ai-studio/backend/internal/service/xlabapi_sso_test.go`:

```go
func TestXlabapiSSOLoginCreatesLocalUserAndIdentity(t *testing.T)
func TestXlabapiSSOLoginReturnsExistingBoundUser(t *testing.T)
func TestXlabapiSSOLoginUsesInternalEmailWhenEmailBelongsToLocalUser(t *testing.T)
func TestXlabapiSSOBindRejectsIdentityBoundToDifferentUser(t *testing.T)
func TestXlabapiSSOConsumeSessionRejectsReplay(t *testing.T)
func TestNormalizeXlabapiUsernameFallsBackToExternalUserID(t *testing.T)
```

The email conflict test must create a local Miku user with `alice@example.com`, consume xlabapi claims with `external_user_id="12345"` and `email="alice@example.com"`, then assert the SSO-created user email is exactly:

```text
xlabapi-12345@internal.miku
```

- [ ] **Step 2: Run the service tests to verify they fail**

Run:

```bash
cd /root/miku-ai-studio/backend
go test ./internal/service -run 'TestXlabapiSSO|TestNormalizeXlabapiUsername' -v
```

Expected: fail because `XlabapiSSOService`, auth identity schema, and SSO session schema do not exist.

- [ ] **Step 3: Add Miku config and Ent schemas**

Add to `/root/miku-ai-studio/backend/internal/config/config.go`:

```go
type XlabapiSSOConfig struct {
	Enabled      bool
	AuthorizeURL string
	TokenURL     string
	ClientID     string
	ClientSecret string
	ProviderKey  string
	EmbedURL     string
	SessionTTLSeconds int
}
```

Add it to `Config` as `XlabapiSSO XlabapiSSOConfig`, and load:

```go
xlabapiSSOEnabled, _ := strconv.ParseBool(getEnv("XLABAPI_SSO_ENABLED", "false"))
xlabapiSSOSessionTTL, _ := strconv.Atoi(getEnv("XLABAPI_SSO_SESSION_TTL_SECONDS", "120"))
```

with values:

```go
XlabapiSSO: XlabapiSSOConfig{
	Enabled: xlabapiSSOEnabled,
	AuthorizeURL: getEnv("XLABAPI_SSO_AUTHORIZE_URL", "https://xlabapi.com/api/v1/miku-sso/authorize"),
	TokenURL: getEnv("XLABAPI_SSO_TOKEN_URL", "https://xlabapi.com/api/v1/miku-sso/token"),
	ClientID: getEnv("XLABAPI_SSO_CLIENT_ID", "miku-prod"),
	ClientSecret: getEnv("XLABAPI_SSO_CLIENT_SECRET", ""),
	ProviderKey: getEnv("XLABAPI_SSO_PROVIDER_KEY", "xlabapi-prod"),
	EmbedURL: getEnv("XLABAPI_SSO_EMBED_URL", "https://ai.mikuapi.org/embed/xlabapi"),
	SessionTTLSeconds: xlabapiSSOSessionTTL,
}
```

Create `/root/miku-ai-studio/backend/internal/ent/schema/auth_identity.go` with fields:

```go
field.Int("user_id"),
field.String("provider_type").MaxLen(20).NotEmpty(),
field.String("provider_key").NotEmpty(),
field.String("provider_subject").NotEmpty(),
field.String("email").Default(""),
field.String("username").Default(""),
field.String("status").Default(""),
field.JSON("metadata", map[string]any{}).Default(func() map[string]any { return map[string]any{} }),
field.Time("verified_at").Optional().Nillable(),
field.Time("last_login_at").Optional().Nillable(),
field.Time("created_at").Default(time.Now).Immutable(),
field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
```

Add indexes:

```go
index.Fields("provider_type", "provider_key", "provider_subject").Unique()
index.Fields("user_id", "provider_type")
```

Create `/root/miku-ai-studio/backend/internal/ent/schema/sso_login_session.go` with fields:

```go
field.String("token_hash").NotEmpty().Unique(),
field.String("provider_type").MaxLen(20).NotEmpty(),
field.String("provider_key").NotEmpty(),
field.String("provider_subject").NotEmpty(),
field.Int("user_id"),
field.String("intent").MaxLen(20).NotEmpty(),
field.String("redirect_to").Default("/"),
field.JSON("claims", map[string]any{}).Default(func() map[string]any { return map[string]any{} }),
field.Time("expires_at"),
field.Time("consumed_at").Optional().Nillable(),
field.Time("created_at").Default(time.Now).Immutable(),
```

Add indexes:

```go
index.Fields("token_hash").Unique()
index.Fields("provider_type", "provider_key", "provider_subject")
index.Fields("expires_at")
```

Modify `User.Edges()` in `/root/miku-ai-studio/backend/internal/ent/schema/user.go` to include:

```go
edge.To("auth_identities", AuthIdentity.Type),
edge.To("sso_login_sessions", SSOLoginSession.Type),
```

- [ ] **Step 4: Implement Miku SSO service**

Create `/root/miku-ai-studio/backend/internal/service/xlabapi_sso.go` with:

```go
const (
	XlabapiProviderType = "xlabapi"
	XlabapiProviderKeyDefault = "xlabapi-prod"
)

type XlabapiClaims struct {
	ProviderKey    string    `json:"provider_key"`
	ExternalUserID string    `json:"external_user_id"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	Role           string    `json:"role"`
	Status         string    `json:"status"`
	SignupSource   string    `json:"signup_source"`
	IssuedAt       time.Time `json:"issued_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	Intent         string    `json:"intent"`
	RedirectTo     string    `json:"redirect_to"`
}
```

Implement:

```go
func (s *XlabapiSSOService) BuildAuthorizeURL(intent string, redirectMode string, redirectTo string, state string) (string, error)
func (s *XlabapiSSOService) ExchangeTicket(ctx context.Context, ticket string, redirectMode string) (*XlabapiClaims, error)
func (s *XlabapiSSOService) LoginOrCreate(ctx context.Context, claims *XlabapiClaims) (*ent.User, error)
func (s *XlabapiSSOService) BindCurrentUser(ctx context.Context, userID int, claims *XlabapiClaims) (*ent.AuthIdentity, error)
func (s *XlabapiSSOService) CreateConsumeSession(ctx context.Context, claims *XlabapiClaims, userID int) (string, error)
func (s *XlabapiSSOService) ConsumeSession(ctx context.Context, session string) (*ent.User, string, error)
```

Rules:

```text
provider_type = "xlabapi"
provider_key = claims.ProviderKey, defaulting to "xlabapi-prod"
provider_subject = claims.ExternalUserID
only status == "active" can log in
never copy xlabapi password
never reuse xlabapi ID as Miku user ID
never grant Miku admin from xlabapi role
never silently merge by email
```

Username generation must use:

```text
claims.Username
email local part
xlabapi_<external_user_id>
```

and then normalize to `[A-Za-z0-9_]`, truncate to leave suffix room, and suffix `_2`, `_3`, `_4` until unique.

- [ ] **Step 5: Run Miku service tests to verify they pass**

Run:

```bash
cd /root/miku-ai-studio/backend
go generate ./internal/ent
go test ./internal/service -run 'TestXlabapiSSO|TestNormalizeXlabapiUsername' -v
```

Expected: pass.

- [ ] **Step 6: Write and implement handler and frame policy tests**

Add handler tests in `/root/miku-ai-studio/backend/internal/handler/xlabapi_sso_test.go`:

```go
func TestXlabapiStartRedirectsToXlabapiAuthorize(t *testing.T)
func TestXlabapiCallbackRejectsMissingTicket(t *testing.T)
func TestXlabapiEmbedCallbackReturnsSessionJSON(t *testing.T)
func TestXlabapiConsumeSessionReturnsMikuTokenAndUser(t *testing.T)
func TestXlabapiConsumeSessionRejectsReplay(t *testing.T)
func TestXlabapiBindCallbackRequiresMikuJWT(t *testing.T)
```

Add frame policy test in `/root/miku-ai-studio/backend/internal/middleware/frame_ancestors_test.go`:

```go
func TestXlabapiEmbedFrameAncestorsAllowsOnlyXlabapi(t *testing.T) {
	r := gin.New()
	r.Use(XlabapiEmbedFrameAncestors("https://xlabapi.com"))
	r.GET("/embed/xlabapi", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	req := httptest.NewRequest(http.MethodGet, "/embed/xlabapi", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Contains(t, w.Header().Get("Content-Security-Policy"), "frame-ancestors https://xlabapi.com")
}
```

Run:

```bash
cd /root/miku-ai-studio/backend
go test ./internal/handler ./internal/middleware -run 'TestXlabapi|Test.*FrameAncestors' -v
```

Expected before implementation: fail.

Implement routes in `/root/miku-ai-studio/backend/cmd/server/main.go`:

```go
api.GET("/auth/xlabapi/start", authHandler.StartXlabapiLogin)
api.GET("/auth/xlabapi/callback", authHandler.XlabapiCallback)
api.POST("/auth/xlabapi/consume-session", authRateLimit, authHandler.ConsumeXlabapiSession)

auth.GET("/auth/xlabapi/bind/start", authHandler.StartXlabapiBind)
auth.GET("/auth/xlabapi/bind/callback", authHandler.XlabapiBindCallback)
```

`StartXlabapiLogin` must redirect to:

```text
https://xlabapi.com/api/v1/miku-sso/authorize?intent=login&redirect_mode=standalone&redirect_to=<internal-path>&state=<random>
```

`XlabapiCallback` must support both browser callback modes:

```text
GET /api/v1/auth/xlabapi/callback?ticket=<ticket>&state=<state>
  -> 302 /sso/xlabapi/callback?session=<session>

GET /api/v1/auth/xlabapi/callback?ticket=<ticket>&state=<state>&redirect_mode=embed&response_mode=json
  -> 200 {"session":"<session>"}
```

Register embed route before SPA fallback handling if present:

```go
r.GET("/embed/xlabapi", middleware.XlabapiEmbedFrameAncestors("https://xlabapi.com"), func(c *gin.Context) {
	c.File("frontend/dist/index.html")
})
```

Run the handler and middleware tests again. Expected: pass.

- [ ] **Step 7: Commit Miku backend**

Run:

```bash
cd /root/miku-ai-studio
git add backend/internal/ent/schema/auth_identity.go backend/internal/ent/schema/sso_login_session.go backend/internal/ent/schema/user.go backend/internal/ent backend/internal/config/config.go backend/internal/service/xlabapi_sso.go backend/internal/service/xlabapi_sso_test.go backend/internal/handler/xlabapi_sso.go backend/internal/handler/xlabapi_sso_test.go backend/internal/handler/auth.go backend/internal/middleware/frame_ancestors.go backend/internal/middleware/frame_ancestors_test.go backend/cmd/server/main.go
git commit -m "feat: add miku xlabapi sso backend"
```

### Task 3: Miku Frontend Login, Callback, and Embed Mode

**Files:**
- Modify: `/root/miku-ai-studio/frontend/src/api/index.ts`
- Modify: `/root/miku-ai-studio/frontend/src/stores/user.ts`
- Modify: `/root/miku-ai-studio/frontend/src/router/index.ts`
- Modify: `/root/miku-ai-studio/frontend/src/views/Login.vue`
- Create: `/root/miku-ai-studio/frontend/src/views/XlabapiCallback.vue`
- Create: `/root/miku-ai-studio/frontend/src/views/XlabapiEmbed.vue`
- Create: `/root/miku-ai-studio/frontend/src/__tests__/xlabapi-sso.test.ts`

- [ ] **Step 1: Write the failing frontend tests**

Create `/root/miku-ai-studio/frontend/src/__tests__/xlabapi-sso.test.ts`:

```ts
import { mount, flushPromises } from '@vue/test-utils'
import { describe, expect, it, vi, beforeEach } from 'vitest'
import Login from '../views/Login.vue'
import XlabapiCallback from '../views/XlabapiCallback.vue'
import XlabapiEmbed from '../views/XlabapiEmbed.vue'

describe('Miku xlabapi SSO frontend', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.restoreAllMocks()
  })

  it('keeps password login primary and renders xlabapi as an alternate login action', () => {
    const wrapper = mount(Login, { global: { stubs: { RouterLink: true, SurfaceCard: true, ActionButton: true } } })
    expect(wrapper.find('input[type="email"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('通过 xlabapi 登录')
  })

  it('callback consumes session and writes the existing Miku localStorage keys', async () => {
    const wrapper = mount(XlabapiCallback, {
      global: { mocks: { $route: { query: { session: 'session-token' } } } },
    })
    await flushPromises()
    expect(localStorage.getItem('token')).toBe('miku-jwt')
    expect(JSON.parse(localStorage.getItem('user') || '{}').email).toBe('alice@example.com')
    expect(wrapper.text()).toContain('登录成功')
  })

  it('embed mode renders fallback login when ticket exchange fails', async () => {
    const wrapper = mount(XlabapiEmbed, {
      global: { mocks: { $route: { query: { ticket: 'bad-ticket' } } } },
    })
    await flushPromises()
    expect(wrapper.text()).toContain('登录工作区')
    expect(wrapper.text()).toContain('通过 xlabapi 登录')
  })
})
```

- [ ] **Step 2: Run the frontend tests to verify they fail**

Run:

```bash
cd /root/miku-ai-studio
npm --prefix frontend test -- xlabapi-sso
```

Expected: fail because the new routes/views/API methods do not exist.

- [ ] **Step 3: Add API methods and store session setter**

Modify `/root/miku-ai-studio/frontend/src/api/index.ts`:

```ts
xlabapiCallback: (params: { ticket: string; state?: string; redirect_mode?: 'standalone' | 'embed' }) =>
  api.get('/auth/xlabapi/callback', { params }),
consumeXlabapiSession: (data: { session: string }) =>
  api.post('/auth/xlabapi/consume-session', data),
```

Modify `/root/miku-ai-studio/frontend/src/stores/user.ts`:

```ts
function setSession(nextToken: string, nextUser: any) {
  token.value = nextToken
  user.value = nextUser
  localStorage.setItem('token', nextToken)
  localStorage.setItem('user', JSON.stringify(nextUser))
}
```

Return `setSession` from the store.

- [ ] **Step 4: Add routes and views**

Modify `/root/miku-ai-studio/frontend/src/router/index.ts`:

```ts
{
  path: '/sso/xlabapi/callback',
  name: 'XlabapiCallback',
  component: () => import('../views/XlabapiCallback.vue'),
  meta: { chrome: false },
},
{
  path: '/embed/xlabapi',
  name: 'XlabapiEmbed',
  component: () => import('../views/XlabapiEmbed.vue'),
  meta: { chrome: false },
},
```

Create `/root/miku-ai-studio/frontend/src/views/XlabapiCallback.vue` with logic:

```ts
const session = String(route.query.session || '')
const res: any = await authAPI.consumeXlabapiSession({ session })
userStore.setSession(res.data.token, res.data.user)
router.replace(res.data.redirect_to || '/')
```

Create `/root/miku-ai-studio/frontend/src/views/XlabapiEmbed.vue` with logic:

```ts
const ticket = String(route.query.ticket || '')
const callback: any = await authAPI.xlabapiCallback({ ticket, state, redirect_mode: 'embed', response_mode: 'json' })
const consumed: any = await authAPI.consumeXlabapiSession({ session: callback.data.session })
userStore.setSession(consumed.data.token, consumed.data.user)
router.replace(consumed.data.redirect_to || '/')
```

On failure, render the same login form behavior as `/login` and keep the xlabapi alternate button visible.

Modify `/root/miku-ai-studio/frontend/src/views/Login.vue` to add:

```vue
<button
  type="button"
  class="auth-card__xlabapi-button"
  data-test="xlabapi-login"
  @click="startXlabapiLogin"
>
  通过 xlabapi 登录
</button>
```

`startXlabapiLogin` should redirect the browser to the Miku backend start endpoint:

```text
/api/v1/auth/xlabapi/start?intent=login&redirect_to=/
```

- [ ] **Step 5: Run Miku frontend tests**

Run:

```bash
cd /root/miku-ai-studio
npm --prefix frontend test -- xlabapi-sso
npm --prefix frontend test -- auth router
```

Expected: pass.

- [ ] **Step 6: Commit Miku frontend**

Run:

```bash
cd /root/miku-ai-studio
git add frontend/src/api/index.ts frontend/src/stores/user.ts frontend/src/router/index.ts frontend/src/views/Login.vue frontend/src/views/XlabapiCallback.vue frontend/src/views/XlabapiEmbed.vue frontend/src/__tests__/xlabapi-sso.test.ts
git commit -m "feat: add miku xlabapi sso frontend"
```

### Task 4: xlabapi Frontend Image Generation Iframe Entry

**Files:**
- Create: `/root/sub2api-src/frontend/src/api/mikuSso.ts`
- Modify: `/root/sub2api-src/frontend/src/api/index.ts`
- Create: `/root/sub2api-src/frontend/src/views/user/ImageGenerationView.vue`
- Create: `/root/sub2api-src/frontend/src/views/user/__tests__/ImageGenerationView.spec.ts`
- Modify: `/root/sub2api-src/frontend/src/router/index.ts`
- Modify: `/root/sub2api-src/frontend/src/components/layout/AppSidebar.vue`
- Modify: `/root/sub2api-src/frontend/src/i18n/locales/en.ts`
- Modify: `/root/sub2api-src/frontend/src/i18n/locales/zh.ts`

- [ ] **Step 1: Write the failing xlabapi frontend tests**

Create `/root/sub2api-src/frontend/src/views/user/__tests__/ImageGenerationView.spec.ts`:

```ts
import { mount, flushPromises } from '@vue/test-utils'
import { describe, expect, it, vi, beforeEach } from 'vitest'
import ImageGenerationView from '../ImageGenerationView.vue'

const authorize = vi.fn()

vi.mock('@/api/mikuSso', () => ({
  mikuSsoAPI: { authorize },
}))

describe('ImageGenerationView', () => {
  beforeEach(() => {
    authorize.mockReset()
  })

  it('loads the Miku embed URL from xlabapi authorize flow', async () => {
    authorize.mockResolvedValue({ data: { redirect_url: 'https://ai.mikuapi.org/embed/xlabapi?ticket=abc&state=s1' } })
    const wrapper = mount(ImageGenerationView, { global: { stubs: { AppLayout: { template: '<div><slot /></div>' }, Icon: true } } })
    await flushPromises()
    const iframe = wrapper.find('iframe')
    expect(iframe.attributes('src')).toBe('https://ai.mikuapi.org/embed/xlabapi?ticket=abc&state=s1')
  })

  it('shows a new-tab fallback link for the same Miku embed URL', async () => {
    authorize.mockResolvedValue({ data: { redirect_url: 'https://ai.mikuapi.org/embed/xlabapi?ticket=abc&state=s1' } })
    const wrapper = mount(ImageGenerationView, { global: { stubs: { AppLayout: { template: '<div><slot /></div>' }, Icon: true } } })
    await flushPromises()
    expect(wrapper.find('a[target="_blank"]').attributes('href')).toBe('https://ai.mikuapi.org/embed/xlabapi?ticket=abc&state=s1')
  })
})
```

Add an assertion to `/root/sub2api-src/frontend/src/components/layout/__tests__/AppSidebar.spec.ts`:

```ts
describe('AppSidebar image generation entry', () => {
  it('contains a dedicated image-generation route', () => {
    expect(componentSource).toContain("path: '/image-generation'")
    expect(componentSource).toContain("t('nav.imageGeneration')")
  })
})
```

- [ ] **Step 2: Run the xlabapi frontend tests to verify they fail**

Run:

```bash
cd /root/sub2api-src
pnpm -C frontend test:run ImageGenerationView AppSidebar
```

Expected: fail because the API wrapper, page, nav item, and route do not exist.

- [ ] **Step 3: Add API wrapper and iframe view**

Create `/root/sub2api-src/frontend/src/api/mikuSso.ts`:

```ts
import { apiClient } from './client'

export const mikuSsoAPI = {
  authorize() {
    return apiClient.get('/miku-sso/authorize', {
      params: {
        intent: 'login',
        redirect_mode: 'embed',
        redirect_to: '/workspace',
        response_mode: 'json',
      },
    })
  },
}
```

Modify `/root/sub2api-src/frontend/src/api/index.ts`:

```ts
export { mikuSsoAPI } from './mikuSso'
```

Create `/root/sub2api-src/frontend/src/views/user/ImageGenerationView.vue` with:

```vue
<template>
  <AppLayout>
    <div class="image-generation-page">
      <div v-if="loading" class="image-generation-state">{{ t('imageGeneration.loading') }}</div>
      <div v-else-if="error" class="image-generation-state">
        <p>{{ t('imageGeneration.loadFailed') }}</p>
        <button class="btn btn-primary" type="button" @click="loadEmbedUrl">{{ t('common.retry') }}</button>
      </div>
      <div v-else class="image-generation-shell">
        <a :href="embedUrl" target="_blank" rel="noopener noreferrer" class="btn btn-secondary btn-sm image-generation-open">
          {{ t('imageGeneration.openInNewTab') }}
        </a>
        <iframe :src="embedUrl" class="image-generation-frame" allow="clipboard-write; fullscreen" />
      </div>
    </div>
  </AppLayout>
</template>
```

Script behavior:

```ts
const embedUrl = ref('')
const loading = ref(false)
const error = ref('')

async function loadEmbedUrl() {
  loading.value = true
  error.value = ''
  try {
    const res: any = await mikuSsoAPI.authorize()
    embedUrl.value = res.data.redirect_url
  } catch (err: any) {
    error.value = err?.message || 'failed'
  } finally {
    loading.value = false
  }
}

onMounted(loadEmbedUrl)
```

Use a full-height unframed layout:

```css
.image-generation-page { height: calc(100vh - 64px - 4rem); }
.image-generation-shell { position: relative; height: 100%; width: 100%; overflow: hidden; }
.image-generation-frame { display: block; width: 100%; height: 100%; border: 0; background: transparent; }
.image-generation-open { position: absolute; right: 12px; top: 12px; z-index: 10; }
```

- [ ] **Step 4: Add route, nav item, and translations**

Modify `/root/sub2api-src/frontend/src/router/index.ts`:

```ts
{
  path: '/image-generation',
  name: 'ImageGeneration',
  component: () => import('@/views/user/ImageGenerationView.vue'),
  meta: {
    requiresAuth: true,
    requiresAdmin: false,
    title: 'Image Generation',
    titleKey: 'imageGeneration.title',
    descriptionKey: 'imageGeneration.description',
  },
},
```

Modify `buildSelfNavItems` in `/root/sub2api-src/frontend/src/components/layout/AppSidebar.vue` so the item order starts:

```ts
items.push(
  { path: '/keys', label: t('nav.apiKeys'), icon: KeyIcon },
  { path: '/image-generation', label: t('nav.imageGeneration'), icon: ChannelIcon },
  { path: '/models', label: t('nav.modelHub'), icon: ChannelIcon },
```

Also add it in simple admin mode next to `/models`:

```ts
filtered.push({ path: '/image-generation', label: t('nav.imageGeneration'), icon: ChannelIcon })
filtered.push({ path: '/models', label: t('nav.modelHub'), icon: ChannelIcon })
```

Add translations:

```ts
nav: {
  imageGeneration: 'Online Image Generation',
}
imageGeneration: {
  title: 'Online Image Generation',
  description: 'Open Miku image generation in xlabapi.',
  loading: 'Loading Miku workspace...',
  loadFailed: 'Failed to load Miku workspace.',
  openInNewTab: 'Open in new tab',
}
```

and Chinese:

```ts
nav: {
  imageGeneration: '在线生图',
}
imageGeneration: {
  title: '在线生图',
  description: '在 xlabapi 中打开 Miku 生图工作区。',
  loading: '正在加载 Miku 工作区...',
  loadFailed: 'Miku 工作区加载失败。',
  openInNewTab: '新窗口打开',
}
```

- [ ] **Step 5: Run xlabapi frontend tests**

Run:

```bash
cd /root/sub2api-src
pnpm -C frontend test:run ImageGenerationView AppSidebar
pnpm -C frontend typecheck
```

Expected: pass.

- [ ] **Step 6: Commit xlabapi frontend**

Run:

```bash
cd /root/sub2api-src
git add frontend/src/api/mikuSso.ts frontend/src/api/index.ts frontend/src/views/user/ImageGenerationView.vue frontend/src/views/user/__tests__/ImageGenerationView.spec.ts frontend/src/router/index.ts frontend/src/components/layout/AppSidebar.vue frontend/src/components/layout/__tests__/AppSidebar.spec.ts frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts
git commit -m "feat: open miku image generation from xlabapi"
```

### Task 5: Deployment, CSP, and End-to-End Verification

**Files:**
- Modify: `/root/sub2api-src/deploy/.env.example`
- Modify: `/root/miku-ai-studio/deploy/.env.example`
- Modify: `/root/miku-ai-studio/nginx.conf`
- Modify: `/root/sub2api-src/backend/internal/server/middleware/security_headers.go`
- Modify: `/root/miku-ai-studio/backend/internal/middleware/frame_ancestors.go`

- [ ] **Step 1: Add deployment env examples**

Add to `/root/sub2api-src/deploy/.env.example`:

```text
MIKU_SSO_ENABLED=true
MIKU_SSO_CLIENT_ID=miku-prod
MIKU_SSO_CLIENT_SECRET=change-me
MIKU_SSO_PROVIDER_KEY=xlabapi-prod
MIKU_SSO_FRONTEND_URL=https://ai.mikuapi.org
MIKU_SSO_ALLOWED_REDIRECT_MODES=standalone,embed
```

Add to `/root/miku-ai-studio/deploy/.env.example`:

```text
XLABAPI_SSO_ENABLED=true
XLABAPI_SSO_AUTHORIZE_URL=https://xlabapi.com/api/v1/miku-sso/authorize
XLABAPI_SSO_TOKEN_URL=https://xlabapi.com/api/v1/miku-sso/token
XLABAPI_SSO_CLIENT_ID=miku-prod
XLABAPI_SSO_CLIENT_SECRET=change-me
XLABAPI_SSO_PROVIDER_KEY=xlabapi-prod
XLABAPI_SSO_EMBED_URL=https://ai.mikuapi.org/embed/xlabapi
XLABAPI_SSO_SESSION_TTL_SECONDS=120
```

- [ ] **Step 2: Verify CSP and iframe headers**

Run:

```bash
cd /root/sub2api-src/backend
go test ./internal/server/middleware -run 'TestSecurityHeaders' -v
```

Expected: pass and include `frame-src https://ai.mikuapi.org` when Miku SSO is enabled.

Run:

```bash
cd /root/miku-ai-studio/backend
go test ./internal/middleware -run 'Test.*FrameAncestors' -v
```

Expected: pass and include `frame-ancestors https://xlabapi.com` for `/embed/xlabapi`.

- [ ] **Step 3: Run backend suites for touched packages**

Run:

```bash
cd /root/sub2api-src/backend
go test ./internal/service ./internal/handler ./internal/server/middleware ./internal/config -run 'TestMikuSSO|TestSecurityHeaders|TestConfig' -v
```

Run:

```bash
cd /root/miku-ai-studio/backend
go test ./internal/service ./internal/handler ./internal/middleware ./internal/config -run 'TestXlabapi|Test.*FrameAncestors|TestConfig' -v
```

Expected: pass.

- [ ] **Step 4: Run frontend suites for touched pages**

Run:

```bash
cd /root/sub2api-src
pnpm -C frontend test:run ImageGenerationView AppSidebar
pnpm -C frontend typecheck
```

Run:

```bash
cd /root/miku-ai-studio
npm --prefix frontend test -- xlabapi-sso auth router
npm --prefix frontend run build
```

Expected: pass.

- [ ] **Step 5: Manual production-domain smoke test**

With both apps deployed to their real domains, verify these exact flows:

```text
1. Visit https://ai.mikuapi.org/ directly.
2. Confirm Miku email/password login remains the primary form.
3. Click "通过 xlabapi 登录".
4. Confirm the browser goes through https://xlabapi.com/api/v1/miku-sso/authorize and returns to Miku logged in.
5. Visit https://xlabapi.com/ and log in.
6. Click "在线生图".
7. Confirm the page renders an iframe whose src begins with https://ai.mikuapi.org/embed/xlabapi?ticket=.
8. Confirm Miku inside the iframe is logged in without a password prompt.
9. Refresh the iframe URL with the same ticket.
10. Confirm ticket replay fails and the fallback Miku login form appears.
11. Temporarily set a test ticket expiry to the past.
12. Confirm expired ticket exchange fails.
```

- [ ] **Step 6: Commit deployment wiring**

Run:

```bash
cd /root/sub2api-src
git add deploy/.env.example backend/internal/server/middleware/security_headers.go
git commit -m "chore: document xlabapi miku sso deployment"
```

Run:

```bash
cd /root/miku-ai-studio
git add deploy/.env.example nginx.conf backend/internal/middleware/frame_ancestors.go
git commit -m "chore: document miku xlabapi sso deployment"
```

---

## Self-Review Notes

- Spec coverage: the plan covers one-time xlabapi tickets, server-side Miku ticket exchange, Miku-local JWT issuance, standalone Miku xlabapi login, xlabapi iframe entry, Miku embed fallback, identity binding, email conflict handling, no password copying, no role/balance/concurrency mapping, and cross-domain iframe CSP.
- Intent rules: `intent=login` creates or returns the bound Miku user; `intent=bind` requires current Miku authentication and rejects cross-user binding conflicts.
- Cookie dependency: the iframe flow uses a URL ticket and Miku `localStorage` on `ai.mikuapi.org`, so it does not depend on third-party cookies.
- Domain policy: xlabapi frames only `https://ai.mikuapi.org`; Miku embed routes allow frame ancestors only from `https://xlabapi.com`.
- Deferred scope: balance exchange, role synchronization, user-facing unbind, and same-email merge remain out of scope.
