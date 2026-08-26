package src

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	m3u "threadfin/src/internal/m3u-parser"
)

// Playlisten parsen
var (
	filterExcludeRegexp   = regexp.MustCompile(`!+[{]+[^.]+[}]`)
	filterIncludeRegexp   = regexp.MustCompile(`[{]+[^.]+[}]`)
	createM3UStreamingURL = createStreamingURL
)

func parsePlaylist(filename, fileType string) (channels []interface{}, err error) {

	content, err := readByteFromFile(filename)
	var id = strings.TrimSuffix(getFilenameFromPath(filename), path.Ext(getFilenameFromPath(filename)))
	var playlistName = getProviderParameter(id, fileType, "name")

	if err == nil {

		switch fileType {
		case "m3u":
			channels, err = m3u.MakeInterfaceFromM3U(content)
		case "hdhr":
			channels, err = makeInteraceFromHDHR(content, playlistName, id)
		}

	}

	return
}

// Streams filtern
func filterThisStream(s interface{}) (status bool, liveEvent bool) {

	var stream = s.(map[string]string)
	liveEvent = false

	for _, filter := range Data.Filter {

		if filter.Rule == "" {
			continue
		}

		var group, name, search string
		var exclude, include string
		var match = false

		var streamValues = strings.Replace(stream["_values"], "\r", "", -1)

		if v, ok := stream["group-title"]; ok {
			group = v
		}

		if v, ok := stream["name"]; ok {
			name = v
		}

		// Unerwünschte Streams !{DEU}
		val := filterExcludeRegexp.FindStringSubmatch(filter.Rule)

		if len(val) == 1 {

			exclude = val[0][2 : len(val[0])-1]
			filter.Rule = strings.Replace(filter.Rule, " "+val[0], "", -1)
			filter.Rule = strings.Replace(filter.Rule, val[0], "", -1)

		}

		// Muss zusätzlich erfüllt sein {DEU}
		val = filterIncludeRegexp.FindStringSubmatch(filter.Rule)

		if len(val) == 1 {

			include = val[0][1 : len(val[0])-1]
			filter.Rule = strings.Replace(filter.Rule, " "+val[0], "", -1)
			filter.Rule = strings.Replace(filter.Rule, val[0], "", -1)

		}

		switch filter.CaseSensitive {

		case false:

			streamValues = strings.ToLower(streamValues)
			filter.Rule = strings.ToLower(filter.Rule)
			exclude = strings.ToLower(exclude)
			include = strings.ToLower(include)
			group = strings.ToLower(group)
			name = strings.ToLower(name)

		}

		switch filter.Type {

		case "group-title":
			search = name

			if group == filter.Rule {
				match = true
			}

		case "custom-filter":
			search = streamValues
			if strings.Contains(search, filter.Rule) {
				match = true
			}
		}

		if match == true {

			if len(exclude) > 0 && !checkConditions(search, exclude, "exclude") {
				continue
			}

			if len(include) > 0 && !checkConditions(search, include, "include") {
				continue
			}

			return true, filter.LiveEvent

		}

	}

	return false, false
}

// Bedingungen für den Filter
func checkConditions(streamValues, conditions, coType string) (status bool) {

	switch coType {

	case "exclude":
		status = true

	case "include":
		status = false

	}

	conditions = strings.Replace(conditions, ", ", ",", -1)
	conditions = strings.Replace(conditions, " ,", ",", -1)

	for key := range strings.SplitSeq(conditions, ",") {

		if strings.Contains(streamValues, key) {

			switch coType {

			case "exclude":
				return false

			case "include":
				return true

			}

		}

	}

	return
}

func parseGroupCountLabel(label string) (name string, count int, ok bool) {
	name, remainder, found := strings.CutLast(label, " (")
	if !found || name == "" {
		return "", 0, false
	}

	countText, _, found := strings.CutLast(remainder, ")")
	if !found || countText == "" {
		return "", 0, false
	}

	count, err := strconv.Atoi(countText)
	if err != nil {
		return "", 0, false
	}

	return name, count, true
}

func compareChannelNumbers(a, b XEPGChannelStruct) int {
	aNumber, aErr := strconv.ParseFloat(a.TvgChno, 64)
	bNumber, bErr := strconv.ParseFloat(b.TvgChno, 64)

	switch {
	case aErr == nil && bErr == nil:
		if aNumber < bNumber {
			return -1
		}
		if aNumber > bNumber {
			return 1
		}
		return 0
	case aErr == nil:
		return -1
	case bErr == nil:
		return 1
	case a.TvgChno < b.TvgChno:
		return -1
	case a.TvgChno > b.TvgChno:
		return 1
	default:
		return 0
	}
}

type m3uTempFile interface {
	io.Writer
	Name() string
	Chmod(os.FileMode) error
	Sync() error
	Close() error
}

type m3uPublicationOps struct {
	createTemp func(string, string) (m3uTempFile, error)
	stat       func(string) (os.FileInfo, error)
	rename     func(string, string) error
	remove     func(string) error
}

