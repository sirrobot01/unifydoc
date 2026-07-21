package asyncapi

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sirrobot01/unifydoc/internal/ir"
	"gopkg.in/yaml.v3"
)

// Plugin implements the AsyncAPI plugin with support for both 2.x and 3.x
// specifications, including local $ref resolution.
type Plugin struct{}

// NewPlugin creates a new AsyncAPI plugin
func NewPlugin() *Plugin { return &Plugin{} }

// Name returns the plugin name
func (p *Plugin) Name() string { return "asyncapi" }

// Version returns the plugin version
func (p *Plugin) Version() string { return "1.0.0" }

// Validate validates an AsyncAPI specification.
func (p *Plugin) Validate(spec []byte) error {
	root, err := unmarshal(spec)
	if err != nil {
		return fmt.Errorf("invalid AsyncAPI YAML: %w", err)
	}
	if asString(root["asyncapi"]) == "" {
		return fmt.Errorf("asyncapi version is required")
	}
	info, _ := asMap(root["info"])
	if asString(info["title"]) == "" {
		return fmt.Errorf("info.title is required")
	}
	if asString(info["version"]) == "" {
		return fmt.Errorf("info.version is required")
	}
	return nil
}

// Parse parses an AsyncAPI specification and converts it to IR.
func (p *Plugin) Parse(spec []byte) (*ir.IR, error) {
	root, err := unmarshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AsyncAPI spec: %w", err)
	}

	version := asString(root["asyncapi"])
	result := ir.NewIR("asyncapi")
	result.Metadata["asyncapi_version"] = version

	info, _ := asMap(root["info"])
	result.Title = asString(info["title"])
	result.Description = asString(info["description"])
	result.Version = asString(info["version"])

	parseServers(root, result)

	if strings.HasPrefix(version, "3") {
		parseOperations3(root, result)
	} else {
		parseChannels2(root, result)
	}

	parseComponentSchemas(root, result)
	return result, nil
}

// parseServers extracts servers, handling both 2.x (url) and 3.x (host+pathname).
func parseServers(root map[string]interface{}, result *ir.IR) {
	servers, _ := asMap(root["servers"])
	names := sortedKeys(servers)
	for _, name := range names {
		s, ok := asMap(servers[name])
		if !ok {
			continue
		}
		url := asString(s["url"])
		if url == "" { // AsyncAPI 3.x
			url = asString(s["host"]) + asString(s["pathname"])
		}
		protocol := asString(s["protocol"])
		result.Servers = append(result.Servers, ir.Server{
			URL:         url,
			Description: strings.TrimSpace(fmt.Sprintf("%s (%s)", asString(s["description"]), protocol)),
			Variables:   map[string]string{"protocol": protocol},
		})
		result.Metadata["server_"+name] = protocol
	}
}

// parseChannels2 handles AsyncAPI 2.x, where each channel carries publish and
// subscribe operations directly.
func parseChannels2(root map[string]interface{}, result *ir.IR) {
	channels, _ := asMap(root["channels"])
	for _, name := range sortedKeys(channels) {
		ch, ok := asMap(channels[name])
		if !ok {
			continue
		}
		chDesc := asString(ch["description"])
		for _, opType := range []string{"subscribe", "publish"} {
			op, ok := asMap(ch[opType])
			if !ok {
				continue
			}
			payload, msgName := messagePayload(root, op["message"])
			result.Resources = append(result.Resources,
				buildResource(name, opType, opType, op, chDesc, payload, msgName))
		}
	}
}

// parseOperations3 handles AsyncAPI 3.x, where operations are top-level and
// reference channels and messages by $ref.
func parseOperations3(root map[string]interface{}, result *ir.IR) {
	operations, _ := asMap(root["operations"])
	for _, opID := range sortedKeys(operations) {
		op, ok := asMap(operations[opID])
		if !ok {
			continue
		}
		action := asString(op["action"]) // send | receive
		opType := normalizeAction(action)

		channelName := opID
		var chMap map[string]interface{}
		if chRef, ok := asMap(op["channel"]); ok {
			if ref := asString(chRef["$ref"]); ref != "" {
				channelName = refName(ref)
				if resolved, ok := asMap(resolveRef(root, ref)); ok {
					chMap = resolved
				}
			}
		}

		// Messages: prefer the operation's messages, else the channel's.
		var msgNode interface{}
		if msgs, ok := op["messages"].([]interface{}); ok && len(msgs) > 0 {
			msgNode = msgs[0]
		} else if chMap != nil {
			if chMsgs, ok := asMap(chMap["messages"]); ok {
				for _, k := range sortedKeys(chMsgs) {
					msgNode = chMsgs[k]
					break
				}
			}
		}

		chDesc := ""
		if chMap != nil {
			chDesc = asString(chMap["description"])
		}
		payload, msgName := messagePayload(root, msgNode)
		result.Resources = append(result.Resources,
			buildResource(channelName, opType, action, op, chDesc, payload, msgName))
	}
}

// buildResource assembles an IR resource for one operation.
func buildResource(channel, opType, action string, op map[string]interface{}, chDesc string, payload *ir.Schema, msgName string) ir.Resource {
	resource := ir.Resource{
		Name:        asString(op["summary"]),
		Path:        channel,
		Method:      opType, // publish | subscribe -> PUB/SUB badge
		Description: asString(op["description"]),
		Request:     payload,
		Metadata:    make(map[string]interface{}),
		Tags:        make([]string, 0),
	}
	if resource.Name == "" {
		resource.Name = fmt.Sprintf("%s %s", opType, channel)
	}
	if resource.Description == "" {
		resource.Description = chDesc
	}
	resource.Metadata["operation_type"] = opType
	resource.Metadata["channel"] = channel
	if action != "" {
		resource.Metadata["action"] = action
	}
	if msgName != "" {
		resource.Metadata["message_name"] = msgName
	}

	if tags, ok := op["tags"].([]interface{}); ok {
		for _, t := range tags {
			if tm, ok := asMap(t); ok {
				if n := asString(tm["name"]); n != "" {
					resource.Tags = append(resource.Tags, n)
				}
			}
		}
	}
	return resource
}

