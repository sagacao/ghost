package mysql

import (
	"fmt"
	"time"
)

//货币产出
type LogFilter struct {
	AccountId uint32 `xorm:"pk notnull index"`
}

//全区全服
//新建账号，导入量，只有一个表
type Account struct {
	AccountId   uint32 `json:"gameuserid" xorm:"pk notnull index"`
	Plat        int    `json:"platform" xorm:"notnull index"`
	Os          int    `json:"os" xorm:"notnull"`
	Mac         string `json:"deviceid" xorm:"varchar(64) index"`
	CreatedTime string `json:"logtime" xorm:"TIMESTAMP index"`
	RoleCount   int    `json:"-" xorm:"notnull"`
	Model       string `json:"model" xorm:"varchar(256) index"`
}

func (self *Account) TableName() string {
	return fmt.Sprintf("account")
}

//账号登录，记录最后一次登录时间和服务器
type AccountLogin struct {
	AccountId uint32 `json:"gameuserid" xorm:"pk notnull index"`
	Plat      int    `json:"platform" xorm:"notnull index"`
	UserName  string `json:"userid" xorm:"varchar(256)"`
	ServerId  int    `json:"serverid" xorm:"notnull"`
	Ip        string `json:"loginip" xorm:"varchar(20)"`
	LoginTime string `json:"logtime" xorm:"TIMESTAMP"`
	Mac       string `json:"deviceid" xorm:"varchar(64) index"`
	Model     string `json:"model" xorm:"varchar(256) index"`
}

func (self *AccountLogin) TableName() string {
	return fmt.Sprintf("%s_accountlogin", getPrefix(time.Now()))
}

//账号退出，记录最后一次退出时间和服务器
type AccountLogout struct {
	AccountId  uint32 `json:"gameuserid" xorm:"pk notnull index"`
	Plat       int    `json:"platform" xorm:"notnull index"`
	LogoutTime string `json:"logtime" xorm:"TIMESTAMP"`
	ServerId   int    `json:"serverid" xorm:"notnull"`
}

func (self *AccountLogout) TableName() string {
	return fmt.Sprintf("%s_accountlogout", getPrefix(time.Now()))
}

//角色退出，记录退出状态
type RoleLogout struct {
	AccountId  uint32  `json:"gameuserid" xorm:"pk notnull index"`
	RoleId     uint32  `json:"roleid" xorm:"pk notnull index"`
	LogTime    string  `json:"runtime" xorm:"TIMESTAMP"`
	Level      uint32  `json:"lev" xorm:"index"`
	Cash       float32 `json:"totalcash" xorm:"notnull"`
	Duration   uint32  `json:"time" xorm:"notnull"`
	MapId      uint32  `json:"mapid" xorm:"index"`
	Posx       float32 `json:"posx"`
	Posy       float32 `json:"posy"`
	TaskId     int     `json:"taskid" xorm:"index"`
	ServerId   int     `json:"serverid" xorm:"notnull"`
	Plat       int     `json:"platform" xorm:"notnull index"`
	Profession int     `json:"prof" xorm:"notnull index"`
	Model      string  `json:"model" xorm:"varchar(256) index"`
}

func (self *RoleLogout) TableName() string {
	return fmt.Sprintf("%s_rolelogout", getPrefix(time.Now()))
}

type CreateRole struct {
	LogTime     string `json:"logtime"`
	Plat        int    `json:"platform"`
	ServerId    int    `json:"serverid"`
	Mac         string `json:"deviceid"`
	Account     uint32 `json:"gameuserid"`
	RoleId      uint32 `json:"roleid"`
	Level       int    `json:"lev"`
	Sex         int    `json:"sex"`
	Prof        int    `json:"prof"`
	Name        string `json:"account"`
	Cash        uint64 `json:"-"`
	BattleValue uint32 `json:"-" `
	Model       string `json:"model"`
}

type LevelUp struct {
	LogTime     string `json:"logtime"`
	Plat        int    `json:"platform"`
	ServerId    int    `json:"serverid"`
	Mac         string `json:"deviceid"`
	Account     uint32 `json:"gameuserid"`
	RoleId      uint32 `json:"roleid"`
	Level       int    `json:"lev"`
	Sex         int    `json:"sex"`
	Prof        int    `json:"prof"`
	Name        string `json:"account"`
	Cash        uint64 `json:"money"`
	BattleValue uint32 `json:"-" `
	Model       string `json:"model"`
}

