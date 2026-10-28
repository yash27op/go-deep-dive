package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestHashAll(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.txt")
	if err := os.WriteFile(path, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	results, err := HashAll(context.Background(), []string{path, path + "missing"}, 2)
	if err != nil || len(results) != 2 {
		t.Fatalf("results=%v err=%v", results, err)
	}
	if results[0].SHA256 != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Fatalf("bad hash: %+v", results[0])
	}
	if results[1].Error == "" {
		t.Fatal("missing file should fail independently")
	}
}

func TestCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := HashAll(ctx, []string{"a", "b"}, 2); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestInvalidWorkers(t *testing.T) {
	if _, err := HashAll(context.Background(), nil, 0); err == nil {
		t.Fatal("expected validation error")
	}
}
