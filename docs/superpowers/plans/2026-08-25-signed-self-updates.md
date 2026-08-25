# Signed Self-Updates Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Require an Ed25519-authenticated SHA-256 digest for every self-update and verify the complete candidate before replacing the current executable.

**Architecture:** The updater embeds a public key and isolates sidecar URL construction, bounded metadata fetching, signature verification, checksum parsing, and verified artifact download in a testable integrity module. Release CI signs a checksum sidecar for each binary with a GitHub-held private key; `DoUpdate` prepares and verifies a temporary candidate before entering the existing replacement/restart phase.

**Tech Stack:** Go 1.27 standard `crypto/ed25519`, `crypto/sha256`, `embed`, OpenSSL Ed25519 signing in GitHub Actions, existing updater and release workflow.

**Spec:** `docs/superpowers/specs/2026-08-25-auth-update-security-design.md`

## Global Constraints

- Publish `<asset>.sha256` and raw 64-byte `<asset>.sha256.sig` beside every release binary.
- Sign the exact lowercase 64-hex-character checksum plus newline.
- Never commit, print, or pass the private key as a command-line argument.
- Store the private key in GitHub secret `THREADFIN_UPDATE_SIGNING_KEY` and `/home/volsch/.config/threadfin/update-signing-key.pem` with mode `0600`.
- Missing, malformed, unsigned, or mismatched metadata must leave the current executable untouched.
- Custom update servers must provide the same signed sidecars; there is no unsigned compatibility mode.

---

### Task 1: Provision the Release Signing Key

**Files:**
- Create outside Git: `/home/volsch/.config/threadfin/update-signing-key.pem`
- Create: `src/internal/up2date/client/update-signing-public-key.pem`
- Mutate repository setting: GitHub Actions secret `THREADFIN_UPDATE_SIGNING_KEY`

**Interfaces:**
- Consumes: repository administrator access through the authenticated `gh` CLI.
- Produces: an owner-only Ed25519 private key, the matching committed public key, and the repository signing secret.

- [ ] **Step 1: Confirm the private key does not already exist**

Run:

```bash
test ! -e /home/volsch/.config/threadfin/update-signing-key.pem
gh repo view --json nameWithOwner,viewerPermission
```

Expected: the file absence check exits zero, and GitHub reports `volschin/Threadfin` with `ADMIN` permission. If the key already exists, stop and derive the public key from it instead of replacing it.

- [ ] **Step 2: Generate the owner-only key and committed public half**

Run without printing key material:

```bash
install -d -m 0700 /home/volsch/.config/threadfin
umask 077
openssl genpkey -algorithm ED25519 -out /home/volsch/.config/threadfin/update-signing-key.pem
openssl pkey -in /home/volsch/.config/threadfin/update-signing-key.pem -pubout -out src/internal/up2date/client/update-signing-public-key.pem
chmod 0600 /home/volsch/.config/threadfin/update-signing-key.pem
chmod 0644 src/internal/up2date/client/update-signing-public-key.pem
test "$(stat -c '%a' /home/volsch/.config/threadfin/update-signing-key.pem)" = 600
```

Expected: all commands exit zero. Do not display either file during execution logs.

- [ ] **Step 3: Store the private key in GitHub Actions**

Run with stdin redirection so the secret never appears in arguments:

```bash
gh secret set THREADFIN_UPDATE_SIGNING_KEY --repo volschin/Threadfin < /home/volsch/.config/threadfin/update-signing-key.pem
gh secret list --repo volschin/Threadfin | rg '^THREADFIN_UPDATE_SIGNING_KEY[[:space:]]'
```

Expected: the secret name is listed without its value.

### Task 2: Update Integrity Primitive

**Files:**
- Create: `src/internal/up2date/client/integrity.go`
- Create: `src/internal/up2date/client/integrity_test.go`
- Track: `src/internal/up2date/client/update-signing-public-key.pem`

