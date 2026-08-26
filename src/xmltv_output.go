package src

import (
	"encoding/xml"
	"fmt"
	"io"
)

type xmlTVDocumentWriter struct {
	out        io.Writer
	encoder    *xml.Encoder
	encodedAny bool
	closed     bool
}

func newXMLTVDocumentWriter(out io.Writer, generator, source string) (*xmlTVDocumentWriter, error) {
	if _, err := io.WriteString(out, xml.Header); err != nil {
		return nil, err
	}
	if _, err := io.WriteString(out, "<tv>\n"); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(out, "  <generator>%s</generator>\n", generator); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(out, "  <source>%s</source>\n", source); err != nil {
		return nil, err
	}
	encoder := xml.NewEncoder(out)
	encoder.Indent("  ", "    ")
	return &xmlTVDocumentWriter{out: out, encoder: encoder}, nil
}

func (w *xmlTVDocumentWriter) writeValue(value any) error {
	if err := w.encoder.Encode(value); err != nil {
		return err
	}
	w.encodedAny = true
	return nil
}

func (w *xmlTVDocumentWriter) WriteChannel(value Channel) error {
	return w.writeValue(value)
}

func (w *xmlTVDocumentWriter) WriteProgram(value *Program) error {
	return w.writeValue(value)
}

func (w *xmlTVDocumentWriter) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true
	if err := w.encoder.Close(); err != nil {
		return err
	}
	if w.encodedAny {
		if _, err := io.WriteString(w.out, "\n"); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w.out, "</tv>\n")
	return err
}
