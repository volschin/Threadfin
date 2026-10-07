package authentication

import (
	"path/filepath"
	"testing"
)

func TestPasswordChangeRevokesOnlyThatUsersAPITokens(t *testing.T) {
	initAuthenticationTest(t)
	id, err := CreateNewUser("changed", "old-password")
	if err != nil {
		t.Fatal(err)
	}
	otherID, err := CreateNewUser("other", "password")
	if err != nil {
		t.Fatal(err)
	}
	for _, userID := range []string{id, otherID} {
		if err := WriteUserData(userID, map[string]interface{}{"authentication.api": true}); err != nil {
			t.Fatal(err)
		}
	}
	first, err := UserAuthentication("changed", "old-password")
	if err != nil {
		t.Fatal(err)
	}
	second, err := UserAuthentication("changed", "old-password")
	if err != nil {
		t.Fatal(err)
	}
	other, err := UserAuthentication("other", "password")
	if err != nil {
		t.Fatal(err)
	}
	if err := ChangeCredentials(id, "", "new-password"); err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{first, second} {
		if _, _, err := AuthorizeTokenPermissions(token, "authentication.api"); err == nil {
			t.Error("pre-change API token still authorizes and rotates")
		}
		if _, err := CheckTheValidityOfTheToken(token); err == nil {
			t.Error("legacy token renewal survived password change")
		}
	}
	if _, _, err := AuthorizeTokenPermissions(other, "authentication.api"); err != nil {
		t.Fatal(err)
	}
	if _, err := UserAuthentication("changed", "new-password"); err != nil {
		t.Fatal(err)
	}
}

func TestFailedPasswordChangePreservesCredentialsAndTokens(t *testing.T) {
	initAuthenticationTest(t)
	id, err := CreateNewUser("unchanged", "old-password")
	if err != nil {
		t.Fatal(err)
	}
	token, err := UserAuthentication("unchanged", "old-password")
	if err != nil {
		t.Fatal(err)
	}
	session, _, err := AuthenticateBrowser("unchanged", "old-password")
	if err != nil {
		t.Fatal(err)
	}
	previousDatabase := database
	database = filepath.Join(t.TempDir(), "missing", "auth.json")
	if err := ChangeCredentials(id, "renamed", "new-password"); err == nil {
		t.Fatal("expected persistence failure")
	}
	database = previousDatabase
	if _, err := CheckTheValidityOfTheToken(token); err != nil {
		t.Fatal(err)
	}
	if _, err := AuthorizeBrowserSession(session); err != nil {
		t.Fatal(err)
	}
	if _, err := UserAuthentication("unchanged", "old-password"); err != nil {
		t.Fatalf("failed save changed in-memory credentials: %v", err)
	}
}
