package src

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

type resettableSegmentReadCloser struct {
	*bytes.Reader
}

func (*resettableSegmentReadCloser) Close() error {
	return nil
}

var benchmarkSegmentContentType string

func BenchmarkSegmentTransfer(b *testing.B) {
	cases := []struct {
		name string
		size int
	}{
		{name: "188KiB", size: 188 << 10},
		{name: "2MiB", size: 2 << 20},
		{name: "10MiB", size: 10 << 20},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			content := bytes.Repeat([]byte{0x47}, tc.size)
			source := &resettableSegmentReadCloser{Reader: bytes.NewReader(content)}
			beforeWrite := func(content []byte) {
				benchmarkSegmentContentType = http.DetectContentType(content)
			}

			b.ReportAllocs()
			for b.Loop() {
				source.Reset(content)
				inputErr, writeErr := transferSegment(io.Discard, source, beforeWrite)
				if inputErr != nil || writeErr != nil {
					b.Fatalf("transferSegment() = (%v, %v)", inputErr, writeErr)
				}
			}
		})
	}
}
