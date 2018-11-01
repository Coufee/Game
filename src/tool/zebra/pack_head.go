package zebra

import (
	l4g "test/tool/log4go"
)

type PackHead struct {
	Length uint32
	Cmd    uint32
	Uid    uint32
	Sid    uint32
}

func DecodePackHead(buf []byte, ph *PackHead) bool {
	if len(buf) < 16 {
		l4g.Error("[PackHead] decode size no enough size:%v", len(buf))
		return false
	}

	ph.Length = DecodeUint32(buf[0:])
	ph.Cmd = DecodeUint32(buf[4:])
	ph.Uid = DecodeUint32(buf[8:])
	ph.Sid = DecodeUint32(buf[12:])
	return true
}

func EncodePackHead(buf []byte, ph *PackHead) bool {
	if len(buf) < 16 {
		l4g.Error("[PackHead] encode size no enough size:%v", len(buf))
		return false
	}

	EncodeUint32(ph.Length, buf[0:])
	EncodeUint32(ph.Cmd, buf[4:])
	EncodeUint32(ph.Uid, buf[8:])
	EncodeUint32(ph.Sid, buf[12:])
	return true
}
