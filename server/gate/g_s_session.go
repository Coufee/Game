package main

import (
	"sync"

	_ "github.com/golang/protobuf/proto"
	"test/tool/zebra"
)

const s_broadcast_info_size = 65536
const max_client_num uint32 = 6000

type SSession struct {
	mtx             sync.Mutex
	id              uint32
	broker          *zebra.Broker
	sid             uint64
	RemoteAddr      string
	ClientMap       map[uint32]*ClientSession
	Clients         [max_client_num]uint32
	hash            uint32
	broadcast_index uint32
	broadcast_info  [s_broadcast_info_size]byte
}

func (this *SSession) Write(ph *zebra.PackHead, msg interface{}) {
	this.broker.Wirte(ph, msg)
}

func (this *SSession) WirteToClient(ph *zebra.PackHead, msg interface{}) {
	cs := this.ClientMap[ph.Sid]
	if cs != nil {
		cs.broker.Wirte(ph, msg)
	}
}

func (this *SSession) Run() {
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

func (this *SSession) Process(ph *zebra.PackHead, date []byte) {
	if ph.Cmd != uint32(111) {
		return
	}

	switch ph.Cmd {
	case 1:
		// this
	default:
		data := 111
		this.WirteToClient(ph, data)
	}
}

func (this *SSession) Init(broker *zebra.Broker) {
	this.broker = broker
	this.broker.AddWaitGroup()
	this.RemoteAddr = this.broker.RemoteAddr()
	this.hash = g_ctp.NextIndex()
	go this.Run()
}

func (this *SSession) Close() {
}

func (this *SSession) Finish() {
	for more := true; more; {
		select {
		case msg := <-this.broker.ReadMsgQuene:
			this.Process(zebra.GetInputMsgPackHead(msg), msg[12:])
		default:
			more = false
		}
	}
}

func (this *SSession) AddClientSession(cs *ClientSession) {
	this.mtx.Lock()
	this.ClientMap[cs.id] = cs
	this.mtx.Unlock()
}

func (this *SSession) DeleteClientSession(id uint32) {
	this.mtx.Lock()
	delete(this.ClientMap, id)
	this.mtx.Unlock()
}

func (this *SSession) RegistServer(ph *zebra.PackHead, buf []byte) {
	// recv := &1231
	// err := proto.Unmarshal(buf, recv)
	// // if err != nil && recv.getsid() == 0 {
	// // 	return
	// // }

	// if !g_ss.AddSidMap(this.id, this.sid) {
	// 	return
	// }

	// head := &zebra.PackHead
	// this.Write(head, data)
}

func (this *SSession) HeartBeat() {
	head := &zebra.PackHead{}
	head.Cmd = 111
	data := 111
	this.Write(head, data)
}

func (this *SSession) Login() {
}

func (this *SSession) Create() {
}

func (this *SSession) Broadcast() {
	//recv := 1231
	// err := proto.Unmarshal(buf, recv)
	// if err != nil {
	// 	return
	// }
	// rs := len(recv)

	// if rs == 0 {

	// } else {
	// 	for i := 0; i < rs; i++ {
	// 		// id := recv
	// 		// head := &zebra.PackHead
	// 		// this.Write(head, data)
	// 	}
	// }

}

func (this *SSession) AppendBoardcastInfo(msg []byte, size uint32) {
	if size > s_broadcast_info_size {
		return
	}
	if size+this.broadcast_index == s_broadcast_info_size {
		this.ForceBoardcast()
	}
	copy(this.broadcast_info[this.broadcast_index:], msg[0:size])
	this.broadcast_index = size
}

func (this *SSession) ForceBoardcast() {
	if this.broadcast_index > 0 {
		online_size := this.GetOnlineClinets()
		for i := uint32(0); i < online_size; i++ {
			cs := this.GetClient(i)
			if cs != nil && cs.State() == S_NORMAL {
				cs.broker.MyWrite(this.broadcast_info[:this.broadcast_index])
			}
		}
		this.broadcast_index = 0
	}
}

func (this *SSession) KickOut(ph *zebra.PackHead) {
	cs := this.ClientMap[ph.Sid]
	if cs != nil {
		cs.Close()
	}
}

func (this *SSession) GetOnlineClinets() uint32 {
	var index uint32
	this.mtx.Lock()
	for _, v := range this.ClientMap {
		this.Clients[index] = v.id
		index++
		if index >= max_client_num {
			break
		}
	}
	this.mtx.Unlock()
	return index
}

func (this *SSession) GetClientSessionSize() int {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	return len(this.ClientMap)
}

func (this *SSession) GetClient(id uint32) *ClientSession {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	iter, exist := this.ClientMap[id]
	if exist {
		return iter
	}
	return nil
}
