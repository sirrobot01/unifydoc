package generator

import (
	"strings"
	"testing"

	"github.com/sirrobot01/unifydoc/internal/ir"
)

func newGen() *ExampleGenerator {
	return NewExampleGenerator([]string{"curl", "javascript", "python", "go"})
}

func TestGenerateHTTP(t *testing.T) {
	g := newGen()
	res := &ir.Resource{Name: "Create pet", Method: "POST", Path: "/pets",
		Request: &ir.Schema{Type: "object", Properties: map[string]*ir.Schema{"name": {Type: "string"}}}}
	ex := g.Generate(res, "openapi", "https://api.example.com")
	if !strings.Contains(ex["curl"], "curl -X POST") {
		t.Errorf("curl example missing POST: %q", ex["curl"])
	}
	if _, ok := ex["grpcurl"]; ok {
		t.Error("HTTP protocol should not emit a grpcurl example")
	}
}

func TestGenerateGRPC(t *testing.T) {
	g := newGen()
	res := &ir.Resource{
		Name:   "GetUser",
		Method: "RPC",
		Path:   "/user.UserService/GetUser",
		Request: &ir.Schema{Type: "object", Properties: map[string]*ir.Schema{
			"user_id": {Type: "string"}}},
		Metadata: map[string]interface{}{
			"service": "UserService", "request_type": "GetUserRequest",
			"response_type": "GetUserResponse", "streaming": "unary"},
	}
	ex := g.Generate(res, "grpc", "")

	if _, ok := ex["curl"]; ok {
		t.Error("gRPC must not produce an HTTP curl example")
	}
	grpcurl := ex["grpcurl"]
	if !strings.Contains(grpcurl, "grpcurl -plaintext") || !strings.Contains(grpcurl, "user.UserService/GetUser") {
		t.Errorf("grpcurl example malformed: %q", grpcurl)
	}
	if !strings.Contains(grpcurl, "localhost:50051") {
		t.Errorf("grpcurl should default host: %q", grpcurl)
	}
	if !strings.Contains(ex["go"], "NewUserServiceClient") {
		t.Errorf("go stub missing client constructor: %q", ex["go"])
	}
	if !strings.Contains(ex["python"], "GetUserRequest") {
		t.Errorf("python stub missing request type: %q", ex["python"])
	}
}

func TestGenerateGRPCStreamingNote(t *testing.T) {
	g := newGen()
	res := &ir.Resource{Name: "StreamUsers", Method: "RPC", Path: "/user.UserService/StreamUsers",
		Metadata: map[string]interface{}{"streaming": "server_streaming", "service": "UserService", "request_type": "Req"}}
	ex := g.Generate(res, "grpc", "")
	if !strings.Contains(ex["grpcurl"], "server streaming") {
		t.Errorf("expected streaming note in grpcurl: %q", ex["grpcurl"])
	}
}

func TestGenerateWebSocketSend(t *testing.T) {
	g := newGen()
	res := &ir.Resource{Name: "message.send", Method: "send",
		Request:  &ir.Schema{Type: "object", Properties: map[string]*ir.Schema{"text": {Type: "string"}}},
		Metadata: map[string]interface{}{"direction": "send"}}
	ex := g.Generate(res, "websocket", "wss://api.example.com/ws")
	if !strings.Contains(ex["javascript"], "new WebSocket") || !strings.Contains(ex["javascript"], "ws.send") {
		t.Errorf("send example should open and send: %q", ex["javascript"])
	}
	if !strings.Contains(ex["wscat"], "wscat -c wss://api.example.com/ws") {
		t.Errorf("wscat example malformed: %q", ex["wscat"])
	}
}

func TestGenerateWebSocketReceive(t *testing.T) {
	g := newGen()
	res := &ir.Resource{Name: "message.received", Method: "receive",
		Metadata: map[string]interface{}{"direction": "receive"}}
	ex := g.Generate(res, "websocket", "wss://x/ws")
	if !strings.Contains(ex["javascript"], "addEventListener('message'") {
		t.Errorf("receive example should listen for messages: %q", ex["javascript"])
	}
}

func TestGenerateMessaging(t *testing.T) {
	g := newGen()
	res := &ir.Resource{Name: "orders.events", Method: "publish",
		Request:  &ir.Schema{Type: "object", Properties: map[string]*ir.Schema{"id": {Type: "string"}}},
		Metadata: map[string]interface{}{"operation_type": "publish", "channel": "orders.events"}}
	ex := g.Generate(res, "asyncapi", "")
	if _, ok := ex["payload"]; !ok {
		t.Error("messaging example should include a payload")
	}
	if !strings.Contains(ex["javascript"], "publish") || !strings.Contains(ex["javascript"], "orders.events") {
		t.Errorf("publish snippet malformed: %q", ex["javascript"])
	}
}

func TestGenerateMCPTool(t *testing.T) {
	g := newGen()
	res := &ir.Resource{Name: "create_order",
		Request:  &ir.Schema{Type: "object", Properties: map[string]*ir.Schema{"sku": {Type: "string"}}},
		Metadata: map[string]interface{}{"type": "mcp-tool"}}
	ex := g.Generate(res, "mcp", "")
	if !strings.Contains(ex["json"], "tools/call") || !strings.Contains(ex["json"], "create_order") {
		t.Errorf("MCP tool JSON-RPC malformed: %q", ex["json"])
	}
	if !strings.Contains(ex["javascript"], "callTool") {
		t.Errorf("MCP client call missing: %q", ex["javascript"])
	}
}

func TestGenerateMCPResource(t *testing.T) {
	g := newGen()
	res := &ir.Resource{Name: "recent orders", Path: "orders://recent",
		Metadata: map[string]interface{}{"type": "mcp-resource"}}
	ex := g.Generate(res, "mcp", "")
	if !strings.Contains(ex["json"], "resources/read") || !strings.Contains(ex["json"], "orders://recent") {
		t.Errorf("MCP resource JSON-RPC malformed: %q", ex["json"])
	}
}
