package asyncapi

import (
	"fmt"

	"github.com/sirrobot01/unifydoc/internal/ir"
	"gopkg.in/yaml.v3"
)

// Plugin implements the AsyncAPI plugin
type Plugin struct{}

// NewPlugin creates a new AsyncAPI plugin
func NewPlugin() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "asyncapi"
}

// Version returns the plugin version
func (p *Plugin) Version() string {
	return "1.0.0"
}

// Spec represents a simplified AsyncAPI specification
type Spec struct {
	AsyncAPI   string                 `yaml:"asyncapi" json:"asyncapi"`
	Info       Info                   `yaml:"info" json:"info"`
	Servers    map[string]Server      `yaml:"servers,omitempty" json:"servers,omitempty"`
	Channels   map[string]Channel     `yaml:"channels" json:"channels"`
	Components map[string]interface{} `yaml:"components,omitempty" json:"components,omitempty"`
}

// Info represents API information
type Info struct {
	Title       string `yaml:"title" json:"title"`
	Version     string `yaml:"version" json:"version"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// Server represents a server
type Server struct {
	URL         string `yaml:"url" json:"url"`
	Protocol    string `yaml:"protocol" json:"protocol"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// Channel represents a channel
type Channel struct {
	Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Subscribe   *Operation             `yaml:"subscribe,omitempty" json:"subscribe,omitempty"`
	Publish     *Operation             `yaml:"publish,omitempty" json:"publish,omitempty"`
	Parameters  map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

// Operation represents a subscribe or publish operation
type Operation struct {
	Summary     string                 `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Message     map[string]interface{} `yaml:"message,omitempty" json:"message,omitempty"`
	Tags        []Tag                  `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// Tag represents a tag
type Tag struct {
	Name string `yaml:"name" json:"name"`
}

// Validate validates an AsyncAPI specification
func (p *Plugin) Validate(spec []byte) error {
	var asyncSpec Spec
	if err := yaml.Unmarshal(spec, &asyncSpec); err != nil {
		return fmt.Errorf("invalid AsyncAPI YAML: %w", err)
	}

	if asyncSpec.AsyncAPI == "" {
		return fmt.Errorf("asyncapi version is required")
	}

	if asyncSpec.Info.Title == "" {
		return fmt.Errorf("info.title is required")
	}

	if asyncSpec.Info.Version == "" {
		return fmt.Errorf("info.version is required")
	}

	return nil
}

// Parse parses an AsyncAPI specification and converts it to IR
func (p *Plugin) Parse(spec []byte) (*ir.IR, error) {
	var asyncSpec Spec
	if err := yaml.Unmarshal(spec, &asyncSpec); err != nil {
		return nil, fmt.Errorf("failed to parse AsyncAPI spec: %w", err)
	}

	result := ir.NewIR("asyncapi")
	result.Title = asyncSpec.Info.Title
	result.Description = asyncSpec.Info.Description
	result.Version = asyncSpec.Info.Version
	result.Metadata["asyncapi_version"] = asyncSpec.AsyncAPI

	// Parse servers
	for name, server := range asyncSpec.Servers {
		result.Servers = append(result.Servers, ir.Server{
			URL:         server.URL,
			Description: fmt.Sprintf("%s (%s)", server.Description, server.Protocol),
			Variables:   map[string]string{"protocol": server.Protocol},
		})
		result.Metadata[fmt.Sprintf("server_%s", name)] = server.Protocol
	}

	// Parse channels
	for channelName, channel := range asyncSpec.Channels {
		// Parse subscribe operation
		if channel.Subscribe != nil {
			resource := p.parseOperation(channelName, "subscribe", channel.Subscribe, channel.Description)
			result.Resources = append(result.Resources, resource)
		}

		// Parse publish operation
		if channel.Publish != nil {
			resource := p.parseOperation(channelName, "publish", channel.Publish, channel.Description)
			result.Resources = append(result.Resources, resource)
		}
	}

	return result, nil
}

// parseOperation parses a subscribe or publish operation
func (p *Plugin) parseOperation(channelName, opType string, op *Operation, channelDesc string) ir.Resource {
	resource := ir.Resource{
		Name:        op.Summary,
		Path:        channelName,
		Method:      opType,
		Description: op.Description,
		Metadata:    make(map[string]interface{}),
		Examples:    make([]ir.Example, 0),
		Tags:        make([]string, 0),
	}

	if resource.Name == "" {
		resource.Name = fmt.Sprintf("%s %s", opType, channelName)
	}

	if resource.Description == "" {
		resource.Description = channelDesc
	}

	resource.Metadata["operation_type"] = opType
	resource.Metadata["channel"] = channelName

	// Parse tags
	for _, tag := range op.Tags {
		resource.Tags = append(resource.Tags, tag.Name)
	}

	// Parse message
	if op.Message != nil {
		if payload, ok := op.Message["payload"].(map[string]interface{}); ok {
			resource.Request = convertMapToSchema(payload)
		}

		if name, ok := op.Message["name"].(string); ok {
			resource.Metadata["message_name"] = name
		}

		if contentType, ok := op.Message["contentType"].(string); ok {
			resource.Metadata["content_type"] = contentType
		}
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
