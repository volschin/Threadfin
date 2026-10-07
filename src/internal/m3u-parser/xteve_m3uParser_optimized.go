package m3u

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Precompiled regex patterns for better performance
var (
	parameterRegex   = regexp.MustCompile(`[a-z-A-Z&=]*(".*?")`)
	channelNameRegex = regexp.MustCompile(`,([^\n]*|,[^\r]*)`)
	quoteReplacer    = strings.NewReplacer(`"`, "")
)

// MakeInterfaceFromM3UOptimized : Optimized version for large M3U files
func MakeInterfaceFromM3UOptimized(byteStream []byte) (allChannels []interface{}, err error) {
	// Use bytes.Contains for faster validation
	if bytes.Contains(byteStream, []byte("#EXT-X-TARGETDURATION")) || bytes.Contains(byteStream, []byte("#EXT-X-MEDIA-SEQUENCE")) {
		err = errors.New("Invalid M3U file, an extended M3U file is required.")
		return
	}

	if !bytes.Contains(byteStream, []byte("#EXTM3U")) {
		err = errors.New("Invalid M3U file, an extended M3U file is required.")
		return
	}

	// Use scanner for line-by-line processing instead of loading full content
	scanner := bufio.NewScanner(bytes.NewReader(byteStream))
	// The playlist is already in memory; allow its full line length so small
	// and large inputs share the same parsing limits.
	scanner.Buffer(make([]byte, 0, 64*1024), len(byteStream)+1)

	var extinfLine string

	// Pre-allocate channels slice with estimated capacity
	estimatedChannels := bytes.Count(byteStream, []byte("#EXTINF"))
	allChannels = make([]interface{}, 0, estimatedChannels)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Process #EXTINF lines
		if strings.HasPrefix(line, "#EXTINF") {
			extinfLine = strings.ReplaceAll(line, "'", `"`)
			continue
		}

		// Skip other # lines
		if strings.HasPrefix(line, "#") {
			continue
		}

		// This is a URL line
		if extinfLine != "" {
			if stream := parseMetaDataOptimized(extinfLine, line); len(stream) > 0 {
				allChannels = append(allChannels, stream)
			}
			extinfLine = ""
		}
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return allChannels, nil
}

// parseMetaDataOptimized : Optimized metadata parsing
func parseMetaDataOptimized(extinfLine, streamURL string) map[string]string {
	stream := make(map[string]string, 12) // Pre-allocate with typical size
	stream["url"] = strings.TrimSpace(streamURL)

	// Extract parameters using pre-compiled regex
	var value strings.Builder
	streamParameter := parameterRegex.FindAllString(extinfLine, -1)

	for _, p := range streamParameter {
		extinfLine = strings.Replace(extinfLine, p, "", 1)
		cleanParam := quoteReplacer.Replace(p)

		if paramParts := strings.SplitN(cleanParam, "=", 2); len(paramParts) == 2 {
			key, val := paramParts[0], paramParts[1]

			// Save TVG Key in lowercase
			if strings.Contains(strings.ToLower(key), "tvg") {
				stream[strings.ToLower(key)] = val
			} else {
				stream[key] = val
			}

			// Build value string (skip URLs)
			if !strings.Contains(val, "://") && len(val) > 0 {
				value.WriteString(val)
				value.WriteByte(' ')
			}
		}
	}

	// Parse channel name using pre-compiled regex
	var channelName string
	if nameMatches := channelNameRegex.FindStringSubmatch(extinfLine); len(nameMatches) > 1 {
		channelName = nameMatches[1]
		channelName = strings.TrimSpace(channelName)
	}

	// Fallback to tvg-name if no channel name found
	if len(channelName) == 0 {
		if tvgName, ok := stream["tvg-name"]; ok {
			channelName = tvgName
		}
	}

	// Skip channels without a name
	if len(channelName) == 0 {
		return nil
	}

	// Generate tvg-id if missing
	if tvgID := stream["tvg-id"]; tvgID == "" || tvgID == "(no tvg-id)" {
		hash := md5.Sum([]byte(stream["url"]))
		stream["tvg-id"] = fmt.Sprintf("threadfin-%x", hash)
	}

	stream["name"] = strings.TrimSpace(channelName)
	value.WriteString(channelName)
	stream["_values"] = value.String()

	// Set UUID for tvg-name (simplified)
	if tvgName, ok := stream["tvg-name"]; ok && len(tvgName) > 0 {
		stream["_uuid.key"] = "tvg-name"
		stream["_uuid.value"] = tvgName
	}

	return stream
}

// Wrapper to maintain backward compatibility
func MakeInterfaceFromM3U(byteStream []byte) (allChannels []interface{}, err error) {
	return MakeInterfaceFromM3UOptimized(byteStream)
}
