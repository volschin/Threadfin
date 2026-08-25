# Password Hashing Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Store all new and changed passwords with Argon2id and transparently migrate valid legacy HMAC-SHA256 credentials on login.

**Architecture:** A focused password helper owns PHC encoding, strict parsing, Argon2id derivation, legacy verification, and constant-time comparison. The existing authentication database keeps its shape; creation and credential-change paths write Argon2id, while the login path detects and persists a legacy upgrade before issuing a token.

**Tech Stack:** Go 1.27, `golang.org/x/crypto/argon2`, `crypto/rand`, PHC strings, existing JSON authentication database.

**Spec:** `docs/superpowers/specs/2026-08-25-auth-update-security-design.md`

## Global Constraints

- Use Argon2id with memory `19456` KiB, iterations `2`, parallelism `1`, a 16-byte random salt, and a 32-byte derived key.
- Preserve `_username`, `_salt`, `_id`, authorization data, external authentication APIs, and existing user passwords.
- Treat malformed `$argon2id$` values as failed credentials; never fall back to the legacy verifier for them.
- Persist a successful legacy migration before issuing a session token.
- Keep the change limited to password storage and its required dependency/vendor updates.

---

### Task 1: Argon2id Password Primitive

**Files:**
- Create: `src/internal/authentication/password.go`
- Create: `src/internal/authentication/password_test.go`
- Modify: `go.mod`
- Modify: `go.sum`
- Regenerate: `vendor/`

**Interfaces:**
- Consumes: existing `SHA256(secret, salt string) string` as the legacy verifier.
- Produces: `hashPassword(password string) (string, error)` and `verifyPassword(password, stored string) (matches bool, legacy bool)`.

- [ ] **Step 1: Write failing primitive tests**

Create `src/internal/authentication/password_test.go` with tests that name the production behavior directly:

```go
package authentication

import (
	"strings"
	"testing"
)

func TestHashPasswordCreatesVerifiableArgon2idPHCString(t *testing.T) {
	encoded, err := hashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(encoded, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Fatalf("unexpected password format: %q", encoded)
	}
	matched, legacy := verifyPassword("correct horse battery staple", encoded)
	if !matched || legacy {
		t.Fatalf("Argon2id password verification = (%v, %v), want (true, false)", matched, legacy)
	}
	matched, legacy = verifyPassword("wrong", encoded)
	if matched || legacy {
		t.Fatalf("wrong password verification = (%v, %v), want (false, false)", matched, legacy)
	}
}

func TestVerifyPasswordRecognizesLegacyVerifier(t *testing.T) {
	stored := SHA256("legacy-password", "ignored-salt")
	matched, legacy := verifyPassword("legacy-password", stored)
	if !matched || !legacy {
		t.Fatalf("legacy verification = (%v, %v), want (true, true)", matched, legacy)
	}
}

func TestVerifyPasswordRejectsMalformedOrExcessiveArgon2id(t *testing.T) {
	values := []string{
		"$argon2id$broken",
		"$argon2id$v=19$m=1048576,t=2,p=1$YWJjZGVmZ2hpamtsbW5vcA$YWJj",
		"$argon2id$v=18$m=19456,t=2,p=1$YWJjZGVmZ2hpamtsbW5vcA$YWJj",
	}
	for _, stored := range values {
		matched, legacy := verifyPassword("password", stored)
		if matched || legacy {
			t.Fatalf("malformed verifier %q returned (%v, %v)", stored, matched, legacy)
		}
	}
}
```

- [ ] **Step 2: Run the primitive tests and confirm RED**

Run:

```bash
GOTOOLCHAIN=go1.27.0 go test -mod=vendor ./src/internal/authentication
```

Expected: compilation fails because `hashPassword` and `verifyPassword` do not exist.

- [ ] **Step 3: Add the Argon2 dependency and minimal implementation**

Run:

```bash
GOTOOLCHAIN=go1.27.0 go get golang.org/x/crypto@v0.55.0
```

Create `src/internal/authentication/password.go`:

```go
package authentication

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	passwordMemory      uint32 = 19456
	passwordIterations  uint32 = 2
	passwordParallelism uint8  = 1
	passwordSaltLength         = 16
	passwordKeyLength   uint32 = 32
)

func hashPassword(password string) (string, error) {
	salt := make([]byte, passwordSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, passwordIterations, passwordMemory, passwordParallelism, passwordKeyLength)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		passwordMemory,
		passwordIterations,
		passwordParallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func verifyPassword(password, stored string) (matches bool, legacy bool) {
	if !strings.HasPrefix(stored, "$argon2id$") {
		legacyHash := SHA256(password, "")
		return subtle.ConstantTimeCompare([]byte(legacyHash), []byte(stored)) == 1, true
	}

	parts := strings.Split(stored, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" || parts[3] != "m=19456,t=2,p=1" {
		return false, false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != passwordSaltLength {
		return false, false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expected) != int(passwordKeyLength) {
		return false, false
	}
	actual := argon2.IDKey([]byte(password), salt, passwordIterations, passwordMemory, passwordParallelism, passwordKeyLength)
	return subtle.ConstantTimeCompare(actual, expected) == 1, false
}
```

