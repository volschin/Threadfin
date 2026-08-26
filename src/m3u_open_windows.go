//go:build windows

package src

import (
	"os"

	"golang.org/x/sys/windows"
)

func openFinalM3U(filename string) (m3uReadFile, error) {
	path, err := windows.UTF16PtrFromString(filename)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: filename, Err: err}
	}

	handle, err := windows.CreateFile(
		path,
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: filename, Err: err}
	}
	return os.NewFile(uintptr(handle), filename), nil
}
