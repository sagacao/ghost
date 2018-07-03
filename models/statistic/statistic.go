package statistic

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "ghost/global"
	"ghost/services/utils"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
)

type ChargeParam struct {
	FuncNo         string `json:"funcNo"`
	Version        string `json:"version"`
	SysType        string `json:"sysType"`
	ChannelID      string `json:"channelID"`
	GameID         string `json:"gameID"`
	UId            string `json:"uId"`
	Level          string `json:"level"`
	Orderno        string `json:"orderno"`
	GameServer     string `json:"gameServer"`
	Itemid         string `json:"itemid"`
	CurrencyAmount string `json:"currencyAmount"`
	VcAmount       string `json:"vcAmount"`
	CurrencyType   string `json:"currencyType"`
}

type ChargeBody struct {
	Param   ChargeParam `json:"param"`
	Sign    string      `json:"sign"`
	GameKey string      `json:"gameKey"`
}

type ReqHandler struct {
	Info map[string]interface{}
	Type int
}

type StatisticEngine struct {
	in         chan *ReqHandler
	msgChanLen int64
	signalChan chan string
	wg         sync.WaitGroup
}

var statisticClient = &http.Client{}

func NewStatisticEngine() *StatisticEngine {
	se := StatisticEngine{}
	se.Init()

	se.in = make(chan *ReqHandler, se.msgChanLen)
	se.signalChan = make(chan string, 1)

	se.wg.Add(1)
	go se.start()

	return &se
}

func (se *StatisticEngine) start() {
	gameOver := false
	for {
		select {
		case myhandler, ok := <-se.in:
			if !ok {
				return
			}

			se.SendStatistic(myhandler)

		case sg := <-se.signalChan:
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

func (se *StatisticEngine) Init() error {
	se.msgChanLen = Config.Output.ChanLen
	if se.msgChanLen <= 0 {
		se.msgChanLen = 1000
	}

	return nil
}

func (se *StatisticEngine) Write(database string, json map[string]interface{}) {
	if fmt.Sprintf("%v", json["platform"]) != "12" {
		return
	}

	handle := &ReqHandler{Info: json}
	logname := json["logname"].(string)
	if logname == "charge" {
		handle.Type = 6
		se.in <- handle
	}
}

func (se *StatisticEngine) WriteRaw(database string, json map[string]interface{}) {

}

func (se *StatisticEngine) Destroy() {
	se.signalChan <- "close"
	se.wg.Wait()
	close(se.in)
	close(se.signalChan)
}

func (se *StatisticEngine) Flush() {
	se.signalChan <- "flush"
	se.wg.Wait()
	se.wg.Add(1)
}

func (se *StatisticEngine) SendStatistic(handle *ReqHandler) {
	url := GetStatisticUrl(handle.Type)
	param := &ChargeParam{
		FuncNo:         "1",
		Version:        "4",
		SysType:        "0",
		ChannelID:      "92504f729bda4862",
		GameID:         "8972e8dbf1cdaf74",
		UId:            handle.Info["userid"].(string),
		Level:          handle.Info["lev"].(string),
		Orderno:        handle.Info["gameorder"].(string),
		GameServer:     handle.Info["serverid"].(string),
		Itemid:         handle.Info["rechargeid"].(string),
		CurrencyAmount: handle.Info["price"].(string),
		VcAmount:       handle.Info["totalcash"].(string),
		CurrencyType:   "CNY",
	}
	b, err := json.Marshal(param)
	if err != nil {
		return
	}

	var buffer bytes.Buffer
	buffer.WriteString(url)
	buffer.WriteString("?param=")
	buffer.WriteString(string(b))
	buffer.WriteString("&sign=")
	buffer.WriteString(GetChuJianSignStr(param))
	//buffer.WriteString("&gameKey=542065188e4849f9ab82752c48cbcc8d")
	//Console.Info(buffer.String())
	reqest, err := http.NewRequest("GET", buffer.String(), nil)
	if err != nil {
		Console.Error("%v", err)
	}
	se.AsynRequest(reqest)
}

func (se *StatisticEngine) AsynRequest(req *http.Request) {
	go func() {
		resp, err := statisticClient.Do(req)
		if err != nil {
			Console.Error("send statistic err:[%v]", err)
		} else {
			defer resp.Body.Close()
			Console.Info("%v", resp)
			res, _ := ioutil.ReadAll(resp.Body)
			Console.Info(string(res))
		}
	}()
}

func GetJsonStatistic(handle *ReqHandler) string {
	if handle.Type == 6 {
		param := &ChargeParam{
			FuncNo:         "1",
			Version:        "4",
			SysType:        "0",
			ChannelID:      "92504f729bda4862",
			GameID:         "92504f729bda48628972e8dbf1cdaf74",
			UId:            handle.Info["userid"].(string),
			Level:          handle.Info["lev"].(string),
			Orderno:        handle.Info["gameorder"].(string),
			GameServer:     handle.Info["serverid"].(string),
			Itemid:         handle.Info["rechargeid"].(string),
			CurrencyAmount: handle.Info["price"].(string),
			VcAmount:       handle.Info["totalcash"].(string),
			CurrencyType:   "CNY",
		}
		b, err := json.Marshal(param)
		if err != nil {
			return ""
		}
		Console.Info("%v", string(b))
		Console.Info("%v", GetChuJianSignStr(param))
		var clusterinfo = url.Values{}
		clusterinfo.Add("param", string(b))
		clusterinfo.Add("sign", GetChuJianSignStr(param))
		clusterinfo.Add("gameKey", "542065188e4849f9ab82752c48cbcc8d")
		data := clusterinfo.Encode()

		return data
	}
	return ""
}

func GetStatisticUrl(stype int) string {
	if stype == 6 {
		return "http://yh.ga.16801.com/inter/v3/rechargeLog.page"
	}
	return ""
}

func GetChuJianSignStr(param *ChargeParam) string {
	sign := fmt.Sprintf("channelID=%v&currencyAmount=%v&currencyType=%v&funcNo=%v&gameID=%v&gameServer=%v&itemid=%v&level=%v&orderno=%v&sysType=%v&uId=%v&vcAmount=%v&version=%v&gameKey=%v",
		param.ChannelID,
		param.CurrencyAmount,
		param.CurrencyType,
		param.FuncNo,
		param.GameID,
		param.GameServer,
		param.Itemid,
		param.Level,
		param.Orderno,
		param.SysType,
		param.UId,
		param.VcAmount,
		param.Version,
		"542065188e4849f9ab82752c48cbcc8d")
	Console.Info(sign)
	return utils.Md5sum([]byte(sign))
}
