package src

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	up2date "threadfin/src/internal/up2date/client"
)

func TestBinaryUpdateHandlesMissingAndInvalidVersions(t *testing.T) {
	for _, test := range []struct {
		name, branch, body string
		status             int
		wantError          bool
	}{
		{"no beta release", "Beta", `[{"tag_name":"v3.2.4","prerelease":false}]`, 200, false},
		{"empty releases", "Main", `[]`, 200, false},
		{"null release", "Main", `[null]`, 200, false},
		{"malformed release", "Main", `[{"tag_name":"invalid","prerelease":false}]`, 200, true},
		{"empty tag", "Main", `[{"tag_name":"","prerelease":false}]`, 200, true},
		{"server failure", "Main", `[]`, 500, true},
		{"malformed custom version", "Custom", `{"status":true,"version":"invalid"}`, 200, true},
		{"disabled custom response", "Custom", `{"status":false}`, 200, false},
		{"current release", "Main", `[{"tag_name":"v3.2.4","prerelease":false}]`, 200, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			restorePersistentState(t)
			previous := up2date.Updater
			t.Cleanup(func() { up2date.Updater = previous })
			up2date.Updater = up2date.ClientInfo{Response: up2date.ServerResponse{Status: true, Version: "v99.0.0", UpdateBIN: "http://stale.test/binary"}}
			Settings = SettingsStruct{ThreadfinAutoUpdate: true}
			System = SystemStruct{Branch: test.branch, Version: "3.2", Build: "4", OS: "linux", ARCH: "amd64"}
			System.GitHub.Update = true
			var assetRequests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if test.branch != "Custom" && r.URL.Path != "/releases" {
					assetRequests.Add(1)
				}
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()
			System.Update.Github = server.URL
			System.Update.Git = server.URL
			Settings.UpdateURL = server.URL
			defer func() {
				if value := recover(); value != nil {
					t.Errorf("update check panicked: %v", value)
				}
			}()
			err := BinaryUpdate()
			if (err != nil) != test.wantError {
				t.Fatalf("error = %v, wantError %v", err, test.wantError)
			}
			if up2date.Updater.Response.UpdateBIN == "http://stale.test/binary" {
				t.Error("retained a stale release candidate")
			}
			if assetRequests.Load() != 0 {
				t.Error("missing, invalid or current release triggered a binary download")
			}
		})
	}
}
