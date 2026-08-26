//go:build linux && amd64

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

type stringList []string

func (values *stringList) String() string { return fmt.Sprint([]string(*values)) }
func (values *stringList) Set(value string) error {
	*values = append(*values, value)
	return nil
}

func main() { os.Exit(realMain(os.Args[1:], os.Stdout, os.Stderr)) }

func realMain(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: pgo-pilot run|compare")
		return 2
	}
	switch args[0] {
	case "run":
		flags := flag.NewFlagSet("run", flag.ContinueOnError)
		flags.SetOutput(stderr)
		var config runConfig
		var guideStart string
		flags.StringVar(&config.Variant, "variant", "", "capture, off, or pgo")
		flags.StringVar(&config.Binary, "binary", "", "absolute Threadfin binary")
		flags.StringVar(&config.Output, "output", "", "absolute result JSON")
		flags.StringVar(&config.Profile, "profile", "", "absolute capture profile")
		flags.IntVar(&config.Pair, "pair", 0, "pair number")
		flags.IntVar(&config.Sequence, "sequence", 0, "run sequence")
		flags.DurationVar(&config.StreamDuration, "stream-duration", 0, "stream load duration")
		flags.IntVar(&config.Clients, "clients", 0, "concurrent clients")
		flags.Int64Var(&config.SampleBytes, "sample-bytes", 0, "bytes per request")
		flags.StringVar(&guideStart, "guide-start", "", "UTC RFC3339 guide start")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 {
			fmt.Fprintln(stderr, "run: unexpected positional arguments")
			return 2
		}
		parsed, err := time.Parse(time.RFC3339, guideStart)
		if err != nil {
			fmt.Fprintln(stderr, "run:", err)
			return 2
		}
		config.GuideStart = parsed
		if _, err := run(context.Background(), config); err != nil {
			fmt.Fprintln(stderr, "run:", err)
			return 1
		}
		return 0
	case "compare":
		flags := flag.NewFlagSet("compare", flag.ContinueOnError)
		flags.SetOutput(stderr)
		var sessions stringList
		var output string
		flags.Var(&sessions, "session", "absolute JSONL session; specify twice")
		flags.StringVar(&output, "output", "", "absolute summary JSON")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 || len(sessions) != 2 || output == "" {
			fmt.Fprintln(stderr, "compare requires two -session values and -output")
			return 2
		}
		compared, err := compareSessions(sessions[0], sessions[1], output)
		if err != nil {
			fmt.Fprintln(stderr, "compare:", err)
			return 1
		}
		fmt.Fprintln(stderr, compared.Verdict)
		return 0
	default:
		fmt.Fprintln(stderr, errors.New("usage: pgo-pilot run|compare"))
		return 2
	}
}
