package network

import (
	"net"
	"sync"
	"time"

	. "ghost/global"
)

type TCPServer struct {
	Addr            string
	MaxConnNum      int
	PendingWriteNum int
	NewAgent        func(*ConnSession) Agent
	ln              net.Listener
	conns           SessionSet
	mutexConns      sync.Mutex
	wgLn            sync.WaitGroup
	wgConns         sync.WaitGroup

	// msg parser
	LenMsgLen    int
	MinMsgLen    uint32
	MaxMsgLen    uint32
	LittleEndian bool
	msgParser    *MsgParser
}

func (server *TCPServer) Start() {
	Console.Info("TCP Server Start On :[%v]", server.Addr)
	Log.Info("TCP Server Start On :[%v]", server.Addr)
	server.init()
	go server.run()
}

func (server *TCPServer) init() {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		Log.Critical("%v", err)
		return
	}

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 100
		Log.Info("invalid MaxConnNum, reset to %v", server.MaxConnNum)
	}
	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 100
		Log.Info("invalid PendingWriteNum, reset to %v", server.PendingWriteNum)
	}
	if server.NewAgent == nil {
		Log.Critical("NewAgent must not be nil")
	}

	server.ln = ln
	server.conns = make(SessionSet)

	// msg parser
	msgParser := NewMsgParser()
	msgParser.SetMsgLen(uint32(server.LenMsgLen), server.MinMsgLen, server.MaxMsgLen)
	msgParser.SetByteOrder(server.LittleEndian)
	server.msgParser = msgParser
}

func (server *TCPServer) run() {
	server.wgLn.Add(1)
	defer server.wgLn.Done()

	var tempDelay time.Duration
	for {
		conn, err := server.ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				Log.Info("accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		tempDelay = 0

		server.mutexConns.Lock()
		if len(server.conns) >= server.MaxConnNum {
			server.mutexConns.Unlock()
			conn.Close()
			Log.Debug("too many connections")
			continue
		}
		server.conns[conn] = struct{}{}
		server.mutexConns.Unlock()

		server.wgConns.Add(1)

		sess := newConnSession(conn, server.PendingWriteNum, server.msgParser)
		agent := server.NewAgent(sess)
		go func() {
			Console.Info("new connection from: [%v]", sess.RemoteAddr())
			Log.Info("new connection from: [%v]", sess.RemoteAddr())
			agent.Run()

			// cleanup
			sess.Close()
			//sess.Destroy()
			server.mutexConns.Lock()
			delete(server.conns, conn)
			server.mutexConns.Unlock()
			agent.OnClose()

			server.wgConns.Done()
		}()
	}
}

func (server *TCPServer) Close() {
	Console.Info("TCP Server Safe Close")
	Log.Info("TCP Server Safe Close")
	server.ln.Close()
	server.wgLn.Wait()

	server.mutexConns.Lock()
	for conn := range server.conns {
		conn.Close()
	}
	server.conns = nil
	server.mutexConns.Unlock()
	server.wgConns.Wait()
}
