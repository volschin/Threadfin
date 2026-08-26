package m3u

import (
	"bytes"
	"testing"
)

var benchmarkM3UResult []interface{}

const benchmarkM3UHeader = "#EXTM3U\n"

const benchmarkM3UFixedRecord = `#EXTINF:-1 tvg-id="channel" tvg-name="Benchmark Channel" tvg-logo="https://example.test/logo.png" group-title="Benchmark",Benchmark Channel
https://example.test/live/channel.ts
`

const benchmarkM3UTailPrefix = `#EXTINF:-1 tvg-id="tail" tvg-name="`
const benchmarkM3UTailSuffix = `" group-title="Benchmark",Tail Channel
https://example.test/live/tail.ts
`

func benchmarkM3UBySize(size int) []byte {
	minimum := len(benchmarkM3UHeader) + len(benchmarkM3UTailPrefix) + len(benchmarkM3UTailSuffix)
	if size < minimum {
		panic("benchmark M3U size is too small for one complete record")
	}

	fixedRecords := (size - minimum) / len(benchmarkM3UFixedRecord)
	padding := size - len(benchmarkM3UHeader) -
		fixedRecords*len(benchmarkM3UFixedRecord) -
		len(benchmarkM3UTailPrefix) - len(benchmarkM3UTailSuffix)

	var playlist bytes.Buffer
	playlist.Grow(size)
	playlist.WriteString(benchmarkM3UHeader)
	for range fixedRecords {
		playlist.WriteString(benchmarkM3UFixedRecord)
	}
	playlist.WriteString(benchmarkM3UTailPrefix)
	playlist.Write(bytes.Repeat([]byte{'x'}, padding))
	playlist.WriteString(benchmarkM3UTailSuffix)
	return playlist.Bytes()
}

func BenchmarkM3UParsingNormal(b *testing.B) {
	cases := []struct {
		name string
		size int
	}{
		{name: "1MiB", size: 1 << 20},
		{name: "5MiB", size: 5 << 20},
		{name: "10MiB", size: 10 << 20},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			input := benchmarkM3UBySize(tc.size)
			if len(input) != tc.size {
				b.Fatalf("fixture size = %d, want %d", len(input), tc.size)
			}
			streams, err := MakeInterfaceFromM3U(input)
			if err != nil {
				b.Fatalf("fixture validation: %v", err)
			}
			if len(streams) == 0 {
				b.Fatal("fixture validation parsed zero streams")
			}

			b.ReportAllocs()
			for b.Loop() {
				streams, err := MakeInterfaceFromM3U(input)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkM3UResult = streams
			}
		})
	}
}
