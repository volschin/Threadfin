package m3u

import (
	"crypto/md5"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestM3UParsingDoesNotChangeAtSizeThreshold(t *testing.T) {
	records := "#EXTINF:-1 tvg-id=\"one\",One\nhttp://example.test/one\n" +
		"#EXTINF:-1 tvg-id=\"two\",Two\nhttp://example.test/two\n"
	for _, size := range []int{10 * 1024 * 1024, 10*1024*1024 + 1} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			padding := size - len("#EXTM3U\n") - len(records)
			input := "#EXTM3U\n" + strings.Repeat("#\n", padding/2) + strings.Repeat("\n", padding%2) + records
			streams, err := MakeInterfaceFromM3U([]byte(input))
			if err != nil {
				t.Fatal(err)
			}
			if len(streams) != 2 {
				t.Fatalf("parsed %d streams, want 2", len(streams))
			}
			for i, name := range []string{"One", "Two"} {
				if got := streams[i].(map[string]string)["name"]; got != name {
					t.Fatalf("stream %d name = %q, want %q", i, got, name)
				}
			}
		})
	}
}

func TestM3UEntryPointsShareMetadataSemantics(t *testing.T) {
	input := []byte("#EXTM3U\r\nstray preamble\r\n" +
		"#EXTINF:-1 tvg-id='one' tvg-name='Station, HD' group-title='News, Local',Station, HD\r\n" +
		"#EXTVLCOPT:http-user-agent=player\r\n\r\nhttp://example.test/one\r\n" +
		"#EXTINF:-1 tvg-name=\"Fallback\",\nhttp://example.test/two\n" +
		"#EXTINF:-1,\nhttp://example.test/nameless\n" +
		"#EXTINF:-1,Missing URL\n")
	want := []interface{}{
		map[string]string{"url": "http://example.test/one", "tvg-id": "one", "tvg-name": "Station, HD", "group-title": "News, Local", "name": "Station, HD", "_values": "one Station, HD News, Local Station, HD", "_uuid.key": "tvg-name", "_uuid.value": "Station, HD"},
		map[string]string{"url": "http://example.test/two", "tvg-id": fmt.Sprintf("threadfin-%x", md5.Sum([]byte("http://example.test/two"))), "tvg-name": "Fallback", "name": "Fallback", "_values": "Fallback Fallback", "_uuid.key": "tvg-name", "_uuid.value": "Fallback"},
	}
	for name, parse := range map[string]func([]byte) ([]interface{}, error){
		"public":    MakeInterfaceFromM3U,
		"original":  makeInterfaceFromM3UOriginal,
		"optimized": MakeInterfaceFromM3UOptimized,
	} {
		t.Run(name, func(t *testing.T) {
			got, err := parse(input)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("streams = %#v, want %#v", got, want)
			}
		})
	}
}

func TestM3UOptimizedParsesLongMetadataLine(t *testing.T) {
	name := strings.Repeat("x", 1024*1024+1)
	got, err := MakeInterfaceFromM3UOptimized([]byte("#EXTM3U\n#EXTINF:-1," + name + "\nhttp://example.test/long\n"))
	if err != nil || len(got) != 1 {
		t.Fatalf("parsed %d streams, error %v; want 1", len(got), err)
	}
	if got[0].(map[string]string)["name"] != name {
		t.Fatal("long channel name changed")
	}
}
