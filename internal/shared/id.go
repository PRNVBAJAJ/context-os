package shared

import "github.com/google/uuid"

// ID is the canonical identifier type for all Context OS domain objects.
// It is represented as a UUID v4 string and is safe for use as a map key.
type ID string

// EmptyID is the zero value of ID.
const EmptyID ID = ""

// NewID returns a new random, globally unique ID.
func NewID() ID {
	return ID(uuid.New().String())
}

// IsEmpty reports whether the ID is the zero value.
func (id ID) IsEmpty() bool {
	return id == EmptyID
}

// String returns the string representation of the ID.
func (id ID) String() string {
	return string(id)
}
