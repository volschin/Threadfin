package src

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"threadfin/src/internal/authentication"
)

type webSocketAuthentication struct {
	browserSessionID string
	legacyToken      string
	persistent       bool
}

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

	systemMutex.Lock()
	configuredProtocol := System.ServerProtocol.WEB
	configuredHTTPS := Settings.ForceHttps && Settings.HttpsThreadfinDomain != ""
	systemMutex.Unlock()

	effectiveScheme := configuredProtocol
	if r.TLS != nil || configuredHTTPS {
		effectiveScheme = "https"
	}
	if !strings.EqualFold(effectiveScheme, "http") && !strings.EqualFold(effectiveScheme, "https") {
		return false
	}

	return strings.EqualFold(parsedOrigin.Scheme, effectiveScheme) && strings.EqualFold(parsedOrigin.Host, r.Host)
}

func authenticateWebSocketRequest(r *http.Request) (webSocketAuthentication, error) {
	systemMutex.Lock()
	authenticationRequired := Settings.AuthenticationWEB
	configurationWizard := System.ConfigurationWizard
	systemMutex.Unlock()

	if !authenticationRequired || configurationWizard {
		return webSocketAuthentication{persistent: true}, nil
	}

	sessionCookiePresent := webSocketCookieNamePresent(r, authentication.BrowserSessionCookieName)
	cookie, err := r.Cookie(authentication.BrowserSessionCookieName)
	if err == nil {
		if _, err = authorizeBrowserRequest(r, "authentication.web"); err != nil {
			return webSocketAuthentication{}, err
		}
		return webSocketAuthentication{browserSessionID: cookie.Value, persistent: true}, nil
	}
	if sessionCookiePresent {
		return webSocketAuthentication{}, errors.New("websocket authentication failed")
	}
	if !errors.Is(err, http.ErrNoCookie) {
		return webSocketAuthentication{}, err
	}

	legacyToken, err := tokenAuthentication(r.URL.Query().Get("Token"))
	if err != nil {
		return webSocketAuthentication{}, err
	}
	if legacyToken == "" {
		return webSocketAuthentication{}, errors.New("websocket authentication failed")
	}
	if err = checkAuthorizationLevel(legacyToken, "authentication.web"); err != nil {
		return webSocketAuthentication{}, err
	}

	return webSocketAuthentication{legacyToken: legacyToken}, nil
}

func webSocketCookieNamePresent(r *http.Request, cookieName string) bool {
	for _, header := range r.Header.Values("Cookie") {
		for part := range strings.SplitSeq(header, ";") {
			name, _, _ := strings.Cut(strings.TrimSpace(part), "=")
			if strings.TrimSpace(name) == cookieName {
				return true
			}
		}
	}
	return false
}
