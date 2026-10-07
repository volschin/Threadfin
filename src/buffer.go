package src

/*
  Render tuner-limit image as video [ffmpeg]
  -loop 1 -i stream-limit.jpg -c:v libx264 -t 1 -pix_fmt yuv420p -vf scale=1920:1080  stream-limit.ts
*/

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/avfs/avfs/vfs/memfs"
)

type BackupStream struct {
	PlaylistID string
	URL        string
}

// Lock owns BufferInformation's nested maps and every BufferClients update.
// Readers outside Lock receive copies; producers update only the current stream.
func playlistSnapshot(id string) (Playlist, bool) {
	Lock.RLock()
	defer Lock.RUnlock()
	value, ok := BufferInformation.Load(id)
	if !ok {
		return Playlist{}, false
	}
	playlist := value.(Playlist)
	playlist.Streams = maps.Clone(playlist.Streams)
	playlist.Clients = maps.Clone(playlist.Clients)
	return playlist, true
}

func getActiveClientCount() (count int) {
	Lock.RLock()
	defer Lock.RUnlock()
	BufferInformation.Range(func(_, value any) bool {
		if playlist, ok := value.(Playlist); ok {
			for _, client := range playlist.Clients {
				count += max(0, client.Connection)
			}
		}
		return true
	})
	return
}

func getActivePlaylistCount() (count int) {
	Lock.RLock()
	defer Lock.RUnlock()
	BufferInformation.Range(func(_, _ any) bool {
		count++
		return true
	})
	return
}

// Registration rechecks the live playlist while holding Lock, even when two
// first clients constructed their playlist configuration concurrently.
var nextBufferSession uint64 // guarded by Lock

func registerBufferClient(candidate Playlist, stream ThisStream, ip, userAgent string) (Playlist, int, bool, bool) {
	Lock.Lock()
	defer Lock.Unlock()
	playlist := candidate
	if value, ok := BufferInformation.Load(candidate.PlaylistID); ok {
		playlist = value.(Playlist)
	} else {
		playlist.Streams = make(map[int]ThisStream)
		playlist.Clients = make(map[int]ThisClient)
	}
	for id, current := range playlist.Streams {
		if current.URL == stream.URL {
			client := playlist.Clients[id]
			client.Connection++
			playlist.Clients[id] = client
			connection, _ := BufferClients.Load(playlist.PlaylistID + current.MD5)
			clients, _ := connection.(ClientConnection)
			clients.Connection = client.Connection
			BufferClients.Store(playlist.PlaylistID+current.MD5, clients)
			playlist.Streams = maps.Clone(playlist.Streams)
			playlist.Clients = maps.Clone(playlist.Clients)
			return playlist, id, false, true
		}
	}
	if len(playlist.Streams) >= playlist.Tuner {
		playlist.Streams = maps.Clone(playlist.Streams)
		playlist.Clients = maps.Clone(playlist.Clients)
		return playlist, 0, false, false
	}
	if playlist.Streams == nil {
		playlist.Streams = make(map[int]ThisStream)
	}
	if playlist.Clients == nil {
		playlist.Clients = make(map[int]ThisClient)
	}
	id := createStreamID(playlist.Streams, ip, userAgent)
	stream.PlaylistID = playlist.PlaylistID
	stream.PlaylistName = playlist.PlaylistName
	stream.MD5 = getMD5(stream.URL)
	nextBufferSession++
	// The previous producer may still be stopping after the last disconnect.
	stream.Folder = fmt.Sprintf("%s%s-%d%c", playlist.Folder, stream.MD5, nextBufferSession, os.PathSeparator)
	stream.done = make(chan struct{})
	playlist.Streams[id] = stream
	playlist.Clients[id] = ThisClient{Connection: 1}
	BufferClients.Store(playlist.PlaylistID+stream.MD5, ClientConnection{Connection: 1})
	BufferInformation.Store(playlist.PlaylistID, playlist)
	playlist.Streams = maps.Clone(playlist.Streams)
	playlist.Clients = maps.Clone(playlist.Clients)
	return playlist, id, true, true
}

