package zebra

import (
	l4g "tool/log4go"
)

type Config struct {
	Address string //ip：port

	//read
	MaxReadMsgSize   int
	ReadMsgQueneSize int
	ReadTimeOut      int

	//write
	MaxWriteMsgSize   int
	WriteMsgQueneSize int
	WriteTimeOut      int
}

func (this *Config) Check() bool {
	if this.MaxWriteMsgSize == 0 {
		l4g.Error("[Config] MaxWriteMsgSize error")
		return false
	}
	if this.WriteMsgQueneSize == 0 {
		l4g.Error("[Config] WriteMsgQueneSize erro r")
		return false
	}
	if this.MaxReadMsgSize == 0 {
		l4g.Error("[Config] MaxReadMsgSize error")
		return false
	}
	return true
}
