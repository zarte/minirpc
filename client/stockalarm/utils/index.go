package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"minirpc/Model"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type StackInfo struct {
	Name string `json:"name"`
	Code string `json:"code"`
	Ocode string `json:"ocode"`
	Gcode string `json:"gcode"`
	Val string `json:"val"`
	Per string `json:"per"`
	//位置
	Ptype string `json:"ptype"`
	Daystr string `json:"daystr"`
}
type StockHistory struct {
	Code string `json:"code"`
	Todaye string `json:"todaye"`
	Ydaye string `json:"ydaye"`
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

func StockAnaly(code string) string {
	var retstr string
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
	retstr +="昨到今：" +list.Ydaye +"~"+list.Todaye
	if checkRaise(list,dayarr) {
		retstr += "  连涨3天!!"
	}
	if checkDec(list,dayarr) {
		retstr += "  连跌3天!!"
	}
	return  retstr

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
func checkDec(list StockHistory,dayarr map[int]string) bool {
	flag := true
	raseCunt :=0
	startval := 0.0
	for n:=0 ; n<len(dayarr); n++ {
		if dayarr[n]!="" {
			for _,v :=range list.List{
				if v.Date == dayarr[n]{
					if startval ==0 {
						startval = v.Eprice
					}else if startval<v.Eprice{
						startval = v.Eprice
						raseCunt++
					}else{
						//fmt.Println(startval,v.Eprice)
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
		//连跌3天
		fmt.Println("连跌"+strconv.Itoa(raseCunt)+"天")
		return true
	}
	return false
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
	nowdaystr := time.Now().Format("2006-01-02")
	nowweek := time.Now().Weekday()
	lastdaystr :=""
	if nowweek==1 {
		lastdaystr = time.Now().AddDate(0,0,-3).Format("2006-01-02")
	}else{
		lastdaystr = time.Now().AddDate(0,0,-1).Format("2006-01-02")
	}
	if err !=nil {
		fmt.Println(err)
		return list
	}else{
		var m  []interface{}
		err = json.Unmarshal([]byte(res), &m)
		if err != nil {
			fmt.Println(data["code"]," gethistory json fail")
			fmt.Println(res)
			fmt.Println(err)
			return list
		}
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
			if nowdaystr == titem.Date{
				list.Todaye = item.([]interface{})[2].(string)
			}
			if lastdaystr ==titem.Date{
				list.Ydaye = item.([]interface{})[2].(string)
			}
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

func GetStockBaseInfo(gcode string,name string) StackInfo{
	//查询数据库
	info,err := Model.GetOneStockInfo(gcode)
	if err!=nil{
		fmt.Println(err)
	}
	if info.Name !="" {
		return  StackInfo{
			Name:info.Name,
			Code:info.Code,
			Gcode:info.Gcode,
			Ptype :info.Type,
		}
	}else {
		//爬取
		info2 :=CrawStockBaseInfo(name)
		if info2.Name !=""{
			//入库
			Model.AddOneStockInfo(info2.Name,info2.Code,info2.Ptype,gcode,name)
			return  StackInfo{
				Name:info2.Name,
				Code:info2.Code,
				Gcode:info2.Gcode,
				Ptype :info2.Ptype,
			}
		}else{
			return  StackInfo{
				Name:"",
			}
		}
	}
}
func CrawStockBaseInfo(name string) StackInfo{
	var stockInfo StackInfo
	//https://quotes.money.163.com/stocksearch/json.do?type=HS&count=5&word=%E4%B8%AD%E9%87%91%E5%85%AC%E5%8F%B8&t=0.40060858273627353

	data := make(map[string]string)
	data["word"] =name
	data["type"] ="HS"
	data["count"] = "1"
	header := make(map[string]string)
	header["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.190 Safari/537.36"

	res,err:=curltest("https://quotes.money.163.com/stocksearch/json.do",data,header,"GET")
	//_ntes_stocksearch_callback([{"code":"0601995","type":"SH","symbol":"601995","tag":"HS MYHS","spell":"zjgs","name":"\u4e2d\u91d1\u516c\u53f8"}])
	if err!= nil {
		fmt.Println(err)
		return stockInfo
	}
	if len(res)<27 {
		fmt.Println("get stockInfo fail")
		return stockInfo
	}
	res = res[27:len(res)-1]
	var m  []interface{}
	err = json.Unmarshal([]byte(res), &m)
	if err != nil {
		fmt.Println(name)
		fmt.Println(res)
		fmt.Println(err)
		return stockInfo
	}
	/*
	if name != m[0].(map[string]interface{})["name"].(string){
		fmt.Println(name)
		fmt.Println(m[0].(map[string]interface{})["name"].(string))
		fmt.Println("get stockInfo no match")
		return stockInfo
	}
	 */
	stockInfo.Name =m[0].(map[string]interface{})["name"].(string)
	stockInfo.Code = m[0].(map[string]interface{})["symbol"].(string)
	stockInfo.Ptype = m[0].(map[string]interface{})["type"].(string)
	return stockInfo
}
func strtof64(str string) (float64,error)  {
	v, err := strconv.ParseFloat(str, 64)
	if err != nil{
		return 0,err
	}
	return v,nil
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
