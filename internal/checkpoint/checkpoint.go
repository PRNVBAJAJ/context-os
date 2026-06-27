package checkpoint

import (
	"time"

	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

// Checkpoint is an immutable recovery snapshot capturing the state of a project
// or workflow at a specific point in time. Checkpoints are append-only.
type Checkpoint struct {
	ID         shared.ID
	WorkflowID shared.ID // EmptyID for project-level checkpoints
	Note       string    // human description of current state; may be empty
	CreatedAt  time.Time
}

// New returns a Checkpoint with a generated ID and the current UTC timestamp.
// workflowID may be shared.EmptyID for project-level checkpoints.
func New(workflowID shared.ID, note string) *Checkpoint {
	return &Checkpoint{
		ID:         shared.NewID(),
		WorkflowID: workflowID,
		Note:       note,
		CreatedAt:  time.Now().UTC(),
	}
}
