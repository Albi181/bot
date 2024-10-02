package main

import (
	"automata/go/src/BOTPAYEER/binance"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var SS float64
var Chipi float64 = 1.005

type pairs string

const (
	BTC_USD pairs = "BTC_USD"
	BTC_RUB pairs = "BTC_RUB"
)

type OrdersReque struct {
	Status bool               `json:"success"`
	Pairs  map[pairs]AsksBids `json:"pairs"`
}

type AsksBids struct {
	Ask  string         `json:"ask"`
	Bid  string         `json:"bid"`
	Asks []AsksBisdInfo `json:"asks"`
	Bids []AsksBisdInfo `json:"bids"`
}

type AsksBisdInfo struct {
	Price  string `json:"price"`
	Amount string `json:"amount"`
	Value  string `json:"value"`
}

var Balas Balancejson

type request struct {
	Timestamp int64 `json:"ts"`
}

type ordersreq struct {
	Pair string `json:"pair"`
}

type postorderreq struct {
	Pair      string `json:"pair"`
	Type      string `json:"type"`
	Action    string `json:"action"`
	Amount    string `json:"amount"`
	Timestamp int64  `json:"ts"`
}

type Balancejson struct {
	Status  bool                  `json:"success"`
	Balance map[currency]balances `json:"balances"`
}
type currency string

const (
	USD  currency = "USD"
	RUB  currency = "RUB"
	EUR  currency = "EUR"
	USDT currency = "USDT"
)

type balances struct {
	Total     float64 `json:"total"`
	Available float64 `json:"available"`
	Hold      float64 `json:"hold"`
}

var binask string // цена бинанс аск (стримится функцией Bintic2)
var ords OrdersReque

func main() {

	var mx sync.Mutex

	pair := "BTC_USD"
	var pairr pairs = "BTC_USD"

	go Bintic2()
	go OrdersRequest(pair, &mx)
	go chooseSell(pairr, &mx)

	select {}

	// var mutexbinance sync.Mutex
	// var mutexpayeer sync.Mutex
	//	typee := "market"
	//	action := "sell"
	//	amount := "0.00000001"
	//PostOrder(pair, typee, action, amount)
	//var price []float64 = ords.Pairs[BTC_USD].Asks[].Price
	//for index, value := range ords.Pairs[BTC_USD].Asks[].Price {
	//	fmt.Println(index, value)
	//}
	//	ords.Pairs[BTC_USD].Asks[3].Price, 64)
}

/*
	func checkOrds(mx *sync.Mutex) {
		var b string
		for {
			mx.Lock()
			if ords.Pairs[BTC_USD].Ask != "5" {
				b = ords.Pairs[BTC_USD].Ask
				fmt.Println(b)
			}
			mx.Unlock()
		}
	}
*/
func Bintic2() {

	Orders2 := binance.NewClient().SubscribeTicker(binance.SYMBOL_BTCUSDT, time.Millisecond*100)

	for Order2 := range Orders2 {

		binask = Order2.AskPrice
		//fmt.Println(binask)
	}
	fmt.Println("There were no reasons to be here =( ... ", binask)
}

func OrdersRequest(pair string, mx *sync.Mutex) {

	var r *http.Response
	baseURL := "https://payeer.com/api/trade/"
	endpoint := "orders"
	req := ordersreq{
		Pair: pair,
	}

	rBody, err := json.Marshal(req)
	if err != nil {
		fmt.Println("rBody Marshal error -> ", err)
	}

	for {
		for {
			r, err = http.Post(baseURL+endpoint, "application/json", bytes.NewBuffer(rBody))
			if err == nil {
				break
			}
			fmt.Println("request error -> ", err)
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println("responce error -> ", err)
		}
		//fmt.Println("response body->", string(bodyBytes))

		mx.Lock()
		json.Unmarshal(bodyBytes, &ords)
		mx.Unlock()
		//fmt.Println(ords.Pairs[BTC_USD].Asks[1].Price)

		time.Sleep(1 * time.Second)
	}
}

func chooseSell(pairr pairs, mx *sync.Mutex) {
	//var amountToSell float64
	// var payeerOrdersResponce OrdersReque

	BalanceRequest()
	fmt.Println(Balas.Balance["BTC"].Available)
	time.Sleep(5 * time.Second)

	for {

		mx.Lock()
		SS = 0
		for _, Asks := range ords.Pairs[pairr].Asks {

			p, err := strconv.ParseFloat(Asks.Price, 64)
			if err != nil {
				fmt.Println("p error -> ", err)
			}
			b, err := strconv.ParseFloat(binask, 64)
			if err != nil {
				fmt.Println("b error -> ", err)
			}
			c, err := strconv.ParseFloat(Asks.Amount, 64)
			if err != nil {
				fmt.Println("c error -> ", err)
			}
			//fmt.Println(p, b, c)
			if p/b < 1.1 { //										вернуть 0.99
				SS += c
			} else {
				break
			}
		}
		fmt.Println("SS -> ", SS)

		mx.Unlock()

		//amountToSell = min(Balas.Balance["BTC"].Available, SS)
		//	fmt.Println("amount avaliable for sell ", amountToSell)

		//	if amountToSell >= 0.00001 {

		//  	PostOrder(pair string, typee string, action string, amount string)
		//		time.Sleep(500 * time.Millisecond)
		//		BalanceRequest()
		//	}
	}
}

//func mincopy() {
//	var amount float64
//
//	BalanceRequest()
//
//	amount = min(Balas.Balance[USD].Total, 0.01) // вместо 0.01 будет полученный эмаунт для выбора
//	if amount >= 0.0001 {                        // минимальное значение для заявки
//		//Post order(amount)
//	}
//	time.Sleep(1 * time.Second)

//}

func BalanceRequest() {
	var resp *http.Response
	baseURL := "https://payeer.com/api/trade/"
	endpoint := "account"
	req := request{
		Timestamp: time.Now().UnixMilli(),
	}

	rBody, err := json.Marshal(req)
	if err != nil {
		fmt.Println("rBody Marshal error #1 -> ", err)
	}

	r, err := http.NewRequest(http.MethodPost, (baseURL + endpoint), bytes.NewBuffer(rBody))
	if err != nil {
		fmt.Println("http.NewRequest error #1 -> ", err)
	}

	client := http.Client{
		Timeout: 10 * time.Hour,
	}

	secret := "plIzgsI8akwDumrU"
	data := endpoint + string(rBody)
	fmt.Println("string(rBody) -> ", string(rBody))

	h := hmac.New(sha256.New, []byte(secret)) // Create a new HMAC by defining the hash type and the key (as byte array)
	h.Write([]byte(data))                     // Write Data to it
	sha := hex.EncodeToString(h.Sum(nil))     // Get result and encode as hexadecimal string

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("API-ID", "3dc19f11-b03f-4d38-be50-c33a2585e12d")
	r.Header.Set("API-SIGN", sha)

	for {
		resp, err = client.Do(r)
		if err == nil {
			break
		}
		fmt.Println("failed to create request #1 -> ", err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed to read resp.Body #1 -> ", err)
	}

	json.Unmarshal(bodyBytes, &Balas)
}

func PostOrder(pair string, typee string, action string, amount string) {

	baseURL := "https://payeer.com/api/trade/"
	endpoint := "order_create"

	req := postorderreq{
		Pair:      pair,
		Type:      typee,
		Action:    action,
		Amount:    amount,
		Timestamp: time.Now().UnixMilli(),
	}

	rBody := bytes.NewBuffer(nil)      //создает буфер бади
	json.NewEncoder(rBody).Encode(req) // энкодит реквест и записывает в бади
	//	fmt.Println(rBody)
	r, _ := http.NewRequest(http.MethodPost, (baseURL + endpoint), rBody)

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	secret := "plIzgsI8akwDumrU"
	data := endpoint + rBody.String()

	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(secret))
	// Write Data to it
	h.Write([]byte(data))

	// Get result and encode as hexadecimal string
	sha := hex.EncodeToString(h.Sum(nil))

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("API-ID", "3dc19f11-b03f-4d38-be50-c33a2585e12d")
	r.Header.Set("API-SIGN", sha)

	resp, err := client.Do(r)
	if err != nil {
		fmt.Printf("failed to create request %v", err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("response body-1>", string(bodyBytes))

	//	json.Unmarshal(bodyBytes, &Balas)

	//	fmt.Printf("%+v", Balas)

	//	fmt.Println(Balas.Balance[USD])

}
