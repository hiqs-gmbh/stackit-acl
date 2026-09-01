package ip

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("1.2.3.4\n"))
	}))
	defer server.Close()

	originalURL := ifconfigURL
	ifconfigURL = server.URL
	defer func() { ifconfigURL = originalURL }()

	addr, err := Fetch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addr != "1.2.3.4" {
		t.Errorf("expected 1.2.3.4, got %s", addr)
	}
}

func TestFetch_TrimsWhitespace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("  10.0.0.1  \n\n"))
	}))
	defer server.Close()

	originalURL := ifconfigURL
	ifconfigURL = server.URL
	defer func() { ifconfigURL = originalURL }()

	addr, err := Fetch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addr != "10.0.0.1" {
		t.Errorf("expected 10.0.0.1, got %s", addr)
	}
}

func TestFetch_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	originalURL := ifconfigURL
	ifconfigURL = server.URL
	defer func() { ifconfigURL = originalURL }()

	_, err := Fetch()
	if err == nil {
		t.Fatal("expected error for non-OK status")
	}
}

func TestFetch_InvalidIP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-an-ip"))
	}))
	defer server.Close()

	originalURL := ifconfigURL
	ifconfigURL = server.URL
	defer func() { ifconfigURL = originalURL }()

	_, err := Fetch()
	if err == nil {
		t.Fatal("expected error for invalid IP")
	}
}

func TestFetch_IPv6(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("2001:db8::1"))
	}))
	defer server.Close()

	originalURL := ifconfigURL
	ifconfigURL = server.URL
	defer func() { ifconfigURL = originalURL }()

	addr, err := Fetch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addr != "2001:db8::1" {
		t.Errorf("expected 2001:db8::1, got %s", addr)
	}
}
