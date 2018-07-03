package mysql

import (
	"encoding/json"
	"fmt"
	. "ghost/global"
	"os/exec"
	"strings"
	"time"
)

func createDatabase(dbname string) error {
	cmd := fmt.Sprintf("%s/bin/mysql", Config.Mysql.Path)
	host_str := strings.Split(Config.Mysql.Host, ":")
	if len(host_str) != 2 {
		return fmt.Errorf("Error Config Of Mysql Host `%s` like this '127.0.0.1:3306'", Config.Mysql.Host)
	}
	optstr := fmt.Sprintf("-h%s -P%s -u%s -p%s -e", host_str[0], host_str[1], Config.Mysql.User, Config.Mysql.Password)

	opt := strings.Split(optstr, " ")
	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbname)

	_, err := exec.Command(cmd, opt[0], opt[1], opt[2], opt[3], opt[4], sql).Output()
	if err != nil {
		err = fmt.Errorf("Create database `%s`, %s", dbname, err.Error())
		return err
	}
	return nil
}

func gettablename(table string, tm time.Time) string {
	return fmt.Sprintf("log_%s_%s", tm.Format("20060102"), table)
}

func getPrefix(tm time.Time) string {
	return fmt.Sprintf("log_%s", tm.Format("20060102"))
}

func map2Struct(m map[string]interface{}, v interface{}) error {
	outrsp, err := json.Marshal(m)
	if err != nil {
		return err
	}

	//fmt.Printf("json: %v ", string(outrsp))

	err = json.Unmarshal(outrsp, &v)
	if err != nil {
		return err
	}
	return nil
}
