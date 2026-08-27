package authentication

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestAuthenticateBrowserCreatesIndependentAuthorizedSessions(t *testing.T) {
	initAuthenticationTest(t)
	userID, err := CreateNewUser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteUserData(userID, map[string]interface{}{"authentication.web": true}); err != nil {
		t.Fatal(err)
	}

	first, firstUser, err := AuthenticateBrowser("browser-user", "password", "authentication.web")
	second, secondUser, err := AuthenticateBrowser("browser-user", "password", "authentication.web")
	if err != nil || first == second || firstUser != userID || secondUser != userID {
		t.Fatalf("sessions/users = %q/%q %q/%q: %v", first, firstUser, second, secondUser, err)
	}
	if _, err := AuthorizeBrowserSession(first, "authentication.web"); err != nil {
		t.Fatal(err)
	}
}

func TestAuthenticateBrowserRejectsWrongPasswordAndMissingPermission(t *testing.T) {
	initAuthenticationTest(t)
	_, err := CreateNewUser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}

	if sessionID, _, err := AuthenticateBrowser("browser-user", "wrong-password"); err == nil || sessionID != "" {
		t.Fatalf("wrong password = session %q, error %v", sessionID, err)
	}
	if sessionID, _, err := AuthenticateBrowser("browser-user", "password", "authentication.web"); err == nil || sessionID != "" {
		t.Fatalf("missing permission = session %q, error %v", sessionID, err)
	}
}

func TestAuthorizeBrowserSessionRemovesExpiredSession(t *testing.T) {
	initAuthenticationTest(t)
	userID, err := CreateNewUser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	sessionID, _, err := AuthenticateBrowser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	browserSessions[sessionID] = browserSession{userID: userID, expires: time.Now().Add(-time.Second)}

	if _, err := AuthorizeBrowserSession(sessionID); err == nil {
		t.Fatal("expired browser session authorized")
	}
	if _, ok := browserSessions[sessionID]; ok {
		t.Fatal("expired browser session remained stored")
	}
}

func TestInvalidateBrowserSessionOnlyInvalidatesSpecifiedSession(t *testing.T) {
	initAuthenticationTest(t)
	userID, err := CreateNewUser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	first, _, err := AuthenticateBrowser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := AuthenticateBrowser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}

	InvalidateBrowserSession(first)
	if _, err := AuthorizeBrowserSession(first); err == nil {
		t.Fatal("invalidated browser session authorized")
	}
	if got, err := AuthorizeBrowserSession(second); err != nil || got != userID {
		t.Fatalf("remaining session = user %q, error %v", got, err)
	}
}

func TestRemoveUserInvalidatesBrowserSessions(t *testing.T) {
	initAuthenticationTest(t)
	userID, err := CreateNewUser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	sessionID, _, err := AuthenticateBrowser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}

	if err := RemoveUser(userID); err != nil {
		t.Fatal(err)
	}
	if _, err := AuthorizeBrowserSession(sessionID); err == nil {
		t.Fatal("removed user's browser session authorized")
	}
}

func TestChangeCredentialsInvalidatesBrowserSessionsAfterPasswordChange(t *testing.T) {
	initAuthenticationTest(t)
	userID, err := CreateNewUser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	sessionID, _, err := AuthenticateBrowser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}

	if err := ChangeCredentials(userID, "", "new-password"); err != nil {
		t.Fatal(err)
	}
	if _, err := AuthorizeBrowserSession(sessionID); err == nil {
		t.Fatal("browser session survived password change")
	}
}

func TestChangeCredentialsPreservesBrowserSessionsForUsernameOnlyChange(t *testing.T) {
	initAuthenticationTest(t)
	userID, err := CreateNewUser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	sessionID, _, err := AuthenticateBrowser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}

	if err := ChangeCredentials(userID, "renamed-browser-user", ""); err != nil {
		t.Fatal(err)
	}
	if got, err := AuthorizeBrowserSession(sessionID); err != nil || got != userID {
		t.Fatalf("username-only change session = user %q, error %v", got, err)
	}
}

