package container

type RestartPolicy string

const (
	Always        RestartPolicy = "always"
	OnFailure     RestartPolicy = "on-failure"
	UnlessStopped RestartPolicy = "unless-stopped"
	Never         RestartPolicy = "never"
)
