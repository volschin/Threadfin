package src

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	up2date "threadfin/src/internal/up2date/client"
)

func TestMaintenanceBinaryUpdateExitsZeroAfterAcknowledgedHandoff(t *testing.T) {
	events := []string{}
	runMaintenanceBinaryUpdate(
		func() error {
			events = append(events, "update")
			return fmt.Errorf("maintenance handoff: %w", up2date.ErrWindowsUpdateHandoff)
		},
		func(code int) { events = append(events, fmt.Sprintf("exit:%d", code)) },
		func(error) { events = append(events, "report") },
	)
	if want := []string{"update", "exit:0"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("maintenance events = %#v, want %#v", events, want)
	}
}

func TestMaintenanceBinaryUpdateReportsOrdinaryFailureWithoutExit(t *testing.T) {
	failure := errors.New("ordinary update failure")
	exitCalled := false
	var reported error
	runMaintenanceBinaryUpdate(
		func() error { return failure },
		func(int) { exitCalled = true },
		func(err error) { reported = err },
	)
	if exitCalled {
		t.Fatal("ordinary update failure exited the serving process")
	}
	if !errors.Is(reported, failure) {
		t.Fatalf("reported error = %v, want %v", reported, failure)
	}
}
