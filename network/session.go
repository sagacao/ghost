package network

import (
	"net"
	"sync"

	. "ghost/global"
)

type Session interface {
	ReadMsg() ([]byte, error)
	WriteMsg(args ...[]byte) error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close()
	Destroy()
}

type SessionSet map[net.Conn]struct{}

type ConnSession struct {
	sync.Mutex
	conn      net.Conn
	writeChan chan []byte
	closeFlag bool
	msgParser *MsgParser
}

func newConnSession(conn net.Conn, pendingWriteNum int, msgParser *MsgParser) *ConnSession {
	sess := new(ConnSession)
	sess.conn = conn
	sess.writeChan = make(chan []byte, pendingWriteNum)
	sess.msgParser = msgParser

	go func() {
		for b := range sess.writeChan {
			if b == nil {
				break
			}

			_, err := conn.Write(b)
			if err != nil {
				break
			}
		}

		conn.Close()
		sess.Lock()
		sess.closeFlag = true
		sess.Unlock()
	}()

	return sess
}

func (sess *ConnSession) doDestroy() {
	sess.conn.(*net.TCPConn).SetLinger(0)
	sess.conn.Close()

	if !sess.closeFlag {
		close(sess.writeChan)
		sess.closeFlag = true
	}
}

func (sess *ConnSession) Destroy() {
	sess.Lock()
	defer sess.Unlock()

	sess.doDestroy()
}

func (sess *ConnSession) Close() {
	sess.Lock()
	defer sess.Unlock()
	if sess.closeFlag {
		return
	}

	sess.doWrite(nil)
	sess.closeFlag = true
}

func (sess *ConnSession) doWrite(b []byte) {
	if len(sess.writeChan) == cap(sess.writeChan) {
		Log.Debug("close conn: channel full")
		sess.doDestroy()
		return
	}

	sess.writeChan <- b
}

// b must not be modified by the others goroutines
func (sess *ConnSession) Write(b []byte) {
	sess.Lock()
	defer sess.Unlock()
	if sess.closeFlag || b == nil {
		return
	}

	sess.doWrite(b)
}

func (sess *ConnSession) Read(b []byte) (int, error) {
	return sess.conn.Read(b)
}

func (sess *ConnSession) LocalAddr() net.Addr {
	return sess.conn.LocalAddr()
}

func (sess *ConnSession) RemoteAddr() net.Addr {
	return sess.conn.RemoteAddr()
}

func (sess *ConnSession) ReadMsg() ([]byte, error) {
	return sess.msgParser.Read(sess)
}

func (sess *ConnSession) WriteMsg(args ...[]byte) error {
	return sess.msgParser.Write(sess, args...)
}
