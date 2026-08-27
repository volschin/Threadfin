# Browser Session and WebSocket Security Design

## Goal

Remove browser session credentials from JavaScript and WebSocket URLs, eliminate dropped UI commands, and close the active dynamic-HTML injection surface without changing Threadfin's external API token contract or making global backend state concurrent.

## Scope

- Introduce server-side browser sessions backed by an opaque cookie.
- Authenticate and authorize browser WebSocket connections before upgrading them.
- Retain same-origin legacy WebSocket query-token compatibility for a transition period.
- Replace the browser's one-connection-per-command behavior with one persistent, queued, request-correlated WebSocket connection.
- Replace active dynamic `innerHTML` sinks with text or explicit DOM construction.
- Add narrowly scoped browser security headers compatible with the currently loaded assets.

The external `/api/` token format, token rotation, configuration API contract, command names, command payloads, response data, persisted settings, and backend mutation order remain unchanged. True parallel command execution, a frontend framework migration, self-hosting third-party browser assets, and removal of the legacy query-token path are outside this change.

## Compatibility Boundary

Threadfin has two distinct authentication consumers after this change:

1. External API clients continue to use the existing rotating token functions and JSON contracts without modification.
2. The Threadfin browser UI uses a separate server-side browser session and never reads or transmits its credential from JavaScript.

The existing `/data/?Token=<token>` WebSocket form remains temporarily available only when the request passes the same origin policy as cookie-authenticated browser traffic. New browser code never uses it. This compatibility path is explicitly transitional; removing it requires a separate usage and release decision.

## Browser Session Model

### Session record

The authentication package owns a dedicated browser-session store. Each record contains:

- a cryptographically random opaque session ID;
- the authenticated user ID;
- an absolute expiry time.

Session identifiers use `crypto/rand`, and random-source errors fail login. They are independent of API tokens, are never persisted to disk, and are invalidated by process restart. Multiple logins for one user create independent sessions so separate browsers and tabs do not invalidate one another.

### Cookie contract

A successful browser login creates a browser session and returns a cookie with:

- `Name=ThreadfinSession`;
- `HttpOnly=true`;
- `SameSite=Strict`;
- `Path=/`;
- an explicit expiry matching the server-side record;
- `Secure=true` when the effective request transport is HTTPS.

Threadfin determines secure transport from the direct request TLS state and its existing configured HTTPS behavior; it does not trust arbitrary forwarded headers. HTTP deployments retain a non-`Secure` cookie so current private-LAN installations continue to work.

Logout removes only the presented browser session and expires its cookie with the same name and path. Expired sessions are rejected and removed when encountered. Password changes and user removal invalidate that user's browser sessions. Permission checks use current authentication data for each command rather than a permission snapshot stored in the session.

### Login compatibility

Browser form login and first-user completion issue the new browser-session cookie. They no longer expose an API token to browser JavaScript. Existing API authentication entry points continue issuing their current tokens. Tests distinguish these routes so later refactoring cannot silently mix the two credential types.

## WebSocket Authentication and Origin Policy

### Pre-upgrade authentication

When web authentication is enabled and setup is complete, `/data/` authenticates before calling the WebSocket upgrader:

1. Validate the request origin.
2. Prefer a valid `ThreadfinSession` cookie.
3. If no browser session is present, allow the legacy `Token` query parameter through the existing token validator.
4. Reject missing, expired, malformed, or unauthorized credentials with HTTP `401` or `403` before protocol upgrade.

When `Settings.AuthenticationWEB` is false or `System.ConfigurationWizard` is true, `/data/` preserves the current unauthenticated setup/private-LAN behavior but still enforces the origin policy for browser requests. Enabling browser sessions must not silently enable authentication for installations that have it disabled.

The authenticated principal is attached to the connection state. Every command rechecks that the user still exists and applies the browser authorization rules already used by that command; this phase introduces no new permission matrix. A session that becomes invalid closes the connection with a policy-violation close code.

### Origin validation

Requests without an `Origin` header remain allowed for non-browser legacy clients after credential validation. Requests with an `Origin` header are accepted only when scheme and host match the effective Threadfin web origin. The comparison uses parsed, normalized origins and the existing configured Threadfin domain/port rules; suffix and substring matches are forbidden. Cross-origin requests fail before either cookie or query-token authentication can open a socket.

## Persistent Queued WebSocket Client

### Wire envelope

Browser requests add a `requestId` string to the existing request object. Browser responses echo the same `requestId`. Command names and command-specific fields are otherwise unchanged. Legacy requests without a `requestId` continue receiving their existing response shape.

Request IDs are unique within the browser page lifetime and are not security credentials. The server continues reading and executing commands in arrival order on one connection.

### Client queue

`Server(cmd).request(data)` remains the UI-facing entry point. Internally, one connection manager owns:

- one persistent WebSocket;
- a FIFO queue of commands not yet sent;
- at most one sent command awaiting a response;
- monotonically increasing request IDs;
- connection, response, and timeout state.

The next queued request is sent only after the active request reaches a terminal state. This preserves Threadfin's current mutation ordering while eliminating the `SERVER_CONNECTION` behavior that discards actions.

### Failure and retry semantics

- A response completes only the request with the matching ID. Missing or mismatched IDs are treated as protocol errors and never settle another request.
- A request-specific timeout fails the active request and reconnects before continuing the queue.
- Commands still waiting locally are retained across reconnects.
- Once a command has been written to the socket, it is never automatically replayed because the server may have applied it before the connection failed.
- A transport failure settles the active request exactly once through the existing screen-specific failure callbacks.
- Logout, authentication failure, or policy-violation closure rejects all queued work and routes to the existing reload/login behavior.

