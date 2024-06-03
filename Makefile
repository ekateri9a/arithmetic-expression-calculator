# Build for Windows ---------------------------------------------------------------

build-windows: build-orchestrator-windows build-agent-windows

build-orchestrator-windows:
	go build -o "main-orchestrator.exe" ./cmd/orchestrator/main.go

build-agent-windows:
	go build -o "main-agent.exe" ./cmd/agent/main.go


# Run for Windows ---------------------------------------------------------------

run-orchestrator-windows: build-orchestrator-windows
	./main-orchestrator.exe

run-agent-windows: build-agent-windows
	./main-agent.exe

# Build for Linux ---------------------------------------------------------------

build-linux: build-orchestrator-linux build-agent-linux

build-orchestrator-linux:
	go build -o "main-orchestrator" ./cmd/orchestrator/main.go

build-agent-linux:
	go build -o "main-agent" ./cmd/agent/main.go


# Run for Linux ---------------------------------------------------------------

run-orchestrator-linux: build-orchestrator-linux
	./main-orchestrator

run-agent-linux: build-agent-linux
	./main-agent


# Go run ---------------------------------------------------------------

run-orchestrator:
	go run ./cmd/orchestrator/main.go

run-agent:
	go run ./cmd/agent/main.go

