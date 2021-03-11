package main

import (
	"encoding/json"
	"fmt"
	"github.com/antchfx/htmlquery"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"minirpc/config"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"
	"github.com/zarte/comutil/Goconfig"
)

type StackInfo struct {
	Name string `json:"name"`
	Code string `json:"code"`
	Val string `json:"val"`
	Per string `json:"per"`
	Ptype string `json:"ptype"`
	Daystr string `json:"daystr"`
}
type StockHistory struct {
	Code string `json:"code"`
	List []StockHistoryItem `json:"list"`
}
type StockHistoryItem struct {
	Date string `json:"date"`
	Sprice float64 `json:"sprice"`
	Eprice float64 `json:"eprice"`
	Cval float64 `json:"cval"`
	Cper float64 `json:"cper"`
	Cnum float64 `json:"cnum"`
	Ctval float64 `json:"ctval"`
	Chand float64 `json:"chand"`
}


var Ofile string
var Gdreg *regexp.Regexp
var Gdreg2 *regexp.Regexp

var otherinfoarr []string

func main() {
	config.SetConfig()
	loadIni(config.Gconfig.CurExePath+"config.ini")
	fmt.Println("begin")
	//getShDayDetail()
	//getSzDayDetail()
	stockAnaly("300059")
	fmt.Println("complete")

}
func getShDayDetail( ){
	data := make(map[string]string)
	data["t"] ="sh"

	header := make(map[string]string)
	header["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.190 Safari/537.36"
	res,err:=curltest("https://www.hkexnews.hk/sdw/search/mutualmarket_c.aspx",data,header,"GET")

	if err!=nil{
		fmt.Println(err)
	}else{
		//解析
		doc,_ := htmlquery.Parse(strings.NewReader(res))
		trlist := htmlquery.Find(doc, `//*[@id="mutualmarket-result"]/tbody/tr`)
		for _,n  := range trlist {
			var stuckinfo StackInfo
			content := htmlquery.FindOne(n,".//td[1]/div[2]")
			stuckinfo.Code = htmlquery.InnerText(content)
			if stuckinfo.Code==""{
				fmt.Println("craw detail fail")
				continue
			}
			content = htmlquery.FindOne(n,".//td[2]/div[2]")
			stuckinfo.Name = htmlquery.InnerText(content)
			stuckinfo.Name = strings.Replace(stuckinfo.Name, " ", "", -1)
			content = htmlquery.FindOne(n,".//td[3]/div[2]")
			stuckinfo.Val = htmlquery.InnerText(content)
			stuckinfo.Val = strings.Replace(stuckinfo.Val, ",", "", -1)
			content = htmlquery.FindOne(n,".//td[4]/div[2]")
			stuckinfo.Per = strings.Trim(htmlquery.InnerText(content),"%")
			stuckinfo.Ptype = "1"
			if AddGgDayInfo(stuckinfo, config.Gconfig.Nowdaystr) == false {
				fmt.Println("insert fail:" + stuckinfo.Code)
			}
		}
	}
}
func getSzDayDetail( ){
	data := make(map[string]string)
	data["t"] ="sz"

	header := make(map[string]string)
	header["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.190 Safari/537.36"
	res,err:=curltest("https://www.hkexnews.hk/sdw/search/mutualmarket_c.aspx",data,header,"GET")

	if err!=nil{
		fmt.Println(err)
	}else{
		//解析
		doc,_ := htmlquery.Parse(strings.NewReader(res))
		trlist := htmlquery.Find(doc, `//*[@id="mutualmarket-result"]/tbody/tr`)
		for _,n  := range trlist {
			var stuckinfo StackInfo
			content := htmlquery.FindOne(n,".//td[1]/div[2]")
			stuckinfo.Code = htmlquery.InnerText(content)
			if stuckinfo.Code==""{
				fmt.Println("craw detail fail")
				continue
			}
			content = htmlquery.FindOne(n,".//td[2]/div[2]")
			stuckinfo.Name = htmlquery.InnerText(content)
			stuckinfo.Name = strings.Replace(stuckinfo.Name, " ", "", -1)
			content = htmlquery.FindOne(n,".//td[3]/div[2]")
			stuckinfo.Val = htmlquery.InnerText(content)
			stuckinfo.Val = strings.Replace(stuckinfo.Val, ",", "", -1)
			content = htmlquery.FindOne(n,".//td[4]/div[2]")
			stuckinfo.Per = strings.Trim(htmlquery.InnerText(content),"%")
			stuckinfo.Ptype = "2"
			if AddGgDayInfo(stuckinfo,config.Gconfig.Nowdaystr) == false {
				fmt.Println("insert fail:" + stuckinfo.Code)
			}
		}
	}
}

func checkRaise(list StockHistory,dayarr map[int]string) bool {
	flag := true
	raseCunt :=0
	startval := 0.0
	for n:=0 ; n<len(dayarr); n++ {
		if dayarr[n]!="" {
			for _,v :=range list.List{
				if v.Date == dayarr[n]{
					if startval ==0 {
						startval = v.Eprice
					}else if startval>v.Eprice{
						startval = v.Eprice
						raseCunt++
					}else{
						fmt.Println(startval,v.Eprice)
						//未涨
						flag = false
					}
					break
				}
			}
		}
		if !flag{
			break
		}
	}
	if raseCunt>=3{
		//连涨3天
		fmt.Println("连涨"+strconv.Itoa(raseCunt)+"天")
		return true
	}
	return false
}
func stockAnaly(code string)  {
	dayNum := 7
	end :=time.Now().Format("20060102")
	start :=time.Now().AddDate(0,0,-dayNum).Format("20060102")
	dayarr := make(map[int]string,dayNum)
	for {
		week := time.Now().AddDate(0,0,-dayNum).Weekday()
		if week ==0 || week==6 {
			dayarr[dayNum] =""
		}else{
			dayarr[dayNum] =  time.Now().AddDate(0,0,-dayNum).Format("2006-01-02")
		}
		dayNum--
		if dayNum<0 {
			break
		}
	}
	list := gethistory(code,start,end)
	fmt.Println(list)
	checkRaise(list,dayarr)

}
func gethistory(code string,start string,end string ) StockHistory  {
	var list StockHistory
	list.List = make([]StockHistoryItem,0)
	//https://q.stock.sohu.com/hisHq?code=cn_002960&start=20210301&end=20210309&stat=1&order=D&period=d&callback=historySearchHandler&rt=json
	data := make(map[string]string)
	data["code"] ="cn_"+code
	data["start"] =start
	data["end"] =end
	data["stat"] ="1"
	data["order"] = "D"
	data["period"] = "d"
	data["callback"] = "historySearchHandler"
	data["rt"] = "json"
	header := make(map[string]string)
	header["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.190 Safari/537.36"

	res,err:=curltest("https://q.stock.sohu.com/hisHq",data,header,"GET")
	//["2018-05-10","3.02","3.05","0.04","1.33%","3.02","3.06","118307","3601.96","0.52%"]
	//[日期，开盘价，收盘价，涨跌额，涨跌幅，最低，最高，成交量，成交额，换手率]
	if err !=nil {
		fmt.Println(err)
		return list
	}else{
		var m  []interface{}
		err = json.Unmarshal([]byte(res), &m)
		tcode := m[0].(map[string]interface{})["code"].(string)
		if tcode!="cn_"+code{
			return list
		}

		dlst := m[0].(map[string]interface{})["hq"].([]interface{})
		for _,item := range dlst{
			var titem StockHistoryItem
			titem.Date = item.([]interface{})[0].(string)
			v, err := strtof64(item.([]interface{})[1].(string))
			if err != nil{
				fmt.Println(err)
				return list
			}
			titem.Sprice = v
			v, err = strtof64(item.([]interface{})[2].(string))
			if err != nil{
				fmt.Println(err)
			}
			titem.Eprice = v
			v, err = strtof64(item.([]interface{})[3].(string))
			if err != nil{
				fmt.Println(err)
			}
			titem.Cval = v
			v, err = strtof64(strings.Trim(item.([]interface{})[4].(string),"%"))
			if err != nil{
				fmt.Println(err)
			}
			titem.Cper = v
			v, err = strtof64(item.([]interface{})[7].(string))
			if err != nil{
				fmt.Println(err)
			}
			titem.Cnum = v
			v, err = strtof64(item.([]interface{})[8].(string))
			if err != nil{
				fmt.Println(err)
			}
			titem.Ctval = v
			v, err = strtof64(strings.Trim(item.([]interface{})[9].(string),"%"))
			if err != nil{
				fmt.Println(err)
			}
			titem.Chand = v
			list.List = append(list.List,titem)
		}
	}
	list.Code = code
	return list
}

func strtof64(str string) (float64,error)  {
	v, err := strconv.ParseFloat(str, 64)
	if err != nil{
		return 0,err
	}
	return v,nil
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
	config.Gconfig.Nowdaystr = time.Now().Format("20060102")
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

func curltest(ourl string,data map[string]string, header map[string]string,dtype string) (string,error) {

	reader := strings.NewReader("")

	var mentype string
	if dtype =="JSON" {
		mentype = "POST"
	}else{
		mentype = dtype
	}
	if data!=nil {
		if mentype == "POST"{
			str, err := json.Marshal(data)
			if err != nil {
				fmt.Println("json.Marshal failed:", err)
				return "",err
			}
			reader = strings.NewReader(string(str))
		} else if mentype== "GET"{
			params := url.Values{}
			parseURL, err := url.Parse(ourl)
			if err != nil {
				log.Println("err")
			}
			for key,val := range data {
				params.Set(key, val)
			}
			//如果参数中有中文参数,这个方法会进行URLEncode
			parseURL.RawQuery = params.Encode()
			ourl = parseURL.String()
		}

	}
	//fmt.Println(ourl)
	// request, err := http.Get(ourl)
	//fmt.Println(reader)
	request, err := http.NewRequest(mentype, ourl, reader)
	if err != nil {
		return "",err
	}
	if dtype =="JSON" {
		request.Header.Set("Content-Type", "application/json;charset=utf-8")
	}


	for key,item := range header{
		request.Header.Set(key,item)
	}

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return "",err
	}
	//utf8Reader := transform.NewReader(resp.Body,simplifiedchinese.GBK.NewDecoder())
	respBytes, err := ioutil.ReadAll(resp.Body)
	//respBytes, err := ioutil.ReadAll(utf8Reader)
	if err != nil {
		return "",err
	}
	//byte数组直接转成string，优化内存

	//utf8 := mahonia.NewDecoder("utf8").ConvertString(string(respBytes))
	//ioutil.WriteFile("./output2.txt", respBytes, 0666) //写入文件(字节数组)
	res := (*string)(unsafe.Pointer(&respBytes))
	return *res,nil
}