// messagePayload resolves a message node (which may be a $ref, an inline
// message, or a oneOf list) into its payload schema and message name.
func messagePayload(root map[string]interface{}, msgNode interface{}) (*ir.Schema, string) {
	msg, ok := asMap(resolveNode(root, msgNode))
	if !ok {
		return nil, ""
	}
	// oneOf: multiple messages — document the first.
	if oneOf, ok := msg["oneOf"].([]interface{}); ok && len(oneOf) > 0 {
		if first, ok := asMap(resolveNode(root, oneOf[0])); ok {
			msg = first
		}
	}
	name := asString(msg["name"])
	payload := convertSchema(root, msg["payload"], map[string]bool{})
	return payload, name
}

// parseComponentSchemas exposes reusable schemas as documented types.
func parseComponentSchemas(root map[string]interface{}, result *ir.IR) {
	components, ok := asMap(root["components"])
	if !ok {
		return
	}
	schemas, ok := asMap(components["schemas"])
	if !ok {
		return
	}
	for _, name := range sortedKeys(schemas) {
		schema := convertSchema(root, schemas[name], map[string]bool{})
		if schema == nil {
			continue
		}
		result.Types = append(result.Types, ir.TypeDef{
			Name:        name,
			Description: schema.Description,
			Schema:      schema,
		})
	}
}

// convertSchema converts a JSON-Schema-like node into an IR schema, resolving
// local $ref and guarding against reference cycles.
func convertSchema(root map[string]interface{}, node interface{}, seen map[string]bool) *ir.Schema {
	m, ok := asMap(node)
	if !ok {
		return nil
	}

	// Resolve $ref, breaking cycles.
	if ref := asString(m["$ref"]); ref != "" {
		if seen[ref] {
			return &ir.Schema{Type: "object", Ref: ref}
		}
		seen = cloneSet(seen, ref)
		resolved, ok := asMap(resolveRef(root, ref))
		if !ok {
			return &ir.Schema{Type: "object", Ref: ref}
		}
		schema := convertSchema(root, resolved, seen)
		if schema != nil && schema.Metadata == nil {
			schema.Metadata = map[string]interface{}{}
		}
		return schema
	}

	schema := &ir.Schema{
		Type:        asString(m["type"]),
		Format:      asString(m["format"]),
		Description: asString(m["description"]),
		Properties:  make(map[string]*ir.Schema),
		Metadata:    make(map[string]interface{}),
	}
	if schema.Type == "" {
		schema.Type = "object"
	}

	if props, ok := asMap(m["properties"]); ok {
		for name, propData := range props {
			if s := convertSchema(root, propData, seen); s != nil {
				schema.Properties[name] = s
			}
		}
	}
	if items := m["items"]; items != nil {
		schema.Items = convertSchema(root, items, seen)
	}
	if required, ok := m["required"].([]interface{}); ok {
		for _, r := range required {
			if rs := asString(r); rs != "" {
				schema.Required = append(schema.Required, rs)
			}
		}
	}
	if enum, ok := m["enum"].([]interface{}); ok {
		schema.Enum = enum
	}
	return schema
}

// ---------- helpers ----------

func unmarshal(spec []byte) (map[string]interface{}, error) {
	var root map[string]interface{}
	if err := yaml.Unmarshal(spec, &root); err != nil {
		return nil, err
	}
	if root == nil {
		return nil, fmt.Errorf("empty document")
	}
	return root, nil
}

func normalizeAction(action string) string {
	// AsyncAPI 3.x: "send" publishes to the channel, "receive" subscribes.
	switch action {
	case "send":
		return "publish"
	case "receive":
		return "subscribe"
	default:
		return action
	}
}

// resolveNode follows a chain of $ref indirections until it reaches a concrete
// node (bounded to avoid cycles).
func resolveNode(root map[string]interface{}, node interface{}) interface{} {
	for i := 0; i < 32; i++ {
		m, ok := asMap(node)
		if !ok {
			return node
		}
		ref := asString(m["$ref"])
		if ref == "" {
			return node
		}
		node = resolveRef(root, ref)
	}
	return node
}

// resolveRef resolves a local JSON pointer like "#/components/messages/Foo".
func resolveRef(root map[string]interface{}, ref string) interface{} {
	if !strings.HasPrefix(ref, "#/") {
		return nil
	}
	parts := strings.Split(strings.TrimPrefix(ref, "#/"), "/")
	var current interface{} = root
	for _, part := range parts {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		m, ok := asMap(current)
		if !ok {
			return nil
		}
		current, ok = m[part]
		if !ok {
			return nil
		}
	}
	return current
}

func refName(ref string) string {
	if i := strings.LastIndex(ref, "/"); i != -1 {
		return ref[i+1:]
	}
	return ref
}

func asMap(v interface{}) (map[string]interface{}, bool) {
	m, ok := v.(map[string]interface{})
	return m, ok
}

func asString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func cloneSet(s map[string]bool, add string) map[string]bool {
	next := make(map[string]bool, len(s)+1)
	for k := range s {
		next[k] = true
	}
	next[add] = true
	return next
}

// GetTemplate returns the custom template
func (p *Plugin) GetTemplate() string { return "" }