func getClientIP(r *http.Request) string {
	// Check the X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For may contain multiple IP addresses; return the first one
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check the X-Real-IP header next
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		// Remove port if present
		ip = strings.Split(ip, ":")[0]
	}

	return ip
}

func createStreamID(stream map[int]ThisStream, ip, userAgent string) (streamID int) {
	streamID = 0
	uniqueIdentifier := fmt.Sprintf("%s-%s", ip, userAgent)

	for i := 0; i <= len(stream); i++ {
		if _, ok := stream[i]; !ok {
			streamID = i
			break
		}
	}

	if _, ok := stream[streamID]; ok && stream[streamID].ClientID == uniqueIdentifier {
		// Return the same ID if the combination already exists
		return streamID
	}

	return
}

const contentTypeSniffBytes = 512

type segmentInputReader struct {
	reader io.Reader
	err    error
}

func (reader *segmentInputReader) Read(buffer []byte) (int, error) {
	n, err := reader.reader.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		reader.err = err
	}
	return n, err
}

type segmentOutputWriter struct {
	writer io.Writer
	err    error
}

func (writer *segmentOutputWriter) Write(buffer []byte) (int, error) {
	n, err := writer.writer.Write(buffer)
	if err != nil {
		writer.err = err
	}
	return n, err
}

func transferSegment(
	destination io.Writer,
	segment io.ReadCloser,
	beforeWrite func([]byte),
) (inputErr, writeErr error) {
	reader := bufio.NewReaderSize(segment, contentTypeSniffBytes)
	prefix, prefixErr := reader.Peek(contentTypeSniffBytes)
	if prefixErr != nil && !errors.Is(prefixErr, io.EOF) {
		return errors.Join(prefixErr, segment.Close()), nil
	}
	if beforeWrite != nil {
		beforeWrite(prefix)
	}

	trackedInput := &segmentInputReader{reader: reader}
	trackedOutput := &segmentOutputWriter{writer: destination}
	_, transferErr := io.Copy(trackedOutput, trackedInput)
	closeErr := segment.Close()
	if trackedInput.err != nil {
		return errors.Join(trackedInput.err, closeErr), nil
	}
	if trackedOutput.err != nil {
		return closeErr, trackedOutput.err
	}
	if transferErr == io.ErrShortWrite {
		transferErr = nil
	}
	return closeErr, transferErr
}

