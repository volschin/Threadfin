package src

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func finiteBufferPayload() []byte {
	packet := bytes.Repeat([]byte{0xff}, 188)
	copy(packet, []byte{0x47, 0x1f, 0xff, 0x10}) // MPEG-TS null packet.
	return bytes.Repeat(packet, 128)
}

func TestFiniteBufferProcessHelper(t *testing.T) {
	if os.Getenv("THREADFIN_FINITE_BUFFER_HELPER") != "1" {
		return
	}
	chunk := finiteBufferPayload()
	if os.Getenv("THREADFIN_FINITE_BUFFER_TAIL") == "1" {
		_, _ = os.Stdout.Write(append(chunk, chunk[:24*188]...))
		os.Exit(0)
	}
	_, _ = os.Stdout.Write(chunk)
	time.Sleep(20 * time.Millisecond)
	_, _ = os.Stdout.Write(chunk)
	os.Exit(0)
}

func TestBufferingStreamDrainsCompletedSegmentsAfterProducerExit(t *testing.T) {
	for _, tail := range []bool{false, true} {
		name := "full segments"
		if tail {
			name = "short final segment without pauses"
		}
		t.Run(name, func(t *testing.T) { testBufferingStreamDrainsFinitePayload(t, tail) })
	}
}

func testBufferingStreamDrainsFinitePayload(t *testing.T, tail bool) {
	vfs := useMemoryBufferVFS(t)
	previousSettings, previousFlag := Settings, System.Flag
	t.Cleanup(func() { Settings = previousSettings; System.Flag = previousFlag })
	System.Flag.Info = true
	System.Flag.Debug = 0
	Settings.BufferTimeout = 0
	Settings.BufferSize = 47 // Half the configured buffer is 128 MPEG-TS packets.
	Settings.UserAgent = ""
	Settings.FFmpegPath = os.Args[0]
	Settings.FFmpegOptions = "-test.run=^TestFiniteBufferProcessHelper$"
	t.Setenv("THREADFIN_FINITE_BUFFER_HELPER", "1")
	if tail {
		t.Setenv("THREADFIN_FINITE_BUFFER_TAIL", "1")
	}
	t.Setenv("GORACE", "atexit_sleep_ms=0")
	const id = "finite-buffer"
	BufferInformation.Store(id, Playlist{PlaylistID: id, Folder: "finite/", Buffer: "ffmpeg", Tuner: 1})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bufferingStream(id, "https://example.test/finite", nil, nil, nil, "Finite stream", w, r)
	}))
	t.Cleanup(server.Close)
	client := server.Client()
	client.Timeout = 3 * time.Second
	response, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	content, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	want := bytes.Repeat(finiteBufferPayload(), 2)
	if tail {
		want = append(finiteBufferPayload(), finiteBufferPayload()[:24*188]...)
	}
	if response.StatusCode != http.StatusOK || !bytes.Equal(content, want) {
		t.Errorf("finite stream response: status=%d, bytes=%d; want 200 and %d complete bytes", response.StatusCode, len(content), len(want))
	}
	if _, ok := playlistSnapshot(id); ok {
		t.Error("last HTTP reader did not release its playlist")
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		entries, err := vfs.ReadDir("finite")
		if err == nil && len(entries) == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Error("finished producer's segments remain after the last reader leaves")
}

