package mysql

import (
	"fmt"
	"ghost/models/mysql/decoder"
	"ghost/services/utils"
	"time"

	. "ghost/global"
)

var mapcoder *decoder.Decoder
var Handlers map[string]func(*MysqlEngine, uint32, map[string]interface{})

func init() {

	mapcoder = decoder.NewDecoder()

	Handlers = map[string]func(*MysqlEngine, uint32, map[string]interface{}){
		"accountlogin":      ctor_accountlogin,
		"accountlogout":     ctor_accountlogout,
		"rolelogin":         ctor_rolelogin,
		"rolelogout":        ctor_rolelogout,
		"onlineuser":        ctor_onlinestate,
		"createrole":        ctor_createrole,
		"levelup":           ctor_levelup,
		"endtask":           ctor_endtask,
		"clientstartnovice": ctor_guide,
		"shoptrade":         ctor_shoptrade,
		"addyuanbao":        ctor_moneyproduct,
		"costyuanbao":       ctor_moneyconsume,
		"leaguecreate":      ctor_leaguecreate,
		"leaguedata":        ctor_leaguedata,
		"leagueactivity":    ctor_leagueactivity,
		"charge":            ctor_chargedata,
		"gainitem":          ctor_gainitem,
		"loseitem":          ctor_costitem,
		"starttask":         ctor_starttask,
		"purchase":          ctor_purchase,
		"award":             ctor_award,
	}
}

//////////////////////////////////////////////////////////
func dumpJson(m map[string]interface{}) {
	for key, value := range m {
		fmt.Println("Key:", key, " --> Value:", value)
	}
}

type AccLogin struct {
	AccountId uint32 `json:"gameuserid"`
	UserName  string `json:"userid"`
	ServerId  int    `json:"serverid"`
	Ip        string `json:"loginip"`
	LoginTime string `json:"logtime"`
	Plat      int    `json:"platform"`
	Os        int    `json:"os"`
	Mac       string `json:"deviceid"`
	Model     string `json:"model"`
}

