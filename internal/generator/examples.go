package generator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirrobot01/unifydoc/internal/ir"
)

// ExampleGenerator generates code examples in multiple languages
type ExampleGenerator struct {
	languages []string
}

// NewExampleGenerator creates a new example generator
func NewExampleGenerator(languages []string) *ExampleGenerator {
	if len(languages) == 0 {
		languages = []string{"curl", "javascript", "python", "go"}
	}
	return &ExampleGenerator{
		languages: languages,
	}
}

// Generate produces protocol-appropriate code examples for a resource.
func (g *ExampleGenerator) Generate(resource *ir.Resource, protocol, baseURL string) map[string]string {
	switch protocol {
	case "grpc":
		return g.generateGRPC(resource, baseURL)
	case "websocket":
		return g.generateWebSocket(resource, baseURL)
	case "asyncapi", "events":
		return g.generateMessaging(resource, baseURL)
	case "mcp":
		return g.generateMCP(resource)
	default: // openapi, api, webhook, custom — HTTP request/response
		return g.generateHTTP(resource, baseURL)
	}
}

// generateHTTP generates HTTP client examples in the configured languages.
func (g *ExampleGenerator) generateHTTP(resource *ir.Resource, baseURL string) map[string]string {
	examples := make(map[string]string)
	for _, lang := range g.languages {
		switch lang {
		case "curl":
			examples["curl"] = g.generateCurl(resource, baseURL)
		case "javascript":
			examples["javascript"] = g.generateJavaScript(resource, baseURL)
		case "python":
			examples["python"] = g.generatePython(resource, baseURL)
		case "go":
			examples["go"] = g.generateGo(resource, baseURL)
		}
	}
	return examples
}

// generateCurl generates a curl example
func (g *ExampleGenerator) generateCurl(resource *ir.Resource, baseURL string) string {
	var sb strings.Builder

	method := strings.ToUpper(resource.Method)
	if method == "" || method == "RPC" || method == "EVENT" || method == "TOOL" || method == "RESOURCE" || method == "PROMPT" {
		method = "GET"
	}

	url := baseURL + resource.Path
	sb.WriteString(fmt.Sprintf("curl -X %s \\\n", method))
	sb.WriteString(fmt.Sprintf("  '%s'", url))

	// Add headers
	hasContentType := false
	for _, param := range resource.Parameters {
		if param.In == "header" {
			sb.WriteString(fmt.Sprintf(" \\\n  -H '%s: %v'", param.Name, param.Example))
			if param.Name == "Content-Type" {
				hasContentType = true
			}
		}
	}

	// Add request body
	if resource.Request != nil && (method == "POST" || method == "PUT" || method == "PATCH") {
		if !hasContentType {
			sb.WriteString(" \\\n  -H 'Content-Type: application/json'")
		}

		var exampleData interface{}
		if len(resource.Examples) > 0 {
			exampleData = resource.Examples[0].Value
		} else {
			exampleData = g.generateExampleData(resource.Request)
		}

		jsonData, _ := json.MarshalIndent(exampleData, "  ", "  ")
		sb.WriteString(fmt.Sprintf(" \\\n  -d '%s'", string(jsonData)))
	}

	return sb.String()
}

