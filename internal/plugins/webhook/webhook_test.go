package webhook

import "testing"

const spec = `
webhooks:
  - name: order.created
    description: Fired when an order is created
    method: POST
    url: https://example.com/hooks/order
    headers:
      - name: X-Signature
        description: HMAC signature
        required: true
    payload:
      type: object
      required: [orderId]
      properties:
        orderId:
          type: string
    example:
      orderId: ord_123
`

func TestMetadata(t *testing.T) {
	p := NewPlugin()
	if p.Name() != "webhook" {
		t.Errorf("Name() = %q", p.Name())
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
		{"valid", spec, false},
		{"missing name", "webhooks:\n  - method: POST\n", true},
		{"missing method", "webhooks:\n  - name: x\n", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := p.Validate([]byte(tt.spec)); (err != nil) != tt.wantErr {
				t.Errorf("Validate() err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestParse(t *testing.T) {
	p := NewPlugin()
	result, err := p.Parse([]byte(spec))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(result.Resources) != 1 {
		t.Fatalf("resources = %d, want 1", len(result.Resources))
	}
	r := result.Resources[0]
	if r.Method != "POST" || r.Path != "https://example.com/hooks/order" {
		t.Errorf("method/url wrong: %q %q", r.Method, r.Path)
	}
	if r.Metadata["type"] != "webhook" {
		t.Errorf("type metadata = %v", r.Metadata["type"])
	}
	if len(r.Parameters) != 1 || r.Parameters[0].In != "header" || !r.Parameters[0].Required {
		t.Errorf("header param wrong: %+v", r.Parameters)
	}
	if r.Request == nil || r.Request.Properties["orderId"] == nil {
		t.Errorf("payload not converted: %+v", r.Request)
	}
	if len(r.Examples) != 1 {
		t.Errorf("expected 1 example, got %d", len(r.Examples))
	}
}
