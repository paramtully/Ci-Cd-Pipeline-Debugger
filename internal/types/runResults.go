package types

type RunResults struct {
	passed bool
	failedStep string
	exitCode int
	tail string
}