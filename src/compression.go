package src

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
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
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	for _, source := range sourceFiles {

		info, err := os.Stat(source)
		if err != nil {
			return nil
		}

		var baseDir string
		if info.IsDir() {
			baseDir = filepath.Base(System.Folder.Data)
		}

		filepath.Walk(source, func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			if baseDir != "" {
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
			defer file.Close()

			_, err = io.Copy(writer, file)

			return err

		})

	}

	return err
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

func extractGZIP(gzipBody []byte, fileSource string) (body []byte, err error) {

	var b = bytes.NewBuffer(gzipBody)

	var r io.Reader
	r, err = gzip.NewReader(b)
	if err != nil {
		// Keine gzip Datei
		body = gzipBody
		err = nil
		return
	}

	showInfo("Extract gzip:" + fileSource)

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		body = gzipBody
		err = nil
		return
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
		w.Write(*data)
		w.Close()
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
	defer out.Close()

	gw := gzip.NewWriter(out)
	defer gw.Close()

	_, err = io.Copy(gw, in)
	return err
}
