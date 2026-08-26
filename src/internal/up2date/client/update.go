package up2date

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var (
	updateHTTPClient  = &http.Client{Timeout: 2 * time.Minute}
	chmodUpdateBinary = os.Chmod
)

// DoUpdate : Update binary
func DoUpdate(fileType, filenameBIN string) (err error) {

	var url string
	switch fileType {
	case "bin":
		url = Updater.Response.UpdateBIN
	case "zip":
		url = Updater.Response.UpdateZIP
	}

	switch runtime.GOOS {
	case "windows":
		filenameBIN = filenameBIN + ".exe"
	}

	if len(url) > 0 {
		log.Println("["+strings.ToUpper(fileType)+"]", "New version ("+Updater.Name+"):", Updater.Response.Version)

		binary, err := os.Executable()
		if err != nil {
			return err
		}
		publicKey, err := embeddedUpdatePublicKey()
		if err != nil {
			return err
		}
		path := getPlatformPath(binary)
		candidate, cleanup, err := prepareVerifiedUpdate(updateHTTPClient, url, fileType, filenameBIN, path, publicKey)
		if err != nil {
			return err
		}

		filename := getFilenameFromPath(binary)
		oldBinary := path + "_old_" + filename
		if err := replacePreparedUpdate(candidate, binary, oldBinary); err != nil {
			if cleanupErr := cleanup(); cleanupErr != nil {
				return fmt.Errorf("%w; cleanup failed: %w", err, cleanupErr)
			}
			return err
		}

		log.Println("["+strings.ToUpper(fileType)+"]", "Update Successful")

		// Restart binary (Windows)
		if runtime.GOOS == "windows" {
			return beginWindowsHandoff(binary, oldBinary, os.Args[1:], cleanup, defaultWindowsUpdateProtocol())
		}

		// Restart binary (Linux and UNIX)
		return restartUnix(binary, oldBinary, cleanup, syscall.Exec)

	}

	return
}

func replacePreparedUpdate(candidate, binary, oldBinary string) error {
	_ = os.Remove(oldBinary)
	if err := os.Rename(binary, oldBinary); err != nil {
		return err
	}
	if err := copyFile(candidate, binary); err != nil {
		return restoreAfterUpdateFailure(err, oldBinary, binary)
	}
	if err := chmodUpdateBinary(binary, 0755); err != nil {
		return restoreAfterUpdateFailure(err, oldBinary, binary)
	}
	return nil
}

func wrapError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func restartUnix(binary, oldBinary string, cleanup func() error, execProcess func(string, []string, []string) error) error {
	if err := cleanup(); err != nil {
		return restoreAfterUpdateFailure(fmt.Errorf("clean temporary update material: %w", err), oldBinary, binary)
	}
	if err := execProcess(binary, os.Args, os.Environ()); err != nil {
		return restoreAfterUpdateFailure(err, oldBinary, binary)
	}
	return nil
}

func restoreAfterUpdateFailure(updateErr error, oldBinary, newBinary string) error {
	if restoreErr := restorOldBinary(oldBinary, newBinary); restoreErr != nil {
		return fmt.Errorf("%w; rollback failed: %w", updateErr, restoreErr)
	}
	return updateErr
}

func restorOldBinary(oldBinary, newBinary string) error {
	if err := os.RemoveAll(newBinary); err != nil {
		return fmt.Errorf("remove failed update: %w", err)
	}
	if err := os.Rename(oldBinary, newBinary); err != nil {
		return fmt.Errorf("restore previous binary: %w", err)
	}
	return nil
}

func getPlatformFile(filename string) string {

	path, file := filepath.Split(filename)
	var newPath = filepath.Dir(path)
	var newFileName = newPath + string(os.PathSeparator) + file

	return newFileName
}

func getFilenameFromPath(path string) string {

	file := filepath.Base(path)

	return file
}

func getPlatformPath(path string) string {

	var newPath = filepath.Dir(path) + string(os.PathSeparator)

	return newPath
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
