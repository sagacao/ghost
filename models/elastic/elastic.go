package elastic

import (
	"bytes"
	"fmt"
	"ghost/services/utils"
	"sync"
	"time"

	. "ghost/global"

	"github.com/OwnLocal/goes"
)

const defaultAsyncMsgLen = 1e3

var logMsgPool *sync.Pool

type ElasticEngine struct {
	lock       sync.Mutex
	msgChanLen int64  `json:"chanlen"`
	host       string `json:"host"`
	port       string `json:"port"`

	msgChan chan *utils.LogMsg

	signalChan chan string
	wg         sync.WaitGroup
	es         *goes.Client

	isusepool bool
	espool    utils.Pool
}

func NewElasticEngine() *ElasticEngine {
	el := new(ElasticEngine)
	err := el.Init()
	if err != nil {
		panic(err)
	}

	el.msgChan = make(chan *utils.LogMsg, el.msgChanLen)
	logMsgPool = &sync.Pool{
		New: func() interface{} {
			return &utils.LogMsg{}
		},
	}
	el.signalChan = make(chan string, 1)

	el.es = goes.NewClient(el.host, el.port)

	if el.isusepool {
		el.newEsPool()
	}

	el.wg.Add(1)
	go el.start()
	return el
}

func (el *ElasticEngine) start() {
	gameOver := false
	for {
		select {
		case bm := <-el.msgChan:
			el.writeMsg(bm.When, bm.Msg, bm.Database)
			logMsgPool.Put(bm)
		case sg := <-el.signalChan:
			// Now should only send "flush" or "close" to bl.signalChan
			for {
				if len(el.msgChan) > 0 {
					bm := <-el.msgChan
					el.writeMsg(bm.When, bm.Msg, bm.Database)
					logMsgPool.Put(bm)
					continue
				}
				break
			}
			if sg == "close" {
				el.Destroy()
				gameOver = true
			}
			el.wg.Done()
		}
		if gameOver {
			break
		}
	}
}

func (el *ElasticEngine) Init() error {
	//el.lock.Lock()
	//defer el.lock.Unlock()
	el.msgChanLen = Config.Output.ChanLen
	if el.msgChanLen <= 0 {
		el.msgChanLen = defaultAsyncMsgLen
	}
	el.host = Config.Output.Host
	if el.host == "" {
		el.host = "127.0.0.1"
	}
	el.port = Config.Output.Port
	if el.port == "" {
		el.port = "9200"
	}
	el.isusepool = true
	return nil
}

func (el *ElasticEngine) Write(database string, json map[string]interface{}) {
	when := time.Now()

	logname, ok := json["logname"]
	if !ok {
		logname = database
	}

	server, ok := json["serverid"]
	if !ok {
		server = "0"
	}

	lm := logMsgPool.Get().(*utils.LogMsg)
	lm.Database = fmt.Sprintf("%s_%s", logname.(string), server.(string))
	lm.Msg = json
	lm.When = when
	el.msgChan <- lm
}

func (el *ElasticEngine) WriteRaw(database string, json map[string]interface{}) {
	d := goes.Document{
		Index:  database,
		Type:   "template_logging", //fmt.Sprintf("template_logging", database),
		Fields: json,
	}
	_, err := el.es.Index(d, nil)
	if err != nil {
		Log.Warn("write error: [%v] ", err)
	}
}

func (el *ElasticEngine) newEsPool() {
	var err error
	el.espool, err = utils.NewTaskPool(&utils.PoolConfig{
		IdleNum: 10,
		Factory: func() (interface{}, error) {
			return goes.NewClient(el.host, el.port), nil
		},
		Close: func(interface{}) error {
			return nil
		},
	})
	if err != nil {
		Console.Warn("newEsPool Create pool error")
	}
}

func (el *ElasticEngine) getIdleClient() *goes.Client {
	esiface, err := el.espool.Get()
	if err != nil {
		return nil
	}
	es, ok := esiface.(*goes.Client)
	if !ok {
		return nil
	}
	return es
}

func (el *ElasticEngine) asyncIndex(index, types string, jsonmsg map[string]interface{}) func() (interface{}, error) {
	return func() (interface{}, error) {
		_, err := el.es.Index(goes.Document{Index: index, Type: types, Fields: jsonmsg}, nil)
		if err != nil {
			Log.Warn("write error: [%v] ", err)
			el.log(index, jsonmsg)
		}
		return nil, err
	}
}

func (el *ElasticEngine) log(database string, jsonmsg map[string]interface{}) {
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

func (el *ElasticEngine) writeMsg(when time.Time, jsonmsg map[string]interface{}, database string) error {
	var timestamp time.Time
	logtime, ok := jsonmsg["logtime"]
	if ok {
		timestamp, _ = time.Parse("2006-01-02 15:04:05", logtime.(string))
	} else {
		timestamp = when
	}
	jsonmsg["@timestamp"] = timestamp.Format(time.RFC3339)

	syncIndex := el.asyncIndex(
		fmt.Sprintf("%s-%04d.%02d.%02d", database, timestamp.Year(), timestamp.Month(), timestamp.Day()),
		jsonmsg["logname"].(string),
		jsonmsg)

	utils.Future(syncIndex)

	return nil
	// d := goes.Document{
	// 	Index:  fmt.Sprintf("%s-%04d.%02d.%02d", database, timestamp.Year(), timestamp.Month(), timestamp.Day()),
	// 	Type:   jsonmsg["logname"].(string),
	// 	Fields: jsonmsg,
	// }
	// var err error
	// if el.espool != nil {
	// 	es := el.getIdleClient()
	// 	if es != nil {
	// 		_, err = es.Index(d, nil)
	// 		el.espool.Put(es)
	// 	} else {
	// 		_, err = el.es.Index(d, nil)
	// 	}
	// } else {
	// 	_, err = el.es.Index(d, nil)
	// }
	// if err != nil {
	// 	Log.Warn("write error: [%v] ", err)
	// 	el.log(database, jsonmsg)
	// }
	// return err
}

func (el *ElasticEngine) Destroy() {
	el.signalChan <- "close"
	el.wg.Wait()
	close(el.msgChan)
	close(el.signalChan)

	el.espool.Release()
}

func (el *ElasticEngine) Flush() {
	el.signalChan <- "flush"
	el.wg.Wait()
	el.wg.Add(1)
}
