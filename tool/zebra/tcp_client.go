package zebra

import "time"

//启动 tcp client
func TCPClientServer(se Sessioner, conf *Config, timeout time.Duration) bool {
	return newBroker(se, conf).Connect(timeout)
}