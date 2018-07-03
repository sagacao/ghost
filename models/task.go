package models

import (
	"encoding/json"
	"ghost/models/elastic"
	"ghost/models/format"
	"ghost/models/mongo"
	"ghost/models/mysql"
	"ghost/models/report"
	"ghost/models/statistic"
	"ghost/services/utils"
	"io/ioutil"
	"sync"

	. "ghost/global"
)

var m *utils.Map
var template map[string]interface{}

type Task struct {
	zoneid uint32
	lock   sync.Mutex

	Es  Engine
	Db  Engine
	Fw  Engine
	Se  Engine
	Mgo Engine
}

func (t *Task) Destory() {
	if t.Es != nil {
		t.Es.Destroy()
	}

	if t.Db != nil {
		t.Db.Destroy()
	}

	if t.Fw != nil {
		t.Fw.Destroy()
	}

	if t.Se != nil {
		t.Se.Destroy()
	}

	if t.Mgo != nil {
		t.Mgo.Destroy()
	}
}

func NewTask() *Task {
	return &Task{zoneid: 0}
}

func (t *Task) Init(zoneid uint32) bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.zoneid != 0 {
		Log.Info("Task had inited .[src: %v --> new: %v] ", t.zoneid, zoneid)
		return false
	}
	t.zoneid = zoneid

	t.Fw = format.NewFormatEngine(zoneid)

	if Config.Optional.Mysql {
		t.Db = mysql.NewMysqlEngine(zoneid)
	}

	if Config.Optional.Elastic {
		t.Es = elastic.NewElasticEngine()
		if t.Es != nil {
			Log.Info("write template to elastic.")
			t.Es.WriteRaw("_template", template)
		}
	}

	if Config.Optional.Mongo {
		t.Mgo = mongo.NewMongoEngine()
	}

	t.Se = statistic.NewStatisticEngine()
	return true
}

func PushTask(zoneid uint32, t *Task) {
	m.Set(zoneid, t)
}

func PopTask(zoneid uint32) {
	task := GetTask(zoneid)
	if task != nil {
		task.Destory()
	}
	m.Del(zoneid)
}

func GetTask(zoneid uint32) *Task {
	task := m.Get(zoneid)
	if task != nil {
		t, ok := task.(*Task)
		if ok && t != nil {
			return t
		}
	}
	return nil
}

func init() {
	rawdata, err := ioutil.ReadFile("conf/template.json")
	if err != nil {
		panic(err)
	}
	m = new(utils.Map)
	template = make(map[string]interface{})
	err = json.Unmarshal(rawdata, &template)
	if err != nil {
		panic(err)
	}
}
