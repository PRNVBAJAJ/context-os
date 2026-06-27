package shared_test

import (
	"strings"
	"testing"

	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

func TestNewID_Format(t *testing.T) {
	id := shared.NewID()

	if id.IsEmpty() {
		t.Fatal("NewID should not return an empty ID")
	}

	// UUID v4 format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx (5 groups).
	parts := strings.Split(id.String(), "-")
	if len(parts) != 5 {
		t.Errorf("expected UUID v4 format with 5 groups, got %q", id)
	}
}

func TestNewID_Uniqueness(t *testing.T) {
	const count = 1000
	seen := make(map[shared.ID]struct{}, count)

	for range count {
		id := shared.NewID()
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate ID generated: %q", id)
		}
		seen[id] = struct{}{}
	}
}

func TestID_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		id   shared.ID
		want bool
	}{
		{"empty constant", shared.EmptyID, true},
		{"zero value literal", shared.ID(""), true},
		{"non-empty", shared.NewID(), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.id.IsEmpty(); got != tc.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestID_String(t *testing.T) {
	raw := "550e8400-e29b-41d4-a716-446655440000"
	id := shared.ID(raw)

	if id.String() != raw {
		t.Errorf("String() = %q, want %q", id.String(), raw)
	}
}