func TestBufferingStreamWithRealFFmpeg(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg is not installed")
	}
	fixture, err := exec.Command(ffmpeg, "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "color=c=blue:s=64x64:r=10", "-t", "0.5", "-c:v", "mpeg2video", "-f", "mpegts", "pipe:1").Output()
	if err != nil {
		t.Fatalf("generate local MPEG-TS fixture: %v", err)
	}
	input := filepath.Join(t.TempDir(), "input.ts")
	if err := os.WriteFile(input, fixture, 0600); err != nil {
		t.Fatal(err)
	}
	want, err := exec.Command(ffmpeg, "-hide_banner", "-loglevel", "error", "-i", input, "-c", "copy", "-f", "mpegts", "pipe:1").Output()
	if err != nil {
		t.Fatalf("generate expected remuxed output: %v", err)
	}
	if len(want) < 512 {
		t.Fatalf("FFmpeg generated only %d bytes", len(want))
	}
	vfs := useMemoryBufferVFS(t)
	previousSettings, previousFlag := Settings, System.Flag
	t.Cleanup(func() { Settings = previousSettings; System.Flag = previousFlag })
	System.Flag.Info = true
	System.Flag.Debug = 0
	Settings.BufferTimeout = 0
	Settings.BufferSize = 1
	Settings.UserAgent = ""
	Settings.FFmpegPath = ffmpeg
	Settings.FFmpegOptions = "-hide_banner -loglevel error -i [URL] -c copy -f mpegts pipe:1"
	const id = "real-ffmpeg-buffer"
	BufferInformation.Store(id, Playlist{PlaylistID: id, Folder: "real-ffmpeg/", Buffer: "ffmpeg", Tuner: 1})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bufferingStream(id, input, nil, nil, nil, "Local FFmpeg fixture", w, r)
	}))
	t.Cleanup(server.Close)
	client := server.Client()
	client.Timeout = 5 * time.Second
	response, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	content, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || !bytes.Equal(content, want) {
		t.Errorf("FFmpeg response: status=%d, bytes=%d; want 200 and %d identical bytes", response.StatusCode, len(content), len(want))
	}
	t.Logf("real FFmpeg remuxed %d fixture bytes into %d bytes delivered through HTTP", len(fixture), len(content))
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		entries, err := vfs.ReadDir("real-ffmpeg")
		if err == nil && len(entries) == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Error("FFmpeg buffer was not removed after its reader finished")
}

