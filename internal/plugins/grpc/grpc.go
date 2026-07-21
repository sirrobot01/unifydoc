package grpc

import (
	"fmt"
	"strings"

	"github.com/sirrobot01/unifydoc/internal/ir"
)

// Plugin implements the gRPC plugin
type Plugin struct{}

// NewPlugin creates a new gRPC plugin
func NewPlugin() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "grpc"
}

// Version returns the plugin version
func (p *Plugin) Version() string {
	return "1.0.0"
}

// Validate validates a Protocol Buffer specification
func (p *Plugin) Validate(spec []byte) error {
	// Basic validation - check if it looks like a proto file
	content := string(spec)
	if !strings.Contains(content, "syntax") && !strings.Contains(content, "service") {
		return fmt.Errorf("invalid proto file: missing syntax or service declaration")
	}
	return nil
}

// Parse parses a Protocol Buffer specification and converts it to IR
func (p *Plugin) Parse(spec []byte) (*ir.IR, error) {
	// This is a simplified proto parser
	// In production, you'd use google.golang.org/protobuf/proto or a full parser
	content := string(spec)

	result := ir.NewIR("grpc")
	result.Title = "gRPC Service"
	result.Description = "gRPC service definition"
	result.Version = "1.0.0"

	// Extract package name
	packageName := extractPackageName(content)
	if packageName != "" {
		result.Metadata["package"] = packageName
	}

	// Extract services and RPCs
	services := p.parseServices(content)
	for _, service := range services {
		for _, rpc := range service.RPCs {
			resource := ir.Resource{
				Name:        rpc.Name,
				Path:        fmt.Sprintf("/%s/%s", service.Name, rpc.Name),
				Method:      "RPC",
				Description: rpc.Description,
				Metadata:    make(map[string]interface{}),
			}

			// Store streaming info
			resource.Metadata["streaming"] = rpc.Streaming
			resource.Metadata["service"] = service.Name
			resource.Metadata["request_type"] = rpc.RequestType
			resource.Metadata["response_type"] = rpc.ResponseType

			// Create simple request/response schemas
			resource.Request = &ir.Schema{
				Type:        "object",
				Description: fmt.Sprintf("Request message: %s", rpc.RequestType),
				Metadata:    map[string]interface{}{"message_type": rpc.RequestType},
			}

			resource.Response = &ir.Schema{
				Type:        "object",
				Description: fmt.Sprintf("Response message: %s", rpc.ResponseType),
				Metadata:    map[string]interface{}{"message_type": rpc.ResponseType},
			}

			result.Resources = append(result.Resources, resource)
		}
	}

	// Extract messages as types
	messages := p.parseMessages(content)
	for _, msg := range messages {
		typeDef := ir.TypeDef{
			Name:        msg.Name,
			Description: msg.Description,
			Schema:      msg.Schema,
		}
		result.Types = append(result.Types, typeDef)
	}

	return result, nil
}

// Service represents a gRPC service
type Service struct {
	Name        string
	Description string
	RPCs        []RPC
}

// RPC represents a gRPC RPC method
type RPC struct {
	Name         string
	Description  string
	RequestType  string
	ResponseType string
	Streaming    string // unary, client_streaming, server_streaming, bidirectional
}

// Message represents a Protocol Buffer message
type Message struct {
	Name        string
	Description string
	Schema      *ir.Schema
}

// parseServices extracts services from proto content
func (p *Plugin) parseServices(content string) []Service {
	services := make([]Service, 0)
	lines := strings.Split(content, "\n")

	var currentService *Service
	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Detect service start
		if strings.HasPrefix(line, "service ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				serviceName := strings.TrimSuffix(parts[1], "{")
				currentService = &Service{
					Name: serviceName,
					RPCs: make([]RPC, 0),
				}

				// Get description from previous line if it's a comment
				if i > 0 {
					prevLine := strings.TrimSpace(lines[i-1])
					if strings.HasPrefix(prevLine, "//") {
						currentService.Description = strings.TrimPrefix(prevLine, "//")
						currentService.Description = strings.TrimSpace(currentService.Description)
					}
				}
			}
		}

		// Detect RPC
		if currentService != nil && strings.HasPrefix(line, "rpc ") {
			rpc := p.parseRPC(line)
			if rpc != nil {
				// Get description from previous line
				if i > 0 {
					prevLine := strings.TrimSpace(lines[i-1])
					if strings.HasPrefix(prevLine, "//") {
						rpc.Description = strings.TrimPrefix(prevLine, "//")
						rpc.Description = strings.TrimSpace(rpc.Description)
					}
				}
				currentService.RPCs = append(currentService.RPCs, *rpc)
			}
		}

		// Detect service end
		if currentService != nil && strings.HasPrefix(line, "}") {
			services = append(services, *currentService)
			currentService = nil
		}
	}

	return services
}

