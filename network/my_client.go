package network

import (
	"fmt"
	"net"
	"time"
)

type MyClient struct {
	Addr            string
	ConnectInterval time.Duration
	PendingWriteNum int
	NewAgent        func(*TCPConn) Agent
	closeFlag       bool
	conn            *TCPConn

	// msg parser
	LenMsgLen    int
	MinMsgLen    uint32
	MaxMsgLen    uint32
	LittleEndian bool
	msgParser    *MsgParser
}

func (client *MyClient) Start() {
	agent := client.NewAgent(client.conn)
	go agent.Run()

	// cleanup
	//client.conn.Destroy()
	//agent.OnClose()
}

func (client *MyClient) Init(addr string) {

	client.Addr = addr
	client.closeFlag = false
	client.PendingWriteNum = 2000
	client.LenMsgLen = 2
	client.MinMsgLen = 2
	client.MaxMsgLen = 4096
	client.LittleEndian = false
	// msg parser
	msgParser := NewMsgParser()
	msgParser.SetMsgLen(client.LenMsgLen, client.MinMsgLen, client.MaxMsgLen)
	msgParser.SetByteOrder(client.LittleEndian)
	client.msgParser = msgParser

	client.connect()
}

func (client *MyClient) dial() net.Conn {
	conn, err := net.Dial("tcp", client.Addr)
	if err == nil {
		fmt.Println("connect success!", client.Addr)
		return conn
	}
	fmt.Println("connect fail!", client.Addr, err.Error())
	return nil
}

func (client *MyClient) connect() {

	conn := client.dial()
	if conn == nil {
		return
	}

	client.conn = newTCPConn(conn, client.PendingWriteNum, client.msgParser)

}

func (client *MyClient) ReadMsg() ([]byte, error) {
	return client.conn.ReadMsg()
}

func (client *MyClient) WriteMsg(args ...[]byte) error {
	return client.conn.WriteMsg(args...)
}

func (client *MyClient) Close() {
	client.closeFlag = true
}
