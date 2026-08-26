package src

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

type benchmarkResponseWriter struct {
	header http.Header
	status int
}

func (writer *benchmarkResponseWriter) Header() http.Header {
	return writer.header
}

func (writer *benchmarkResponseWriter) Write(content []byte) (int, error) {
	return len(content), nil
}

func (writer *benchmarkResponseWriter) WriteHeader(status int) {
	writer.status = status
}

func BenchmarkCachedImageDelivery(b *testing.B) {
	cases := []struct {
		name string
		size int
	}{
		{name: "32KiB", size: 32 << 10},
		{name: "1MiB", size: 1 << 20},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			root := b.TempDir()
			filename := "cached-image.png"
			content := bytes.Repeat([]byte{0x5a}, tc.size)
			if err := os.WriteFile(filepath.Join(root, filename), content, 0o600); err != nil {
				b.Fatal(err)
			}

			systemMutex.Lock()
			previousCache := System.Folder.ImagesCache
			System.Folder.ImagesCache = root + string(os.PathSeparator)
			systemMutex.Unlock()
			b.Cleanup(func() {
				systemMutex.Lock()
				System.Folder.ImagesCache = previousCache
				systemMutex.Unlock()
			})

			request := httptest.NewRequest(http.MethodGet, "/images/"+filename, nil)
			writer := &benchmarkResponseWriter{header: make(http.Header)}
			b.ReportAllocs()
			for b.Loop() {
				clear(writer.header)
				writer.status = 0
				Images(writer, request)
				if writer.status != http.StatusOK {
					b.Fatalf("status = %d, want %d", writer.status, http.StatusOK)
				}
			}
		})
	}
}
