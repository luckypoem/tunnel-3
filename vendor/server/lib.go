// -*-coding:utf-8-unix;-*-
package server

import (
	"context"
	"crypto/tls"
	"github.com/golang/glog"
	"net"
	"time"
)

type TCPServer struct {
	Laddr string
	Error chan error
}

type TLSServer struct {
	TCPServer
	TLSConfig *tls.Config
}

func NewTCPServer(laddr string) Server {
	return &TCPServer{
		Laddr: laddr,
		Error: make(chan error, 1),
	}
}

func NewTLSServer(laddr string, config *tls.Config) Server {
	return &TLSServer{
		TCPServer: TCPServer{
			Laddr: laddr,
			Error: make(chan error, 1),
		},
		TLSConfig: config,
	}
}

func acceptLoop(ch chan error, ln net.Listener, h Handler) {
	defer ln.Close()
	var tempDelay time.Duration = 0 // how long to sleep on accept failure

	rootCtx := context.Background()
	acceptCtx, cancel := context.WithCancel(rootCtx)

	for {
		select {
		case <-acceptCtx.Done():
			return
		default:
		}
		conn, err := ln.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				glog.Warningf("Accept error: %v; retrying in %v",
					nerr, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			glog.Error(err)
			ch <- err
			cancel()
		} else {
			tempDelay = 0
			go h.Handle(acceptCtx, conn)
		}
	}
}

func (self *TCPServer) Serve(h Handler) (net.Listener, error) {
	ln, err := net.Listen("tcp", self.Laddr)
	if err != nil {
		self.Error <- err
		return nil, err
	}

	go acceptLoop(self.Error, ln, h)
	return ln, nil
}

func (self *TLSServer) Serve(h Handler) (net.Listener, error) {
	ln, err := tls.Listen("tcp", self.Laddr, self.TLSConfig)
	if err != nil {
		self.Error <- err
		return nil, err
	}

	go acceptLoop(self.Error, ln, h)
	return ln, nil
}

func (self *TCPServer) Err() chan error {
	return self.Error
}
