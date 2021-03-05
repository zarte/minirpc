package Model

import (
    "time"
    "strconv"
	"minirpc/config"
)

type Luacheck struct {
    Id      int    `xorm:"INT(10)"`
    GameId    int `xorm:"INT(11)`
    ServerId    int `xorm:"INT(11)`
    LogStr string `xorm:"VARCHAR(512)"`
	State    int `xorm:"TINYINT(4)`
	AddTime  string    `xorm:"DATETIME"`
	LogTime  string    `xorm:"DATETIME"`
}


func AddLuaCheckLog(list []Luacheck,logtype string ,serverId string,gameid string) bool {
    engine := GetInstance()

    log := new(Luacheck)
    curtime := time.Now().Format("2006-01-02 15:04:05")
    for _,item := range list{
			log.GameId, _ = strconv.Atoi(gameid)
			log.ServerId, _ = strconv.Atoi(serverId)
			log.State = 0
			log.AddTime = curtime
			log.LogTime = item.LogTime

			_, err := engine.Table("test").Insert(log)
			if err != nil {
				config.Gconfig.GLoger.ErrorLog(err.Error())
			}
    }
    return true
}


type Ggdaydetail struct {
	Id      int    `xorm:"INT(10)"`
	Code   string `xorm:"VARCHAR(512)"`
	Name   string `xorm:"VARCHAR(512)"`
	Per   float64 `xorm:"decimal(20,2)"`
	Val   float64 `xorm:"decimal(20,2)"`
	Ctime  string    `xorm:"DATETIME"`
	Daystr  string    `xorm:"VARCHAR(32)"`
	Ptype  int    `xorm:"INT(10)"`
}


func  GgdayList(daystr string) ([]Ggdaydetail, error) {
	engine := GetInstance()
	list := make([]Ggdaydetail, 0)
	err := engine.Table("go_ggdaydetail").Where("daystr=?", daystr).Find(&list)
	if err != nil {
		return nil, err
	}
	return list, err
}

