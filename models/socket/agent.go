package socket

import (
	"encoding/binary"
	"fmt"
	"ghost/network"
	"net"
	"reflect"

	. "ghost/global"
)

type Agent struct {
	Sess network.Session
}

func (a *Agent) Run() {
	for {
		data, err := a.Sess.ReadMsg()
		if err != nil {
			Log.Debug("read message: %v", err)
			Console.Info("read message: %v", err)
			break
		}
		Console.Info("ReadMsg:[%v]", data)
	}
}

func (a *Agent) OnClose() {
	Log.Error("AgentClient OnClose")
}

func (a *Agent) Bind(idx uint32) {
}

func (a *Agent) GetIndex() uint32 {
	return 0
}

func (a *Agent) pack(cmd uint16, msg interface{}) ([][]byte, error) {
	logheader := make([]byte, 6)
	binary.BigEndian.PutUint16(logheader, cmd)
	binary.BigEndian.PutUint32(logheader, 2)

	var err error
	// data
	data, ok := msg.([]byte)
	if !ok {
		err = fmt.Errorf("message [%v] Marshal data error", cmd)
	}
	return [][]byte{logheader, data}, err
}

func (a *Agent) WriteMsg(cmd uint16, msg interface{}) error {
	data, err := a.pack(cmd, msg)
	if err != nil {
		return err
	}

	err = a.Sess.WriteMsg(data...)
	if err != nil {
		Log.Error("write message %v error: %v", reflect.TypeOf(msg), err)
		return err
	}
	return nil
}

func (a *Agent) LocalAddr() net.Addr {
	return a.Sess.LocalAddr()
}

func (a *Agent) RemoteAddr() net.Addr {
	return a.Sess.RemoteAddr()
}

func (a *Agent) Close() {
	a.Sess.Close()
}

func (a *Agent) Destroy() {
	a.Sess.Destroy()
}
