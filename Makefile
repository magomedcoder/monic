.PHONY: run-agent
run-agent:
	MONIC_WEBHOOK_URL=http://127.0.0.1:8000/webhook MONIC_SHARED_SECRET=secret go run ./cmd/monic-agent

.PHONY: build
build:
	CGO_ENABLED=1 go build -o build/monic-agent ./cmd/monic-agent