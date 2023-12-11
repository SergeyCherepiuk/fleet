package task

import "slices"

type State string

const (
	Pending   State = "pending"
	Scheduled State = "scheduled"
	Running   State = "running"
	Finished  State = "finished"
	Failed    State = "failed"
)

type ErrStateTransitionNotAllowed error

var allowedTransitions = map[State][]State{
	Pending:   {Scheduled, Failed},
	Scheduled: {Running, Failed},
	Running:   {Finished, Failed},
	Finished:  {},
	Failed:    {},
}

func IsStateTransitionAllowed(from State, to State) bool {
	allowed, ok := allowedTransitions[from]
	return ok && slices.Contains(allowed, to)
}
