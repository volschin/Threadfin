package src

import "testing"

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
