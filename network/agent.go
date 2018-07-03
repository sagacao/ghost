package network

import (
	"net"
	"reflect"

	. "ghost/global"
)

type Agent interface {
	Run()
	OnClose()

	Bind(idx uint32)
	GetIndex() uint32

	WriteMsg(cmd uint16, msg interface{}) error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close()
	Destroy()
}

type AgentInstance struct {
	Index  uint32
	Sess   Session
	Worker Processor
}

func (a *AgentInstance) Run() {
	for {
		data, err := a.Sess.ReadMsg()
		if err != nil {
			Log.Debug("read message: %v", err)
			Console.Info("read message: %v", err)
			break
		}

		//		Console.Info("ReadMsg:[%v]", data)

		if a.Worker != nil {
			cmd, msg, err := a.Worker.Unmarshal(data)
			//Console.Info("ReadMsg:[%v][%v]", cmd, msg)
			if err == nil {
				err = a.Worker.Route(cmd, msg, a)
				if err != nil {
					Log.Debug("route message error: %v", err)
					break
				}
			} else {
				Log.Debug("unmarshal message error: %v", err)
			}
		}
	}
}

func (a *AgentInstance) OnClose() {
	Log.Error("agent OnClose")
}

func (a *AgentInstance) Bind(idx uint32) {
	a.Index = idx
}

func (a *AgentInstance) GetIndex() uint32 {
	return a.Index
}

func (a *AgentInstance) WriteMsg(cmd uint16, msg interface{}) error {
	if a.Worker != nil {
		data, err := a.Worker.Marshal(cmd, msg)
		if err != nil {
			Log.Error("marshal message %v error: %v", reflect.TypeOf(msg), err)
			return err
		}
		err = a.Sess.WriteMsg(data...)
		if err != nil {
			Log.Error("write message %v error: %v", reflect.TypeOf(msg), err)
			return err
		}
	}
	return nil
}

func (a *AgentInstance) LocalAddr() net.Addr {
	return a.Sess.LocalAddr()
}

func (a *AgentInstance) RemoteAddr() net.Addr {
	return a.Sess.RemoteAddr()
}

func (a *AgentInstance) Close() {
	a.Sess.Close()
}

func (a *AgentInstance) Destroy() {
	a.Sess.Destroy()
}

///////////////////////////////////////////////////
