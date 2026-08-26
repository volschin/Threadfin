package src

import (
	"errors"
	"io"
	"math"
	"strings"
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

func TestConfiguredProviderRequestTimeoutNormalizesBounds(t *testing.T) {
	tests := []struct {
		name          string
		bufferTimeout float64
		want          time.Duration
	}{
		{name: "negative uses default", bufferTimeout: -1, want: 30 * time.Second},
		{name: "zero uses default", bufferTimeout: 0, want: 30 * time.Second},
		{name: "NaN uses default", bufferTimeout: math.NaN(), want: 30 * time.Second},
		{name: "positive sub-millisecond clamps", bufferTimeout: 0.0005, want: time.Millisecond},
		{name: "one millisecond boundary", bufferTimeout: 0.001, want: time.Millisecond},
		{name: "normal configured timeout", bufferTimeout: 500, want: 500 * time.Second},
		{name: "overflow saturates", bufferTimeout: math.MaxFloat64, want: time.Duration(1<<63 - 1)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := configuredProviderRequestTimeout(test.bufferTimeout); got != test.want {
				t.Fatalf("configured provider request timeout = %v, want %v", got, test.want)
			}
		})
	}
}

func TestWindowsUpdateReadinessBudgetUsesNormalizedProviderTimeout(t *testing.T) {
	var settings SettingsStruct
	settings.BufferTimeout = 0.0005
	settings.FilesUpdate = true
	settings.Files.M3U = map[string]interface{}{
		"remote": map[string]interface{}{"file.source": "https://example.test/playlist.m3u"},
	}

	if got, want := windowsUpdateReadinessBudget(settings), 2*time.Minute+time.Millisecond; got != want {
		t.Fatalf("Windows update readiness budget = %v, want %v", got, want)
	}
}

func TestWindowsUpdateReadinessBudgetUsesDefaultProviderTimeoutForNaN(t *testing.T) {
	var settings SettingsStruct
	settings.BufferTimeout = math.NaN()
	settings.FilesUpdate = true
	settings.Files.M3U = map[string]interface{}{
		"remote": map[string]interface{}{"file.source": "https://example.test/playlist.m3u"},
	}

	if got, want := windowsUpdateReadinessBudget(settings), 2*time.Minute+30*time.Second; got != want {
		t.Fatalf("Windows update readiness budget = %v, want %v", got, want)
	}
}

var errGitHubRead = errors.New("GitHub response read failed")
var errGitHubClose = errors.New("GitHub response close failed")

type githubBody struct {
	io.Reader
	closeErr error
}

func (b githubBody) Close() error { return b.closeErr }

type githubReadFailure struct{}

func (githubReadFailure) Read([]byte) (int, error) { return 0, errGitHubRead }

func TestDecodeGitHubReleasesClosesAndJoinsErrors(t *testing.T) {
	body := githubBody{Reader: githubReadFailure{}, closeErr: errGitHubClose}
	_, err := decodeGitHubReleases(body)
	if !errors.Is(err, errGitHubRead) || !errors.Is(err, errGitHubClose) {
		t.Fatalf("error=%v", err)
	}
}

func TestDecodeGitHubReleasesRejectsTrailingValue(t *testing.T) {
	if _, err := decodeGitHubReleases(githubBody{Reader: strings.NewReader(`[] [])`)}); err == nil {
		t.Fatal("expected trailing-value error")
	}
}
