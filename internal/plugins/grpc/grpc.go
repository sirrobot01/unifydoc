package grpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bufbuild/protocompile"
	"github.com/sirrobot01/unifydoc/internal/ir"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// mainFile is the synthetic filename used to compile the in-memory spec.
const mainFile = "unifidoc_input.proto"

// Plugin implements the gRPC plugin backed by a real .proto parser.
type Plugin struct{}

// NewPlugin creates a new gRPC plugin
func NewPlugin() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name
func (p *Plugin) Name() string { return "grpc" }

// Version returns the plugin version
func (p *Plugin) Version() string { return "1.0.0" }

// Validate performs a lightweight syntactic check on a proto file.
func (p *Plugin) Validate(spec []byte) error {
	content := string(spec)
	if !strings.Contains(content, "syntax") && !strings.Contains(content, "service") {
		return fmt.Errorf("invalid proto file: missing syntax or service declaration")
	}
	return nil
}

// compile parses and links the in-memory proto source into a FileDescriptor.
func compile(spec []byte) (protoreflect.FileDescriptor, error) {
	resolver := protocompile.WithStandardImports(&protocompile.SourceResolver{
		Accessor: func(path string) (io.ReadCloser, error) {
			if path == mainFile {
				return io.NopCloser(bytes.NewReader(spec)), nil
			}
			return nil, os.ErrNotExist
		},
	})
	compiler := protocompile.Compiler{Resolver: resolver}
	files, err := compiler.Compile(context.Background(), mainFile)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no proto file compiled")
	}
	return files[0], nil
}

// Parse parses a Protocol Buffer specification and converts it to IR.
func (p *Plugin) Parse(spec []byte) (*ir.IR, error) {
	fd, err := compile(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proto: %w", err)
	}

	result := ir.NewIR("grpc")
	result.Title = "gRPC Service"
	result.Version = "1.0.0"
	pkg := string(fd.Package())
	if pkg != "" {
		result.Metadata["package"] = pkg
	}

	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		svc := services.Get(i)
		if i == 0 {
			result.Title = string(svc.Name())
			result.Description = leadingComment(svc)
		}

		methods := svc.Methods()
		for j := 0; j < methods.Len(); j++ {
			m := methods.Get(j)
			result.Resources = append(result.Resources, buildRPC(pkg, svc, m))
		}
	}

	// Top-level messages and enums become documented types.
	msgs := fd.Messages()
	for i := 0; i < msgs.Len(); i++ {
		md := msgs.Get(i)
		result.Types = append(result.Types, ir.TypeDef{
			Name:        string(md.Name()),
			Description: leadingComment(md),
			Schema:      messageToSchema(md, nil),
		})
	}
	enums := fd.Enums()
	for i := 0; i < enums.Len(); i++ {
		ed := enums.Get(i)
		result.Types = append(result.Types, ir.TypeDef{
			Name:        string(ed.Name()),
			Description: leadingComment(ed),
			Schema:      enumToSchema(ed),
		})
	}

	return result, nil
}

// buildRPC converts a single RPC method into an IR resource with resolved
// request/response schemas and streaming metadata.
func buildRPC(pkg string, svc protoreflect.ServiceDescriptor, m protoreflect.MethodDescriptor) ir.Resource {
	svcPath := string(svc.Name())
	if pkg != "" {
		svcPath = pkg + "." + svcPath
	}

	resource := ir.Resource{
		Name:        string(m.Name()),
		Path:        fmt.Sprintf("/%s/%s", svcPath, m.Name()),
		Method:      "RPC",
		Description: leadingComment(m),
		Request:     messageToSchema(m.Input(), nil),
		Response:    messageToSchema(m.Output(), nil),
		Metadata:    make(map[string]interface{}),
	}

	resource.Metadata["streaming"] = streamingMode(m)
	resource.Metadata["service"] = string(svc.Name())
	resource.Metadata["request_type"] = string(m.Input().Name())
	resource.Metadata["response_type"] = string(m.Output().Name())
	return resource
}

func streamingMode(m protoreflect.MethodDescriptor) string {
	switch {
	case m.IsStreamingClient() && m.IsStreamingServer():
		return "bidirectional"
	case m.IsStreamingClient():
		return "client_streaming"
	case m.IsStreamingServer():
		return "server_streaming"
	default:
		return "unary"
	}
}

