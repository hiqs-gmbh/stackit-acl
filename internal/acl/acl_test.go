package acl

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hiqs-gmbh/stackit-acl/internal/services"
)

func TestToCIDR(t *testing.T) {
	tests := []struct {
		ip     string
		prefix int
		want   string
	}{
		{"1.2.3.4", 32, "1.2.3.4/32"},
		{"1.2.3.4", 24, "1.2.3.0/24"},
		{"1.2.3.4", 16, "1.2.0.0/16"},
		{"1.2.3.4", 8, "1.0.0.0/8"},
		{"10.0.0.1", 8, "10.0.0.0/8"},
		{"::1", 128, "::1/128"},
		{"::1", 64, "::/64"},
	}
	for _, tt := range tests {
		got := ToCIDR(tt.ip, tt.prefix)
		if got != tt.want {
			t.Errorf("ToCIDR(%s, %d) = %s, want %s", tt.ip, tt.prefix, got, tt.want)
		}
	}
}

func TestAppendCIDR_New(t *testing.T) {
	existing := []string{"1.2.3.0/24", "5.6.7.0/24"}
	result := AppendCIDR(existing, "9.10.11.12/32")
	if len(result) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result))
	}
	if result[2] != "9.10.11.12/32" {
		t.Errorf("expected 9.10.11.12/32 at index 2, got %s", result[2])
	}
}

func TestAppendCIDR_Duplicate(t *testing.T) {
	existing := []string{"1.2.3.0/24", "9.10.11.12/32"}
	result := AppendCIDR(existing, "9.10.11.12/32")
	if len(result) != 2 {
		t.Fatalf("expected 2 entries (no duplicate), got %d", len(result))
	}
}

func TestAppendCIDR_EmptyList(t *testing.T) {
	result := AppendCIDR([]string{}, "9.10.11.12/32")
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if result[0] != "9.10.11.12/32" {
		t.Errorf("expected 9.10.11.12/32, got %s", result[0])
	}
}

func TestAppendCIDR_NilList(t *testing.T) {
	result := AppendCIDR(nil, "9.10.11.12/32")
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
}

func TestContains_True(t *testing.T) {
	list := []string{"1.2.3.0/24", "9.10.11.12/32"}
	if !Contains(list, "9.10.11.12/32") {
		t.Error("expected Contains to return true")
	}
}

func TestContains_False(t *testing.T) {
	list := []string{"1.2.3.0/24", "5.6.7.0/24"}
	if Contains(list, "9.10.11.12/32") {
		t.Error("expected Contains to return false")
	}
}

func TestContains_EmptyList(t *testing.T) {
	if Contains([]string{}, "9.10.11.12/32") {
		t.Error("expected Contains to return false for empty list")
	}
}

func TestContains_NilList(t *testing.T) {
	if Contains(nil, "9.10.11.12/32") {
		t.Error("expected Contains to return false for nil list")
	}
}

func TestRemoveCIDR_Existing(t *testing.T) {
	existing := []string{"1.2.3.0/24", "9.10.11.12/32", "5.6.7.0/24"}
	result := RemoveCIDR(existing, "9.10.11.12/32")
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result[0] != "1.2.3.0/24" || result[1] != "5.6.7.0/24" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestRemoveCIDR_NotPresent(t *testing.T) {
	existing := []string{"1.2.3.0/24", "5.6.7.0/24"}
	result := RemoveCIDR(existing, "9.10.11.12/32")
	if len(result) != 2 {
		t.Fatalf("expected 2 entries (unchanged), got %d", len(result))
	}
}

func TestRemoveCIDR_EmptyList(t *testing.T) {
	result := RemoveCIDR([]string{}, "9.10.11.12/32")
	if len(result) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result))
	}
}

func TestRemoveCIDR_NilList(t *testing.T) {
	result := RemoveCIDR(nil, "9.10.11.12/32")
	if len(result) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result))
	}
}

