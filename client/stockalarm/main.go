package main

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/zarte/comutil/Goconfig"
	"minirpc/client/stockalarm/utils"
	"minirpc/rpc/service"
	"os"
	"strconv"
	"minirpc/config"
	"minirpc/Model"
	"strings"
	"time"
)
type strockInfo struct{
	Code string
	Name string
	Msg string
}
func main()  {
	//查询数据库
	config.SetConfig()
	loadIni(config.Gconfig.CurExePath+"config.ini")
	nowdaylist,err := Model.GgdayList(time.Now().Format("20060102"))
	if err != nil {
		fmt.Println(err)
		return
	}
	nowweek := int(time.Now().Weekday())
	var ydaylist []Model.Ggdaydetail
	if nowweek== 1 {
		ydaylist,err = Model.GgdayList(time.Now().AddDate(0,0,-3).Format("20060102"))
	}else{
		ydaylist,err = Model.GgdayList(time.Now().AddDate(0,0,-1).Format("20060102"))
	}

	if err != nil {
		fmt.Println(err)
		return
	}
	runtimestart := time.Now()
	var newCode []strockInfo
	var insCode []strockInfo
	var desCode []strockInfo
	for _,item := range nowdaylist{
		newflag :=true
		for _,yitem := range ydaylist{
			if item.Code == yitem.Code {
				newflag = false
				val :=item.Per-yitem.Per
				if val>=config.Gconfig.Alarmper {
					//获取最近情况
					var otherstr string
					stockinfo := utils.GetStockBaseInfo(item.Code,item.Name)
					if stockinfo.Name != ""{
						//抓取
						otherstr = utils.StockAnaly(stockinfo.Code)
					}
					insCode = append(insCode,strockInfo{
						Code :item.Code,
						Name :item.Name,
						Msg :strconv.FormatFloat(item.Per,'f',2,64) +"% +" +strconv.FormatFloat(val,'f',2,64)+"% "+ otherstr,
					})
				} else if val<=-config.Gconfig.Alarmper {
					var otherstr string
					stockinfo := utils.GetStockBaseInfo(item.Code,item.Name)
					if stockinfo.Name != ""{
						//抓取
						otherstr = utils.StockAnaly(stockinfo.Code)
					}
					desCode = append(desCode,strockInfo{
						Code :item.Code,
						Name :item.Name,
						Msg :strconv.FormatFloat(item.Per,'f',2,64) +"% " +strconv.FormatFloat(val,'f',2,64)+"% "+ otherstr,
					})
				}
				break
			}
		}
		//不存在
		if newflag {
			newCode = append(newCode,strockInfo{
				Code :item.Code,
				Name :item.Name,
			})
		}
	}

	if len(insCode)>100{
		insCode = []strockInfo{
			strockInfo{
				Code :"111",
				Name :"sss",
				Msg :"超过100",
			},
		}
	}
	if len(desCode)>100{
		desCode = []strockInfo{
			strockInfo{
				Code :"111",
				Name :"sss",
				Msg :"超过100",
			},
		}
	}
	if len(newCode)>100{
		newCode = []strockInfo{
			strockInfo{
				Code :"111",
				Name :"sss",
				Msg :"超过100",
			},
		}
	}
	if len(insCode)>0 ||len(desCode)>0 ||len(newCode)>0 {
		//生成邮件内容
		var mailcontent string
		mailcontent += "增持:<br/>"
		for _,item := range insCode{
			mailcontent += item.Name+"("+item.Code+")"+" "+item.Msg+"<br/>"
		}
		mailcontent += "减持:<br/>"
		for _,item := range desCode{
			mailcontent +=  item.Name+"("+item.Code+")"+" "+item.Msg+"<br/>"
		}
		mailcontent += "新增:<br/>"
		for _,item := range newCode{
			mailcontent +=  item.Name+"("+item.Code+")"+"<br/>"
		}
		fmt.Println(mailcontent)
		sendAlarm(mailcontent)
	}
	fmt.Println("run time:" ,time.Since(runtimestart))
}
func loadIni(configFile string)  {
	locconfig, errr := Goconfig.LoadConfigFile(configFile)
	if errr!=nil{
		fmt.Println(errr)
		os.Exit(1)
	}

	Dbdsn, _ := locconfig.GetValue(Goconfig.DEFAULT_SECTION, "Dbdsn")
	Emaillist,_:=locconfig.GetValue(Goconfig.DEFAULT_SECTION, "Emaillist")
	DbPrefix, _ := locconfig.GetValue(Goconfig.DEFAULT_SECTION, "DbPrefix")
	Debug, _ := locconfig.Bool(Goconfig.DEFAULT_SECTION, "debug")
	if Dbdsn == ""{
		fmt.Println("no set dbdsn")
		os.Exit(1)
	}
	Alarmper, _ := locconfig.GetValue(Goconfig.DEFAULT_SECTION, "Alarmper")
	config.Gconfig.Dbdns = Dbdsn
	config.Gconfig.Debug = Debug
	config.Gconfig.DbPrefix = DbPrefix
	config.Gconfig.Emaillist = Emaillist

	val, err := strconv.ParseFloat(Alarmper, 64)
	if err!=nil{
		fmt.Println("Alarmper fail")
		os.Exit(1)
	}
	if val == 0 {
		val = 0.1
	}
	config.Gconfig.Alarmper = val
}
func  sendAlarm(content string)  {
	var lastIndex uint64
	apiconfig := api.DefaultConfig()
	apiconfig.Address = "127.0.0.1:8500" //consul server

	client, err := api.NewClient(apiconfig)
	if err != nil {
		fmt.Println("api new client is failed, err:", err)
		return
	}
	mservices, metainfo, err := client.Health().Service("sendmailrpc", "v1.0", true, &api.QueryOptions{
		WaitIndex: lastIndex, // 同步点，这个调用将一直阻塞，直到有新的更新
	})
	if err != nil {
		fmt.Println("error Consul: %v", err)
		return
	}
	lastIndex = metainfo.LastIndex

	if len(mservices)<=0{
		fmt.Println("no find rpc server")
	}

	emailarr := strings.Split(config.Gconfig.Emaillist,"|")
	for _, iservice := range mservices {
		for _,email :=range emailarr{
			email  = strings.TrimSpace(email)
			if email == ""{
				continue
			}
			service.Sendmail(iservice.Service.Address,strconv.Itoa(iservice.Service.Port),service.EmailContent{
				Touser: email,
				Title:   "每日预警邮件",
				Content: content,
			})
		}
		break
	}
}