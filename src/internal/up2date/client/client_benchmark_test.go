package up2date

import (
	"bytes"
	"testing"
)

var benchmarkServerResponseResult ServerResponse

func benchmarkUpdaterJSON(prefix, suffix string, size int) []byte {
	padding := size - len(prefix) - len(suffix)
	if padding < 0 {
		panic("updater JSON fixture size is too small")
	}
	result := make([]byte, 0, size)
	result = append(result, prefix...)
	result = append(result, bytes.Repeat([]byte{'u'}, padding)...)
	result = append(result, suffix...)
	return result
}

func BenchmarkUpdateJSONDecoding(b *testing.B) {
	cases := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "Small",
			payload: []byte(`{"status":true,"version":"3.1.0","update.url.bin":"https://example.test/threadfin"}`),
		},
		{
			name: "Large_1MiB",
			payload: benchmarkUpdaterJSON(
				`{"status":true,"version":"3.1.0","update.url.bin":"https://example.test/threadfin","padding":"`,
				`"}`,
				1<<20,
			),
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			reader := bytes.NewReader(tc.payload)
			b.ReportAllocs()
			for b.Loop() {
				reader.Reset(tc.payload)
				response, err := decodeServerResponse(reader)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkServerResponseResult = response
			}
		})
	}
}
