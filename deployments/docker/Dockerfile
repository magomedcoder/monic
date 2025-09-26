FROM golang:1.24.7-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends build-essential pkg-config ca-certificates make \
    libsystemd-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /usr/src/monic

COPY . .

RUN go mod download

RUN --mount=type=cache,target=/root/.cache/go-build make build

FROM ubuntu:22.04 AS server

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/src/monic/build/monic-server /usr/local/bin/monic-server

EXPOSE 8000 50051

ENTRYPOINT ["/usr/local/bin/monic-server"]
