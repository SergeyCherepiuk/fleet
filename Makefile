manager:
	go run -race cli/main.go manager

worker:
	go run -race cli/main.go worker --manager $(manager)

task-run:
	go run -race cli/main.go task run --manager $(manager) $(manifest)
