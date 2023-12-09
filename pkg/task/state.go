package task

type State int

const (
	Pending State = iota
	Failed
	Scheduled
	Running
	Finished
)