func TestBrowserSessionsAreIndependentFromAPITokenRotation(t *testing.T) {
	initAuthenticationTest(t)
	userID, err := CreateNewUser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteUserData(userID, map[string]interface{}{"authentication.web": true}); err != nil {
		t.Fatal(err)
	}
	sessionID, _, err := AuthenticateBrowser("browser-user", "password", "authentication.web")
	if err != nil {
		t.Fatal(err)
	}
	token, err := UserAuthentication("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := AuthorizeTokenPermissions(token, "authentication.web"); err != nil {
		t.Fatal(err)
	}

	if got, err := AuthorizeBrowserSession(sessionID, "authentication.web"); err != nil || got != userID {
		t.Fatalf("browser session after API rotation = user %q, error %v", got, err)
	}
}

func TestInitInvalidatesExistingBrowserSessions(t *testing.T) {
	databasePath := initAuthenticationTest(t)
	if _, err := CreateNewUser("browser-user", "password"); err != nil {
		t.Fatal(err)
	}
	sessionID, _, err := AuthenticateBrowser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}

	if err := Init(databasePath, 60); err != nil {
		t.Fatal(err)
	}
	if _, err := AuthorizeBrowserSession(sessionID); err == nil {
		t.Fatal("browser session survived Init")
	}
}

func TestAuthorizeBrowserSessionRejectsRevokedPermission(t *testing.T) {
	initAuthenticationTest(t)
	userID, err := CreateNewUser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteUserData(userID, map[string]interface{}{"authentication.web": true}); err != nil {
		t.Fatal(err)
	}
	sessionID, _, err := AuthenticateBrowser("browser-user", "password", "authentication.web")
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteUserData(userID, map[string]interface{}{"authentication.web": false}); err != nil {
		t.Fatal(err)
	}
	if _, exists := browserSessions[sessionID]; !exists {
		t.Fatal("permission change unexpectedly removed browser session")
	}

	if _, err := AuthorizeBrowserSession(sessionID, "authentication.web"); err == nil {
		t.Fatal("browser session authorized after permission revocation")
	}
}

func TestBrowserSessionCookieUsesStoredExpiryAndStrictAttributes(t *testing.T) {
	initAuthenticationTest(t)
	if _, err := CreateNewUser("browser-user", "password"); err != nil {
		t.Fatal(err)
	}
	sessionID, _, err := AuthenticateBrowser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}

	cookie, err := BrowserSessionCookie(sessionID, true)
	if err != nil {
		t.Fatal(err)
	}
	if cookie.Name != "ThreadfinSession" || cookie.Value != sessionID || cookie.Path != "/" || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || !cookie.Secure {
		t.Fatalf("browser cookie = %#v", cookie)
	}
	if !cookie.Expires.Equal(browserSessions[sessionID].expires) {
		t.Fatalf("cookie expiry = %s, want stored expiry %s", cookie.Expires, browserSessions[sessionID].expires)
	}

	cookie, err = BrowserSessionCookie(sessionID, false)
	if err != nil {
		t.Fatal(err)
	}
	if cookie.Secure {
		t.Fatal("insecure browser cookie has Secure set")
	}
}

func TestExpiredBrowserSessionCookieDeletesWithStrictAttributes(t *testing.T) {
	cookie := ExpiredBrowserSessionCookie(true)
	if cookie.Name != "ThreadfinSession" || cookie.Value != "" || cookie.Path != "/" || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || !cookie.Secure {
		t.Fatalf("expired browser cookie = %#v", cookie)
	}
	if !cookie.Expires.Equal(time.Unix(1, 0)) || cookie.MaxAge >= 0 {
		t.Fatalf("expired browser cookie expiry/max age = %s/%d", cookie.Expires, cookie.MaxAge)
	}
}

func TestAuthenticateBrowserRejectsRandomFailureWithoutCreatingSession(t *testing.T) {
	initAuthenticationTest(t)
	if _, err := CreateNewUser("browser-user", "password"); err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("random failure")
	previous := readBrowserSessionRandom
	readBrowserSessionRandom = func([]byte) (int, error) { return 0, sentinel }
	t.Cleanup(func() { readBrowserSessionRandom = previous })

	sessionID, _, err := AuthenticateBrowser("browser-user", "password")
	if !errors.Is(err, sentinel) || sessionID != "" {
		t.Fatalf("random failure = session %q, error %v", sessionID, err)
	}
	if len(browserSessions) != 0 {
		t.Fatalf("sessions after random failure = %d, want 0", len(browserSessions))
	}
}
