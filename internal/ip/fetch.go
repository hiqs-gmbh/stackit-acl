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

func Fetch() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(ifconfigURL)
	if err != nil {
		return "", fmt.Errorf("fetch external IP: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch external IP: unexpected status code %d", resp.StatusCode)
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
