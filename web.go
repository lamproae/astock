package main

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	tt "text/template"
)

var StockDB *sql.DB
var StockInMemDB map[string][]Stock

type prefixError struct {
	Code  int
	Stock string
}

type Stock struct {
	StockCode             string
	ShortName             string
	TypeACode             string
	TypeAShortName        string
	TypeAOnSaleDate       string
	TypeATotalValue       int64
	TypeACirculationValue int64
}

var ShenZhenStartupPrefix string = "300"
var ShenZhenMiddleLittePrefix string = "002"
var ShenZhenMainPrefix string = "000"
var ShangHaiMainPrefix string = "60"
var YahooGetHistoryURL string = "http://table.finance.yahoo.com/table.csv?s="

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

func PrintStock(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Hello, world\n")
}

func PrintTest(w http.ResponseWriter, req *http.Request) {
	tmpl, err := template.ParseFiles("index.tmpl")
	checkError(err)

	tmpl.Execute(w, tmpl)
}

func StockList(w http.ResponseWriter, req *http.Request) {
	tmpl, err := template.ParseFiles("index.tmpl")
	checkError(err)

	tmpl.Execute(w, tmpl)
}

func GetStatic(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.Method)
	fmt.Println(req.URL)
	fmt.Println(req.Proto)
	io.WriteString(w, fmt.Sprintf("%v", req.URL))
	file, _ := os.Open(fmt.Sprintf(".%v", req.URL))
	c, _ := ioutil.ReadAll(file)
	w.Write(c)
}

func GetJS(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.Method)
	fmt.Println(req.URL)
	fmt.Println(req.Proto)
	if strings.Contains(fmt.Sprintf("%v", req.URL), ".js") {
		w.Header().Set("Content-Type", "application/javascript")
		t, err := tt.ParseFiles(fmt.Sprintf("./%v", req.URL))
		checkError(err)
		t.Execute(w, nil)
	}
}

func PrintMain(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL)
	t, err := tt.ParseFiles(fmt.Sprintf(".%v", req.URL))
	checkError(err)
	t.Execute(w, nil)
}

func ShowStockList(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL)
	t, err := template.ParseFiles("stock.tmpl")
	if err != nil {
		log.Println(err.Error())
		return
	}

	List, _ := StockInMemDB["shanghai"]
	t.Execute(w, List)
}

func DumpAllStock(market string) []Stock {
	allStock, err := StockDB.Query("select * from " + market)
	defer allStock.Close()
	if err != nil {
		log.Fatal(err.Error())
	}

	stockList := make([]Stock, 0, 1000)

	for allStock.Next() {
		var stock Stock
		if err := allStock.Scan(&stock.StockCode, &stock.ShortName, &stock.TypeACode, &stock.TypeAOnSaleDate, &stock.TypeAOnSaleDate, &stock.TypeATotalValue, &stock.TypeACirculationValue); err != nil {
			log.Println(err.Error())
			continue
		}

		//		log.Println(stock)
		if err := allStock.Err(); err != nil {
			log.Println("Rows error: ", err.Error(), " happend")
			continue
		}
		stockList = append(stockList, stock)
	}

	return stockList
}

func main() {
	http.HandleFunc("/test", PrintTest)
	http.HandleFunc("/static/css/", GetStatic)
	http.HandleFunc("/static/js/", GetJS)
	http.HandleFunc("/static/fonts", GetStatic)
	http.HandleFunc("/static/favicons", GetStatic)
	http.HandleFunc("/stock", ShowStockList)
	http.HandleFunc("/", PrintMain)
	http.ListenAndServe(":1234", nil)
	//	DumpAllStock("shanghai")
	//	DumpAllStock("shenzhen")
}

func init() {
	StockDB, _ = sql.Open("mysql", "kkkmmu:leeweop@/stock?charset=utf8")
	StockInMemDB = make(map[string][]Stock, 2)
	StockInMemDB["shenzhen"] = DumpAllStock("shenzhen")
	StockInMemDB["shanghai"] = DumpAllStock("shanghai")
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
