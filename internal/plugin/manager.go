package plugin

import (
	"fmt"
	"sync"
)

// Manager manages plugin lifecycle and discovery
type Manager struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

// NewManager creates a new plugin manager
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
	}
}

// Register registers a plugin
func (m *Manager) Register(plugin Plugin) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin '%s' already registered", name)
	}

	m.plugins[name] = plugin
	return nil
}

// Get retrieves a plugin by name
func (m *Manager) Get(name string) (Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}

	return plugin, nil
}

// List returns metadata for all registered plugins
func (m *Manager) List() []Metadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]Metadata, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		metadata := Metadata{
			Name:     plugin.Name(),
			Version:  plugin.Version(),
			BuiltIn:  true, // All plugins are built-in for now
			Protocols: []string{plugin.Name()},
		}
		list = append(list, metadata)
	}

	return list
}

// Unregister removes a plugin
func (m *Manager) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[name]; !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	delete(m.plugins, name)
	return nil
}

// Count returns the number of registered plugins
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.plugins)
}

// Has checks if a plugin exists
func (m *Manager) Has(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.plugins[name]
	return exists
}
