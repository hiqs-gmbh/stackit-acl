package stackit

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"stackit-acl/internal/services"
)

type Client struct {
	projectID string
	region    string
}

func New(projectID, region string) *Client {
	return &Client{
		projectID: projectID,
		region:    region,
	}
}

func CheckAvailable() error {
	_, err := exec.LookPath("stackit")
	if err != nil {
		return fmt.Errorf("stackit CLI not found in PATH. Please install it first: https://github.com/stackitcloud/stackit-cli")
	}
	return nil
}

func (c *Client) globalArgs() []string {
	args := []string{"-p", c.projectID, "--verbosity", "error"}
	if c.region != "" {
		args = append(args, "--region", c.region)
	}
	return args
}

func (c *Client) DescribeInstance(cfg services.ServiceConfig, resourceID string) ([]byte, error) {
	args := []string{cfg.Name, cfg.ResourceGroup, "describe", resourceID, "--output-format", "json"}
	args = append(args, c.globalArgs()...)
	return c.run(args)
}

func (c *Client) UpdateInstanceACL(cfg services.ServiceConfig, resourceID string, cidrs []string) error {
	args := []string{cfg.Name, cfg.ResourceGroup, "update", resourceID}
	for _, cidr := range cidrs {
		args = append(args, "--acl", cidr)
	}
	args = append(args, "--assume-yes")
	args = append(args, c.globalArgs()...)
	_, err := c.run(args)
	return err
}

func (c *Client) GeneratePayload(cfg services.ServiceConfig, resourceID string) ([]byte, error) {
	args := []string{cfg.Name, cfg.ResourceGroup, "generate-payload", "--cluster-name", resourceID}
	args = append(args, c.globalArgs()...)
	return c.run(args)
}

func (c *Client) UpdateClusterACL(cfg services.ServiceConfig, resourceID string, payload []byte) error {
	tmpFile, err := os.CreateTemp("", "stackit-acl-payload-*.json")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(payload); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	args := []string{cfg.Name, cfg.ResourceGroup, "update", resourceID, "--payload", "@" + tmpFile.Name(), "--assume-yes"}
	args = append(args, c.globalArgs()...)
	_, err = c.run(args)
	return err
}

func (c *Client) run(args []string) ([]byte, error) {
	cmd := exec.Command("stackit", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("stackit %s: %w\n%s", strings.Join(args, " "), err, stderr.String())
	}

	return stdout.Bytes(), nil
}
