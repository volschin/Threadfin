//go:build linux && amd64

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	playlistEntryCount = 100_000
	xmltvChannelCount  = 100
	xmltvProgramCount  = 10_000
	streamChannelCount = 4
	fixtureSeed        = 20260826
)

func writePlaylist(w io.Writer, baseURL string) error {
	if _, err := io.WriteString(w, "#EXTM3U\n"); err != nil {
		return err
	}
	for i := range playlistEntryCount {
		id := fmt.Sprintf("channel-%06d", i)
		_, err := fmt.Fprintf(w,
			"#EXTINF:-1 tvg-id=\"%s\" tvg-name=\"Channel %06d\" tvg-chno=\"%d\" group-title=\"Group %02d\",Channel %06d\n%s/stream/%d.ts?channel=%d\n",
			id, i, i+1, i%10, i, baseURL, i%streamChannelCount, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeGuide(w io.Writer, start time.Time) error {
	start = start.UTC().Truncate(time.Minute)
	if _, err := io.WriteString(w, xml.Header+"<tv generator-info-name=\"Threadfin PGO Pilot\">\n"); err != nil {
		return err
	}
	for channel := range xmltvChannelCount {
		id := fmt.Sprintf("channel-%06d", channel)
		if _, err := fmt.Fprintf(w, "  <channel id=\"%s\"><display-name>Channel %06d</display-name></channel>\n", id, channel); err != nil {
			return err
		}
	}
	for channel := range xmltvChannelCount {
		id := fmt.Sprintf("channel-%06d", channel)
		for programme := range xmltvProgramCount / xmltvChannelCount {
			from := start.Add(time.Duration(programme) * 30 * time.Minute)
			to := from.Add(30 * time.Minute)
			if _, err := fmt.Fprintf(w,
				"  <programme start=\"%s\" stop=\"%s\" channel=\"%s\"><title lang=\"en\">Programme %03d-%03d</title><desc lang=\"en\">Seed %d</desc></programme>\n",
				from.Format("20060102150405 -0700"), to.Format("20060102150405 -0700"), id, channel, programme, fixtureSeed); err != nil {
				return err
			}
		}
	}
	_, err := io.WriteString(w, "</tv>\n")
	return err
}

func settingsDocument(providerBaseURL, ffmpegPath, port string) ([]byte, error) {
	settings := map[string]any{
		"api":                false,
		"authentication.api": false,
		"authentication.m3u": false,
		"authentication.pms": false,
		"authentication.web": false,
		"authentication.xml": false,
		"bindIpAddress":      "127.0.0.1",
		"buffer":             "ffmpeg",
		"buffer.size.kb":     1024,
		"buffer.timeout":     0,
		"cache.images":       false,
		"epgSource":          "XEPG",
		"ffmpeg.path":        ffmpegPath,
		"ffmpeg.options":     "-hide_banner -loglevel error -i [URL] -map 0:v -map 0:a:0 -c copy -f mpegts pipe:1",
		"files": map[string]any{
			"hdhr": map[string]any{},
			"m3u": map[string]any{"M1000": map[string]any{
				"name": "PGO M3U", "file.source": providerBaseURL + "/playlist.m3u",
				"buffer": "ffmpeg", "tuner": 4, "http_proxy.ip": "", "http_proxy.port": "",
			}},
			"xmltv": map[string]any{"X1000": map[string]any{
				"name": "PGO XMLTV", "file.source": providerBaseURL + "/guide.xml",
				"http_proxy.ip": "", "http_proxy.port": "",
			}},
		},
		"files.update":                true,
		"filter":                      map[string]any{},
		"mapping.first.channel":       1000,
		"port":                        port,
		"ssdp":                        false,
		"tuner":                       4,
		"ThreadfinAutoUpdate":         false,
		"update":                      []string{"0000"},
		"uuid":                        "20260826-PGO-PILOT",
		"version":                     "0.5.0",
		"xepg.replace.missing.images": false,
		"xepg.replace.channel.title":  false,
	}
	body, err := json.MarshalIndent(settings, "", "  ")
	return append(body, '\n'), err
}

type fixtureSnapshot struct {
	PlaylistStarted  time.Time
	PlaylistFinished time.Time
	GuideStarted     time.Time
	GuideFinished    time.Time
	StreamRequests   [streamChannelCount]int
	StreamBytes      uint64
	Cancellations    uint64
}

type fixtureSet struct {
	server    *httptest.Server
	tempDir   string
	playlist  []byte
	guide     []byte
	transport []byte
	closeOnce sync.Once
	mu        sync.Mutex
	observed  fixtureSnapshot
}

func (f *fixtureSet) baseURL() string           { return f.server.URL }
func (f *fixtureSet) playlistBytes() []byte     { return append([]byte(nil), f.playlist...) }
func (f *fixtureSet) guideBytes() []byte        { return append([]byte(nil), f.guide...) }
func (f *fixtureSet) snapshot() fixtureSnapshot { f.mu.Lock(); defer f.mu.Unlock(); return f.observed }
func (f *fixtureSet) Close() {
	f.closeOnce.Do(func() {
		if f.server != nil {
			f.server.Close()
		}
		if f.tempDir != "" {
			_ = os.RemoveAll(f.tempDir)
		}
	})
}

func newFixtureSet(ctx context.Context, ffmpegPath string, guideStart time.Time) (*fixtureSet, error) {
	tempDir, err := os.MkdirTemp("", "threadfin-pgo-fixture-")
	if err != nil {
		return nil, err
	}
	sourcePath := filepath.Join(tempDir, "source.ts")
	ffmpegCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ffmpegCtx, ffmpegPath,
		"-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "testsrc2=size=640x360:rate=25",
		"-f", "lavfi", "-i", "sine=frequency=1000:sample_rate=48000",
		"-t", "2", "-c:v", "mpeg2video", "-b:v", "2M", "-c:a", "mp2",
		"-f", "mpegts", sourcePath)
	if output, runErr := command.CombinedOutput(); runErr != nil {
		_ = os.RemoveAll(tempDir)
		return nil, fmt.Errorf("generate MPEG-TS fixture: %w: %s", runErr, strings.TrimSpace(string(output)))
	}
	transport, err := os.ReadFile(sourcePath)
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, err
	}
	if len(transport) == 0 {
		_ = os.RemoveAll(tempDir)
		return nil, errors.New("FFmpeg produced an empty fixture")
	}
	fixture := &fixtureSet{tempDir: tempDir, transport: transport}
	mux := http.NewServeMux()
	mux.HandleFunc("/playlist.m3u", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().UTC()
		fixture.mu.Lock()
		if fixture.observed.PlaylistStarted.IsZero() {
			fixture.observed.PlaylistStarted = now
		}
		fixture.mu.Unlock()
		w.Header().Set("Content-Type", "audio/x-mpegurl")
		w.Header().Set("Content-Length", strconv.Itoa(len(fixture.playlist)))
		_, _ = w.Write(fixture.playlist)
		fixture.mu.Lock()
		fixture.observed.PlaylistFinished = time.Now().UTC()
		fixture.mu.Unlock()
	})
	mux.HandleFunc("/guide.xml", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().UTC()
		fixture.mu.Lock()
		if fixture.observed.GuideStarted.IsZero() {
			fixture.observed.GuideStarted = now
		}
		fixture.mu.Unlock()
		w.Header().Set("Content-Type", "application/xml")
		w.Header().Set("Content-Length", strconv.Itoa(len(fixture.guide)))
		_, _ = w.Write(fixture.guide)
		fixture.mu.Lock()
		fixture.observed.GuideFinished = time.Now().UTC()
		fixture.mu.Unlock()
	})
	for index := range streamChannelCount {
		path := fmt.Sprintf("/stream/%d.ts", index)
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			fixture.mu.Lock()
			fixture.observed.StreamRequests[index]++
			fixture.mu.Unlock()
			w.Header().Set("Content-Type", "video/mp2t")
			flusher, _ := w.(http.Flusher)
			for {
				select {
				case <-r.Context().Done():
					fixture.mu.Lock()
					fixture.observed.Cancellations++
					fixture.mu.Unlock()
					return
				default:
				}
				written, writeErr := w.Write(fixture.transport)
				fixture.mu.Lock()
				fixture.observed.StreamBytes += uint64(written)
				fixture.mu.Unlock()
				if flusher != nil {
					flusher.Flush()
				}
				if writeErr != nil {
					fixture.mu.Lock()
					fixture.observed.Cancellations++
					fixture.mu.Unlock()
					return
				}
			}
		})
	}
	fixture.server = httptest.NewServer(mux)
	var playlist, guide bytes.Buffer
	if err := writePlaylist(&playlist, fixture.server.URL); err != nil {
		fixture.Close()
		return nil, err
	}
	if err := writeGuide(&guide, guideStart); err != nil {
		fixture.Close()
		return nil, err
	}
	fixture.playlist, fixture.guide = playlist.Bytes(), guide.Bytes()
	return fixture, nil
}