// generateJavaScript generates a JavaScript fetch example
func (g *ExampleGenerator) generateJavaScript(resource *ir.Resource, baseURL string) string {
	var sb strings.Builder

	method := strings.ToUpper(resource.Method)
	if method == "" || method == "RPC" || method == "EVENT" || method == "TOOL" || method == "RESOURCE" || method == "PROMPT" {
		method = "GET"
	}

	url := baseURL + resource.Path

	sb.WriteString(fmt.Sprintf("fetch('%s', {\n", url))
	sb.WriteString(fmt.Sprintf("  method: '%s',\n", method))

	// Add headers
	headers := make(map[string]string)
	for _, param := range resource.Parameters {
		if param.In == "header" {
			headers[param.Name] = fmt.Sprintf("%v", param.Example)
		}
	}

	if resource.Request != nil && (method == "POST" || method == "PUT" || method == "PATCH") {
		headers["Content-Type"] = "application/json"

		var exampleData interface{}
		if len(resource.Examples) > 0 {
			exampleData = resource.Examples[0].Value
		} else {
			exampleData = g.generateExampleData(resource.Request)
		}

		if len(headers) > 0 {
			sb.WriteString("  headers: {\n")
			for k, v := range headers {
				sb.WriteString(fmt.Sprintf("    '%s': '%s',\n", k, v))
			}
			sb.WriteString("  },\n")
		}

		jsonData, _ := json.MarshalIndent(exampleData, "  ", "  ")
		sb.WriteString(fmt.Sprintf("  body: JSON.stringify(%s)\n", string(jsonData)))
	} else if len(headers) > 0 {
		sb.WriteString("  headers: {\n")
		for k, v := range headers {
			sb.WriteString(fmt.Sprintf("    '%s': '%s',\n", k, v))
		}
		sb.WriteString("  }\n")
	}

	sb.WriteString("})\n")
	sb.WriteString("  .then(response => response.json())\n")
	sb.WriteString("  .then(data => console.log(data))\n")
	sb.WriteString("  .catch(error => console.error('Error:', error));")

	return sb.String()
}

// generatePython generates a Python requests example
func (g *ExampleGenerator) generatePython(resource *ir.Resource, baseURL string) string {
	var sb strings.Builder

	method := strings.ToLower(resource.Method)
	if method == "" || method == "rpc" || method == "event" || method == "tool" || method == "resource" || method == "prompt" {
		method = "get"
	}

	url := baseURL + resource.Path

	sb.WriteString("import requests\n\n")

	// Prepare headers
	hasHeaders := false
	for _, param := range resource.Parameters {
		if param.In == "header" {
			hasHeaders = true
			break
		}
	}

	if hasHeaders || resource.Request != nil {
		sb.WriteString("headers = {\n")
		for _, param := range resource.Parameters {
			if param.In == "header" {
				sb.WriteString(fmt.Sprintf("    '%s': '%v',\n", param.Name, param.Example))
			}
		}
		if resource.Request != nil && (method == "post" || method == "put" || method == "patch") {
			sb.WriteString("    'Content-Type': 'application/json',\n")
		}
		sb.WriteString("}\n\n")
	}

	// Prepare data
	if resource.Request != nil && (method == "post" || method == "put" || method == "patch") {
		var exampleData interface{}
		if len(resource.Examples) > 0 {
			exampleData = resource.Examples[0].Value
		} else {
			exampleData = g.generateExampleData(resource.Request)
		}

		jsonData, _ := json.MarshalIndent(exampleData, "", "    ")
		sb.WriteString(fmt.Sprintf("data = %s\n\n", string(jsonData)))
	}

	// Make request
	sb.WriteString(fmt.Sprintf("response = requests.%s(\n", method))
	sb.WriteString(fmt.Sprintf("    '%s'", url))

	if hasHeaders || resource.Request != nil {
		sb.WriteString(",\n    headers=headers")
	}

	if resource.Request != nil && (method == "post" || method == "put" || method == "patch") {
		sb.WriteString(",\n    json=data")
	}

	sb.WriteString("\n)\n\n")
	sb.WriteString("print(response.json())")

	return sb.String()
}

