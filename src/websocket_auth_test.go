package src

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"threadfin/src/internal/authentication"

	"github.com/gorilla/websocket"
)

func TestWebSocketOriginAllowed(t *testing.T) {
	restorePersistentState(t)

	tests := []struct {
		name            string
		origin          string
		directTLS       bool
		configuredHTTPS bool
		want            bool
	}{
		{name: "absent", want: true},
		{name: "exact HTTP", origin: "http://threadfin.example:34400", want: true},
		{name: "exact HTTPS from direct TLS", origin: "https://threadfin.example:34400", directTLS: true, want: true},
		{name: "exact HTTPS from configuration", origin: "https://THREADFIN.EXAMPLE:34400", configuredHTTPS: true, want: true},
		{name: "wrong scheme", origin: "https://threadfin.example:34400"},
		{name: "unsupported scheme", origin: "ws://threadfin.example:34400"},
		{name: "wrong port", origin: "http://threadfin.example:34401"},
		{name: "userinfo", origin: "http://user@threadfin.example:34400"},
		{name: "path", origin: "http://threadfin.example:34400/web"},
		{name: "query", origin: "http://threadfin.example:34400?source=web"},
		{name: "fragment", origin: "http://threadfin.example:34400#web"},
		{name: "malformed", origin: "http://threadfin.example:34400/%zz"},
		{name: "host prefix", origin: "http://evilthreadfin.example:34400"},
		{name: "host suffix", origin: "http://threadfin.example.evil:34400"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			systemMutex.Lock()
			Settings = SettingsStruct{}
			System.ServerProtocol.WEB = "http"
			if test.configuredHTTPS {
				Settings.ForceHttps = true
				Settings.HttpsThreadfinDomain = "threadfin.example"
			}
			systemMutex.Unlock()

			request := httptest.NewRequest(http.MethodGet, "http://threadfin.example:34400/data/", nil)
			if test.directTLS {
				request.TLS = &tls.ConnectionState{}
			}
			if test.origin != "" {
				request.Header.Set("Origin", test.origin)
			}

			if got := webSocketOriginAllowed(request); got != test.want {
				t.Fatalf("webSocketOriginAllowed(%q) = %v, want %v", test.origin, got, test.want)
			}
		})
	}
}

