package container

type RestartPolicy int

const (
	Always RestartPolicy = iota
	OnFailure
	UnlessStopped
	Never
)
