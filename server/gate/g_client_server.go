package main

import (
	"sync"
	l4g "test/tool/log4go"
	"test/tool/zebra"
)

type ClientServer struct {
	mutex    sync.Mutex
	sessions map[uint32]*ClientSession
}

func (this *ClientServer) NewSession() zebra.Sessioner {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	session := &ClientSession{}
	session.id = GetGid()
	session.client_server = this
	this.Add(session)
	l4g.Info("new clent session, id: %d addr: %s",
		session.id, session.RemoteAddr)
	return session
}

func (this *ClientServer) Close() {
	return
}

func (this *ClientServer) Delete(id uint32) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	delete(this.sessions, id)
}

func (this *ClientServer) Add(cs *ClientSession) {
	this.sessions[cs.id] = cs
}

func (this *ClientServer) Timer() {
	l4g.Info("Wait login user count: %d", this.GetSessionSize())
}

func (this *ClientServer) GetSessionSize() int {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return len(this.sessions)
}
