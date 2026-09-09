package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestDigestBytes(t *testing.T) {
	sum := sha256.Sum256([]byte("content"))
	digest := "sha256:" + hex.EncodeToString(sum[:])

	got, ok := DigestBytes(digest)
	if !ok {
		t.Fatal("expected digest to be valid")
	}
	if string(got) != string(sum[:]) {
		t.Errorf("digest = %x, want %x", got, sum)
	}
}

func TestDigestBytes_Invalid(t *testing.T) {
	tests := []string{
		"",
		"sha256:",
		"sha256:not-hex!",
		"sha256:0123",
		"sha512:" + hex.EncodeToString(make([]byte, sha256.Size)),
	}
	for _, tt := range tests {
		if _, ok := DigestBytes(tt); ok {
			t.Errorf("DigestBytes(%q) = ok, want invalid", tt)
		}
	}
}

func TestParseChecksums(t *testing.T) {
	sum := sha256.Sum256([]byte("archive"))
	data := []byte("1111111111111111111111111111111111111111111111111111111111111111  stackit-acl_0.4.0_linux_amd64.tar.gz\n" +
		hex.EncodeToString(sum[:]) + "  stackit-acl_0.4.0_darwin_arm64.tar.gz\n")

	got, err := ParseChecksums(data, "stackit-acl_0.4.0_darwin_arm64.tar.gz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(sum[:]) {
		t.Errorf("checksum = %x, want %x", got, sum)
	}
}

func TestParseChecksums_NotFound(t *testing.T) {
	data := []byte("1111111111111111111111111111111111111111111111111111111111111111  stackit-acl_0.4.0_linux_amd64.tar.gz\n")
	if _, err := ParseChecksums(data, "missing.tar.gz"); err == nil {
		t.Fatal("expected error for missing checksum entry")
	}
}

func TestParseChecksums_InvalidHash(t *testing.T) {
	data := []byte("xyz  stackit-acl_0.4.0_linux_amd64.tar.gz\n")
	if _, err := ParseChecksums(data, "stackit-acl_0.4.0_linux_amd64.tar.gz"); err == nil {
		t.Fatal("expected error for invalid checksum entry")
	}
}

func TestVerifyChecksum(t *testing.T) {
	sum := sha256.Sum256([]byte("data"))
	if !VerifyChecksum([]byte("data"), sum[:]) {
		t.Error("expected checksum to match")
	}
	if VerifyChecksum([]byte("other"), sum[:]) {
		t.Error("expected checksum mismatch")
	}
}