**Interfaces:**
- Consumes: the public key created in Task 1 and an `*http.Client`.
- Produces: `embeddedUpdatePublicKey() (ed25519.PublicKey, error)`, `sidecarURL(string, string) (string, error)`, `fetchExpectedChecksum(*http.Client, string, ed25519.PublicKey) ([32]byte, error)`, and `downloadVerified(*http.Client, string, string, [32]byte) error`.

- [ ] **Step 1: Write failing URL and signature tests**

Create `src/internal/up2date/client/integrity_test.go`:

```go
package up2date

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func signedUpdateServer(t *testing.T, artifact []byte, checksum []byte, signature []byte) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Threadfin_linux_amd64":
			_, _ = w.Write(artifact)
		case "/Threadfin_linux_amd64.sha256":
			_, _ = w.Write(checksum)
		case "/Threadfin_linux_amd64.sha256.sig":
			_, _ = w.Write(signature)
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestSidecarURLPreservesQueryAndAddsSuffixToPath(t *testing.T) {
	got, err := sidecarURL("https://updates.example/Threadfin_linux_amd64?token=abc", ".sha256")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://updates.example/Threadfin_linux_amd64.sha256?token=abc"
	if got != want {
		t.Fatalf("sidecar URL = %q, want %q", got, want)
	}
}

func TestFetchExpectedChecksumVerifiesSignature(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte("verified update"))
	checksum := []byte(hex.EncodeToString(digest[:]) + "\n")
	server := signedUpdateServer(t, []byte("verified update"), checksum, ed25519.Sign(privateKey, checksum))
	defer server.Close()

	got, err := fetchExpectedChecksum(server.Client(), server.URL+"/Threadfin_linux_amd64", publicKey)
	if err != nil {
		t.Fatal(err)
	}
	if got != digest {
		t.Fatalf("digest = %x, want %x", got, digest)
	}
}

func TestFetchExpectedChecksumRejectsInvalidSignature(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte("update"))
	checksum := []byte(hex.EncodeToString(digest[:]) + "\n")
	server := signedUpdateServer(t, []byte("update"), checksum, make([]byte, ed25519.SignatureSize))
	defer server.Close()

	if _, err := fetchExpectedChecksum(server.Client(), server.URL+"/Threadfin_linux_amd64", publicKey); err == nil {
		t.Fatal("invalid signature was accepted")
	}
}

func TestFetchExpectedChecksumRejectsMissingSignature(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte("update"))
	checksum := []byte(hex.EncodeToString(digest[:]) + "\n")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Threadfin_linux_amd64.sha256" {
			_, _ = w.Write(checksum)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	if _, err := fetchExpectedChecksum(server.Client(), server.URL+"/Threadfin_linux_amd64", publicKey); err == nil {
		t.Fatal("missing signature was accepted")
	}
}

func TestFetchExpectedChecksumRejectsSignedMalformedChecksum(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	checksum := []byte("not-a-sha256\n")
	server := signedUpdateServer(t, []byte("update"), checksum, ed25519.Sign(privateKey, checksum))
	defer server.Close()

	if _, err := fetchExpectedChecksum(server.Client(), server.URL+"/Threadfin_linux_amd64", publicKey); err == nil {
		t.Fatal("signed malformed checksum was accepted")
	}
}
```

- [ ] **Step 2: Run integrity tests and confirm RED**

Run:

```bash
GOTOOLCHAIN=go1.27.0 go test -mod=vendor ./src/internal/up2date/client
```

Expected: compilation fails because the integrity functions do not exist.

- [ ] **Step 3: Implement embedded key parsing and signed checksum fetching**

Create `src/internal/up2date/client/integrity.go` with these exact boundaries:

