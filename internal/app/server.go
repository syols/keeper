package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"keeper/config"
	"keeper/internal/pkg"
)

// Server struct
type Server struct {
	ctx         context.Context
	authorizer  pkg.Authorizer
	grpcService pkg.GrpcService
	settings    config.Config
}

// NewServer creates server struct
func NewServer(ctx context.Context, settings config.Config) (Server, error) {
	grpcService, err := pkg.NewGrpcService(ctx, settings)
	if err != nil {
		return Server{}, err
	}
	return Server{
		ctx:         ctx,
		grpcService: grpcService,
		settings:    settings,
	}, nil
}

func (s *Server) Run() {
	ctx, cancel := signal.NotifyContext(s.ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	s.shutdown(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := s.grpcService.Run(s.settings.Server.Address.Port)
		if err != nil {
			log.Println(err.Error())
		}
	}()
	wg.Wait()
}

func (s *Server) shutdown(ctx context.Context) {
	go func() {
		<-ctx.Done()
		s.grpcService.Shutdown()
		os.Exit(0)
	}()
}
