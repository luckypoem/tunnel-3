// -*-coding:utf-8-unix;-*-
package server

import (
	"context"
	"flag"
	"github.com/golang/glog"
	"io"
	"net"
	"testing"
	"time"
)

func init() {
	flag.Parse()
}

type Echo struct {
}

func (self Echo) Handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := io.CopyN(conn, conn, 1024)
			if err != io.EOF {
				glog.Info(conn.RemoteAddr(), err)
				return
			}
		}
	}
}

func TestServer1(t *testing.T) {
	var h Echo
	s := NewTCPServer(":1023")
	s.Serve(h)
	err := <-s.Err()
	glog.Info(err)
}

func TestServer2(t *testing.T) {
	var h Echo
	s := NewTCPServer(":9000")
	ln, err := s.Serve(h)
	if err == nil {
		time.Sleep(6 * time.Second)
		err := ln.Close()
		if err != nil {
			glog.Warning(err)
		}
		err = <-s.Err()
		glog.Error(err)
	}
}
