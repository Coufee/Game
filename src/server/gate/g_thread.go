package main

import (
	"container/list"
	"fmt"
	"sync"
	"test/tool/zebra"
	"time"
)

type Thread struct {
	is_running bool
	mtx        sync.Mutex
	msg_quene  chan []byte
}

func NewThread() *Thread {
	ret := new(Thread)
	ret.is_running = true
	ret.msg_quene = make(chan []byte, 100000)
	return ret
}

func DeleteThread(this *Thread) {
	if this.is_running {
		this.Stop()
	}
}

func (this *Thread) Stop() {
	this.is_running = false
	this.mtx.Lock()
}

func (this *Thread) Append(sid uint32, msg []byte, msg_id uint32) {
	head := &zebra.PackHead{}
	head.Cmd = msg_id
	head.Length = uint32(len(msg) + 16)
	head.Sid = sid
	head.Uid = 0
	info := make([]byte, head.Length)
	copy(info[16:], msg)
	zebra.EncodePackHead(info, head)
	this.msg_quene <- info
}

func (this *Thread) Run() {
	var temp_quene chan []byte = make(chan []byte, 1024)
	var temp_quene_len uint32
	var msg []byte = nil
	var id_vec *list.List = list.New().Init()
	ticker := time.NewTicker(500 * time.Millisecond)
	for this.is_running {
		if this.is_running {
			select {
			case msg = <-this.msg_quene:
				temp_quene_len++
				temp_quene <- msg
			case <-ticker.C:
				break
			}
		} else {
			return
		}
		// fmt.Println("cq1")
		msg = nil
		for temp_quene_len != 0 {
			temp_quene_len--
			fmt.Println(temp_quene_len, "----")
			msg = <-temp_quene
			head := &zebra.PackHead{}
			zebra.DecodePackHead(msg, head)
			fmt.Println(msg, head.Sid)
			ss := g_ss.GetSessionById(head.Sid)
			if ss != nil {
				// ss.
			}
		}
		// fmt.Println("cq2")
		for v := id_vec.Front(); v != nil; v = v.Next() {
			ss := g_ss.GetSessionById(v.Value.(uint32))
			if ss != nil {
				// ss.r
			}
		}
	}
}