Regenerate the dependency snapshot:

```bash
GOTOOLCHAIN=go1.27.0 go mod tidy
GOTOOLCHAIN=go1.27.0 go mod vendor
```

- [ ] **Step 4: Run primitive tests and confirm GREEN**

Run:

```bash
GOTOOLCHAIN=go1.27.0 go test -mod=vendor ./src/internal/authentication
```

Expected: all password primitive tests pass.

- [ ] **Step 5: Commit the password primitive**

```bash
git add docs/superpowers/plans/2026-08-25-password-hashing-migration.md src/internal/authentication/password.go src/internal/authentication/password_test.go go.mod go.sum vendor
git commit -m "feat: add Argon2id password verification"
```

### Task 2: Authentication Database Migration

**Files:**
- Modify: `src/internal/authentication/authentication.go:128-240`
- Modify: `src/internal/authentication/authentication.go:386-410`
- Modify: `src/internal/authentication/authentication.go:546-556`
- Modify: `src/internal/authentication/password_test.go`

**Interfaces:**
- Consumes: `hashPassword(password string) (string, error)` and `verifyPassword(password, stored string) (bool, bool)` from Task 1.
- Produces: creation, login migration, and password-change flows that persist Argon2id in the existing `_password` field.

- [ ] **Step 1: Add failing integration tests**

Append tests that use the real JSON database and package functions:

```go
func initAuthenticationTest(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	data = make(map[string]interface{})
	tokens = make(map[string]interface{})
	initAuthentication = false
	if err := Init(filepath.Join(root, "config"), 60); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(root, databaseFile)
}

func TestCreateNewUserStoresArgon2idPassword(t *testing.T) {
	initAuthenticationTest(t)
	userID, err := CreateNewUser("new-user", "new-password")
	if err != nil {
		t.Fatal(err)
	}
	stored := data["users"].(map[string]interface{})[userID].(map[string]interface{})["_password"].(string)
	if !strings.HasPrefix(stored, "$argon2id$") {
		t.Fatalf("new password stored as %q", stored)
	}
}

func TestSuccessfulLegacyLoginMigratesPasswordAndPersists(t *testing.T) {
	databasePath := initAuthenticationTest(t)
	userID, err := CreateNewUser("legacy-user", "temporary")
	if err != nil {
		t.Fatal(err)
	}
	user := data["users"].(map[string]interface{})[userID].(map[string]interface{})
	user["_password"] = SHA256("legacy-password", user["_salt"].(string))
	if err := saveDatabase(data); err != nil {
		t.Fatal(err)
	}

	if _, err := UserAuthentication("legacy-user", "legacy-password"); err != nil {
		t.Fatal(err)
	}
	persisted, err := os.ReadFile(databasePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(persisted, []byte(`"_password": "$argon2id$`)) {
		t.Fatalf("legacy verifier was not migrated: %s", persisted)
	}

	data = make(map[string]interface{})
	if err := loadDatabase(); err != nil {
		t.Fatal(err)
	}
	if _, err := UserAuthentication("legacy-user", "legacy-password"); err != nil {
		t.Fatalf("migrated password failed after reload: %v", err)
	}
}

