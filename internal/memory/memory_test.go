package memory_test

import (
	"errors"
	"testing"

	"github.com/PRNVBAJAJ/context-os/internal/memory"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

func TestNew_ValidInputs(t *testing.T) {
	m, err := memory.New("auth-strategy", "Auth Strategy", "We use JWT with RS256.")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if m.ID.IsEmpty() {
		t.Error("ID should not be empty")
	}
	if m.Key != "auth-strategy" {
		t.Errorf("Key = %q, want %q", m.Key, "auth-strategy")
	}
	if m.Title != "Auth Strategy" {
		t.Errorf("Title = %q, want %q", m.Title, "Auth Strategy")
	}
	if m.Content != "We use JWT with RS256." {
		t.Errorf("Content = %q, want %q", m.Content, "We use JWT with RS256.")
	}
	if m.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestNew_TitleDefaultsToKey(t *testing.T) {
	m, err := memory.New("db-schema", "", "The schema uses UUID primary keys.")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if m.Title != "db-schema" {
		t.Errorf("Title = %q, want key %q", m.Title, "db-schema")
	}
}

func TestNew_InvalidKey(t *testing.T) {
	cases := []struct {
		key  string
		desc string
	}{
		{"", "empty key"},
		{"Auth Strategy", "key with spaces"},
		{"auth/strategy", "key with slash"},
		{"Auth-Strategy", "key with uppercase"},
		{"auth_strategy", "key with underscore"},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := memory.New(tc.key, "", "some content")
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var domainErr *shared.Error
			if !errors.As(err, &domainErr) {
				t.Fatalf("expected *shared.Error, got %T", err)
			}
			if domainErr.Code != shared.CodeInvalidInput {
				t.Errorf("Code = %q, want %q", domainErr.Code, shared.CodeInvalidInput)
			}
		})
	}
}

func TestNew_EmptyContent(t *testing.T) {
	_, err := memory.New("auth-strategy", "Title", "")
	if err == nil {
		t.Fatal("expected error for empty content, got nil")
	}
	var domainErr *shared.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected *shared.Error, got %T", err)
	}
	if domainErr.Code != shared.CodeInvalidInput {
		t.Errorf("Code = %q, want %q", domainErr.Code, shared.CodeInvalidInput)
	}
}

func TestNew_ValidKeyFormats(t *testing.T) {
	validKeys := []string{"a", "auth", "auth-strategy", "db-schema-v2", "123", "a1b2-c3"}
	for _, key := range validKeys {
		t.Run(key, func(t *testing.T) {
			if _, err := memory.New(key, "", "content"); err != nil {
				t.Errorf("New(%q) should succeed: %v", key, err)
			}
		})
	}
}
