package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strings"
)

type Stock struct {
	StockCode             string
	ShortName             string
	TypeACode             string
	TypeAShortName        string
	TypeAOnSaleDate       string
	TypeATotalValue       int64
	TypeACirculationValue int64
}

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

var StockDB *sql.DB
var StockInfoDB *sql.DB
var GlobalError error

var ShenZhenStartupPrefix string = "300"
var ShenZhenMiddleLittePrefix string = "002"
var ShenZhenMainPrefix string = "000"
var ShangHaiMainPrefix string = "60"
var YahooGetHistoryURL string = "http://table.finance.yahoo.com/table.csv?s="

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

func main() {
	rows, err := StockDB.Query("select StockCode from shanghai")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var code string
	var stock StockInfo
	for rows.Next() {
		if err := rows.Scan(&code); err != nil {
			log.Println("Fetch Stock information error with: ", err.Error())
			continue
		}
		log.Println(code)
		if err := rows.Err(); err != nil {
			log.Println(err.Error())
			continue
		}

		name, err := getStockSuffix(code)
		if err != nil {
			log.Println("Unknown stock ", code)
			continue
		}
		dayinfos, err := StockInfoDB.Query("select * from " + name + code)
		defer dayinfos.Close()
		if err != nil {
			log.Println("Query for ", code, " with error: ", err.Error())
			continue
		}

		for dayinfos.Next() {
			if err := dayinfos.Scan(&stock.Date, &stock.Open, &stock.High, &stock.Low, &stock.Close, &stock.Volume, &stock.Adj); err != nil {
				log.Println("Get stock information for ", code, " error with: ", err.Error())
			}
			log.Println("==========================================================================")
			log.Println(stock)
			if err := rows.Err(); err != nil {
				log.Println("Rows error: ", err.Error(), " happend")
				continue
			}
		}
	}
}

func init() {
	StockDB, GlobalError = sql.Open("mysql", "kkkmmu:leeweop@/stock?charset=utf8")
	if GlobalError != nil {
		log.Println(GlobalError.Error())
	}
	StockInfoDB, GlobalError = sql.Open("mysql", "kkkmmu:leeweop@/stockinfo?charset=utf8")
	if GlobalError != nil {
		log.Println(GlobalError.Error())
	}
}