// parseRPC parses an RPC line
func (p *Plugin) parseRPC(line string) *RPC {
	// Format: rpc GetUser (GetUserRequest) returns (GetUserResponse);
	// or: rpc StreamUsers (stream StreamUsersRequest) returns (stream StreamUsersResponse);

	parts := strings.Fields(line)
	if len(parts) < 5 {
		return nil
	}

	rpc := &RPC{
		Name:      parts[1],
		Streaming: "unary",
	}

	// Parse request
	requestStart := strings.Index(line, "(")
	requestEnd := strings.Index(line, ")")
	if requestStart != -1 && requestEnd != -1 {
		requestPart := line[requestStart+1 : requestEnd]
		requestPart = strings.TrimSpace(requestPart)
		if strings.HasPrefix(requestPart, "stream ") {
			rpc.RequestType = strings.TrimPrefix(requestPart, "stream ")
			rpc.Streaming = "client_streaming"
		} else {
			rpc.RequestType = requestPart
		}
	}

	// Parse response
	returnsIdx := strings.Index(line, "returns")
	if returnsIdx != -1 {
		responsePart := line[returnsIdx+7:]
		responseStart := strings.Index(responsePart, "(")
		responseEnd := strings.Index(responsePart, ")")
		if responseStart != -1 && responseEnd != -1 {
			responsePart = responsePart[responseStart+1 : responseEnd]
			responsePart = strings.TrimSpace(responsePart)
			if strings.HasPrefix(responsePart, "stream ") {
				rpc.ResponseType = strings.TrimPrefix(responsePart, "stream ")
				if rpc.Streaming == "client_streaming" {
					rpc.Streaming = "bidirectional"
				} else {
					rpc.Streaming = "server_streaming"
				}
			} else {
				rpc.ResponseType = responsePart
			}
		}
	}

	return rpc
}

// parseMessages extracts messages from proto content
func (p *Plugin) parseMessages(content string) []Message {
	messages := make([]Message, 0)
	lines := strings.Split(content, "\n")

	var currentMessage *Message
	var currentSchema *ir.Schema

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Detect message start
		if strings.HasPrefix(line, "message ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				messageName := strings.TrimSuffix(parts[1], "{")
				currentMessage = &Message{
					Name: messageName,
				}
				currentSchema = &ir.Schema{
					Type:       "object",
					Properties: make(map[string]*ir.Schema),
					Metadata:   make(map[string]interface{}),
				}

				// Get description from previous line
				if i > 0 {
					prevLine := strings.TrimSpace(lines[i-1])
					if strings.HasPrefix(prevLine, "//") {
						currentMessage.Description = strings.TrimPrefix(prevLine, "//")
						currentMessage.Description = strings.TrimSpace(currentMessage.Description)
					}
				}
			}
		}

		// Parse fields
		if currentMessage != nil && currentSchema != nil && !strings.HasPrefix(line, "message") && !strings.HasPrefix(line, "}") && line != "" && !strings.HasPrefix(line, "//") {
			field := p.parseField(line)
			if field != nil {
				currentSchema.Properties[field.Name] = field.Schema
			}
		}

		// Detect message end
		if currentMessage != nil && strings.HasPrefix(line, "}") {
			currentMessage.Schema = currentSchema
			messages = append(messages, *currentMessage)
			currentMessage = nil
			currentSchema = nil
		}
	}

	return messages
}

// parseField parses a message field
func (p *Plugin) parseField(line string) *struct {
	Name   string
	Schema *ir.Schema
} {
	// Format: string name = 1;
	// or: repeated string tags = 2;
	parts := strings.Fields(line)
	if len(parts) < 4 {
		return nil
	}

	idx := 0
	repeated := false
	if parts[0] == "repeated" {
		repeated = true
		idx = 1
	}

	protoType := parts[idx]
	fieldName := parts[idx+1]

	schema := &ir.Schema{
		Type:     mapProtoTypeToJSON(protoType),
		Metadata: make(map[string]interface{}),
	}

	if repeated {
		schema = &ir.Schema{
			Type:  "array",
			Items: schema,
		}
	}

	return &struct {
		Name   string
		Schema *ir.Schema
	}{
		Name:   fieldName,
		Schema: schema,
	}
}

// mapProtoTypeToJSON maps proto types to JSON types
func mapProtoTypeToJSON(protoType string) string {
	mapping := map[string]string{
		"string":   "string",
		"int32":    "integer",
		"int64":    "integer",
		"uint32":   "integer",
		"uint64":   "integer",
		"sint32":   "integer",
		"sint64":   "integer",
		"fixed32":  "integer",
		"fixed64":  "integer",
		"sfixed32": "integer",
		"sfixed64": "integer",
		"bool":     "boolean",
		"float":    "number",
		"double":   "number",
		"bytes":    "string",
	}

	if jsonType, ok := mapping[protoType]; ok {
		return jsonType
	}
	return "object" // For custom message types
}

// extractPackageName extracts the package name from proto content
func extractPackageName(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return strings.TrimSuffix(parts[1], ";")
			}
		}
	}
	return ""
}

// GetTemplate returns the custom template
func (p *Plugin) GetTemplate() string {
	return ""
}
