package src

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const expectedBrowserContentSecurityPolicy = "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; font-src 'self' data: https://cdnjs.cloudflare.com; img-src 'self' data: http: https:; connect-src 'self' ws: wss:; form-action 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'self'"

var expectedBrowserSecurityHeaders = map[string]string{
	"Content-Security-Policy": expectedBrowserContentSecurityPolicy,
	"X-Content-Type-Options":  "nosniff",
	"Referrer-Policy":         "no-referrer",
	"X-Frame-Options":         "DENY",
}

func TestWebBrowserSecurityHeaders(t *testing.T) {
	restorePersistentState(t)
	previousWebUI := webUI
	t.Cleanup(func() { webUI = previousWebUI })

	System = SystemStruct{}
	Settings = SettingsStruct{Language: "en"}
	Settings.Files.M3U = map[string]interface{}{}
	Settings.Files.HDHR = map[string]interface{}{}
	Data = DataStruct{}
	webUI = map[string]interface{}{
		"html/lang/en.json": base64.StdEncoding.EncodeToString([]byte(`{}`)),
		"html/index.html":   base64.StdEncoding.EncodeToString([]byte(`<!doctype html><html><body>Threadfin</body></html>`)),
		"html/js/test.js":   base64.StdEncoding.EncodeToString([]byte(`globalThis.threadfinAsset = true`)),
	}

	t.Run("HTML gets the exact browser policy", func(t *testing.T) {
		response := httptest.NewRecorder()
		Web(response, httptest.NewRequest(http.MethodGet, "/web/", nil))

		if response.Code != http.StatusOK {
			t.Fatalf("HTML status = %d, want %d; body=%q", response.Code, http.StatusOK, response.Body.String())
		}
		for name, want := range expectedBrowserSecurityHeaders {
			if got := response.Header().Get(name); got != want {
				t.Errorf("HTML %s = %q, want %q", name, got, want)
			}
		}
	})

	t.Run("JavaScript asset keeps its existing contract", func(t *testing.T) {
		response := httptest.NewRecorder()
		Web(response, httptest.NewRequest(http.MethodGet, "/web/js/test.js", nil))

		if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/javascript" {
			t.Fatalf("JavaScript response = status %d content type %q", response.Code, response.Header().Get("Content-Type"))
		}
		assertNoBrowserSecurityHeaders(t, response.Header())
	})

	t.Run("JSON API keeps its existing contract", func(t *testing.T) {
		Settings.API = true
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/", strings.NewReader(`{"cmd":"status"}`))
		request.Header.Set("Content-Type", "application/json")
		API(response, request)

		if response.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("API Content-Type = %q, want application/json", response.Header().Get("Content-Type"))
		}
		assertNoBrowserSecurityHeaders(t, response.Header())
	})

	t.Run("stream redirect keeps its existing contract", func(t *testing.T) {
		const streamID = "0123456789abcdef0123456789abcdef"
		Settings.Files.M3U = map[string]interface{}{
			"M1": map[string]interface{}{"buffer": "-"},
		}
		Data.Cache.StreamingURLS = map[string]StreamInfo{
			streamID: {
				Name:       "Security header fixture",
				PlaylistID: "M1",
				URL:        "http://provider.invalid/watch",
				URLid:      streamID,
			},
		}
		response := httptest.NewRecorder()
		Stream(response, httptest.NewRequest(http.MethodGet, "/stream/"+streamID, nil))

		if response.Code != http.StatusFound {
			t.Fatalf("stream status = %d, want %d; body=%q", response.Code, http.StatusFound, response.Body.String())
		}
		assertNoBrowserSecurityHeaders(t, response.Header())
	})
}

func assertNoBrowserSecurityHeaders(t *testing.T, header http.Header) {
	t.Helper()
	for name := range expectedBrowserSecurityHeaders {
		if got := header.Get(name); got != "" {
			t.Errorf("non-HTML %s = %q, want absent", name, got)
		}
	}
}
