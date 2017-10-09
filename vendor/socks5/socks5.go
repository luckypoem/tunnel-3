// -*-coding:utf-8-unix;-*-
package socks5

import (
	"encoding/binary"
	"errors"
	"net"
	"strconv"
)

const (
	s5MaxAuthBytes = 1 + 1 + 255
	s5MinAuthBytes = 1 + 1 + 1
	s5Version      = 0x05
	s5AuthNone     = 0x00
	s5AuthUnaccept = 0xff
)

const (
	ver      = 0
	nmethods = 1
	cmd      = 1
	srv      = 2
	atyp     = 3
	rep      = 1
	addr     = 4
)

const (
	s5RSV             = 0x00
	s5MinRequestBytes = 1 + 1 + 1 + 1 + 2
)

const (
	S5Connect = 0x01
	S5Bind    = 0x2
	S5UDP     = 0x3
)

const (
	s5IPv4    = 0x01
	s5IPv4Len = 4
	s5Domain  = 0x03
	s5IPv6    = 0x04
	s5IPv6Len = 16
)

const (
	s5Succeeded = iota
	s5GeneralFailure
	s5ConnectionNotAllowed
	s5NetworkUnreachable
	s5HostUnreachable
	s5ConnectionRefused
	s5TTLExpired
	s5CommandNotSupported
	s5AddressNotSupported
	s5Undefined = 0xff
)

var s5Errors = []string{
	"",
	"socks5: general failure",
	"socks5: connection forbidden",
	"socks5: network unreachable",
	"socks5: host unreachable",
	"socks5: connection refused",
	"socks5: TTL expired",
	"socks5: command not supported",
	"socks5: address type not supported",
}

type SOCKS5 struct {
	Errno byte
	Host  string
	Port  uint16
	CMD   uint8
}

func New() *SOCKS5 {
	return &SOCKS5{
		Errno: s5Undefined,
		Host:  "",
		Port:  0,
		CMD:   0,
	}
}

func (self *SOCKS5) Command() uint8 {
	return self.CMD
}

func (self *SOCKS5) Address() string {
	if self.Errno != s5Succeeded {
		return ""
	}
	port := strconv.FormatUint(uint64(self.Port), 10)
	return net.JoinHostPort(self.Host, port)
}

func (self *SOCKS5) ParseHandshake(buf []byte) (n int, err error) {
	self.Errno = s5Succeeded

	// more buffer
	if len(buf) < s5MinAuthBytes {
		self.Errno = s5GeneralFailure
		return 0, errors.New(s5Errors[self.Errno])
	}

	// invalid version
	if buf[ver] != s5Version {
		self.Errno = s5GeneralFailure
		return 0, errors.New(s5Errors[self.Errno])
	}

	// more buffer
	n = int(buf[nmethods]) + 2
	if len(buf) < n {
		self.Errno = s5GeneralFailure
		return 0, errors.New(s5Errors[self.Errno])
	}

	return n, nil
}

func (self *SOCKS5) NewShakehand() []byte {
	if self.Errno != s5Succeeded {
		return []byte{s5Version, s5AuthUnaccept}
	} else {
		return []byte{s5Version, s5AuthNone}
	}
}

func (self *SOCKS5) ParseRequest(buf []byte) (n int, err error) {
	self.Errno = s5Succeeded

	// more buffer
	if len(buf) <= s5MinRequestBytes {
		self.Errno = s5GeneralFailure
		return 0, errors.New(s5Errors[self.Errno])
	}

	// invalid version
	if buf[ver] != s5Version {
		self.Errno = s5GeneralFailure
		return 0, errors.New(s5Errors[self.Errno])
	}

	// is supported command
	switch buf[cmd] {
	case S5Connect, S5UDP:
	default:
		self.Errno = s5CommandNotSupported
		return 0, errors.New(s5Errors[s5CommandNotSupported])
	}
	self.CMD = buf[cmd]

	// invalid SRV
	if buf[srv] != s5RSV {
		self.Errno = s5CommandNotSupported
		return 0, errors.New(s5Errors[s5CommandNotSupported])
	}

	// address type check
	offset := 0
	addrlen := 0
	switch buf[atyp] {
	case s5IPv4:
		offset = addr
		addrlen = s5IPv4Len
	case s5Domain:
		offset = addr + 1
		addrlen = int(buf[addr])
	case s5IPv6:
		offset = addr
		addrlen = s5IPv6Len
	default:
		self.Errno = s5AddressNotSupported
		return 0, errors.New(s5Errors[s5AddressNotSupported])
	}

	if addrlen == 0 {
		self.Errno = s5AddressNotSupported
		return 0, errors.New(s5Errors[s5AddressNotSupported])
	}

	// more buffer
	if len(buf) < addrlen+s5MinRequestBytes {
		self.Errno = s5GeneralFailure
		return 0, errors.New(s5Errors[self.Errno])
	}

	var host string
	switch buf[atyp] {
	case s5IPv4:
		ip := net.IPv4(buf[offset], buf[offset+1],
			buf[offset+2], buf[offset+3])
		host = ip.String()
	case s5Domain:
		host = string(buf[offset : offset+addrlen])
	case s5IPv6:
		self.Errno = s5AddressNotSupported
		return 0, errors.New(s5Errors[s5AddressNotSupported])
	}
	self.Host = host
	port := binary.BigEndian.Uint16(buf[offset+addrlen:])
	self.Port = port

	n = offset + addrlen + 2

	return n, nil
}

func (self *SOCKS5) NewErrorReply(errno byte) []byte {
	var reply = []byte{s5Version, s5Undefined, s5RSV, s5IPv4,
		0, 0, 0, 0, 0, 0}
	reply[rep] = errno
	return reply
}

func (self *SOCKS5) NewReply() []byte {
	return self.NewErrorReply(self.Errno)
}
