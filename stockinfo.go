package main

import (
	"database/sql"
	//  "encoding/json"
	//"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	//"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// all company of shenzhen
// http://www.szse.cn/szseWeb/ShowReport.szse?SHOWTYPE=EXCEL&CATALOGID=1110&tab1PAGENUM=1&ENCODE=1&TABKEY=tab1

// all company of shanghai
//"http://query.sse.com.cn/security/stock/downloadStockListFile.do?csrcCode=&stockCode=&areaName=&stockType=1",
//"http://query.sse.com.cn/security/stock/downloadStockListFile.do?csrcCode=&stockCode=&areaName=&stockType=2",

// show columns from XXXX ----> get the property of each columns

var StockDB *sql.DB
var StockInfoDB *sql.DB
var wg sync.WaitGroup

var ShenZhenStartupPrefix string = "300"
var ShenZhenMiddleLittePrefix string = "002"
var ShenZhenMainPrefix string = "000"
var ShangHaiMainPrefix string = "60"
var YahooGetHistoryURL string = "http://table.finance.yahoo.com/table.csv?s="

type StockInfo struct {
	Name   string
	Code   string
	Date   string
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume uint64
	Adj    float64
}

type prefixError struct {
	Code  int
	Stock string
}

func (e prefixError) Error() string {
	return fmt.Sprintf("Unknown Stock Code %s, Error code is %d", e.Stock, e.Code)
}

func getStockSuffix(code string) (string, error) {
	if strings.HasPrefix(code, ShenZhenStartupPrefix) {
		return "sz", nil
	} else if strings.HasPrefix(code, ShenZhenMiddleLittePrefix) {
		return "sz", nil
	} else if strings.HasPrefix(code, ShenZhenMainPrefix) {
		return "sz", nil
	} else if strings.HasPrefix(code, ShangHaiMainPrefix) {
		return "ss", nil
	} else {
		return "", prefixError{Code: -1, Stock: code}
	}
}

func dropAllTables() {
	db, err := sql.Open("mysql", "kkkmmu:leeweop@/yahoo?charset=utf8")
	checkError(err)
	defer db.Close()
	tables, err := db.Query("show table status from yahoo")
	for tables.Next() {
		var name sql.NullString
		tables.Scan(&name)
		col, _ := tables.Columns()
		fmt.Println(col)
		fmt.Println(name)
	}
}

func FetchAllTypeAStockHistory(market string) error {
	allStock, err := StockDB.Query("select StockCode from " + market)
	checkError(err)

	limiter := time.Tick(time.Second * 5)
	for allStock.Next() {
		<-limiter
		var code string
		allStock.Scan(&code)

		//		wg.Add(1)
		//		go func(code string) {
		//			defer wg.Done()
		log.Println(code)
		suffix, err := getStockSuffix(code)
		if err != nil {
			fmt.Println(err.Error)
			continue
			//return errors.New("Invalid code")
		}

		table := suffix + code

		client := http.Client{
		/*
			Transport: &http.Transport{
				Dial: func(netw, addr string) (net.Conn, error) {
					deadline := time.Now().Add(5 * time.Second)
					c, err := net.DialTimeout(netw, addr, time.Second*30)
					if err != nil {
						return nil, err
					}
					c.SetDeadline(deadline)
					return c, nil
				},
			},
		*/
		}

		url := YahooGetHistoryURL + code + "." + suffix
		fmt.Println(url)
		request, err := http.NewRequest("GET", url, nil)
		checkError(err)

		response, err := client.Do(request)
		if err != nil {
			log.Println("Get ", url, " expired: ", err.Error(), ", retry agian")
			continue
			//response, err = client.Do(request)
		}
		if response.StatusCode == 404 {
			log.Println("Cannot find: ", url)
			continue
		}
		result, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println("Errore while reading body of ", url)
			continue
		}
		log.Println("==========response.StatusCode: ", response.StatusCode, " =======================")
		log.Println(string(result))

		defer response.Body.Close()
		//fmt.Println(string(result))

		_, err = StockInfoDB.Query("CREATE TABLE IF NOT EXISTS " + table + " (Date char(64) not null primary key, Open double not null, High double not null, Low double not null, Close double not null, Volume bigint(64) not null, Adj double not null)")
		checkError(err)

		insert, err := StockInfoDB.Prepare("INSERT IGNORE INTO " + table + "(Date, Open, High, Low, Close, Volume, Adj) VALUES (?, ?, ?, ?, ?, ?, ?)")
		checkError(err)

		str := strings.Split(string(result), "\n")
		for i, r := range str {
			//fmt.Println(r)
			if i == 0 {
				continue
			}

			if len(r) == 0 {
				continue
			}

			token := strings.Split(r, ",")
			var stock StockInfo
			stock.Date = token[0]
			stock.Open, _ = strconv.ParseFloat(token[1], 64)
			stock.High, _ = strconv.ParseFloat(token[2], 64)
			stock.Low, _ = strconv.ParseFloat(token[3], 64)
			stock.Close, _ = strconv.ParseFloat(token[4], 64)
			stock.Volume, _ = strconv.ParseUint(token[5], 10, 64)
			stock.Adj, _ = strconv.ParseFloat(token[6], 64)
			fmt.Println(stock)

			_, err = insert.Exec(stock.Date, stock.Open, stock.High, stock.Low, stock.Close, stock.Volume, stock.Adj)
			checkError(err)
		}
		insert.Close()
		//		}(code)
	}

	if err = allStock.Err(); err != nil {
		return err
	}
	return nil
}

func main() {
	FetchAllTypeAStockHistory("shenzhen")
	FetchAllTypeAStockHistory("shanghai")
	wg.Wait()
}

func checkError(err error) {
	if err != nil {
		if driverErr, ok := err.(*mysql.MySQLError); ok {
			if driverErr.Number == 1062 {
				return
			}
		}
		fmt.Println(err.Error())
		fmt.Println(err)
	}
}

func init() {
	StockDB, _ = sql.Open("mysql", "kkkmmu:leeweop@/stock?charset=utf8")
	StockInfoDB, _ = sql.Open("mysql", "kkkmmu:leeweop@/stockinfo?charset=utf8")
}
