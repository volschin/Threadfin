package m3u

import "testing"

func TestMakeInterfaceFromM3UOriginalPreambleAndDelimiterBehavior(t *testing.T) {
	input := []byte("#EXTM3U\nstray preamble URL\n#EXTINF:-1 tvg-id=\"one\",One\nhttp://example/one\n#EXTINF:-1 tvg-id=\"two\",Two\nhttp://example/two")

	got, err := makeInterfaceFromM3UOriginal(input)
	if err != nil {
		t.Fatalf("makeInterfaceFromM3UOriginal() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("makeInterfaceFromM3UOriginal() returned %d streams, want 2", len(got))
	}

	for i, want := range []struct {
		name string
		url  string
	}{
		{name: "One", url: "http://example/one"},
		{name: "Two", url: "http://example/two"},
	} {
		stream, ok := got[i].(map[string]string)
		if !ok {
			t.Fatalf("stream %d has type %T, want map[string]string", i, got[i])
		}
		if stream["name"] != want.name || stream["url"] != want.url {
			t.Fatalf("stream %d = name %q, URL %q; want name %q, URL %q", i, stream["name"], stream["url"], want.name, want.url)
		}
	}
}