func bufferingStream(playlistID string, streamingURL string, backupStream1 *BackupStream, backupStream2 *BackupStream, backupStream3 *BackupStream, channelName string, w http.ResponseWriter, r *http.Request) {
	systemMutex.Lock()
	bufferTimeout := Settings.BufferTimeout
	systemMutex.Unlock()
	select {
	case <-r.Context().Done():
		return
	case <-time.After(time.Duration(bufferTimeout) * time.Millisecond):
	}

	var streaming bool
	var debug string
	var timeOut int
	w.Header().Set("Connection", "close")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	playlist, exists := playlistSnapshot(playlistID)
	if !exists {
		playlist.Folder = System.Folder.Temp + playlistID + string(os.PathSeparator)
		playlist.PlaylistID = playlistID
		if err := checkVFSFolder(playlist.Folder, bufferVFS); err != nil {
			ShowError(err, 0)
			httpStatusError(w, r, 404)
			return
		}
		playlistType := "m3u"
		if strings.HasPrefix(playlistID, "H") {
			playlistType = "hdhr"
		}
		systemMutex.Lock()
		provider := Settings.Files.M3U[playlistID]
		if provider == nil {
			provider = Settings.Files.HDHR[playlistID]
		}
		playlist.Buffer = "-"
		if values, ok := provider.(map[string]interface{}); ok {
			if buffer, ok := values["buffer"].(string); ok {
				playlist.Buffer = buffer
			}
		}
		systemMutex.Unlock()
		playlist.Tuner = getTuner(playlistID, playlistType)
		playlist.PlaylistName = getProviderParameter(playlistID, playlistType, "name")
		playlist.HttpProxyIP = getProviderParameter(playlistID, playlistType, "http_proxy.ip")
		playlist.HttpProxyPort = getProviderParameter(playlistID, playlistType, "http_proxy.port")
		playlist.HttpUserOrigin = getProviderParameter(playlistID, playlistType, "http_headers.origin")
		playlist.HttpUserReferer = getProviderParameter(playlistID, playlistType, "http_headers.referer")
	}
	playlist, streamID, newStream, accepted := registerBufferClient(playlist, ThisStream{
		URL: streamingURL, ChannelName: channelName,
		BackupChannel1: backupStream1, BackupChannel2: backupStream2, BackupChannel3: backupStream3,
	}, getClientIP(r), r.UserAgent())
	if !accepted {
		for _, backup := range []*BackupStream{backupStream1, backupStream2, backupStream3} {
			if backup != nil {
				switch backup {
				case backupStream1:
					bufferingStream(backup.PlaylistID, backup.URL, nil, backupStream2, backupStream3, channelName, w, r)
				case backupStream2:
					bufferingStream(backup.PlaylistID, backup.URL, nil, nil, backupStream3, channelName, w, r)
				default:
					bufferingStream(backup.PlaylistID, backup.URL, nil, nil, nil, channelName, w, r)
				}
				return
			}
		}
		showInfo(fmt.Sprintf("Streaming Status:Playlist: %s - No new connections available. Tuner = %d", playlist.PlaylistName, playlist.Tuner))
		if value, ok := webUI["html/video/stream-limit.ts"]; ok {
			content := GetHTMLString(value.(string))
			w.Header().Set("Content-type", "video/mpeg")
			w.WriteHeader(200)
			for range 59 {
				if _, err := w.Write([]byte(content)); err != nil {
					return
				}
				select {
				case <-r.Context().Done():
					return
				case <-time.After(500 * time.Millisecond):
				}
			}
		}
		return
	}
	stream := playlist.Streams[streamID]
	defer releaseBufferClient(streamID, stream, false)
	if newStream {
		switch playlist.Buffer {
		case "ffmpeg", "vlc":
			go thirdPartyBuffer(streamID, playlist, stream)
		}
	}

	w.WriteHeader(200)

	for { //Loop 1: Wait until the first segment has been downloaded through the buffer
		select {
		case <-r.Context().Done():
			return
		case <-stream.done:
			return
		default:
		}

		if playlist, ok := playlistSnapshot(playlistID); ok {

			if stream, ok := playlist.Streams[streamID]; ok {
				connection, _ := BufferClients.Load(playlistID + stream.MD5)
				clients, _ := connection.(ClientConnection)

				if !stream.Status && clients.Error == nil {

					timeOut++

					time.Sleep(time.Duration(100) * time.Millisecond)

					if c, ok := BufferClients.Load(playlistID + stream.MD5); ok {

						var clients = c.(ClientConnection)

						if clients.Error != nil {
							// Re-read terminal state after the sleep so even a short,
							// closed first segment can be delivered.
							continue
						}
						if timeOut > 200 && stream.BackupChannel1 == nil && stream.BackupChannel2 == nil && stream.BackupChannel3 == nil {
							return
						}

					}

					continue
				}

				var oldSegments []string

				for { // Loop 2: Temporary files are present, data can be sent to the client
					var streamEnded bool

					// Monitor HTTP client connection

					ctx := r.Context()
					if ok {

						select {

						case <-ctx.Done():

							return

						default:
							if c, ok := BufferClients.Load(playlistID + stream.MD5); ok {

								var clients = c.(ClientConnection)
								streamEnded = clients.Error != nil

							} else {

								return

							}

						}

					}

					if _, err := bufferVFS.Stat(stream.Folder); fsIsNotExistErr(err) {

						return
					}

					var tmpFiles = getBufTmpFiles(&stream, streamEnded)
					if streamEnded && len(tmpFiles) == 0 {
						return
					}
					//fmt.Println("Buffer Loop:", stream.Connection)

					for _, f := range tmpFiles {

						if _, err := bufferVFS.Stat(stream.Folder); fsIsNotExistErr(err) {

							return
						}

						oldSegments = append(oldSegments, f)

						var fileName = stream.Folder + f

						file, err := bufferVFS.Open(fileName)
						if err != nil {
							debug = fmt.Sprintf("Buffer Open (%s)", fileName)
							showDebug(debug, 2)
							return
						}
						inputErr, writeErr := transferSegment(w, file, func(buffer []byte) {
							debug = fmt.Sprintf("Buffer Status:Send to client (%s)", fileName)
							showDebug(debug, 2)

							if !streaming {
								contentType := http.DetectContentType(buffer)
								w.Header().Set("Content-type", contentType)
								w.Header().Set("Content-Length", "0")
								w.Header().Set("Connection", "close")
							}
						})
						if inputErr != nil {
							ShowError(inputErr, 0)

							return
						}
						if writeErr != nil {

							return
						}
						streaming = true

						var n = slices.Index(oldSegments, f)

						if n > 20 {

							var fileToRemove = stream.Folder + oldSegments[0]
							if err = bufferVFS.RemoveAll(getPlatformFile(fileToRemove)); err != nil {
								ShowError(err, 4007)
							}
							oldSegments = append(oldSegments[:0], oldSegments[0+1:]...)

						}

					}

					if len(tmpFiles) == 0 {
						time.Sleep(time.Duration(100) * time.Millisecond)
					}

				} // End Loop 2

			} else {

				// Stream not found
				showDebug("Streaming Status:Stream not found. Killing Connection", 3)

				showInfo(fmt.Sprintf("Streaming Status:Playlist: %s - Tuner: %d / %d", playlist.PlaylistName, len(playlist.Streams), playlist.Tuner))
				return

			}

		} else {
			return
		} // End BufferInformation

	} // End Loop 1

}

