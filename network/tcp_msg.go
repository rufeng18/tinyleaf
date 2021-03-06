package network

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/xxtea/xxtea-go/xxtea"
)

var ENCRYPT_KEY = "0123456789"

// ------------------
// | len |head| data |
// ------------------
type MsgParser struct {
	lenMsgLen     int
	lenExtHeadLen int // 协议头长度
	minMsgLen     uint32
	maxMsgLen     uint32
	littleEndian  bool
	encrypt       bool // 加密标示
}

func NewMsgParser() *MsgParser {
	p := new(MsgParser)
	p.lenMsgLen = 2
	p.lenExtHeadLen = 2
	p.minMsgLen = 1
	p.maxMsgLen = 4096
	p.littleEndian = false
	p.encrypt = false

	return p
}

// It's dangerous to call the method on reading or writing
func (p *MsgParser) SetMsgLen(lenMsgLen int, minMsgLen uint32, maxMsgLen uint32) {
	if lenMsgLen == 1 || lenMsgLen == 2 || lenMsgLen == 4 {
		p.lenMsgLen = lenMsgLen
	}
	if minMsgLen != 0 {
		p.minMsgLen = minMsgLen
	}
	if maxMsgLen != 0 {
		p.maxMsgLen = maxMsgLen
	}

	var max uint32
	switch p.lenMsgLen {
	case 1:
		max = math.MaxUint8
	case 2:
		max = math.MaxUint16
	case 4:
		max = math.MaxUint32
	}
	if p.minMsgLen > max {
		p.minMsgLen = max
	}
	if p.maxMsgLen > max {
		p.maxMsgLen = max
	}
}

// It's dangerous to call the method on reading or writing
func (p *MsgParser) SetExtHeadLen(l int) {
	p.lenExtHeadLen = l
}

// It's dangerous to call the method on reading or writing
func (p *MsgParser) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

func (p *MsgParser) SetEncrypt(encrypt bool) {
	p.encrypt = encrypt
}

// goroutine safe
func (p *MsgParser) Read(conn *TCPConn) ([]byte, error) {
	var b [4]byte
	bufMsgLen := b[:p.lenMsgLen] // 读取头数据
	// read len
	if _, err := io.ReadFull(conn, bufMsgLen); err != nil {
		return nil, err
	}

	// parse len
	var msgLen uint32
	switch p.lenMsgLen {
	case 1:
		msgLen = uint32(bufMsgLen[0])
	case 2:
		if p.littleEndian {
			msgLen = uint32(binary.LittleEndian.Uint16(bufMsgLen))
		} else {
			msgLen = uint32(binary.BigEndian.Uint16(bufMsgLen))
		}
	case 4:
		if p.littleEndian {
			msgLen = binary.LittleEndian.Uint32(bufMsgLen)
		} else {
			msgLen = binary.BigEndian.Uint32(bufMsgLen)
		}
	}

	// check len
	if msgLen > p.maxMsgLen {
		return nil, errors.New("message too long")
	} else if msgLen < p.minMsgLen {
		return nil, errors.New("message too short")
	}

	fmt.Println("tcp_msg.go.MsgParse.Read rawDate Len:", msgLen)

	// data
	msgData := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgData); err != nil {
		return nil, err
	}

	fmt.Println("tcp_msg.go.MsgParse.Read msgData:", msgData)
	bodyData := msgData[p.lenExtHeadLen:] // 跳过消息头
	fmt.Println("tcp_msg.go.MsgParse.Read bodyData:", string(bodyData))
	// decrypt data
	if p.encrypt {
		decrypt_data := xxtea.Decrypt(bodyData, []byte(ENCRYPT_KEY))
		return decrypt_data, nil
	} else {
		return bodyData, nil
	}
}

// goroutine safe
func (p *MsgParser) Write(conn *TCPConn, args ...[]byte) error {
	fmt.Println("tcp_msg.go.MsgParse.Write", conn, args)
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

	msgLen = msgLen + uint32(p.lenExtHeadLen)
	msg := make([]byte, uint32(p.lenMsgLen)+msgLen)

	// write len
	switch p.lenMsgLen {
	case 1:
		msg[0] = byte(msgLen)
	case 2:
		if p.littleEndian {
			binary.LittleEndian.PutUint16(msg, uint16(msgLen))
		} else {
			binary.BigEndian.PutUint16(msg, uint16(msgLen))
		}
	case 4:
		if p.littleEndian {
			binary.LittleEndian.PutUint32(msg, msgLen)
		} else {
			binary.BigEndian.PutUint32(msg, msgLen)
		}
	}

	// write head data (忽略)

	// write data
	l := p.lenMsgLen + p.lenExtHeadLen
	for i := 0; i < len(args); i++ {
		copy(msg[l:], args[i])
		l += len(args[i])
	}

	fmt.Println("tcp_msg.go.MsgParse.Write Data Len:", l)
	conn.Write(msg)

	return nil
}

// goroutine safe
func (p *MsgParser) Write1(conn *TCPConn, args ...[]byte) error {
	fmt.Println("tcp_msg.go.MsgParse.Write1", conn, args)
	l := 0
	var dataLen uint32
	for i := 0; i < len(args); i++ {
		l += len(args[i])
		dataLen += uint32(len(args[i]))
	}

	data := make([]byte, dataLen)
	for i := 0; i < len(args); i++ {
		copy(data[l:], args[i])
		l += len(args[i])
	}

	if p.encrypt {
		data = xxtea.Encrypt(data, []byte(ENCRYPT_KEY))
	}

	msgLen := uint32(len(data))
	//fmt.Println("tcp_msg.go.MsgParse.Write config ", p)
	//fmt.Println("tcp_msg.go.MsgParse.Write msgLen:", msgLen, ",data:")
	// check len
	if msgLen > p.maxMsgLen {
		return errors.New("message too long")
	} else if msgLen < p.minMsgLen {
		return errors.New("message too short")
	}

	msg := make([]byte, uint32(p.lenMsgLen)+msgLen)

	// write len
	switch p.lenMsgLen {
	case 1:
		msg[0] = byte(msgLen)
	case 2:
		if p.littleEndian {
			binary.LittleEndian.PutUint16(msg, uint16(msgLen))
		} else {
			binary.BigEndian.PutUint16(msg, uint16(msgLen))
		}
	case 4:
		if p.littleEndian {
			binary.LittleEndian.PutUint32(msg, msgLen)
		} else {
			binary.BigEndian.PutUint32(msg, msgLen)
		}
	}

	copy(msg[p.lenMsgLen:], []byte(data))
	//fmt.Println("############")
	//fmt.Println(msg)
	//fmt.Println("############")
	conn.Write(msg)

	return nil
}
