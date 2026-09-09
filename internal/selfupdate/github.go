package selfupdate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultBaseURL = "https://api.github.com/repos/hiqs-gmbh/stackit-acl/releases"
	httpTimeout    = 60 * time.Second
)

var ErrReleaseNotFound = errors.New("release not found")

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Digest             string `json:"digest"`
}

type Release struct {
	TagName string  `json:"tag_name"`
	URL     string  `json:"html_url"`
	Assets  []Asset `json:"assets"`
}

type Releases struct {
	Client  *http.Client
	BaseURL string
}

func NewReleases() *Releases {
	return &Releases{
		Client:  &http.Client{Timeout: httpTimeout},
		BaseURL: DefaultBaseURL,
	}
}

func (r *Releases) Latest() (*Release, error) {
	return r.fetchRelease(r.BaseURL + "/latest")
}

func (r *Releases) ReleaseByTag(tag string) (*Release, error) {
	return r.fetchRelease(r.BaseURL + "/tags/" + Tag(tag))
}

func (r *Releases) fetchRelease(url string) (*Release, error) {
	resp, err := r.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrReleaseNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API responded with %s: %s", resp.Status, apiErrorMessage(resp.Body))
	}

	release := &Release{}
	if err := json.NewDecoder(resp.Body).Decode(release); err != nil {
		return nil, fmt.Errorf("decode release from %s: %w", url, err)
	}
	return release, nil
}

func (r *Releases) Download(url string) ([]byte, error) {
	resp, err := r.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading %s failed with %s", url, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", url, err)
	}
	return data, nil
}

func apiErrorMessage(body io.Reader) string {
	var payload struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(io.LimitReader(body, 4096)).Decode(&payload); err != nil || payload.Message == "" {
		return "unexpected error"
	}
	return payload.Message
}
