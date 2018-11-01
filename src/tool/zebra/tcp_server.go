package zebra

import "net"

type Server interface {
	NewSession() Sessioner
	Close()
}

func TCPServe(srv Server, conf *Config) {
	defer srv.Close()

	l, e := net.Listen("tcp", conf.Address)
	if e != nil {
		panic(e.Error())
	}

	defer l.Close()

	for {
		rw, e := l.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				continue
			}
			return
		}
		newBroker(srv.NewSession(), conf).serve(rw)
	}
}