// defaultM3UPublishedMode keeps a first publication owner-only. Replacements
// preserve the explicit mode of an existing published playlist.
const defaultM3UPublishedMode os.FileMode = 0o600

func publishM3UFile(filename string, write func(io.StringWriter) error) error {
	return publishM3UFileWithOps(filename, write, m3uPublicationOps{
		createTemp: func(directory, pattern string) (m3uTempFile, error) {
			return os.CreateTemp(directory, pattern)
		},
		stat:   os.Stat,
		rename: os.Rename,
		remove: os.Remove,
	})
}

func publishM3UFileWithOps(filename string, write func(io.StringWriter) error, ops m3uPublicationOps) (err error) {
	publishedMode := defaultM3UPublishedMode
	stat := ops.stat
	if stat == nil {
		stat = os.Stat
	}
	info, statErr := stat(filename)
	switch {
	case statErr == nil:
		publishedMode = info.Mode().Perm()
	case !errors.Is(statErr, fs.ErrNotExist):
		return fmt.Errorf("inspect published M3U mode: %w", statErr)
	}

	temporary, err := ops.createTemp(filepath.Dir(filename), ".threadfin.m3u-*")
	if err != nil {
		return fmt.Errorf("create temporary M3U file: %w", err)
	}

	temporaryName := temporary.Name()
	closed := false
	published := false
	defer func() {
		if !closed {
			err = errors.Join(err, temporary.Close())
		}
		if !published {
			removeErr := ops.remove(temporaryName)
			if removeErr != nil && !errors.Is(removeErr, fs.ErrNotExist) {
				err = errors.Join(err, fmt.Errorf("remove temporary M3U file: %w", removeErr))
			}
		}
	}()

	if err := temporary.Chmod(publishedMode); err != nil {
		return fmt.Errorf("set temporary M3U mode: %w", err)
	}

	writer := bufio.NewWriterSize(temporary, 1<<20)
	if err := write(writer); err != nil {
		return fmt.Errorf("write complete M3U: %w", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush complete M3U: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync complete M3U: %w", err)
	}
	closed = true
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close complete M3U: %w", err)
	}
	if err := ops.rename(temporaryName, filename); err != nil {
		return fmt.Errorf("publish complete M3U: %w", err)
	}
	published = true
	return nil
}

