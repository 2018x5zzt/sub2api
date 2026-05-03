# xlabapi and Miku SSO Design

Date: 2026-05-03

## Summary

This design adds xlabapi as an upstream identity provider for Miku AI Studio.
The first release covers two entry paths:

- xlabapi opens Miku inside an iframe for the "Image generation" entry.
- Miku keeps its normal standalone username/password login, with xlabapi login
  as an optional alternate path.

The first release only covers direct login and permanent account binding.
Balance exchange from xlabapi to Miku credits is intentionally out of scope and
will be added later on top of the same identity mapping.

The user-facing goals are:

- Miku login page shows a "Use xlabapi to log in" button.
- xlabapi's "Image generation" entry opens Miku in an iframe and auto-signs
  the user in with xlabapi identity.
- Miku personal center shows an xlabapi binding section.
- A Miku user can bind one xlabapi account.
- Once an xlabapi account is bound, users cannot unbind it themselves.

## Current Context

xlabapi is based on Sub2API. Its user model is richer than Miku's user model:

- `users.id` is `int64`.
- `email` is present and soft-delete uniqueness is handled through partial indexes.
- `username` can be empty and allows up to 100 characters.
- `password_hash` exists but must not be copied to Miku.
- `role`, `status`, `balance`, `concurrency`, `rpm_limit`, `signup_source`,
  `last_login_at`, and `last_active_at` exist.
- xlabapi already has canonical identity tables: `auth_identities` and
  `auth_identity_channels`.

Miku currently has a local user model:

- `users.id` is `int`.
- `email` is unique and required.
- `username` is unique, required, and limited to 50 characters.
- `password_hash` is required.
- `avatar`, `balance`, `vip_level`, `rpm_limit`, `concurrency_limit`, `role`,
  and `status` exist.
- Miku's JWT is issued by Miku and stored by the frontend in `localStorage.token`
  and `localStorage.user`.

The systems have incompatible account schemas, so Miku must not reuse xlabapi
user IDs as local user IDs and must not silently merge users by email.

## Decisions

### Identity Key

Miku will treat xlabapi as an external identity provider. The canonical binding
key is:

```text
provider_type = "xlabapi"
provider_key = "xlabapi-prod"
provider_subject = string(xlabapi.users.id)
```

`provider_subject` must use the xlabapi user ID, not the email address. Email is
only a display and audit snapshot.

### Passwords

Miku will never copy or validate xlabapi passwords. xlabapi remains responsible
for authenticating the user. Miku only consumes a short-lived one-time ticket
and then issues its own Miku JWT.

For SSO-created users, Miku will generate a non-user-facing random password hash
to satisfy the current `users.password_hash` requirement.

### Role Mapping

xlabapi admins do not automatically become Miku admins. All SSO-created users
default to Miku `role = "user"` unless a future explicit admin-side mapping is
added.

### Balance Mapping

No balance or credit values are mapped in this release. xlabapi `balance` and
Miku `balance` remain independent. Balance exchange will be implemented later as
a separate explicit feature.

### Permanent Binding

User-facing unbind is not supported. If an incorrect binding must be repaired,
it is an administrator-only remediation operation outside this release.

## Data Model

### Miku `auth_identities`

Add an external identity table to Miku:

```text
auth_identities
- id
- user_id
- provider_type
- provider_key
- provider_subject
- email
- username
- status
- metadata
- verified_at
- last_login_at
- created_at
- updated_at
```

Constraints and indexes:

```text
unique(provider_type, provider_key, provider_subject)
index(user_id, provider_type)
```

Recommended field types:

- `provider_type`: string, max 20.
- `provider_key`: string.
- `provider_subject`: string.
- `email`: string.
- `username`: string.
- `status`: string.
- `metadata`: JSON object.
- `verified_at`: nullable timestamp.
- `last_login_at`: nullable timestamp.

### Miku SSO Sessions

Add a short-lived SSO session table:

```text
sso_login_sessions
- id
- token_hash
- provider_type
- provider_key
- provider_subject
- user_id
- intent
- redirect_to
- claims
- expires_at
- consumed_at
- created_at
```

Constraints and indexes:

