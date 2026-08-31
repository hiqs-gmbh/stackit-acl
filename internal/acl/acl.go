package acl

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"stackit-acl/internal/services"
)

func ToCIDR(ip string, prefix int) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return fmt.Sprintf("%s/%d", ip, prefix)
	}

	var bits int
	if parsed.To4() != nil {
		bits = 32
	} else {
		bits = 128
	}

	mask := net.CIDRMask(prefix, bits)
	masked := parsed.Mask(mask)
	return fmt.Sprintf("%s/%d", masked, prefix)
}

func Contains(list []string, cidr string) bool {
	for _, c := range list {
		if c == cidr {
			return true
		}
	}
	return false
}

func AppendCIDR(existing []string, cidr string) []string {
	if Contains(existing, cidr) {
		return existing
	}
	return append(existing, cidr)
}

func RemoveCIDR(existing []string, cidr string) []string {
	result := make([]string, 0, len(existing))
	for _, c := range existing {
		if c != cidr {
			result = append(result, c)
		}
	}
	return result
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

func ExtractName(jsonData []byte, cfg services.ServiceConfig) string {
	if cfg.NameJSONPath == "" {
		return ""
	}
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return ""
	}
	val := navigatePath(data, cfg.NameJSONPath)
	if val == nil {
		return ""
	}
	s, ok := val.(string)
	if !ok {
		return ""
	}
	return s
}
