// -*- coding:utf-8-unix -*-
package main

import (
	"flag"
	"github.com/golang/glog"
	"tunnel"
)

func init() {
	flag.Parse()
}

func main() {
	tunnel := tunnel.NewClientTunnel()
	err := tunnel.Serve()
	if err != nil {
		glog.Error(err)
	}
}
