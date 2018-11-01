package zebra

import (
	l4g "test/tool/log4go"
)

type Command interface {
	Excute(Sessioner, *PackHead, []byte) bool
}

type CommandM struct {
	cmdm map[uint32]Command
}

func NewCommandM() *CommandM {
	return &CommandM{
		cmdm: make(map[uint32]Command),
	}
}

func (this *CommandM) Register(id uint32, cmd Command) {
	this.cmdm[id] = cmd
}

func (this *CommandM) Dispatcher(session Sessioner, ph *PackHead, data []byte) bool {
	if cmd, exist := this.cmdm[ph.Cmd]; exist {
		return cmd.Excute(session, ph, data)
	}

	l4g.Error("[Command] no find cmd: %d %d", ph.Sid, ph.Cmd)
	return false
}
