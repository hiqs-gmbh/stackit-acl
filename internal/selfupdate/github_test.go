package selfupdate

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestReleases(server *httptest.Server) *Releases {
	return &Releases{Client: server.Client(), BaseURL: server.URL}
}

func TestReleases_Latest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/latest" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
  "tag_name": "v0.4.0",
  "html_url": "https://github.com/hiqs-gmbh/stackit-acl/releases/tag/v0.4.0",
  "assets": [
    {"name": "stackit-acl_0.4.0_darwin_arm64.tar.gz", "browser_download_url": "https://example.com/stackit-acl_0.4.0_darwin_arm64.tar.gz", "digest": "sha256:0123"},
    {"name": "checksums.txt", "browser_download_url": "https://example.com/checksums.txt"}
  ]
}`))
	}))
	defer server.Close()

	release, err := newTestReleases(server).Latest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if release.TagName != "v0.4.0" {
		t.Errorf("TagName = %q, want %q", release.TagName, "v0.4.0")
	}
	if release.URL != "https://github.com/hiqs-gmbh/stackit-acl/releases/tag/v0.4.0" {
		t.Errorf("URL = %q", release.URL)
	}
	if len(release.Assets) != 2 {
		t.Fatalf("assets = %d, want 2", len(release.Assets))
	}
	asset := release.Assets[0]
	if asset.Name != "stackit-acl_0.4.0_darwin_arm64.tar.gz" {
		t.Errorf("asset name = %q", asset.Name)
	}
	if asset.BrowserDownloadURL != "https://example.com/stackit-acl_0.4.0_darwin_arm64.tar.gz" {
		t.Errorf("asset download URL = %q", asset.BrowserDownloadURL)
	}
	if asset.Digest != "sha256:0123" {
		t.Errorf("asset digest = %q", asset.Digest)
	}
}

func TestReleases_ReleaseByTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tags/v0.4.0" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tag_name": "v0.4.0", "assets": []}`))
	}))
	defer server.Close()

	release, err := newTestReleases(server).ReleaseByTag("0.4.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if release.TagName != "v0.4.0" {
		t.Errorf("TagName = %q, want %q", release.TagName, "v0.4.0")
	}
}

func TestReleases_ReleaseByTag_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	_, err := newTestReleases(server).ReleaseByTag("v9.9.9")
	if !errors.Is(err, ErrReleaseNotFound) {
		t.Fatalf("error = %v, want ErrReleaseNotFound", err)
	}
}

func TestReleases_Latest_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(http.NotFound))
	defer server.Close()

	_, err := newTestReleases(server).Latest()
	if !errors.Is(err, ErrReleaseNotFound) {
		t.Fatalf("error = %v, want ErrReleaseNotFound", err)
	}
}

func TestReleases_FetchRelease_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message": "API rate limit exceeded"}`))
	}))
	defer server.Close()

	_, err := newTestReleases(server).Latest()
	if err == nil {
		t.Fatal("expected error for non-OK status")
	}
	if !strings.Contains(err.Error(), "API rate limit exceeded") {
		t.Errorf("error = %v, want it to contain the API message", err)
	}
}

func TestReleases_Download(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/asset" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("binary"))
	}))
	defer server.Close()

	data, err := newTestReleases(server).Download(server.URL + "/asset")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "binary" {
		t.Errorf("data = %q, want %q", data, "binary")
	}
}

func TestReleases_Download_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := newTestReleases(server).Download(server.URL + "/asset")
	if err == nil {
		t.Fatal("expected error for non-OK status")
	}
}
