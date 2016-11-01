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
var StockInfoDB *sql.DB
var StockInMemDB map[string][]Stock
var StockCodeNameMap map[string]string

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

type StockWebInfo struct {
	Name string
	Data []*StockInfo
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

func GetCSS(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.Method)
	fmt.Println(req.URL)
	fmt.Println(req.Proto)
	if strings.Contains(fmt.Sprintf("%v", req.URL), ".css") {
		w.Header().Set("Content-Type", "text/css")
		t, err := tt.ParseFiles(fmt.Sprintf("./%v", req.URL))
		checkError(err)
		t.Execute(w, nil)
	}
}

func GetFonts(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.Method)
	fmt.Println(req.URL)
	fmt.Println(req.Proto)
	if strings.Contains(fmt.Sprintf("%v", req.URL), ".css") {
		if strings.Contains(fmt.Sprintf("%v", req.URL), ".svg") {
			w.Header().Set("Content-Type", "image/svg+xmz")
		} else if strings.Contains(fmt.Sprintf("%v", req.URL), ".woff") {
			w.Header().Set("Content-Type", "application/x-font-woff")
		}
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

func MainPage(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL)
	t, err := template.ParseFiles("index.tmpl")
	if err != nil {
		log.Println(err.Error())
		return
	}

	t.Execute(w, nil)
}

func Register(w http.ResponseWriter, req *http.Request) {
	t, err := template.ParseFiles("register.tmpl")
	if err != nil {
		log.Println(err.Error())
		return
	}
	t.Execute(w, nil)
}

func Login(w http.ResponseWriter, req *http.Request) {
	t, err := template.ParseFiles("login.tmpl")
	if err != nil {
		log.Println(err.Error())
		return
	}
	t.Execute(w, nil)
}

func ShowForum(w http.ResponseWriter, req *http.Request) {
	t, err := template.ParseFiles("forum.tmpl")
	if err != nil {
		log.Println(err.Error())
		return
	}
	t.Execute(w, nil)
}

func GetStockInfo(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL)
	url := fmt.Sprintf("%v", req.URL)
	parts := strings.Split(url, "/")
	fmt.Println(parts)
	fmt.Println(len(parts))

	//Routing problem :(, more check
	if len(parts) != 3 {
		//not found
		return
	}

	t, err := template.ParseFiles("stockinfo.tmpl")
	if err != nil {
		log.Println(err.Error())
		return
	}
	sifs := DumpStockInfo(parts[len(parts)-1])
	t.Execute(w, sifs)
}

func Logout(w http.ResponseWriter, req *http.Request) {

}

func DumpStockInfo(code string) *StockWebInfo {
	suffix, err := getStockSuffix(code)
	if err != nil {
		fmt.Println(err.Error)
		//return errors.New("Invalid code")
	}

	table := suffix + code
	rows, err := StockInfoDB.Query("select * from " + table)
	defer rows.Close()
	checkError(err)

	var swf StockWebInfo
	swf.Data = make([]*StockInfo, 0, 1000)
	for rows.Next() {
		var stock StockInfo
		if err := rows.Scan(&stock.Date, &stock.Open, &stock.High, &stock.Low, &stock.Close, &stock.Volume, &stock.Adj); err != nil {
			log.Println("Get stock information for ", code, " error with: ", err.Error())
			continue
		}
		if err := rows.Err(); err != nil {
			log.Println("Rows error: ", err.Error(), " happend")
			continue
		}
		swf.Data = append(swf.Data, &stock)
	}
	swf.Name = StockCodeNameMap[code]

	return &swf
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
		StockCodeNameMap[stock.StockCode] = stock.ShortName
	}

	return stockList
}

func main() {
	http.HandleFunc("/test", PrintTest)
	http.HandleFunc("/static/css/", GetCSS)
	http.HandleFunc("/static/js/", GetJS)
	http.HandleFunc("/static/fonts", GetFonts)
	http.HandleFunc("/static/favicons", GetStatic)
	http.HandleFunc("/stock", ShowStockList)
	http.HandleFunc("/register", Register)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/logout", ShowStockList)
	http.HandleFunc("/si/", GetStockInfo)
	http.HandleFunc("/forum", ShowForum)
	http.HandleFunc("/", MainPage)
	http.ListenAndServe(":1234", nil)
	//	DumpAllStock("shanghai")
	//	DumpAllStock("shenzhen")
}

func init() {
	StockDB, _ = sql.Open("mysql", "kkkmmu:leeweop@/stock?charset=utf8")
	StockInfoDB, _ = sql.Open("mysql", "kkkmmu:leeweop@/stockinfo?charset=utf8")
	StockInMemDB = make(map[string][]Stock, 2)
	StockCodeNameMap = make(map[string]string, 10000)
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
