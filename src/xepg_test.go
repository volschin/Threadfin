package src

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateXMLTVFileProducesMatchingPlainAndGzip(t *testing.T) {
	restorePersistentState(t)
	root := t.TempDir()
	System.Folder.ImagesCache = root + string(os.PathSeparator)
	System.File.XML = filepath.Join(root, "threadfin.xml")
	System.Compressed.GZxml = filepath.Join(root, "threadfin.xml.gz")
	System.Name, System.Branch, System.Version = "Threadfin", "main", "3.0.0"
	Data.XMLTV.Files = []string{"fixture"}
	Data.XEPG.Channels = map[string]interface{}{}
	if err := createXMLTVFile(); err != nil {
		t.Fatal(err)
	}
	plain, err := os.ReadFile(System.File.XML)
	if err != nil {
		t.Fatal(err)
	}
	compressed, err := os.Open(System.Compressed.GZxml)
	if err != nil {
		t.Fatal(err)
	}
	zr, gzipErr := gzip.NewReader(compressed)
	if gzipErr != nil {
		_ = compressed.Close()
		t.Fatal(gzipErr)
	}
	unzipped, readErr := io.ReadAll(zr)
	if err = errors.Join(readErr, zr.Close(), compressed.Close()); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(unzipped, plain) {
		t.Fatalf("gzip differs from plain XML:\nplain=%q\ngzip=%q", plain, unzipped)
	}
}

var errXMLFlush = errors.New("XML flush failed")
var errXMLClose = errors.New("XML close failed")

type xmlFlushErrorWriter struct{}

func (xmlFlushErrorWriter) Write([]byte) (int, error) { return 0, errXMLFlush }

type xmlCloseError struct{}

func (xmlCloseError) Close() error { return errXMLClose }

func TestFinalizeXMLTVOutputJoinsFlushAndCloseErrors(t *testing.T) {
	writer := bufio.NewWriter(xmlFlushErrorWriter{})
	if _, err := writer.WriteString("buffered XML"); err != nil {
		t.Fatal(err)
	}
	err := finalizeXMLTVOutput(writer, xmlCloseError{})
	if !errors.Is(err, errXMLFlush) || !errors.Is(err, errXMLClose) {
		t.Fatalf("error = %v, want both failures", err)
	}
}
