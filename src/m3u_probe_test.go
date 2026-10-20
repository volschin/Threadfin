package src

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestProbeChannelWithJellyfinFFmpegDirectory(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg is not installed")
	}
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		t.Skip("ffprobe is not installed")
	}
	dir := filepath.Join(t.TempDir(), "jellyfin-ffmpeg")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	probeName := "ffprobe"
	ffmpegName := "ffmpeg"
	if runtime.GOOS == "windows" {
		probeName += ".exe"
		ffmpegName += ".exe"
	}
	if err := os.Symlink(ffprobe, filepath.Join(dir, probeName)); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("symlinks unavailable: %v", err)
		}
		t.Fatal(err)
	}
	input := filepath.Join(t.TempDir(), "input.ts")
	cmd := exec.Command(ffmpeg, "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "color=s=64x64:r=25", "-t", "1", "-c:v", "mpeg2video", "-f", "mpegts", input)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generate probe fixture: %v: %s", err, output)
	}
	previousPath := Settings.FFmpegPath
	t.Cleanup(func() { Settings.FFmpegPath = previousPath })
	t.Chdir(dir)
	t.Setenv("PATH", "")
	for _, executable := range []string{filepath.Join(dir, ffmpegName), "." + string(filepath.Separator) + ffmpegName} {
		t.Run(executable, func(t *testing.T) {
			Settings.FFmpegPath = executable
			resolution, frameRate, _, err := probeChannel(RequestStruct{ProbeURL: input})
			if err != nil {
				t.Fatal(err)
			}
			if resolution != "64p" || frameRate != "25" {
				t.Fatalf("probe = %s at %s fps, want 64p at 25 fps", resolution, frameRate)
			}
		})
	}
}
