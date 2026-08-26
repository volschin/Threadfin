package src

import (
	"encoding/json"
	"testing"
)

var benchmarkMapJSONResult string

func benchmarkChannelValue() map[string]interface{} {
	return map[string]interface{}{
		"_file.m3u.id":   "M3U-001",
		"_file.m3u.name": "Benchmark Provider",
		"group-title":    "News",
		"name":           "Benchmark News HD",
		"tvg-id":         "benchmark.news",
		"tvg-logo":       "https://example.test/images/news.png",
		"tvg-name":       "Benchmark News",
		"tvg-chno":       "101",
		"url":            "https://example.test/live/news.ts",
		"x-active":       true,
		"x-category":     "News",
		"x-channelID":    "threadfin-101",
		"x-name":         "Benchmark News HD",
		"backup_channel_1": map[string]interface{}{
			"PlaylistID": "M3U-002",
			"URL":        "https://backup.example.test/live/news.ts",
		},
	}
}

func benchmarkConfigurationValue() SettingsStruct {
	settings := SettingsStruct{
		API:                     true,
		AuthenticationAPI:       true,
		BackupKeep:              10,
		Branch:                  "main",
		Buffer:                  "ffmpeg",
		BufferSize:              4096,
		CacheImages:             true,
		EnableNonAscii:          true,
		EpgSource:               "XEPG",
		Language:                "en",
		Port:                    "34400",
		ThreadfinAutoUpdate:     true,
		Tuner:                   4,
		UserAgent:               "Threadfin benchmark",
		XepgReplaceChannelTitle: true,
	}
	settings.Files.M3U = map[string]interface{}{
		"M3U-001": map[string]interface{}{
			"file.source": "https://example.test/playlist.m3u",
			"name":        "Benchmark Provider",
			"tuner":       4,
		},
	}
	settings.Files.XMLTV = map[string]interface{}{
		"XMLTV-001": map[string]interface{}{
			"file.source": "https://example.test/guide.xml",
			"name":        "Benchmark Guide",
		},
	}
	settings.Filter = map[int64]interface{}{
		0: map[string]interface{}{
			"active":  true,
			"include": "sports,news,movies",
			"name":    "Include groups",
			"type":    "group-title",
		},
	}
	return settings
}

func BenchmarkMapToJSON(b *testing.B) {
	cases := []struct {
		name  string
		value interface{}
	}{
		{name: "Channel", value: benchmarkChannelValue()},
		{name: "Configuration", value: benchmarkConfigurationValue()},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			encoded := mapToJSON(tc.value)
			if !json.Valid([]byte(encoded)) {
				b.Fatalf("fixture produced invalid JSON: %q", encoded)
			}
			b.ReportAllocs()
			for b.Loop() {
				benchmarkMapJSONResult = mapToJSON(tc.value)
			}
		})
	}
}