```go
package up2date

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/subtle"
	_ "embed"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"crypto/x509"
)

const updateMetadataLimit int64 = 1024

//go:embed update-signing-public-key.pem
var updateSigningPublicKeyPEM []byte

func embeddedUpdatePublicKey() (ed25519.PublicKey, error) {
	block, rest := pem.Decode(updateSigningPublicKeyPEM)
	if block == nil || len(rest) != 0 {
		return nil, fmt.Errorf("invalid embedded update public key PEM")
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse embedded update public key: %w", err)
	}
	publicKey, ok := parsed.(ed25519.PublicKey)
	if !ok || len(publicKey) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("embedded update key is not Ed25519")
	}
	return publicKey, nil
}

func sidecarURL(artifactURL, suffix string) (string, error) {
	u, err := url.Parse(artifactURL)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("unsupported update URL scheme %q", u.Scheme)
	}
	u.Path += suffix
	u.RawPath = ""
	return u.String(), nil
}

func fetchBounded(client *http.Client, requestURL string, limit int64) ([]byte, error) {
	resp, err := client.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update metadata returned %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("update metadata exceeds %d bytes", limit)
	}
	return body, nil
}

func fetchExpectedChecksum(client *http.Client, artifactURL string, publicKey ed25519.PublicKey) ([32]byte, error) {
	var digest [32]byte
	checksumURL, err := sidecarURL(artifactURL, ".sha256")
	if err != nil {
		return digest, err
	}
	signatureURL, err := sidecarURL(artifactURL, ".sha256.sig")
	if err != nil {
		return digest, err
	}
	checksum, err := fetchBounded(client, checksumURL, updateMetadataLimit)
	if err != nil {
		return digest, err
	}
	signature, err := fetchBounded(client, signatureURL, ed25519.SignatureSize)
	if err != nil {
		return digest, err
	}
	if len(signature) != ed25519.SignatureSize || !ed25519.Verify(publicKey, checksum, signature) {
		return digest, fmt.Errorf("invalid update checksum signature")
	}
	if len(checksum) != sha256.Size*2+1 || checksum[len(checksum)-1] != '\n' {
		return digest, fmt.Errorf("invalid signed SHA-256 checksum format")
	}
	decoded, err := hex.DecodeString(string(checksum[:len(checksum)-1]))
	if err != nil || hex.EncodeToString(decoded) != string(checksum[:len(checksum)-1]) {
		return digest, fmt.Errorf("invalid signed SHA-256 checksum")
	}
	copy(digest[:], decoded)
	return digest, nil
}
```

Run `gofmt` after creation; it will order `crypto/x509` with the other crypto imports.

- [ ] **Step 4: Add failing verified-download tests**

Append:

```go
func TestDownloadVerifiedWritesMatchingArtifact(t *testing.T) {
	artifact := []byte("verified update")
	digest := sha256.Sum256(artifact)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(artifact)
	}))
	defer server.Close()
	destination := filepath.Join(t.TempDir(), "candidate")
	if err := downloadVerified(server.Client(), server.URL, destination, digest); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(artifact) {
		t.Fatalf("candidate = %q, want %q", got, artifact)
	}
}

func TestDownloadVerifiedRemovesChecksumMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "tampered")
	}))
	defer server.Close()
	destination := filepath.Join(t.TempDir(), "candidate")
	want := sha256.Sum256([]byte("trusted"))
	if err := downloadVerified(server.Client(), server.URL, destination, want); err == nil {
		t.Fatal("checksum mismatch was accepted")
	}
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("unverified candidate remains: %v", err)
	}
}
```

Remove unused `time` from the test imports if the compiler reports it.

- [ ] **Step 5: Run download tests and confirm RED**

Run:

```bash
GOTOOLCHAIN=go1.27.0 go test -mod=vendor ./src/internal/up2date/client
```

Expected: compilation fails because `downloadVerified` does not exist.

- [ ] **Step 6: Implement verified temporary download**

Append to `integrity.go`:

```go
func downloadVerified(client *http.Client, artifactURL, destination string, expected [32]byte) (err error) {
	resp, err := client.Get(artifactURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update artifact returned %s", resp.Status)
	}

	out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	succeeded := false
	defer func() {
		if !succeeded {
			_ = os.Remove(destination)
		}
	}()

	hash := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(out, hash), resp.Body)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if subtle.ConstantTimeCompare(hash.Sum(nil), expected[:]) != 1 {
		return fmt.Errorf("update artifact SHA-256 mismatch")
	}
	succeeded = true
	return nil
}
```

