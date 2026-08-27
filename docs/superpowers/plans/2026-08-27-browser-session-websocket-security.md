# Browser Session and WebSocket Security Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (- [ ]) syntax for tracking.

**Goal:** Move the browser UI to an HttpOnly server-side session, enforce same-origin WebSocket authentication, queue and correlate commands on one persistent connection, and remove dynamic HTML injection.

**Architecture:** Keep external API tokens unchanged and add a separate in-memory browser-session store. Authenticate /data/ before upgrade, preserve a same-origin one-request legacy query-token path, and use one cookie-authenticated FIFO browser connection with request IDs while server commands remain sequential. Harden only loaded TypeScript renderers and HTML responses, then regenerate checked-in assets.

**Tech Stack:** Go 1.27, checked-in vendor tree, Gorilla WebSocket, TypeScript, Node DOM harnesses launched from Go tests, embedded assets in src/webUI.go.

**Spec:** docs/superpowers/specs/2026-08-27-browser-session-websocket-security-design.md

## Global Constraints

- Keep /api/ token format, rotation, permissions, request schema, and response schema unchanged.
- Preserve credential-free /data/ when AuthenticationWEB is false or ConfigurationWizard is true, while enforcing browser origins.
- Preserve same-origin /data/?Token=... for one-command legacy clients; new browser code never uses it.
- Execute at most one command at a time and never replay a command written to the socket.
- Add no frontend framework, dependency, state library, or concurrent backend mutation.
- Change TypeScript in ts/, compile to html/js/, and regenerate src/webUI.go.
- Use gofmt and -mod=vendor. Preserve unrelated .claude/ and .serena/ content.

---

### Task 1: Add isolated browser-session primitives

**Files:**
- Create: src/internal/authentication/browser_session.go
- Create: src/internal/authentication/browser_session_test.go
- Modify: src/internal/authentication/authentication.go
- Modify: src/internal/authentication/password_test.go

**Interfaces:**
- Produces BrowserSessionCookieName.
- Produces AuthenticateBrowser(username, password string, permissions ...string) (sessionID, userID string, err error).
- Produces AuthorizeBrowserSession(sessionID string, permissions ...string) (userID string, err error).
- Produces InvalidateBrowserSession(sessionID string).
- Produces BrowserSessionCookie(sessionID string, secure bool) (*http.Cookie, error).
- Produces ExpiredBrowserSessionCookie(secure bool) *http.Cookie.

- [ ] **Step 1: Write failing lifecycle tests**

Use initAuthenticationTest(t). Cover independent sessions, wrong password, missing permission, expiry removal, one-session invalidation, user removal, password-change invalidation, username-only preservation, and independence from API token rotation.

Core assertion:

    first, firstUser, err := AuthenticateBrowser("browser-user", "password", "authentication.web")
    second, secondUser, err := AuthenticateBrowser("browser-user", "password", "authentication.web")
    if err != nil || first == second || firstUser != userID || secondUser != userID {
        t.Fatalf("sessions/users = %q/%q %q/%q: %v", first, firstUser, second, secondUser, err)
    }
    if _, err := AuthorizeBrowserSession(first, "authentication.web"); err != nil {
        t.Fatal(err)
    }

- [ ] **Step 2: Verify RED**

    go test -count=1 -mod=vendor ./src/internal/authentication -run 'TestAuthenticateBrowser|TestAuthorizeBrowserSession|TestBrowserSession|TestChangeCredentialsInvalidatesBrowserSessions|TestRemoveUserInvalidatesBrowserSessions'

Expected: compilation fails because the session API is absent.

- [ ] **Step 3: Extract credential verification**

Extract the locked credential check from UserAuthentication into:

    func authenticateUserLocked(username, password string) (userID string, err error)

Retain Argon2id verification, successful legacy migration, persistence-before-success, and exact errors. UserAuthentication locks, calls the helper, then returns setToken(userID, "-"). Run existing password tests and require PASS.

- [ ] **Step 4: Implement the minimal session store**

Use:

    const BrowserSessionCookieName = "ThreadfinSession"

    type browserSession struct {
        userID  string
        expires time.Time
    }

    var browserSessions = make(map[string]browserSession)
    var readBrowserSessionRandom = rand.Read

