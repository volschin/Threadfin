//go:build windows

package up2date

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestWindowsSubprocessPreservesArgumentsByteForByte(t *testing.T) {
	const childMarker = "threadfin-update-argument-child"
	for index, arg := range os.Args {
		if arg != childMarker {
			continue
		}
		if index+1 >= len(os.Args) {
			t.Fatal("missing subprocess output path")
		}
		encoded, err := json.Marshal(os.Args[index+2:])
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(os.Args[index+1], encoded, 0600); err != nil {
			t.Fatal(err)
		}
		return
	}

	output := filepath.Join(t.TempDir(), "args.json")
	original := []string{`C:\path with spaces\`, `--quoted="a b"`, `trailing\\`, "plain"}
	args := append([]string{"-test.run=^TestWindowsSubprocessPreservesArgumentsByteForByte$", childMarker, output}, original...)
	executable, err := filepath.Abs(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	process, err := startOSUpdateProcess(filepath.Clean(executable), args)
	if err != nil {
		t.Fatal(err)
	}
	if err := process.Wait(10 * time.Second); err != nil {
		t.Fatal(err)
	}
	encoded, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, original) {
		t.Fatalf("subprocess args = %#v, want %#v", got, original)
	}
}
