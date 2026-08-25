//go:build !windows

package up2date

import "errors"

func startOSUpdateProcess(string, []string) (updateProcess, error) {
	return nil, errors.New("Windows update processes are unavailable on this platform")
}

func findOSUpdateProcess(int) (updateProcess, error) {
	return nil, errors.New("Windows update processes are unavailable on this platform")
}
