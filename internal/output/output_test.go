package output

import (
	"os"
	"testing"
)

// TestNewOutput_truncatesFile ensures that creating a new Output with an existing file truncates it.
func TestNewOutput_truncatesFile(t *testing.T) {
	path := "test_output.txt"
	// Create file with some data
	err := os.WriteFile(path, []byte("old-data"), 0o644)
	if err != nil {
		t.Fatalf("failed to prepare test file: %v", err)
	}

	o, err := NewOutput(path, false)
	if err != nil {
		t.Fatalf("failed to create output: %v", err)
	}
	defer func() {
		_ = o.receivedDataFile.Close()
		_ = os.Remove(path)
	}()

	// Write new content
	_, err = o.receivedDataFile.WriteString("new")
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	_ = o.receivedDataFile.Close()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}
	if string(b) != "new" {
		t.Fatalf("file not truncated, got: %q", string(b))
	}
}
