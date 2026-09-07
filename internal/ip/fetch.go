package ip

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

var ifconfigURL = "https://ifconfig.schwarz"

var (
	maxAttempts = 3
	retryDelay  = 2 * time.Second
)

func Fetch() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	attempts := max(maxAttempts, 1)

	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		var addr string
		addr, err = fetchOnce(client)
		if err == nil {
			return addr, nil
		}
		if attempt < attempts {
			time.Sleep(retryDelay)
		}
	}
	return "", fmt.Errorf("fetch external IP after %d attempts: %w", attempts, err)
}

func fetchOnce(client *http.Client) (string, error) {
	resp, err := client.Get(ifconfigURL)
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