func TestWebSocketAuthenticationBeforeUpgrade(t *testing.T) {
	t.Run("valid session cookie", func(t *testing.T) {
		restorePersistentState(t)
		sessionID, _, _ := initializeWebSocketAuthentication(t, 60, true)
		configureWebSocketAuthentication(true, false)
		server, webSocketURL := newWebSocketTestServer(t)

		headers := http.Header{"Origin": []string{server.URL}}
		headers.Add("Cookie", (&http.Cookie{Name: authentication.BrowserSessionCookieName, Value: sessionID}).String())
		conn := dialWebSocket(t, webSocketURL, headers)
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("missing credentials", func(t *testing.T) {
		restorePersistentState(t)
		initializeWebSocketAuthentication(t, 60, true)
		configureWebSocketAuthentication(true, false)
		server, webSocketURL := newWebSocketTestServer(t)

		expectWebSocketHandshakeFailure(t, webSocketURL, http.Header{"Origin": []string{server.URL}}, http.StatusUnauthorized)
	})

	t.Run("expired session", func(t *testing.T) {
		restorePersistentState(t)
		sessionID, _, _ := initializeWebSocketAuthentication(t, 0, true)
		configureWebSocketAuthentication(true, false)
		server, webSocketURL := newWebSocketTestServer(t)

		headers := http.Header{"Origin": []string{server.URL}}
		headers.Add("Cookie", (&http.Cookie{Name: authentication.BrowserSessionCookieName, Value: sessionID}).String())
		expectWebSocketHandshakeFailure(t, webSocketURL, headers, http.StatusUnauthorized)
	})

	t.Run("permission denied session", func(t *testing.T) {
		restorePersistentState(t)
		sessionID, _, _ := initializeWebSocketAuthentication(t, 60, false)
		configureWebSocketAuthentication(true, false)
		server, webSocketURL := newWebSocketTestServer(t)

		headers := http.Header{"Origin": []string{server.URL}}
		headers.Add("Cookie", (&http.Cookie{Name: authentication.BrowserSessionCookieName, Value: sessionID}).String())
		expectWebSocketHandshakeFailure(t, webSocketURL, headers, http.StatusUnauthorized)
	})

	t.Run("invalid session does not fall back to token", func(t *testing.T) {
		restorePersistentState(t)
		_, token, _ := initializeWebSocketAuthentication(t, 60, true)
		configureWebSocketAuthentication(true, false)
		server, webSocketURL := newWebSocketTestServer(t)

		headers := http.Header{"Origin": []string{server.URL}}
		headers.Add("Cookie", (&http.Cookie{Name: authentication.BrowserSessionCookieName, Value: "invalid-session"}).String())
		expectWebSocketHandshakeFailure(t, webSocketURL+"?Token="+url.QueryEscape(token), headers, http.StatusUnauthorized)
		if _, err := authentication.GetUserID(token); err != nil {
			t.Fatalf("rejected fallback rotated the query token: %v", err)
		}
	})

	t.Run("malformed session cookie does not fall back to token", func(t *testing.T) {
		restorePersistentState(t)
		_, token, _ := initializeWebSocketAuthentication(t, 60, true)
		configureWebSocketAuthentication(true, false)
		server, webSocketURL := newWebSocketTestServer(t)

		headers := http.Header{"Origin": []string{server.URL}}
		headers.Set("Cookie", authentication.BrowserSessionCookieName+`=bad\value`)
		expectWebSocketHandshakeFailure(t, webSocketURL+"?Token="+url.QueryEscape(token), headers, http.StatusUnauthorized)
		if _, err := authentication.GetUserID(token); err != nil {
			t.Fatalf("malformed session cookie rotated the fallback token: %v", err)
		}
	})

	t.Run("comma session cookie does not fall back to token", func(t *testing.T) {
		restorePersistentState(t)
		_, token, _ := initializeWebSocketAuthentication(t, 60, true)
		configureWebSocketAuthentication(true, false)
		server, webSocketURL := newWebSocketTestServer(t)

		headers := http.Header{"Origin": []string{server.URL}}
		headers.Set("Cookie", authentication.BrowserSessionCookieName+"=bad,value")
		expectWebSocketHandshakeFailure(t, webSocketURL+"?Token="+url.QueryEscape(token), headers, http.StatusUnauthorized)
		if _, err := authentication.GetUserID(token); err != nil {
			t.Fatalf("comma session cookie rotated the fallback token: %v", err)
		}
	})

	t.Run("similarly named malformed cookie permits token", func(t *testing.T) {
		restorePersistentState(t)
		_, token, _ := initializeWebSocketAuthentication(t, 60, true)
		configureWebSocketAuthentication(true, false)
		server, webSocketURL := newWebSocketTestServer(t)

		headers := http.Header{"Origin": []string{server.URL}}
		headers.Set("Cookie", "Not"+authentication.BrowserSessionCookieName+`=bad\value`)
		conn := dialWebSocket(t, webSocketURL+"?Token="+url.QueryEscape(token), headers)
		if err := conn.WriteJSON(RequestStruct{Cmd: "getServerConfig"}); err != nil {
			t.Fatal(err)
		}
		var response ResponseStruct
		if err := conn.ReadJSON(&response); err != nil {
			t.Fatal(err)
		}
		if response.Token == "" || response.Token == token {
			t.Fatalf("similarly named cookie blocked legacy token rotation: %q", response.Token)
		}
	})

	t.Run("permission denied legacy token", func(t *testing.T) {
		restorePersistentState(t)
		_, token, _ := initializeWebSocketAuthentication(t, 60, false)
		configureWebSocketAuthentication(true, false)
		server, webSocketURL := newWebSocketTestServer(t)

		expectWebSocketHandshakeFailure(t, webSocketURL+"?Token="+url.QueryEscape(token), http.Header{"Origin": []string{server.URL}}, http.StatusUnauthorized)
	})

	for _, credential := range []string{"session cookie", "legacy token"} {
		t.Run("cross origin rejects "+credential, func(t *testing.T) {
			restorePersistentState(t)
			sessionID, token, _ := initializeWebSocketAuthentication(t, 60, true)
			configureWebSocketAuthentication(true, false)
			_, webSocketURL := newWebSocketTestServer(t)
			headers := http.Header{"Origin": []string{"http://threadfin.example.evil:34400"}}
			if credential == "session cookie" {
				headers.Add("Cookie", (&http.Cookie{Name: authentication.BrowserSessionCookieName, Value: sessionID}).String())
			} else {
				webSocketURL += "?Token=" + url.QueryEscape(token)
			}

			expectWebSocketHandshakeFailure(t, webSocketURL, headers, http.StatusForbidden)
			if credential == "legacy token" {
				if _, err := authentication.GetUserID(token); err != nil {
					t.Fatalf("cross-origin rejection rotated the query token: %v", err)
				}
			}
		})
	}

	for _, mode := range []struct {
		name           string
		authentication bool
		setup          bool
	}{
		{name: "authentication disabled", authentication: false},
		{name: "configuration setup", authentication: true, setup: true},
	} {
		t.Run(mode.name+" exact origin", func(t *testing.T) {
			restorePersistentState(t)
			initializeWebSocketAuthentication(t, 60, true)
			configureWebSocketAuthentication(mode.authentication, mode.setup)
			server, webSocketURL := newWebSocketTestServer(t)

			conn := dialWebSocket(t, webSocketURL, http.Header{"Origin": []string{server.URL}})
			if err := conn.Close(); err != nil {
				t.Fatal(err)
			}
		})

		t.Run(mode.name+" cross origin", func(t *testing.T) {
			restorePersistentState(t)
			initializeWebSocketAuthentication(t, 60, true)
			configureWebSocketAuthentication(mode.authentication, mode.setup)
			_, webSocketURL := newWebSocketTestServer(t)

			expectWebSocketHandshakeFailure(t, webSocketURL, http.Header{"Origin": []string{"http://threadfin.example.evil:34400"}}, http.StatusForbidden)
		})
	}

	for _, withOrigin := range []bool{true, false} {
		name := "no Origin legacy access"
		if withOrigin {
			name = "same-origin legacy token rotation"
		}
		t.Run(name, func(t *testing.T) {
			restorePersistentState(t)
			_, token, _ := initializeWebSocketAuthentication(t, 60, true)
			configureWebSocketAuthentication(true, false)
			server, webSocketURL := newWebSocketTestServer(t)
			headers := make(http.Header)
			if withOrigin {
				headers.Set("Origin", server.URL)
			}

			conn := dialWebSocket(t, webSocketURL+"?Token="+url.QueryEscape(token), headers)
			if err := conn.WriteJSON(RequestStruct{Cmd: "getServerConfig"}); err != nil {
				t.Fatal(err)
			}
			var response ResponseStruct
			if err := conn.ReadJSON(&response); err != nil {
				t.Fatal(err)
			}
			if !response.Status || response.Token == "" || response.Token == token {
				t.Fatalf("legacy response = status %v token %q, want successful rotation", response.Status, response.Token)
			}
			if err := checkAuthorizationLevel(response.Token, "authentication.web"); err != nil {
				t.Fatalf("rotated token lacks web permission: %v", err)
			}
			if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
				t.Fatal(err)
			}
			if err := conn.ReadJSON(&response); err == nil {
				t.Fatal("legacy query-token connection returned more than one response")
			}
			if err := conn.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRequestStructRequestIDJSON(t *testing.T) {
	encoded, err := json.Marshal(RequestStruct{Cmd: "getServerConfig", RequestID: "request-1"})
	if err != nil {
		t.Fatal(err)
	}
	var message map[string]interface{}
	if err := json.Unmarshal(encoded, &message); err != nil {
		t.Fatal(err)
	}
	if got := message["requestId"]; got != "request-1" {
		t.Fatalf("requestId = %#v, want request-1 in %s", got, encoded)
	}
	if _, exists := message["RequestID"]; exists {
		t.Fatalf("request used Go field spelling: %s", encoded)
	}

	legacy, err := json.Marshal(RequestStruct{Cmd: "getServerConfig"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(legacy), "requestId") {
		t.Fatalf("legacy request contains requestId: %s", legacy)
	}
}

func TestResponseStructRequestIDJSON(t *testing.T) {
	encoded, err := json.Marshal(ResponseStruct{Status: true, RequestID: "request-1"})
	if err != nil {
		t.Fatal(err)
	}
	var message map[string]interface{}
	if err := json.Unmarshal(encoded, &message); err != nil {
		t.Fatal(err)
	}
	if got := message["requestId"]; got != "request-1" {
		t.Fatalf("requestId = %#v, want request-1 in %s", got, encoded)
	}
	if _, exists := message["RequestID"]; exists {
		t.Fatalf("response used Go field spelling: %s", encoded)
	}

	legacy, err := json.Marshal(ResponseStruct{Status: true})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(legacy), "requestId") {
		t.Fatalf("legacy response contains requestId: %s", legacy)
	}
}

func TestWebSocketPersistentRequestIDs(t *testing.T) {
	restorePersistentState(t)
	sessionID, _, _ := initializeWebSocketAuthentication(t, 60, true)
	configureWebSocketAuthentication(true, false)
	server, webSocketURL := newWebSocketTestServer(t)
	header := http.Header{"Origin": []string{server.URL}}
	header.Add("Cookie", (&http.Cookie{Name: authentication.BrowserSessionCookieName, Value: sessionID}).String())
	conn := dialWebSocket(t, webSocketURL, header)
	defer conn.Close()

	for _, request := range []RequestStruct{
		{Cmd: "updateLog", RequestID: "request-1"},
		{Cmd: "getServerConfig", RequestID: "request-2"},
	} {
		if err := conn.WriteJSON(request); err != nil {
			t.Fatal(err)
		}
		var response ResponseStruct
		if err := conn.ReadJSON(&response); err != nil {
			t.Fatal(err)
		}
		if !response.Status || response.RequestID != request.RequestID {
			t.Fatalf("response = status %v requestId %q, want true %q", response.Status, response.RequestID, request.RequestID)
		}
	}
}

func TestWebSocketPersistentMessagesResetRequestAndResponse(t *testing.T) {
	restorePersistentState(t)
	sessionID, _, _ := initializeWebSocketAuthentication(t, 60, true)
	configureWebSocketAuthentication(true, false)
	systemMutex.Lock()
	System.Folder.ImagesUpload = t.TempDir() + string(filepath.Separator)
	System.ServerProtocol.XML = "http"
	systemMutex.Unlock()
	server, webSocketURL := newWebSocketTestServer(t)
	header := http.Header{"Origin": []string{server.URL}}
	header.Add("Cookie", (&http.Cookie{Name: authentication.BrowserSessionCookieName, Value: sessionID}).String())
	conn := dialWebSocket(t, webSocketURL, header)

	if err := conn.WriteJSON(RequestStruct{
		Cmd:       "uploadLogo",
		RequestID: "request-with-logo",
		Base64:    "data:application/octet-stream;base64,YQ==",
		Filename:  "logo.png",
	}); err != nil {
		t.Fatal(err)
	}
	var first ResponseStruct
	if err := conn.ReadJSON(&first); err != nil {
		t.Fatal(err)
	}
	if first.RequestID != "request-with-logo" || first.LogoURL == "" {
		t.Fatalf("first response = requestId %q logoURL %q, want populated values", first.RequestID, first.LogoURL)
	}

	if err := conn.WriteJSON(RequestStruct{Cmd: "uploadLogo", RequestID: "request-without-logo"}); err != nil {
		t.Fatal(err)
	}
	var second ResponseStruct
	if err := conn.ReadJSON(&second); err != nil {
		t.Fatal(err)
	}
	if second.RequestID != "request-without-logo" || second.LogoURL != "" {
		t.Fatalf("second response = requestId %q logoURL %q, want reset optional request and response fields", second.RequestID, second.LogoURL)
	}
}

func TestWebSocketRevokedSessionClosesWithPolicyViolation(t *testing.T) {
	restorePersistentState(t)
	previousWebScreenLog := WebScreenLog
	t.Cleanup(func() { WebScreenLog = previousWebScreenLog })
	sessionID, _, userID := initializeWebSocketAuthentication(t, 60, true)
	configureWebSocketAuthentication(true, false)
	server, webSocketURL := newWebSocketTestServer(t)
	header := http.Header{"Origin": []string{server.URL}}
	header.Add("Cookie", (&http.Cookie{Name: authentication.BrowserSessionCookieName, Value: sessionID}).String())
	conn := dialWebSocket(t, webSocketURL, header)
	defer conn.Close()
	WebScreenLog.Log = []string{"must remain"}
	WebScreenLog.Errors = 2
	WebScreenLog.Warnings = 3

	if err := authentication.WriteUserData(userID, map[string]interface{}{"authentication.web": false}); err != nil {
		t.Fatal(err)
	}
	if err := conn.WriteJSON(RequestStruct{Cmd: "resetLogs", RequestID: "revoked-request"}); err != nil {
		t.Fatal(err)
	}
	_, _, err := conn.ReadMessage()
	var closeError *websocket.CloseError
	if !errors.As(err, &closeError) {
		t.Fatalf("revoked session read error = %v, want WebSocket close", err)
	}
	if closeError.Code != websocket.ClosePolicyViolation {
		t.Fatalf("revoked session close code = %d, want %d", closeError.Code, websocket.ClosePolicyViolation)
	}
	if len(WebScreenLog.Log) != 1 || WebScreenLog.Log[0] != "must remain" || WebScreenLog.Errors != 2 || WebScreenLog.Warnings != 3 {
		t.Fatalf("revoked command executed: log=%v errors=%d warnings=%d", WebScreenLog.Log, WebScreenLog.Errors, WebScreenLog.Warnings)
	}
}

func TestWebSocketServerClosesConnections(t *testing.T) {
	previousWebScreenLog := WebScreenLog
	t.Cleanup(func() { WebScreenLog = previousWebScreenLog })

	t.Run("legacy response exit", func(t *testing.T) {
		restorePersistentState(t)
		_, token, _ := initializeWebSocketAuthentication(t, 60, true)
		configureWebSocketAuthentication(true, false)
		server, webSocketURL, listener, handlerDone := newTrackedWebSocketTestServer(t)
		conn := dialWebSocket(t, webSocketURL+"?Token="+url.QueryEscape(token), http.Header{"Origin": []string{server.URL}})
		serverConn := listener.nextConnection(t)

		if err := conn.WriteJSON(RequestStruct{Cmd: "getServerConfig"}); err != nil {
			t.Fatal(err)
		}
		var response ResponseStruct
		if err := conn.ReadJSON(&response); err != nil {
			t.Fatal(err)
		}
		waitWebSocketHandler(t, handlerDone)
		serverConn.requireClosed(t)
	})

	t.Run("revocation exit", func(t *testing.T) {
		restorePersistentState(t)
		sessionID, _, userID := initializeWebSocketAuthentication(t, 60, true)
		configureWebSocketAuthentication(true, false)
		server, webSocketURL, listener, handlerDone := newTrackedWebSocketTestServer(t)
		header := http.Header{"Origin": []string{server.URL}}
		header.Add("Cookie", (&http.Cookie{Name: authentication.BrowserSessionCookieName, Value: sessionID}).String())
		conn := dialWebSocket(t, webSocketURL, header)
		serverConn := listener.nextConnection(t)

		if err := authentication.WriteUserData(userID, map[string]interface{}{"authentication.web": false}); err != nil {
			t.Fatal(err)
		}
		if err := conn.WriteJSON(RequestStruct{Cmd: "getServerConfig"}); err != nil {
			t.Fatal(err)
		}
		_, _, err := conn.ReadMessage()
		var closeError *websocket.CloseError
		if !errors.As(err, &closeError) || closeError.Code != websocket.ClosePolicyViolation {
			t.Fatalf("revocation read error = %v, want close code %d", err, websocket.ClosePolicyViolation)
		}
		waitWebSocketHandler(t, handlerDone)
		serverConn.requireClosed(t)
	})

	t.Run("read exit", func(t *testing.T) {
		restorePersistentState(t)
		initializeWebSocketAuthentication(t, 60, true)
		configureWebSocketAuthentication(false, false)
		server, webSocketURL, listener, handlerDone := newTrackedWebSocketTestServer(t)
		conn := dialWebSocket(t, webSocketURL, http.Header{"Origin": []string{server.URL}})
		serverConn := listener.nextConnection(t)

		if err := conn.WriteMessage(websocket.TextMessage, []byte("{")); err != nil {
			t.Fatal(err)
		}
		waitWebSocketHandler(t, handlerDone)
		serverConn.requireClosed(t)
	})

	t.Run("write exit", func(t *testing.T) {
		restorePersistentState(t)
		initializeWebSocketAuthentication(t, 60, true)
		configureWebSocketAuthentication(false, false)
		server, webSocketURL, listener, handlerDone := newTrackedWebSocketTestServer(t)
		conn := dialWebSocket(t, webSocketURL, http.Header{"Origin": []string{server.URL}})
		serverConn := listener.nextConnection(t)
		serverConn.failWrites.Store(true)

		if err := conn.WriteJSON(RequestStruct{Cmd: "getServerConfig"}); err != nil {
			t.Fatal(err)
		}
		waitWebSocketHandler(t, handlerDone)
		serverConn.requireClosed(t)
	})
}

func initializeWebSocketAuthentication(t *testing.T, validityMinutes int, webPermission bool) (sessionID, token, userID string) {
	t.Helper()
	if err := authentication.Init(filepath.Join(t.TempDir(), "config"), validityMinutes); err != nil {
		t.Fatal(err)
	}
	userID, err := authentication.CreateNewUser("websocket-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	if err := authentication.WriteUserData(userID, map[string]interface{}{"authentication.web": webPermission}); err != nil {
		t.Fatal(err)
	}
	sessionID, _, err = authentication.AuthenticateBrowser("websocket-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	token, err = authentication.UserAuthentication("websocket-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	return sessionID, token, userID
}

func configureWebSocketAuthentication(authenticationRequired, configurationWizard bool) {
	systemMutex.Lock()
	Settings = SettingsStruct{AuthenticationWEB: authenticationRequired}
	System.ConfigurationWizard = configurationWizard
	System.ServerProtocol.WEB = "http"
	systemMutex.Unlock()
}

func newWebSocketTestServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	handlerDone := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(handlerDone)
		WS(w, r)
	}))
	t.Cleanup(func() {
		server.Close()
		select {
		case <-handlerDone:
		case <-time.After(5 * time.Second):
			t.Error("WebSocket test handler did not stop")
		}
	})
	return server, "ws" + strings.TrimPrefix(server.URL, "http") + "/data/"
}

var errTrackedWebSocketWrite = errors.New("tracked WebSocket write failure")

type trackedWebSocketConn struct {
	net.Conn
	closed     chan struct{}
	closeOnce  sync.Once
	failWrites atomic.Bool
}

func (conn *trackedWebSocketConn) Write(payload []byte) (int, error) {
	if conn.failWrites.Load() {
		return 0, errTrackedWebSocketWrite
	}
	return conn.Conn.Write(payload)
}

func (conn *trackedWebSocketConn) Close() (err error) {
	conn.closeOnce.Do(func() {
		close(conn.closed)
		err = conn.Conn.Close()
	})
	return err
}

func (conn *trackedWebSocketConn) requireClosed(t *testing.T) {
	t.Helper()
	select {
	case <-conn.closed:
	default:
		t.Fatal("WebSocket handler returned without closing the server connection")
	}
}

type trackedWebSocketListener struct {
	net.Listener
	accepted chan *trackedWebSocketConn
}

func (listener *trackedWebSocketListener) Accept() (net.Conn, error) {
	conn, err := listener.Listener.Accept()
	if err != nil {
		return nil, err
	}
	tracked := &trackedWebSocketConn{Conn: conn, closed: make(chan struct{})}
	listener.accepted <- tracked
	return tracked, nil
}

func (listener *trackedWebSocketListener) nextConnection(t *testing.T) *trackedWebSocketConn {
	t.Helper()
	select {
	case conn := <-listener.accepted:
		return conn
	case <-time.After(5 * time.Second):
		t.Fatal("tracked WebSocket server accepted no connection")
		return nil
	}
}

func newTrackedWebSocketTestServer(t *testing.T) (*httptest.Server, string, *trackedWebSocketListener, <-chan struct{}) {
	t.Helper()
	handlerDone := make(chan struct{})
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(handlerDone)
		WS(w, r)
	}))
	listener := &trackedWebSocketListener{Listener: server.Listener, accepted: make(chan *trackedWebSocketConn, 1)}
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)
	return server, "ws" + strings.TrimPrefix(server.URL, "http") + "/data/", listener, handlerDone
}

