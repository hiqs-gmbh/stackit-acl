package selfupdate

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const ChecksumsName = "checksums.txt"

func DigestBytes(digest string) ([]byte, bool) {
	value, ok := strings.CutPrefix(digest, "sha256:")
	if !ok {
		return nil, false
	}
	hash, err := hex.DecodeString(value)
	if err != nil || len(hash) != sha256.Size {
		return nil, false
	}
	return hash, true
}

func ParseChecksums(data []byte, filename string) ([]byte, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		hash, name, ok := strings.Cut(scanner.Text(), "  ")
		if !ok || name != filename {
			continue
		}
		want, err := hex.DecodeString(hash)
		if err != nil || len(want) != sha256.Size {
			return nil, fmt.Errorf("invalid checksum for %s", filename)
		}
		return want, nil
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read checksums: %w", err)
	}
	return nil, fmt.Errorf("%s not found in checksums", filename)
}

func VerifyChecksum(data, want []byte) bool {
	got := sha256.Sum256(data)
	return bytes.Equal(got[:], want)
}
