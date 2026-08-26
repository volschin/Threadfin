package src

import (
	"html"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStreamRouteContract(t *testing.T) {
	restorePersistentState(t)

	previousWebScreenLog := WebScreenLog
	WebScreenLog = WebScreenLogStruct{}
	t.Cleanup(func() {
		WebScreenLog = previousWebScreenLog
	})

	const (
		knownID   = "0123456789abcdef0123456789abcdef"
		unknownID = "ffffffffffffffffffffffffffffffff"
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Stream-Head", "copied")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)
	streamURL := upstream.URL + "/watch?token=a&mode=b"

	Settings = SettingsStruct{}
	Settings.Files.M3U = map[string]interface{}{
		"playlist": map[string]interface{}{
			"buffer": "-",
		},
	}
	Data.Cache.StreamingURLS = map[string]StreamInfo{
		knownID: {
			Name:       "Contract Channel",
			PlaylistID: "playlist",
			URL:        streamURL,
			URLid:      knownID,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/stream/", Stream)

	tests := []struct {
		name           string
		method         string
		target         string
		wantStatus     int
		wantLocation   string
		wantBody       string
		checkBody      bool
		wantCORS       string
		wantHeadHeader string
	}{
		{
			name:         "known canonical ID with GET",
			method:       http.MethodGet,
			target:       "/stream/" + knownID,
			wantStatus:   http.StatusFound,
			wantLocation: streamURL,
			wantBody:     "<a href=\"" + html.EscapeString(streamURL) + "\">Found</a>.\n\n",
			checkBody:    true,
			wantCORS:     "*",
		},
		{
			name:           "known canonical ID with HEAD",
			method:         http.MethodHead,
			target:         "/stream/" + knownID,
			wantStatus:     http.StatusOK,
			wantHeadHeader: "copied",
		},
		{
			name:         "known canonical ID with POST",
			method:       http.MethodPost,
			target:       "/stream/" + knownID,
			wantStatus:   http.StatusFound,
			wantLocation: streamURL,
			wantCORS:     "*",
		},
		{
			name:         "known canonical ID with unqualified method",
			method:       http.MethodPatch,
			target:       "/stream/" + knownID,
			wantStatus:   http.StatusFound,
			wantLocation: streamURL,
			wantCORS:     "*",
		},
		{
			name:       "unknown ID with GET",
			method:     http.MethodGet,
			target:     "/stream/" + unknownID,
			wantStatus: http.StatusNotFound,
			wantBody:   "Not Found [404]\n",
			checkBody:  true,
		},
		{
			name:       "unknown ID with HEAD",
			method:     http.MethodHead,
			target:     "/stream/" + unknownID,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "unknown ID with POST",
			method:     http.MethodPost,
			target:     "/stream/" + unknownID,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "empty suffix",
			method:     http.MethodGet,
			target:     "/stream/",
			wantStatus: http.StatusNotFound,
			wantBody:   "Not Found [404]\n",
			checkBody:  true,
		},
		{
			name:       "nested suffix",
			method:     http.MethodGet,
			target:     "/stream/" + knownID + "/segment.ts",
			wantStatus: http.StatusNotFound,
			wantBody:   "Not Found [404]\n",
			checkBody:  true,
		},
		{
			name:       "known ID with trailing slash",
			method:     http.MethodGet,
			target:     "/stream/" + knownID + "/",
			wantStatus: http.StatusNotFound,
			wantBody:   "Not Found [404]\n",
			checkBody:  true,
		},
		{
			name:       "escaped known ID",
			method:     http.MethodGet,
			target:     "/stream/%30" + knownID[1:],
			wantStatus: http.StatusNotFound,
			wantBody:   "Not Found [404]\n",
			checkBody:  true,
		},
		{
			name:       "escaped route prefix",
			method:     http.MethodGet,
			target:     "/%73tream/" + knownID,
			wantStatus: http.StatusNotFound,
			wantBody:   "Not Found [404]\n",
			checkBody:  true,
		},
		{
			name:       "escaped slash in suffix",
			method:     http.MethodGet,
			target:     "/stream/not%2Fa-channel",
			wantStatus: http.StatusNotFound,
			wantBody:   "Not Found [404]\n",
			checkBody:  true,
		},
		{
			name:       "query string on known ID",
			method:     http.MethodGet,
			target:     "/stream/" + knownID + "?token=abc",
			wantStatus: http.StatusNotFound,
			wantBody:   "Not Found [404]\n",
			checkBody:  true,
		},
		{
			name:       "empty query delimiter on known ID",
			method:     http.MethodGet,
			target:     "/stream/" + knownID + "?",
			wantStatus: http.StatusNotFound,
			wantBody:   "Not Found [404]\n",
			checkBody:  true,
		},
		{
			name:       "absolute form known ID",
			method:     http.MethodGet,
			target:     "http://threadfin.example/stream/" + knownID,
			wantStatus: http.StatusNotFound,
			wantBody:   "Not Found [404]\n",
			checkBody:  true,
		},
		{
			name:         "subtree root without trailing slash",
			method:       http.MethodGet,
			target:       "/stream",
			wantStatus:   http.StatusTemporaryRedirect,
			wantLocation: "/stream/",
			wantBody:     "<a href=\"/stream/\">Temporary Redirect</a>.\n\n",
			checkBody:    true,
		},
		{
			name:         "subtree root redirect preserves query",
			method:       http.MethodGet,
			target:       "/stream?token=abc",
			wantStatus:   http.StatusTemporaryRedirect,
			wantLocation: "/stream/?token=abc",
			wantBody:     "<a href=\"/stream/?token=abc\">Temporary Redirect</a>.\n\n",
			checkBody:    true,
		},
		{
			name:         "repeated slash clean path redirect",
			method:       http.MethodGet,
			target:       "/stream//" + knownID,
			wantStatus:   http.StatusTemporaryRedirect,
			wantLocation: "/stream/" + knownID,
			wantBody:     "<a href=\"/stream/" + knownID + "\">Temporary Redirect</a>.\n\n",
			checkBody:    true,
		},
		{
			name:         "dot segment clean path redirect",
			method:       http.MethodGet,
			target:       "/stream/stale/../" + knownID,
			wantStatus:   http.StatusTemporaryRedirect,
			wantLocation: "/stream/" + knownID,
			wantBody:     "<a href=\"/stream/" + knownID + "\">Temporary Redirect</a>.\n\n",
			checkBody:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, nil)
			recorder := httptest.NewRecorder()

			mux.ServeHTTP(recorder, request)

			response := recorder.Result()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if err := response.Body.Close(); err != nil {
				t.Fatal(err)
			}

			if response.StatusCode != test.wantStatus {
				t.Errorf("status = %d, want %d", response.StatusCode, test.wantStatus)
			}
			if got := response.Header.Get("Location"); got != test.wantLocation {
				t.Errorf("Location = %q, want %q", got, test.wantLocation)
			}
			if test.checkBody && string(body) != test.wantBody {
				t.Errorf("body = %q, want %q", body, test.wantBody)
			}
			if got := response.Header.Get("Access-Control-Allow-Origin"); got != test.wantCORS {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, test.wantCORS)
			}
			if got := response.Header.Get("X-Stream-Head"); got != test.wantHeadHeader {
				t.Errorf("X-Stream-Head = %q, want %q", got, test.wantHeadHeader)
			}
		})
	}
}
