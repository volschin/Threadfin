package src

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"reflect"
	"testing"
)

type xmlTVWireContract struct {
	Generator string    `xml:"generator"`
	Source    string    `xml:"source"`
	Channels  []Channel `xml:"channel"`
	Programs  []Program `xml:"programme"`
}

func xmlTVOutputFixture() (Channel, *Program) {
	return Channel{
			ID: "channel-1", DisplayName: []DisplayName{{Value: "News & Weather"}},
			Icon: Icon{Src: "https://example.invalid/logo?a=1&b=2"}, Live: true, Active: true,
		}, &Program{
			Channel: "channel-1", Start: "20260826080000 +0000", Stop: "20260826090000 +0000",
			Title: []*Title{{Lang: "en", Value: "Morning <News>"}},
			Desc:  []*Desc{{Lang: "en", Value: "Headlines & weather"}},
		}
}

func legacyXMLTVDocument(t *testing.T, channels []Channel, programs []*Program) []byte {
	t.Helper()
	var out bytes.Buffer
	_, _ = io.WriteString(&out, xml.Header)
	_, _ = io.WriteString(&out, "<tv>\n")
	_, _ = io.WriteString(&out, "  <generator>Threadfin</generator>\n")
	_, _ = io.WriteString(&out, "  <source>Threadfin - 3.0.0</source>\n")
	for _, value := range channels {
		encoded, err := xml.MarshalIndent(value, "  ", "    ")
		if err != nil {
			t.Fatal(err)
		}
		out.Write(encoded)
		out.WriteByte('\n')
	}
	for _, value := range programs {
		encoded, err := xml.MarshalIndent(value, "  ", "    ")
		if err != nil {
			t.Fatal(err)
		}
		out.Write(encoded)
		out.WriteByte('\n')
	}
	_, _ = io.WriteString(&out, "</tv>\n")
	return out.Bytes()
}

func TestXMLTVDocumentWriterPreservesBytesAndSemantics(t *testing.T) {
	channel, program := xmlTVOutputFixture()
	var got bytes.Buffer
	document, err := newXMLTVDocumentWriter(&got, "Threadfin", "Threadfin - 3.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if err = document.WriteChannel(channel); err != nil {
		t.Fatal(err)
	}
	if err = document.WriteProgram(program); err != nil {
		t.Fatal(err)
	}
	if err = document.Close(); err != nil {
		t.Fatal(err)
	}
	want := legacyXMLTVDocument(t, []Channel{channel}, []*Program{program})
	if !bytes.Equal(got.Bytes(), want) {
		t.Fatalf("XML changed:\ngot:\n%s\nwant:\n%s", got.Bytes(), want)
	}
	var gotWire, wantWire xmlTVWireContract
	if err = xml.Unmarshal(got.Bytes(), &gotWire); err != nil {
		t.Fatal(err)
	}
	if err = xml.Unmarshal(want, &wantWire); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotWire, wantWire) {
		t.Fatalf("semantics changed: %#v %#v", gotWire, wantWire)
	}
}

func TestXMLTVDocumentWriterPreservesZeroValueFormatting(t *testing.T) {
	var got bytes.Buffer
	document, err := newXMLTVDocumentWriter(&got, "Threadfin", "source")
	if err != nil {
		t.Fatal(err)
	}
	if err = document.Close(); err != nil {
		t.Fatal(err)
	}
	want := xml.Header + "<tv>\n  <generator>Threadfin</generator>\n  <source>source</source>\n</tv>\n"
	if got.String() != want {
		t.Fatalf("got %q want %q", got.String(), want)
	}
}

var errXMLDocumentWrite = errors.New("XML document write failed")

type xmlSuffixErrorWriter struct {
	bytes.Buffer
	suffix []byte
}

func (w *xmlSuffixErrorWriter) Write(p []byte) (int, error) {
	if bytes.Contains(p, w.suffix) {
		return 0, errXMLDocumentWrite
	}
	return w.Buffer.Write(p)
}

func (w *xmlSuffixErrorWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func TestXMLTVDocumentWriterReturnsEncodeAndCloseErrors(t *testing.T) {
	channel, _ := xmlTVOutputFixture()
	encodeWriter := &xmlSuffixErrorWriter{suffix: []byte("<channel")}
	document, err := newXMLTVDocumentWriter(encodeWriter, "Threadfin", "source")
	if err != nil {
		t.Fatal(err)
	}
	if err = document.WriteChannel(channel); !errors.Is(err, errXMLDocumentWrite) {
		t.Fatalf("encode error=%v", err)
	}

	closeWriter := &xmlSuffixErrorWriter{suffix: []byte("</tv>")}
	document, err = newXMLTVDocumentWriter(closeWriter, "Threadfin", "source")
	if err != nil {
		t.Fatal(err)
	}
	if err = document.Close(); !errors.Is(err, errXMLDocumentWrite) {
		t.Fatalf("close error=%v", err)
	}
}
