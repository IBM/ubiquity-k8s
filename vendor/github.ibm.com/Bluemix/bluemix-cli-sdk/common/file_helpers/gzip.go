package file_helpers

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

func ExtractTgz(src string, dest string) error {
	fd, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fd.Close()

	gReader, err := gzip.NewReader(fd)
	if err != nil {
		return err
	}
	defer gReader.Close()

	tarReader := tar.NewReader(gReader)

	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if hdr.Name == "." {
			continue
		}

		err = extractFileInArchive(tarReader, hdr, dest)
		if err != nil {
			return err
		}
	}

	return nil
}

func extractFileInArchive(r io.Reader, hdr *tar.Header, dest string) error {
	fi := hdr.FileInfo()
	path := filepath.Join(dest, hdr.Name)

	if fi.IsDir() {
		return os.MkdirAll(path, fi.Mode())
	} else {
		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			return err
		}

		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(f, r)
		return err
	}
}
