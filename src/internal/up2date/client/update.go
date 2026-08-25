package up2date

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
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
		defer cleanup()

		filename := getFilenameFromPath(binary)
		oldBinary := path + "_old_" + filename
		if err := replacePreparedUpdate(candidate, binary, oldBinary); err != nil {
			return err
		}

		log.Println("["+strings.ToUpper(fileType)+"]", "Update Successful")

		// Restart binary (Windows)
		if runtime.GOOS == "windows" {
			return restartWindows(binary, oldBinary, start)
		}

		// Restart binary (Linux and UNIX)
		return restartUnix(binary, oldBinary, cleanup, syscall.Exec)

	}

	return
}

func start(args ...string) (p *os.Process, err error) {

	if args[0], err = exec.LookPath(args[0]); err == nil {
		//fmt.Println(args[0])
		var procAttr os.ProcAttr
		procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
		p, err := os.StartProcess(args[0], args, &procAttr)

		if err == nil {
			return p, nil
		}

	}

	return nil, err
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

func restartWindows(binary, oldBinary string, startProcess func(...string) (*os.Process, error)) error {
	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		return restoreAfterUpdateFailure(err, oldBinary, binary)
	}
	proc, startErr := startProcess(binary)
	if startErr != nil {
		return restoreAfterUpdateFailure(startErr, oldBinary, binary)
	}
	_ = os.RemoveAll(oldBinary)
	_ = process.Kill()
	_, _ = proc.Wait()
	return nil
}

func restartUnix(binary, oldBinary string, cleanup func(), execProcess func(string, []string, []string) error) error {
	cleanup()
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

func extractZIP(archive, target string) (err error) {

	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {

		path, err := archiveDestination(target, file.Name)
		if err != nil {
			return err
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.Mode()); err != nil {
				return err
			}
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			fileReader.Close()
			return err
		}

		_, copyErr := io.Copy(targetFile, fileReader)
		closeTargetErr := targetFile.Close()
		closeSourceErr := fileReader.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeTargetErr != nil {
			return closeTargetErr
		}
		if closeSourceErr != nil {
			return closeSourceErr
		}

	}

	return nil
}

func archiveDestination(target, name string) (string, error) {
	cleanName := filepath.Clean(filepath.FromSlash(name))
	if filepath.IsAbs(cleanName) || cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("archive entry %q escapes target directory", name)
	}

	targetRoot, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	destination, err := filepath.Abs(filepath.Join(targetRoot, cleanName))
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(targetRoot, destination)
	if err != nil {
		return "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("archive entry %q escapes target directory", name)
	}

	return destination, nil
}
