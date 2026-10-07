package src

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"threadfin/src/internal/imgcache"
)

func setupPublishedXMLTV(t *testing.T) (plain, compressed []byte) {
	t.Helper()
	restorePersistentState(t)
	root := t.TempDir()
	System.Folder.ImagesCache = root + string(os.PathSeparator)
	System.Folder.Data = root + string(os.PathSeparator)
	System.File.XML = filepath.Join(root, "threadfin.xml")
	System.Compressed.GZxml = filepath.Join(root, "threadfin.xml.gz")
	System.Name, System.Branch, System.Version = "Threadfin", "main", "test"
	Data.XMLTV.Files = []string{"fixture"}
	Data.XEPG.Channels = map[string]interface{}{}
	var err error
	Data.Cache.Images, err = imgcache.New(root, "/images/", false)
	if err != nil {
		t.Fatal(err)
	}
	plain = []byte("<?xml version=\"1.0\"?><tv><channel id=\"old\"/></tv>")
	var buffer bytes.Buffer
	w := gzip.NewWriter(&buffer)
	if _, err := w.Write(plain); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	compressed = buffer.Bytes()
	for name, content := range map[string][]byte{System.File.XML: plain, System.Compressed.GZxml: compressed} {
		if err := os.WriteFile(name, content, 0o640); err != nil {
			t.Fatal(err)
		}
	}
	return plain, compressed
}

func assertPublishedBytes(t *testing.T, filename string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(filename)
	if err != nil || !bytes.Equal(got, want) {
		t.Errorf("published %s = %q, error %v; want previous contents %q", filepath.Base(filename), got, err, want)
	}
}

func assertNoXMLTVTemporaryFiles(t *testing.T) {
	t.Helper()
	files, err := os.ReadDir(filepath.Dir(System.File.XML))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if file.Name() != "threadfin.xml" && file.Name() != "threadfin.xml.gz" {
			t.Errorf("temporary artifact remains: %s", file.Name())
		}
	}
}

func TestXMLTVGenerationFailurePreservesPublishedPair(t *testing.T) {
	plain, compressed := setupPublishedXMLTV(t)
	Data.XEPG.Channels["broken"] = XEPGChannelStruct{XActive: true, XName: "Broken", XChannelID: "broken", XmltvFile: "missing-guide.xml", XMapping: "-"}
	if err := createXMLTVFile(); err == nil {
		t.Error("expected a source generation error")
	}
	assertPublishedBytes(t, System.File.XML, plain)
	assertPublishedBytes(t, System.Compressed.GZxml, compressed)
	assertNoXMLTVTemporaryFiles(t)
}

func TestXMLTVUnmappedChannelsRemainPublishable(t *testing.T) {
	for _, source := range []string{"", "-"} {
		t.Run("source="+source, func(t *testing.T) {
			setupPublishedXMLTV(t)
			Data.XEPG.Channels["unmapped"] = XEPGChannelStruct{
				XActive: true, XName: "Unmapped", XChannelID: "unmapped",
				XmltvFile: source, XMapping: "-",
			}
			if err := createXMLTVFile(); err != nil {
				t.Fatalf("publish active unmapped channel: %v", err)
			}
			plain, err := os.ReadFile(System.File.XML)
			if err != nil {
				t.Fatal(err)
			}
			var guide XMLTV
			if err := xml.Unmarshal(plain, &guide); err != nil {
				t.Fatal(err)
			}
			if len(guide.Channel) != 1 || guide.Channel[0].ID != "unmapped" || len(guide.Program) != 0 {
				t.Fatalf("unmapped output has %d channels and %d programs", len(guide.Channel), len(guide.Program))
			}
			assertNoXMLTVTemporaryFiles(t)
		})
	}
}

func TestXMLTVCompressionFailurePreservesPublishedPlain(t *testing.T) {
	plain, compressed := setupPublishedXMLTV(t)
	previousGzip := System.Compressed.GZxml
	System.Compressed.GZxml = filepath.Join(previousGzip, "not-a-directory.gz")
	if err := createXMLTVFile(); err == nil {
		t.Fatal("expected compression target error")
	}
	assertPublishedBytes(t, System.File.XML, plain)
	assertPublishedBytes(t, previousGzip, compressed)
	assertNoXMLTVTemporaryFiles(t)
}