// messageToSchema converts a message descriptor to an IR object schema,
// recursively resolving nested message fields. The seen set breaks cycles
// along the current ancestry (e.g. a message that references itself).
func messageToSchema(md protoreflect.MessageDescriptor, seen map[protoreflect.FullName]bool) *ir.Schema {
	schema := &ir.Schema{
		Type:       "object",
		Properties: make(map[string]*ir.Schema),
		Metadata:   map[string]interface{}{"message_type": string(md.FullName())},
	}
	seen = cloneWith(seen, md.FullName())

	fields := md.Fields()
	var required []string
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		s := fieldToSchema(f, seen)
		if c := leadingComment(f); c != "" {
			s.Description = c
		}
		schema.Properties[string(f.Name())] = s
		if f.Cardinality() == protoreflect.Required {
			required = append(required, string(f.Name()))
		}
	}
	schema.Required = required
	return schema
}

// fieldToSchema converts a field descriptor, handling maps, repeated fields,
// nested messages, enums and scalars.
func fieldToSchema(f protoreflect.FieldDescriptor, seen map[protoreflect.FullName]bool) *ir.Schema {
	if f.IsMap() {
		return &ir.Schema{
			Type:        "object",
			Description: fmt.Sprintf("map<%s, %s>", scalarType(f.MapKey()), valueLabel(f.MapValue())),
			Metadata:    map[string]interface{}{"map": true},
		}
	}

	base := scalarOrMessage(f, seen)
	if f.IsList() {
		return &ir.Schema{Type: "array", Items: base}
	}
	return base
}

func scalarOrMessage(f protoreflect.FieldDescriptor, seen map[protoreflect.FullName]bool) *ir.Schema {
	switch f.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		md := f.Message()
		if seen[md.FullName()] {
			// Cycle: reference by name instead of expanding.
			return &ir.Schema{Type: "object", Ref: string(md.FullName())}
		}
		return messageToSchema(md, seen)
	case protoreflect.EnumKind:
		return enumToSchema(f.Enum())
	default:
		return &ir.Schema{Type: scalarKind(f.Kind())}
	}
}

func enumToSchema(ed protoreflect.EnumDescriptor) *ir.Schema {
	values := ed.Values()
	enum := make([]interface{}, 0, values.Len())
	for i := 0; i < values.Len(); i++ {
		enum = append(enum, string(values.Get(i).Name()))
	}
	return &ir.Schema{Type: "string", Enum: enum}
}

// valueLabel returns a short type label for a map value field.
func valueLabel(f protoreflect.FieldDescriptor) string {
	if f.Kind() == protoreflect.MessageKind || f.Kind() == protoreflect.GroupKind {
		return string(f.Message().Name())
	}
	if f.Kind() == protoreflect.EnumKind {
		return string(f.Enum().Name())
	}
	return scalarType(f)
}

func scalarType(f protoreflect.FieldDescriptor) string {
	return scalarKind(f.Kind())
}

func scalarKind(k protoreflect.Kind) string {
	switch k {
	case protoreflect.BoolKind:
		return "boolean"
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return "number"
	case protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Uint32Kind,
		protoreflect.Uint64Kind, protoreflect.Sint32Kind, protoreflect.Sint64Kind,
		protoreflect.Fixed32Kind, protoreflect.Fixed64Kind, protoreflect.Sfixed32Kind,
		protoreflect.Sfixed64Kind:
		return "integer"
	case protoreflect.StringKind, protoreflect.BytesKind:
		return "string"
	default:
		return "string"
	}
}

// leadingComment returns the trimmed leading comment for a descriptor, if any.
func leadingComment(d protoreflect.Descriptor) string {
	loc := d.ParentFile().SourceLocations().ByDescriptor(d)
	return strings.TrimSpace(loc.LeadingComments)
}

// cloneWith returns a copy of seen with name added, so sibling branches don't
// share cycle-breaking state.
func cloneWith(seen map[protoreflect.FullName]bool, name protoreflect.FullName) map[protoreflect.FullName]bool {
	next := make(map[protoreflect.FullName]bool, len(seen)+1)
	for k := range seen {
		next[k] = true
	}
	next[name] = true
	return next
}

// GetTemplate returns the custom template
func (p *Plugin) GetTemplate() string { return "" }