Generate 32 bytes with crypto/rand, reject errors and short reads, encode with base64.RawURLEncoding, and store an absolute expiry. AuthenticateBrowser checks requested permissions atomically. AuthorizeBrowserSession removes expired records and checks current user data/permissions. Reset sessions in Init. Invalidate a user's sessions only after successful password change or removal.

- [ ] **Step 5: Write cookie/random-failure tests, verify RED, then implement**

Assert exact name, value, HttpOnly, SameSiteStrictMode, Path "/", expiry, conditional Secure, and deletion MaxAge below zero. Replace readBrowserSessionRandom with a sentinel-error function and restore with t.Cleanup; no session may remain. The live cookie uses the stored expiry. The expired cookie uses empty value, time.Unix(1, 0), MaxAge -1, HttpOnly, SameSite Strict, Path "/", and requested Secure.

- [ ] **Step 6: Verify and commit**

    gofmt -w src/internal/authentication/browser_session.go src/internal/authentication/browser_session_test.go src/internal/authentication/authentication.go src/internal/authentication/password_test.go
    go test -count=1 -mod=vendor ./src/internal/authentication
    go test -race -count=1 -mod=vendor ./src/internal/authentication
    git diff --check
    git add src/internal/authentication/browser_session.go src/internal/authentication/browser_session_test.go src/internal/authentication/authentication.go src/internal/authentication/password_test.go
    git commit -m "feat(auth): add isolated browser sessions"

---

### Task 2: Move browser login and logout to the HttpOnly session

**Files:**
- Modify: src/authentication.go
- Modify: src/webserver.go
- Modify: src/authentication_test.go
- Modify: src/ui_admin_test.go
- Modify: ts/menu_ts.ts
- Generate: html/js/menu_ts.js

**Interfaces:**
- Consumes Task 1 session functions.
- Produces browserCookieSecure(*http.Request) bool.
- Produces authorizeBrowserRequest(*http.Request, ...string) (string, error).
- Produces WebLogout(http.ResponseWriter, *http.Request).

- [ ] **Step 1: Write failing login-session tests**

Change the old token-cookie test to require ThreadfinSession, HttpOnly, SameSite Strict, Path "/", no password leakage, HTTP without Secure, HTTPS with Secure, and authenticated GET with the returned cookie. Add a user lacking authentication.web and an empty-credential POST that expires the cookie.

- [ ] **Step 2: Add failing logout and API-compatibility tests**

Create two sessions, POST /web/logout with one cookie, and prove only that session is invalid. Non-POST returns 405. Add a /api/ test proving Token still rotates and ThreadfinSession is never emitted.

    go test -count=1 -mod=vendor ./src -run 'TestUIAdminAuthentication|TestWebLogout|TestAPIBrowserSessionCompatibility'

Expected: old Token behavior and missing WebLogout fail.

- [ ] **Step 3: Integrate normal and first-user login**

Change createFirstUserForAuthentication to return a browser session. After CreateDefaultUser, authenticate without a permission argument to obtain both the new session and user ID:

    sessionID, userID, err := authentication.AuthenticateBrowser(username, password)

Write the existing first-user permission map through WriteUserData(userID, userData), then require AuthorizeBrowserSession(sessionID, "authentication.web") before returning it. Invalidate the session if permission persistence or authorization fails. Normal login calls AuthenticateBrowser with authentication.web directly. Set BrowserSessionCookie with http.SetCookie. GET reads only ThreadfinSession and calls AuthorizeBrowserSession. Missing or invalid sessions render login without issuing/rotating an API token.

- [ ] **Step 4: Implement logout and Secure selection**

Register /web/logout. Accept POST only, invalidate presented session, set ExpiredBrowserSessionCookie, and redirect to /web/ with 303. browserCookieSecure returns true only for direct TLS or configured HTTPS protocol and ignores forwarded headers.

- [ ] **Step 5: Change browser logout**

Replace document.cookie access with:

    var form = document.createElement("form")
    form.method = "post"
    form.action = "/web/logout"
    document.body.appendChild(form)
    form.submit()

Update the Node harness to assert method/action and no active Token-cookie access.

- [ ] **Step 6: Verify and commit**

    gofmt -w src/authentication.go src/webserver.go src/authentication_test.go src/ui_admin_test.go
    tsc -p ./ts/tsconfig.json
    go test -count=1 -mod=vendor ./src -run 'TestUIAdminAuthentication|TestWebLogout|TestAPI|TestGenerated.*Logout'
    git diff --check
    git add src/authentication.go src/webserver.go src/authentication_test.go src/ui_admin_test.go ts/menu_ts.ts html/js/menu_ts.js
    git commit -m "feat(auth): use HttpOnly browser sessions"

