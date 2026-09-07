package server

import (
	"context"
	"os"

	"go.kenn.io/kit/daemon"
)

func (s *Server) registerHealthRoutes() {
	group := newRouteGroup(s.api, "/api", "Health")

	get(s, group, "/ping", "Ping daemon", s.humaPing)
}

func (s *Server) humaPing(
	_ context.Context,
	_ *emptyInput,
) (*jsonOutput[daemon.PingInfo], error) {
	return &jsonOutput[daemon.PingInfo]{
		Body: daemon.PingInfo{
			OK:      true,
			Service: daemonService,
			Version: s.version.Version,
			PID:     os.Getpid(),
		},
	}, nil
}
