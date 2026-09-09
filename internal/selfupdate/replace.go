package selfupdate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func VerifyAccess(target string) error {
	if info, err := os.Lstat(target); err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file, specify a different path using --output", target)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat %s: %w", target, err)
	}

	if !writable(filepath.Dir(target)) {
		return notWritable(filepath.Dir(target))
	}
	return nil
}

func notWritable(path string) error {
	return fmt.Errorf("%s is not writable, run self-update as root or specify a different path using --output", path)
}

func writable(dir string) bool {
	file, err := os.CreateTemp(dir, "."+binaryName+"-*")
	if err != nil {
		return false
	}
	name := file.Name()
	_ = file.Close()
	return os.Remove(name) == nil
}

func Replace(target string, content []byte) (err error) {
	mode := os.FileMode(0o755)
	if info, statErr := os.Lstat(target); statErr == nil {
		mode = info.Mode().Perm()
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return fmt.Errorf("stat %s: %w", target, statErr)
	}

	file, err := os.CreateTemp(filepath.Dir(target), "."+binaryName+"-*")
	if err != nil {
		return fmt.Errorf("create temporary file next to %s: %w", target, err)
	}
	defer func() {
		_ = file.Close()
		if err != nil {
			_ = os.Remove(file.Name())
		}
	}()

	if _, err = file.Write(content); err != nil {
		return fmt.Errorf("write new binary: %w", err)
	}
	if err = file.Chmod(mode); err != nil {
		return fmt.Errorf("set permissions on new binary: %w", err)
	}
	if err = file.Close(); err != nil {
		return fmt.Errorf("close temporary file: %w", err)
	}
	if err = os.Rename(file.Name(), target); err != nil {
		return fmt.Errorf("replace %s: %w", target, err)
	}
	return nil
}
