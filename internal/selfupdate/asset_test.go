package selfupdate

import (
	"fmt"
	"runtime"
	"testing"
)

func TestFindAsset(t *testing.T) {
	expected := fmt.Sprintf("stackit-acl_0.4.0_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	release := &Release{
		TagName: "v0.4.0",
		Assets: []Asset{
			{Name: "stackit-acl_0.4.0_plan9_386.tar.gz"},
			{Name: "checksums.txt"},
			{Name: expected},
		},
	}

	asset, err := FindAsset(release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if asset.Name != expected {
		t.Errorf("asset = %q, want %q", asset.Name, expected)
	}
}

func TestFindAsset_NotFound(t *testing.T) {
	release := &Release{
		TagName: "v0.4.0",
		Assets:  []Asset{{Name: "checksums.txt"}},
	}
	if _, err := FindAsset(release); err == nil {
		t.Fatal("expected error for missing platform asset")
	}
}

func TestFindAssetByName(t *testing.T) {
	release := &Release{
		Assets: []Asset{
			{Name: "stackit-acl_0.4.0_darwin_arm64.tar.gz"},
			{Name: "checksums.txt"},
		},
	}

	if asset := FindAssetByName(release, "checksums.txt"); asset == nil {
		t.Error("expected checksums asset, got nil")
	}
	if asset := FindAssetByName(release, "missing.txt"); asset != nil {
		t.Errorf("expected nil for missing asset, got %q", asset.Name)
	}
}
