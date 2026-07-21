package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirrobot01/unifydoc/internal/plugins/grpc"
	"github.com/sirrobot01/unifydoc/internal/plugins/webhook"
)

func TestParseProtocolSpecDirMerges(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("a.yaml", "webhooks:\n  - {name: order.created, method: POST, url: https://x/o}\n")
	write("b.yml", "webhooks:\n  - {name: order.refunded, method: POST, url: https://x/r}\n")
	write("ignore.txt", "not a spec")

	result, err := parseProtocolSpec(webhook.NewPlugin(), dir)
	if err != nil {
		t.Fatalf("parseProtocolSpec(dir) error: %v", err)
	}
	if len(result.Resources) != 2 {
		t.Fatalf("merged resources = %d, want 2", len(result.Resources))
	}
}

func TestParseProtocolSpecEmptyDir(t *testing.T) {
	if _, err := parseProtocolSpec(webhook.NewPlugin(), t.TempDir()); err == nil {
		t.Error("empty directory should error")
	}
}

func TestParseProtocolSpecUsesPathParser(t *testing.T) {
	// grpc implements PathParser, so a directory of .proto is handled by it,
	// not the generic yaml/json dir merge.
	dir := t.TempDir()
	proto := "syntax = \"proto3\";\npackage p;\nservice S { rpc M (Req) returns (Resp); }\nmessage Req { string a = 1; }\nmessage Resp { string b = 1; }\n"
	if err := os.WriteFile(filepath.Join(dir, "s.proto"), []byte(proto), 0644); err != nil {
		t.Fatal(err)
	}
	result, err := parseProtocolSpec(grpc.NewPlugin(), dir)
	if err != nil {
		t.Fatalf("parseProtocolSpec via PathParser error: %v", err)
	}
	if result.Protocol != "grpc" || len(result.Resources) != 1 {
		t.Fatalf("unexpected result: protocol=%s resources=%d", result.Protocol, len(result.Resources))
	}
}
