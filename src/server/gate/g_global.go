package main

import (
	"os"
	"sync/atomic"

	"test/tool/common"
)

var g_ss *SServer = &SServer{
	session:    make(map[uint32]*SSession),
	sceneidmap: make(map[uint64]uint32),
}

var g_cs *ClientServer = &ClientServer{
	sessions: make(map[uint32]*ClientSession),
}

var g_ctp *ThreadPool
var g_timer *common.TimerManager = common.NewTimerManager(1024)
var g_signal = make(chan os.Signal, 1)
var gid uint32

func GetGid() uint32 {
	return atomic.AddUint32(&gid, 1)
}