// generateGo generates a Go example
func (g *ExampleGenerator) generateGo(resource *ir.Resource, baseURL string) string {
	var sb strings.Builder

	method := strings.ToUpper(resource.Method)
	if method == "" || method == "RPC" || method == "EVENT" || method == "TOOL" || method == "RESOURCE" || method == "PROMPT" {
		method = "GET"
	}

	url := baseURL + resource.Path

	sb.WriteString("package main\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("    \"bytes\"\n")
	sb.WriteString("    \"encoding/json\"\n")
	sb.WriteString("    \"fmt\"\n")
	sb.WriteString("    \"io\"\n")
	sb.WriteString("    \"net/http\"\n")
	sb.WriteString(")\n\n")

	sb.WriteString("func main() {\n")

	// Prepare request body
	if resource.Request != nil && (method == "POST" || method == "PUT" || method == "PATCH") {
		var exampleData interface{}
		if len(resource.Examples) > 0 {
			exampleData = resource.Examples[0].Value
		} else {
			exampleData = g.generateExampleData(resource.Request)
		}

		jsonData, _ := json.MarshalIndent(exampleData, "    ", "    ")
		sb.WriteString(fmt.Sprintf("    data := []byte(`%s`)\n", string(jsonData)))
		sb.WriteString(fmt.Sprintf("    req, err := http.NewRequest(\"%s\", \"%s\", bytes.NewBuffer(data))\n", method, url))
	} else {
		sb.WriteString(fmt.Sprintf("    req, err := http.NewRequest(\"%s\", \"%s\", nil)\n", method, url))
	}

	sb.WriteString("    if err != nil {\n")
	sb.WriteString("        panic(err)\n")
	sb.WriteString("    }\n\n")

	// Add headers
	for _, param := range resource.Parameters {
		if param.In == "header" {
			sb.WriteString(fmt.Sprintf("    req.Header.Set(\"%s\", \"%v\")\n", param.Name, param.Example))
		}
	}

	if resource.Request != nil && (method == "POST" || method == "PUT" || method == "PATCH") {
		sb.WriteString("    req.Header.Set(\"Content-Type\", \"application/json\")\n\n")
	}

	sb.WriteString("    client := &http.Client{}\n")
	sb.WriteString("    resp, err := client.Do(req)\n")
	sb.WriteString("    if err != nil {\n")
	sb.WriteString("        panic(err)\n")
	sb.WriteString("    }\n")
	sb.WriteString("    defer resp.Body.Close()\n\n")

	sb.WriteString("    body, _ := io.ReadAll(resp.Body)\n")
	sb.WriteString("    fmt.Println(string(body))\n")
	sb.WriteString("}")

	return sb.String()
}

// ---------- Protocol-aware example generators ----------

// requestExample returns illustrative request/payload data for a resource,
// preferring an explicit example then falling back to the request schema.
func (g *ExampleGenerator) requestExample(resource *ir.Resource) interface{} {
	if len(resource.Examples) > 0 {
		return resource.Examples[0].Value
	}
	if resource.Request != nil {
		return g.generateExampleData(resource.Request)
	}
	return nil
}

func metaString(resource *ir.Resource, key string) string {
	if resource.Metadata == nil {
		return ""
	}
	if v, ok := resource.Metadata[key].(string); ok {
		return v
	}
	return ""
}

// generateGRPC produces a runnable grpcurl command plus Go and Python stubs.
func (g *ExampleGenerator) generateGRPC(resource *ir.Resource, baseURL string) map[string]string {
	host := grpcHost(baseURL)
	fullMethod := strings.TrimPrefix(resource.Path, "/") // pkg.Service/Method
	service := metaString(resource, "service")
	reqType := metaString(resource, "request_type")
	streaming := metaString(resource, "streaming")
	reqData := g.requestExample(resource)

	examples := make(map[string]string)

	// grpcurl — accurate and runnable.
	var gc strings.Builder
	if streaming != "" && streaming != "unary" {
		gc.WriteString(fmt.Sprintf("# %s\n", strings.ReplaceAll(streaming, "_", " ")))
	}
	gc.WriteString("grpcurl -plaintext \\\n")
	if reqData != nil {
		gc.WriteString(fmt.Sprintf("  -d '%s' \\\n", compactJSON(reqData)))
	}
	gc.WriteString(fmt.Sprintf("  %s \\\n  %s", host, fullMethod))
	examples["grpcurl"] = gc.String()

	// Go client stub.
	var goc strings.Builder
	goc.WriteString(fmt.Sprintf("conn, err := grpc.Dial(%q, grpc.WithTransportCredentials(insecure.NewCredentials()))\n", host))
	goc.WriteString("if err != nil {\n    panic(err)\n}\n")
	goc.WriteString("defer conn.Close()\n\n")
	goc.WriteString(fmt.Sprintf("client := pb.New%sClient(conn)\n", service))
	if reqData != nil {
		goc.WriteString(fmt.Sprintf("// request: %s\n", compactJSON(reqData)))
	}
	goc.WriteString(fmt.Sprintf("resp, err := client.%s(ctx, &pb.%s{})\n", resource.Name, reqType))
	goc.WriteString("if err != nil {\n    panic(err)\n}\nfmt.Println(resp)")
	examples["go"] = goc.String()

	// Python client stub.
	var py strings.Builder
	py.WriteString("import grpc\n\n")
	py.WriteString(fmt.Sprintf("channel = grpc.insecure_channel(%q)\n", host))
	py.WriteString(fmt.Sprintf("stub = %sStub(channel)\n", service))
	if reqData != nil {
		py.WriteString(fmt.Sprintf("# request: %s\n", compactJSON(reqData)))
	}
	py.WriteString(fmt.Sprintf("response = stub.%s(%s())\n", resource.Name, reqType))
	py.WriteString("print(response)")
	examples["python"] = py.String()

	return examples
}