func getBufTmpFiles(stream *ThisStream, streamEnded bool) (tmpFiles []string) {

	var tmpFolder = stream.Folder
	var fileIDs []int

	if _, err := bufferVFS.Stat(tmpFolder); !fsIsNotExistErr(err) {

		files, err := bufferVFS.ReadDir(getPlatformPath(tmpFolder))
		if err != nil {
			ShowError(err, 000)
			return
		}

		for _, file := range files {
			if !file.Type().IsRegular() {
				continue
			}
			fileID, ok := parseSegmentFilename(file.Name())
			if ok {
				fileIDs = append(fileIDs, fileID)
			}
		}

		if len(fileIDs) > 0 {
			sort.Ints(fileIDs)
			// Only the running producer's highest file may still be changing.
			// Terminal state is published after all producer files are closed.
			if !streamEnded {
				fileIDs = fileIDs[:len(fileIDs)-1]
			}

			for _, fileID := range fileIDs {
				fileName := strconv.Itoa(fileID) + ".ts"
				if slices.Index(stream.OldSegments, fileName) == -1 {
					tmpFiles = append(tmpFiles, fileName)
					stream.OldSegments = append(stream.OldSegments, fileName)
				}
			}
		}

	}

	return
}

func killClientConnection(streamID int, playlistID string, force bool) {
	playlist, ok := playlistSnapshot(playlistID)
	if !ok {
		return
	}
	if stream, ok := playlist.Streams[streamID]; ok {
		releaseBufferClient(streamID, stream, force)
	}
}

func releaseBufferClient(streamID int, stream ThisStream, force bool) {
	Lock.Lock()
	defer Lock.Unlock()
	value, ok := BufferInformation.Load(stream.PlaylistID)
	if !ok {
		return
	}
	playlist := value.(Playlist)
	current, ok := playlist.Streams[streamID]
	if !ok || current.done != stream.done || current.MD5 != stream.MD5 {
		return
	}
	client := playlist.Clients[streamID]
	client.Connection = max(0, client.Connection-1)
	if force {
		client.Connection = 0
	}
	if client.Connection == 0 {
		if current.done != nil {
			close(current.done)
		}
		BufferClients.Delete(stream.PlaylistID + stream.MD5)
		delete(playlist.Streams, streamID)
		delete(playlist.Clients, streamID)
		if len(playlist.Streams) == 0 {
			BufferInformation.Delete(stream.PlaylistID)
		}
	} else {
		playlist.Clients[streamID] = client
		value, _ := BufferClients.Load(stream.PlaylistID + stream.MD5)
		clients, _ := value.(ClientConnection)
		clients.Connection = client.Connection
		BufferClients.Store(stream.PlaylistID+stream.MD5, clients)
	}
}

