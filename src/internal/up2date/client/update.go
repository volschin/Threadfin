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

var updateHTTPClient = &http.Client{Timeout: 2 * time.Minute}

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
		_ = os.Remove(oldBinary)
		if err := os.Rename(binary, oldBinary); err != nil {
			return err
		}
		if err := copyFile(candidate, binary); err != nil {
			restorOldBinary(oldBinary, binary)
			return err
		}
		if err := os.Chmod(binary, 0755); err != nil {
			restorOldBinary(oldBinary, binary)
			return err
		}

		log.Println("["+strings.ToUpper(fileType)+"]", "Update Successful")

		// Restart binary (Windows)
		if runtime.GOOS == "windows" {

			bin, err := os.Executable()

			if err != nil {
				restorOldBinary(oldBinary, binary)
				return err
			}

			var pid = os.Getpid()
			var process, _ = os.FindProcess(pid)

			if proc, err := start(bin); err == nil {

				os.RemoveAll(oldBinary)
				process.Kill()
				proc.Wait()

			} else {
				restorOldBinary(oldBinary, binary)
			}

		} else {

			// Restart binary (Linux and UNIX)
			err = syscall.Exec(binary, os.Args, os.Environ())
			if err != nil {
				restorOldBinary(oldBinary, binary)
				return err
			}

		}

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

func restorOldBinary(oldBinary, newBinary string) {
	os.RemoveAll(newBinary)
	os.Rename(oldBinary, newBinary)
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