// generateWebSocket produces browser, Python and wscat examples honouring the
// event direction (send vs receive).
func (g *ExampleGenerator) generateWebSocket(resource *ir.Resource, baseURL string) map[string]string {
	url := baseURL
	if url == "" {
		url = "wss://example.com/ws"
	}
	direction := metaString(resource, "direction")
	payload := g.requestExample(resource)
	examples := make(map[string]string)

	var js strings.Builder
	js.WriteString(fmt.Sprintf("const ws = new WebSocket(%q);\n\n", url))
	if direction == "receive" {
		js.WriteString("ws.addEventListener('message', (event) => {\n")
		js.WriteString("  const data = JSON.parse(event.data);\n")
		js.WriteString(fmt.Sprintf("  // %s\n", resource.Name))
		js.WriteString("  console.log(data);\n});")
	} else {
		js.WriteString("ws.addEventListener('open', () => {\n")
		if payload != nil {
			js.WriteString(fmt.Sprintf("  ws.send(JSON.stringify(%s));\n", indentJSON(payload, "  ")))
		} else {
			js.WriteString(fmt.Sprintf("  ws.send(JSON.stringify({ event: %q }));\n", resource.Name))
		}
		js.WriteString("});")
	}
	examples["javascript"] = js.String()

	var py strings.Builder
	py.WriteString("import asyncio, json, websockets\n\n")
	py.WriteString("async def main():\n")
	py.WriteString(fmt.Sprintf("    async with websockets.connect(%q) as ws:\n", url))
	if direction == "receive" {
		py.WriteString("        message = json.loads(await ws.recv())\n")
		py.WriteString("        print(message)\n\n")
	} else {
		if payload != nil {
			py.WriteString(fmt.Sprintf("        await ws.send(json.dumps(%s))\n\n", compactJSON(payload)))
		} else {
			py.WriteString(fmt.Sprintf("        await ws.send(json.dumps({\"event\": %q}))\n\n", resource.Name))
		}
	}
	py.WriteString("asyncio.run(main())")
	examples["python"] = py.String()

	// wscat — quick manual testing.
	wscat := fmt.Sprintf("wscat -c %s", url)
	if direction != "receive" && payload != nil {
		wscat += fmt.Sprintf("\n> %s", compactJSON(payload))
	}
	examples["wscat"] = wscat

	return examples
}

// generateMessaging produces a message payload plus publish/subscribe snippets
// for event-driven protocols (AsyncAPI, Events).
func (g *ExampleGenerator) generateMessaging(resource *ir.Resource, baseURL string) map[string]string {
	channel := metaString(resource, "channel")
	if channel == "" {
		channel = metaString(resource, "topic")
	}
	if channel == "" {
		channel = resource.Name
	}
	op := metaString(resource, "operation_type") // publish / subscribe (asyncapi)
	payload := g.requestExample(resource)
	examples := make(map[string]string)

	if payload != nil {
		examples["payload"] = prettyJSON(payload)
	}

	subscribe := op == "subscribe" || metaString(resource, "type") == "event"

	var js strings.Builder
	if subscribe {
		js.WriteString(fmt.Sprintf("// subscribe to %q\n", channel))
		js.WriteString(fmt.Sprintf("client.subscribe(%q, (message) => {\n", channel))
		js.WriteString("  console.log(JSON.parse(message));\n});")
	} else {
		js.WriteString(fmt.Sprintf("// publish to %q\n", channel))
		if payload != nil {
			js.WriteString(fmt.Sprintf("client.publish(%q, JSON.stringify(%s));", channel, indentJSON(payload, "")))
		} else {
			js.WriteString(fmt.Sprintf("client.publish(%q, JSON.stringify({}));", channel))
		}
	}
	examples["javascript"] = js.String()

	return examples
}

