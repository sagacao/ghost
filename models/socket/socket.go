package socket

import (
	"bytes"
	"fmt"
	"ghost/network"
	"ghost/services/utils"
	"math"
	"sync"
	"time"

	. "ghost/global"
)

const defaultAsyncMsgLen = 1e3

var socketMsgPool *sync.Pool

type RawLogHeader struct {
	cmd     uint16
	logtype uint32
}

type SocketEngine struct {
	lock       sync.Mutex
	msgChanLen int64 `json:"chanlen"`
	msgChan    chan *utils.LogMsg

	signalChan chan string
	wg         sync.WaitGroup

	client    *network.TCPClient
	logHeader *RawLogHeader
	agent     network.Agent
}

func NewSocketEngine(worker network.Processor) *SocketEngine {
	se := new(SocketEngine)
	err := se.Init()
	if err != nil {
		panic(err)
	}

	se.newSocketClient(worker)

	se.msgChan = make(chan *utils.LogMsg, se.msgChanLen)
	socketMsgPool = &sync.Pool{
		New: func() interface{} {
			return &utils.LogMsg{}
		},
	}
	se.signalChan = make(chan string, 1)

	se.wg.Add(1)
	go se.start()
	return se
}

func (se *SocketEngine) start() {
	gameOver := false
	for {
		select {
		case bm := <-se.msgChan:
			se.writeMsg(bm.When, bm.Msg, bm.Database)
			socketMsgPool.Put(bm)
		case sg := <-se.signalChan:
			// Now should only send "flush" or "close" to bl.signalChan
			for {
				if len(se.msgChan) > 0 {
					bm := <-se.msgChan
					se.writeMsg(bm.When, bm.Msg, bm.Database)
					socketMsgPool.Put(bm)
					continue
				}
				break
			}
			if sg == "close" {
				se.Destroy()
				gameOver = true
			}
			se.wg.Done()
		}
		if gameOver {
			break
		}
	}
}

func (se *SocketEngine) Init() error {
	//el.lock.Lock()
	//defer el.lock.Unlock()
	se.msgChanLen = Config.Output.ChanLen
	if se.msgChanLen <= 0 {
		se.msgChanLen = defaultAsyncMsgLen
	}

	se.logHeader = &RawLogHeader{cmd: 10010, logtype: 2}

	return nil
}

func (se *SocketEngine) newSocketClient(worker network.Processor) {
	se.client = new(network.TCPClient)
	se.client.Addr = Config.Tcp.Net
	se.client.ConnNum = 1
	se.client.ConnectInterval = 3 * time.Second
	se.client.PendingWriteNum = 1000
	se.client.AutoReconnect = true
	se.client.LenMsgLen = 4
	se.client.MaxMsgLen = math.MaxUint32
	se.client.NewAgent = func(sess *network.ConnSession) network.Agent {
		se.agent = &Agent{Sess: sess}
		return se.agent
	}

	if se.client == nil {
		panic("err")
	}
}

func (se *SocketEngine) newLaPool(nethost string) error {

	return nil
}

func (se *SocketEngine) Write(database string, json map[string]interface{}) {
	when := time.Now()

	lm := socketMsgPool.Get().(*utils.LogMsg)
	lm.Database = database
	lm.Msg = json
	lm.When = when
	se.msgChan <- lm
}

func (se *SocketEngine) WriteRaw(database string, json map[string]interface{}) {

}

func (se *SocketEngine) log(database string, jsonmsg map[string]interface{}) {
	var buf bytes.Buffer
	buf.WriteString(" [")
	buf.WriteString(database)
	buf.WriteString("] ")
	for k, v := range jsonmsg {
		buf.WriteString(k)
		buf.WriteString(":")
		buf.WriteString(fmt.Sprintf("%v", v))
		buf.WriteString(",")
	}

	Log.Debug("%v", buf.String())
}

func (se *SocketEngine) writeMsg(when time.Time, jsonmsg map[string]interface{}, database string) error {
	var msg []byte
	var err error

	if se.agent != nil {
		err = se.agent.WriteMsg(se.logHeader.cmd, msg)
	} else {
		err = fmt.Errorf("socket agent closed!\n")
	}
	if err != nil {
		Log.Warn("write error: [%v] ", err)
		se.log(database, jsonmsg)
	}
	return err
}

func (se *SocketEngine) Destroy() {
	se.signalChan <- "close"
	se.wg.Wait()
	close(se.msgChan)
	close(se.signalChan)

	if se.client != nil {
		se.client.Close()
	}
}

func (se *SocketEngine) Flush() {
	se.signalChan <- "flush"
	se.wg.Wait()
	se.wg.Add(1)
}