func TestXMLTVReadersSeePreviousGuideUntilGenerationCompletes(t *testing.T) {
	plain, compressed := setupPublishedXMLTV(t)
	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	Data.Cache.Images.Image.GetURL = func(src, domain, port string, force bool, httpsPort int, httpsDomain string) string {
		close(started)
		<-release
		return src
	}
	Data.XEPG.Channels["new"] = XEPGChannelStruct{XActive: true, XName: "New", XChannelID: "new"}
	done := make(chan error, 1)
	go func() { done <- createXMLTVFile() }()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		unblock()
		<-done
		t.Fatal("generation did not start")
	}
	assertPublishedBytes(t, System.File.XML, plain)
	assertPublishedBytes(t, System.Compressed.GZxml, compressed)
	unblock()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	newPlain, err := os.ReadFile(System.File.XML)
	if err != nil {
		t.Fatal(err)
	}
	var guide XMLTV
	if err := xml.Unmarshal(newPlain, &guide); err != nil || len(guide.Channel) != 1 || guide.Channel[0].ID != "new" {
		t.Fatalf("new guide = %s, error %v", newPlain, err)
	}
	newGzip, err := os.ReadFile(System.Compressed.GZxml)
	if err != nil {
		t.Fatal(err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(newGzip))
	if err != nil {
		t.Fatal(err)
	}
	unzipped, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	if err := zr.Close(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(unzipped, newPlain) {
		t.Fatal("new plain and compressed guides differ")
	}
	if os.PathSeparator == '/' {
		for _, filename := range []string{System.File.XML, System.Compressed.GZxml} {
			info, err := os.Stat(filename)
			if err != nil || info.Mode().Perm() != 0o640 {
				t.Fatalf("published permissions changed, stat error %v", err)
			}
		}
	}
	assertNoXMLTVTemporaryFiles(t)
}

func TestXMLTVRenderingFailureDiscardsPartiallyWrittenOutputs(t *testing.T) {
	plain, compressed := setupPublishedXMLTV(t)
	renderErr := errors.New("XMLTV renderer failed after writing")
	err := publishXMLTVFiles(System.File.XML, System.Compressed.GZxml, func(w io.Writer) error {
		if _, err := io.WriteString(w, strings.Repeat("partially rendered XML", 100000)); err != nil {
			return err
		}
		return renderErr
	})
	if !errors.Is(err, renderErr) {
		t.Fatalf("publication error = %v, want %v", err, renderErr)
	}
	assertPublishedBytes(t, System.File.XML, plain)
	assertPublishedBytes(t, System.Compressed.GZxml, compressed)
	assertNoXMLTVTemporaryFiles(t)
}

func TestXMLTVConcurrentPublishersSerializeGenerations(t *testing.T) {
	directory := t.TempDir()
	plainPath := filepath.Join(directory, "threadfin.xml")
	gzipPath := plainPath + ".gz"
	firstXML := `<tv><channel id="first"/></tv>`
	secondXML := `<tv><channel id="second"/></tv>`
	firstRendering := make(chan struct{})
	finishFirst := make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(finishFirst) }) }
	defer release()
	firstDone := make(chan error, 1)
	go func() {
		firstDone <- publishXMLTVFiles(plainPath, gzipPath, func(w io.Writer) error {
			close(firstRendering)
			<-finishFirst
			_, err := io.WriteString(w, firstXML)
			return err
		})
	}()
	select {
	case <-firstRendering:
	case err := <-firstDone:
		t.Fatalf("first publisher failed before rendering: %v", err)
	}
	secondAttempted := make(chan struct{})
	secondRendering := make(chan struct{})
	secondDone := make(chan error, 1)
	go func() {
		close(secondAttempted)
		secondDone <- publishXMLTVFiles(plainPath, gzipPath, func(w io.Writer) error {
			close(secondRendering)
			_, err := io.WriteString(w, secondXML)
			return err
		})
	}()
	<-secondAttempted
	select {
	case <-secondRendering:
		t.Error("a second generation began rendering while the first still owned publication")
	case <-time.After(100 * time.Millisecond):
	}
	release()
	for _, done := range []<-chan error{firstDone, secondDone} {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	assertPublishedBytes(t, plainPath, []byte(secondXML))
	compressed, err := os.ReadFile(gzipPath)
	if err != nil {
		t.Fatal(err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatal(err)
	}
	decoded, readErr := io.ReadAll(zr)
	if err := errors.Join(readErr, zr.Close()); err != nil {
		t.Fatal(err)
	}
	if string(decoded) != secondXML {
		t.Fatalf("compressed generation = %q, want %q", decoded, secondXML)
	}
}