// generateMCP produces a JSON-RPC request and client call for MCP tools,
// resources and prompts.
func (g *ExampleGenerator) generateMCP(resource *ir.Resource) map[string]string {
	kind := metaString(resource, "type") // mcp-tool / mcp-resource / mcp-prompt
	examples := make(map[string]string)

	switch kind {
	case "mcp-resource":
		uri := resource.Path
		if uri == "" {
			uri = resource.Name
		}
		examples["json"] = prettyJSON(map[string]interface{}{
			"method": "resources/read",
			"params": map[string]interface{}{"uri": uri},
		})
		examples["javascript"] = fmt.Sprintf("const result = await client.readResource({ uri: %q });\nconsole.log(result.contents);", uri)
	case "mcp-prompt":
		examples["json"] = prettyJSON(map[string]interface{}{
			"method": "prompts/get",
			"params": map[string]interface{}{"name": resource.Name},
		})
		examples["javascript"] = fmt.Sprintf("const prompt = await client.getPrompt({ name: %q });\nconsole.log(prompt.messages);", resource.Name)
	default: // mcp-tool
		args := g.requestExample(resource)
		if args == nil {
			args = map[string]interface{}{}
		}
		examples["json"] = prettyJSON(map[string]interface{}{
			"method": "tools/call",
			"params": map[string]interface{}{"name": resource.Name, "arguments": args},
		})
		examples["javascript"] = fmt.Sprintf("const result = await client.callTool({\n  name: %q,\n  arguments: %s,\n});\nconsole.log(result);", resource.Name, indentJSON(args, "  "))
		examples["python"] = fmt.Sprintf("result = await session.call_tool(%q, arguments=%s)\nprint(result)", resource.Name, compactJSON(args))
	}

	return examples
}

// ---------- JSON helpers ----------

func compactJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func prettyJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}

// indentJSON pretty-prints v with every line after the first prefixed by
// indent, so it embeds cleanly inside a surrounding snippet.
func indentJSON(v interface{}, indent string) string {
	b, err := json.MarshalIndent(v, indent, "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}

// grpcHost derives a host:port target from a base URL, defaulting to a common
// local gRPC address.
func grpcHost(baseURL string) string {
	host := baseURL
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "grpc://")
	host = strings.TrimSuffix(host, "/")
	if host == "" {
		return "localhost:50051"
	}
	return host
}

// generateExampleData generates example data from a schema
func (g *ExampleGenerator) generateExampleData(schema *ir.Schema) interface{} {
	if schema == nil {
		return nil
	}

	// If there's an example, use it
	if schema.Example != nil {
		return schema.Example
	}

	// Generate based on type
	switch schema.Type {
	case "object":
		obj := make(map[string]interface{})
		for name, prop := range schema.Properties {
			obj[name] = g.generateExampleData(prop)
		}
		return obj

	case "array":
		if schema.Items != nil {
			return []interface{}{g.generateExampleData(schema.Items)}
		}
		return []interface{}{}

	case "string":
		if schema.Format == "email" {
			return "user@example.com"
		}
		if schema.Format == "date-time" {
			return "2025-01-01T00:00:00Z"
		}
		if len(schema.Enum) > 0 {
			return schema.Enum[0]
		}
		return "string"

	case "integer":
		return 0

	case "number":
		return 0.0

	case "boolean":
		return true

	default:
		return nil
	}
}
