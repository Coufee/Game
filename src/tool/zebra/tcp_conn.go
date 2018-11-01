package zebra

import (
	"bufio"
	"io"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	l4g "test/tool/log4go"
)

const msgSizeBytes uint32 = 4

type conn struct {
	//流量统计
	readMsgCount    uint32
	writeMsgCount   uint32
	readMsgSize     uint64
	writeMsgSize    uint64
	readMsgMaxSize  uint32
	writeMsgMaxSize uint32

	rwc        net.Conn
	remoteAddr string
	localAddr  string

	msgLength []byte

	broker       *Broker
	writeTimeOut time.Duration
	readTimeOut  time.Duration
}

func newconn(cn net.Conn, bk *Broker) *conn {
	c := &conn{
		rwc:        cn,
		remoteAddr: cn.RemoteAddr().String(),
		localAddr:  cn.LocalAddr().String(),
		msgLength:  make([]byte, msgSizeBytes, msgSizeBytes),
		broker:     bk,
	}
	if c.broker.conf.WriteTimeOut > 0 {
		c.writeTimeOut = time.Duration(c.broker.conf.WriteTimeOut) * time.Second
	} else {
		c.writeTimeOut = WRITE_TIME_OUT
	}

	if c.broker.conf.ReadTimeOut > 0 {
		c.readTimeOut = time.Duration(c.broker.conf.ReadTimeOut) * time.Second
	} else {
		c.readTimeOut = READ_TIME_OUT
	}
	return c
}

func (this *conn) read(r io.Reader) ([]byte, error) {
	_, err := io.ReadFull(r, this.msgLength)
	if err != nil {
		l4g.Error("[conn] io read length error:%s &s %v",
			this.localAddr, this.remoteAddr, err)
		return nil, err
	}

	msgSize := DecodeUint32(this.msgLength)
	if msgSize < PACK_HEAD_LEN ||
		msgSize > uint32(this.broker.conf.ReadMsgQueneSize) {
		l4g.Error("[conn] pack length error: %s %s len: %d",
			this.localAddr, this.remoteAddr, msgSize)
		return nil, ReadOverFlow
	}

	buf := make([]byte, msgSize-msgSizeBytes, msgSize-msgSizeBytes)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		l4g.Error("[conn] io data error:%s %s %v",
			this.localAddr, this.remoteAddr, err)
		return nil, err
	}

	this.readMsgCount++
	this.readMsgSize += uint64(msgSize)
	if msgSize > this.readMsgMaxSize {
		this.readMsgMaxSize = msgSize
	}
	if this.readMsgCount&0xff == 0 {
		l4g.Error("[conn] socket(%s-&s) read info:%d %d %d",
			this.localAddr, this.remoteAddr, this.readMsgCount, this.readMsgMaxSize, this.readMsgSize)
	}

	return buf, nil
}

func (this *conn) readLoop() {
	rbuf := bufio.NewReader(this.rwc)
	for {
		this.rwc.SetReadDeadline(time.Now().Add(this.readTimeOut))
		buf, err := this.read(rbuf)
		if err != nil {
			l4g.Error("[conn] read error: %s %s %s",
				this.localAddr, this.remoteAddr, err)
			this.broker.Stop()
			goto exit
		}
		this.broker.transmitOrProcessMsg(buf)
	}
exit:
	this.broker.wg.Done()
}

func (this *conn) Wirte(msg []byte) bool {
	n, err := this.rwc.Write(msg)
	msg = nil
	if err != nil {
		l4g.Error("[conn] write error: %s %s %v",
			this.localAddr, this.remoteAddr, err)
		this.broker.Stop()
		return false
	}
	this.writeMsgCount++
	this.writeMsgSize += uint64(n)
	if uint32(n) > this.writeMsgMaxSize {
		this.writeMsgMaxSize = uint32(n)
	}
	if this.writeMsgSize&0xff == 0 {
		l4g.Error("[conn] socket(%s-&s) write info:%d %d %d",
			this.localAddr, this.remoteAddr, this.readMsgCount, this.writeMsgMaxSize, this.writeMsgSize)
	}

	return true
}

func (this *conn) writeLoop() {
	max_size := this.broker.conf.MaxWriteMsgSize
	write_buff := make([]byte, max_size)
	head_buff := make([]byte, PACK_HEAD_LEN)
	data_buff := make([]byte, max_size-PACK_HEAD_LEN)
	for {
		select {
		case msg := <-this.broker.writeMsgQueue:
			length, data, err := Marshal(msg.PH, msg.info, max_size, head_buff, data_buff)
			if err == nil {
				index := 0
				copy(write_buff, head_buff)
				copy(write_buff[PACK_HEAD_LEN:], data)
				index += length
				for more := true; more; {
					select {
					case msg := <-this.broker.writeMsgQueue:
						length, data, err := Marshal(msg.PH, msg.info, max_size, head_buff, data_buff)
						if err == nil {
							if index+length <= max_size {
								copy(write_buff[index:], head_buff)
								copy(write_buff[index+PACK_HEAD_LEN:], data)
								index += length
							} else {
								if !this.Wirte(write_buff[:index]) {
									goto exit
								}
								index = 0
								copy(write_buff[index:], head_buff)
								copy(write_buff[index+PACK_HEAD_LEN:], data)
								index += length
							}
						}
					case <-this.broker.CloseChan:
						goto exit
					default:
						more = false
					}
				}
				if !this.Wirte(write_buff[:index]) {
					goto exit
				}
			}
		case <-this.broker.CloseChan:
			goto exit
		}
	}
exit:
	this.broker.wg.Done()
}

func (this *conn) close() {
	this.broker = nil
}

func Marshal(ph *PackHead, msg interface{}, max_size int, head_buff, data_buff []byte) (int, []byte, error) {
	var data []byte
	switch v := msg.(type) {
	case []byte:
		data = v
	case proto.Message:
		if mdata, err := proto.Marshal(v); err == nil {
			data = mdata
		} else {
			l4g.Error("[Broker] proto marshal cmd: %d sid: %d uid: %d error: %v",
				ph.Cmd, ph.Sid, ph.Uid, err)
			return 0, nil, err
		}
	default:
		l4g.Error("[Broker] error msg type cmd: %d sid: %d uid: %d",
			ph.Cmd, ph.Sid, ph.Uid)
		return 0, nil, ErrorMsgType
	}

	length := len(data) + PACK_HEAD_LEN
	l4g.Fine("CmdId:%d MsgLength:%d UserId:%d", ph.Cmd, length, ph.Uid)
	if length > max_size {
		l4g.Error("[Broker] write msg size overflow cmd: %d sid: %d uid: %d",
			ph.Cmd, ph.Sid, ph.Uid)
		return 0, nil, WriteOverFlow
	}

	ph.Length = uint32(length)
	l4g.Debug("[Broker] write head %v", ph)
	EncodePackHead(head_buff, ph)
	return length, data, nil
}
