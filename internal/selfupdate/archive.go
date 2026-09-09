package selfupdate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"path"
)

func ExtractBinary(archive []byte) ([]byte, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("open release archive: %w", err)
	}
	defer func() { _ = gzipReader.Close() }()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read release archive: %w", err)
		}
		if header.Typeflag != tar.TypeReg || path.Base(header.Name) != binaryName {
			continue
		}
		binary, err := io.ReadAll(tarReader)
		if err != nil {
			return nil, fmt.Errorf("read %s from release archive: %w", header.Name, err)
		}
		return binary, nil
	}
	return nil, fmt.Errorf("%s not found in release archive", binaryName)
}
