package config

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Project.Name != "API Documentation" {
		t.Errorf("Expected default project name 'API Documentation', got '%s'", cfg.Project.Name)
	}

	if cfg.Output.Dir != "./docs" {
		t.Errorf("Expected default output dir './docs', got '%s'", cfg.Output.Dir)
	}

	if cfg.Output.Format != "html" {
		t.Errorf("Expected default format 'html', got '%s'", cfg.Output.Format)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}

	if !cfg.Features.Search {
		t.Error("Expected search to be enabled by default")
	}

	if !cfg.Features.DarkMode {
		t.Error("Expected dark mode to be enabled by default")
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := &Config{}

	err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Check defaults were set
	if cfg.Project.Name == "" {
		t.Error("Expected project name to be set")
	}

	if cfg.Output.Dir == "" {
		t.Error("Expected output dir to be set")
	}

	if cfg.Output.Format == "" {
		t.Error("Expected format to be set")
	}
}

func TestConfigProtocols(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Protocols = []ProtocolConfig{
		{
			Plugin:  "openapi",
			Spec:    "./api.yaml",
			Enabled: true,
		},
		{
			Plugin:  "grpc",
			Spec:    "./service.proto",
			Enabled: false,
		},
	}

	if len(cfg.Protocols) != 2 {
		t.Errorf("Expected 2 protocols, got %d", len(cfg.Protocols))
	}

	if cfg.Protocols[0].Plugin != "openapi" {
		t.Errorf("Expected first plugin to be 'openapi', got '%s'", cfg.Protocols[0].Plugin)
	}

	if cfg.Protocols[1].Enabled {
		t.Error("Expected second protocol to be disabled")
	}
}
