package asyncapi

import (
	"testing"

	"github.com/sirrobot01/unifydoc/internal/ir"
)

// AsyncAPI 2.x with a $ref from the operation message to components.
const spec2x = `
asyncapi: 2.6.0
info:
  title: Account Service
  version: 1.0.0
  description: Emits user events
servers:
  production:
    url: broker.example.com:9092
    protocol: kafka
    description: Kafka broker
channels:
  user/signedup:
    description: User signup channel
    subscribe:
      summary: Receive user signups
      message:
        $ref: '#/components/messages/UserSignedUp'
    publish:
      summary: Send user signups
      message:
        $ref: '#/components/messages/UserSignedUp'
components:
  messages:
    UserSignedUp:
      name: UserSignedUp
      contentType: application/json
      payload:
        $ref: '#/components/schemas/User'
  schemas:
    User:
      type: object
      required: [id]
      properties:
        id:
          type: string
        email:
          type: string
`

// AsyncAPI 3.x with operations referencing channels and messages.
const spec3x = `
asyncapi: 3.0.0
info:
  title: Order Service
  version: 2.0.0
servers:
  prod:
    host: broker.example.com:9092
    protocol: kafka
channels:
  orderEvents:
    address: order.events
    description: Order events channel
    messages:
      OrderCreated:
        $ref: '#/components/messages/OrderCreated'
operations:
  sendOrder:
    action: send
    channel:
      $ref: '#/channels/orderEvents'
    messages:
      - $ref: '#/channels/orderEvents/messages/OrderCreated'
components:
  messages:
    OrderCreated:
      name: OrderCreated
      payload:
        $ref: '#/components/schemas/Order'
  schemas:
    Order:
      type: object
      properties:
        orderId:
          type: string
        amount:
          type: number
`

func TestMetadata(t *testing.T) {
	p := NewPlugin()
	if p.Name() != "asyncapi" {
		t.Errorf("Name() = %q, want asyncapi", p.Name())
	}
	if p.Version() == "" {
		t.Error("Version() empty")
	}
}

func TestValidate(t *testing.T) {
	p := NewPlugin()
	tests := []struct {
		name    string
		spec    string
		wantErr bool
	}{
		{"valid 2.x", spec2x, false},
		{"valid 3.x", spec3x, false},
		{"missing version", "info:\n  title: X\n  version: 1.0.0\n", true},
		{"missing title", "asyncapi: 2.6.0\ninfo:\n  version: 1.0.0\n", true},
		{"not yaml", "\t\tbroken: [", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := p.Validate([]byte(tt.spec)); (err != nil) != tt.wantErr {
				t.Errorf("Validate() err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestParse2xResolvesRefs(t *testing.T) {
	p := NewPlugin()
	result, err := p.Parse([]byte(spec2x))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if result.Title != "Account Service" || result.Version != "1.0.0" {
		t.Errorf("info not parsed: %q %q", result.Title, result.Version)
	}
	if len(result.Servers) != 1 || result.Servers[0].URL != "broker.example.com:9092" {
		t.Errorf("server not parsed: %+v", result.Servers)
	}
	// subscribe + publish => 2 resources
	if len(result.Resources) != 2 {
		t.Fatalf("resources = %d, want 2", len(result.Resources))
	}
	sub := findByOp(result, "subscribe")
	if sub == nil {
		t.Fatal("subscribe operation missing")
	}
	if sub.Metadata["channel"] != "user/signedup" {
		t.Errorf("channel = %v", sub.Metadata["channel"])
	}
	// The $ref chain message -> payload -> schema must resolve to real fields.
	if sub.Request == nil || sub.Request.Properties["email"] == nil {
		t.Fatalf("payload $ref not resolved: %+v", sub.Request)
	}
	if sub.Metadata["message_name"] != "UserSignedUp" {
		t.Errorf("message_name = %v", sub.Metadata["message_name"])
	}
	// components.schemas exposed as a type
	if findType(result, "User") == nil {
		t.Error("User schema should be a documented type")
	}
}

func TestParse3xOperations(t *testing.T) {
	p := NewPlugin()
	result, err := p.Parse([]byte(spec3x))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if result.Metadata["asyncapi_version"] != "3.0.0" {
		t.Errorf("version metadata = %v", result.Metadata["asyncapi_version"])
	}
	if len(result.Servers) != 1 || result.Servers[0].URL != "broker.example.com:9092" {
		t.Errorf("3.x server host+pathname: %+v", result.Servers)
	}
	if len(result.Resources) != 1 {
		t.Fatalf("resources = %d, want 1", len(result.Resources))
	}
	op := result.Resources[0]
	if op.Metadata["operation_type"] != "publish" { // send -> publish
		t.Errorf("send action should normalize to publish, got %v", op.Metadata["operation_type"])
	}
	if op.Metadata["action"] != "send" {
		t.Errorf("original action lost: %v", op.Metadata["action"])
	}
	if op.Metadata["channel"] != "orderEvents" {
		t.Errorf("channel = %v", op.Metadata["channel"])
	}
	if op.Request == nil || op.Request.Properties["orderId"] == nil {
		t.Fatalf("3.x message payload not resolved: %+v", op.Request)
	}
}

func findByOp(result *ir.IR, opType string) *ir.Resource {
	for i := range result.Resources {
		if result.Resources[i].Metadata["operation_type"] == opType {
			return &result.Resources[i]
		}
	}
	return nil
}

func findType(result *ir.IR, name string) *ir.TypeDef {
	for i := range result.Types {
		if result.Types[i].Name == name {
			return &result.Types[i]
		}
	}
	return nil
}
