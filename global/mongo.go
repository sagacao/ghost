package global

import (
	"encoding/json"
	"io/ioutil"
)

var indexconfig map[string]interface{}

func initMongoIndex() {
	rawdata, err := ioutil.ReadFile("conf/mongoindex.json")
	if err != nil {
		panic(err)
	}
	indexconfig = make(map[string]interface{})
	err = json.Unmarshal(rawdata, &indexconfig)
	if err != nil {
		panic(err)
	}
}
