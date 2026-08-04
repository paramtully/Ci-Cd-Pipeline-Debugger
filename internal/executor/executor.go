// Package executor runs CI jobs in Docker behind the Executor interface.
package executor

import (
	"context"

	"github.com/paramtully/Ci-Cd-Pipeline-Debugger/internal/types"
)

type Executor interface {
	// Executes a job step and returns the outcome (including capped log tail on failure).
	// REQUIRES: 
	// MODIFIES: docker instance
	// EFFECTS: runs the job; returns RunResults for the attempt;
	// returns a non-nil error only on infrastructure/invocation failure (not CI step failure)
	Run(ctx context.Context, job types.Job) (types.RunResults, error)
}