Embedding parity is completed in Task 6.

---

### Task 3: Authenticate WebSockets before upgrade and correlate responses

**Files:**
- Create: src/websocket_auth.go
- Create: src/websocket_auth_test.go
- Modify: src/webserver.go
- Modify: src/struct-webserver.go

**Interfaces:**
- Produces RequestStruct.RequestID and ResponseStruct.RequestID with JSON name requestId and omitempty.
- Produces webSocketOriginAllowed(*http.Request) bool.
- Produces authenticateWebSocketRequest(*http.Request) (webSocketAuthentication, error).

- [ ] **Step 1: Write origin tests and verify RED**

Cover absent origin, exact HTTP/HTTPS origin, wrong scheme/port, userinfo, path/query/fragment, malformed values, evilthreadfin.example, and threadfin.example.evil for host threadfin.example:34400.

    go test -count=1 -mod=vendor ./src -run TestWebSocketOriginAllowed

Expected: helper missing.

- [ ] **Step 2: Implement exact parsed-origin comparison**

Parse Origin with url.Parse. Require http/https, no userinfo/path/query/fragment, the effective scheme from direct TLS or Threadfin's configured web protocol, and exact case-insensitive host equality with r.Host. Missing Origin remains allowed for credentialed non-browser clients.

- [ ] **Step 3: Write pre-upgrade auth tests**

Using httptest.NewServer and Gorilla DefaultDialer, cover valid session cookie, missing/expired/permission-denied sessions, cross-origin rejection with both credentials, auth-disabled/setup exact-origin access, same-origin legacy token rotation, and no-Origin legacy access. Require failures before upgrade with 401/403.

- [ ] **Step 4: Implement the auth boundary**

Use:

    type webSocketAuthentication struct {
        browserSessionID string
        legacyToken      string
        persistent       bool
    }

Snapshot AuthenticationWEB and ConfigurationWizard under systemMutex. When auth is required and ThreadfinSession is present, reject an invalid session without falling back to a query token. Only when the session cookie is absent may the query Token use the existing validator; require authentication.web through checkAuthorizationLevel after rotation. Cookie and auth-disabled/setup connections persist. Legacy token connections process one command and return the rotated token. Check origin and credentials before Upgrade.

- [ ] **Step 5: Add request-ID tests, verify RED, then fields**

Add a RequestID string field with JSON name requestId and omitempty to both structs. Test exact JSON spelling, omission for legacy messages, and two cookie-authenticated messages returning request-1 then request-2.

- [ ] **Step 6: Make the loop safely persistent**

Initialize request and response inside every read loop:

    for {
        var request RequestStruct
        if err := conn.ReadJSON(&request); err != nil {
            return
        }
        response := ResponseStruct{Status: true, RequestID: request.RequestID}
    }

Immediately after response initialization, reauthorize authentication.web for session connections; on failure send close code 1008 and return. Then execute the existing command switch once and write response once. Preserve locks and lockConfigMutationForCommand; launch no goroutines. Break after one response only for legacy query-token connections. Reset response fields per command.

- [ ] **Step 7: Verify and commit**

    gofmt -w src/websocket_auth.go src/websocket_auth_test.go src/webserver.go src/struct-webserver.go
    go test -count=1 -mod=vendor ./src -run 'TestWebSocket|TestRequestStruct|TestResponseStruct'
    go test -race -count=1 -mod=vendor ./src -run TestWebSocket
    go vet -mod=vendor ./src
    git diff --check
    git add src/websocket_auth.go src/websocket_auth_test.go src/webserver.go src/struct-webserver.go
    git commit -m "feat(websocket): authenticate and correlate browser commands"

---

### Task 4: Replace dropped commands with a persistent FIFO client

**Files:**
- Create: src/ui_network_test.go
- Modify: ts/network_ts.ts
- Modify: src/ui_sources_test.go
- Modify: src/ui_filter_test.go
- Modify: src/ui_mapping_test.go
- Modify: src/ui_operations_test.go
- Modify: src/ui_overview_test.go
- Generate: html/js/network_ts.js

