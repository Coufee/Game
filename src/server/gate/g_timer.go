package main

import (
	"runtime"
	l4g "test/tool/log4go"
	"time"
)

type TenSecondTimer struct{}

func (this *TenSecondTimer) TimeOuter(now int64) {
	g_ss.Timer()
	g_cs.Timer()
}

type TenMinuteTimer struct{}

func (this *TenMinuteTimer) TimeOuter(now int64) {
	times := time.Now()
	runtime.GC()
	l4g.Info("GC time %v", time.Now().Sub(times))
}
