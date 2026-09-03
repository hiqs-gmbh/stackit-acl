package stackit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hiqs-gmbh/stackit-acl/internal/services"
)

type Client struct {
	projectID string
	region    string
}

type Project struct {
	ID   string
	Name string
}

type Instance struct {
	ID   string
	Name string
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

func ListProjects() ([]Project, error) {
	args := []string{"project", "list", "--output-format", "json", "--verbosity", "error"}
	out, err := runRaw(args)
	if err != nil {
		return nil, err
	}

	var raw []struct {
		ProjectID string `json:"projectId"`
		Name      string `json:"name"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse project list JSON: %w", err)
	}

	projects := make([]Project, 0, len(raw))
	for _, p := range raw {
		projects = append(projects, Project{ID: p.ProjectID, Name: p.Name})
	}
	return projects, nil
}

func ListInstances(projectID, region string, cfg services.ServiceConfig) ([]Instance, error) {
	args := []string{cfg.Name, cfg.ResourceGroup, "list", "--output-format", "json", "--verbosity", "error", "-p", projectID}
	if region != "" {
		args = append(args, "--region", region)
	}
	out, err := runRaw(args)
	if err != nil {
		return nil, err
	}

	var raw []struct {
		ID         string `json:"id"`
		InstanceID string `json:"instanceId"`
		Name       string `json:"name"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse instance list JSON: %w", err)
	}

	instances := make([]Instance, 0, len(raw))
	for _, i := range raw {
		id := i.ID
		if id == "" {
			id = i.InstanceID
		}
		instances = append(instances, Instance{ID: id, Name: i.Name})
	}
	return instances, nil
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
	return c.run(args)
}

func (c *Client) UpdateInstanceACL(cfg services.ServiceConfig, resourceID string, cidrs []string) error {
	args := []string{cfg.Name, cfg.ResourceGroup, "update", resourceID}
	for _, cidr := range cidrs {
		args = append(args, "--acl", cidr)
	}
	args = append(args, "--assume-yes")
	_, err := c.run(args)
	return err
}

func (c *Client) GeneratePayload(cfg services.ServiceConfig, resourceID string) ([]byte, error) {
	args := []string{cfg.Name, cfg.ResourceGroup, "generate-payload", "--cluster-name", resourceID}
	return c.run(args)
}

func (c *Client) UpdateClusterACL(cfg services.ServiceConfig, resourceID string, payload []byte) error {
	tmpFile, err := os.CreateTemp("", "stackit-acl-payload-*.json")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.Write(payload); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	args := []string{cfg.Name, cfg.ResourceGroup, "update", resourceID, "--payload", "@" + tmpFile.Name(), "--assume-yes"}
	_, err = c.run(args)
	return err
}

func (c *Client) run(args []string) ([]byte, error) {
	args = append(args, c.globalArgs()...)
	return runRaw(args)
}

type commandError struct {
	args   []string
	stderr string
	err    error
}

func (e *commandError) Error() string {
	return fmt.Sprintf("stackit %s: %v\n%s", strings.Join(e.args, " "), e.err, e.stderr)
}

func IsAuthError(err error) bool {
	var cmdErr *commandError
	if !errors.As(err, &cmdErr) {
		return false
	}
	msg := strings.ToLower(cmdErr.stderr)
	return strings.Contains(msg, "not authenticated") ||
		strings.Contains(msg, "unauthorized") ||
		strings.Contains(msg, "401")
}

func runRaw(args []string) ([]byte, error) {
	cmd := exec.Command("stackit", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, &commandError{args: args, stderr: stderr.String(), err: err}
	}

	return stdout.Bytes(), nil
}
