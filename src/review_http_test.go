package src

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"threadfin/src/internal/authentication"

	"github.com/gorilla/websocket"
)

func TestBackupHEADPreservesArchiveAndGETCleansItUp(t *testing.T) {
	restorePersistentState(t)
	Settings = SettingsStruct{}
	System.Folder.Temp = t.TempDir()
	const name = "threadfin_backup_20261007_1100.zip"
	filename := filepath.Join(System.Folder.Temp, name)
	if err := os.WriteFile(filename, []byte("archive"), 0600); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	Download(w, httptest.NewRequest(http.MethodHead, "/download/"+name, nil))
	if w.Code != 200 || w.Body.Len() != 0 {
		t.Fatalf("HEAD: %d %q", w.Code, w.Body.String())
	}
	if _, err := os.Stat(filename); err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	Download(w, httptest.NewRequest(http.MethodGet, "/download/"+name, nil))
	if w.Code != 200 || w.Body.String() != "archive" {
		t.Fatalf("GET: %d %q", w.Code, w.Body.String())
	}
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		t.Fatalf("downloaded archive not cleaned up: %v", err)
	}
}

func TestStreamHTTPLogsDoNotExposeURLCredentials(t *testing.T) {
	restorePersistentState(t)
	previousOutput, previousScreen := log.Writer(), WebScreenLog
	var output bytes.Buffer
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previousOutput); WebScreenLog = previousScreen })
	System.Flag.Info = false
	Settings = SettingsStruct{}
	Settings.Files.M3U = map[string]interface{}{"provider": map[string]interface{}{"buffer": "-"}}
	for _, source := range []string{
		"https://secret-user:secret-password@provider.test/live/path-user/path-password/1?token=query-secret#fragment-secret",
		"rtsp://secret-user:secret-password@provider.test/live/path-user/path-password/1?token=query-secret",
	} {
		Data.Cache.StreamingURLS = map[string]StreamInfo{"channel": {URL: source, PlaylistID: "provider"}}
		r := httptest.NewRequest(http.MethodGet, "/stream/channel", nil)
		w := httptest.NewRecorder()
		Stream(w, r)
		if w.Code != http.StatusFound || w.Header().Get("Location") != source {
			t.Errorf("redirect changed provider address: %d %s", w.Code, w.Header().Get("Location"))
		}
	}
	Data.Cache.StreamingURLS = map[string]StreamInfo{"channel": {URL: "http://secret-user:secret-password@provider.test/live/path-user/path-password/%zz?token=query-secret", PlaylistID: "provider"}}
	Stream(httptest.NewRecorder(), httptest.NewRequest(http.MethodHead, "/stream/channel", nil))
	logged := output.String() + mapToJSON(WebScreenLog)
	for _, secret := range []string{"secret-user", "secret-password", "path-user", "path-password", "query-secret", "fragment-secret"} {
		if strings.Contains(logged, secret) {
			t.Errorf("stream credential leaked to logs: %s", secret)
		}
	}
}

func TestBackupDownloadRequiresAdministrativeSession(t *testing.T) {
	restorePersistentState(t)
	Settings = SettingsStruct{AuthenticationWEB: true}
	System.Folder.Temp = t.TempDir() + string(os.PathSeparator)
	session, token, userID := initializeWebSocketAuthentication(t, 60, true)
	name := "threadfin_backup_20261007_1000.zip"
	filename := filepath.Join(System.Folder.Temp, name)
	for _, credential := range []string{"anonymous", "api-token-cookie", "invalid-session", "valid-session"} {
		t.Run(credential, func(t *testing.T) {
			if err := os.WriteFile(filename, []byte("private archive"), 0600); err != nil {
				t.Fatal(err)
			}
			r := httptest.NewRequest(http.MethodGet, "/download/"+name, nil)
			switch credential {
			case "api-token-cookie":
				r.AddCookie(&http.Cookie{Name: "Token", Value: token})
			case "invalid-session":
				r.AddCookie(&http.Cookie{Name: authentication.BrowserSessionCookieName, Value: "invalid"})
			case "valid-session":
				r.AddCookie(&http.Cookie{Name: authentication.BrowserSessionCookieName, Value: session})
			}
			w := httptest.NewRecorder()
			Download(w, r)
			if credential == "valid-session" {
				if w.Code != http.StatusOK || w.Body.String() != "private archive" {
					t.Fatalf("authorized download: %d %q", w.Code, w.Body.String())
				}
			} else {
				if w.Code != http.StatusForbidden || strings.Contains(w.Body.String(), "private archive") {
					t.Errorf("unauthorized download: %d", w.Code)
				}
				if _, err := os.Stat(filename); err != nil {
					t.Errorf("unauthorized request removed archive: %v", err)
				}
			}
		})
	}
	if err := authentication.WriteUserData(userID, map[string]interface{}{"authentication.web": false}); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodGet, "/download/"+name, nil)
	r.AddCookie(&http.Cookie{Name: authentication.BrowserSessionCookieName, Value: session})
	w := httptest.NewRecorder()
	Download(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatalf("revoked permission got %d", w.Code)
	}
}

