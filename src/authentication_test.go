package src

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"threadfin/src/internal/authentication"
)

func TestCheckAuthorizationLevelPreservesUserDataPersistenceError(t *testing.T) {
	root := t.TempDir()
	if err := authentication.Init(filepath.Join(root, "config"), 60); err != nil {
		t.Fatal(err)
	}
	if _, err := authentication.CreateNewUser("user", "password"); err != nil {
		t.Fatal(err)
	}
	token, err := authentication.UserAuthentication("user", "password")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(root, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}

	err = checkAuthorizationLevel(token, "missing.permission")
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("checkAuthorizationLevel() error = %v, want persistence path error", err)
	}
}
