package up2date

import (
	"errors"
	"net"
	"strings"
	"testing"
)

func TestServerRequestReturnsTransportErrorWithoutPanicking(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	accepted := make(chan error, 1)
	go func() {
		for range 2 {
			connection, acceptErr := listener.Accept()
			if acceptErr != nil {
				accepted <- acceptErr
				return
			}
			_ = connection.Close()
		}
		accepted <- nil
	}()

	previousUpdater := Updater
	Updater = ClientInfo{URL: "http://" + listener.Addr().String()}
	t.Cleanup(func() { Updater = previousUpdater })

	if err := serverRequest(); err == nil {
		t.Fatal("serverRequest() error = nil, want transport error")
	}
	if err := <-accepted; err != nil {
		t.Fatalf("accepting test connections: %v", err)
	}
}

type failingServerResponseReader struct {
	err error
}

func (reader failingServerResponseReader) Read([]byte) (int, error) {
	return 0, reader.err
}

func TestDecodeServerResponsePreservesCurrentJSONSemantics(t *testing.T) {
	response, err := decodeServerResponse(strings.NewReader(
		`{"STATUS":true,"version":"3.1.0","update.url.bin":"https://example.test/threadfin","unknown":"ignored"}`,
	))
	if err != nil {
		t.Fatal(err)
	}
	if !response.Status || response.Version != "3.1.0" || response.UpdateBIN != "https://example.test/threadfin" {
		t.Fatalf("decoded response = %#v", response)
	}

	for _, body := range []string{"", `{"status":`, `{"status":true} trailing`} {
		if _, err := decodeServerResponse(strings.NewReader(body)); err == nil {
			t.Fatalf("decodeServerResponse(%q) error = nil", body)
		}
	}

	readErr := errors.New("updater response read failed")
	if _, err := decodeServerResponse(failingServerResponseReader{err: readErr}); !errors.Is(err, readErr) {
		t.Fatalf("decodeServerResponse() error = %v, want %v", err, readErr)
	}
}
