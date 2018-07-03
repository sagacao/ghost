package global

import (
	"github.com/astaxie/beego/logs"
)

var (
	Console *logs.BeeLogger
	Log     *logs.BeeLogger
)

func init() {
	initJson()
	initMongoIndex()

	Console = logs.NewLogger()
	Console.SetLogger(logs.AdapterConsole, `{"level":7}`)

	Log = logs.NewLogger()
	Log.SetLogger(logs.AdapterFile, Config.Log.Cfg)

	Console.Info("Global init Success ")
}
