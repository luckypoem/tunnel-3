// -*- coding:utf-8-unix -*-
package tunnel

import (
	"bufio"
	"context"
	"errors"
	"github.com/golang/glog"
	"io"
	"net"
	"server"
	"socks5"
)

const (
	maxSOCKS5BufLen = 2048
)

type ServerTunnel struct {
	S5     *socks5.SOCKS5
	Reader *bufio.Reader
	Writer *bufio.Writer
}

func NewServerTunnel() Tunnel {
	return &ServerTunnel{
		S5:     socks5.New(),
		Writer: nil,
		Reader: nil,
	}
}

func (self *ServerTunnel) init(conn net.Conn) {
	self.Reader = bufio.NewReader(conn)
	self.Writer = bufio.NewWriter(conn)
}

func (self *ServerTunnel) writeAll(buf []byte) error {
	_, err := self.Writer.Write(buf)
	if err != nil {
		return err
	}
	return self.Writer.Flush()
}

func (self *ServerTunnel) readSome() ([]byte, error) {
	var buf [maxSOCKS5BufLen]byte
	n, err := self.Reader.Read(buf[:])
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (self *ServerTunnel) shakehand() error {
	buf, err := self.readSome()
	if err != nil {
		return err
	}
	_, err = self.S5.ParseHandshake(buf)
	if err != nil {
		return err
	}
	shakehand := self.S5.NewShakehand()
	return self.writeAll(shakehand)
}

func (self *ServerTunnel) request() error {
	buf, err := self.readSome()
	if err != nil {
		return err
	}
	_, err = self.S5.ParseRequest(buf)
	if err != nil {
		return err
	}
	reply := self.S5.NewReply()
	err = self.writeAll(reply)
	if err != nil {
		return err
	}
	return nil
}

func (self *ServerTunnel) dial() (net.Conn, error) {
	addr := self.S5.Address()
	cmd := self.S5.Command()
	if cmd != socks5.S5Connect {
		return nil, errors.New("socks5: command not support")
	} else {
		return net.Dial("tcp", addr)
	}
}

func (self *ServerTunnel) Handle(_ context.Context, conn net.Conn) {
	defer conn.Close()

	self.init(conn)

	err := self.shakehand()
	if err != nil {
		if err != io.EOF {
			glog.Warning(err)
		}
		return
	}
	err = self.request()
	if err != nil && err != io.EOF {
		if err != io.EOF {
			glog.Warning(err)
		}
		return
	}
	remote, err := self.dial()
	if err != nil {
		glog.Warning(err)
		return
	}
	defer remote.Close()

	go io.Copy(conn, remote)
	io.Copy(remote, conn)
}

func (self *ServerTunnel) Serve() error {
	config, err := NewTLSServerConfig()
	if err != nil {
		return err
	}
	s := server.NewTLSServer(ServerAddr, config)
	_, err = s.Serve(self)
	if err != nil {
		return err
	}
	err = <-s.Err()
	return err
}
