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

// PathParser is an optional interface for plugins that need filesystem access
// rather than a single in-memory spec — for example to resolve imports between
// files, or to parse a whole directory of specs. When a plugin implements it,
// the generator calls ParsePath with the configured spec path (file or dir)
// instead of reading the bytes itself.
type PathParser interface {
	ParsePath(path string) (*ir.IR, error)
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
