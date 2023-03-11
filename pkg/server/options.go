package server

import "time"

type Option func(server *Server)

func WithWriteTimeout(dur time.Duration) Option {
	return func(server *Server) {
		server.srv.WriteTimeout = dur
	}
}

func WithReadTimeout(dur time.Duration) Option {
	return func(server *Server) {
		server.srv.ReadTimeout = dur
	}
}

func WithShutdownTimeout(dur time.Duration) Option {
	return func(server *Server) {
		server.shutdownTimeout = dur
	}
}
