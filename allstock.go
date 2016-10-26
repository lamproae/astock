package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/guotie/gogb2312"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// shenzhen all
// http://www.szse.cn/szseWeb/ShowReport.szse?SHOWTYPE=EXCEL&CATALOGID=1110&tab1PAGENUM=1&ENCODE=1&TABKEY=tab1

// shanhai all
//"http://query.sse.com.cn/security/stock/downloadStockListFile.do?csrcCode=&stockCode=&areaName=&stockType=1"
//"http://query.sse.com.cn/security/stock/downloadStockListFile.do?csrcCode=&stockCode=&areaName=&stockType=2"

var ShangHanStock = regexp.MustCompile(`(?P<Code>[0-9]{6})[[:space:]]+(?P<ShortName>[[\p{Han}[:word:]\*]+)[[:space:]]+(?P<aCode>[0-9]{6})[[:space:]]+(?P<aName>[[\p{Han}[:word:]\*]+)[[:space:]]+(?P<date>[0-9\-]+)[[:space:]]+(?P<totoal>[0-9\.]+)[[:space:]]+(?P<circulation>[0-9\.]+)`)

type Stock struct {
	StockCode             string
	ShortName             string
	TypeACode             string
	TypeAShortName        string
	TypeAOnSaleDate       string
	TypeATotalValue       int64
	TypeACirculationValue int64
}

func main() {
	db, err := sql.Open("mysql", "kkkmmu:leeweop@/stock?charset=utf8")
	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = db.Query("CREATE TABLE IF NOT EXISTS shenzhen (StockCode char(64) not null primary key, ShortName char(64) not null, TypeACode char(64) not null, TypeAShortName char(64) not null, TypeAOnSaleDate char(64) not null, TypeATotalValue int(64) not null, TypeACirculationValue int(64) not null) DEFAULT CHARSET=utf8")
	if err != nil {
		log.Fatal(err.Error())
	}

	req, err := http.Get("http://www.szse.cn/szseWeb/ShowReport.szse?SHOWTYPE=EXCEL&CATALOGID=1110&tab1PAGENUM=1&ENCODE=1&TABKEY=tab1")
	if err != nil {
		log.Fatal(err.Error())
	}

	result, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err.Error())
	}

	strResult, err, _, _ := gogb2312.ConvertGB2312String(string(result))
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println(strResult)

	stocks := strings.Split(strResult, "<td  class='cls-data-td' style='mso-number-format:\\@' align='center' >")
	//stocks := strings.Split(strResult, "<tr  class='cls-data-tr' bgcolor='#F8F8F8'><td  class='cls-data-td' style='mso-number-format:\\@' align='center' >")