No command cancellation protocol or concurrent server execution is introduced.

## DOM Injection Boundary

All values originating from providers, M3U/XMLTV content, mapping state, probes, logs, settings, or server responses are untrusted display text. Active TypeScript renderers must insert those values using `textContent`, input values, safe attributes, or explicitly created nodes.

`innerHTML` remains permitted only for clearing an element with the literal empty string or inserting a repository-controlled static template that cannot contain runtime data. Mixed static markup and runtime values, such as probe details or connection counts, is constructed from separate elements and text nodes.

The change covers only scripts loaded by the current HTML pages. Unloaded historical JavaScript assets are not rewritten in this phase; their later removal belongs to legacy asset cleanup.

Tests inject representative strings containing tags, event handlers, entities, and script-closing sequences into each corrected runtime data path and assert that the payload is displayed as text and creates no executable element or handler.

## Browser Security Headers

HTML responses receive a centralized policy containing:

- `Content-Security-Policy` restricted to the application origin plus the exact CDN origins currently required for Bootstrap, Font Awesome, and Clipboard.js;
- `X-Content-Type-Options: nosniff`;
- `Referrer-Policy: no-referrer`;
- `X-Frame-Options: DENY` and equivalent CSP `frame-ancestors 'none'` protection.

The baseline CSP is explicit: `default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; font-src 'self' data: https://cdnjs.cloudflare.com; img-src 'self' data: http: https:; connect-src 'self' ws: wss:; form-action 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'self'`. The inline allowances remain only because current HTML and generated controls still use inline handlers and styles. Removing those allowances is a separate hardening task and must not be claimed by this phase. Streaming, M3U, XMLTV, HDHomeRun, image, and API responses retain their existing cross-origin and content contracts.

## Components and Files

- `src/internal/authentication/browser_session.go`: browser-session lifecycle and cookie construction.
- `src/internal/authentication/browser_session_test.go`: session, cookie, expiry, invalidation, and concurrency contracts.
- `src/webserver.go` and focused webserver tests: pre-upgrade authentication, origin policy, principal revalidation, and request-ID echo.
- `ts/network_ts.ts`: persistent connection manager and FIFO request queue behind the existing `Server` interface.
- existing UI test harnesses under `src/ui_*_test.go`: queue, correlation, failure, and injection regressions.
- active TypeScript renderers containing runtime `innerHTML`: replace only confirmed dynamic sinks.
- `src/html-build.go` or the narrow HTTP response path responsible for HTML: centralized browser headers.
- generated `html/js/*` and `src/webUI.go`: regenerated only after TypeScript source changes pass.

## Testing Strategy

Implementation follows red-green-refactor cycles. Required automated coverage includes:

### Authentication

- successful browser login sets every cookie attribute;
- HTTP and HTTPS produce the intended `Secure` behavior;
- two sessions for one user remain independently valid;
- logout invalidates only the presented session;
- expiry, user removal, and password change invalidate access;
- random-source failure issues neither a session nor a cookie;
- API token behavior and rotation remain unchanged.

### WebSocket

- valid browser cookie upgrades without a query token;
- invalid or expired cookie fails before upgrade;
- authentication-disabled and setup-mode sockets remain credential-free but enforce browser origins;
- allowed same-origin and rejected cross-origin cases cover scheme, host, port, malformed origins, and misleading suffixes;
- credentialed requests without `Origin` retain legacy non-browser compatibility;
- a valid same-origin legacy query token still works;
- each response echoes its request ID, while legacy messages remain compatible;
- permission removal during a connection blocks the next command.

### Client transport

- rapid commands are queued rather than dropped;
- FIFO order and one-active-request behavior are preserved;
- response IDs settle only their matching request;
- malformed responses, timeout, error-plus-close, and reconnect settle once;
- unsent work resumes after reconnect;
- sent work is not replayed automatically;
- authentication closure rejects the queue and follows the login path.

### DOM and headers

- malicious provider, mapping, probe, log, and server values remain inert text;
- static labels and intended markup retain their rendered structure;
- HTML responses contain the policy headers;
- non-HTML API and streaming response contracts remain unchanged.

### Project gates

- `tsc -p ./ts/tsconfig.json`;
- deterministic generated-asset regeneration and parity tests;
- `go test -count=1 -mod=vendor ./...`;
- `go test -race -count=1 -mod=vendor ./...`;
- `go vet -mod=vendor ./...`;
- existing Linux builds and Windows cross-build checks;
- headless browser acceptance for login, navigation, queued rapid actions, mapping save protection, logout, and representative injection payloads.

## Delivery Slices

1. Add browser-session primitives and route-level login/logout integration while preserving API tokens.
2. Move WebSocket authentication before upgrade and enforce the origin policy with legacy query-token compatibility.
3. Introduce request-ID echo and the persistent queued browser client without changing command payloads.
4. Remove active dynamic HTML sinks and add the compatible browser headers.
5. Regenerate embedded assets and run full automated and headless acceptance gates.

Each slice must leave its focused tests green and must not mix unrelated module cleanup.

## Acceptance Criteria

- Browser JavaScript cannot read the session credential.
- Browser WebSocket URLs contain no credential.
- Cross-origin browser WebSocket upgrades are rejected before authentication.
- External `/api/` clients observe no token-contract change.
- Same-origin legacy WebSocket query-token clients continue to work during the transition.
- Rapid UI actions are queued and completed in order rather than discarded.
- A transport ambiguity never automatically repeats a potentially mutating command.
- Confirmed runtime values cannot create DOM elements or event handlers through the corrected renderers.
- HTML responses carry the defined browser security headers without breaking current UI workflows.
- No backend command is executed concurrently as part of this change.
