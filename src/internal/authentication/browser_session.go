package authentication

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"time"
)

const BrowserSessionCookieName = "ThreadfinSession"

type browserSession struct {
	userID  string
	expires time.Time
}

var browserSessions = make(map[string]browserSession)
var readBrowserSessionRandom = rand.Read

// AuthenticateBrowser verifies credentials and creates a browser-only session.
func AuthenticateBrowser(username, password string, permissions ...string) (sessionID, userID string, err error) {
	authenticationMutex.Lock()
	defer authenticationMutex.Unlock()

	if err = checkInit(); err != nil {
		return "", "", err
	}
	if userID, err = authenticateUserLocked(username, password); err != nil {
		return "", "", err
	}
	if !userHasPermissionsLocked(userID, permissions...) {
		return "", "", createError(010)
	}
	if sessionID, err = newBrowserSessionID(); err != nil {
		return "", "", err
	}
	browserSessions[sessionID] = browserSession{
		userID:  userID,
		expires: time.Now().Local().Add(time.Minute * time.Duration(tokenValidity)),
	}
	return sessionID, userID, nil
}

// AuthorizeBrowserSession validates a browser session and requested permissions.
func AuthorizeBrowserSession(sessionID string, permissions ...string) (userID string, err error) {
	authenticationMutex.Lock()
	defer authenticationMutex.Unlock()

	if err = checkInit(); err != nil {
		return "", err
	}
	session, ok := browserSessions[sessionID]
	if !ok || !session.expires.After(time.Now().Local()) {
		delete(browserSessions, sessionID)
		return "", createError(011)
	}
	if !userHasPermissionsLocked(session.userID, permissions...) {
		return "", createError(011)
	}
	return session.userID, nil
}

// InvalidateBrowserSession removes one browser session.
func InvalidateBrowserSession(sessionID string) {
	authenticationMutex.Lock()
	defer authenticationMutex.Unlock()
	delete(browserSessions, sessionID)
}

// BrowserSessionCookie returns the live cookie for a browser session.
func BrowserSessionCookie(sessionID string, secure bool) (*http.Cookie, error) {
	authenticationMutex.Lock()
	defer authenticationMutex.Unlock()

	if err := checkInit(); err != nil {
		return nil, err
	}
	session, ok := browserSessions[sessionID]
	if !ok || !session.expires.After(time.Now().Local()) {
		delete(browserSessions, sessionID)
		return nil, createError(011)
	}
	return &http.Cookie{
		Name:     BrowserSessionCookieName,
		Value:    sessionID,
		Path:     "/",
		Expires:  session.expires,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   secure,
	}, nil
}

// ExpiredBrowserSessionCookie returns a browser session deletion cookie.
func ExpiredBrowserSessionCookie(secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     BrowserSessionCookieName,
		Path:     "/",
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   secure,
	}
}

func newBrowserSessionID() (string, error) {
	for {
		bytes := make([]byte, 32)
		read, err := readBrowserSessionRandom(bytes)
		if err != nil {
			return "", err
		}
		if read != len(bytes) {
			return "", io.ErrUnexpectedEOF
		}
		sessionID := base64.RawURLEncoding.EncodeToString(bytes)
		if _, exists := browserSessions[sessionID]; !exists {
			return sessionID, nil
		}
	}
}

func userHasPermissionsLocked(userID string, permissions ...string) bool {
	user, ok := data["users"].(map[string]interface{})[userID].(map[string]interface{})
	if !ok {
		return false
	}
	userData, ok := user["data"].(map[string]interface{})
	if !ok {
		return false
	}
	for _, permission := range permissions {
		granted, exists := userData[permission].(bool)
		if !exists || !granted {
			return false
		}
	}
	return true
}

func invalidateUserBrowserSessionsLocked(userID string) {
	for sessionID, session := range browserSessions {
		if session.userID == userID {
			delete(browserSessions, sessionID)
		}
	}
}