func clientConnection(stream ThisStream) bool {
	Lock.RLock()
	defer Lock.RUnlock()
	if stream.done != nil {
		select {
		case <-stream.done:
			return false
		default:
		}
	}
	value, ok := BufferClients.Load(stream.PlaylistID + stream.MD5)
	return ok && value.(ClientConnection).Connection > 0
}

// A late producer must neither resurrect a deleted playlist nor overwrite a
// replacement stream (or another client's updated connection count).
func updateBufferStream(streamID int, stream ThisStream, ready bool, streamErr error) {
	Lock.Lock()
	defer Lock.Unlock()
	value, ok := BufferInformation.Load(stream.PlaylistID)
	if !ok {
		return
	}
	playlist := value.(Playlist)
	current, ok := playlist.Streams[streamID]
	if !ok || current.done != stream.done || current.MD5 != stream.MD5 {
		return
	}
	if ready {
		current.Status = true
		playlist.Streams[streamID] = current
	}
	if streamErr != nil {
		if value, ok := BufferClients.Load(stream.PlaylistID + stream.MD5); ok {
			clients := value.(ClientConnection)
			clients.Error = streamErr
			BufferClients.Store(stream.PlaylistID+stream.MD5, clients)
		}
	}
}

func switchBandwidth(stream *ThisStream) (err error) {

	bandwidth := slices.Sorted(maps.Keys(stream.DynamicStream))
	var dynamicStream DynamicStream
	var segment Segment

	if len(bandwidth) > 0 {

		for i := range bandwidth {

			segment.StreamInf.Bandwidth = stream.DynamicStream[bandwidth[i]].Bandwidth
			dynamicStream = stream.DynamicStream[bandwidth[0]]

			if stream.NetworkBandwidth == 0 {

				dynamicStream = stream.DynamicStream[bandwidth[0]]
				break

			} else {

				if bandwidth[i] > stream.NetworkBandwidth {
					break
				}

				dynamicStream = stream.DynamicStream[bandwidth[i]]

			}

		}

	} else {

		err = errors.New("M3U8 does not contain streaming URLs")
		return

	}

	segment.URL = dynamicStream.URL
	segment.Duration = 0
	stream.Segment = append(stream.Segment, segment)

	return
}

var errSegmentNumberOverflow = errors.New("segment number overflow")

func incrementSegmentNumber(segment int) (int, error) {
	maxInt := int(^uint(0) >> 1)
	if segment == maxInt {
		return segment, errSegmentNumberOverflow
	}
	return segment + 1, nil
}

func isFirstSegment(segment, startSegment int) bool {
	return segment == startSegment
}

func parseSegmentFilename(name string) (int, bool) {
	numberText, ok := strings.CutSuffix(name, ".ts")
	if !ok || numberText == "" {
		return 0, false
	}
	number, err := strconv.Atoi(numberText)
	if err != nil || number < 1 || strconv.Itoa(number) != numberText {
		return 0, false
	}
	return number, true
}

func nextBackupSegment(folder string) (int, error) {
	entries, err := bufferVFS.ReadDir(getPlatformPath(folder))
	if err != nil {
		return 0, err
	}

	highest := 0
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}
		number, ok := parseSegmentFilename(entry.Name())
		if !ok {
			continue
		}
		if number > highest {
			highest = number
		}
	}
	if highest == 0 {
		return 1, nil
	}
	return incrementSegmentNumber(highest)
}

func prepareSegmentDirectory(folder string, useBackup bool) (int, error) {
	if !useBackup {
		if err := bufferVFS.RemoveAll(getPlatformPath(folder)); err != nil {
			ShowError(err, 4005)
		}
	}
	if err := checkVFSFolder(folder, bufferVFS); err != nil {
		return 0, err
	}
	if !useBackup {
		return 1, nil
	}
	return nextBackupSegment(folder)
}

// streamURLForLog intentionally omits the complete path: IPTV providers often
// put usernames and passwords in /live/user/password/channel, not just userinfo.
func streamURLForLog(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "[redacted URL]"
	}
	return parsed.Scheme + "://" + parsed.Host
}

