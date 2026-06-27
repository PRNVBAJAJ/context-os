package memory

import (
	"regexp"
	"time"

	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

// keyPattern enforces slug-style keys: lowercase letters, digits, hyphens only.
var keyPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// Memory is a named, durable piece of project knowledge.
// Keys are unique within a project; content is the markdown body.
type Memory struct {
	ID        shared.ID
	Key       string
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// New validates inputs and returns an initialized Memory.
// Title defaults to Key when empty.
func New(key, title, content string) (*Memory, error) {
	if key == "" {
		return nil, shared.NewError(shared.CodeInvalidInput, "memory key must not be empty")
	}
	if !keyPattern.MatchString(key) {
		return nil, shared.NewError(shared.CodeInvalidInput, "memory key must match [a-z0-9-]+ (lowercase letters, digits, hyphens)")
	}
	if content == "" {
		return nil, shared.NewError(shared.CodeInvalidInput, "memory content must not be empty")
	}
	if title == "" {
		title = key
	}
	now := time.Now().UTC()
	return &Memory{
		ID:        shared.NewID(),
		Key:       key,
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