func TestXMLTVOnlyServesPublishedGuideArtifacts(t *testing.T) {
	setupM3UDeliveryTest(t)
	Settings.AuthenticationM3U = true
	System.File.XML = filepath.Join(System.Folder.Data, "threadfin.xml")
	System.Compressed.GZxml = filepath.Join(System.Folder.Data, "threadfin.xml.gz")
	for _, name := range []string{"threadfin.xml", "threadfin.xml.gz", "threadfin.m3u", "provider.m3u", "provider.xml"} {
		if err := os.WriteFile(filepath.Join(System.Folder.Data, name), []byte("artifact "+name), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, target := range []string{"/xmltv/threadfin.m3u", "/xmltv/provider.m3u", "/xmltv/provider.xml", "/xmltv/nested/threadfin.xml", "/xmltv/%2e%2e/threadfin.m3u", "/xmltv/m3u/threadfin.m3u"} {
		w := requestThreadfinM3U(t, http.MethodGet, target, nil)
		if w.Code != http.StatusNotFound || strings.Contains(w.Body.String(), "artifact") {
			t.Errorf("%s: %d %q", target, w.Code, w.Body.String())
		}
	}
	for _, name := range []string{"threadfin.xml", "threadfin.xml.gz"} {
		w := requestThreadfinM3U(t, http.MethodGet, "/xmltv/"+name, nil)
		if w.Code != http.StatusOK || w.Body.String() != "artifact "+name {
			t.Errorf("published %s: %d %q", name, w.Code, w.Body.String())
		}
	}
}

func TestPPVRequiresPOSTAndAdministrativeAuthorization(t *testing.T) {
	for _, enable := range []bool{true, false} {
		for _, test := range []struct {
			name, method               string
			auth, session, crossOrigin bool
			want                       int
		}{
			{"anonymous GET", http.MethodGet, true, false, false, http.StatusMethodNotAllowed},
			{"anonymous POST", http.MethodPost, true, false, false, http.StatusForbidden},
			{"authenticated GET", http.MethodGet, true, true, false, http.StatusMethodNotAllowed},
			{"cross-origin POST", http.MethodPost, true, true, true, http.StatusForbidden},
			{"authenticated POST", http.MethodPost, true, true, false, http.StatusOK},
			{"auth disabled POST", http.MethodPost, false, false, false, http.StatusOK},
		} {
			t.Run(test.name+map[bool]string{true: " enable", false: " disable"}[enable], func(t *testing.T) {
				setupIsolatedConfigDomainState(t)
				Settings.AuthenticationWEB = test.auth
				System.ScanInProgress = 1
				mapping := map[string]interface{}{"ppv": map[string]interface{}{"x-mapping": "PPV", "x-active": !enable}}
				if err := saveMapToJSONFile(System.File.XEPG, mapping); err != nil {
					t.Fatal(err)
				}
				r := httptest.NewRequest(test.method, "http://threadfin.test/ppv/enable", nil)
				if test.session {
					session, _, _ := initializeWebSocketAuthentication(t, 60, true)
					r.AddCookie(&http.Cookie{Name: authentication.BrowserSessionCookieName, Value: session})
				}
				if test.crossOrigin {
					r.Header.Set("Origin", "https://attacker.test")
				}
				w := httptest.NewRecorder()
				if enable {
					enablePPV(w, r)
				} else {
					disablePPV(w, r)
				}
				if w.Code != test.want {
					t.Errorf("status %d, want %d", w.Code, test.want)
				}
				stored, err := loadJSONFileToMap(System.File.XEPG)
				if err != nil {
					t.Fatal(err)
				}
				wantActive := !enable
				if test.want == http.StatusOK {
					wantActive = enable
				}
				if stored["ppv"].(map[string]interface{})["x-active"] != wantActive {
					t.Error("unexpected persisted PPV state")
				}
			})
		}
	}
}

func TestLegacyAPIErrorIsSingleJSONResponse(t *testing.T) {
	restorePersistentState(t)
	Settings = SettingsStruct{API: true}
	w := httptest.NewRecorder()
	API(w, httptest.NewRequest(http.MethodPost, "/api/", strings.NewReader(`{"cmd":"unsupported"}`)))
	var result APIResponseStruct
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON response: %q: %v", w.Body.String(), err)
	}
	if result.Status || result.Error == "" {
		t.Fatalf("error response = %#v", result)
	}
}

func TestLegacyAPICommandErrorReturnsRotatedToken(t *testing.T) {
	restorePersistentState(t)
	Settings = SettingsStruct{API: true, AuthenticationAPI: true}
	System.ConfigurationWizard = false
	_, token, userID := initializeWebSocketAuthentication(t, 60, true)
	if err := authentication.WriteUserData(userID, map[string]interface{}{"authentication.api": true}); err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(APIRequestStruct{Cmd: "unsupported", Token: token})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	API(w, httptest.NewRequest(http.MethodPost, "/api/", bytes.NewReader(body)))
	var result APIResponseStruct
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status || result.Error == "" || result.Token == "" || result.Token == token {
		t.Fatalf("command failure did not return its rotated token (status=%t, error=%q, token present=%t)", result.Status, result.Error, result.Token != "")
	}
	if _, _, err := authentication.AuthorizeTokenPermissions(result.Token, "authentication.api"); err != nil {
		t.Fatalf("returned token cannot continue session: %v", err)
	}
}

func TestRequestHostsDoNotChangePublicationAddress(t *testing.T) {
	for _, handler := range []struct {
		name, target string
		serve        http.HandlerFunc
	}{
		{"index", "/discover.json", Index}, {"playlist", "/xmltv/missing.xml", Threadfin},
		{"web", "/web/", Web}, {"api", "/api/", API},
	} {
		t.Run(handler.name, func(t *testing.T) {
			setupIsolatedConfigDomainState(t)
			Settings.API = true
			Settings.EpgSource = "XEPG"
			setGlobalDomain("published.test:34400")
			before := System.Addresses
			r := httptest.NewRequest(http.MethodPost, "http://attacker.test"+handler.target, strings.NewReader(`{"cmd":"status"}`))
			w := httptest.NewRecorder()
			handler.serve(w, r)
			if System.Domain != "published.test:34400" || System.Addresses != before {
				t.Errorf("request host changed publication address to %q", System.Domain)
			}
			url, err := createStreamingURL("M3U", "provider", "1", "channel", "http://provider.test/1", nil, nil, nil)
			if err != nil || !strings.Contains(url, "://published.test:34400/stream/") {
				t.Errorf("published stream = %q, %v", url, err)
			}
		})
	}
}

func TestWebSocketHostDoesNotChangePublicationAddress(t *testing.T) {
	setupIsolatedConfigDomainState(t)
	System.Domain = "published.test:34400"
	done := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { WS(w, r); close(done) }))
	defer server.Close()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/data/", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	waitWebSocketHandler(t, done)
	if System.Domain != "published.test:34400" {
		t.Fatalf("WebSocket replaced publication domain: %s", System.Domain)
	}
}

func TestSettingsUpdatePublicationDomainWithoutRequestHost(t *testing.T) {
	setupIsolatedConfigDomainState(t)
	Settings.Port = "34400"
	Settings.EpgSource = "XEPG"
	System.ServerProtocol.M3U = "http"
	var request RequestStruct
	domain := "configured.test:8443"
	request.Settings.HttpThreadfinDomain = &domain
	if _, err := updateServerSettings(request); err != nil {
		t.Fatal(err)
	}
	if System.Domain != domain || System.Addresses.M3U != "http://configured.test:8443/m3u/threadfin.m3u" {
		t.Fatalf("saved domain not published: %s %s", System.Domain, System.Addresses.M3U)
	}
	Settings.ForceHttps = true
	Settings.HttpsThreadfinDomain = "secure.test"
	stream, err := createStreamingURL("M3U", "provider", "1", "Channel", "http://provider.test/stream", nil, nil, nil)
	if err != nil || !strings.HasPrefix(stream, "https://secure.test/stream/") {
		t.Fatalf("HTTPS stream: %q %v", stream, err)
	}
	if System.Domain != domain {
		t.Fatalf("HTTPS stream changed canonical domain: %s", System.Domain)
	}
}
