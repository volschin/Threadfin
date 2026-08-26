package src

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type publicationTestFile struct {
	bytes.Buffer
	name     string
	writeErr error
	chmodErr error
	syncErr  error
	closeErr error
	mode     os.FileMode
}

func (f *publicationTestFile) Name() string { return f.name }
func (f *publicationTestFile) Write(p []byte) (int, error) {
	if f.writeErr != nil {
		return 0, f.writeErr
	}
	return f.Buffer.Write(p)
}
func (f *publicationTestFile) WriteString(s string) (int, error) { return f.Write([]byte(s)) }
func (f *publicationTestFile) Chmod(mode os.FileMode) error {
	f.mode = mode
	return f.chmodErr
}
func (f *publicationTestFile) Sync() error  { return f.syncErr }
func (f *publicationTestFile) Close() error { return f.closeErr }

type publicationTestFileInfo struct {
	mode os.FileMode
}

func (i publicationTestFileInfo) Name() string       { return "threadfin.m3u" }
func (i publicationTestFileInfo) Size() int64        { return 0 }
func (i publicationTestFileInfo) Mode() os.FileMode  { return i.mode }
func (i publicationTestFileInfo) ModTime() time.Time { return time.Time{} }
func (i publicationTestFileInfo) IsDir() bool        { return false }
func (i publicationTestFileInfo) Sys() interface{}   { return nil }

func TestPublishM3UFilePublishesCompleteFile(t *testing.T) {
	directory := t.TempDir()
	filename := filepath.Join(directory, "threadfin.m3u")

	err := publishM3UFile(filename, func(writer io.StringWriter) error {
		_, err := writer.WriteString("#EXTM3U\nnew playlist\n")
		return err
	})
	if err != nil {
		t.Fatalf("publishM3UFile() error = %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(content), "#EXTM3U\nnew playlist\n"; got != want {
		t.Fatalf("published content = %q, want %q", got, want)
	}
	if os.PathSeparator == '/' {
		info, err := os.Stat(filename)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := info.Mode().Perm(), os.FileMode(0o600); got != want {
			t.Fatalf("published mode = %04o, want %04o", got, want)
		}
	}
	temporary, err := filepath.Glob(filepath.Join(directory, ".threadfin.m3u-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(temporary) != 0 {
		t.Fatalf("temporary files remain after publication: %v", temporary)
	}
}

func TestPublishM3UFilePreservesCompatibleMode(t *testing.T) {
	tests := []struct {
		name         string
		existingMode os.FileMode
		statErr      error
		wantMode     os.FileMode
	}{
		{name: "preserves existing published mode", existingMode: 0o640, wantMode: 0o640},
		{name: "uses owner-only mode for missing file", statErr: fs.ErrNotExist, wantMode: 0o600},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			temporary := &publicationTestFile{name: ".threadfin.m3u-test"}
			err := publishM3UFileWithOps("threadfin.m3u", func(writer io.StringWriter) error {
				_, err := writer.WriteString("#EXTM3U\nreplacement\n")
				return err
			}, m3uPublicationOps{
				createTemp: func(string, string) (m3uTempFile, error) { return temporary, nil },
				stat: func(string) (os.FileInfo, error) {
					if test.statErr != nil {
						return nil, test.statErr
					}
					return publicationTestFileInfo{mode: test.existingMode}, nil
				},
				rename: func(string, string) error { return nil },
				remove: func(string) error { return nil },
			})
			if err != nil {
				t.Fatalf("publishM3UFileWithOps() error = %v", err)
			}
			if got := temporary.mode; got != test.wantMode {
				t.Fatalf("temporary mode = %04o, want %04o", got, test.wantMode)
			}
		})
	}
}

