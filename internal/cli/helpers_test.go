package cli

import (
	"os"
	"testing"
)

// readExample loads an example spec file for use in tests, failing the test if
// it cannot be read.
func readExample(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read example %s: %v", path, err)
	}
	return data
}
