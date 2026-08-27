//go:build !windows

package src

import "os"

func openFinalM3U(filename string) (m3uReadFile, error) {
	return os.Open(filename)
}
