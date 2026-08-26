package archive

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ExtractZIP extracts archivePath into targetDir without permitting archive
// path traversal or symbolic-link resolution outside targetDir. os.Root does
// not prevent filesystem-boundary traversal, Linux bind mounts, special
// filesystems, device-file access, or writes through pre-existing hard links.
func ExtractZIP(archivePath, targetDir string) (err error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open ZIP archive %q: %w", archivePath, err)
	}
	defer func() {
		err = errors.Join(err, wrapOperationError("close ZIP archive", reader.Close()))
	}()

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create ZIP target %q: %w", targetDir, err)
	}
	targetRoot, err := os.OpenRoot(targetDir)
	if err != nil {
		return fmt.Errorf("open ZIP target %q: %w", targetDir, err)
	}
	defer func() {
		err = errors.Join(err, wrapOperationError("close ZIP target", targetRoot.Close()))
	}()

	for _, entry := range reader.File {
		if err := extractEntry(targetRoot, entry); err != nil {
			return err
		}
	}
	return nil
}

func extractEntry(targetRoot *os.Root, entry *zip.File) error {
	localName, err := localArchiveName(entry.Name)
	if err != nil {
		return err
	}
	if entry.FileInfo().IsDir() {
		parent := filepath.Dir(localName)
		if parent != "." {
			if err := targetRoot.MkdirAll(parent, 0755); err != nil {
				return fmt.Errorf("archive entry %q: create parent directories: %w", entry.Name, err)
			}
		}
		if err := targetRoot.MkdirAll(localName, entry.Mode().Perm()); err != nil {
			return fmt.Errorf("archive entry %q: create directory: %w", entry.Name, err)
		}
		return nil
	}
	if err := targetRoot.MkdirAll(filepath.Dir(localName), 0755); err != nil {
		return fmt.Errorf("archive entry %q: create parent directories: %w", entry.Name, err)
	}

	source, err := entry.Open()
	if err != nil {
		return fmt.Errorf("archive entry %q: open source: %w", entry.Name, err)
	}
	destination, err := targetRoot.OpenFile(localName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, entry.Mode().Perm())
	if err != nil {
		return errors.Join(
			fmt.Errorf("archive entry %q: open destination: %w", entry.Name, err),
			wrapOperationError(fmt.Sprintf("archive entry %q: close source", entry.Name), source.Close()),
		)
	}

	operationErr := copyAndClose(destination, source)
	if operationErr == nil {
		return nil
	}
	removeErr := targetRoot.Remove(localName)
	if removeErr != nil && !errors.Is(removeErr, fs.ErrNotExist) {
		operationErr = errors.Join(operationErr, fmt.Errorf("remove incomplete archive entry %q: %w", entry.Name, removeErr))
	}
	return fmt.Errorf("archive entry %q: %w", entry.Name, operationErr)
}

func copyAndClose(destination io.WriteCloser, source io.ReadCloser) error {
	_, copyErr := io.Copy(destination, source)
	return errors.Join(
		wrapOperationError("copy entry data", copyErr),
		wrapOperationError("close destination", destination.Close()),
		wrapOperationError("close source", source.Close()),
	)
}

func localArchiveName(name string) (string, error) {
	cleanName := path.Clean(name)
	if path.IsAbs(cleanName) {
		return "", fmt.Errorf("archive entry %q is absolute", name)
	}
	if cleanName == ".." || strings.HasPrefix(cleanName, "../") {
		return "", fmt.Errorf("archive entry %q traverses outside target directory", name)
	}
	if !filepath.IsLocal(cleanName) {
		return "", fmt.Errorf("archive entry %q is not local", name)
	}
	localName, err := filepath.Localize(cleanName)
	if err != nil {
		return "", fmt.Errorf("archive entry %q is invalid on this platform: %w", name, err)
	}
	return localName, nil
}

func wrapOperationError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}
