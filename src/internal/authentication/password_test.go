package authentication

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestCreateDefaultUserReturnsPersistenceError(t *testing.T) {
	initAuthenticationTest(t)
	database = filepath.Join(t.TempDir(), "missing-directory", databaseFile)

	err := CreateDefaultUser("default-user", "password")
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("CreateDefaultUser() error = %v, want persistence path error", err)
	}
	if got := len(data["users"].(map[string]interface{})); got != 0 {
		t.Fatalf("users after failed CreateDefaultUser() = %d, want 0", got)
	}
}

func TestCreateNewUserReturnsPersistenceError(t *testing.T) {
	initAuthenticationTest(t)
	database = filepath.Join(t.TempDir(), "missing-directory", databaseFile)

	userID, err := CreateNewUser("new-user", "password")
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("CreateNewUser() error = %v, want persistence path error", err)
	}
	if userID != "" {
		t.Fatalf("CreateNewUser() userID = %q after persistence failure, want empty", userID)
	}
	if got := len(data["users"].(map[string]interface{})); got != 0 {
		t.Fatalf("users after failed CreateNewUser() = %d, want 0", got)
	}
}

func TestGetAllUserDataReturnsInitializationPersistenceError(t *testing.T) {
	initAuthenticationTest(t)
	data = make(map[string]interface{})
	database = filepath.Join(t.TempDir(), "missing-directory", databaseFile)

	users, err := GetAllUserData()
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("GetAllUserData() error = %v, want persistence path error", err)
	}
	if users != nil {
		t.Fatalf("GetAllUserData() users = %#v after persistence failure, want nil", users)
	}
}

func TestRemoveUserRollsBackPersistenceFailure(t *testing.T) {
	initAuthenticationTest(t)
	userID, err := CreateNewUser("remove-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	database = filepath.Join(t.TempDir(), "missing-directory", databaseFile)

	err = RemoveUser(userID)
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("RemoveUser() error = %v, want persistence path error", err)
	}
	if _, ok := data["users"].(map[string]interface{})[userID]; !ok {
		t.Fatal("RemoveUser() discarded user from memory after persistence failure")
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

func TestLegacyMigrationSaveFailureWithMultipleUsersReturnsPersistenceError(t *testing.T) {
	initAuthenticationTest(t)
	userID, err := CreateNewUser("legacy-user", "temporary")
	if err != nil {
		t.Fatal(err)
	}
	user := data["users"].(map[string]interface{})[userID].(map[string]interface{})
	legacy := SHA256("legacy-password", user["_salt"].(string))
	user["_password"] = legacy

	users := data["users"].(map[string]interface{})
	for i := 0; i < 1024; i++ {
		otherUser := make(map[string]interface{}, len(user))
		for key, value := range user {
			otherUser[key] = value
		}
		otherUser["_id"] = fmt.Sprintf("other-user-%d", i)
		otherUser["_username"] = SHA256("other-user", user["_salt"].(string))
		users[otherUser["_id"].(string)] = otherUser
	}
	database = filepath.Join(t.TempDir(), "missing-directory", databaseFile)

	token, err := UserAuthentication("legacy-user", "legacy-password")
	if token != "" {
		t.Fatalf("migration persistence failure returned token %q", token)
	}
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("migration persistence failure returned %v, want filesystem persistence error", err)
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

func TestHashPasswordPreservesExactPHCFields(t *testing.T) {
	hash, err := hashPassword("benchmark-password")
	if err != nil {
		t.Fatal(err)
	}

	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Fatalf("PHC field count = %d, want 6: %q", len(parts), hash)
	}
	wantPrefix := []string{"", "argon2id", "v=19", "m=19456,t=2,p=1"}
	for i, want := range wantPrefix {
		if parts[i] != want {
			t.Fatalf("PHC field %d = %q, want %q", i, parts[i], want)
		}
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		t.Fatalf("decode salt: %v", err)
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		t.Fatalf("decode key: %v", err)
	}
	if len(salt) != 16 || len(key) != int(passwordKeyLength) {
		t.Fatalf("decoded lengths = salt:%d key:%d, want salt:16 key:%d", len(salt), len(key), passwordKeyLength)
	}
	matches, legacy := verifyPassword("benchmark-password", hash)
	if !matches || legacy {
		t.Fatalf("verify generated hash = matches:%t legacy:%t, want true false", matches, legacy)
	}
}

func TestSerializeArgon2IDHashMatchesCurrentFormat(t *testing.T) {
	salt := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
	}
	key := []byte{
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
		0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
		0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27,
		0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
	}
	const want = "$argon2id$v=19$m=19456,t=2,p=1$AAECAwQFBgcICQoLDA0ODw$EBESExQVFhcYGRobHB0eHyAhIiMkJSYnKCkqKywtLi8"
	if got := serializeArgon2IDHash(salt, key); got != want {
		t.Fatalf("serializeArgon2IDHash() = %q, want %q", got, want)
	}
}

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
