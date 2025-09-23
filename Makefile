.PHONY: run-server
run-server:
	MONIC_SERVER_ADDR=:8000 \
	#MONIC_SERVER_CLICKHOUSE_DSN="tcp://127.0.0.1:9000?database=monic_db&username=default&password=default" \
	MONIC_SERVER_SHARED_SECRET=secret \
	go run ./cmd/monic-server

.PHONY: run-agent
run-agent:
	MONIC_WEBHOOK_URL=http://127.0.0.1:8000/webhook \
	MONIC_SHARED_SECRET=secret \
	go run ./cmd/monic-agent

.PHONY: build
build:
	CGO_ENABLED=1 go build -o build/monic-server ./cmd/monic-server \
	&& CGO_ENABLED=1 go build -o build/monic-agent ./cmd/monic-agent
