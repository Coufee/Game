package main

import (
	"runtime"
	"time"
	l4g "tool/log4go"
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
