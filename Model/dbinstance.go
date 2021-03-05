package Model

import (
    _ "github.com/go-sql-driver/mysql"
    "xorm.io/xorm"
    "time"
    "minirpc/config"
    "fmt"
    "xorm.io/core"
)



var engine *xorm.Engine

func GetInstance() *xorm.Engine {
    if engine == nil {
        var err error
        engine, err = xorm.NewEngine("mysql", config.Gconfig.Dbdns)
        if err != nil {
            fmt.Println("NewEngine create fail:",err)
        }
        err2:=engine.Ping()
        if err2 != nil {
            fmt.Println("Ping error:",err2)
        }
        mapper := core.NewPrefixMapper(core.SnakeMapper{}, config.Gconfig.DbPrefix)
        engine.SetTableMapper(mapper)
        engine.SetMaxIdleConns(2)
        engine.SetMaxOpenConns(5)
        go func() {
            for {
                time.Sleep(time.Second * 240)
                engine.Ping()
            }
        }()
    }
    return engine
}
