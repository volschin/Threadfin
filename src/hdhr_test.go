package src

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestHDHRWireJSONIsCompact(t *testing.T) {
	restorePersistentState(t)
	System.ServerProtocol.WEB, System.ServerProtocol.DVR = "http", "http"
	System.Domain, System.Name, System.AppName = "127.0.0.1:34400", "Threadfin", "threadfin"
	System.Version, System.DeviceID, Settings.Tuner = "3.0.0", "device", 2
	System.File.URLS = filepath.Join(t.TempDir(), "urls.json")
	for _, tc := range []struct {
		name string
		call func() ([]byte, error)
	}{
		{"discover", getDiscover}, {"lineup-status", getLineupStatus}, {"lineup", getLineup},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.call()
			if err != nil {
				t.Fatal(err)
			}
			if bytes.ContainsAny(got, "\n\t") {
				t.Fatalf("indented wire JSON: %q", got)
			}
		})
	}
}
