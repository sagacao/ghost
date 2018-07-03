package global

import (
	"encoding/json"
	"io/ioutil"
)

var Config struct {
	Optional struct {
		Elastic bool
		Mongo   bool
		Mysql   bool
		BDC     bool
	}

	Mysql struct {
		Host     string
		User     string `json:"user"`
		Password string
		Db       string
		Path     string `json:"path"`
	}

	Mongo struct {
		Worker  int
		ChanLen int64
		URL     string
		// Filter  []string `json:"filter"`
	}

	Tcp struct {
		Addr string
	}

	Report struct {
		URL string
	}

	Output struct {
		Worker  int
		ChanLen int64
		Host    string
		Port    string
	}

	Log struct {
		Cfg     string
		Console bool
	}

	File struct {
		FilePath string `json:"filename"`
		MaxLines int    `json:"maxlines"`
		MaxSize  int    `json:"maxsize"`
	}

	Bdc struct {
		Gameid    string
		Versionid string
		Host      string
		Db        string
		Password  string
		User      string
		Path      string
		Region    string

		Channelid struct {
			Plat         string
			Ios          string
			Android      string
			Coupon       string
			Chargecoupon string
		}
	}

	Hdc struct {
		Gameid   string
		Appkey   string
		Url      string
		Currency string
		Output   string
	}
}

func initJson() {
	data, err := ioutil.ReadFile("conf/elastic.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &Config)
	if err != nil {
		panic(err)
	}
}