func thirdPartyBuffer(streamID int, playlist Playlist, stream ThisStream) {
	defer func() {
		// The process has stopped, but readers may still need its completed
		// segments. The last reader releases this producer's unique directory.
		if stream.done != nil {
			<-stream.done
		}
		_ = bufferVFS.RemoveAll(stream.Folder)
	}()
	sources := []string{stream.URL, "", "", ""}
	for index, backup := range []*BackupStream{stream.BackupChannel1, stream.BackupChannel2, stream.BackupChannel3} {
		if backup != nil {
			sources[index+1] = backup.URL
		}
	}
	var streamErr error
	for index := range sources {
		if sources[index] == "" {
			continue
		}
		if !clientConnection(stream) {
			return
		}
		if index > 0 {
			showInfo(fmt.Sprintf("Backup Channel %d:%s", index, streamURLForLog(sources[index])))
		}
		streamErr = thirdPartyBufferAttempt(streamID, playlist, stream, sources[index], index > 0)
		if errors.Is(streamErr, errNoBufferClients) {
			return
		}
	}
	if streamErr != nil {
		updateBufferStream(streamID, stream, false, streamErr)
	}
}

var errNoBufferClients = errors.New("no clients remain for stream")
var errBufferStartupTimeout = errors.New("buffer startup timed out")

// A separate pipe reader lets timeout and cancellation win even when the child
// never writes stdout. This function owns Start, Kill, Wait and its reader.
func consumeBufferProcess(cmd *exec.Cmd, stream ThisStream, timeout time.Duration, write func([]byte) (bool, error)) (resultErr error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.New("cannot open buffer process output")
	}
	// Leave stderr connected to the null device. Child diagnostics can repeat
	// passwords, query tokens, headers and private URL paths verbatim.
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		_ = stdout.Close()
		return errors.New("cannot start buffer process")
	}
	type readResult struct {
		data []byte
		err  error
	}
	results := make(chan readResult)
	stop := make(chan struct{})
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		for {
			buffer := make([]byte, 4096)
			n, err := stdout.Read(buffer)
			select {
			case results <- readResult{buffer[:n], err}:
			case <-stop:
				return
			}
			if err != nil {
				return
			}
		}
	}()
	defer func() {
		close(stop)
		_ = stdout.Close()
		resultErr = errors.Join(resultErr, terminateProcess(cmd))
		<-readerDone
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	deadline := timer.C
	clients := time.NewTicker(100 * time.Millisecond)
	defer clients.Stop()
	for {
		select {
		case <-stream.done:
			return errNoBufferClients
		case <-clients.C:
			if !clientConnection(stream) {
				return errNoBufferClients
			}
		case <-deadline:
			return errBufferStartupTimeout
		case result := <-results:
			if len(result.data) > 0 {
				ready, err := write(result.data)
				if err != nil {
					return err
				}
				if ready {
					timer.Stop()
					deadline = nil
				}
			}
			if result.err != nil {
				return errors.New("buffer process stopped producing data")
			}
		}
	}
}

