package network

import (
	"encoding/binary"
	"errors"
	. "ghost/global"
	"io"
)

// --------------
// | len | data |
// --------------
type MsgParser struct {
	lenMsgLen    uint32
	minMsgLen    uint32
	maxMsgLen    uint32
	littleEndian bool
}

func NewMsgParser() *MsgParser {
	p := new(MsgParser)
	p.lenMsgLen = 2
	p.minMsgLen = 1
	p.maxMsgLen = 4096
	p.littleEndian = false

	return p
}

// It's dangerous to call the method on reading or writing
func (p *MsgParser) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

// It's dangerous to call the method on reading or writing
func (p *MsgParser) SetMsgLen(msglen, minlen, maxlen uint32) {
	p.lenMsgLen = msglen
	p.minMsgLen = minlen
	p.maxMsgLen = maxlen
}

// goroutine safe
func (p *MsgParser) Read(conn *ConnSession) ([]byte, error) {
	var b [4]byte
	bufMsgLen := b[:p.lenMsgLen]

	// read len
	if n, err := io.ReadFull(conn, bufMsgLen); err != nil {
		Log.Warning("read header failed, ip:%v reason:%v size:%v", conn.RemoteAddr(), err, n)
		return nil, err
	}

	// parse len
	var msgLen uint32
	if p.littleEndian {
		if p.lenMsgLen == 2 {
			msgLen = uint32(binary.LittleEndian.Uint16(bufMsgLen))
		} else if p.lenMsgLen == 4 {
			msgLen = binary.LittleEndian.Uint32(bufMsgLen)
		} else {
			Log.Warning("read header failed, not support msglen(%v)", p.lenMsgLen)
		}

	} else {
		if p.lenMsgLen == 2 {
			msgLen = uint32(binary.BigEndian.Uint16(bufMsgLen))
		} else if p.lenMsgLen == 4 {
			msgLen = binary.BigEndian.Uint32(bufMsgLen)
		} else {
			Log.Warning("read header failed, not support msglen(%v)", p.lenMsgLen)
		}
	}

	// check len
	if msgLen > p.maxMsgLen {
		return nil, errors.New("message too long")
	} else if msgLen < p.minMsgLen {
		return nil, errors.New("message too short")
	}
	//Console.Info("len ............. [%v][%v]", p.lenMsgLen, msgLen)
	// data
	msgData := make([]byte, msgLen-p.lenMsgLen)
	if n, err := io.ReadFull(conn, msgData); err != nil {
		Log.Warning("read payload failed, ip:%v reason:%v size:%v", conn.RemoteAddr(), err, n)
		return nil, err
	}
	return msgData, nil
}

// goroutine safe
func (p *MsgParser) Write(conn *ConnSession, args ...[]byte) error {
	// get len
	var msgLen uint32
	for i := 0; i < len(args); i++ {
		msgLen += uint32(len(args[i]))
	}

	// check len
	if msgLen > p.maxMsgLen {
		return errors.New("message too long")
	} else if msgLen < p.minMsgLen {
		return errors.New("message too short")
	}

	msg := make([]byte, p.lenMsgLen+msgLen)

	// write len
	if p.littleEndian {
		binary.LittleEndian.PutUint16(msg, uint16(msgLen+p.lenMsgLen))
	} else {
		binary.BigEndian.PutUint16(msg, uint16(msgLen+p.lenMsgLen))
	}

	// write data
	offset := int(p.lenMsgLen)
	for i := 0; i < len(args); i++ {
		copy(msg[offset:], args[i])
		offset += len(args[i])
	}

	conn.Write(msg)

	return nil
}
