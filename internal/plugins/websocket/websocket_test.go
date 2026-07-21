package websocket

import (
	"testing"
)

const validSpec = `
websocket:
  url: wss://api.example.com/ws/chat
  description: Real-time chat
  version: 2.0.0
  events:
    - name: message.send
      direction: send
      description: Send a message
      payload:
        type: object
        properties:
          text:
            type: string
          roomId:
            type: string
        required:
          - text
          - roomId
      example:
        text: "hi"
        roomId: "room-1"
    - name: message.received
      direction: receive
      description: Receive a message
`

func TestPluginMetadata(t *testing.T) {
	p := NewPlugin()
	if p.Name() != "websocket" {
		t.Errorf("Name() = %q, want %q", p.Name(), "websocket")
	}
	if p.Version() == "" {
		t.Error("Version() should not be empty")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantErr bool
	}{
		{"valid", validSpec, false},
		{
			name:    "missing url",
			spec:    "websocket:\n  description: no url\n  events: []\n",
			wantErr: true,
		},
		{
			name:    "invalid direction",
			spec:    "websocket:\n  url: wss://x\n  events:\n    - name: e\n      direction: sideways\n",
			wantErr: true,
		},
		{
			name:    "event missing name",
			spec:    "websocket:\n  url: wss://x\n  events:\n    - direction: send\n",
			wantErr: true,
		},
		{
			name:    "not yaml",
			spec:    "\t\tnot: [valid",
			wantErr: true,
		},
	}

	p := NewPlugin()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.Validate([]byte(tt.spec))
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParse(t *testing.T) {
	p := NewPlugin()
	result, err := p.Parse([]byte(validSpec))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	if result.Protocol != "websocket" {
		t.Errorf("Protocol = %q, want %q", result.Protocol, "websocket")
	}
	if result.Version != "2.0.0" {
		t.Errorf("Version = %q, want %q", result.Version, "2.0.0")
	}
	if len(result.Servers) != 1 || result.Servers[0].URL != "wss://api.example.com/ws/chat" {
		t.Errorf("Servers = %+v, want the chat URL", result.Servers)
	}

	// Two events => two resources.
	if len(result.Resources) != 2 {
		t.Fatalf("Resources = %d, want 2", len(result.Resources))
	}

	send := result.Resources[0]
	if send.Name != "message.send" {
		t.Errorf("Resource[0].Name = %q, want message.send", send.Name)
	}
	if send.Metadata["direction"] != "send" {
		t.Errorf("direction metadata = %v, want send", send.Metadata["direction"])
	}
	if send.Request == nil {
		t.Fatal("send event should have a request schema from its payload")
	}
	if _, ok := send.Request.Properties["text"]; !ok {
		t.Error("send payload schema should include a 'text' property")
	}
	if len(send.Request.Required) != 2 {
		t.Errorf("send payload required = %v, want [text roomId]", send.Request.Required)
	}
	if len(send.Examples) != 1 {
		t.Errorf("send event should carry 1 example, got %d", len(send.Examples))
	}
}

func TestParseVersionDefault(t *testing.T) {
	p := NewPlugin()
	result, err := p.Parse([]byte("websocket:\n  url: wss://x\n  events: []\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if result.Version != "1.0.0" {
		t.Errorf("Version = %q, want default 1.0.0", result.Version)
	}
}
