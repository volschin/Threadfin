# Threadfin Envoy WebSocket Origin Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox ( - [ ] ) syntax for tracking.

**Goal:** Allow the Threadfin administration WebSocket to operate behind an HTTPS-terminating Envoy proxy without enabling global ForceHttps or trusting forwarded headers.

**Architecture:** Keep the existing strict Origin parser and exact authority comparison, but stop requiring the browser Origin scheme to equal the backend transport scheme. Only HTTP and HTTPS Origins with the exact request Host remain valid. Release the change as Threadfin v3.2.3 so HomeOps can pin the resulting signed multi-architecture image digest.

**Tech Stack:** Go 1.27.0, vendored Go modules, gorilla/websocket, GitHub Actions, GHCR, Cosign.

**Spec:** /home/volsch/projekte/homeops/docs/superpowers/specs/2026-08-30-threadfin-easyepg-migration-design.md

## Global Constraints

- Preserve the unrelated untracked .claude/ and .serena/ directories.
- Use an isolated worktree created through superpowers:using-git-worktrees.
- Use the Go version declared in go.mod and build with -mod=vendor.
- Do not read X-Forwarded-Proto or any other forwarded header.
- Continue rejecting non-HTTP(S) schemes, mismatched authorities, userinfo, paths, queries, fragments, opaque URLs, and malformed Origins.
- Do not change authentication, cookies, API behavior, media URLs, ForceHttps, or generated browser assets.
- Keep Version and APIVersion synchronized for the v3.2.3 release.
- Never move or overwrite an existing release tag.
- Do not publish, merge, tag, or release until the applicable explicit delivery approval has been confirmed.

## File Structure

- Modify src/websocket_auth.go: implement the transport-independent same-authority Origin rule.
- Modify src/websocket_auth_test.go: specify the allowed proxy case and all retained rejection cases.
- Modify threadfin.go only in the version-only release change.
- No TypeScript or generated web asset changes are required.
- No dependency or vendor changes are allowed.

---

### Task 1: Specify the HTTPS-Origin/HTTP-Upstream Contract

**Files:**

- Modify: src/websocket_auth_test.go
- Test: src/websocket_auth_test.go

**Interfaces:**

- Consumes: webSocketOriginAllowed(r *http.Request) bool
- Produces: a table-driven contract in TestWebSocketOriginAllowed and an upgrade-level regression test

- [ ] **Step 1: Create an isolated feature worktree**

Run the superpowers:using-git-worktrees workflow from /home/volsch/projekte/Threadfin. Create a branch named fix/envoy-websocket-origin without staging, deleting, or copying .claude/ or .serena/.

Expected: the new worktree starts from current origin/main and git status --short shows no unrelated files.

- [ ] **Step 2: Change the unit test before the implementation**

Replace the scheme-dependent cases in TestWebSocketOriginAllowed with this contract:

```go
tests := []struct {
    name   string
    origin string
    want   bool
}{
    {name: "absent", want: true},
    {name: "exact HTTP", origin: "http://threadfin.example:34400", want: true},
    {name: "exact HTTPS over HTTP upstream", origin: "https://threadfin.example:34400", want: true},
    {name: "case insensitive HTTPS authority", origin: "https://THREADFIN.EXAMPLE:34400", want: true},
    {name: "unsupported websocket scheme", origin: "ws://threadfin.example:34400"},
    {name: "unsupported file scheme", origin: "file://threadfin.example:34400"},
    {name: "wrong port", origin: "https://threadfin.example:34401"},
    {name: "userinfo", origin: "https://user@threadfin.example:34400"},
    {name: "path", origin: "https://threadfin.example:34400/web"},
    {name: "query", origin: "https://threadfin.example:34400?source=web"},
    {name: "fragment", origin: "https://threadfin.example:34400#web"},
    {name: "malformed", origin: "https://threadfin.example:34400/%zz"},
    {name: "host prefix", origin: "https://evilthreadfin.example:34400"},
    {name: "host suffix", origin: "https://threadfin.example.evil:34400"},
}
```

Build every request with:

```go
request := httptest.NewRequest(http.MethodGet, "http://threadfin.example:34400/data/", nil)
```

Remove directTLS and configuredHTTPS from the test table and remove crypto/tls from the imports if it becomes unused.

- [ ] **Step 3: Add an upgrade-level proxy regression test**

Add this test next to TestWebSocketOriginAllowed:

```go
func TestWebSocketOriginAllowsHTTPSAtHTTPUpstream(t *testing.T) {
    restorePersistentState(t)
    configureWebSocketAuthentication(false, false)
    server, webSocketURL := newWebSocketTestServer(t)

    origin := "https" + strings.TrimPrefix(server.URL, "http")
    conn := dialWebSocket(t, webSocketURL, http.Header{"Origin": []string{origin}})
    if err := conn.Close(); err != nil {
        t.Fatal(err)
    }
}
```

This uses an HTTP test server while presenting the exact same authority as an HTTPS browser Origin.

- [ ] **Step 4: Run the focused tests and verify the new case fails**

Run:

```bash
go test -count=1 -mod=vendor ./src -run 'TestWebSocketOrigin(Allowed|AllowsHTTPSAtHTTPUpstream)$'
```

Expected: FAIL because exact HTTPS over an HTTP upstream is still rejected. All retained rejection cases must continue to report the expected false result.

- [ ] **Step 5: Commit only after Task 2 is complete**

Do not commit the deliberately failing test by itself. The test and minimal implementation form one atomic commit in Task 2.

---

### Task 2: Implement the Same-Authority Rule

**Files:**

- Modify: src/websocket_auth.go
- Modify: src/websocket_auth_test.go
- Test: src/websocket_auth_test.go

**Interfaces:**

- Consumes: parsed Origin and r.Host
- Produces: webSocketOriginAllowed(r *http.Request) bool with no global settings or forwarded-header dependency

- [ ] **Step 1: Remove transport-derived scheme selection**

Replace webSocketOriginAllowed with the following minimal implementation:

```go
func webSocketOriginAllowed(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    if origin == "" {
        return true
    }

    parsedOrigin, err := url.Parse(origin)
    if err != nil || parsedOrigin.Scheme == "" || parsedOrigin.Host == "" {
        return false
    }
    if !strings.EqualFold(parsedOrigin.Scheme, "http") && !strings.EqualFold(parsedOrigin.Scheme, "https") {
        return false
    }
    if parsedOrigin.User != nil || parsedOrigin.Opaque != "" || parsedOrigin.Path != "" || parsedOrigin.RawPath != "" || parsedOrigin.RawQuery != "" || parsedOrigin.ForceQuery || parsedOrigin.Fragment != "" || parsedOrigin.RawFragment != "" || strings.Contains(origin, "#") {
        return false
    }

    return strings.EqualFold(parsedOrigin.Host, r.Host)
}
```

Do not add proxy CIDRs, trusted proxies, forwarded-header parsing, configurable aliases, or wildcard hosts.

- [ ] **Step 2: Format the changed Go files**

Run:

```bash
gofmt -w src/websocket_auth.go src/websocket_auth_test.go
```

Expected: gofmt exits zero and changes no unrelated file.

- [ ] **Step 3: Run the focused tests**

Run:

```bash
go test -count=1 -mod=vendor ./src -run 'TestWebSocketOrigin(Allowed|AllowsHTTPSAtHTTPUpstream)$'
```

Expected: PASS.

- [ ] **Step 4: Run the complete WebSocket authentication suite**

Run:

```bash
go test -count=1 -mod=vendor ./src -run 'TestWebSocket'
```

Expected: PASS, including authentication-before-upgrade, origin rejection, revocation, shutdown, and persistent request tests.

- [ ] **Step 5: Review the exact diff**

Run:

```bash
git diff --check
git diff -- src/websocket_auth.go src/websocket_auth_test.go
```

Expected: only the scheme-coupling removal and its tests are present. There must be no authentication relaxation for a different Host.

- [ ] **Step 6: Commit the feature**

Run:

```bash
git add src/websocket_auth.go src/websocket_auth_test.go
git commit -m "fix: allow HTTPS WebSocket origin behind proxy"
```

Expected: one atomic feature commit.

---

### Task 3: Verify the Fork Change

**Files:**

- Verify: all non-vendored Go packages
- Verify: go.mod, go.sum, vendor/

**Interfaces:**

- Consumes: the Task 2 commit
- Produces: local evidence that the feature is safe to deliver

- [ ] **Step 1: Run the repository-required test suite**

Run:

```bash
go test -count=1 -mod=vendor ./...
```

Expected: PASS.

- [ ] **Step 2: Run repository-required vet**

Run:

```bash
go vet -mod=vendor ./...
```

Expected: PASS.

- [ ] **Step 3: Compile the platform targets affected by shared Go code**

Run:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o /tmp/threadfin-linux-amd64 .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod=vendor -o /tmp/threadfin-linux-arm64 .
GOOS=windows GOARCH=amd64 go build -mod=vendor -o /tmp/threadfin-windows-amd64.exe .
```

Expected: all three builds succeed. Remove only these explicit /tmp artifacts afterward.

- [ ] **Step 4: Prove dependencies were untouched**

Run:

```bash
git diff HEAD^ -- go.mod go.sum vendor/
git status --short
```

Expected: no dependency diff and no worktree changes after verification.

---

### Task 4: Deliver the Feature PR

**Files:**

- No additional source changes

**Interfaces:**

- Consumes: verified feature commit
- Produces: merged exact feature head on main

- [ ] **Step 1: Stop for delivery authorization**

Confirm that an explicit authorization covers pushing the feature branch and creating and merging its PR. Do not infer release or tag authorization from design approval alone.

- [ ] **Step 2: Push the feature branch and open the PR**

Run through non-interactive git and gh commands:

```bash
git push -u origin fix/envoy-websocket-origin
pr_url=$(gh pr create \
  --base main \
  --head fix/envoy-websocket-origin \
  --title "fix: allow HTTPS WebSocket origin behind proxy" \
  --body "Allow an HTTPS browser Origin when Envoy connects to Threadfin over HTTP, while retaining exact authority validation and all malformed/cross-host rejections. No forwarded headers are trusted. Verified with focused WebSocket tests, the full Go test suite, vet, and cross-builds.")