//实时在线状态
type OnlineState struct {
	Logtime    string `json:"logtime" xorm:"TIMESTAMP index"`
	OnlineUser uint32 `json:"currentuser"`
	LeagueNum  uint32 `json:"leaguenum"`
	ServerId   int    `json:"serverid" xorm:"notnull"`
	Plat       int    `json:"platform" xorm:"notnull"`
}

func (self *OnlineState) TableName() string {
	return fmt.Sprintf("%s_onlinestate", getPrefix(time.Now()))
}

//等级分布
type LevelDistribution struct {
	Level    int    `json:"lev" xorm:"unique notnull index"`
	Num      int    `json:"-"`
	LogTime  string `json:"logtime" xorm:"TIMESTAMP"`
	ServerId int    `json:"serverid"`
	Plat     int    `json:"platform" xorm:"notnull"`
}

func (self *LevelDistribution) TableName() string {
	return "leveldistribution"
}

type RoleInfo struct {
	AccountId   uint32 `json:"gameuserid" xorm:"pk notnull index"`
	RoleId      uint32 `json:"roleid" xorm:"pk notnull index"`
	Plat        int    `json:"platform" xorm:"notnull index"`
	CreatedTime string `json:"logtime" xorm:"TIMESTAMP index"`
	Mac         string `json:"deviceid" xorm:"varchar(64)"`
	LogTime     string `json:"logtime" xorm:"TIMESTAMP index"`
	Sex         int    `json:"sex" `
	Profession  int    `json:"prof" `
	Level       uint32 `json:"lev" xorm:"index"`
	Cash        uint64 `json:"totalcash" `
	Name        string `json:"account" xorm:"varchar(64)"`
	GuideId     uint32 `json:"-" `
	BattleValue uint32 `json:"-" `
	Model       string `json:"model" xorm:"varchar(256) index"`
}

func (self *RoleInfo) TableName() string {
	return "roleinfo"
	//return fmt.Sprintf("%s_createrole", getPrefix(time.Now()))
}

//等级分布
type TaskFinish struct {
	TaskTypeId int    `json:"tasktypeid" xorm:"pk notnull index"`
	TaskId     int    `json:"taskid" xorm:"pk notnull index"`
	Num        int    `json:"-"`
	Fight      int    `json:"fight"`
	LogTime    string `json:"logtime" xorm:"TIMESTAMP"`
	ServerId   int    `json:"serverid"`
	Plat       int    `json:"platform" xorm:"notnull"`
}

func (self *TaskFinish) TableName() string {
	//return "taskfinish"
	return fmt.Sprintf("%s_taskfinish", getPrefix(time.Now()))
}

type GuideInfo struct {
	HelpId   int    `json:"helpid" xorm:"pk notnull index"`
	Num      int    `json:"-"`
	LogTime  string `json:"logtime" xorm:"TIMESTAMP"`
	ServerId int    `json:"serverid"`
	Plat     int    `json:"platform" xorm:"notnull"`
}

func (self *GuideInfo) TableName() string {
	//return "guideinfo"
	return fmt.Sprintf("%s_guideinfo", getPrefix(time.Now()))
}

type RoleLogin struct {
	AccountId  uint32 `json:"gameuserid" xorm:"pk index"`
	RoleId     uint32 `json:"roleid" xorm:"pk index"`
	LogTime    string `json:"logtime" xorm:"varchar(32)"`
	Level      int    `json:"lev" xorm:"index"`
	Cash       uint64 `json:"totalcash"`
	ServerId   int    `json:"serverid"`
	Plat       int    `json:"platform" xorm:"index"`
	IsFirst    int    `json:"isfirst" xorm:"-"`
	Profession int    `json:"prof" xorm:"notnull index"`
	Model      string `json:"model" xorm:"varchar(256) index"`
}

func (self *RoleLogin) TableName() string {
	return fmt.Sprintf("%s_rolelogin", getPrefix(time.Now()))
}

