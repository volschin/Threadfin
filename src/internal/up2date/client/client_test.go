package up2date

import (
	"net"
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