```text
unique(token_hash)
index(provider_type, provider_key, provider_subject)
index(expires_at)
```

Allowed `intent` values:

```text
login
bind
```

### xlabapi SSO Tickets

Add a short-lived SSO ticket table to xlabapi:

```text
miku_sso_tickets
- id
- ticket_hash
- user_id
- intent
- redirect_mode
- redirect_to
- state
- claims
- expires_at
- consumed_at
- created_at
```

Constraints and indexes:

```text
unique(ticket_hash)
index(user_id)
index(expires_at)
```

Tickets expire after 2 minutes and are single-use.

## Field Mapping

| xlabapi field | Miku destination | Rule |
| --- | --- | --- |
| `users.id` | `auth_identities.provider_subject` | Required canonical binding key. Store as string. |
| deployment name | `auth_identities.provider_key` | Use `xlabapi-prod` for the production xlabapi instance. |
| `email` | `auth_identities.email` | Snapshot for display and audit. Not the binding key. |
| `username` | `auth_identities.username`, initial `users.username` | Snapshot on identity. For local user creation, normalize and deduplicate. |
| `status` | login gate, `auth_identities.status` | Only `active` can log in. Save latest snapshot. |
| `role` | `auth_identities.metadata.role` | Audit only. Do not grant Miku admin by default. |
| `balance` | none | Not mapped in this release. |
| `concurrency` | none | Miku keeps its own `concurrency_limit`. |
| `rpm_limit` | none | Miku keeps its own rate-limit behavior. |
| `signup_source` | `auth_identities.metadata.signup_source` | Audit only. |
| `last_login_at` | none | Miku records its own `last_login_at` on the identity. |

## Username and Email Handling

### Username

Miku usernames are unique and limited to 50 characters. When creating a new
Miku user from xlabapi SSO:

1. Start with xlabapi `username`.
2. If empty, use the email local part.
3. If still empty, use `xlabapi_<external_user_id>`.
4. Normalize unsupported characters.
5. Truncate the base value to leave suffix room.
6. If the username already exists, append a numeric suffix such as `_2`, `_3`.

### Email

Miku must not silently merge users by email.

If the xlabapi email is available and not already used by a Miku user, Miku can
use it for the newly created local user.

If the email already exists in Miku and the existing Miku user is not already
bound to the same xlabapi identity, create the SSO user with an internal email:

```text
xlabapi-<external_user_id>@internal.miku
```

The real xlabapi email remains in `auth_identities.email`.

If a user wants to connect an existing Miku account with the same email, they
must first log in to that Miku account and use the personal-center binding flow.

## User Flows

### Flow 1: Miku Login Page

Miku login page adds a button:

```text
Use xlabapi to log in
```

The button starts an xlabapi login handoff:

```text
GET https://xlabapi.com/api/v1/miku-sso/authorize
  ?intent=login
  &redirect_mode=standalone
  &redirect_to=/
  &state=<random>
```

If the user is not logged in to xlabapi, xlabapi asks them to log in first and
then resumes the handoff. If the user is already logged in, xlabapi immediately
creates a one-time ticket and redirects to the Miku callback.

### Flow 2: xlabapi "Image Generation" Entry

xlabapi adds or changes its "Image generation" entry to open the Miku embed
experience:

```text
GET /api/v1/miku-sso/authorize?intent=login&redirect_mode=embed&redirect_to=/workspace
```

After successful authorization, xlabapi loads Miku inside an iframe URL:

```text
https://ai.mikuapi.org/embed/xlabapi?ticket=<ticket>&state=<state>
```

Miku consumes the ticket, signs in the user locally, and renders the embed
workspace. The iframe mode must not rely on third-party cookies.

### Flow 3: Miku Personal Center Binding

Miku personal center adds a third-party account section:

```text
Third-party accounts
xlabapi: Not bound
[Bind xlabapi account]
```

When clicked, Miku starts an SSO authorization with `intent=bind`.

The bind callback requires the user to already have a valid Miku JWT.

After binding:

```text
xlabapi: Bound
Account: alice@example.com
Binding cannot be removed by users.
```

No unbind button is shown.

### Flow 4: Standalone Miku Account Login

