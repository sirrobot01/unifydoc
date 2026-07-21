package webhook

import (
	"fmt"

	"github.com/sirrobot01/unifydoc/internal/ir"
	"gopkg.in/yaml.v3"
)

// Plugin implements the Webhook plugin
type Plugin struct{}

// NewPlugin creates a new Webhook plugin
func NewPlugin() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "webhook"
}

// Version returns the plugin version
func (p *Plugin) Version() string {
	return "1.0.0"
}

// WebhookSpec represents the webhook specification
type WebhookSpec struct {
	Webhooks []Webhook `yaml:"webhooks" json:"webhooks"`
}

// Webhook represents a webhook event
type Webhook struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description" json:"description"`
	Method      string                 `yaml:"method" json:"method"`
	URL         string                 `yaml:"url" json:"url"`
	Headers     []Header               `yaml:"headers,omitempty" json:"headers,omitempty"`
	Payload     map[string]interface{} `yaml:"payload,omitempty" json:"payload,omitempty"`
	Example     interface{}            `yaml:"example,omitempty" json:"example,omitempty"`
}

// Header represents an HTTP header
type Header struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Required    bool   `yaml:"required" json:"required"`
}

// Validate validates a webhook specification
func (p *Plugin) Validate(spec []byte) error {
	var webhookSpec WebhookSpec
	if err := yaml.Unmarshal(spec, &webhookSpec); err != nil {
		return fmt.Errorf("invalid webhook YAML: %w", err)
	}

	for i, webhook := range webhookSpec.Webhooks {
		if webhook.Name == "" {
			return fmt.Errorf("webhook %d: name is required", i)
		}
		if webhook.Method == "" {
			return fmt.Errorf("webhook %s: method is required", webhook.Name)
		}
	}

	return nil
}

// Parse parses a webhook specification and converts it to IR
func (p *Plugin) Parse(spec []byte) (*ir.IR, error) {
	var webhookSpec WebhookSpec
	if err := yaml.Unmarshal(spec, &webhookSpec); err != nil {
		return nil, fmt.Errorf("failed to parse webhook spec: %w", err)
	}

	result := ir.NewIR("webhook")
	result.Title = "Webhooks"
	result.Description = "Webhook event definitions"
	result.Version = "1.0.0"

	// Parse webhooks as resources
	for _, webhook := range webhookSpec.Webhooks {
		resource := ir.Resource{
			Name:        webhook.Name,
			Path:        webhook.URL,
			Method:      webhook.Method,
			Description: webhook.Description,
			Parameters:  make([]ir.Parameter, 0),
			Examples:    make([]ir.Example, 0),
			Metadata:    make(map[string]interface{}),
		}

		resource.Metadata["type"] = "webhook"

		// Parse headers as parameters
		for _, header := range webhook.Headers {
			param := ir.Parameter{
				Name:        header.Name,
				In:          "header",
				Description: header.Description,
				Required:    header.Required,
				Type:        "string",
			}
			resource.Parameters = append(resource.Parameters, param)
		}

		// Parse payload
		if webhook.Payload != nil {
			resource.Request = convertPayloadToSchema(webhook.Payload)
		}

		// Add example
		if webhook.Example != nil {
			resource.Examples = append(resource.Examples, ir.Example{
				Name:  "Webhook Payload Example",
				Value: webhook.Example,
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

	return schema
}

// GetTemplate returns the custom template
func (p *Plugin) GetTemplate() string {
	return ""
}
