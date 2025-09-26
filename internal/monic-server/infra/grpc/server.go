package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/magomedcoder/monic/api/pb"
	"github.com/magomedcoder/monic/internal/monic-server/config"
	"github.com/magomedcoder/monic/internal/monic-server/domain"
	"github.com/magomedcoder/monic/internal/monic-server/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net"
	"strings"
)

type Server struct {
	cfg config.Config
	app ports.Enqueuer
}

func NewServer(cfg config.Config, app ports.Enqueuer) *Server {
	return &Server{
		cfg: cfg,
		app: app,
	}
}

func (s *Server) Start(ctx context.Context) error {
	if s.cfg.GRPCAddr == "" {
		return nil
	}

	lis, err := net.Listen("tcp", s.cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("grpc listen: %w", err)
	}

	var opts []grpc.ServerOption
	if s.cfg.TLSCertFile != "" && s.cfg.TLSKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(s.cfg.TLSCertFile, s.cfg.TLSKeyFile)
		if err != nil {
			return fmt.Errorf("load tls certs: %w", err)
		}
		opts = append(opts, grpc.Creds(credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		})))
	} else {
		opts = append(opts, grpc.Creds(insecure.NewCredentials()))
	}

	gs := grpc.NewServer(opts...)
	pb.RegisterEventsServiceServer(gs, &ingestSvc{
		app:          s.app,
		sharedSecret: s.cfg.SharedSecret,
	})

	go func() {
		<-ctx.Done()
		gs.GracefulStop()
	}()

	log.Printf("[Monic] gRPC listening on %s (tls=%v)", s.cfg.GRPCAddr, s.cfg.TLSCertFile != "")

	if err := gs.Serve(lis); err != nil {
		return fmt.Errorf("grpc serve: %w", err)
	}

	return nil
}

type ingestSvc struct {
	pb.UnimplementedEventsServiceServer
	app          ports.Enqueuer
	sharedSecret string
}

func (s *ingestSvc) Ingest(ctx context.Context, req *pb.IngestRequest) (*pb.IngestResponse, error) {
	if s.sharedSecret != "" {
		md, _ := metadata.FromIncomingContext(ctx)
		auth := ""
		if vals := md.Get("authorization"); len(vals) > 0 {
			auth = vals[0]
		}
		const pfx = "Bearer "
		if !strings.HasPrefix(auth, pfx) || strings.TrimPrefix(auth, pfx) != s.sharedSecret {
			return nil, fmt.Errorf("unauthorized")
		}
	}

	ev := req.GetEvent()
	if ev == nil {
		return nil, fmt.Errorf("empty event")
	}

	if dateTime := ev.GetDateTime(); dateTime == nil || dateTime.AsTime().IsZero() || dateTime.AsTime().After(timestamppb.Now().AsTime().AddDate(0, 0, 1)) {
		return nil, fmt.Errorf("bad ts")
	}

	d := domain.Event{
		DateTime: ev.GetDateTime().AsTime(),
		Server:   ev.GetServer(),
		Type:     ev.GetType(),
		User:     ev.GetUser(),
		RemoteIP: ev.GetRemoteIp(),
		Port:     ev.GetPort(),
		Method:   ev.GetMethod(),
		Message:  ev.GetMessage(),
		Raw:      ev.GetRaw(),
	}

	if err := s.app.Enqueue(d); err != nil {
		return nil, fmt.Errorf("enqueue: %w", err)
	}

	return &pb.IngestResponse{}, nil
}