When a user enters `https://ai.mikuapi.org/` directly, Miku keeps the existing
email/password login as the primary path.

The login page also shows:

```text
Use xlabapi to log in
```

This alternate path uses the same xlabapi ticket exchange as the standalone
login flow above. The difference is only in the redirect target after Miku
issues its own JWT.

## SSO Protocol

### xlabapi Authorize Endpoint

```http
GET /api/v1/miku-sso/authorize
```

Query parameters:

```text
intent=login|bind
redirect_mode=standalone|embed
redirect_to=<internal Miku path>
state=<random state>
```

Behavior:

1. Require an authenticated xlabapi user.
2. Validate `intent`.
3. Validate `redirect_mode`.
4. Validate `redirect_to` as an internal Miku path only.
5. Reject users whose xlabapi `status` is not `active`.
6. Create a short-lived one-time ticket.
7. If `redirect_mode=embed`, redirect the browser to the Miku iframe entry
   URL with the ticket.
8. If `redirect_mode=standalone`, redirect the browser to the Miku callback
   URL with the ticket.

### xlabapi Ticket Exchange Endpoint

```http
POST /api/v1/miku-sso/token
Authorization: Basic <miku_client_id:miku_client_secret>
```

Request:

```json
{
  "ticket": "one-time-ticket",
  "redirect_mode": "standalone"
}
```

Response:

```json
{
  "provider_key": "xlabapi-prod",
  "external_user_id": "12345",
  "email": "alice@example.com",
  "username": "alice",
  "role": "user",
  "status": "active",
  "signup_source": "email",
  "issued_at": "2026-05-03T12:00:00Z",
  "expires_at": "2026-05-03T12:02:00Z",
  "intent": "login",
  "redirect_to": "/"
}
```

Behavior:

1. Authenticate Miku as an SSO client.
2. Hash and look up the ticket.
3. Ensure ticket exists, has not expired, and has not been consumed.
4. Ensure the requested redirect mode matches the original authorize request.
5. Mark the ticket consumed.
6. Return xlabapi identity claims.

### Miku Callback Endpoint

```http
GET /api/v1/auth/xlabapi/callback?ticket=<ticket>&state=<state>
```

Behavior:

1. Exchange ticket with xlabapi token endpoint.
2. Validate returned `status == "active"`.
3. Dispatch by intent:
   - `login`: find or create a Miku user for this xlabapi identity.
   - `bind`: require current Miku authentication, then bind the identity to the
     current user.
4. Create a short-lived Miku frontend consume session.
5. Redirect to Miku frontend route:

```text
/sso/xlabapi/callback?session=<session-token>
```

### Miku Frontend Consume Endpoint

```http
POST /api/v1/auth/xlabapi/consume-session
```

Request:

```json
{
  "session": "short-lived-session-token"
}
```

Response:

```json
{
  "token": "miku-jwt",
  "user": {
    "id": 88,
    "username": "alice",
    "email": "alice@example.com",
    "balance": 0,
    "role": "user"
  },
  "redirect_to": "/"
}
```

The frontend stores `token` and `user` in the same localStorage keys used by the
existing login flow.

### Miku Embedded Entry

```http
GET /embed/xlabapi?ticket=<ticket>&state=<state>
```

Behavior:

1. Read the ticket from the iframe URL.
2. Exchange the ticket with the Miku backend.
3. If the session is valid, write the normal Miku token and user payload to the
   iframe origin's localStorage.
4. Render the normal Miku workspace inside the embed shell.
5. If the exchange fails, show the Miku login form with the xlabapi option as a
   fallback path.

## Login and Binding Rules

### `intent=login`

1. If the xlabapi identity is already bound, log in the bound Miku user.
2. If the xlabapi identity is not bound, create a Miku user and bind it.
3. Do not merge by email.
4. If xlabapi email conflicts with an existing Miku user, use an internal Miku
   email for the newly created user.

### `intent=bind`

1. The user must already be logged in to Miku.
2. If the xlabapi identity is already bound to the current Miku user, return
   success.
3. If the xlabapi identity is already bound to a different Miku user, reject.
4. If the current Miku user already has an xlabapi identity, reject.
5. Otherwise, bind the xlabapi identity to the current Miku user.

