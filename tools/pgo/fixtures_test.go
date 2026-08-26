//go:build linux && amd64

package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestFixtureCardinalityAndDeterminism(t *testing.T) {
	start := time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC)
	var firstPlaylist, secondPlaylist bytes.Buffer
	if err := writePlaylist(&firstPlaylist, "http://127.0.0.1:12345"); err != nil {
		t.Fatal(err)
	}
	if err := writePlaylist(&secondPlaylist, "http://127.0.0.1:12345"); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstPlaylist.Bytes(), secondPlaylist.Bytes()) {
		t.Fatal("playlist is not deterministic")
	}
	if got := bytes.Count(firstPlaylist.Bytes(), []byte("#EXTINF:")); got != playlistEntryCount {
		t.Fatalf("playlist entries = %d", got)
	}
	for i := 0; i < streamChannelCount; i++ {
		want := fmt.Sprintf("/stream/%d.ts?channel=%d", i, i)
		if !bytes.Contains(firstPlaylist.Bytes(), []byte(want)) {
			t.Fatalf("missing %s", want)
		}
	}

	var firstGuide, secondGuide bytes.Buffer
	if err := writeGuide(&firstGuide, start); err != nil {
		t.Fatal(err)
	}
	if err := writeGuide(&secondGuide, start); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstGuide.Bytes(), secondGuide.Bytes()) {
		t.Fatal("guide is not deterministic")
	}
	var guide struct {
		Channels []struct {
			ID string `xml:"id,attr"`
		} `xml:"channel"`
		Programmes []struct {
			Channel string `xml:"channel,attr"`
		} `xml:"programme"`
	}
	if err := xml.Unmarshal(firstGuide.Bytes(), &guide); err != nil {
		t.Fatal(err)
	}
	if len(guide.Channels) != xmltvChannelCount || len(guide.Programmes) != xmltvProgramCount {
		t.Fatalf("guide cardinality = %d/%d", len(guide.Channels), len(guide.Programmes))
	}
}

func TestPlaylistProvenanceIgnoresEphemeralServerEndpoint(t *testing.T) {
	firstServer := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(firstServer.Close)
	secondServer := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(secondServer.Close)

	newFixture := func(server *httptest.Server) *fixtureSet {
		t.Helper()
		var playlist bytes.Buffer
		if err := writePlaylist(&playlist, server.URL); err != nil {
			t.Fatal(err)
		}
		return &fixtureSet{server: server, playlist: playlist.Bytes()}
	}
	first, second := newFixture(firstServer), newFixture(secondServer)

	if bytes.Equal(first.playlistBytes(), second.playlistBytes()) {
		t.Fatal("operational playlists unexpectedly use the same endpoint")
	}
	if !bytes.Contains(first.playlistBytes(), []byte(firstServer.URL+"/stream/0.ts?channel=0")) ||
		!bytes.Contains(second.playlistBytes(), []byte(secondServer.URL+"/stream/0.ts?channel=0")) {
		t.Fatal("operational playlist does not retain its fixture server endpoint")
	}
	if got, want := bytesSHA256(first.playlistProvenanceBytes()), bytesSHA256(second.playlistProvenanceBytes()); got != want {
		t.Fatalf("playlist provenance hash differs across ephemeral endpoints: %s != %s", got, want)
	}
}

func TestSettingsDocumentUsesCurrentFields(t *testing.T) {
	body, err := settingsDocument("http://127.0.0.1:12345", "/usr/bin/ffmpeg", "34401")
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	checks := map[string]any{
		"version": "0.5.0", "files.update": true, "epgSource": "XEPG",
		"buffer": "ffmpeg", "buffer.size.kb": float64(1024), "buffer.timeout": float64(0),
		"cache.images": false, "ssdp": false, "ThreadfinAutoUpdate": false,
		"port": "34401", "bindIpAddress": "127.0.0.1", "tuner": float64(4),
	}
	for key, want := range checks {
		if !reflect.DeepEqual(got[key], want) {
			t.Fatalf("%s = %#v, want %#v", key, got[key], want)
		}
	}
	files := got["files"].(map[string]any)
	m3u := files["m3u"].(map[string]any)["M1000"].(map[string]any)
	xmltv := files["xmltv"].(map[string]any)["X1000"].(map[string]any)
	if m3u["file.source"] != "http://127.0.0.1:12345/playlist.m3u" || m3u["buffer"] != "ffmpeg" || m3u["tuner"] != float64(4) {
		t.Fatalf("invalid M3U provider: %#v", m3u)
	}
	if xmltv["file.source"] != "http://127.0.0.1:12345/guide.xml" {
		t.Fatalf("invalid XMLTV provider: %#v", xmltv)
	}
}
