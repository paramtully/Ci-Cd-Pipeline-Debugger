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
	// EFFECTS: runs steps in order, returns RunResult and error if any step fails
	Run(ctx context.Context, job types.Job) (types.RunResults, error)
}
