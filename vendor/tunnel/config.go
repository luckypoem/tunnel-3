// -*- coding:utf-8-unix -*-
package tunnel

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
)

const (
	ServerAddr = ":9443"
	ClientAddr = ":1080"
)

func newCertPool() (*x509.CertPool, error) {
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM([]byte(CACert)) {
		return nil, errors.New("tunnel: AppendCertsFromPEM failed")
	}
	return cp, nil
}

func NewTLSServerConfig() (*tls.Config, error) {
	cp, err := newCertPool()
	if err != nil {
		return nil, err
	}
	cert, err := tls.X509KeyPair([]byte(ServerCert), []byte(ServerKey))
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    cp,
	}, nil
}

func NewTLSClientConfig(name string) (*tls.Config, error) {
	cp, err := newCertPool()
	if err != nil {
		return nil, err
	}
	cert, err := tls.X509KeyPair([]byte(ClientCert), []byte(ClientKey))
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName:   name,
		RootCAs:      cp,
	}, nil
}