func waitWebSocketHandler(t *testing.T, handlerDone <-chan struct{}) {
	t.Helper()
	select {
	case <-handlerDone:
	case <-time.After(5 * time.Second):
		t.Fatal("WebSocket handler did not return")
	}
}

func dialWebSocket(t *testing.T, webSocketURL string, headers http.Header) *websocket.Conn {
	t.Helper()
	conn, response, err := websocket.DefaultDialer.Dial(webSocketURL, headers)
	if err != nil {
		if response != nil {
			_ = response.Body.Close()
		}
		t.Fatalf("WebSocket dial failed: %v", err)
	}
	if response != nil {
		_ = response.Body.Close()
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func expectWebSocketHandshakeFailure(t *testing.T, webSocketURL string, headers http.Header, wantStatus int) {
	t.Helper()
	conn, response, err := websocket.DefaultDialer.Dial(webSocketURL, headers)
	if conn != nil {
		_ = conn.Close()
		t.Fatal("unauthorized WebSocket request upgraded successfully")
	}
	if err == nil {
		t.Fatal("unauthorized WebSocket request returned no error")
	}
	if response == nil {
		t.Fatalf("handshake error %v returned no HTTP response", err)
	}
	defer response.Body.Close()
	if response.StatusCode != wantStatus {
		t.Fatalf("handshake status = %d, want %d (%v)", response.StatusCode, wantStatus, err)
	}
	if response.Header.Get("Sec-WebSocket-Accept") != "" {
		t.Fatalf("failed handshake included Sec-WebSocket-Accept: %q", response.Header.Get("Sec-WebSocket-Accept"))
	}
}
