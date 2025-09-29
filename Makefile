.PHONY: run-server-http
run-server-http:
	MONIC_SERVER_HTTP_ADDR=:8000 \
	MONIC_SERVER_CLICKHOUSE_DSN="tcp://127.0.0.1:9000?database=monic_db&" \
	MONIC_SECRET=secret \
	go run ./cmd/monic-server

.PHONY: run-server-grpc
run-server-grpc:
	MONIC_SERVER_GRPC_ADDR=:50051 \
	MONIC_SERVER_CLICKHOUSE_DSN="tcp://127.0.0.1:9000?database=monic_db" \
	MONIC_SECRET=secret \
	go run ./cmd/monic-server

.PHONY: run-agent-http
run-agent-http:
	MONIC_HTTP_URL=http://127.0.0.1:8000 \
	MONIC_SECRET=secret \
	go run ./cmd/monic-agent

.PHONY: run-agent-grpc
run-agent-grpc:
	MONIC_GRPC_ADDR=127.0.0.1:50051 \
 	MONIC_GRPC_INSECURE=true \
 	MONIC_SECRET=secret \
	go run ./cmd/monic-agent

.PHONY: gen
gen:
	protoc --proto_path=./api/proto \
	   --go_out=paths=source_relative:./api/pb \
	   --go-grpc_out=paths=source_relative:./api/pb \
	   ./api/proto/*.proto

.PHONY: build
build:
	CGO_ENABLED=1 go build -o build/monic-server ./cmd/monic-server \
	&& CGO_ENABLED=1 go build -o build/monic-agent ./cmd/monic-agent
