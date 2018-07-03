package format

import (
	"fmt"
	"ghost/services/utils"
	"path"
	"sync"
	"time"

	. "ghost/global"
)

type FormatEngine struct {
	lock sync.Mutex

	msgChanLen int64
	msgChan    chan *utils.LogMsg

	signalChan chan string
	wg         sync.WaitGroup
	writer     *fileLogWriter
	fwpool     utils.Pool

	zoneid   uint32
	filepath string
	filename string
}

var logMsgPool *sync.Pool

func NewFormatEngine(zoneid uint32) *FormatEngine {
	fe := &FormatEngine{zoneid: zoneid}
	err := fe.Init()
	if err != nil {
		panic(err)
	}

	fe.msgChan = make(chan *utils.LogMsg, fe.msgChanLen)
	logMsgPool = &sync.Pool{
		New: func() interface{} {
			return &utils.LogMsg{}
		},
	}
	fe.signalChan = make(chan string, 1)

	err = fe.newWriterPool()
	if err != nil {
		Log.Critical("NewFormatEngine err : %v", err)
	}

	go fe.start()
	return fe
}

func (fe *FormatEngine) start() {
	gameOver := false
	for {
		select {
		case bm := <-fe.msgChan:
			fe.writeMsg(bm.When, bm.Msg, bm.Database)
			logMsgPool.Put(bm)
		case sg := <-fe.signalChan:
			// Now should only send "flush" or "close" to bl.signalChan
			for {
				if len(fe.msgChan) > 0 {
					bm := <-fe.msgChan
					fe.writeMsg(bm.When, bm.Msg, bm.Database)
					logMsgPool.Put(bm)
					continue
				}
				break
			}
			if sg == "close" {
				fe.Destroy()
				gameOver = true
			}
			fe.wg.Done()
		}
		if gameOver {
			break
		}
	}
}

func (fe *FormatEngine) Init() error {
	//el.lock.Lock()
	//defer el.lock.Unlock()
	fe.msgChanLen = Config.Output.ChanLen
	if fe.msgChanLen <= 0 {
		fe.msgChanLen = 1e3
	}

	realpath := fmt.Sprintf(Config.File.FilePath, fe.zoneid)
	fe.filepath, fe.filename = path.Split(realpath)
	err := createFilePath(fe.filepath)
	if err != nil {
		return err
	}

	return nil
}

func (fe *FormatEngine) Write(database string, jsonmsg map[string]interface{}) {
	//fmt.Printf("logtime: %v", json["logtime"].(string))
	lm := logMsgPool.Get().(*utils.LogMsg)
	lm.Database = database
	lm.Msg = jsonmsg
	lm.When = time.Now()
	fe.msgChan <- lm
}

func (fe *FormatEngine) WriteRaw(database string, json map[string]interface{}) {

}

func (fe *FormatEngine) newWriterPool() error {
	fw, err := newFileWriter(fmt.Sprintf("%s%s", fe.filepath, fe.filename), Config.File.MaxLines, Config.File.MaxSize)
	if err != nil {
		Log.Critical("newFileWriter err: %v", err)
		return err
	}
	fe.writer = fw

	return nil

	// var err error
	// fe.fwpool, err = utils.NewTaskPool(&utils.PoolConfig{
	// 	Factory: func() (interface{}, error) {
	// 		return newFileWriter(fmt.Sprintf("%s%s", fe.filepath, fe.filename), Config.File.MaxLines, Config.File.MaxSize)
	// 	},
	// 	Close: func(interface{}) error {
	// 		return nil
	// 	},
	// })
	// if err != nil {
	// 	Console.Warn("newEsPool Create pool error")
	// }
	// return err
}

func (fe *FormatEngine) getWriter() *fileLogWriter {
	return fe.writer

	// fwiface, err := fe.fwpool.Get()
	// if err != nil {
	// 	return nil
	// }
	// fw, ok := fwiface.(*fileLogWriter)
	// if !ok {
	// 	return nil
	// }
	// return fw
}

func (fe *FormatEngine) writeMsg(when time.Time, jsonmsg map[string]interface{}, database string) error {
	w := fe.getWriter()
	if w == nil {
		Log.Critical("writeMsg no writer found.")
	}

	msg, ok := jsonmsg["msg"].([]byte)
	if ok {
		w.WriteMsg(when, string(msg))
	}

	return nil
}

func (fe *FormatEngine) Destroy() {
	fe.signalChan <- "close"
	fe.wg.Wait()
	close(fe.msgChan)
	close(fe.signalChan)
}

func (fe *FormatEngine) Flush() {
	fe.signalChan <- "flush"
	fe.wg.Wait()
	fe.wg.Add(1)
}