- [ ] **Step 7: Run integrity tests and confirm GREEN**

Run:

```bash
gofmt -w src/internal/up2date/client/integrity.go src/internal/up2date/client/integrity_test.go
GOTOOLCHAIN=go1.27.0 go test -mod=vendor ./src/internal/up2date/client
```

Expected: all integrity and ZIP tests pass.

- [ ] **Step 8: Commit the integrity primitive**

```bash
git add docs/superpowers/plans/2026-08-25-signed-self-updates.md src/internal/up2date/client/integrity.go src/internal/up2date/client/integrity_test.go src/internal/up2date/client/update-signing-public-key.pem
git commit -m "feat: verify signed update checksums"
```

### Task 3: Verify Before Executable Replacement

**Files:**
- Modify: `src/internal/up2date/client/update.go:18-163`
- Modify: `src/internal/up2date/client/integrity.go`
- Modify: `src/internal/up2date/client/integrity_test.go`

**Interfaces:**
- Consumes: integrity functions from Task 2 and existing `extractZIP`/`copyFile` helpers.
- Produces: `prepareVerifiedUpdate(*http.Client, string, string, string, string, ed25519.PublicKey) (candidate string, cleanup func(), err error)` and a `DoUpdate` flow that calls replacement only after preparation succeeds.

- [ ] **Step 1: Write failing preparation tests**

Append tests that exercise the real network, signature, checksum, and filesystem behavior:

```go
func TestPrepareVerifiedUpdateReturnsCandidateWithoutTouchingCurrentBinary(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	artifact := []byte("new executable")
	digest := sha256.Sum256(artifact)
	checksum := []byte(hex.EncodeToString(digest[:]) + "\n")
	server := signedUpdateServer(t, artifact, checksum, ed25519.Sign(privateKey, checksum))
	defer server.Close()
	directory := t.TempDir()
	current := filepath.Join(directory, "threadfin")
	if err := os.WriteFile(current, []byte("current executable"), 0755); err != nil {
		t.Fatal(err)
	}

	candidate, cleanup, err := prepareVerifiedUpdate(server.Client(), server.URL+"/Threadfin_linux_amd64", "bin", "", directory, publicKey)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	gotCurrent, _ := os.ReadFile(current)
	if string(gotCurrent) != "current executable" {
		t.Fatalf("current executable changed during verification: %q", gotCurrent)
	}
	gotCandidate, _ := os.ReadFile(candidate)
	if string(gotCandidate) != string(artifact) {
		t.Fatalf("candidate = %q, want %q", gotCandidate, artifact)
	}
}

func TestPrepareVerifiedUpdateMismatchLeavesNoCandidate(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	trusted := sha256.Sum256([]byte("trusted"))
	checksum := []byte(hex.EncodeToString(trusted[:]) + "\n")
	server := signedUpdateServer(t, []byte("tampered"), checksum, ed25519.Sign(privateKey, checksum))
	defer server.Close()
	directory := t.TempDir()

	if _, _, err := prepareVerifiedUpdate(server.Client(), server.URL+"/Threadfin_linux_amd64", "bin", "", directory, publicKey); err == nil {
		t.Fatal("tampered update was prepared")
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("temporary update material remains: %v", entries)
	}
}

func TestPrepareVerifiedZipReturnsExpectedBinary(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	var archiveBody bytes.Buffer
	archiveWriter := zip.NewWriter(&archiveBody)
	entry, err := archiveWriter.Create("Threadfin_linux_amd64")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("verified zipped executable")); err != nil {
		t.Fatal(err)
	}
	if err := archiveWriter.Close(); err != nil {
		t.Fatal(err)
	}
	artifact := archiveBody.Bytes()
	digest := sha256.Sum256(artifact)
	checksum := []byte(hex.EncodeToString(digest[:]) + "\n")
	server := signedUpdateServer(t, artifact, checksum, ed25519.Sign(privateKey, checksum))
	defer server.Close()

	candidate, cleanup, err := prepareVerifiedUpdate(server.Client(), server.URL+"/Threadfin_linux_amd64", "zip", "Threadfin_linux_amd64", t.TempDir(), publicKey)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	got, err := os.ReadFile(candidate)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "verified zipped executable" {
		t.Fatalf("zipped candidate = %q", got)
	}
}
```

