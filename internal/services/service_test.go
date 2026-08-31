package services

import (
	"sort"
	"testing"
)

func TestGet_ValidService(t *testing.T) {
	cfg, ok := Get("mongodbflex", "instance")
	if !ok {
		t.Fatal("expected mongodbflex instance to be found")
	}
	if cfg.Name != "mongodbflex" {
		t.Errorf("expected Name mongodbflex, got %s", cfg.Name)
	}
	if cfg.ResourceGroup != "instance" {
		t.Errorf("expected ResourceGroup instance, got %s", cfg.ResourceGroup)
	}
	if cfg.ACLJSONPath != "acl.items" {
		t.Errorf("expected ACLJSONPath acl.items, got %s", cfg.ACLJSONPath)
	}
	if cfg.ACLType != ACLArray {
		t.Errorf("expected ACLType array, got %s", cfg.ACLType)
	}
	if cfg.UpdateStrategy != FlagStrategy {
		t.Errorf("expected UpdateStrategy flag, got %s", cfg.UpdateStrategy)
	}
}

func TestGet_SKE(t *testing.T) {
	cfg, ok := Get("ske", "cluster")
	if !ok {
		t.Fatal("expected ske cluster to be found")
	}
	if cfg.ACLJSONPath != "extensions.acl.allowedCidrs" {
		t.Errorf("expected ACLJSONPath extensions.acl.allowedCidrs, got %s", cfg.ACLJSONPath)
	}
	if cfg.UpdateStrategy != PayloadStrategy {
		t.Errorf("expected UpdateStrategy payload, got %s", cfg.UpdateStrategy)
	}
}

func TestGet_UnknownService(t *testing.T) {
	_, ok := Get("unknown", "instance")
	if ok {
		t.Fatal("expected unknown service to not be found")
	}
}

func TestGet_WrongResourceGroup(t *testing.T) {
	_, ok := Get("mongodbflex", "cluster")
	if ok {
		t.Fatal("expected mongodbflex cluster to not be found")
	}
}

func TestGet_BrokerService(t *testing.T) {
	cfg, ok := Get("redis", "instance")
	if !ok {
		t.Fatal("expected redis instance to be found")
	}
	if cfg.ACLJSONPath != "parameters.sgw_acl" {
		t.Errorf("expected ACLJSONPath parameters.sgw_acl, got %s", cfg.ACLJSONPath)
	}
	if cfg.ACLType != ACLCommaString {
		t.Errorf("expected ACLType comma_string, got %s", cfg.ACLType)
	}
}

func TestSupported(t *testing.T) {
	supported := Supported()
	if len(supported) != 10 {
		t.Errorf("expected 10 supported services, got %d", len(supported))
	}

	sort.Strings(supported)
	for _, s := range supported {
		found := false
		for _, expected := range []string{
			"mongodbflex instance", "postgresflex instance", "sqlserverflex instance",
			"redis instance", "valkey instance", "opensearch instance",
			"rabbitmq instance", "mariadb instance", "logme instance",
			"ske cluster",
		} {
			if s == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("unexpected service in supported list: %s", s)
		}
	}
}