func TestPublishM3UFilePreservesPublishedFileOnFailure(t *testing.T) {
	writerErr := errors.New("writer failure")
	chmodErr := errors.New("chmod failure")
	flushErr := errors.New("flush failure")
	syncErr := errors.New("sync failure")
	closeErr := errors.New("close failure")
	renameErr := errors.New("rename failure")

	tests := []struct {
		name   string
		file   func(string) *publicationTestFile
		write  func(io.StringWriter) error
		rename func(string, string) error
		want   error
	}{
		{
			name: "chmod",
			file: func(name string) *publicationTestFile {
				return &publicationTestFile{name: name, chmodErr: chmodErr}
			},
			write: func(io.StringWriter) error {
				return errors.New("write called after chmod failure")
			},
			want: chmodErr,
		},
		{
			name: "writer",
			file: func(name string) *publicationTestFile {
				return &publicationTestFile{name: name, writeErr: writerErr}
			},
			write: func(writer io.StringWriter) error {
				_, err := writer.WriteString(strings.Repeat("x", (1<<20)+1))
				return err
			},
			want: writerErr,
		},
		{
			name: "flush",
			file: func(name string) *publicationTestFile {
				return &publicationTestFile{name: name, writeErr: flushErr}
			},
			write: func(writer io.StringWriter) error {
				_, err := writer.WriteString("small write")
				return err
			},
			want: flushErr,
		},
		{
			name: "sync",
			file: func(name string) *publicationTestFile {
				return &publicationTestFile{name: name, syncErr: syncErr}
			},
			write: func(writer io.StringWriter) error {
				_, err := writer.WriteString("small write")
				return err
			},
			want: syncErr,
		},
		{
			name: "close",
			file: func(name string) *publicationTestFile {
				return &publicationTestFile{name: name, closeErr: closeErr}
			},
			write: func(writer io.StringWriter) error {
				_, err := writer.WriteString("small write")
				return err
			},
			want: closeErr,
		},
		{
			name: "rename",
			file: func(name string) *publicationTestFile {
				return &publicationTestFile{name: name}
			},
			write: func(writer io.StringWriter) error {
				_, err := writer.WriteString("small write")
				return err
			},
			rename: func(string, string) error { return renameErr },
			want:   renameErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			filename := filepath.Join(directory, "threadfin.m3u")
			if err := os.WriteFile(filename, []byte("published playlist"), 0o600); err != nil {
				t.Fatal(err)
			}
			temporaryName := filepath.Join(directory, ".threadfin.m3u-test")
			if err := os.WriteFile(temporaryName, nil, 0o600); err != nil {
				t.Fatal(err)
			}

			ops := m3uPublicationOps{
				createTemp: func(string, string) (m3uTempFile, error) {
					return test.file(temporaryName), nil
				},
				rename: os.Rename,
				remove: os.Remove,
			}
			if test.rename != nil {
				ops.rename = test.rename
			}

			err := publishM3UFileWithOps(filename, test.write, ops)
			if !errors.Is(err, test.want) {
				t.Fatalf("publishM3UFileWithOps() error = %v, want wrapping %v", err, test.want)
			}
			content, err := os.ReadFile(filename)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := string(content), "published playlist"; got != want {
				t.Fatalf("published content = %q, want %q", got, want)
			}
			if _, err := os.Stat(temporaryName); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("temporary file remains after failure: %v", err)
			}
		})
	}
}

