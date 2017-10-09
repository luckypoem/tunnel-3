// -*- coding:utf-8-unix -*-
package tunnel

import (
	"context"
	"crypto/tls"
	"github.com/golang/glog"
	"io"
	"net"
	"server"
)

type ClientTunnel struct {
	TLSConfig *tls.Config
}

func NewClientTunnel() *ClientTunnel {
	return &ClientTunnel{
		TLSConfig: nil,
	}
}

func (self *ClientTunnel) Handle(_ context.Context, conn net.Conn) {
	defer conn.Close()
	remote, err := tls.Dial("tcp", ServerAddr, self.TLSConfig)
	if err != nil {
		glog.Warning(err)
		return
	}
	go io.Copy(conn, remote)
	io.Copy(remote, conn)
}

func (self *ClientTunnel) Serve() error {
	config, err := NewTLSClientConfig("localhost")
	if err != nil {
		return err
	}
	self.TLSConfig = config
	s := server.NewTCPServer(ClientAddr)
	_, err = s.Serve(self)
	if err != nil {
		return err
	}
	err = <-s.Err()
	return err
}
