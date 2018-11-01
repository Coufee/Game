package main

import (
	"sync"
	"sync/atomic"
	l4g "test/tool/log4go"
	"test/tool/zebra"
)

const (
	S_NULL         int32 = 0
	S_LOGIN        int32 = 1
	S_WAIT_CRATE   int32 = 2
	S_CREATE       int32 = 3
	S_NORMAL       int32 = 4
	S_LOGOUT       int32 = 5
	S_LOGIN_REPECT int32 = 6
)

type ClientSession struct {
	broker        *zebra.Broker
	id            uint32
	sid           uint32
	uid           uint32
	state         int32
	RemoteAddr    string
	mtx           sync.Mutex
	ss            *SSession
	client_server *ClientServer
	closing       bool
}

func (this *ClientSession) Write(ph *zebra.PackHead, msg interface{}) {
	this.broker.Wirte(ph, msg)
}

func (this *ClientSession) Init(broker *zebra.Broker) {
	this.broker = broker
	this.broker.AddWaitGroup()
	this.RemoteAddr = this.broker.RemoteAddr()
	go this.Run()
}

func (this *ClientSession) Run() {
	defer this.broker.DecWaitGroup()
	defer this.Finish()
	for {
		select {
		case msg := <-this.broker.ReadMsgQuene:
			this.Process(zebra.GetInputMsgPackHead(msg), msg[12:])
		case <-this.broker.CloseChan:
			return
		}
	}
}

func (this *ClientSession) Finish() {
	for more := true; more; {
		select {
		case msg := <-this.broker.ReadMsgQuene:
			this.Process(zebra.GetInputMsgPackHead(msg), msg[12:])
		default:
			more = false
		}
	}
}

func (this *ClientSession) Process(ph *zebra.PackHead, data []byte) {
	l4g.Debug("[Command] msg head: %v", ph)
	if this.sid == 0 || ph.Cmd == uint32(1111) {
	} else {
		l4g.Error("first msg not register: %s", this.broker.RemoteAddr())
	}

	if ph.Cmd != uint32(2222) {
		if ph.Cmd == uint32(333) {
			this.KeepAlive()
			return
		}

		if ph.Sid != this.id {
			l4g.Error("client msg head error, id: %d head_sid: %d cmd: %d", this.sid, ph.Sid, ph.Cmd)
			return
		}
	}

	state := this.State()
	switch state {
	case S_NULL:
		if ph.Cmd == uint32(111) {
			// this.Login(ph, data)
			return
		}
	case S_LOGIN:
		fallthrough
	case S_CREATE:
	case S_LOGOUT:
	case S_NORMAL:
		if ph.Cmd != 111 {

		}
	default:
		l4g.Error("client error state, id: %d, state: %d", this.id, this.State())
	}
}

func (this *ClientSession) Close() {
	this.mtx.Lock()
	// if 1 {
	// }
	this.mtx.Unlock()
	this.client_server.Delete(this.id)
	if this.State() != S_NULL && this.State() != S_LOGIN_REPECT {
		head := &zebra.PackHead{}
		head.Uid = this.uid
		head.Sid = this.sid
		head.Cmd = uint32(1111)
		data := 111
		this.Write(head, data)
	}
}

func (this *ClientSession) KeepAlive() {
	head := &zebra.PackHead{}
	head.Uid = this.uid
	head.Sid = this.sid
	head.Cmd = uint32(11111)
	data := 111
	this.Write(head, data)
}

func (this *ClientSession) WriteToScene(ph *zebra.PackHead, msg interface{}) {
	this.ss.broker.Wirte(ph, msg)
}

func (this *ClientSession) Login() {
}

func (this *ClientSession) State() int32 {
	return atomic.LoadInt32(&this.state)
}
