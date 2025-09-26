package grpcsender

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/magomedcoder/monic/api/pb"
	"github.com/magomedcoder/monic/internal/monic-agent/domain"
	"github.com/magomedcoder/monic/internal/monic-agent/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type GRPCSender struct {
	cc     *grpc.ClientConn
	c      pb.EventsServiceClient
	secret string
}

func NewGRPCSender(addr string, secret string, insecureTLS bool) (ports.EventSender, error) {
	var dialOpt grpc.DialOption
	if insecureTLS {
		dialOpt = grpc.WithTransportCredentials(insecure.NewCredentials())
	} else {
		dialOpt = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			MinVersion: tls.VersionTLS12,
		}))
	}

	cc, err := grpc.Dial(addr, dialOpt, grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}

	return &GRPCSender{
		cc:     cc,
		c:      pb.NewEventsServiceClient(cc),
		secret: secret,
	}, nil
}

func (g *GRPCSender) Send(ctx context.Context, ev *domain.Event) error {
	req := &pb.IngestRequest{
		Event: &pb.Event{
			DateTime: timestamppb.New(ev.DateTime),
			Server:   ev.Server,
			Type:     ev.Type,
			User:     ev.User,
			RemoteIp: ev.RemoteIP,
			Port:     ev.Port,
			Method:   ev.Method,
			Message:  ev.Message,
			Raw:      ev.Raw,
		},
	}

	if g.secret != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+g.secret)
	}

	_, err := g.c.Ingest(ctx, req)
	return err
}

func (g *GRPCSender) Close() error {
	return g.cc.Close()
}
