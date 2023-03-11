package server

import (
	"family-catering/pkg/logger"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/context"
)

type Server struct {
	srv             http.Server
	notify          chan error
	shutdownTimeout time.Duration
}

func New(address string, handler http.Handler, opts ...Option) *Server {
	srv := http.Server{
		Handler: handler,
		Addr:    address,
	}

	server := &Server{
		srv:    srv,
		notify: make(chan error, 1),
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

func (s *Server) Start() {
	go func() {
		logger.Info("server.Server.Start: starting server at %s\n", s.srv.Addr)
		err := s.srv.ListenAndServe()
		if err != nil {
			err = fmt.Errorf("server.Server.Start: %w", err)
		}
		s.notify <- err
		close(s.notify)
	}()
}

func (s *Server) Notify() <-chan error {
	return s.notify
}

func (s *Server) Shutdown() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), s.shutdownTimeout)

	defer cancelFunc()

	err := s.srv.Shutdown(ctx)
	if err != nil {
		err = fmt.Errorf("server.Server.Shutdown: %w", err)
		return err
	}

	return nil
}