func ctor_accountlogin(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {

	//dumpJson(jsonData)

	var info AccLogin
	if err := mapcoder.Decode(&info, jsonData); err != nil {
		Log.Error("ctor_accountlogin >>>> Decode Error :", err)
		return
	}

	modelstr := utils.SplitModel(info.Model)
	sql := fmt.Sprintf(`INSERT INTO account VALUES ('%d','%d','%d','%s','%s', '%d', '%s') `,
		info.AccountId, info.Plat, info.Os, info.Mac, info.LoginTime, 0, modelstr)

	//Console.Info("%s, %v", modelstr, sql)
	if _, err := self.engine.Exec(sql); err != nil {
		if !utils.CatchError("Duplicate", err) {
			Log.Error("ctor_accountlogin >>>> DBEngine Insert(account) sql Error :(%v)", err)
		}
	}

	var accountlogininfo AccountLogin
	tablename := accountlogininfo.TableName()
	if tm, err := time.Parse("2006-01-02", accountlogininfo.LoginTime); err == nil {
		tablename = gettablename("accountlogin", tm)
	}

	sql = fmt.Sprintf(`replace into %s VALUES ('%d', '%v', '%s','%d','%s','%s','%s','%s') `,
		tablename, info.AccountId, info.Plat, info.UserName, info.ServerId, info.Ip, info.LoginTime, info.Mac, modelstr)

	if _, err := self.engine.Exec(sql); err != nil {
		Log.Error("ctor_accountlogin >>>> DBEngine Insert(%v) sql Error :(%v)", tablename, sql)
	}
}

func statisticsAccountRoleCount(self *MysqlEngine, accountid uint32) {
	sql := fmt.Sprintf(`UPDATE account SET role_count = role_count + 1 where account_id='%v'`, accountid)
	if _, err := self.engine.Exec(sql); err != nil {
		Log.Error("ctor_levelup >>>> DBEngine Insert(account) sql Error :", sql)
	}
}

func ctor_rolelogin(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {
	var info RoleLogin
	if err := mapcoder.Decode(&info, jsonData); err != nil {
		Log.Error("ctor_rolelogin >>>> Decode Error :", err)
		return
	}
	// if info.IsFirst == 1 {
	// 	statisticsAccountRoleCount(self, info.AccountId)
	// }

	tablename := info.TableName()
	if tm, err := time.Parse("2006-01-02", info.LogTime); err == nil {
		tablename = gettablename("rolelogin", tm)
	}

	modelstr := utils.SplitModel(info.Model)
	sql := fmt.Sprintf(`INSERT INTO %s VALUES ('%v', '%v', '%s', '%v', '%v', '%d', '%d', '%d', '%s')`,
		tablename, info.AccountId, info.RoleId, info.LogTime, info.Level, info.Cash, info.ServerId, info.Plat, info.Profession, modelstr)
	sql += fmt.Sprintf(` ON DUPLICATE KEY UPDATE level=%v, cash=%v, log_time='%s' `,
		info.Level, info.Cash, info.LogTime)

	if _, err := self.engine.Exec(sql); err != nil {
		Log.Error("ctor_rolelogin >>>> DBEngine Insert sql Error :", sql)
	}
}

func ctor_rolelogout(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {
	var info RoleLogout
	if err := mapcoder.Decode(&info, jsonData); err != nil {
		Log.Error("ctor_rolelogout >>>> Decode Error :", err)
		return
	}

	tablename := info.TableName()
	if tm, err := time.Parse("2006-01-02", info.LogTime); err == nil {
		tablename = gettablename("rolelogout", tm)
	}

	modelstr := utils.SplitModel(info.Model)
	sql := fmt.Sprintf(`INSERT INTO %s VALUES ('%v', '%v', '%s', '%v', '%d', '%v', '%v', '%04f', '%04f', '%d', '%d','%v', '%d', '%s')`,
		tablename, info.AccountId, info.RoleId, info.LogTime, info.Level, info.Cash, info.Duration, info.MapId, info.Posx, info.Posy, info.TaskId, info.ServerId, info.Plat, info.Profession, modelstr)
	sql += fmt.Sprintf(` ON DUPLICATE KEY UPDATE level=%v, cash=%v, duration=duration+%v, map_id=%v, posx='%04f', posy='%04f', task_id=%d, log_time='%s' `,
		info.Level, info.Cash, info.Duration, info.MapId, info.Posx, info.Posy, info.TaskId, info.LogTime)

	if _, err := self.engine.Exec(sql); err != nil {
		Log.Error("ctor_rolelogout >>>> DBEngine Insert sql Error :", sql)
	}

	// if _, err := self.engine.Table(tablename).Exec(sql); err != nil {
	// 	Log.Error("DBEngine Insert sql Error :", err)
	// }
}

func ctor_onlinestate(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {
	var info OnlineState
	//dumpJson(jsonData)

	if err := mapcoder.Decode(&info, jsonData); err != nil {
		Log.Error("ctor_onlinestate >>>> Decode Error :", err)
		return
	}

	tablename := info.TableName()
	if tm, err := time.Parse("2006-01-02", info.Logtime); err == nil {
		tablename = gettablename("onlinestate", tm)
		Log.Error("ctor_onlinestate >>>> time Parse Error :", err)
	}

	sql := fmt.Sprintf(`INSERT INTO %s VALUES ('%s', '%d', '%d', '%d', '%d') `,
		tablename, info.Logtime, info.OnlineUser, info.LeagueNum, info.ServerId, info.Plat)
	if _, err := self.engine.Exec(sql); err != nil {
		Log.Error("ctor_onlinestate >>>> DBEngine Insert sql Error :", sql)
	}
}

func ctor_createrole(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {
	var info CreateRole
	if err := mapcoder.Decode(&info, jsonData); err != nil {
		Log.Error("ctor_createrole >>>> Decode Error :", err)
		return
	}

	// sql := fmt.Sprintf(`INSERT INTO leveldistribution VALUES ('%d', '%d', '%s', '%d', '%d') ON DUPLICATE KEY UPDATE num=num+1, log_time='%s' `,
	// 	info.Level, 1, info.LogTime, info.ServerId, info.Plat, info.LogTime)
	// if _, err := self.engine.Exec(sql); err != nil {
	// 	Log.Error("ctor_levelup >>>> DBEngine Insert(leveldistribution) sql Error :", err)
	// }

	modelstr := utils.SplitModel(info.Model)
	namestr := utils.SqlStringCheck(info.Name)
	sql := fmt.Sprintf(`INSERT INTO roleinfo VALUES ('%d', '%d', '%d', '%s',  '%s', '%s', '%d', '%d', '%d', '%d', '%s', '%d', '%v', '%v')`,
		info.Account, info.RoleId, info.Plat, info.LogTime, info.Mac, info.LogTime, info.Sex, info.Prof, info.Level, 0, namestr, 0, info.BattleValue, modelstr)
	sql += fmt.Sprintf(` ON DUPLICATE KEY UPDATE level='%d', cash='%d', log_time='%s', name='%s', battle_value='%v'`,
		info.Level, info.Cash, info.LogTime, namestr, info.BattleValue)
	if _, err := self.engine.Exec(sql); err != nil {
		Log.Error("ctor_createrole >>>> DBEngine Insert(roleinfo) sql Error :", sql)
	}
}

func ctor_levelup(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {
	var info LevelUp
	if err := mapcoder.Decode(&info, jsonData); err != nil {
		Log.Error("ctor_levelup >>>> Decode Error :", err)
		return
	}

	// sql := fmt.Sprintf(`INSERT INTO leveldistribution VALUES ('%d', '%d', '%s', '%d', '%d') ON DUPLICATE KEY UPDATE num=num+1, log_time='%s' `,
	// 	info.Level, 1, info.LogTime, info.ServerId, info.Plat, info.LogTime)
	// if _, err := self.engine.Exec(sql); err != nil {
	// 	Log.Error("ctor_levelup >>>> DBEngine Insert(leveldistribution) sql Error :", err)
	// }

	modelstr := utils.SplitModel(info.Model)
	namestr := utils.SqlStringCheck(info.Name)
	sql := fmt.Sprintf(`INSERT INTO roleinfo VALUES ('%d', '%d', '%d', '%s',  '%s', '%s', '%d', '%d', '%d', '%d', '%s', '%d', '%v', '%v')`,
		info.Account, info.RoleId, info.Plat, info.LogTime, info.Mac, info.LogTime, info.Sex, info.Prof, info.Level, info.Cash, namestr, 0, info.BattleValue, modelstr)
	sql += fmt.Sprintf(` ON DUPLICATE KEY UPDATE level='%d', cash='%d', log_time='%s', name='%s', battle_value='%v'`,
		info.Level, info.Cash, info.LogTime, namestr, info.BattleValue)
	if _, err := self.engine.Exec(sql); err != nil {
		Log.Error("ctor_levelup >>>> DBEngine Insert(roleinfo) sql Error :", sql)
	}
}

func ctor_endtask(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {
	var info TaskFinish
	if err := mapcoder.Decode(&info, jsonData); err != nil {
		Log.Error("ctor_endtask >>>> Decode Error :", err)
		return
	}

	tablename := info.TableName()
	if tm, err := time.Parse("2006-01-02", info.LogTime); err == nil {
		tablename = gettablename("taskfinish", tm)
	}
	sql := fmt.Sprintf(`INSERT INTO %s VALUES ('%d', '%d', '%d', '%d', '%s', '%d', '%d') `,
		tablename, info.TaskTypeId, info.TaskId, 1, info.Fight, info.LogTime, info.ServerId, info.Plat)
	sql += fmt.Sprintf(` ON DUPLICATE KEY UPDATE num=num+1, log_time='%s' `, info.LogTime)
	if _, err := self.engine.Exec(sql); err != nil {
		Log.Error("ctor_endtask >>>> DBEngine Insert sql Error :", sql)
	}
}

func ctor_guide(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {
	var info GuideInfo
	if err := mapcoder.Decode(&info, jsonData); err != nil {
		Log.Error("ctor_guide >>>> Decode Error :", err)
		return
	}

	tablename := info.TableName()
	if tm, err := time.Parse("2006-01-02", info.LogTime); err == nil {
		tablename = gettablename("guideinfo", tm)
	}
	sql := fmt.Sprintf(`INSERT INTO %s VALUES ('%d', '%d', '%s', '%d', '%d')`,
		tablename, info.HelpId, 1, info.LogTime, info.ServerId, info.Plat)
	sql += fmt.Sprintf(` ON DUPLICATE KEY UPDATE num=num+1, log_time='%s' `, info.LogTime)
	if _, err := self.engine.Exec(sql); err != nil {
		Log.Error("ctor_guide >>>> DBEngine Insert sql Error :", sql)
	}
}

func ctor_shoptrade(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {
	var info ShopTrade
	if err := mapcoder.Decode(&info, jsonData); err != nil {
		Log.Error("ctor_shoptrade >>>> Decode Error :", err)
		return
	}

	tablename := info.TableName()
	if tm, err := time.Parse("2006-01-02", info.LogTime); err == nil {
		tablename = gettablename("shoptrade", tm)
	}
	sql := fmt.Sprintf(`INSERT INTO %s VALUES ('%d', '%d', '%s', '%d', '%d') ON DUPLICATE KEY UPDATE num=num+1, log_time='%s' `,
		tablename, info.MallId, 1, info.LogTime, info.ServerId, info.Plat, info.LogTime)
	if _, err := self.engine.Exec(sql); err != nil {
		Log.Error("ctor_shoptrade >>>> DBEngine Insert sql Error :", sql)
	}
}

func ctor_moneyproduct(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {
	var info MoneyProduct
	if err := mapcoder.Decode(&info, jsonData); err != nil {
		Log.Error("ctor_moneyproduct >>>> Decode Error :", err)
		return
	}

	if info.MoneyType != 1 && info.MoneyType != 3 && info.MoneyType != 4 {
		return
	}

	tablename := info.TableName()
	if tm, err := time.Parse("2006-01-02", info.LogTime); err == nil {
		tablename = gettablename("moneyproduct", tm)
	}
	sql := fmt.Sprintf(`INSERT INTO %s VALUES ('%d', '%d', '%d', '%v')`, tablename, info.MoneyType, info.Reason, info.Value, info.LogTime)
	if _, err := self.engine.Exec(sql); err != nil {
		Log.Error("ctor_moneyproduct >>>> DBEngine Insert sql Error :", sql)
	}
}

func ctor_moneyconsume(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {
	var info MoneyConsume
	if err := mapcoder.Decode(&info, jsonData); err != nil {
		Log.Error("ctor_moneyconsume >>>> Decode Error :", err)
		return
	}

	if info.MoneyType != 1 && info.MoneyType != 3 && info.MoneyType != 4 {
		return
	}

	tablename := info.TableName()
	if tm, err := time.Parse("2006-01-02", info.LogTime); err == nil {
		tablename = gettablename("moneyconsume", tm)
	}
	sql := fmt.Sprintf(`INSERT INTO %s VALUES ('%d', '%d', '%v', '%v')`, tablename, info.MoneyType, info.Reason, info.Value, info.LogTime)
	if _, err := self.engine.Exec(sql); err != nil {
		Log.Error("ctor_moneyconsume >>>> DBEngine Insert sql Error :", sql)
	}
}

func ctor_leagueactivity(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {
	var info LeagueActivity
	if err := mapcoder.Decode(&info, jsonData); err != nil {
		Log.Error("ctor_leagueactivity >>>> Decode Error :", err)
		return
	}

	tablename := info.TableName()
	sql := fmt.Sprintf(`INSERT INTO %s VALUES ('%d', '%d', '%v', '%d', '%d')`,
		tablename, info.ActivityId, info.DailyNum, info.League, info.ActivePlayerNum, info.MemberNum)
	if _, err := self.engine.Exec(sql); err != nil {
		Log.Error("ctor_leagueactivity >>>> DBEngine Insert sql Error :", sql)
	}
}

func ctor_leaguecreate(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {
	var info LeagueInfo
	if err := mapcoder.Decode(&info, jsonData); err != nil {
		Log.Error("ctor_leaguecreate >>>> Decode Error :", err)
		return
	}

	namestr := utils.SqlStringCheck(info.Name)
	sql := fmt.Sprintf(`INSERT INTO leagueinfo VALUES ('%d', '%s', '0', '%d', '0', '%s', '%v', '%v')`,
		info.League, namestr, info.Camp, info.Leader, info.LeaderId, info.Brithday)
	if _, err := self.engine.Exec(sql); err != nil {
		Log.Error("ctor_leaguecreate >>>> DBEngine Insert sql Error :", sql)
	}
}

func ctor_leaguedata(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {
	var info LeagueData
	if err := mapcoder.Decode(&info, jsonData); err != nil {
		Log.Error("ctor_leaguedata >>>> Decode Error :", err)
		return
	}

	sql := fmt.Sprintf(`UPDATE leagueinfo SET state='%d', level='%d', leader='%s' where league='%d'`,
		info.State, info.Level, info.Leader, info.League)
	// sql := fmt.Sprintf(`replace leagueinfo ('league', 'name', 'level', 'state', 'leader') values ('%d', '%s', '%d', '%d', '%s')`,
	// 	info.League, info.Name, info.Level, info.State, info.Leader)
	if _, err := self.engine.Exec(sql); err != nil {
		Log.Error("ctor_leaguedata >>>> DBEngine replace sql Error :", err)
	}

	tablename := info.TableName()
	if tm, err := time.Parse("2006-01-02", info.Logtime); err == nil {
		tablename = gettablename("leaguedata", tm)
	}
	sql = fmt.Sprintf(`INSERT INTO %s VALUES ('%d', '%d', '%d', '%v', '%d', '%d', '%d', '%d', '%v')`,
		tablename, info.League, info.State, info.Level, info.Assets, info.Credits, info.Health, info.Member, info.Student, info.Logtime)
	if _, err := self.engine.Exec(sql); err != nil {
		Log.Error("ctor_leaguedata >>>> DBEngine Insert sql Error :", sql)
	}
}

func ctor_chargedata(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {
	var info ChargeInfo
	if err := mapcoder.Decode(&info, jsonData); err != nil {
		Log.Error("ctor_chargedata >>>> Decode Error :", err)
		return
	}

	// tablename := info.TableName()
	// if tm, err := time.Parse("2006-01-02", info.LogTime); err == nil {
	// 	tablename = gettablename("chargeinfo", tm)
	// }
	namestr := utils.SqlStringCheck(info.Name)
	sql := fmt.Sprintf(`INSERT INTO chargeinfo VALUES ('%s', '%d', '%d', '%d', '%v', '%v', '%s', '%s', '%d', '%v', '%v', '%d', '%d', '%s', '%d', '%s')`,
		info.LogTime, info.Plat, info.Os, info.ServerId, info.Account, info.RoleId, namestr, info.RegTime, info.Level,
		info.TotalCash, info.Cash, info.RechargeId, info.RechargeType, info.Order, info.IsFirst, info.PlatOrder)
	if _, err := self.engine.Exec(sql); err != nil {
		Log.Error("ctor_chargedata >>>> DBEngine Insert sql Error :", sql)
	}
}

func ctor_gainitem(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {

}

func ctor_costitem(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {

}

func ctor_starttask(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {

}

func ctor_accountlogout(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {

}

func ctor_purchase(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {

}

func ctor_award(self *MysqlEngine, zoneId uint32, jsonData map[string]interface{}) {

}
