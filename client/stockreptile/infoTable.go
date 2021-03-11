package main

import (
    "time"
    "fmt"
	"minirpc/Model"
)


type GgdayinfoModel struct {
	Id      int    `xorm:"INT(10)"`
	Name string `xorm:"VARCHAR(255)"`
	Code string `xorm:"VARCHAR(255)"`
	Val string `xorm:"VARCHAR(255)"`
	Per string `xorm:"VARCHAR(255)"`
	Daystr string `xorm:"VARCHAR(255)"`
	Ptype string `xorm:"VARCHAR(255)"`
	Ctime  string    `xorm:"DATETIME"`
}

func AddGgDayInfo(info StackInfo, Daystr string) bool {
	engine := Model.GetInstance()

	detail := new(GgdayinfoModel)
	detail.Code = info.Code
	detail.Name = info.Name
	detail.Val = info.Val
	detail.Per = info.Per
	detail.Daystr = Daystr
	detail.Ptype = info.Ptype
	detail.Code = info.Code
	detail.Ctime = time.Now().Format("2006-01-02 15:04:05")
	_, err := engine.Table("go_ggdaydetail").Insert(detail)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true

}