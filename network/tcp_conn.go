package network

import (
	"fmt"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"github.com/rufeng18/tinyleaf/conf"
	"github.com/rufeng18/tinyleaf/log"
)

type ConnSet map[net.Conn]struct{}

type TCPConn struct {
	sync.Mutex
	conn      net.Conn
	writeChan chan []byte
	closeFlag bool
	verified  bool
	msgParser *MsgParser
}

func newTCPConn(conn net.Conn, pendingWriteNum int, msgParser *MsgParser) *TCPConn {
	tcpConn := new(TCPConn)
	tcpConn.conn = conn
	tcpConn.writeChan = make(chan []byte, pendingWriteNum)
	tcpConn.msgParser = msgParser
	log.Debug("%s connnection succ!!!", conn.RemoteAddr())
	go func() {
		timeout := time.NewTimer(conf.VerifyInterval)
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("[异常] ", err, "\n", string(debug.Stack()))
			}
			//fmt.Println("write goroutine is close")
			conn.Close()
			tcpConn.Lock()
			tcpConn.closeFlag = true
			tcpConn.Unlock()
		}()

		for {
			select {
			case b := <-tcpConn.writeChan:
				if b == nil {
					//fmt.Println("write goroutine b nil")
					return //break
				}
				//fmt.Println("write goroutine  data", b)
				_, err := conn.Write(b)
				if err != nil {
					//fmt.Println("write goroutine err", err.Error())
					return //break
				}
			case <-timeout.C:
				timeout.Stop()
				if !tcpConn.verified {
					log.Debug("%s timeout to verify , fail !!!", conn.RemoteAddr())
					return
				}
			}
		}

	}()

	return tcpConn
}

func newTCPConn1(conn net.Conn, pendingWriteNum int, msgParser *MsgParser) *TCPConn {
	tcpConn := new(TCPConn)
	tcpConn.conn = conn
	tcpConn.writeChan = make(chan []byte, pendingWriteNum)
	tcpConn.msgParser = msgParser

	go func() {
		for b := range tcpConn.writeChan {
			if b == nil {
				fmt.Println("write goroutine b nil")
				break
			}
			//fmt.Println("write goroutine  data", b)
			_, err := conn.Write(b)
			if err != nil {
				fmt.Println("write goroutine err", err.Error())
				break
			}
		}
		fmt.Println("write goroutine is close")
		conn.Close()
		tcpConn.Lock()
		tcpConn.closeFlag = true
		tcpConn.Unlock()
	}()

	return tcpConn
}

func (tcpConn *TCPConn) doDestroy() {
	tcpConn.conn.(*net.TCPConn).SetLinger(0)
	tcpConn.conn.Close()

	if !tcpConn.closeFlag {
		close(tcpConn.writeChan)
		tcpConn.closeFlag = true
	}
}

func (tcpConn *TCPConn) Destroy() {
	tcpConn.Lock()
	defer tcpConn.Unlock()

	tcpConn.doDestroy()
}

func (tcpConn *TCPConn) Close() {
	tcpConn.Lock()
	defer tcpConn.Unlock()
	if tcpConn.closeFlag {
		return
	}

	tcpConn.doWrite(nil)
	tcpConn.closeFlag = true
}

func (tcpConn *TCPConn) Verify() {
	tcpConn.verified = true
	log.Debug("%s conn verify succ!!!", tcpConn.conn.RemoteAddr())
}

func (tcpConn *TCPConn) doWrite(b []byte) {
	if len(tcpConn.writeChan) == cap(tcpConn.writeChan) {
		log.Debug("close conn: channel full")
		tcpConn.doDestroy()
		return
	}

	tcpConn.writeChan <- b
}

// b must not be modified by the others goroutines
func (tcpConn *TCPConn) Write(b []byte) {
	tcpConn.Lock()
	defer tcpConn.Unlock()
	if tcpConn.closeFlag || b == nil {
		return
	}

	tcpConn.doWrite(b)
}

func (tcpConn *TCPConn) Read(b []byte) (int, error) {
	return tcpConn.conn.Read(b)
}

func (tcpConn *TCPConn) LocalAddr() net.Addr {
	return tcpConn.conn.LocalAddr()
}

func (tcpConn *TCPConn) RemoteAddr() net.Addr {
	return tcpConn.conn.RemoteAddr()
}

func (tcpConn *TCPConn) ReadMsg() ([]byte, error) {
	return tcpConn.msgParser.Read(tcpConn)
}

func (tcpConn *TCPConn) WriteMsg(args ...[]byte) error {
	return tcpConn.msgParser.Write(tcpConn, args...)
}
