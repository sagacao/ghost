package report

import (
	"encoding/json"
	"ghost/network"
	"ghost/services/utils"
	"sync"

	. "ghost/global"
)

const defaultAsyncMsgLen = 1e3

var logMsgPool *sync.Pool

type ReportEngine struct {
	lock sync.Mutex
	url  string `json:"url"`

	signalChan chan string
	wg         sync.WaitGroup
	client     *network.HttpClient
}

func NewReportEngine() *ReportEngine {
	re := new(ReportEngine)
	err := re.Init()
	if err != nil {
		panic(err)
	}

	re.signalChan = make(chan string, 1)

	re.client = new(network.HttpClient)
	re.client.PendingWriteNum = 1000

	if re.client != nil {
		re.client.Start()
	}

	re.wg.Add(1)
	return re
}

func (re *ReportEngine) Init() error {
	//el.lock.Lock()
	//defer el.lock.Unlock()
	re.url = Config.Report.URL
	return nil
}

func (re *ReportEngine) Write(database string, jsonmap map[string]interface{}) {
	logname, ok := jsonmap["logname"]
	if !ok || re.client == nil {
		return
	}

	if logname == "createrole" || logname == "levelup" {
		//var jsondata reportData
		jsondata := &reportData{
			User:    utils.ToUint32(utils.GetInterfaceString("roleid", jsonmap)),
			Account: utils.GetInterfaceString("userid", jsonmap),
			Server:  utils.ToUint32(utils.GetInterfaceString("serverid", jsonmap)),
			Level:   utils.ToUint32(utils.GetInterfaceString("lev", jsonmap)),
			Prof:    utils.ToUint32(utils.GetInterfaceString("prof", jsonmap)),
			Sex:     utils.ToUint32(utils.GetInterfaceString("sex", jsonmap)),
		}
		text, err := json.Marshal(jsondata)
		if err != nil {
			Log.Warn("ReportEngine Write json Marshal err:(%v)", err)
			return
		}
		re.client.Write("POST", re.url, text)
	}
}

func (re *ReportEngine) WriteRaw(database string, json map[string]interface{}) {

}

func (re *ReportEngine) Destroy() {
	re.signalChan <- "close"
	re.wg.Wait()
	close(re.signalChan)
}

func (re *ReportEngine) Flush() {
	re.signalChan <- "flush"
	re.wg.Wait()
	re.wg.Add(1)
}
