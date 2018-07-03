package handler

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"ghost/network"
	"ghost/services/raw"
	"ghost/services/utils"
	"regexp"
	"sync"

	. "ghost/global"

	"ghost/models"
)

var taskmutex sync.RWMutex

var (
	rawExp = utils.ExRegexp{regexp.MustCompile(`(?P<logtime>\d{4}\-\d{1,12}\-\d{1,31} \d{1,24}:\d{2}:\d{2}) (?P<index>[A-Z]{2}_\d_\d) \[(?P<logname>\b\w+\b)\] \[(?P<level>\b\w+\b)\](?P<msg>.*)`)}
)

func catch() {
	if r := recover(); r != nil {
		Log.Critical("%v", r)
	}
}

///////////////////////////////////////////////////////
//var task = models.NewTask()

///////////////////////////////////////////////////////

func dumpmap(m map[string]interface{}) {
	Console.Info("Dump map -------------------> begin")
	for k, v := range m {
		Console.Info("%s=%v;", k, v)
	}
	Console.Info("Dump map -------------------> end")
}

func onRegister(a network.Agent, arg interface{}) {
	//defer catch()
	rawmsg, ok := arg.([]byte)
	if !ok {
		Log.Warn("->>>> error meybe protocol changed.")
		return
	}

	zoneid := binary.LittleEndian.Uint32(rawmsg[0:])
	taskmutex.Lock()
	task := models.GetTask(zoneid)
	if task == nil {
		task := models.NewTask()
		if task == nil {
			Log.Warn("onRegister task is nil.")
			taskmutex.Unlock()
			return
		}
		task.Init(zoneid)
		Log.Info("onRegister zone %v success ", zoneid)
		Console.Info("onRegister:zone %v success ", zoneid)
		models.PushTask(zoneid, task)
		a.Bind(zoneid)
	} else {
		Log.Info("onRegister zone %v had registed ", zoneid)
		Console.Info("onRegister:zone %v had registed ", zoneid)
		a.Bind(zoneid)
	}
	taskmutex.Unlock()
}

func onRawMessage(a network.Agent, arg interface{}) {
	rawmsg, ok := arg.([]byte)
	if !ok {
		Log.Warn("->>>> error meybe protocol changed.")
		return
	}

	//Console.Info("onRawMessage:%v", string(rawmsg))
	zoneid := a.GetIndex() //binary.LittleEndian.Uint32(rawmsg[0:])
	task := models.GetTask(zoneid)
	if task == nil {
		Log.Warn("onRawMessage task is nil. zone: %v ", zoneid)
		return
	}

	if task == nil {
		Log.Critical("onRawMessage zone %v closed.", zoneid)
		a.Destroy()
		return
	}
	result := rawExp.FindStringSubmatchMap(string(rawmsg[4:]))
	_, ok = result["logname"]
	if !ok {
		Log.Warn("onRawMessage Unmarshal error: [%v]", string(rawmsg[4:]))
		return
	}
	//dumpmap(result)
	if task.Es != nil {
		task.Es.Write(fmt.Sprintf("raw_%v", zoneid), result)
	}
}

func onFormatMessage(a network.Agent, arg interface{}) {
	rawmsg, ok := arg.([]byte)
	if !ok {
		Log.Warn("->>>> error meybe protocol changed.")
		return
	}

	//Console.Info("onFormatMessage:%v", string(rawmsg))
	zoneid := a.GetIndex() //binary.LittleEndian.Uint32(rawmsg[0:])
	task := models.GetTask(zoneid)
	if task == nil {
		Log.Warn("onFormatMessage task is nil. zone: %v ", zoneid)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rawmsg[4:], &result); err != nil {
		Log.Warn("onFormatMessage Unmarshal err(%v)", err)
		return
	}
	//dumpmap(result)

	if task.Fw != nil {
		rawMap := make(map[string]interface{})
		rawMap["msg"] = rawmsg[4:]
		task.Fw.Write(fmt.Sprintf("zone_%v", zoneid), rawMap)
	}

	if task.Db != nil {
		logtype := binary.LittleEndian.Uint32(rawmsg[0:])
		if logtype&2 != 0 {
			dbjson := utils.DeepCopy(result)
			if valueMap, ok := dbjson.(map[string]interface{}); ok {
				task.Db.Write(fmt.Sprintf("format_%v", zoneid), valueMap)
			} else {
				Log.Warn("onFormatMessage DeepCopy for log (%v)", logtype)
			}
		}
	}

	if task.Se != nil {
		myjson := utils.DeepCopy(result)
		if valueMap, ok := myjson.(map[string]interface{}); ok {
			task.Se.Write(fmt.Sprintf("format_%v", zoneid), valueMap)
		} else {
			Log.Warn("onFormatMessage DeepCopy for log")
		}
	}

	if task.Mgo != nil {
		mgojson := utils.DeepCopy(result)
		if valueMap, ok := mgojson.(map[string]interface{}); ok {
			task.Mgo.Write("format", valueMap)
		} else {
			Log.Warn("onFormatMessage DeepCopy for log")
		}
	}

	if task.Es != nil {
		task.Es.Write("format", result)
	}
}

////////////////////////////////////////////////
var Processor = raw.NewProcessor()

func init() {
	Processor.SetByteOrder(true)

	Processor.Register(10012, onRegister)
	Processor.Register(10010, onRawMessage)
	Processor.Register(10014, onFormatMessage)
}
