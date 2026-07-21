package events

import (
	"fmt"

	"github.com/sirrobot01/unifydoc/internal/ir"
	"gopkg.in/yaml.v3"
)

// Plugin implements the Events plugin
type Plugin struct{}

// NewPlugin creates a new Events plugin
func NewPlugin() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "events"
}

// Version returns the plugin version
func (p *Plugin) Version() string {
	return "1.0.0"
}

// Spec represents the events specification
type Spec struct {
	Events Config `yaml:"events" json:"events"`
}

// Config represents events configuration
type Config struct {
	Title       string  `yaml:"title,omitempty" json:"title,omitempty"`
	Description string  `yaml:"description,omitempty" json:"description,omitempty"`
	Version     string  `yaml:"version,omitempty" json:"version,omitempty"`
	Topics      []Topic `yaml:"topics,omitempty" json:"topics,omitempty"`
	EventTypes  []Event `yaml:"eventTypes,omitempty" json:"eventTypes,omitempty"`
}

// Topic represents an event topic
type Topic struct {
	Name        string  `yaml:"name" json:"name"`
	Description string  `yaml:"description" json:"description"`
	Events      []Event `yaml:"events,omitempty" json:"events,omitempty"`
}

// Event represents an event type
type Event struct {
	Name        string                 `yaml:"name" json:"name"`
	Type        string                 `yaml:"type" json:"type"`
	Description string                 `yaml:"description" json:"description"`
	Schema      map[string]interface{} `yaml:"schema,omitempty" json:"schema,omitempty"`
	Example     interface{}            `yaml:"example,omitempty" json:"example,omitempty"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// Validate validates an events specification
func (p *Plugin) Validate(spec []byte) error {
	var eventsSpec Spec
	if err := yaml.Unmarshal(spec, &eventsSpec); err != nil {
		return fmt.Errorf("invalid events YAML: %w", err)
	}

	return nil
}

// Parse parses an events specification and converts it to IR
func (p *Plugin) Parse(spec []byte) (*ir.IR, error) {
	var eventsSpec Spec
	if err := yaml.Unmarshal(spec, &eventsSpec); err != nil {
		return nil, fmt.Errorf("failed to parse events spec: %w", err)
	}

	result := ir.NewIR("events")
	result.Title = eventsSpec.Events.Title
	if result.Title == "" {
		result.Title = "Event System"
	}
	result.Description = eventsSpec.Events.Description
	result.Version = eventsSpec.Events.Version
	if result.Version == "" {
		result.Version = "1.0.0"
	}

	// Parse topics
	for _, topic := range eventsSpec.Events.Topics {
		for _, event := range topic.Events {
			resource := p.convertEventToResource(event, topic.Name)
			result.Resources = append(result.Resources, resource)
		}
	}

	// Parse standalone event types
	for _, event := range eventsSpec.Events.EventTypes {
		resource := p.convertEventToResource(event, "")
		result.Resources = append(result.Resources, resource)
	}

	return result, nil
}

// convertEventToResource converts an Event to IR Resource
func (p *Plugin) convertEventToResource(event Event, topicName string) ir.Resource {
	resource := ir.Resource{
		Name:        event.Name,
		Path:        event.Type,
		Method:      "EVENT",
		Description: event.Description,
		Examples:    make([]ir.Example, 0),
		Metadata:    event.Metadata,
	}

	if resource.Metadata == nil {
		resource.Metadata = make(map[string]interface{})
	}
	resource.Metadata["type"] = "event"
	resource.Metadata["event_type"] = event.Type

	if topicName != "" {
		resource.Metadata["topic"] = topicName
		resource.Tags = []string{topicName}
	}

	// Parse schema
	if event.Schema != nil {
		resource.Request = convertMapToSchema(event.Schema)
	}

	// Add example
	if event.Example != nil {
		resource.Examples = append(resource.Examples, ir.Example{
			Name:  "Event Example",
			Value: event.Example,
		})
	}

	return resource
}

// convertMapToSchema converts a map to IR Schema
func convertMapToSchema(data map[string]interface{}) *ir.Schema {
	if data == nil {
		return nil
	}

	schema := &ir.Schema{
		Type:       "object",
		Properties: make(map[string]*ir.Schema),
		Metadata:   make(map[string]interface{}),
	}

	if typ, ok := data["type"].(string); ok {
		schema.Type = typ
	}

	if desc, ok := data["description"].(string); ok {
		schema.Description = desc
	}

	if props, ok := data["properties"].(map[string]interface{}); ok {
		for name, propData := range props {
			if propMap, ok := propData.(map[string]interface{}); ok {
				schema.Properties[name] = convertMapToSchema(propMap)
			}
		}
	}

	if required, ok := data["required"].([]interface{}); ok {
		schema.Required = make([]string, 0)
		for _, r := range required {
			if rStr, ok := r.(string); ok {
				schema.Required = append(schema.Required, rStr)
			}
		}
	}

	if items, ok := data["items"].(map[string]interface{}); ok {
		schema.Items = convertMapToSchema(items)
	}

	return schema
}

// GetTemplate returns the custom template
func (p *Plugin) GetTemplate() string {
	return ""
}
