package config

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/zarte/comutil/Zloger"
)

var Gconfig = new(Sysconfig)
type Sysconfig struct {
	Debug bool
	Dbdns string
	DbPrefix string
	Emaillist string
	CurExePath string
	Alarmper float64
	GLoger *Zloger.Loger
}
func SetConfig() {
	if false {
		Gconfig.CurExePath = "./"
	}else{
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			fmt.Println(err)
		}
		Gconfig.CurExePath = dir+"/"
	}

	Gconfig.GLoger = Zloger.NewLog(Gconfig.CurExePath +"logs")

}