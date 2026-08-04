package types

type Job struct {
	name string
	image string
	steps []Step
}