func TestRemoveCIDR_LastEntry(t *testing.T) {
	existing := []string{"9.10.11.12/32"}
	result := RemoveCIDR(existing, "9.10.11.12/32")
	if len(result) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(result))
	}
}

func TestRemoveCIDR_MultipleDuplicates(t *testing.T) {
	existing := []string{"9.10.11.12/32", "1.2.3.0/24", "9.10.11.12/32"}
	result := RemoveCIDR(existing, "9.10.11.12/32")
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if result[0] != "1.2.3.0/24" {
		t.Errorf("expected 1.2.3.0/24, got %s", result[0])
	}
}

func TestExtractACLs_ArrayType(t *testing.T) {
	cfg := services.ServiceConfig{
		ACLJSONPath: "acl.items",
		ACLType:     services.ACLArray,
	}
	jsonData := []byte(`{"acl": {"items": ["1.2.3.0/24", "5.6.7.0/24"]}}`)

	acls, err := ExtractACLs(jsonData, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(acls) != 2 {
		t.Fatalf("expected 2 ACLs, got %d", len(acls))
	}
	if acls[0] != "1.2.3.0/24" || acls[1] != "5.6.7.0/24" {
		t.Errorf("unexpected ACLs: %v", acls)
	}
}

func TestExtractACLs_ArrayType_NestedPath(t *testing.T) {
	cfg := services.ServiceConfig{
		ACLJSONPath: "network.acl",
		ACLType:     services.ACLArray,
	}
	jsonData := []byte(`{"network": {"acl": ["10.0.0.0/8"]}}`)

	acls, err := ExtractACLs(jsonData, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(acls) != 1 || acls[0] != "10.0.0.0/8" {
		t.Errorf("unexpected ACLs: %v", acls)
	}
}

func TestExtractACLs_ArrayType_Empty(t *testing.T) {
	cfg := services.ServiceConfig{
		ACLJSONPath: "acl.items",
		ACLType:     services.ACLArray,
	}
	jsonData := []byte(`{"acl": {"items": []}}`)

	acls, err := ExtractACLs(jsonData, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(acls) != 0 {
		t.Errorf("expected 0 ACLs, got %d", len(acls))
	}
}

func TestExtractACLs_ArrayType_MissingPath(t *testing.T) {
	cfg := services.ServiceConfig{
		ACLJSONPath: "acl.items",
		ACLType:     services.ACLArray,
	}
	jsonData := []byte(`{"name": "test"}`)

	acls, err := ExtractACLs(jsonData, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(acls) != 0 {
		t.Errorf("expected 0 ACLs for missing path, got %d", len(acls))
	}
}

func TestExtractACLs_CommaStringType(t *testing.T) {
	cfg := services.ServiceConfig{
		ACLJSONPath: "parameters.sgw_acl",
		ACLType:     services.ACLCommaString,
	}
	jsonData := []byte(`{"parameters": {"sgw_acl": "1.2.3.0/24,5.6.7.0/24"}}`)

	acls, err := ExtractACLs(jsonData, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(acls) != 2 {
		t.Fatalf("expected 2 ACLs, got %d", len(acls))
	}
	if acls[0] != "1.2.3.0/24" || acls[1] != "5.6.7.0/24" {
		t.Errorf("unexpected ACLs: %v", acls)
	}
}

func TestExtractACLs_CommaStringType_Empty(t *testing.T) {
	cfg := services.ServiceConfig{
		ACLJSONPath: "parameters.sgw_acl",
		ACLType:     services.ACLCommaString,
	}
	jsonData := []byte(`{"parameters": {"sgw_acl": ""}}`)

	acls, err := ExtractACLs(jsonData, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(acls) != 0 {
		t.Errorf("expected 0 ACLs, got %d", len(acls))
	}
}

func TestExtractACLs_CommaStringType_MissingPath(t *testing.T) {
	cfg := services.ServiceConfig{
		ACLJSONPath: "parameters.sgw_acl",
		ACLType:     services.ACLCommaString,
	}
	jsonData := []byte(`{"name": "test"}`)

	acls, err := ExtractACLs(jsonData, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(acls) != 0 {
		t.Errorf("expected 0 ACLs for missing path, got %d", len(acls))
	}
}

func TestExtractACLs_CommaStringType_WithSpaces(t *testing.T) {
	cfg := services.ServiceConfig{
		ACLJSONPath: "parameters.sgw_acl",
		ACLType:     services.ACLCommaString,
	}
	jsonData := []byte(`{"parameters": {"sgw_acl": "1.2.3.0/24, 5.6.7.0/24"}}`)

	acls, err := ExtractACLs(jsonData, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(acls) != 2 {
		t.Fatalf("expected 2 ACLs, got %d", len(acls))
	}
	if acls[0] != "1.2.3.0/24" || acls[1] != "5.6.7.0/24" {
		t.Errorf("unexpected ACLs: %v", acls)
	}
}

func TestExtractACLs_SKE_DeepNested(t *testing.T) {
	cfg := services.ServiceConfig{
		ACLJSONPath: "extensions.acl.allowedCidrs",
		ACLType:     services.ACLArray,
	}
	jsonData := []byte(`{"extensions": {"acl": {"allowedCidrs": ["1.2.3.4/32", "10.0.0.0/8"]}}}`)

	acls, err := ExtractACLs(jsonData, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(acls) != 2 {
		t.Fatalf("expected 2 ACLs, got %d", len(acls))
	}
}

func TestExtractACLs_InvalidJSON(t *testing.T) {
	cfg := services.ServiceConfig{
		ACLJSONPath: "acl.items",
		ACLType:     services.ACLArray,
	}
	jsonData := []byte(`not json`)

	_, err := ExtractACLs(jsonData, cfg)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestSetACLs(t *testing.T) {
	cfg := services.ServiceConfig{
		ACLJSONPath: "extensions.acl.allowedCidrs",
		ACLType:     services.ACLArray,
	}
	jsonData := []byte(`{"extensions": {"acl": {"allowedCidrs": ["1.2.3.4/32"]}}, "kubernetes": {"version": "1.29"}}`)

	updated, err := SetACLs(jsonData, cfg, []string{"1.2.3.4/32", "5.6.7.8/32"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(updated, &result); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	cidrs := navigatePath(result, "extensions.acl.allowedCidrs")
	arr, ok := cidrs.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", cidrs)
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 CIDRs, got %d", len(arr))
	}

	version := navigatePath(result, "kubernetes.version")
	if version != "1.29" {
		t.Errorf("expected kubernetes.version to be preserved as 1.29, got %v", version)
	}
}

func TestSetACLs_CreatesMissingPath(t *testing.T) {
	cfg := services.ServiceConfig{
		ACLJSONPath: "extensions.acl.allowedCidrs",
		ACLType:     services.ACLArray,
	}
	jsonData := []byte(`{"kubernetes": {"version": "1.29"}}`)

	updated, err := SetACLs(jsonData, cfg, []string{"1.2.3.4/32"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(updated, &result); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	cidrs := navigatePath(result, "extensions.acl.allowedCidrs")
	arr, ok := cidrs.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", cidrs)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 CIDR, got %d", len(arr))
	}
}

func TestSetACLs_PreservesOtherFields(t *testing.T) {
	cfg := services.ServiceConfig{
		ACLJSONPath: "extensions.acl.allowedCidrs",
		ACLType:     services.ACLArray,
	}
	jsonData := []byte(`{"extensions": {"acl": {"allowedCidrs": ["1.2.3.4/32"]}, "hibernation": {"enabled": true}}, "kubernetes": {"version": "1.29"}}`)

	updated, err := SetACLs(jsonData, cfg, []string{"1.2.3.4/32", "5.6.7.8/32"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(string(updated), "hibernation") {
		t.Error("expected hibernation field to be preserved")
	}
	if !strings.Contains(string(updated), "1.29") {
		t.Error("expected kubernetes.version to be preserved")
	}
}