Next:
	for _, s := range stocks {
		var stock Stock
		info := strings.Split(s, "<td  class='cls-data-td' null align=")
		for i, p := range info {
			if i == 0 {
				stock.StockCode = p[:len(p)-5]
				if strings.Contains(stock.StockCode, "http-equiv") {
					continue Next
				}
			} else if i == 1 {
				stock.ShortName = p[10 : len(p)-5]
			} else if i == 5 {
				stock.TypeACode = p[10 : len(p)-5]
			} else if i == 6 {
				stock.TypeAShortName = p[10 : len(p)-5]
			} else if i == 7 {
				stock.TypeAOnSaleDate = p[10 : len(p)-5]
			} else if i == 8 {
				total := strings.Replace(p[10:len(p)-5], ",", "", -1)
				log.Println(total)

				stock.TypeATotalValue, _ = strconv.ParseInt(total, 10, 64)
			} else if i == 9 {
				circulation := strings.Replace(p[10:len(p)-5], ",", "", -1)
				log.Println(circulation)
				stock.TypeACirculationValue, _ = strconv.ParseInt(circulation, 10, 64)
			}

		}
		log.Println(stock)
		insert, err := db.Prepare("INSERT INTO shenzhen (StockCode, ShortName, TypeACode, TypeAShortName, TypeAOnSaleDate, TypeATotalValue, TypeACirculationValue) VALUES (?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE ShortName=?, TypeACode=?, TypeAShortName=?, TypeAOnSaleDate=?, TypeATotalValue=?, TypeACirculationValue=?")
		//insert, err := db.Prepare("INSERT INTO shenzhen (StockCode, ShortName, TypeACode, TypeAShortName, TypeAOnSaleDate, TypeATotalValue, TypeACirculationValue) VALUES (?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err.Error())
		}

		_, err = insert.Exec(stock.StockCode, stock.ShortName, stock.TypeACode, stock.TypeAShortName, stock.TypeAOnSaleDate, stock.TypeATotalValue, stock.TypeACirculationValue, stock.ShortName, stock.TypeACode, stock.TypeAShortName, stock.TypeAOnSaleDate, stock.TypeATotalValue, stock.TypeACirculationValue)
		//_, err = insert.Exec(stock.StockCode, stock.ShortName, stock.TypeACode, stock.TypeAShortName, stock.TypeAOnSaleDate, stock.TypeATotalValue, stock.TypeACirculationValue)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	//(stock.ShortName, stock.ShortName, stock.TypeACode, stock.TypeAShortName, stock.TypeAOnSaleDate, stock.TypeATotalValue, stock.TypeACirculationValue) ON DUPLICATE KEY UPDATE")

	request, err := http.NewRequest("GET", "http://query.sse.com.cn/security/stock/downloadStockListFile.do?csrcCode=&stockCode=&areaName=&stockType=1", nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	request.Header.Set("Referer", "http://www.sse.com.cn/assortment/stock/list/share/")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer response.Body.Close()

	_, err = db.Query("CREATE TABLE IF NOT EXISTS shanghai (StockCode char(64) not null primary key, ShortName char(64) not null, TypeACode char(64) not null, TypeAShortName char(64) not null, TypeAOnSaleDate char(64) not null, TypeATotalValue int(64) not null, TypeACirculationValue int(64) not null) DEFAULT CHARSET=utf8")
	if err != nil {
		log.Fatal(err.Error())
	}

	result, _ = ioutil.ReadAll(response.Body)
	strResult, _, _, _ = gogb2312.ConvertGB2312String(string(result))
	log.Println(strResult)
	sts := ShangHanStock.FindAllStringSubmatch(strResult, -1)
	for _, s := range sts {
		var stock Stock
		log.Println(s[1], s[2], s[3], s[4], s[5], s[6], s[7])
		stock.StockCode = s[1]
		stock.ShortName = s[2]
		stock.TypeACode = s[3]
		stock.TypeAShortName = s[4]
		stock.TypeAOnSaleDate = s[5]
		// This part is not correct
		stock.TypeATotalValue, _ = strconv.ParseInt(strings.Replace(s[6], ".", "", -1), 10, 64)
		stock.TypeACirculationValue, _ = strconv.ParseInt(strings.Replace(s[7], ".", "", -1), 10, 64)
		log.Println(stock)

		insert, err := db.Prepare("INSERT INTO shanghai (StockCode, ShortName, TypeACode, TypeAShortName, TypeAOnSaleDate, TypeATotalValue, TypeACirculationValue) VALUES (?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE ShortName=?, TypeACode=?, TypeAShortName=?, TypeAOnSaleDate=?, TypeATotalValue=?, TypeACirculationValue=?")
		if err != nil {
			log.Fatal(err.Error())
		}

		_, err = insert.Exec(stock.StockCode, stock.ShortName, stock.TypeACode, stock.TypeAShortName, stock.TypeAOnSaleDate, stock.TypeATotalValue, stock.TypeACirculationValue, stock.ShortName, stock.TypeACode, stock.TypeAShortName, stock.TypeAOnSaleDate, stock.TypeATotalValue, stock.TypeACirculationValue)

	}
	//log.Println(strResult)
}
