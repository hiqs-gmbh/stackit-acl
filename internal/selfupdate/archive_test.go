package selfupdate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"testing"
)

func testArchive(t *testing.T, entries map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)

	for name, content := range entries {
		if err := tarWriter.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     name,
			Mode:     0o755,
			Size:     int64(len(content)),
		}); err != nil {
			t.Fatalf("write tar header: %v", err)
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			t.Fatalf("write tar entry: %v", err)
		}
	}

	if err := tarWriter.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	return buf.Bytes()
}

func TestExtractBinary(t *testing.T) {
	archive := testArchive(t, map[string]string{
		"stackit-acl_0.4.0_darwin_arm64/LICENSE":     "license",
		"stackit-acl_0.4.0_darwin_arm64/stackit-acl": "binary content",
	})

	got, err := ExtractBinary(archive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "binary content" {
		t.Errorf("binary = %q, want %q", got, "binary content")
	}
}

func TestExtractBinary_RootEntry(t *testing.T) {
	archive := testArchive(t, map[string]string{
		"stackit-acl": "binary content",
	})

	got, err := ExtractBinary(archive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "binary content" {
		t.Errorf("binary = %q, want %q", got, "binary content")
	}
}

func TestExtractBinary_NotFound(t *testing.T) {
	archive := testArchive(t, map[string]string{
		"stackit-acl_0.4.0_darwin_arm64/LICENSE": "license",
	})
	if _, err := ExtractBinary(archive); err == nil {
		t.Fatal("expected error for missing binary in archive")
	}
}

func TestExtractBinary_InvalidArchive(t *testing.T) {
	if _, err := ExtractBinary([]byte("not a gzip archive")); err == nil {
		t.Fatal("expected error for invalid archive")
	}
}
