package mongo

import (
	"bytes"
	"fmt"
	"ghost/services/utils"
	"sync"
	"time"

	. "ghost/global"

	"gopkg.in/mgo.v2"
)

const defaultAsyncMsgLen = 1e3

var logMsgPool *sync.Pool

// MongoEngine engine define
type MongoEngine struct {
	lock       sync.Mutex
	msgChanLen int64
	url        string

	msgChan chan *utils.LogMsg

	signalChan chan string
	wg         sync.WaitGroup
	mgo        *mgo.Session
}

// NewMongoEngine engine define
func NewMongoEngine() *MongoEngine {
	me := new(MongoEngine)
	err := me.Init()
	if err != nil {
		panic(err)
	}

	me.msgChan = make(chan *utils.LogMsg, me.msgChanLen)
	logMsgPool = &sync.Pool{
		New: func() interface{} {
			return &utils.LogMsg{}
		},
	}
	me.signalChan = make(chan string, 1)

	me.mgo, err = mgo.Dial(me.url)
	if err != nil {
		Log.Warn("NewMongoEngine error: url(%v) ", me.url)
		panic(err)
	}
	Log.Info("Connect to mongodb(%v) success", me.url)
	//Optional. Switch the session to a monotonic behavior.
	me.mgo.SetMode(mgo.Monotonic, true)
	me.mgo.SetSafe(&mgo.Safe{W: 0})

	me.wg.Add(1)
	go me.start()
	return me
}

func (me *MongoEngine) start() {
	gameOver := false
	for {
		select {
		case bm := <-me.msgChan:
			me.writeMsg(bm.When, bm.Msg, bm.Database)
			logMsgPool.Put(bm)
		case sg := <-me.signalChan:
			// Now should only send "flush" or "close" to bl.signalChan
			for {
				if len(me.msgChan) > 0 {
					bm := <-me.msgChan
					me.writeMsg(bm.When, bm.Msg, bm.Database)
					logMsgPool.Put(bm)
					continue
				}
				break
			}
			if sg == "close" {
				me.Destroy()
				gameOver = true
			}
			me.wg.Done()
		}
		if gameOver {
			break
		}
	}
}

// Init init function
func (me *MongoEngine) Init() error {
	//el.lock.Lock()
	//defer el.lock.Unlock()
	me.msgChanLen = Config.Mongo.ChanLen
	if me.msgChanLen <= 0 {
		me.msgChanLen = defaultAsyncMsgLen
	}
	me.url = Config.Mongo.URL
	if me.url == "" {
		me.url = "127.0.0.1:27017"
	}
	return nil
}

func (me *MongoEngine) Write(database string, json map[string]interface{}) {
	when := time.Now()

	lm := logMsgPool.Get().(*utils.LogMsg)
	lm.Database = database
	lm.Msg = json
	lm.When = when
	me.msgChan <- lm
}

// WriteRaw raw message interface
func (me *MongoEngine) WriteRaw(database string, json map[string]interface{}) {

}

func (me *MongoEngine) log(database string, jsonmsg map[string]interface{}) {
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

func (me *MongoEngine) writeMsg(when time.Time, jsonmsg map[string]interface{}, database string) error {
	logname, ok := jsonmsg["logname"].(string)
	if !ok {
		Log.Warn("no logname found ")
		me.log(database, jsonmsg)
		return nil
	}
	c := me.mgo.DB(database).C(logname)
	err := c.Insert(jsonmsg)
	if err != nil {
		Log.Warn("write error: [%v] ", err)
		me.log(database, jsonmsg)
	}
	return err
}

// Destroy destroy
func (me *MongoEngine) Destroy() {
	me.signalChan <- "close"
	me.wg.Wait()
	close(me.msgChan)
	close(me.signalChan)
	me.mgo.Close()
}

// Flush flush
func (me *MongoEngine) Flush() {
	me.signalChan <- "flush"
	me.wg.Wait()
	me.wg.Add(1)
}

///////////////////////
func (me *MongoEngine) createMongoIndex() {
	timer := time.NewTimer(time.Second * 5)
	go func() {
		<-timer.C

		Log.Debug("createMongoIndex Timestamp(%s)", time.Now())
	}()
}
