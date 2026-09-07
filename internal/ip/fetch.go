package ip

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

var endpoints = []string{
	"https://ifconfig.schwarz",
	"https://checkip.amazonaws.com",
	"https://v4.ident.me",
}

var (
	maxAttempts = 3
	retryDelay  = 2 * time.Second
)

func Fetch() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	var errs []error
	for _, endpoint := range endpoints {
		addr, err := fetchFrom(client, endpoint)
		if err == nil {
			return addr, nil
		}
		errs = append(errs, err)
	}
	return "", fmt.Errorf("fetch external IP: %w", errors.Join(errs...))
}

func fetchFrom(client *http.Client, endpoint string) (string, error) {
	attempts := max(maxAttempts, 1)

	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		var addr string
		addr, err = fetchOnce(client, endpoint)
		if err == nil {
			return addr, nil
		}
		if attempt < attempts {
			time.Sleep(retryDelay)
		}
	}
	return "", fmt.Errorf("%s (after %d attempts): %w", endpoint, attempts, err)
}

func fetchOnce(client *http.Client, endpoint string) (string, error) {
	resp, err := client.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read IP response: %w", err)
	}

	addr := strings.TrimSpace(string(body))
	if net.ParseIP(addr) == nil {
		return "", fmt.Errorf("invalid IP address received: %s", addr)
	}

	return addr, nil
}
