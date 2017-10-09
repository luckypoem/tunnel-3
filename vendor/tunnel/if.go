// -*- coding:utf-8-unix -*-
package tunnel

type Tunnel interface {
	Serve() error
}