func TestFinishedBufferRetainsSegmentsUntilFinalClientRelease(t *testing.T) {
	vfs := useMemoryBufferVFS(t)
	previousSettings, previousFlag := Settings, System.Flag
	t.Cleanup(func() { Settings = previousSettings; System.Flag = previousFlag })
	System.Flag.Info = true
	System.Flag.Debug = 0
	Settings.BufferSize = 47
	Settings.UserAgent = ""
	Settings.FFmpegPath = os.Args[0]
	Settings.FFmpegOptions = "-test.run=^TestFiniteBufferProcessHelper$"
	t.Setenv("THREADFIN_FINITE_BUFFER_HELPER", "1")
	t.Setenv("GORACE", "atexit_sleep_ms=0")
	const id = "finite-shared-buffer"
	candidate := Playlist{PlaylistID: id, Folder: "finite-shared/", Buffer: "ffmpeg", Tuner: 1}
	playlist, streamID, _, _ := registerBufferClient(candidate, ThisStream{URL: "https://example.test/finite"}, "", "")
	_, _, _, accepted := registerBufferClient(candidate, ThisStream{URL: "https://example.test/finite"}, "", "")
	if !accepted {
		t.Fatal("second client was not attached")
	}
	stream := playlist.Streams[streamID]
	done := make(chan struct{})
	go func() { thirdPartyBuffer(streamID, playlist, stream); close(done) }()
	t.Cleanup(func() { releaseBufferClient(streamID, stream, true); <-done })
	deadline := time.Now().Add(3 * time.Second)
	finished := false
	for time.Now().Before(deadline) {
		value, _ := BufferClients.Load(id + stream.MD5)
		if value.(ClientConnection).Error != nil {
			finished = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !finished {
		t.Fatal("finite producer did not finish")
	}
	releaseBufferClient(streamID, stream, false)
	if content, err := vfs.ReadFile(stream.Folder + "1.ts"); err != nil || len(content) < 24064 || !bytes.HasPrefix(bytes.Repeat(finiteBufferPayload(), 2), content) {
		t.Fatalf("remaining reader lost its completed segment: bytes=%d, err=%v", len(content), err)
	}
	select {
	case <-done:
		t.Fatal("producer removed its buffer before the last reader left")
	default:
	}
	releaseBufferClient(streamID, stream, false)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("producer did not finish cleanup after final release")
	}
	if _, err := vfs.Stat(stream.Folder); !fsIsNotExistErr(err) {
		t.Fatalf("retired segment directory remains: %v", err)
	}
}

func TestBufferMonitoringPreservesSharedClients(t *testing.T) {
	const id = "monitor-shared-clients"
	playlist := Playlist{PlaylistID: id, Clients: map[int]ThisClient{0: {Connection: 3}, 1: {Connection: 0}}}
	BufferInformation.Store(id, playlist)
	t.Cleanup(func() { BufferInformation.Delete(id) })
	if count := getActiveClientCount(); count != 3 {
		t.Errorf("active clients = %d, want 3", count)
	}
	if playlist.Clients[0].Connection != 3 || len(playlist.Clients) != 2 {
		t.Errorf("monitoring modified live clients: %v", playlist.Clients)
	}
}

func TestInitBufferVFSPreservesActiveSegments(t *testing.T) {
	vfs := useMemoryBufferVFS(t)
	writeBufferTestFile(t, vfs, "active.ts", "playing")
	initBufferVFS()
	content, err := bufferVFS.ReadFile("active.ts")
	if err != nil || string(content) != "playing" {
		t.Fatalf("active segment after reinitialization = %q, %v", content, err)
	}
}

// The test executable stands in for a silent ffmpeg without shell children.
func TestSilentBufferProcessHelper(t *testing.T) {
	if os.Getenv("THREADFIN_BUFFER_HELPER") != "1" {
		return
	}
	if len(os.Args) > 2 {
		if source := os.Args[len(os.Args)-1]; strings.HasPrefix(source, "http") {
			// Simulate a transcoder echoing all input credentials in diagnostics.
			_, _ = os.Stderr.WriteString("failed source: " + source + "\n")
			if strings.Contains(source, "primary") {
				if os.Getenv("THREADFIN_BUFFER_SILENT_PRIMARY") == "1" {
					time.Sleep(time.Minute)
				}
				os.Exit(1)
			}
			_, _ = os.Stdout.Write(bytes.Repeat([]byte{0x47}, 1024))
		}
	}
	if err := os.WriteFile(os.Getenv("THREADFIN_BUFFER_PID"), []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
		os.Exit(2)
	}
	time.Sleep(time.Minute)
	os.Exit(0)
}

func TestBufferRegistrationRechecksDeletedSnapshot(t *testing.T) {
	const id = "stale-registration"
	candidate := Playlist{PlaylistID: id, Tuner: 2}
	old, streamID, _, _ := registerBufferClient(candidate, ThisStream{URL: "https://example.test/a"}, "", "")
	defer BufferInformation.Delete(id)
	defer BufferClients.Delete(id + getMD5("https://example.test/a"))
	releaseBufferClient(streamID, old.Streams[streamID], false)
	replacement, replacementID, newStream, accepted := registerBufferClient(old, ThisStream{URL: "https://example.test/a"}, "", "")
	if !accepted || !newStream || replacement.Clients[replacementID].Connection != 1 {
		t.Fatalf("registration from stale snapshot: new=%v, accepted=%v, clients=%v", newStream, accepted, replacement.Clients)
	}
	if replacement.Streams[replacementID].done == old.Streams[streamID].done {
		t.Fatal("replacement reused a cancelled producer")
	}
}

func TestBufferConcurrentRegistrationMonitoringAndRelease(t *testing.T) {
	const id = "concurrent-clients"
	candidate := Playlist{PlaylistID: id, Tuner: 32}
	anchor, streamID, _, _ := registerBufferClient(candidate, ThisStream{URL: "https://example.test/shared"}, "", "")
	t.Cleanup(func() { releaseBufferClient(streamID, anchor.Streams[streamID], true) })
	var group sync.WaitGroup
	for range 16 {
		group.Go(func() {
			for range 50 {
				playlist, id, fresh, ok := registerBufferClient(candidate, ThisStream{URL: "https://example.test/shared"}, "", "")
				if !ok || fresh {
					t.Error("shared stream was not reused")
					return
				}
				_ = getActiveClientCount()
				_ = getActivePlaylistCount()
				updateBufferStream(id, playlist.Streams[id], true, nil)
				// A snapshot must remain safe while other goroutines modify live maps.
				for _, client := range playlist.Clients {
					if client.Connection < 1 {
						t.Error("invalid snapshot count")
					}
				}
				releaseBufferClient(id, playlist.Streams[id], false)
			}
		})
	}
	group.Wait()
	playlist, ok := playlistSnapshot(id)
	if !ok || len(playlist.Streams) != 1 || playlist.Clients[streamID].Connection != 1 {
		t.Fatalf("remaining anchor client = %+v, exists=%v", playlist.Clients, ok)
	}
}

func TestConcurrentFirstBufferClientsRespectTunerLimit(t *testing.T) {
	const id = "first-buffer-clients"
	candidate := Playlist{PlaylistID: id, Tuner: 4}
	type registered struct {
		stream ThisStream
		id     int
	}
	accepted := make(chan registered, 24)
	var group sync.WaitGroup
	for index := range 24 {
		group.Go(func() {
			playlist, streamID, _, ok := registerBufferClient(candidate, ThisStream{URL: "https://example.test/" + strconv.Itoa(index)}, "", "")
			if ok {
				accepted <- registered{playlist.Streams[streamID], streamID}
			}
		})
	}
	group.Wait()
	close(accepted)
	defer func() {
		for client := range accepted {
			releaseBufferClient(client.id, client.stream, true)
		}
	}()
	playlist, ok := playlistSnapshot(id)
	if !ok || len(playlist.Streams) != 4 || len(accepted) != 4 {
		t.Fatalf("accepted streams = %d, live streams = %d, exists = %v", len(accepted), len(playlist.Streams), ok)
	}
}

func TestLateBufferProducerCannotModifyReplacement(t *testing.T) {
	const id = "late-producer"
	candidate := Playlist{PlaylistID: id, Tuner: 2}
	old, oldID, _, _ := registerBufferClient(candidate, ThisStream{URL: "https://example.test/shared"}, "", "")
	releaseBufferClient(oldID, old.Streams[oldID], false)
	updateBufferStream(oldID, old.Streams[oldID], true, errors.New("old producer error"))
	if _, ok := playlistSnapshot(id); ok {
		t.Fatal("late producer resurrected removed playlist")
	}
	current, currentID, _, _ := registerBufferClient(candidate, ThisStream{URL: "https://example.test/shared"}, "", "")
	t.Cleanup(func() { releaseBufferClient(currentID, current.Streams[currentID], true) })
	updateBufferStream(oldID, old.Streams[oldID], true, errors.New("old producer error"))
	releaseBufferClient(oldID, old.Streams[oldID], true)
	actual, ok := playlistSnapshot(id)
	if !ok || actual.Streams[currentID].Status || actual.Clients[currentID].Connection != 1 {
		t.Fatalf("late producer changed replacement: %+v, exists=%v", actual, ok)
	}
	clients, _ := BufferClients.Load(id + current.Streams[currentID].MD5)
	if clients.(ClientConnection).Error != nil {
		t.Fatal("late error leaked into replacement")
	}
}

func TestReplacementBufferSegmentsCannotBeOverwrittenByOldProducer(t *testing.T) {
	vfs := useMemoryBufferVFS(t)
	const id = "replacement-segments"
	candidate := Playlist{PlaylistID: id, Tuner: 1, Folder: "replacement/"}
	old, oldID, _, _ := registerBufferClient(candidate, ThisStream{URL: "https://example.test/shared"}, "", "")
	if _, err := prepareSegmentDirectory(old.Streams[oldID].Folder, false); err != nil {
		t.Fatal(err)
	}
	releaseBufferClient(oldID, old.Streams[oldID], false)
	current, currentID, _, _ := registerBufferClient(candidate, ThisStream{URL: "https://example.test/shared"}, "", "")
	t.Cleanup(func() { releaseBufferClient(currentID, current.Streams[currentID], true) })
	if _, err := prepareSegmentDirectory(current.Streams[currentID].Folder, false); err != nil {
		t.Fatal(err)
	}
	currentFile := current.Streams[currentID].Folder + "2.ts"
	writeBufferTestFile(t, vfs, currentFile, "replacement data")
	// A pipe chunk already selected when cancellation arrives can still finish.
	writeBufferTestFile(t, vfs, old.Streams[oldID].Folder+"2.ts", "old producer data")
	actual, err := vfs.ReadFile(currentFile)
	if err != nil || string(actual) != "replacement data" {
		t.Fatalf("replacement segment = %q, %v", actual, err)
	}
}

func TestSilentBufferStartupTimeoutReapsProcess(t *testing.T) {
	stream := ThisStream{PlaylistID: "startup-timeout", MD5: "silent", done: make(chan struct{})}
	BufferClients.Store(stream.PlaylistID+stream.MD5, ClientConnection{Connection: 1})
	t.Cleanup(func() { BufferClients.Delete(stream.PlaylistID + stream.MD5) })
	cmd := exec.Command(os.Args[0], "-test.run=^TestSilentBufferProcessHelper$")
	cmd.Env = append(os.Environ(), "THREADFIN_BUFFER_HELPER=1", "THREADFIN_BUFFER_PID="+filepath.Join(t.TempDir(), "pid"))
	err := consumeBufferProcess(cmd, stream, 100*time.Millisecond, func([]byte) (bool, error) { t.Error("silent process produced output"); return false, nil })
	if !errors.Is(err, errBufferStartupTimeout) {
		t.Fatalf("startup error = %v", err)
	}
	if cmd.ProcessState == nil {
		t.Fatal("timeout did not reap the subprocess")
	}
}

func TestBufferLogsOmitCredentialsAndFailoverKeepsClients(t *testing.T) {
	useMemoryBufferVFS(t)
	previousSettings, previousFlag := Settings, System.Flag
	previousOutput := log.Writer()
	var output bytes.Buffer
	log.SetOutput(&output)
	t.Cleanup(func() { Settings = previousSettings; System.Flag = previousFlag; log.SetOutput(previousOutput) })
	System.Flag.Info = false
	System.Flag.Debug = 3
	Settings.UserAgent = ""
	Settings.BufferSize = 1
	Settings.FFmpegPath = os.Args[0]
	Settings.FFmpegOptions = "-test.run=^TestSilentBufferProcessHelper$ -- [URL]"
	t.Setenv("THREADFIN_BUFFER_HELPER", "1")
	t.Setenv("THREADFIN_BUFFER_PID", filepath.Join(t.TempDir(), "pid"))
	const id = "safe-failover"
	playlist, streamID, _, _ := registerBufferClient(Playlist{PlaylistID: id, Buffer: "ffmpeg", Folder: "failover/", Tuner: 1}, ThisStream{
		URL:            "https://alice:primary-secret@example.test/live/path-user/path-password/primary?token=query-secret#fragment-secret",
		BackupChannel2: &BackupStream{URL: "https://backup-user:backup-password@example.test/live/backup-path-user/backup-path-password/stream?token=backup-token"},
	}, "", "")
	stream := playlist.Streams[streamID]
	done := make(chan struct{})
	go func() { thirdPartyBuffer(streamID, playlist, stream); close(done) }()
	t.Cleanup(func() { releaseBufferClient(streamID, stream, true); <-done })
	deadline := time.Now().Add(3 * time.Second)
	ready := false
	for time.Now().Before(deadline) {
		current, ok := playlistSnapshot(id)
		if ok && current.Streams[streamID].Status {
			ready = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !ready {
		t.Fatal("backup in slot 2 never became ready")
	}
	if count := getActiveClientCount(); count != 1 {
		t.Errorf("failover lost client: %d", count)
	}
	releaseBufferClient(streamID, stream, true)
	<-done
	for _, secret := range []string{"alice", "primary-secret", "path-user", "path-password", "query-secret", "fragment-secret", "backup-user", "backup-password", "backup-path-user", "backup-path-password", "backup-token"} {
		if strings.Contains(output.String(), secret) {
			t.Errorf("log exposed credential %q", secret)
		}
	}
}

func TestSilentPrimaryBufferTimesOutAndStartsBackup(t *testing.T) {
	if testing.Short() {
		t.Skip("exercises the production 20-second startup timeout")
	}
	useMemoryBufferVFS(t)
	previousSettings, previousFlag := Settings, System.Flag
	t.Cleanup(func() { Settings = previousSettings; System.Flag = previousFlag })
	System.Flag.Info = true
	System.Flag.Debug = 0
	Settings.UserAgent = ""
	Settings.BufferSize = 1
	Settings.FFmpegPath = os.Args[0]
	Settings.FFmpegOptions = "-test.run=^TestSilentBufferProcessHelper$ -- [URL]"
	t.Setenv("THREADFIN_BUFFER_HELPER", "1")
	t.Setenv("THREADFIN_BUFFER_SILENT_PRIMARY", "1")
	t.Setenv("THREADFIN_BUFFER_PID", filepath.Join(t.TempDir(), "pid"))
	const id = "silent-failover"
	playlist, streamID, _, _ := registerBufferClient(Playlist{PlaylistID: id, Buffer: "ffmpeg", Folder: "silent-failover/", Tuner: 1}, ThisStream{
		URL:            "https://example.test/primary",
		BackupChannel1: &BackupStream{URL: "https://example.test/backup"},
	}, "", "")
	stream := playlist.Streams[streamID]
	done := make(chan struct{})
	go func() { thirdPartyBuffer(streamID, playlist, stream); close(done) }()
	t.Cleanup(func() { releaseBufferClient(streamID, stream, true); <-done })
	deadline := time.Now().Add(25 * time.Second)
	for time.Now().Before(deadline) {
		current, ok := playlistSnapshot(id)
		if ok && current.Streams[streamID].Status {
			if current.Clients[streamID].Connection != 1 {
				t.Fatal("timeout failover lost its client")
			}
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("silent primary never failed over to backup")
}

func TestStreamURLForLogOmitsEveryCredentialLocation(t *testing.T) {
	for _, test := range []struct{ raw, want string }{
		{"https://user:secret@example.test/live/path-user/path-secret/channel?token=query-secret#fragment-secret", "https://example.test"},
		{"rtsp://user:secret@example.test:8554/private", "rtsp://example.test:8554"},
		{"https://example.test/movie/user/secret/123.ts", "https://example.test"},
		{"http://%zz/secret", "[redacted URL]"},
		{"/live/user/secret", "[redacted URL]"},
	} {
		if got := streamURLForLog(test.raw); got != test.want {
			t.Errorf("redacted endpoint = %q, want %q", got, test.want)
		}
	}
}

func TestThirdPartyBufferStopsSilentProcessWhenLastClientLeaves(t *testing.T) {
	useMemoryBufferVFS(t)
	previousSettings := Settings
	previousInfo := System.Flag.Info
	System.Flag.Info = true
	t.Cleanup(func() { Settings = previousSettings; System.Flag.Info = previousInfo })
	Settings.FFmpegPath = os.Args[0]
	Settings.FFmpegOptions = "-test.run=^TestSilentBufferProcessHelper$"
	Settings.UserAgent = ""
	Settings.BufferSize = 1024
	pidFile := filepath.Join(t.TempDir(), "pid")
	t.Setenv("THREADFIN_BUFFER_HELPER", "1")
	t.Setenv("THREADFIN_BUFFER_PID", pidFile)
	const id = "silent-buffer"
	stream := ThisStream{PlaylistID: id, MD5: "silent", Folder: "silent/", URL: "http://example.test/stream"}
	BufferInformation.Store(id, Playlist{PlaylistID: id, Buffer: "ffmpeg", Streams: map[int]ThisStream{0: stream}, Clients: map[int]ThisClient{0: {Connection: 1}}})
	BufferClients.Store(id+stream.MD5, ClientConnection{Connection: 1})
	done := make(chan struct{})
	go func() { playlist, _ := playlistSnapshot(id); thirdPartyBuffer(0, playlist, stream); close(done) }()
	var child *os.Process
	t.Cleanup(func() {
		if child != nil {
			_ = child.Kill()
		}
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("buffer worker did not finish during cleanup")
		}
		BufferInformation.Delete(id)
		BufferClients.Delete(id + stream.MD5)
	})
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if raw, err := os.ReadFile(pidFile); err == nil {
			pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
			if err != nil {
				t.Fatal(err)
			}
			child, err = os.FindProcess(pid)
			if err != nil {
				t.Fatal(err)
			}
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if child == nil {
		t.Fatal("silent process did not start")
	}
	killClientConnection(0, id, false)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("silent process blocks buffer shutdown after the last client leaves")
	}
}