//商城
type ShopTrade struct {
	MallId   int    `json:"mallid" xorm:"index"`
	Num      int    `json:"itemcount"`
	LogTime  string `json:"logtime" xorm:"TIMESTAMP"`
	ServerId int    `json:"serverid"`
	Plat     int    `json:"platform" xorm:"notnull"`
}

func (self *ShopTrade) TableName() string {
	return fmt.Sprintf("%s_shoptrade", getPrefix(time.Now()))
}

//货币产出
type MoneyProduct struct {
	MoneyType int    `json:"type" xorm:"index"`
	Reason    int    `json:"reason" xorm:"index"`
	Value     uint64 `json:"changecount" xorm:"notnull"`
	LogTime   string `json:"logtime" xorm:"TIMESTAMP"`
}

func (self *MoneyProduct) TableName() string {
	return fmt.Sprintf("%s_moneyproduct", getPrefix(time.Now()))
}

//货币消耗
type MoneyConsume struct {
	MoneyType int    `json:"type" xorm:"index"`
	Reason    int    `json:"reason" xorm:"index"`
	Value     uint64 `json:"changecount" xorm:"notnull"`
	LogTime   string `json:"logtime" xorm:"TIMESTAMP"`
}

func (self *MoneyConsume) TableName() string {
	return fmt.Sprintf("%s_moneyconsume", getPrefix(time.Now()))
}

//工会活动
type LeagueActivity struct {
	ActivityId      int    `json:"id" xorm:"index"`
	DailyNum        int    `json:"dailyplaynum" xorm:"index"`
	League          uint32 `json:"leagueid" xorm:"notnull"`
	ActivePlayerNum int    `json:"activenum"`
	MemberNum       int    `json:"totalnum"`
}

func (self *LeagueActivity) TableName() string {
	return fmt.Sprintf("%s_leagueactivity", getPrefix(time.Now()))
}

type LeagueInfo struct {
	League   int    `json:"roleid" xorm:"notnull index"`
	Name     string `json:"account" xorm:"varchar(64)"`
	Level    int    `json:"-"` //userid
	Camp     int    `json:"leaguecamp"`
	State    int    `json:"-"`
	Leader   string `json:"leadername" xorm:"varchar(64)"`
	LeaderId uint32 `json:"gameuserid"`
	Brithday string `json:"brithday" xorm:"TIMESTAMP"`
}

func (self *LeagueInfo) TableName() string {
	return "leagueinfo"
}

type LeagueData struct {
	League  int    `json:"roleid" xorm:"notnull index"`
	Name    string `json:"account" xorm:"-"`
	State   int    `json:"state"`
	Level   int    `json:"userid"`
	Assets  uint64 `json:"assets"`
	Credits int    `json:"credits"`
	Health  int    `json:"health"`
	Member  int    `json:"member"`
	Student int    `json:"student"`
	Logtime string `json:"logtime" xorm:"TIMESTAMP"`
	Leader  string `json:"leadername" xorm:"-"`
}

func (self *LeagueData) TableName() string {
	return fmt.Sprintf("%s_leaguedata", getPrefix(time.Now()))
}

type ChargeInfo struct {
	LogTime      string  `json:"createtime" xorm:"TIMESTAMP index"` //createtime
	Plat         int     `json:"platform"`
	Os           int     `json:"os" xorm:"notnull"`
	ServerId     int     `json:"serverid"`
	Account      uint32  `json:"gameuserid" xorm:"notnull"`
	RoleId       uint32  `json:"roleid" xorm:"notnull"`
	Name         string  `json:"account" xorm:"varchar(64)"`
	RegTime      string  `json:"regtime" xorm:"TIMESTAMP"` //regtime
	Level        int     `json:"lev" xorm:"notnull"`
	TotalCash    float32 `json:"totalcash" xorm:"notnull"`
	Cash         float32 `json:"price" xorm:"notnull"`
	RechargeId   uint32  `json:"rechargeid"`
	RechargeType int     `json:"chargetype"`
	Order        string  `json:"gameorder" xorm:"varchar(64)"`
	IsFirst      uint32  `json:"isfirst"`
	PlatOrder    string  `json:"platformorder" xorm:"varchar(64)"`
}

func (self *ChargeInfo) TableName() string {
	return "chargeinfo"
}
