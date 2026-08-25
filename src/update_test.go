package src

import (
	"testing"
	"time"
)

func TestOfficialUpdateAssetNameMatchesReleaseTargets(t *testing.T) {
	tests := []struct {
		goos   string
		goarch string
		want   string
	}{
		{goos: "linux", goarch: "arm64", want: "Threadfin_linux_arm64"},
		{goos: "linux", goarch: "arm", want: "Threadfin_linux_arm"},
		{goos: "linux", goarch: "amd64", want: "Threadfin_linux_amd64"},
		{goos: "freebsd", goarch: "amd64", want: "Threadfin_freebsd_amd64"},
		{goos: "freebsd", goarch: "arm", want: "Threadfin_freebsd_arm"},
		{goos: "darwin", goarch: "arm64", want: "Threadfin_darwin_arm64"},
		{goos: "darwin", goarch: "amd64", want: "Threadfin_darwin_amd64"},
		{goos: "windows", goarch: "amd64", want: "Threadfin_windows_amd64.exe"},
	}

	for _, test := range tests {
		t.Run(test.goos+"/"+test.goarch, func(t *testing.T) {
			if got := officialUpdateAssetName(test.goos, test.goarch); got != test.want {
				t.Fatalf("official update asset = %q, want %q", got, test.want)
			}
		})
	}
}

func TestWindowsUpdateReadinessBudgetIncludesSequentialConfiguredProviderTimeouts(t *testing.T) {
	var settings SettingsStruct
	settings.BufferTimeout = 500
	settings.FilesUpdate = true
	settings.EpgSource = "XEPG"
	settings.Files.M3U = map[string]interface{}{
		"remote": map[string]interface{}{"file.source": "https://example.test/playlist.m3u"},
		"local":  map[string]interface{}{"file.source": `/srv/threadfin/playlist.m3u`},
	}
	settings.Files.HDHR = map[string]interface{}{
		"tuner": map[string]interface{}{"file.source": "192.0.2.10"},
	}
	settings.Files.XMLTV = map[string]interface{}{
		"remote": map[string]interface{}{"file.source": "https://example.test/guide.xml"},
		"local":  map[string]interface{}{"file.source": `/srv/threadfin/guide.xml`},
	}

	if got, want := windowsUpdateReadinessBudget(settings), 27*time.Minute; got != want {
		t.Fatalf("Windows update readiness budget = %v, want %v", got, want)
	}
}
