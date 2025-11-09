package helper

import (
	"strconv"
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	id := GenerateID("test")
	if !strings.HasPrefix(id, "test_") {
		t.Fatalf("expected prefix 'test_', got %q", id)
	}

	numeric := strings.TrimPrefix(id, "test_")
	if _, err := strconv.ParseInt(numeric, 10, 64); err != nil {
		t.Fatalf("expected numeric suffix, got %q: %v", numeric, err)
	}
}

func TestNewIDGeneratorWithPrefix(t *testing.T) {
	g := NewIDGenerator(1, "foo")
	id := g.GenerateID()
	if !strings.HasPrefix(id, "foo_") {
		t.Fatalf("expected prefix 'foo_', got %q", id)
	}

	numeric := strings.TrimPrefix(id, "foo_")
	if _, err := strconv.ParseInt(numeric, 10, 64); err != nil {
		t.Fatalf("expected numeric suffix, got %q: %v", numeric, err)
	}
}

func TestIDGeneratorAdditionalPrefixes(t *testing.T) {
	g := NewIDGenerator(1, "foo")
	id := g.GenerateID("bar", "baz")
	expectedPrefix := "foo_bar_baz_"
	if !strings.HasPrefix(id, expectedPrefix) {
		t.Fatalf("expected prefix %q, got %q", expectedPrefix, id)
	}

	numeric := strings.TrimPrefix(id, expectedPrefix)
	if _, err := strconv.ParseInt(numeric, 10, 64); err != nil {
		t.Fatalf("expected numeric suffix, got %q: %v", numeric, err)
	}
}

func TestGetGeneratorInvalidSeedPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid seed")
		}
	}()

	_ = GetGenerator(-1)
}
