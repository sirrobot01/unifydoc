package plugin

import (
	"fmt"
	"os"
	"path/filepath"
)

// DiscoveryPath represents a plugin discovery location
type DiscoveryPath struct {
	Path     string
	Priority int
}

// Discoverer handles plugin discovery
type Discoverer struct {
	paths []DiscoveryPath
}

// NewDiscoverer creates a new plugin discoverer
func NewDiscoverer() *Discoverer {
	return &Discoverer{
		paths: getDefaultDiscoveryPaths(),
	}
}

// getDefaultDiscoveryPaths returns the default plugin discovery paths
func getDefaultDiscoveryPaths() []DiscoveryPath {
	paths := make([]DiscoveryPath, 0)

	// Project plugins (highest priority)
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, DiscoveryPath{
			Path:     filepath.Join(cwd, "plugins"),
			Priority: 1,
		})
	}

	// User plugins
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, DiscoveryPath{
			Path:     filepath.Join(home, ".config", "unifidoc", "plugins"),
			Priority: 2,
		})
	}

	// System plugins (lowest priority)
	paths = append(paths, DiscoveryPath{
		Path:     "/etc/unifidoc/plugins",
		Priority: 3,
	})

	return paths
}

// AddPath adds a custom discovery path
func (d *Discoverer) AddPath(path string, priority int) {
	d.paths = append(d.paths, DiscoveryPath{
		Path:     path,
		Priority: priority,
	})
}

// Discover searches for plugins in all discovery paths
func (d *Discoverer) Discover() ([]string, error) {
	discovered := make([]string, 0)

	for _, dp := range d.paths {
		if _, err := os.Stat(dp.Path); os.IsNotExist(err) {
			continue
		}

		files, err := os.ReadDir(dp.Path)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			// Check for .so files (Go plugins)
			if filepath.Ext(file.Name()) == ".so" {
				discovered = append(discovered, filepath.Join(dp.Path, file.Name()))
			}
		}
	}

	return discovered, nil
}

// LoadPlugin loads a plugin from a file (placeholder for future external plugin support)
func (d *Discoverer) LoadPlugin(path string) (Plugin, error) {
	// This is a placeholder for loading external plugins
	// For now, all plugins are built-in and registered during initialization
	return nil, fmt.Errorf("external plugin loading not yet implemented")
}