Add `archive/zip` and `bytes` to the test imports.

- [ ] **Step 2: Run preparation tests and confirm RED**

Run:

```bash
GOTOOLCHAIN=go1.27.0 go test -mod=vendor ./src/internal/up2date/client
```

Expected: compilation fails because `prepareVerifiedUpdate` does not exist.

- [ ] **Step 3: Implement verified candidate preparation**

Add to `integrity.go` and its imports:

```go
func prepareVerifiedUpdate(client *http.Client, artifactURL, fileType, filename, directory string, publicKey ed25519.PublicKey) (string, func(), error) {
	expected, err := fetchExpectedChecksum(client, artifactURL, publicKey)
	if err != nil {
		return "", func() {}, err
	}
	temporary, err := os.CreateTemp(directory, ".threadfin-update-*")
	if err != nil {
		return "", func() {}, err
	}
	downloadPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		_ = os.Remove(downloadPath)
		return "", func() {}, err
	}
	if err := os.Remove(downloadPath); err != nil {
		return "", func() {}, err
	}
	cleanupPaths := []string{downloadPath}
	cleanup := func() {
		for _, cleanupPath := range cleanupPaths {
			_ = os.RemoveAll(cleanupPath)
		}
	}
	if err := downloadVerified(client, artifactURL, downloadPath, expected); err != nil {
		cleanup()
		return "", func() {}, err
	}
	if fileType != "zip" {
		return downloadPath, cleanup, nil
	}
	extractDirectory, err := os.MkdirTemp(directory, ".threadfin-update-extract-*")
	if err != nil {
		cleanup()
		return "", func() {}, err
	}
	cleanupPaths = append(cleanupPaths, extractDirectory)
	if err := extractZIP(downloadPath, extractDirectory); err != nil {
		cleanup()
		return "", func() {}, err
	}
	candidate := filepath.Join(extractDirectory, filepath.Base(filename))
	info, err := os.Stat(candidate)
	if err != nil || !info.Mode().IsRegular() {
		cleanup()
		return "", func() {}, fmt.Errorf("verified update archive does not contain %q", filepath.Base(filename))
	}
	return candidate, cleanup, nil
}
```

Add `path/filepath` to imports.

- [ ] **Step 4: Refactor `DoUpdate` to prepare first and replace second**

Use one package-level client with a finite timeout:

```go
var updateHTTPClient = &http.Client{Timeout: 2 * time.Minute}
```

At the start of the URL-present branch, resolve the executable and embedded public key, then prepare the candidate before calling `os.Rename`, `copyFile`, `os.Chmod`, process restart, or `syscall.Exec`:

```go
binary, err := os.Executable()
if err != nil {
	return err
}
publicKey, err := embeddedUpdatePublicKey()
if err != nil {
	return err
}
path := getPlatformPath(binary)
candidate, cleanup, err := prepareVerifiedUpdate(updateHTTPClient, url, fileType, filenameBIN, path, publicKey)
if err != nil {
	return err
}
defer cleanup()
```

After preparation succeeds, preserve the existing backup/rollback naming, but copy only `candidate` into the executable path:

```go
filename := getFilenameFromPath(binary)
oldBinary := path + "_old_" + filename
_ = os.Remove(oldBinary)
if err := os.Rename(binary, oldBinary); err != nil {
	return err
}
if err := copyFile(candidate, binary); err != nil {
	restorOldBinary(oldBinary, binary)
	return err
}
if err := os.Chmod(binary, 0755); err != nil {
	restorOldBinary(oldBinary, binary)
	return err
}
```