**Interfaces:**
- Produces ThreadfinConnection.enqueue(command, data), one THREADFIN_CONNECTION, and unchanged new Server(cmd).request(data) call sites.

- [ ] **Step 1: Write queue test and verify RED**

Queue three commands rapidly. Assert credential-free /data/ URL, one initial send, unique IDs, FIFO release, mismatched-ID protocol failure, one settlement for error plus close, unsent work retained, sent work never replayed, and close 1008 rejecting the queue/reloading.

    go test -count=1 -mod=vendor ./src -run TestGeneratedWebSocketQueue

Expected: busy rejection, per-command sockets, and ?Token= violate the test.

- [ ] **Step 2: Implement explicit queue state**

Use:

    type QueuedThreadfinRequest = {
      command: string
      data: any
      requestId: string
      sent: boolean
      settled: boolean
      timeoutId: any
    }

    class ThreadfinConnection {
      socket: WebSocket = null
      queue: QueuedThreadfinRequest[] = []
      active: QueuedThreadfinRequest = null
      nextRequestId: number = 1
      enqueue(command: string, data: any): void
      connect(): void
      pump(): void
      settleResponse(response: any): void
      settleTransportFailure(): void
    }

Server.request copies input, sets cmd on the copy, and enqueues. Move existing callbacks into named success/failure helpers invoked once. Use a 30-second timeout. On ambiguity fail/discard active work, retain unsent work, reconnect only with queued work.

- [ ] **Step 3: Adapt fake sockets**

Give affected fakes readyState, OPEN, close(), and multiple sends. Replace obsolete busy assertions with queued completion. Preserve source/filter/mapping/wizard/overview/activity and malformed-response behavior.

- [ ] **Step 4: Compile, verify, and commit**

    tsc -p ./ts/tsconfig.json
    go test -count=1 -mod=vendor ./src -run 'TestGenerated(WebSocketQueue|Task5|Filter|Mapping|Operations|Overview|Wizard|Source)'
    git diff --check

Stage the exact Task 4 paths; unchanged paths are harmless:

    git add ts/network_ts.ts html/js/network_ts.js src/ui_network_test.go src/ui_sources_test.go src/ui_filter_test.go src/ui_mapping_test.go src/ui_operations_test.go src/ui_overview_test.go

Commit:

    git commit -m "feat(ui): queue commands on one websocket"

---

### Task 5: Remove active dynamic HTML sinks and add HTML security headers

**Files:**
- Create: src/web_security.go
- Create: src/web_security_test.go
- Create: src/ui_security_test.go
- Modify: src/webserver.go
- Modify: ts/network_ts.ts
- Modify: ts/menu_ts.ts
- Generate: html/js/network_ts.js
- Generate: html/js/menu_ts.js

- [ ] **Step 1: Write header tests and verify RED**

Require exact CSP:

    default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; font-src 'self' data: https://cdnjs.cloudflare.com; img-src 'self' data: http: https:; connect-src 'self' ws: wss:; form-action 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'self'

Also require nosniff, no-referrer, DENY. Prove JSON API and streaming responses do not receive HTML-only headers.

- [ ] **Step 2: Implement narrow headers**

Define browserContentSecurityPolicy and setBrowserSecurityHeaders in web_security.go. Call from Web only when final contentType is text/html and before WriteHeader. Do not wrap the mux or alter other responses.

- [ ] **Step 3: Write injection tests and verify RED**

Inject as provider/mapping/group/probe/server display data:

    </span><img src=x onerror="globalThis.threadfinInjected=true"><script>globalThis.threadfinInjected=true</script>

Run real compiled renderers. Assert exact visible text, no IMG/SCRIPT/onerror, and flag remains false. Add a source contract allowing innerHTML only for empty clearing or repository-controlled constants without runtime interpolation.

    go test -count=1 -mod=vendor ./src -run 'TestWebBrowserSecurityHeaders|TestGeneratedRuntimeValuesRemainText'

Expected: missing headers and a confirmed parsed runtime sink.

- [ ] **Step 4: Replace confirmed sinks**

Build probe rows and connection capacity with explicit nodes. Use textContent for mapping/provider/user/server values and dynamic headings/descriptions. Keep innerHTML = "" only for clearing. Do not rewrite unloaded historical JavaScript.

