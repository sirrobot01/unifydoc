package config

import (
	"os"
	"path/filepath"
	"strings"
)

// AutoDetector handles automatic detection of specification files
type AutoDetector struct {
	searchDirs []string
}

// NewAutoDetector creates a new auto-detector
func NewAutoDetector() *AutoDetector {
	cwd, _ := os.Getwd()
	return &AutoDetector{
		searchDirs: []string{
			cwd,
			filepath.Join(cwd, "docs"),
			filepath.Join(cwd, "api"),
			filepath.Join(cwd, "specs"),
			filepath.Join(cwd, ".unifidoc"),
		},
	}
}

// Detect automatically detects specification files
func (d *AutoDetector) Detect() ([]ProtocolConfig, error) {
	protocols := make([]ProtocolConfig, 0)

	// Detection patterns from PRD
	patterns := map[string][]string{
		"openapi": {
			"openapi.yaml", "openapi.yml", "openapi.json",
			"api.yaml", "api.yml", "api.json",
			"swagger.yaml", "swagger.yml", "swagger.json",
		},
		"asyncapi": {
			"asyncapi.yaml", "asyncapi.yml", "asyncapi.json",
		},
		"grpc": {
			"*.proto",
		},
		"websocket": {
			"websocket.yaml", "websocket.yml",
			"ws.yaml", "ws.yml",
		},
		"webhook": {
			"webhooks.yaml", "webhooks.yml",
			"webhook.yaml", "webhook.yml",
		},
		"events": {
			"events.yaml", "events.yml",
		},
		"mcp": {
			"mcp.yaml", "mcp.yml", "mcp.json",
		},
		"api": {
			"rest.yaml", "rest.yml",
		},
		"custom": {
			"custom.yaml", "custom.yml",
		},
	}

	for plugin, filePatterns := range patterns {
		for _, dir := range d.searchDirs {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				continue
			}

			for _, pattern := range filePatterns {
				// Handle glob patterns
				if strings.Contains(pattern, "*") {
					matches, err := filepath.Glob(filepath.Join(dir, pattern))
					if err != nil {
						continue
					}
					for _, match := range matches {
						protocols = append(protocols, ProtocolConfig{
							Plugin:  plugin,
							Spec:    match,
							Enabled: true,
						})
					}
				} else {
					// Direct file check
					path := filepath.Join(dir, pattern)
					if _, err := os.Stat(path); err == nil {
						protocols = append(protocols, ProtocolConfig{
							Plugin:  plugin,
							Spec:    path,
							Enabled: true,
						})
					}
				}
			}
		}
	}

	return protocols, nil
}

// DetectAndCreateConfig creates a config with auto-detected specs
func DetectAndCreateConfig() (*Config, error) {
	detector := NewAutoDetector()
	protocols, err := detector.Detect()
	if err != nil {
		return nil, err
	}

	config := DefaultConfig()

	if len(protocols) > 0 {
		config.Protocols = protocols
	}

	return config, nil
}
