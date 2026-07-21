package websocket

import (
	"fmt"

	"github.com/sirrobot01/unifydoc/internal/ir"
	"gopkg.in/yaml.v3"
)

// Plugin implements the WebSocket plugin
type Plugin struct{}

// NewPlugin creates a new WebSocket plugin
func NewPlugin() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "websocket"
}

// Version returns the plugin version
func (p *Plugin) Version() string {
	return "1.0.0"
}

// Spec represents the WebSocket specification format
type Spec struct {
	WebSocket Config `yaml:"websocket" json:"websocket"`
}

// Config represents WebSocket configuration
type Config struct {
	URL         string  `yaml:"url" json:"url"`
	Description string  `yaml:"description" json:"description"`
	Version     string  `yaml:"version,omitempty" json:"version,omitempty"`
	Events      []Event `yaml:"events" json:"events"`
}

// Event represents a WebSocket event
type Event struct {
	Name        string                 `yaml:"name" json:"name"`
	Direction   string                 `yaml:"direction" json:"direction"` // send, receive, bidirectional
	Description string                 `yaml:"description" json:"description"`
	Payload     map[string]interface{} `yaml:"payload,omitempty" json:"payload,omitempty"`
	Example     interface{}            `yaml:"example,omitempty" json:"example,omitempty"`
}

// Validate validates a WebSocket specification
func (p *Plugin) Validate(spec []byte) error {
	var wsSpec Spec
	if err := yaml.Unmarshal(spec, &wsSpec); err != nil {
		return fmt.Errorf("invalid WebSocket YAML: %w", err)
	}

	if wsSpec.WebSocket.URL == "" {
		return fmt.Errorf("WebSocket URL is required")
	}

	for i, event := range wsSpec.WebSocket.Events {
		if event.Name == "" {
			return fmt.Errorf("event %d: name is required", i)
		}
		if event.Direction != "send" && event.Direction != "receive" && event.Direction != "bidirectional" {
			return fmt.Errorf("event %s: invalid direction '%s', must be 'send', 'receive', or 'bidirectional'", event.Name, event.Direction)
		}
	}

	return nil
}

// Parse parses a WebSocket specification and converts it to IR
func (p *Plugin) Parse(spec []byte) (*ir.IR, error) {
	var wsSpec Spec
	if err := yaml.Unmarshal(spec, &wsSpec); err != nil {
		return nil, fmt.Errorf("failed to parse WebSocket spec: %w", err)
	}

	result := ir.NewIR("websocket")
	result.Title = "WebSocket API"
	result.Description = wsSpec.WebSocket.Description
	result.Version = wsSpec.WebSocket.Version

	if result.Version == "" {
		result.Version = "1.0.0"
	}

	// Add server
	result.Servers = append(result.Servers, ir.Server{
		URL:         wsSpec.WebSocket.URL,
		Description: wsSpec.WebSocket.Description,
	})

	// Parse events as resources
	for _, event := range wsSpec.WebSocket.Events {
		resource := ir.Resource{
			Name:        event.Name,
			Path:        wsSpec.WebSocket.URL,
			Method:      event.Direction,
			Description: event.Description,
			Metadata:    make(map[string]interface{}),
		}

		// Store direction in metadata
		resource.Metadata["direction"] = event.Direction
		resource.Metadata["type"] = "websocket-event"

		// Parse payload as request schema
		if event.Payload != nil {
			schema := convertPayloadToSchema(event.Payload)
			resource.Request = schema
		}

		// Add example
		if event.Example != nil {
			resource.Examples = append(resource.Examples, ir.Example{
				Name:  "Event Example",
				Value: event.Example,
			})
		}

		result.Resources = append(result.Resources, resource)
	}

	return result, nil
}

// convertPayloadToSchema converts a payload map to IR schema
func convertPayloadToSchema(payload map[string]interface{}) *ir.Schema {
	schema := &ir.Schema{
		Type:       "object",
		Properties: make(map[string]*ir.Schema),
		Metadata:   make(map[string]interface{}),
	}

	if typ, ok := payload["type"].(string); ok {
		schema.Type = typ
	}

	if desc, ok := payload["description"].(string); ok {
		schema.Description = desc
	}

	if props, ok := payload["properties"].(map[string]interface{}); ok {
		for name, propData := range props {
			if propMap, ok := propData.(map[string]interface{}); ok {
				schema.Properties[name] = convertPayloadToSchema(propMap)
			}
		}
	}

	if required, ok := payload["required"].([]interface{}); ok {
		schema.Required = make([]string, 0)
		for _, r := range required {
			if rStr, ok := r.(string); ok {
				schema.Required = append(schema.Required, rStr)
			}
		}
	}

	if items, ok := payload["items"].(map[string]interface{}); ok {
		schema.Items = convertPayloadToSchema(items)
	}

	if example, ok := payload["example"]; ok {
		schema.Example = example
	}

	return schema
}

// GetTemplate returns the custom template
func (p *Plugin) GetTemplate() string {
	return ""
}