func thirdPartyBufferAttempt(streamID int, playlist Playlist, stream ThisStream, source string, backup bool) error {
	systemMutex.Lock()
	settings := Settings
	systemMutex.Unlock()
	var path, options string
	bufferType := strings.ToUpper(playlist.Buffer)
	url := source
	switch playlist.Buffer {
	case "ffmpeg":
		if settings.FFmpegForceHttp {
			url = strings.Replace(url, "https://", "http://", 1)
		}
		path, options = settings.FFmpegPath, settings.FFmpegOptions
	case "vlc":
		path, options = settings.VLCPath, settings.VLCOptions
	default:
		return errors.New("unsupported buffer process")
	}
	segment, err := prepareSegmentDirectory(stream.Folder, backup)
	if err != nil {
		return err
	}
	fileName := fmt.Sprintf("%s%d.ts", stream.Folder, segment)
	file, err := bufferVFS.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()
	showInfo("Streaming URL:" + streamURLForLog(url))
	showInfo(bufferType + ":Processing data")
	// Set User-Agent
	var args []string

	for i, a := range strings.Split(options, " ") {

		switch bufferType {
		case "FFMPEG":
			a = strings.Replace(a, "[URL]", url, -1)
			if i == 0 {
				if len(settings.UserAgent) != 0 {
					args = []string{"-user_agent", settings.UserAgent}
				}

				if playlist.HttpProxyIP != "" && playlist.HttpProxyPort != "" {
					args = append(args, "-http_proxy", fmt.Sprintf("http://%s:%s", playlist.HttpProxyIP, playlist.HttpProxyPort))
				}

				var headers string
				if len(playlist.HttpUserReferer) != 0 {
					headers += fmt.Sprintf("Referer: %s\r\n", playlist.HttpUserReferer)
				}
				if len(playlist.HttpUserOrigin) != 0 {
					headers += fmt.Sprintf("Origin: %s\r\n", playlist.HttpUserOrigin)
				}
				if headers != "" {
					args = append(args, "-headers", headers)
				}
			}

			args = append(args, a)

		case "VLC":
			if a == "[URL]" {
				a = strings.Replace(a, "[URL]", url, -1)
				args = append(args, a)

				if len(settings.UserAgent) != 0 {
					args = append(args, fmt.Sprintf(":http-user-agent=%s", settings.UserAgent))
				}

				if len(playlist.HttpUserReferer) != 0 {
					args = append(args, fmt.Sprintf(":http-referrer=%s", playlist.HttpUserReferer))
				}

				if playlist.HttpProxyIP != "" && playlist.HttpProxyPort != "" {
					args = append(args, fmt.Sprintf(":http-proxy=%s:%s", playlist.HttpProxyIP, playlist.HttpProxyPort))
				}

			} else {
				args = append(args, a)
			}

		}

	}
	cmd := exec.Command(path, args...)
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
	fileSize := 0
	ready := false
	return consumeBufferProcess(cmd, stream, 20*time.Second, func(data []byte) (bool, error) {
		if _, err := file.Write(data); err != nil {
			return ready, err
		}
		fileSize += len(data)
		if fileSize < max(1, settings.BufferSize*1024/2) {
			return ready, nil
		}
		if err := file.Close(); err != nil {
			file = nil
			return ready, err
		}
		file = nil
		segment, err = incrementSegmentNumber(segment)
		if err != nil {
			return ready, err
		}
		fileName = fmt.Sprintf("%s%d.ts", stream.Folder, segment)
		file, err = bufferVFS.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return ready, err
		}
		fileSize = 0
		if !ready {
			updateBufferStream(streamID, stream, true, nil)
			ready = true
		}
		return ready, nil
	})
}

func createBufferFile(path string) error {
	file, err := bufferVFS.Create(path)
	if err != nil {
		return err
	}
	return file.Close()
}

func getTuner(id, playlistType string) (tuner int) {

	var playListBuffer string
	systemMutex.Lock()
	playListInterface := Settings.Files.M3U[id]
	if playListInterface == nil {
		playListInterface = Settings.Files.HDHR[id]
	}
	if playListMap, ok := playListInterface.(map[string]interface{}); ok {
		if buffer, ok := playListMap["buffer"].(string); ok {
			playListBuffer = buffer
		} else {
			playListBuffer = "-"
		}
	}
	systemMutex.Unlock()

	switch playListBuffer {

	case "-":
		tuner = Settings.Tuner

	case "threadfin", "ffmpeg", "vlc":

		i, err := strconv.Atoi(getProviderParameter(id, playlistType, "tuner"))
		if err == nil {
			tuner = i
		} else {
			ShowError(err, 0)
			tuner = 1
		}

	}

	return
}

var bufferVFSInitMutex sync.Mutex

func initBufferVFS() {
	bufferVFSInitMutex.Lock()
	defer bufferVFSInitMutex.Unlock()
	if bufferVFS == nil {
		bufferVFS = memfs.New()
	}
}

func terminateProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	killErr := cmd.Process.Kill()
	waitErr := cmd.Wait()
	if killErr != nil && !errors.Is(killErr, os.ErrProcessDone) {
		return killErr
	}
	if waitErr != nil {
		if _, ok := errors.AsType[*exec.ExitError](waitErr); !ok {
			return waitErr
		}
	}
	return nil
}