Remove the direct `http.Get`, direct response-to-executable copy, and ZIP extraction from `DoUpdate`; `prepareVerifiedUpdate` now owns those operations. Keep platform restart behavior after verified replacement. On `syscall.Exec` failure, restore `oldBinary` before returning the error; do not delete the backup before calling `syscall.Exec`.

- [ ] **Step 5: Run updater tests and confirm GREEN**

Run:

```bash
gofmt -w src/internal/up2date/client/integrity.go src/internal/up2date/client/integrity_test.go src/internal/up2date/client/update.go
GOTOOLCHAIN=go1.27.0 go test -mod=vendor ./src/internal/up2date/client
GOTOOLCHAIN=go1.27.0 go test -race -mod=vendor ./src/internal/up2date/client
```

Expected: signature, checksum, candidate preservation, ZIP traversal, and race tests all pass.

- [ ] **Step 6: Commit verified update preparation**

```bash
git add src/internal/up2date/client/integrity.go src/internal/up2date/client/integrity_test.go src/internal/up2date/client/update.go
git commit -m "fix: verify updates before replacing executable"
```

### Task 4: Sign Every Release Binary

**Files:**
- Modify: `.github/workflows/release.yml:30-82`

**Interfaces:**
- Consumes: GitHub secret `THREADFIN_UPDATE_SIGNING_KEY`, committed public PEM, and `dist/Threadfin_*` binaries.
- Produces: a matching `.sha256` and `.sha256.sig` for each uploaded and released binary.

- [ ] **Step 1: Add the release signing step**

Insert immediately after `Build Binaries` and before `Upload build artifacts`:

```yaml
      - name: Sign release artifacts
        env:
          UPDATE_SIGNING_KEY: ${{ secrets.THREADFIN_UPDATE_SIGNING_KEY }}
        run: |
          set -euo pipefail
          test -n "$UPDATE_SIGNING_KEY"
          signing_key=$(mktemp)
          cleanup() {
            rm -f "$signing_key"
          }
          trap cleanup EXIT
          chmod 600 "$signing_key"
          printf '%s' "$UPDATE_SIGNING_KEY" > "$signing_key"
          openssl pkey -in "$signing_key" -pubout | cmp - src/internal/up2date/client/update-signing-public-key.pem

          assets=(dist/Threadfin_*)
          test "${#assets[@]}" -gt 0
          for asset in "${assets[@]}"; do
            checksum=$(sha256sum "$asset")
            printf '%s\n' "${checksum%% *}" > "${asset}.sha256"
            openssl pkeyutl -sign -rawin -inkey "$signing_key" -in "${asset}.sha256" -out "${asset}.sha256.sig"
            openssl pkeyutl -verify -pubin -inkey src/internal/up2date/client/update-signing-public-key.pem -rawin -in "${asset}.sha256" -sigfile "${asset}.sha256.sig"
          done
```

The existing artifact upload and release patterns already use `dist/` and `dist/*`, so they will publish all sidecars without further glob changes.

- [ ] **Step 2: Validate workflow syntax**

Run:

```bash
if command -v actionlint >/dev/null 2>&1; then actionlint; else "$(go env GOPATH)/bin/actionlint"; fi
```

Expected: no output and exit zero.

- [ ] **Step 3: Exercise the exact signing commands locally without exposing the key**

Run in a temporary directory using the recovery key and public key:

```bash
sign_test_dir=$(mktemp -d)
cleanup_sign_test() {
  find "$sign_test_dir" -type f -delete
  rmdir "$sign_test_dir"
}
trap cleanup_sign_test EXIT
printf 'release candidate' > "$sign_test_dir/Threadfin_linux_amd64"
asset="$sign_test_dir/Threadfin_linux_amd64"
checksum=$(sha256sum "$asset")
printf '%s\n' "${checksum%% *}" > "${asset}.sha256"
openssl pkeyutl -sign -rawin -inkey /home/volsch/.config/threadfin/update-signing-key.pem -in "${asset}.sha256" -out "${asset}.sha256.sig"
openssl pkeyutl -verify -pubin -inkey src/internal/up2date/client/update-signing-public-key.pem -rawin -in "${asset}.sha256" -sigfile "${asset}.sha256.sig"
```

