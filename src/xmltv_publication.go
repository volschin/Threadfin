package src

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var xmltvPublicationMutex sync.Mutex

// Finish both representations before replacing either published artifact.
// Each rename is atomic for readers; two separate paths cannot be swapped
// as one filesystem transaction.
func publishXMLTVFiles(plainPath, gzipPath string, write func(io.Writer) error) (err error) {
	// Core rebuilds and the image-cache refresh can both publish a guide.
	// Keep rendering and both renames under one owner so their generations
	// cannot interleave and leave the plain and gzip outputs mismatched.
	xmltvPublicationMutex.Lock()
	defer xmltvPublicationMutex.Unlock()

	type output struct {
		file   *os.File
		path   string
		closed bool
	}
	var outputs []*output
	defer func() {
		for _, out := range outputs {
			if !out.closed {
				err = errors.Join(err, out.file.Close())
			}
			if removeErr := os.Remove(out.file.Name()); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				err = errors.Join(err, removeErr)
			}
		}
	}()

	for _, filename := range []string{plainPath, gzipPath} {
		mode := os.FileMode(0o600)
		info, statErr := os.Stat(filename)
		switch {
		case statErr == nil:
			if !info.Mode().IsRegular() {
				return fmt.Errorf("XMLTV output is not a regular file: %s", filename)
			}
			mode = info.Mode().Perm()
		case !errors.Is(statErr, os.ErrNotExist):
			return fmt.Errorf("inspect published XMLTV: %w", statErr)
		}
		file, createErr := os.CreateTemp(filepath.Dir(filename), ".threadfin-xmltv-*")
		if createErr != nil {
			return fmt.Errorf("create temporary XMLTV: %w", createErr)
		}
		outputs = append(outputs, &output{file: file, path: filename})
		if err := file.Chmod(mode); err != nil {
			return err
		}
	}

	compressed := gzip.NewWriter(outputs[1].file)
	writer := bufio.NewWriterSize(io.MultiWriter(outputs[0].file, compressed), 1<<20)
	if err := write(writer); err != nil {
		return err
	}
	if err := errors.Join(writer.Flush(), compressed.Close()); err != nil {
		return err
	}
	for _, out := range outputs {
		if err := out.file.Sync(); err != nil {
			return err
		}
		out.closed = true
		if err := out.file.Close(); err != nil {
			return err
		}
	}
	for _, out := range outputs {
		if err := os.Rename(out.file.Name(), out.path); err != nil {
			return fmt.Errorf("publish complete XMLTV: %w", err)
		}
	}
	return nil
}
