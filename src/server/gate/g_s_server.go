package main

import (
	"sync"
	"test/tool/zebra"
)

type SServer struct {
	mutex      sync.Mutex
	session    map[uint32]*SSession
	sceneidmap map[uint64]uint32
}

func (tihs *SServer) Close() {}
func (this *SServer) NewSession() zebra.Sessioner {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	session := &SSession{
		ClientMap: make(map[uint32]*ClientSession),
	}
	session.id = GetGid()
	this.Add(session)
	return session
}

func (this *SServer) Add(ss *SSession) {
	this.session[ss.id] = ss
}

func (this *SServer) Timer() {
	var count uint32
	this.mutex.Lock()
	defer this.mutex.Unlock()
	for _, v := range this.session {
		count += uint32(v.GetClientSessionSize())
	}
}

func (this *SServer) Delete(id uint32, sid uint64) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	i, exist := this.sceneidmap[sid]
	if exist && i == id {
		delete(this.sceneidmap, sid)
	}
	delete(this.session, id)
}

func (this *SServer) GetSessionBySceneId(a uint64) (ret *SSession) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	i, exist := this.sceneidmap[a]
	if exist {
		ret, _ = this.session[i]
	}
	return ret
}

func (this *SServer) GetSessionById(a uint32) (ret *SSession) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	ret, _ = this.session[a]
	return ret
}

func (this *SServer) AddSidMap(id uint32, sid uint64) bool {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	_, exist1 := this.sceneidmap[sid]
	if exist1 {
		return false
	}
	_, exist2 := this.session[id]
	if exist2 {
		return false
	}
	if !exist1 && exist2 {
		this.sceneidmap[sid] = id
	}
	return true
}