Expected: OpenSSL prints `Signature Verified Successfully` without printing key material.

- [ ] **Step 4: Commit release signing**

```bash
git add .github/workflows/release.yml
git commit -m "ci: sign self-update artifacts"
```

### Task 5: Signed Update Verification Gate

**Files:**
- Verify only; no source changes expected.

**Interfaces:**
- Consumes: completed signed-update implementation and workflow.
- Produces: evidence for every update-integrity acceptance criterion.

- [ ] **Step 1: Run formatting, static, workflow, dependency, and vulnerability checks**

```bash
gofmt -w src/internal/up2date/client/integrity.go src/internal/up2date/client/integrity_test.go src/internal/up2date/client/update.go
git diff --check -- . ':(exclude)vendor/**'
GOTOOLCHAIN=go1.27.0 go vet -mod=vendor ./...
GOTOOLCHAIN=go1.27.0 go mod verify
GOTOOLCHAIN=go1.27.0 go run golang.org/x/vuln/cmd/govulncheck@latest -mode=source ./...
if command -v actionlint >/dev/null 2>&1; then actionlint; else "$(go env GOPATH)/bin/actionlint"; fi
```

Expected: all commands exit zero and `govulncheck` prints `No vulnerabilities found.`

- [ ] **Step 2: Run complete unit and race tests**

```bash
GOTOOLCHAIN=go1.27.0 go test -mod=vendor ./...
GOTOOLCHAIN=go1.27.0 go test -race -mod=vendor ./...
```

Expected: all packages pass.

- [ ] **Step 3: Cross-build all release targets**

Run:

```bash
build_test_dir=$(mktemp -d)
cleanup_build_test() {
  find "$build_test_dir" -type f -delete
  rmdir "$build_test_dir"
}
trap cleanup_build_test EXIT
targets=(
  "linux arm64"
  "linux arm"
  "linux amd64"
  "freebsd amd64"
  "freebsd arm"
  "darwin arm64"
  "darwin amd64"
  "windows amd64"
)
pids=()
names=()
for target in "${targets[@]}"; do
  read -r target_os target_arch <<< "$target"
  output="$build_test_dir/threadfin-${target_os}-${target_arch}"
  if [ "$target_os" = windows ]; then
    output="${output}.exe"
  fi
  GOOS="$target_os" GOARCH="$target_arch" GOTOOLCHAIN=go1.27.0 go build -mod=vendor -o "$output" . &
  pids+=("$!")
  names+=("${target_os}/${target_arch}")
done
for index in "${!pids[@]}"; do
  wait "${pids[$index]}"
  printf 'PASS %s\n' "${names[$index]}"
done
```

Expected: all eight builds exit zero.

- [ ] **Step 4: Build the standard and ARM containers**

```bash
docker buildx build --platform linux/amd64 --build-arg TARGETARCH=amd64 --output=type=cacheonly .
docker buildx build --platform linux/arm/v7 --target builder --output=type=cacheonly -f Dockerfile.arm .
```

Expected: both builds exit zero.

- [ ] **Step 5: Verify secret and repository boundaries**

```bash
test "$(stat -c '%a' /home/volsch/.config/threadfin/update-signing-key.pem)" = 600
gh secret list --repo volschin/Threadfin | rg '^THREADFIN_UPDATE_SIGNING_KEY[[:space:]]'
if git grep -n 'BEGIN PRIVATE KEY' -- ':!docs/superpowers/plans'; then
  exit 1
fi
git status --short
```

Expected: the local key is mode `0600`, the secret name exists, no private key is tracked, and the worktree is clean after the implementation commits.
