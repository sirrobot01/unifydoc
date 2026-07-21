package plugin

import "github.com/sirrobot01/unifydoc/internal/ir"

// Plugin is the interface that all protocol plugins must implement
type Plugin interface {
	// Name returns the plugin name (e.g., 'openapi', 'grpc')
	Name() string

	// Version returns the plugin version
	Version() string

	// Parse parses a specification and converts it to IR
	Parse(spec []byte) (*ir.IR, error)

	// Validate validates a specification without full parsing
	Validate(spec []byte) error

	// GetTemplate returns the custom template for rendering (optional)
	GetTemplate() string
}

// Metadata represents plugin metadata
type Metadata struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author,omitempty"`
	Protocols   []string `json:"protocols"`
	BuiltIn     bool     `json:"built_in"`
}

// Registry is an interface for plugin registration
type Registry interface {
	Register(plugin Plugin) error
	Get(name string) (Plugin, error)
	List() []Metadata
}
