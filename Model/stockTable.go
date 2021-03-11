package Model

import (
	"fmt"
	"time"
)

type StockInfo struct {
    Id      int    `xorm:"INT(10)"`
	Code   string `xorm:"VARCHAR(512)"`
	Name   string
	Fname   string `xorm:"VARCHAR(512)"`
	Gcode   string `xorm:"VARCHAR(512)"`
	Type  string   `xorm:"VARCHAR(512)"`
	Ctime  string    `xorm:"DATETIME"`
	Stime  string    `xorm:"DATETIME"`
}


func GetOneStockInfo(gcode string) (StockInfo,error) {
	engine := GetInstance()
	info := make([]StockInfo, 0)
	var err error
	if gcode!= ""{
		err = engine.Table("go_stock").Where("gcode=?", gcode).Limit(1).Find(&info)
	}
	if err != nil {
		fmt.Println("GetOneStockInfo fail ")
		return StockInfo{}, err
	}
	if len(info)>0{
		return info[0], nil
	}else{
		return StockInfo{},nil
	}

}


func  AddOneStockInfo(name string,code string,ptype string,gcode string,fname string)  bool {
	engine := GetInstance()
	log := new(StockInfo)
	curtime := time.Now().Format("2006-01-02 15:04:05")
	log.Code = code
	log.Name = name
	log.Fname = fname
	log.Type = ptype
	log.Gcode = gcode
	log.Stime = curtime
	log.Ctime = curtime
	_, err := engine.Table("go_stock").Insert(log)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