## Security

Required security checks:

- SSO tickets are single-use.
- SSO tickets expire after 2 minutes.
- Miku consume sessions are single-use.
- Miku consume sessions expire after a short duration such as 2 minutes.
- xlabapi only accepts internal Miku paths for `redirect_to`.
- xlabapi rejects SSO for non-active users.
- Miku authenticates to xlabapi token endpoint with a client secret or HMAC.
- Miku does not trust identity claims passed directly through browser query
  parameters.
- `intent=bind` requires existing Miku authentication.
- One xlabapi identity can bind to only one Miku user.
- One Miku user can bind to only one xlabapi identity.
- User-facing unbind is not implemented.
- SSO success and failure events are logged for audit.
- Miku embed mode must not rely on third-party cookies.
- Miku embed mode must only allow framing from `https://xlabapi.com`.
- Miku must explicitly set CSP `frame-ancestors https://xlabapi.com` on the
  embed routes.
- Xlabapi must explicitly allow `https://ai.mikuapi.org` in its iframe source
  policy for the embed container.

## Admin and Configuration

xlabapi configuration:

```text
MIKU_SSO_ENABLED=true
MIKU_SSO_CLIENT_ID=miku-prod
MIKU_SSO_CLIENT_SECRET=...
MIKU_SSO_PROVIDER_KEY=xlabapi-prod
MIKU_SSO_ALLOWED_REDIRECT_MODES=standalone,embed
MIKU_FRONTEND_URL=https://ai.mikuapi.org
```

Miku configuration:

```text
XLABAPI_SSO_ENABLED=true
XLABAPI_SSO_AUTHORIZE_URL=https://xlabapi.com/api/v1/miku-sso/authorize
XLABAPI_SSO_TOKEN_URL=https://xlabapi.com/api/v1/miku-sso/token
XLABAPI_SSO_CLIENT_ID=miku-prod
XLABAPI_SSO_CLIENT_SECRET=...
XLABAPI_SSO_PROVIDER_KEY=xlabapi-prod
XLABAPI_SSO_EMBED_URL=https://ai.mikuapi.org/embed/xlabapi
```

Admin UI is optional for the first release. Environment configuration is
acceptable for the initial deployment.

## Testing

Backend tests should cover:

- xlabapi authorize rejects unauthenticated users.
- xlabapi authorize rejects invalid redirect modes and unsafe redirect paths.
- xlabapi authorize rejects non-active users.
- xlabapi token exchange consumes a ticket once.
- xlabapi token exchange rejects expired tickets.
- Miku login creates a new user and identity for an unbound xlabapi user.
- Miku login returns the existing Miku user for a bound xlabapi identity.
- Miku login does not merge by email when a Miku local user already has that
  email.
- Miku embed consumes a ticket and renders the workspace without third-party
  cookie dependence.
- Miku bind requires current Miku authentication.
- Miku bind rejects an xlabapi identity already bound to another Miku user.
- Miku bind rejects a Miku user that already has an xlabapi binding.
- Miku consume session can be used only once.

Frontend tests should cover:

- Miku login page renders the xlabapi login button.
- Clicking the xlabapi login button starts the authorize flow.
- Miku personal center shows unbound and bound xlabapi states.
- Bound state does not render an unbind button.
- xlabapi "Image generation" entry opens the Miku iframe embed entry.
- Miku embed login fallback still renders the xlabapi login option.

## Out of Scope

These features are explicitly deferred:

- xlabapi balance to Miku credit exchange.
- Miku credit to xlabapi balance return.
- xlabapi proxying Miku image generation requests.
- Automatic role synchronization.
- Automatic concurrency or RPM synchronization.
- Silent same-email account merging.
- User-facing unbind.

## Open Implementation Notes

- The exact frontend route names can follow Miku's existing router conventions.
- Miku can initially use backend HTML to write localStorage, but the preferred
  design is the frontend consume-session route because it is easier to test and
  gives cleaner error handling.
- If operational support later requires unbinding, add an admin-only remediation
  action with audit logging instead of a user-facing unbind feature.
