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
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	originalURL := ifconfigURL
	ifconfigURL = server.URL
	defer func() { ifconfigURL = originalURL }()

	originalDelay := retryDelay
	retryDelay = 0
	defer func() { retryDelay = originalDelay }()

	_, err := Fetch()
	if err == nil {
		t.Fatal("expected error for non-OK status")
	}
	if calls != maxAttempts {
		t.Errorf("expected %d calls, got %d", maxAttempts, calls)
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

	originalDelay := retryDelay
	retryDelay = 0
	defer func() { retryDelay = originalDelay }()

	_, err := Fetch()
	if err == nil {
		t.Fatal("expected error for invalid IP")
	}
}

func TestFetch_RetriesTransientError(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("1.2.3.4\n"))
	}))
	defer server.Close()

	originalURL := ifconfigURL
	ifconfigURL = server.URL
	defer func() { ifconfigURL = originalURL }()

	originalDelay := retryDelay
	retryDelay = 0
	defer func() { retryDelay = originalDelay }()

	addr, err := Fetch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addr != "1.2.3.4" {
		t.Errorf("expected 1.2.3.4, got %s", addr)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
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
