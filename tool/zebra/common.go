package zebra

import (
	"errors"
	"time"
)

const (
	PACK_HEAD_LEN         = 16
	READ_TIME_OUT         = 60 * time.Second
	WRITE_TIME_OUT        = 10 * time.Second
	HIGH_WATER_MARK_SCALE = 0.9
)

var (
	NoConnect     = errors.New("tcp_conn: no connect")
	ReadOverFlow  = errors.New("tcp_conn: read buffer overflow")
	WriteOverFlow = errors.New("tcp_conn: write buffer overflow")
	ErrorMsgType  = errors.New("tcp_conn: error msg type")
)

func DecodeUint32(data []byte) uint32 {
	return (uint32(data[0]) << 24) | (uint32(data[1]) << 16) | (uint32(data[2]) << 8) | uint32(data[3])
}

func EncodeUint32(n uint32, b []byte) {
	b[3] = byte(n & 0xFF)
	b[2] = byte((n >> 8) & 0xFF)
	b[1] = byte((n >> 16) & 0xFF)
	b[0] = byte((n >> 24) & 0xFF)
}