- [ ] **Step 5: Compile, verify, and commit**

    gofmt -w src/web_security.go src/web_security_test.go src/ui_security_test.go src/webserver.go
    tsc -p ./ts/tsconfig.json
    go test -count=1 -mod=vendor ./src -run 'TestWebBrowserSecurityHeaders|TestGeneratedRuntimeValuesRemainText|TestGenerated.*(Mapping|Operations|Admin|Source)'
    git diff --check

Stage the exact Task-5 files and commit:

    git add src/web_security.go src/web_security_test.go src/ui_security_test.go src/webserver.go ts/network_ts.ts ts/menu_ts.ts html/js/network_ts.js html/js/menu_ts.js

    git commit -m "fix(ui): harden runtime rendering and browser headers"

---

### Task 6: Regenerate embedded assets and complete acceptance

**Files:**
- Modify generated: src/webUI.go
- Modify focused source/tests only if acceptance exposes a regression

- [ ] **Step 1: Regenerate deterministically**

Run the documented development path with an owner-only temporary config, then repeat and compare:

    tsc -p ./ts/tsconfig.json
    THREADFIN_EMBED_DIR=$(mktemp -d -p /tmp threadfin-embed.XXXXXX)
    chmod 700 "$THREADFIN_EMBED_DIR"
    go build -mod=vendor -o "$THREADFIN_EMBED_DIR/threadfin" .
    timeout --signal=INT 10s "$THREADFIN_EMBED_DIR/threadfin" -dev -config "$THREADFIN_EMBED_DIR/config/" -bind 127.0.0.1 -port 0 || embed_status=$?
    test "${embed_status:-0}" -eq 0 || test "$embed_status" -eq 124 || test "$embed_status" -eq 130
    cp src/webUI.go "$THREADFIN_EMBED_DIR/first-webUI.go"
    unset embed_status
    timeout --signal=INT 10s "$THREADFIN_EMBED_DIR/threadfin" -dev -config "$THREADFIN_EMBED_DIR/config/" -bind 127.0.0.1 -port 0 || embed_status=$?
    test "${embed_status:-0}" -eq 0 || test "$embed_status" -eq 124 || test "$embed_status" -eq 130
    cmp -s "$THREADFIN_EMBED_DIR/first-webUI.go" src/webUI.go

After all checks, delete only files below the validated THREADFIN_EMBED_DIR and remove that directory.

- [ ] **Step 2: Run focused gates**

    go test -count=1 -mod=vendor ./src/internal/authentication
    go test -count=1 -mod=vendor ./src -run 'Test(WebSocket|UIAdminAuthentication|WebLogout|WebBrowserSecurityHeaders|Generated)'
    go test -count=1 -mod=vendor ./src -run 'Test.*Embedded.*Match|TestHTML'

- [ ] **Step 3: Run complete verification**

    go test -count=1 -mod=vendor ./...
    go test -race -count=1 -mod=vendor ./...
    go vet -mod=vendor ./...

- [ ] **Step 4: Run platform builds**

Use owner-only THREADFIN_VERIFY_DIR:

    GOOS=windows GOARCH=amd64 go build -mod=vendor -o "$THREADFIN_VERIFY_DIR/Threadfin_windows_amd64.exe" .
    GOOS=linux GOARCH=amd64 go build -mod=vendor -o "$THREADFIN_VERIFY_DIR/Threadfin_linux_amd64" .
    GOOS=linux GOARCH=arm64 go build -mod=vendor -o "$THREADFIN_VERIFY_DIR/Threadfin_linux_arm64" .

Remove only that validated temporary directory.

- [ ] **Step 5: Run loopback headless acceptance**

With updates and SSDP disabled, verify: HttpOnly cookie and credential-free socket URL; rapid actions complete FIFO; Mapping guard/save works; payloads display inertly; logout invalidates current session only; second session survives; cross-origin socket fails. Record redacted status/protocol evidence only.

- [ ] **Step 6: Review scope and commit embed**

    git status --short
    git diff --check
    git diff --stat
    git diff --name-only

Stage only exact implementation/test/compiled/embed files, never .claude/ or .serena/, then commit:

    git commit -m "build(ui): embed secure session client assets"

- [ ] **Step 7: Report evidence**

Report commits, changed files, RED/GREEN evidence, test/race/vet/build outcomes, browser acceptance, remaining legacy query-token boundary, and worktree status. Do not push, open a PR, or merge without separate authorization.
