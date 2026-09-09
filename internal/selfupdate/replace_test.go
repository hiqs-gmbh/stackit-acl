package selfupdate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyAccess_NewFile(t *testing.T) {
	dir := t.TempDir()
	if err := VerifyAccess(filepath.Join(dir, "stackit-acl")); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVerifyAccess_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "stackit-acl")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if err := VerifyAccess(target); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVerifyAccess_Directory(t *testing.T) {
	dir := t.TempDir()
	if err := VerifyAccess(dir); err == nil {
		t.Error("expected error for directory target")
	}
}

func TestVerifyAccess_MissingDir(t *testing.T) {
	target := filepath.Join(t.TempDir(), "missing", "stackit-acl")
	if err := VerifyAccess(target); err == nil {
		t.Error("expected error for missing parent directory")
	}
}

func TestReplace_NewFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "stackit-acl")

	if err := Replace(target, []byte("binary")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(content) != "binary" {
		t.Errorf("content = %q, want %q", content, "binary")
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat target: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o755 {
		t.Errorf("permissions = %o, want 755", perm)
	}
}

func TestReplace_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "stackit-acl")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatalf("write target: %v", err)
	}

	if err := Replace(target, []byte("new")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(content) != "new" {
		t.Errorf("content = %q, want %q", content, "new")
	}
}

func TestReplace_PreservesPermissions(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "stackit-acl")
	if err := os.WriteFile(target, []byte("old"), 0o700); err != nil {
		t.Fatalf("write target: %v", err)
	}

	if err := Replace(target, []byte("new")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat target: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o700 {
		t.Errorf("permissions = %o, want 700", perm)
	}
}

func TestReplace_MissingDir(t *testing.T) {
	target := filepath.Join(t.TempDir(), "missing", "stackit-acl")
	if err := Replace(target, []byte("binary")); err == nil {
		t.Error("expected error for missing parent directory")
	}
}