func TestIncorrectLegacyLoginDoesNotMigratePassword(t *testing.T) {
	databasePath := initAuthenticationTest(t)
	userID, err := CreateNewUser("legacy-user", "temporary")
	if err != nil {
		t.Fatal(err)
	}
	user := data["users"].(map[string]interface{})[userID].(map[string]interface{})
	legacy := SHA256("legacy-password", user["_salt"].(string))
	user["_password"] = legacy
	if err := saveDatabase(data); err != nil {
		t.Fatal(err)
	}
	if _, err := UserAuthentication("legacy-user", "wrong-password"); err == nil {
		t.Fatal("incorrect legacy password authenticated")
	}
	persisted, err := os.ReadFile(databasePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(persisted, []byte(legacy)) {
		t.Fatal("incorrect login modified the legacy verifier")
	}
}

func TestLegacyMigrationSaveFailureDoesNotIssueToken(t *testing.T) {
	initAuthenticationTest(t)
	userID, err := CreateNewUser("legacy-user", "temporary")
	if err != nil {
		t.Fatal(err)
	}
	user := data["users"].(map[string]interface{})[userID].(map[string]interface{})
	legacy := SHA256("legacy-password", user["_salt"].(string))
	user["_password"] = legacy
	database = filepath.Join(t.TempDir(), "missing-directory", databaseFile)

	token, err := UserAuthentication("legacy-user", "legacy-password")
	if err == nil || token != "" {
		t.Fatalf("migration persistence failure returned token %q and error %v", token, err)
	}
	if user["_password"] != legacy {
		t.Fatal("failed migration changed the in-memory legacy verifier")
	}
}

func TestChangeCredentialsStoresArgon2idPassword(t *testing.T) {
	initAuthenticationTest(t)
	userID, err := CreateNewUser("user", "old-password")
	if err != nil {
		t.Fatal(err)
	}
	if err := ChangeCredentials(userID, "", "changed-password"); err != nil {
		t.Fatal(err)
	}
	stored := data["users"].(map[string]interface{})[userID].(map[string]interface{})["_password"].(string)
	if !strings.HasPrefix(stored, "$argon2id$") {
		t.Fatalf("changed password stored as %q", stored)
	}
}
```

Add `bytes`, `os`, and `path/filepath` to the test imports.

- [ ] **Step 2: Run integration tests and confirm RED**

Run:

```bash
GOTOOLCHAIN=go1.27.0 go test -mod=vendor ./src/internal/authentication
```

Expected: new and changed password assertions fail because those paths still call `SHA256`; the legacy migration assertion fails because login leaves the old verifier unchanged.

- [ ] **Step 3: Route creation and password changes through Argon2id**

Change `defaultsForNewUser` to return an error and hash the password before building the map:

```go
func defaultsForNewUser(username, password string) (map[string]interface{}, error) {
	passwordHash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	defaults := make(map[string]interface{})
	salt := randomString(saltLength)
	defaults["_username"] = SHA256(username, salt)
	defaults["_password"] = passwordHash
	defaults["_salt"] = salt
	defaults["_id"] = "id-" + randomID(idLength)
	defaults["data"] = make(map[string]interface{})
	return defaults, nil
}
```

Update `CreateDefaultUser` and `CreateNewUser` to check the returned error before inserting the user. In `ChangeCredentials`, replace the password assignment with:

```go
if len(password) > 0 {
	passwordHash, hashErr := hashPassword(password)
	if hashErr != nil {
		return hashErr
	}
	userData.(map[string]interface{})["_password"] = passwordHash
}
```

Set the `hash` value to `argon2id` only in both new-empty-database default maps.

- [ ] **Step 4: Migrate a valid legacy verifier before issuing a token**

Replace the password comparison inside `UserAuthentication` with this flow:

```go
if SHA256(username, salt) != loginUsername {
	return createError(010)
}
matched, legacy := verifyPassword(password, loginPassword)
if !matched {
	return createError(010)
}
if legacy {
	migrated, hashErr := hashPassword(password)
	if hashErr != nil {
		return hashErr
	}
	loginData["_password"] = migrated
	if saveErr := saveDatabase(data); saveErr != nil {
		loginData["_password"] = loginPassword
		return saveErr
	}
}
return nil
```

Keep token creation after this login helper returns `nil`, ensuring persistence precedes token issuance.

- [ ] **Step 5: Run authentication and full tests and confirm GREEN**

Run:

```bash
GOTOOLCHAIN=go1.27.0 go test -mod=vendor ./src/internal/authentication
GOTOOLCHAIN=go1.27.0 go test -mod=vendor ./...
GOTOOLCHAIN=go1.27.0 go test -race -mod=vendor ./src/internal/authentication
```

Expected: all tests pass and the legacy reload test authenticates with the persisted Argon2id verifier.

- [ ] **Step 6: Commit the authentication migration**

```bash
git add src/internal/authentication/authentication.go src/internal/authentication/password_test.go
git commit -m "feat: migrate passwords to Argon2id"
```

### Task 3: Password Migration Verification

**Files:**
- Verify only; no source changes expected.

**Interfaces:**
- Consumes: completed password primitive and authentication migration.
- Produces: evidence that the password-storage acceptance criteria hold across the project.

- [ ] **Step 1: Verify formatting, static checks, dependencies, and vulnerabilities**

Run:

```bash
gofmt -w src/internal/authentication/authentication.go src/internal/authentication/password.go src/internal/authentication/password_test.go
git diff --check -- . ':(exclude)vendor/**'
GOTOOLCHAIN=go1.27.0 go vet -mod=vendor ./...
GOTOOLCHAIN=go1.27.0 go mod verify
GOTOOLCHAIN=go1.27.0 go run golang.org/x/vuln/cmd/govulncheck@latest -mode=source ./...
```

Expected: every command exits zero and `govulncheck` prints `No vulnerabilities found.`

- [ ] **Step 2: Verify the complete project test surface**

Run:

```bash
GOTOOLCHAIN=go1.27.0 go test -mod=vendor ./...
GOTOOLCHAIN=go1.27.0 go test -race -mod=vendor ./...
```

Expected: all packages pass.

- [ ] **Step 3: Confirm the commit boundary is clean**

Run:

```bash
git status --short
git log -3 --oneline
```

Expected: no uncommitted authentication changes remain, and the two password commits are visible after the design commit.
