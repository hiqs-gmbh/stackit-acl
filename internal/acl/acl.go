package acl

import (
	"encoding/json"
	"fmt"
	"strings"

	"stackit-acl/internal/services"
)

func ToCIDR(ip string, prefix int) string {
	return fmt.Sprintf("%s/%d", ip, prefix)
}

func AppendCIDR(existing []string, cidr string) []string {
	for _, c := range existing {
		if c == cidr {
			return existing
		}
	}
	return append(existing, cidr)
}

func ExtractACLs(jsonData []byte, cfg services.ServiceConfig) ([]string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("parse JSON: %w\nRaw output: %s", err, string(jsonData))
	}

	val := navigatePath(data, cfg.ACLJSONPath)

	switch cfg.ACLType {
	case services.ACLArray:
		if val == nil {
			return []string{}, nil
		}
		arr, ok := val.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected array at path %s, got %T", cfg.ACLJSONPath, val)
		}
		result := make([]string, 0, len(arr))
		for _, item := range arr {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("expected string in ACL array, got %T", item)
			}
			result = append(result, s)
		}
		return result, nil

	case services.ACLCommaString:
		if val == nil {
			return []string{}, nil
		}
		s, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("expected string at path %s, got %T", cfg.ACLJSONPath, val)
		}
		if s == "" {
			return []string{}, nil
		}
		parts := strings.Split(s, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unknown ACL type: %s", cfg.ACLType)
	}
}

func SetACLs(jsonData []byte, cfg services.ServiceConfig, cidrs []string) ([]byte, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	pathParts := strings.Split(cfg.ACLJSONPath, ".")
	setNestedValue(data, pathParts, cidrs)

	return json.MarshalIndent(data, "", "  ")
}

func setNestedValue(data map[string]interface{}, pathParts []string, cidrs []string) {
	if len(pathParts) == 1 {
		data[pathParts[0]] = cidrs
		return
	}

	next, ok := data[pathParts[0]].(map[string]interface{})
	if !ok {
		next = make(map[string]interface{})
		data[pathParts[0]] = next
	}
	setNestedValue(next, pathParts[1:], cidrs)
}

func navigatePath(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current, ok = m[part]
		if !ok {
			return nil
		}
	}

	return current
}
