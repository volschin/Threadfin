package src

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func zipFiles(sourceFiles []string, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}

	writeErr := writeZIP(sourceFiles, zipfile)
	closeErr := zipfile.Close()
	return removeIncompleteFile(target, errors.Join(writeErr, closeErr))
}

func removeIncompleteFile(path string, operationErr error) error {
	if operationErr == nil {
		return nil
	}
	if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
		return errors.Join(operationErr, removeErr)
	}
	return operationErr
}

func writeZIP(sourceFiles []string, target io.Writer) (err error) {
	archive := zip.NewWriter(target)
	defer func() {
		err = errors.Join(err, archive.Close())
	}()

	for _, source := range sourceFiles {
		info, err := os.Stat(source)
		if err != nil {
			return err
		}

		sourceIsDir := info.IsDir()

		if err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			if sourceIsDir {
				header.Name = filepath.Join(strings.TrimPrefix(path, System.Folder.Config))
			}

			if info.IsDir() {
				header.Name += string(os.PathSeparator)
			} else {
				header.Method = zip.Deflate
			}

			writer, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}

			_, copyErr := io.Copy(writer, file)
			closeErr := file.Close()
			return errors.Join(copyErr, closeErr)
		}); err != nil {
			return err
		}

	}

	return nil
}

func extractGZIP(gzipBody []byte, fileSource string) (body []byte, err error) {
	return extractGZIPBounded(gzipBody, fileSource, providerResponseBodyLimit)
}

func extractGZIPBounded(gzipBody []byte, fileSource string, limit int64) (body []byte, err error) {

	var b = bytes.NewBuffer(gzipBody)

	r, err := gzip.NewReader(b)
	if err != nil {
		// Keine gzip Datei
		body = gzipBody
		err = nil
		return
	}

	showInfo("Extract gzip:" + sanitizeProviderSourceForLog(fileSource))

	var resB bytes.Buffer
	_, readErr := io.Copy(&resB, io.LimitReader(r, limit+1))
	closeErr := r.Close()
	if readErr != nil || closeErr != nil {
		return nil, errors.New("provider compressed response could not be read")
	}
	if int64(resB.Len()) > limit {
		return nil, errProviderResponseTooLarge
	}

	body = resB.Bytes()
	return
}

func compressGZIP(data *[]byte, file string) (err error) {

	if len(file) != 0 {

		f, err := os.Create(file)
		if err != nil {
			return err
		}

		w := gzip.NewWriter(f)
		_, writeErr := w.Write(*data)
		writerCloseErr := w.Close()
		fileCloseErr := f.Close()
		err = removeIncompleteFile(file, errors.Join(writeErr, writerCloseErr, fileCloseErr))
	}

	return
}

func compressGZIPFile(sourcePath, targetPath string) (err error) {
	in, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}

	gw := gzip.NewWriter(out)
	_, copyErr := io.Copy(gw, in)
	writerCloseErr := gw.Close()
	fileCloseErr := out.Close()
	return removeIncompleteFile(targetPath, errors.Join(copyErr, writerCloseErr, fileCloseErr))
}
