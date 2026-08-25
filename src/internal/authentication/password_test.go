package authentication

import (
	"bytes"
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
