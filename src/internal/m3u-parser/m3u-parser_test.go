package m3u

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
)

type M3UStream struct {
	GroupTitle string `json:"group-title"`
	Name       string `json:"name"`
	TvgID      string `json:"tvg-id"`
	TvgLogo    string `json:"tvg-logo"`
	TvgName    string `json:"tvg-name"`
	URL        string `json:"url"`
	UUIDKey    string `json:"_uuid.key,omitempty"`
	UUIDValue  string `json:"_uuid.value,omitempty"`
}

func TestStream1(t *testing.T) {

	var file = "test_list_1.m3u"
	var content, err = os.ReadFile(file)
	if err != nil {
		t.Error(err)
		return
	}

	streams, err := MakeInterfaceFromM3U(content)

	if err != nil {
		t.Error(err)
	}

	err = checkStream(streams)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Streams:", len(streams))
	t.Log(streams)

}

func checkStream(streamInterface []interface{}) (err error) {

	for i, s := range streamInterface {

		var stream = s.(map[string]string)
		var m3uStream M3UStream

		jsonString, err := json.MarshalIndent(stream, "", "  ")

		if err == nil {

			err = json.Unmarshal(jsonString, &m3uStream)
			if err == nil {

				log.Print(fmt.Sprintf("Stream:        %d", i))
				log.Print(fmt.Sprintf("Name*:         %s", m3uStream.Name))
				log.Print(fmt.Sprintf("URL*:          %s", m3uStream.URL))
				log.Print(fmt.Sprintf("tvg-name:      %s", m3uStream.TvgName))
				log.Print(fmt.Sprintf("tvg-id**:      %s", m3uStream.TvgID))
				log.Print(fmt.Sprintf("tvg-logo:      %s", m3uStream.TvgLogo))
				log.Print(fmt.Sprintf("group-title**: %s", m3uStream.GroupTitle))

				if len(m3uStream.UUIDKey) > 0 {
					log.Print(fmt.Sprintf("UUID key***:   %s", m3uStream.UUIDKey))
					log.Print(fmt.Sprintf("UUID value:    %s", m3uStream.UUIDValue))
				} else {
					log.Print(fmt.Sprintf("UUID key:    false"))
				}

			}

		}

		log.Println(fmt.Sprintf("- - - - - (*: Required) | (**: Nice to have) | (***: Love it) - - - - -"))
	}

	return
}

func TestMakeInterfaceFromM3URemovesCommentsAndBlankLines(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantURL string
	}{
		{
			name: "adjacent comments and blank lines",
			content: "#EXTM3U\n" +
				"#EXTINF:-1 tvg-id=\"one\" tvg-name=\"One\",One\n" +
				"# first comment\n" +
				"# second comment\n" +
				"\n" +
				"http://example.com/one\n",
			wantURL: "http://example.com/one",
		},
		{
			name: "no removable lines",
			content: "#EXTM3U\n" +
				"#EXTINF:-1 tvg-id=\"two\" tvg-name=\"Two\",Two\n" +
				"http://example.com/two",
			wantURL: "http://example.com/two",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			streams, err := MakeInterfaceFromM3U([]byte(tt.content))
			if err != nil {
				t.Fatalf("MakeInterfaceFromM3U() error = %v", err)
			}
			if len(streams) != 1 {
				t.Fatalf("MakeInterfaceFromM3U() returned %d streams, want 1", len(streams))
			}
			stream, ok := streams[0].(map[string]string)
			if !ok {
				t.Fatalf("stream type = %T, want map[string]string", streams[0])
			}
			if got := stream["url"]; got != tt.wantURL {
				t.Fatalf("stream URL = %q, want %q", got, tt.wantURL)
			}
		})
	}
}
