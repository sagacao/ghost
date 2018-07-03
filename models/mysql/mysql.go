package mysql

import (
	"fmt"
	. "ghost/global"
	"ghost/services/utils"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
)

var dbMsgPool *sync.Pool

type MysqlEngine struct {
	zoneid     uint32
	dbname     string
	dataSource string
	engine     *xorm.Engine
	done       bool

	msgChan    chan *utils.LogMsg
	msgChanLen int64
	signalChan chan string
	wg         sync.WaitGroup
}

func NewMysqlEngine(zoneid uint32) *MysqlEngine {
	me := MysqlEngine{zoneid: zoneid, done: false}
	me.Init()
	err := createDatabase(me.dbname)
	if err != nil {
		Log.Warn("Create database '%s' Error : (%v)", me.dbname, err)
	}

	me.engine, err = xorm.NewEngine("mysql", me.dataSource)
	if err != nil {
		Log.Warn("Create db engine '%s' Error : (%v)", me.dbname, err)
	}
	me.engine.Charset("utf8")
	me.engine.SetMaxIdleConns(10)
	me.engine.ShowSQL(true)
	me.engine.Logger().SetLevel(core.LOG_DEBUG)

	me.CreateTables()
	err = me.engine.Ping()
	if err != nil {
		Log.Warn("db engine '%s' Ping Error : (%v)", me.dbname, err)
		//timer
	}

	me.msgChan = make(chan *utils.LogMsg, me.msgChanLen)
	dbMsgPool = &sync.Pool{
		New: func() interface{} {
			return &utils.LogMsg{}
		},
	}

	me.wg.Add(1)
	go me.start()
	me.startTimer()

	return &me
}

func (me *MysqlEngine) Init() error {
	//el.lock.Lock()
	//defer el.lock.Unlock()
	me.msgChanLen = Config.Output.ChanLen
	if me.msgChanLen <= 0 {
		me.msgChanLen = 1000
	}

	me.dbname = fmt.Sprintf("%s_%v", Config.Mysql.Db, me.zoneid)
	me.dataSource = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8",
		Config.Mysql.User, Config.Mysql.Password, Config.Mysql.Host, me.dbname)

	return nil
}

func (me *MysqlEngine) Write(database string, json map[string]interface{}) {
	when := time.Now()

	lm := dbMsgPool.Get().(*utils.LogMsg)
	lm.Database = me.dbname
	lm.Msg = json
	lm.When = when
	me.msgChan <- lm
}

func (me *MysqlEngine) WriteRaw(database string, json map[string]interface{}) {

}

func (me *MysqlEngine) start() {
	gameOver := false
	for {
		select {
		case bm := <-me.msgChan:
			me.writeMsg(bm.When, bm.Msg, bm.Database)
			dbMsgPool.Put(bm)
		case sg := <-me.signalChan:
			// Now should only send "flush" or "close" to bl.signalChan
			for {
				if len(me.msgChan) > 0 {
					bm := <-me.msgChan
					me.writeMsg(bm.When, bm.Msg, bm.Database)
					dbMsgPool.Put(bm)
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

func (me *MysqlEngine) startTimer() {
	go func() {
		for {
			me.CreateTables()

			now := time.Now()
			next := now.Add(time.Hour * 24)
			next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
			t := time.NewTimer(next.Sub(now))
			<-t.C

			Log.Debug("Service StartTimer Timestamp(%s)", time.Now())
		}
	}()
}

func (me *MysqlEngine) CreateTables() {
	if !me.done {
		err := me.engine.Sync2(
			new(LogFilter),
			new(Account),
			new(AccountLogin),
			new(RoleLogin),
			new(RoleLogout),
			new(RoleInfo),
			new(OnlineState),
			new(LevelDistribution),
			new(TaskFinish),
			new(GuideInfo),
			new(MoneyProduct),
			new(MoneyConsume),
			new(LeagueInfo),
			new(LeagueData),
			new(LeagueActivity),
			new(ShopTrade),
			new(ChargeInfo))
		if err != nil {
			Log.Error("CreateTables error: (%v)", err)
		}

		me.done = true
	} else {
		me.CreateOneTable("accountlogin", new(AccountLogin))
		me.CreateOneTable("rolelogout", new(RoleLogout))
		me.CreateOneTable("onlinestate", new(OnlineState))
		me.CreateOneTable("rolelogin", new(RoleLogin))
		me.CreateOneTable("shoptrade", new(ShopTrade))
		me.CreateOneTable("taskfinish", new(TaskFinish))
		me.CreateOneTable("guideinfo", new(GuideInfo))
		me.CreateOneTable("moneyproduct", new(MoneyProduct))
		me.CreateOneTable("moneyconsume", new(MoneyConsume))
		me.CreateOneTable("leaguedata", new(LeagueData))
		me.CreateOneTable("leagueactivity", new(LeagueActivity))
	}
}

func (me *MysqlEngine) CreateOneTable(table string, bean interface{}) {
	tablename := gettablename(table, time.Now().Add(time.Hour*24))
	err := me.engine.Table(tablename).CreateTable(bean)
	if err != nil {
		Log.Error("CreateOneTable error: (%v)", err)
	}
	err = me.engine.Table(tablename).CreateIndexes(bean)
	if err != nil {
		Log.Error("CreateOneTable error: (%v)", err)
	}

	Log.Info("CreateOneTable >>>>>>> `%s` success ", tablename)
}

func (me *MysqlEngine) writeMsg(when time.Time, jsonmsg map[string]interface{}, database string) {
	logname, ok := jsonmsg["logname"].(string)
	if !ok {
		Log.Error("writeMsg error (no logname): (%v) ", me.zoneid)
		return
	}

	if h := Handlers[logname]; h != nil {
		h(me, me.zoneid, jsonmsg)
	} else {
		Log.Error("db insert error: not regist the name of (%v) ", logname)
		return
	}
}

func (me *MysqlEngine) Destroy() {
	me.signalChan <- "close"
	me.wg.Wait()
	close(me.msgChan)
	close(me.signalChan)
	me.engine.Close()
}

func (me *MysqlEngine) Flush() {
	me.signalChan <- "flush"
	me.wg.Wait()
	me.wg.Add(1)
}
