package main

// Config holds CLI options parsed once in cmd and passed into the run loop.
type Config struct {
	RepoDir       string
	Workflow      string
	Job           string
	Image         string
	EnvFile       string
	AI            bool
	MaxIterations int
	Verbose       bool
}