pr_number=\${pr_url##*/}
test -n "$pr_number"
```

The PR body must state the exact same-authority rule, retained rejection cases, tests run, and the absence of forwarded-header trust. Do not include credentials or private hostnames.

- [ ] **Step 3: Wait for required CI and verify exact head**

Run:

```bash
gh pr checks --watch "$pr_number"
reviewed_head=$(gh pr view "$pr_number" --json headRefOid --jq .headRefOid)
test "$reviewed_head" = "$(git rev-parse HEAD)"
gh pr view "$pr_number" --json mergeStateStatus,statusCheckRollup
```

Expected: all required checks are green and headRefOid equals the reviewed local HEAD.

- [ ] **Step 4: Merge only the exact reviewed head**

Use the repository's protected merge method. Abort if the head changed or any required check is not green.

- [ ] **Step 5: Refresh main**

Run:

```bash
git fetch origin
git switch main
git pull --ff-only
```

Expected: main contains the feature commit and is clean.

---

### Task 5: Prepare and Release v3.2.3

**Files:**

- Modify: threadfin.go
- Test: threadfin.go and complete repository

**Interfaces:**

- Consumes: merged feature on main
- Produces: signed v3.2.3 release and immutable GHCR manifest digest for HomeOps

- [ ] **Step 1: Create a clean version-only branch**

Create a new isolated branch release/v3.2.3 from current origin/main.

- [ ] **Step 2: Update only Version and APIVersion**

Apply exactly:

```go
const Version = "3.2.3"
const APIVersion = "3.2.3"
```

Do not change DBVersion or any other source.

- [ ] **Step 3: Verify the version-only diff**

Run:

```bash
gofmt -w threadfin.go
git diff --check
git diff -- threadfin.go
go test -count=1 -mod=vendor ./...
go vet -mod=vendor ./...
```

Expected: tests and vet pass; the diff changes exactly two version literals.

- [ ] **Step 4: Commit the version bump**

Run:

```bash
git add threadfin.go
git commit -m "chore: release v3.2.3"
```

- [ ] **Step 5: Stop for release delivery authorization**

Confirm explicit authorization for the version PR, merge, annotated tag, and resulting GitHub/GHCR release. A feature-PR approval does not automatically authorize tagging.

- [ ] **Step 6: Deliver the version-only PR**

Push release/v3.2.3, open a PR, wait for all required checks, compare the reviewed head SHA, and merge only that exact SHA.

- [ ] **Step 7: Wait for post-merge main CI**

Run:

```bash
merge_sha=$(git rev-parse origin/main)
main_ci_run_id=$(gh run list --branch main --workflow ci.yml --limit 20 --json databaseId,headSha --jq '.[] | select(.headSha == "'"$merge_sha"'") | .databaseId' | head -n1)
test -n "$main_ci_run_id"
gh run watch "$main_ci_run_id" --exit-status
```

Expected: the CI run for the exact merge commit succeeds.

- [ ] **Step 8: Create and push the annotated tag**

After proving v3.2.3 does not already exist:

```bash
test -z "$(git tag -l v3.2.3)"
git tag -a v3.2.3 -m "Threadfin v3.2.3" "$merge_sha"
git push origin v3.2.3
```

Expected: the tag points to the exact green main merge commit.

- [ ] **Step 9: Verify the release workflow and artifacts**

Run:

```bash
release_run_id=$(gh run list --workflow release.yml --limit 20 --json databaseId,headBranch --jq '.[] | select(.headBranch == "v3.2.3") | .databaseId' | head -n1)
test -n "$release_run_id"
gh run watch "$release_run_id" --exit-status
gh release view v3.2.3
docker buildx imagetools inspect ghcr.io/volschin/threadfin:v3.2.3
docker buildx imagetools inspect ghcr.io/volschin/threadfin:v3.2.3-nvidia
```

Expected: GitHub release exists; signed binary/checksum/signature assets exist; both image variants contain linux/amd64 and linux/arm64 manifests.

- [ ] **Step 10: Record the immutable standard-image digest**

Run:

```bash
docker buildx imagetools inspect ghcr.io/volschin/threadfin:v3.2.3 --format '{{json .Manifest}}' | jq -r .digest
```

Expected: exactly one index digest matching the regular expression sha256:[0-9a-f]{64}. Pass this literal digest to the HomeOps plan without substituting the NVIDIA variant.

---

## Self-Review Checklist

- Spec coverage: HTTPS browser Origin through an HTTP upstream, exact Host enforcement, no forwarded-header trust, tests, versioning, release, and immutable digest are all assigned.
- Placeholder scan: every operational identifier is resolved into a checked shell variable before use; no source file or command contains a placeholder.
- Type consistency: webSocketOriginAllowed retains its existing signature and all callers remain unchanged.
- Scope: no UI asset, authentication, cookie, dependency, media URL, Dockerfile, or API change is included.