// Threadfin M3U Datei erstellen
func buildM3U(groups []string) (m3u string, err error) {

	var imgc = Data.Cache.Images
	// Preserve every active channel as a distinct entry by iteration order, not keyed by number
	type channelEntry struct {
		idx int
		ch  XEPGChannelStruct
	}
	var entries []channelEntry

	// Build a map of group -> expectedCount from Data.Playlist.M3U.Groups.Text (format: "Group (N)")
	expectedGroupCount := make(map[string]int)
	for _, label := range Data.Playlist.M3U.Groups.Text {
		name, count, ok := parseGroupCountLabel(label)
		if ok {
			expectedGroupCount[name] = count
		}
	}

	// Count deactivated channels per group
	deactivatedPerGroup := make(map[string]int)
	for _, dxc := range Data.XEPG.Channels {
		var ch XEPGChannelStruct
		if err := json.Unmarshal([]byte(mapToJSON(dxc)), &ch); err == nil {
			group := ch.XGroupTitle
			if ch.XCategory != "" {
				group = ch.XCategory
			}
			if group == "" {
				group = ch.GroupTitle
			}
			if !ch.XActive || ch.XHideChannel {
				deactivatedPerGroup[group] = deactivatedPerGroup[group] + 1
			}
		}
	}

	for _, dxc := range Data.XEPG.Channels {
		var xepgChannel XEPGChannelStruct
		err := json.Unmarshal([]byte(mapToJSON(dxc)), &xepgChannel)
		if err == nil {
			if xepgChannel.TvgName == "" {
				xepgChannel.TvgName = xepgChannel.Name
			}
			if xepgChannel.XActive && !xepgChannel.XHideChannel {
				if len(groups) > 0 {

					if slices.Index(groups, xepgChannel.XGroupTitle) == -1 {
						goto Done
					}

				}
				entries = append(entries, channelEntry{idx: len(entries), ch: xepgChannel})
			}
		}

	Done:
	}

	// Prepare header
	var xmltvURL = fmt.Sprintf("%s://%s/xmltv/threadfin.xml", System.ServerProtocol.XML, System.Domain)
	if Settings.ForceHttps && Settings.HttpsThreadfinDomain != "" {
		xmltvURL = fmt.Sprintf("https://%s/xmltv/threadfin.xml", Settings.HttpsThreadfinDomain)
	}
	header := fmt.Sprintf(`#EXTM3U url-tvg="%s" x-tvg-url="%s"`+"\n", xmltvURL, xmltvURL)

	// Sort entries by tvg-chno numerically
	slices.SortFunc(entries, func(a, b channelEntry) int {
		return compareChannelNumbers(a.ch, b.ch)
	})

	render := func(output io.StringWriter) error {
		if _, err := output.WriteString(header); err != nil {
			return err
		}

		// Avoid duplicate exact stream URLs within the same group and cap per-group by expected minus deactivated
		seenURLInGroup := make(map[string]struct{})
		emittedGroupCount := make(map[string]int)
		for _, e := range entries {
			var channel = e.ch

			group := channel.XGroupTitle
			if channel.XCategory != "" {
				group = channel.XCategory
			}
			if group == "" {
				group = channel.GroupTitle
			}

			// Determine allowed active count = expected - deactivated
			if expected, ok := expectedGroupCount[group]; ok {
				allowed := max(expected-deactivatedPerGroup[group], 0)
				if emittedGroupCount[group] >= allowed {
					continue
				}
			}

			// Disabling so not to rewrite stream to https domain when disable stream from https set
			if Settings.ForceHttps && Settings.HttpsThreadfinDomain != "" && Settings.ExcludeStreamHttps == false {
				u, err := url.Parse(channel.URL)
				if err == nil {
					u.Scheme = "https"
					host_split := strings.Split(u.Host, ":")
					if len(host_split) > 0 {
						u.Host = host_split[0]
					}
					if u.RawQuery != "" {
						channel.URL = fmt.Sprintf("https://%s:%d%s?%s", u.Host, Settings.HttpsPort, u.Path, u.RawQuery)
					} else {
						channel.URL = fmt.Sprintf("https://%s:%d%s", u.Host, Settings.HttpsPort, u.Path)
					}
				}
			}

			logo := ""
			if channel.TvgLogo != "" {
				logo = imgc.Image.GetURL(channel.TvgLogo, Settings.HttpThreadfinDomain, Settings.Port, Settings.ForceHttps, Settings.HttpsPort, Settings.HttpsThreadfinDomain)
			}
			parameter := fmt.Sprintf(`#EXTINF:0 channelID="%s" tvg-chno="%s" tvg-name="%s" tvg-id="%s" tvg-logo="%s" group-title="%s",%s`+"\n", channel.XEPG, channel.XChannelID, channel.XName, channel.XChannelID, logo, group, channel.XName)
			stream, err := createM3UStreamingURL("M3U", channel.FileM3UID, channel.XChannelID, channel.XName, channel.URL, channel.BackupChannel1, channel.BackupChannel2, channel.BackupChannel3)
			if err != nil {
				return err
			}
			key := group + "|" + stream
			if _, ok := seenURLInGroup[key]; ok {
				continue
			}
			seenURLInGroup[key] = struct{}{}
			if _, err := output.WriteString(parameter); err != nil {
				return err
			}
			if _, err := output.WriteString(stream + "\n"); err != nil {
				return err
			}
			emittedGroupCount[group] = emittedGroupCount[group] + 1
		}

		return nil
	}

	filename := filepath.Join(System.Folder.Data, "threadfin.m3u")
	if len(groups) == 0 {
		err = publishM3UFile(filename, render)
		return "", err
	}

	var output strings.Builder
	if err = render(&output); err != nil {
		return "", err
	}
	return output.String(), nil
}

func probeChannel(request RequestStruct) (string, string, string, error) {

	ffmpegPath := Settings.FFmpegPath
	ffprobePath := strings.Replace(ffmpegPath, "ffmpeg", "ffprobe", 1)

	cmd := exec.Command(ffprobePath, "-v", "error", "-show_streams", "-of", "json", request.ProbeURL)
	output, err := cmd.Output()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to execute ffprobe: %v", err)
	}

	var ffprobeOutput FFProbeOutput
	err = json.Unmarshal(output, &ffprobeOutput)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse ffprobe output: %v", err)
	}

	var resolution, frameRate, audioChannels string

	for _, stream := range ffprobeOutput.Streams {
		if stream.CodecType == "video" {
			resolution = fmt.Sprintf("%dp", stream.Height)
			frameRateParts := strings.Split(stream.RFrameRate, "/")
			if len(frameRateParts) == 2 {
				frameRate = fmt.Sprintf("%d", parseFrameRate(frameRateParts))
			} else {
				frameRate = stream.RFrameRate
			}
		}
		if stream.CodecType == "audio" {
			audioChannels = stream.ChannelLayout
			if audioChannels == "" {
				switch stream.Channels {
				case 1:
					audioChannels = "Mono"
				case 2:
					audioChannels = "Stereo"
				case 6:
					audioChannels = "5.1"
				case 8:
					audioChannels = "7.1"
				default:
					audioChannels = fmt.Sprintf("%d channels", stream.Channels)
				}
			}
		}
	}

	return resolution, frameRate, audioChannels, nil
}

func parseFrameRate(parts []string) int {
	if len(parts) != 2 {
		return 0
	}
	numerator, numeratorErr := strconv.Atoi(parts[0])
	denom, denominatorErr := strconv.Atoi(parts[1])
	if numeratorErr != nil || denominatorErr != nil || denom == 0 {
		return 0
	}
	return int(math.Round(float64(numerator) / float64(denom)))
}
