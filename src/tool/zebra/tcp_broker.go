package zebra

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	l4g "test/tool/log4go"
)

type Sessioner interface {
	Init(*Broker)              //初始化
	Process(*PackHead, []byte) //处理
	Close()                    //消除
}

const (
	StateInit = iota
	StateDisconnected
	StateConnected
)

type Broker struct {
	cn      *conn
	session Sessioner
	conf    *Config

	ReadMsgQuene  chan []byte
	writeMsgQueue chan *Message

	state     int32
	CloseChan chan struct{}

	wg sync.WaitGroup
}

type Message struct {
	PH   *PackHead
	info interface{}
}

func newBroker(se Sessioner, cf *Config) *Broker {
	return &Broker{
		session: se,
		conf:    cf,
	}
}

func (this *Broker) LocalAddr() string { return this.cn.localAddr }

func (this *Broker) RemoteAddr() string { return this.cn.remoteAddr }

func (this *Broker) State() int32 {
	return atomic.LoadInt32(&this.state)
}

func (this *Broker) Connect(timeout time.Duration) bool {
	rw, err := net.DialTimeout("tcp", this.conf.Address, timeout)
	if err != nil {
		l4g.Error("[Broker] Connet Error: %v", err)
		return false
	}
	if !this.serve(rw) {
		rw.Close()
		return false
	}
	return true
}

func (this *Broker) serve(rwc net.Conn) bool {
	if !atomic.CompareAndSwapInt32(&this.state, StateInit, StateConnected) {
		return false
	}

	this.cn = newconn(rwc, this)
	this.CloseChan = make(chan struct{})
	this.writeMsgQueue = make(chan *Message, this.conf.WriteMsgQueneSize)
	if this.conf.ReadMsgQueneSize > 0 {
		this.ReadMsgQuene = make(chan []byte, this.conf.ReadMsgQueneSize)
	}

	this.wg.Add(1)
	go this.cn.writeLoop()
	this.wg.Add(1)
	go this.cn.readLoop()

	this.session.Init(this)
	return true
}

func (this *Broker) AddWaitGroup() {
	this.wg.Add(1)
}

func (this *Broker) DecWaitGroup() {
	this.wg.Done()
}

func (this *Broker) transmitOrProcessMsg(buf []byte) {
	if this.conf.ReadMsgQueneSize > 0 {
		select {
		case this.ReadMsgQuene <- buf:
		case <-this.CloseChan:
		}
	} else {
		this.session.Process(GetInputMsgPackHead(buf), buf[12:])
	}
}

func GetInputMsgPackHead(buf []byte) *PackHead {
	return &PackHead{
		Length: uint32(len(buf) + 4),
		Cmd:    DecodeUint32(buf[0:]),
		Uid:    DecodeUint32(buf[4:]),
		Sid:    DecodeUint32(buf[8:]),
	}
}

func (this *Broker) Stop() {
	if !atomic.CompareAndSwapInt32(&this.state, StateConnected, StateDisconnected) {
		return
	}

	close(this.CloseChan)
	this.cn.rwc.Close()
	go func() {
		this.wg.Wait()
		this.cn.close()
		this.session.Close()
		l4g.Info("[Broker] close, addr: (%s %s)", this.cn.localAddr, this.cn.remoteAddr)
		//this.cn = nil
		this.session = nil
		atomic.StoreInt32(&this.state, StateInit)
	}()
}

func (this *Broker) Wirte(ph *PackHead, msg interface{}) bool {
	select {
	case this.writeMsgQueue <- &Message{ph, msg}:
		return true
	case <-this.CloseChan:
		l4g.Error("session close: %v %v", ph, msg)
		return false
	}

	// if data, err := this.marshal(ph, msg); err == nil {
	// 	mq_len := len(this.writeMsgQueue)
	// 	mq_cap := cap(this.writeMsgQueue)
	// 	if mq_len > int(HIGH_WATER_MARK_SCALE*float64(mq_cap)) {
	// 		l4g.Warn("[Broker] writeMsgQuene is HighWaterMark, len:%d cap: %d addr: (%s %s)",
	// 			mq_len, mq_cap, this.cn.localAddr, this.cn.remoteAddr)
	// 	}
	// 	select{
	// 	case this.writeMsgQueue <- data:
	// 	case <- this.writeMsgQueue:
	// 	}
	// }
}

func (this *Broker) Marshal(ph *PackHead, msg interface{}) ([]byte, error) {
	var data []byte
	switch v := msg.(type) {
	case []byte:
		data := make([]byte, len(v)+PACK_HEAD_LEN)
		copy(data[PACK_HEAD_LEN:], v)
	case proto.Message:

		data = make([]byte, PACK_HEAD_LEN, 64)
		if mdata, err := proto.Marshal(v); err == nil {
			data = mdata
		} else {
			l4g.Error("[Broker] proto marshal cmd: %d sid: %d uid: %d error: %v",
				ph.Cmd, ph.Sid, ph.Uid, err)
			return nil, err
		}
	default:
		l4g.Error("[Broker] error msg type cmd: %d sid: %d uid: %d",
			ph.Cmd, ph.Sid, ph.Uid)
		return nil, ErrorMsgType
	}

	length := len(data)
	if length > this.conf.MaxWriteMsgSize {
		l4g.Error("[Broker] error write msg size overflow cmd: %d sid: %d uid: %d length: %v",
			ph.Cmd, ph.Sid, ph.Uid, length)
		return nil, WriteOverFlow
	}

	ph.Length = uint32(length)
	l4g.Debug("[Broker] write head %v", ph)
	EncodePackHead(data, ph)
	return data, nil
}

func (this *Broker) MyWrite(buf []byte) {
	this.cn.Wirte(buf)
}