func TestPublishM3UFileWithOpsReportsCreateTempError(t *testing.T) {
	createErr := errors.New("create temporary file failure")
	directory := t.TempDir()
	filename := filepath.Join(directory, "threadfin.m3u")
	if err := os.WriteFile(filename, []byte("published playlist"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := publishM3UFileWithOps(filename, func(io.StringWriter) error {
		return nil
	}, m3uPublicationOps{
		createTemp: func(string, string) (m3uTempFile, error) {
			return nil, createErr
		},
	})
	if !errors.Is(err, createErr) {
		t.Fatalf("publishM3UFileWithOps() error = %v, want wrapping %v", err, createErr)
	}
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(content), "published playlist"; got != want {
		t.Fatalf("published content = %q, want %q", got, want)
	}
}

func TestPublishM3UFileWithOpsJoinsCleanupError(t *testing.T) {
	writerErr := errors.New("writer failure")
	removeErr := errors.New("remove temporary file failure")
	directory := t.TempDir()
	filename := filepath.Join(directory, "threadfin.m3u")
	temporaryName := filepath.Join(directory, ".threadfin.m3u-test")
	if err := os.WriteFile(filename, []byte("published playlist"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(temporaryName, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	err := publishM3UFileWithOps(filename, func(writer io.StringWriter) error {
		_, err := writer.WriteString(strings.Repeat("x", (1<<20)+1))
		return err
	}, m3uPublicationOps{
		createTemp: func(string, string) (m3uTempFile, error) {
			return &publicationTestFile{name: temporaryName, writeErr: writerErr}, nil
		},
		rename: os.Rename,
		remove: func(name string) error {
			if err := os.Remove(name); err != nil {
				return err
			}
			return removeErr
		},
	})
	if !errors.Is(err, writerErr) {
		t.Fatalf("publishM3UFileWithOps() error = %v, want wrapping %v", err, writerErr)
	}
	if !errors.Is(err, removeErr) {
		t.Fatalf("publishM3UFileWithOps() error = %v, want joining %v", err, removeErr)
	}
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(content), "published playlist"; got != want {
		t.Fatalf("published content = %q, want %q", got, want)
	}
	if _, err := os.Stat(temporaryName); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("temporary file remains after cleanup failure: %v", err)
	}
}

func TestBuildM3UFiltersProduceExpectedPlaylist(t *testing.T) {
	restorePersistentState(t)
	Settings = SettingsStruct{}
	Data = DataStruct{}
	System.ServerProtocol.XML = "http"
	System.Domain = "threadfin.example"
	Data.XEPG.Channels = map[string]interface{}{
		"channel": XEPGChannelStruct{
			FileM3UID:   "playlist",
			URL:         "https://provider.example/channel",
			XActive:     true,
			XChannelID:  "1",
			XEPG:        "channel-id",
			XGroupTitle: "News",
			XName:       "Channel",
		},
	}

	previousCreateM3UStreamingURL := createM3UStreamingURL
	createM3UStreamingURL = func(string, string, string, string, string, *BackupStream, *BackupStream, *BackupStream) (string, error) {
		return "http://stream.example/channel", nil
	}
	t.Cleanup(func() {
		createM3UStreamingURL = previousCreateM3UStreamingURL
	})

	got, err := buildM3U([]string{"News"})
	if err != nil {
		t.Fatalf("buildM3U() error = %v", err)
	}
	want := "#EXTM3U url-tvg=\"http://threadfin.example/xmltv/threadfin.xml\" x-tvg-url=\"http://threadfin.example/xmltv/threadfin.xml\"\n" +
		"#EXTINF:0 channelID=\"channel-id\" tvg-chno=\"1\" tvg-name=\"Channel\" tvg-id=\"1\" tvg-logo=\"\" group-title=\"News\",Channel\n" +
		"http://stream.example/channel\n"
	if got != want {
		t.Fatalf("buildM3U() = %q, want %q", got, want)
	}
}

func TestBuildM3UReturnsStreamingURLError(t *testing.T) {
	streamErr := errors.New("stream URL failure")
	restorePersistentState(t)
	Data = DataStruct{}
	Data.XEPG.Channels = map[string]interface{}{
		"channel": XEPGChannelStruct{
			FileM3UID:   "playlist",
			URL:         "https://provider.example/channel",
			XActive:     true,
			XChannelID:  "1",
			XGroupTitle: "Group",
			XName:       "Channel",
		},
	}

	previousCreateM3UStreamingURL := createM3UStreamingURL
	createM3UStreamingURL = func(string, string, string, string, string, *BackupStream, *BackupStream, *BackupStream) (string, error) {
		return "", streamErr
	}
	t.Cleanup(func() {
		createM3UStreamingURL = previousCreateM3UStreamingURL
	})

	_, err := buildM3U([]string{"Group"})
	if !errors.Is(err, streamErr) {
		t.Fatalf("buildM3U() error = %v, want wrapping %v", err, streamErr)
	}
}
