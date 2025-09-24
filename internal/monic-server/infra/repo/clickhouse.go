package repo

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/magomedcoder/monic/internal/monic-server/app"
	"github.com/magomedcoder/monic/internal/monic-server/config"
	"github.com/magomedcoder/monic/internal/monic-server/domain"
)

type ClickHouse struct {
	conn clickhouse.Conn
}

func NewClickHouse(ctx context.Context, cfg config.Config) (*ClickHouse, error) {
	addr := cfg.ClickHouseDSN

	opts, err := clickhouse.ParseDSN(addr)
	if err != nil {
		return nil, err
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	return &ClickHouse{
		conn: conn,
	}, nil
}

func (c *ClickHouse) Ensure(ctx context.Context) error {
	return c.conn.Exec(ctx, fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS ssh_events(
		ts DateTime64(3),
		server String,
		type LowCardinality(String),
		user String,
		remote_ip String,
		port UInt16,
		method LowCardinality(String),
		message LowCardinality(String),
		raw String,
		received_at DateTime()
	) ENGINE = MergeTree
	ORDER BY (ts, server)
	SETTINGS index_granularity = 8192;`))
}

func (c *ClickHouse) InsertBatch(ctx context.Context, batch []domain.IngestedEvent) error {
	b, err := c.conn.PrepareBatch(ctx, "INSERT INTO ssh_events (ts, server, type, user, remote_ip, port, method, message, raw, received_at)")
	if err != nil {
		return err
	}

	for _, e := range batch {
		if err := b.Append(e.TS, e.Server, e.Type, e.User, e.RemoteIP, app.ParsePort(e.Port), e.Method, e.Message, e.Raw, e.ReceivedAt); err != nil {
			return err
		}
	}

	return b.Send()
}

func (c *ClickHouse) Close() error {
	return c.conn.Close()
}
