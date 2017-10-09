// -*-coding:utf-8-unix;-*-
package server

import (
	"context"
	"net"
)

type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
}

type Server interface {
	Serve(h Handler) (net.Listener, error)
	Err() chan error
}
