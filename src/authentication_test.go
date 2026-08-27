package src

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"threadfin/src/internal/authentication"
)

func TestBrowserCookieSecureUsesDirectTLSOrConfiguredHTTPS(t *testing.T) {
	restorePersistentState(t)
	request := httptest.NewRequest(http.MethodGet, "http://threadfin.example/web/", nil)
	request.Header.Set("X-Forwarded-Proto", "https")
	if browserCookieSecure(request) {
		t.Fatal("forwarded HTTPS header selected Secure cookie")
	}

	systemMutex.Lock()
	Settings = SettingsStruct{ForceHttps: true}
	systemMutex.Unlock()
	if browserCookieSecure(request) {
		t.Fatal("force HTTPS without configured domain selected Secure cookie")
	}
	systemMutex.Lock()
	Settings.HttpsThreadfinDomain = "threadfin.example"
	systemMutex.Unlock()
	if !browserCookieSecure(request) {
		t.Fatal("configured HTTPS domain did not select Secure cookie")
	}

	systemMutex.Lock()
	Settings = SettingsStruct{}
	systemMutex.Unlock()
	request.TLS = &tls.ConnectionState{}
	if !browserCookieSecure(request) {
		t.Fatal("direct TLS did not select Secure cookie")
	}
}

func TestWebLogoutInvalidatesOnlyPresentedBrowserSession(t *testing.T) {
	restorePersistentState(t)
	Settings = SettingsStruct{AuthenticationWEB: true}
	if err := authentication.Init(filepath.Join(t.TempDir(), "config"), 60); err != nil {
		t.Fatal(err)
	}
	userID, err := authentication.CreateNewUser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	if err := authentication.WriteUserData(userID, map[string]interface{}{"authentication.web": true}); err != nil {
		t.Fatal(err)
	}
	first, _, err := authentication.AuthenticateBrowser("browser-user", "password", "authentication.web")
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := authentication.AuthenticateBrowser("browser-user", "password", "authentication.web")
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/web/logout", nil)
	request.AddCookie(&http.Cookie{Name: authentication.BrowserSessionCookieName, Value: first})
	response := httptest.NewRecorder()
	WebLogout(response, request)
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/web/" {
		t.Fatalf("logout = status %d location %q, want 303 /web/", response.Code, response.Header().Get("Location"))
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != authentication.BrowserSessionCookieName || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode || cookies[0].Path != "/" || cookies[0].MaxAge >= 0 {
		t.Fatalf("logout expiry cookie = %#v", cookies)
	}
	if _, err := authentication.AuthorizeBrowserSession(first, "authentication.web"); err == nil {
		t.Fatal("logout left presented session authorized")
	}
	if _, err := authentication.AuthorizeBrowserSession(second, "authentication.web"); err != nil {
		t.Fatalf("logout invalidated a different browser session: %v", err)
	}
}

func TestWebLogoutRejectsNonPOST(t *testing.T) {
	restorePersistentState(t)
	if err := authentication.Init(filepath.Join(t.TempDir(), "config"), 60); err != nil {
		t.Fatal(err)
	}
	userID, err := authentication.CreateNewUser("browser-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	if err := authentication.WriteUserData(userID, map[string]interface{}{"authentication.web": true}); err != nil {
		t.Fatal(err)
	}
	sessionID, _, err := authentication.AuthenticateBrowser("browser-user", "password", "authentication.web")
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/web/logout", nil)
	request.AddCookie(&http.Cookie{Name: authentication.BrowserSessionCookieName, Value: sessionID})
	response := httptest.NewRecorder()
	WebLogout(response, request)
	if response.Code != http.StatusMethodNotAllowed || response.Header().Get("Allow") != http.MethodPost {
		t.Fatalf("GET logout = status %d allow %q, want 405 POST", response.Code, response.Header().Get("Allow"))
	}
	if _, err := authentication.AuthorizeBrowserSession(sessionID, "authentication.web"); err != nil {
		t.Fatalf("non-POST logout invalidated browser session: %v", err)
	}
}

func TestAPIBrowserSessionCompatibility(t *testing.T) {
	restorePersistentState(t)
	System.ConfigurationWizard = false
	Settings = SettingsStruct{API: true, AuthenticationAPI: true}
	if err := authentication.Init(filepath.Join(t.TempDir(), "config"), 60); err != nil {
		t.Fatal(err)
	}
	userID, err := authentication.CreateNewUser("api-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	if err := authentication.WriteUserData(userID, map[string]interface{}{"authentication.api": true}); err != nil {
		t.Fatal(err)
	}

	loginResponse := httptest.NewRecorder()
	API(loginResponse, httptest.NewRequest(http.MethodPost, "/api/", strings.NewReader(`{"cmd":"login","username":"api-user","password":"password"}`)))
	var login struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &login); err != nil {
		t.Fatalf("decode API login: %v\n%s", err, loginResponse.Body.String())
	}
	if loginResponse.Code != http.StatusOK || login.Token == "" || loginResponse.Header().Get("Set-Cookie") != "" {
		t.Fatalf("API login = status %d token %q cookie %q, want token-only response", loginResponse.Code, login.Token, loginResponse.Header().Get("Set-Cookie"))
	}

	statusResponse := httptest.NewRecorder()
	API(statusResponse, httptest.NewRequest(http.MethodPost, "/api/", strings.NewReader(`{"cmd":"status","token":"`+login.Token+`"}`)))
	var status struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(statusResponse.Body.Bytes(), &status); err != nil {
		t.Fatalf("decode API status: %v\n%s", err, statusResponse.Body.String())
	}
	if statusResponse.Code != http.StatusOK || status.Token == "" || status.Token == login.Token || statusResponse.Header().Get("Set-Cookie") != "" {
		t.Fatalf("API status = status %d token %q cookie %q, want rotated token without browser session", statusResponse.Code, status.Token, statusResponse.Header().Get("Set-Cookie"))
	}
}

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
	if _, ok := errors.AsType[*os.PathError](err); !ok {
		t.Fatalf("checkAuthorizationLevel() error = %v, want persistence path error", err)
	}
}

func TestSaveUserDataPreservesRemoveUserPersistenceError(t *testing.T) {
	root := t.TempDir()
	if err := authentication.Init(filepath.Join(root, "config"), 60); err != nil {
		t.Fatal(err)
	}
	userID, err := authentication.CreateNewUser("remove-user", "password")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(root, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}

	request := RequestStruct{UserData: map[string]interface{}{
		userID: map[string]interface{}{"delete": true},
	}}
	err = saveUserData(request)
	if _, ok := errors.AsType[*os.PathError](err); !ok {
		t.Fatalf("saveUserData() error = %v, want persistence path error", err)
	}
}
