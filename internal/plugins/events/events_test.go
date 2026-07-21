package events

import "testing"

const spec = `
events:
  title: Order Events
  version: 2.0.0
  description: Domain events
  topics:
    - name: orders
      description: Order lifecycle
      events:
        - name: order.created
          type: com.shop.order.created
          description: An order was created
          schema:
            type: object
            required: [orderId]
            properties:
              orderId:
                type: string
              amount:
                type: number
          example:
            orderId: ord_1
            amount: 42
  eventTypes:
    - name: system.heartbeat
      type: com.shop.system.heartbeat
      description: Liveness ping
`

func TestMetadata(t *testing.T) {
	p := NewPlugin()
	if p.Name() != "events" {
		t.Errorf("Name() = %q", p.Name())
	}
	if p.Version() == "" {
		t.Error("Version() empty")
	}
}

func TestValidate(t *testing.T) {
	p := NewPlugin()
	if err := p.Validate([]byte(spec)); err != nil {
		t.Errorf("valid spec rejected: %v", err)
	}
	if err := p.Validate([]byte("\tnot: [yaml")); err == nil {
		t.Error("malformed YAML should fail validation")
	}
}

func TestParse(t *testing.T) {
	p := NewPlugin()
	result, err := p.Parse([]byte(spec))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if result.Title != "Order Events" || result.Version != "2.0.0" {
		t.Errorf("header wrong: %q %q", result.Title, result.Version)
	}
	// 1 topic event + 1 standalone event
	if len(result.Resources) != 2 {
		t.Fatalf("resources = %d, want 2", len(result.Resources))
	}

	created := result.Resources[0]
	if created.Method != "EVENT" {
		t.Errorf("method = %q, want EVENT", created.Method)
	}
	if created.Metadata["event_type"] != "com.shop.order.created" {
		t.Errorf("event_type = %v", created.Metadata["event_type"])
	}
	if created.Metadata["topic"] != "orders" {
		t.Errorf("topic = %v", created.Metadata["topic"])
	}
	if len(created.Tags) != 1 || created.Tags[0] != "orders" {
		t.Errorf("tags = %v, want [orders]", created.Tags)
	}
	if created.Request == nil || created.Request.Properties["orderId"] == nil {
		t.Errorf("schema not converted: %+v", created.Request)
	}
	if len(created.Examples) != 1 {
		t.Errorf("expected 1 example, got %d", len(created.Examples))
	}

	// Standalone event has no topic tag.
	standalone := result.Resources[1]
	if _, ok := standalone.Metadata["topic"]; ok {
		t.Error("standalone event should not have a topic")
	}
}
