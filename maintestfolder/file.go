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
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type pairs string

type OrdersResponce struct {
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

type Request struct {
	Pair      string `json:"pair"`
	Type      string `json:"type"`
	Action    string `json:"action"`
	Amount    string `json:"amount"`
	Value     string `json:"value"`
	Price     string `json:"price"`
	Timestamp int64  `json:"ts"`
	Order_id  int    `json:"order_id"`
}

type Errror struct {
	Errror ErrrorCode `json:"error"`
}

type ErrrorCode struct {
	ErrrorCode string `json:"code"`
}

const (
	INVALID_SIGNATURE         string = "INVALID_SIGNATURE"
	INVALID_IP_ADDRESS        string = "INVALID_IP_ADDRESS"
	LIMIT_EXCEEDED            string = "LIMIT_EXCEEDED"
	INVALID_TIMESTAMP         string = "INVALID_TIMESTAMP"
	ACCESS_DENIED             string = "ACCESS_DENIED"
	INVALID_PARAMETER         string = "INVALID_PARAMETER"
	PARAMETER_EMPTY           string = "PARAMETER_EMPTY"
	INVALID_STATUS_FOR_REFUND string = "INVALID_STATUS_FOR_REFUND"
	REFUND_LIMIT              string = "REFUND_LIMIT"
	UNKNOWN_ERROR             string = "UNKNOWN_ERROR"
	INVALID_DATE_RANGE        string = "INVALID_DATE_RANGE"
	INSUFFICIENT_FUNDS        string = "INSUFFICIENT_FUNDS"
	INSUFFICIENT_VOLUME       string = "INSUFFICIENT_VOLUME"
	INCORRECT_PRICE           string = "INCORRECT_PRICE"
	MIN_AMOUNT                string = "MIN_AMOUNT"
	MIN_VALUE                 string = "MIN_VALUE"
)

type PostOrderResponce struct {
	Success  bool                    `json:"success"`
	Order_id int                     `json:"order_id"`
	Params   PostOrderResponceParams `json:"params"`
}

type PostOrderResponceParams struct {
	Pair      string `json:"pair"`
	Type      string `json:"type"`
	Action    string `json:"action"`
	Amount    string `json:"amount"`
	Price     string `json:"price"`
	Value     string `json:"value"`
	StopPrice string `json:"stop_price"`
}

type CancelOrderResponce struct {
	Success bool       `json:"success"`
	Errror  ErrrorCode `json:"error"`
}

type BalanceResponce struct {
	Status  bool                  `json:"success"`
	Balance map[currency]balances `json:"balances"`
}
type currency string

type balances struct {
	Total     float64 `json:"total"`
	Available float64 `json:"available"`
	Hold      float64 `json:"hold"`
}

const (
	BTC_USD  pairs = "BTC_USD"
	BTC_RUB  pairs = "BTC_RUB"
	ETH_USDT pairs = "ETH_USDT"
)

const (
	USD  currency = "USD"
	RUB  currency = "RUB"
	EUR  currency = "EUR"
	USDT currency = "USDT"
	ETH  currency = "ETH"
)

var Balas BalanceResponce
var LimitOrdId PostOrderResponce
var CancelOrd CancelOrderResponce
var ss float64

var binask string // цена бинанс аск (стримится функцией Bintic2)
var binbid string
var ords OrdersResponce

var buyorders int64
var sellorders int64

func main() {

	var buy, sell bool
	var mx sync.Mutex
	typee := "market"
	const binpair binance.Symbol = "ETHUSDT"
	pair := "ETH_USDT"
	const pairr pairs = "ETH_USDT"
	const base currency = "ETH"
	const nonbase currency = "USDT"
	buy = true
	sell = true

	var percentOfBalance decimal.Decimal = decimal.NewFromFloat(0.15) // процент от баланса идущий на лимит ордер
	var EVtoLimitBuy decimal.Decimal = decimal.NewFromFloat(0.97)     // процент дельты от цены бинанса для лимит покупки
	BalanceRequest()

	go Bintic2(binpair)
	go OrdersRequest(pair, &mx)

	time.Sleep(10 * time.Second)

	//go LimitSell(pairr, &mx)
	go LimitBuy(pairr, pair, &mx, percentOfBalance, nonbase, EVtoLimitBuy)
	go MarketSell(pairr, &mx, pair, typee, sell, base)
	go MarketBuy(pairr, &mx, pair, typee, buy, nonbase)

	select {}

}

func Bintic2(binpair binance.Symbol) {
	Orders2 := binance.NewClient().SubscribeTicker(binpair, time.Millisecond*100)
	for Order2 := range Orders2 {
		binask = Order2.AskPrice
		binbid = Order2.BidPrice
	}
}

func OrdersRequest(pair string, mx *sync.Mutex) {

	var r *http.Response
	baseURL := "https://payeer.com/api/trade/"
	endpoint := "orders"
	req := Request{
		Pair: pair,
	}

	rBody, err := json.Marshal(req)
	if err != nil {
		fmt.Println("rBody Marshal error #2 -> ", err)
	}
	// fmt.Println("rBody Marshal  -> ", bytes.NewBuffer(rBody))
	for {
		for {
			r, err = http.Post(baseURL+endpoint, "application/json", bytes.NewBuffer(rBody))
			if err == nil {
				break
			}
			fmt.Println("request error #2 -> ", err)
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println("responce error #2 -> ", err)
		}

		mx.Lock()
		json.Unmarshal(bodyBytes, &ords)
		mx.Unlock()
		//		fmt.Println(ords.Status)
		time.Sleep(250 * time.Millisecond)
	}
}

func MarketSell(pairr pairs, mx *sync.Mutex, pair string, typee string, sell bool, base currency) {
	var amountToSell float64
	if sell {
		for {
			time.Sleep(5 * time.Millisecond)
			mx.Lock()
			ss = 0
			for _, Bids := range ords.Pairs[pairr].Bids {

				p, err := strconv.ParseFloat(Bids.Price, 64)
				if err != nil {
					fmt.Println("p error -> ", err)
				}
				b, err := strconv.ParseFloat(binask, 64)
				if err != nil {
					fmt.Println("b error -> ", err)
				}
				c, err := strconv.ParseFloat(Bids.Amount, 64)
				if err != nil {
					fmt.Println("c error -> ", err)
				}

				if p/b > 1.01 {
					ss += c
				} else {
					break
				}
			}
			//fmt.Println("ss -> ", ss)
			amountToSell = min(Balas.Balance[base].Available, ss)
			mx.Unlock()

			//fmt.Println("amount avaliable for sell ", amountToSell)

			if amountToSell >= 0.0001 { // 													задать минимальный эмаунт для пары

				PostOrder(pair, typee, "sell", strconv.FormatFloat(amountToSell, 'f', 8, 32), "", "")
				time.Sleep(500 * time.Millisecond)
				BalanceRequest()
			}
		}
	}
}
func MarketBuy(pairr pairs, mx *sync.Mutex, pair string, typee string, buy bool, nonbase currency) {
	var valueToBuy float64
	if buy {
		for {
			time.Sleep(5 * time.Millisecond)
			mx.Lock()
			ss = 0
			for _, Asks := range ords.Pairs[pairr].Asks {

				p, err := strconv.ParseFloat(Asks.Price, 64)
				if err != nil {
					fmt.Println("p error -> ", err)
				}
				b, err := strconv.ParseFloat(binbid, 64)
				if err != nil {
					fmt.Println("b error -> ", err)
				}
				c, err := strconv.ParseFloat(Asks.Value, 64)
				if err != nil {
					fmt.Println("c error -> ", err)
				}

				if p/b < 0.99 {
					ss += c
				} else {
					break
				}
			}
			//fmt.Println("ss -> ", ss)
			valueToBuy = min(Balas.Balance[nonbase].Available, ss)
			mx.Unlock()

			//fmt.Println("value avaliable for buy ", valueToBuy)

			if valueToBuy >= 5 { // 															задать минимальный эмаунт для пары
				PostOrder(pair, typee, "buy", "", strconv.FormatFloat(valueToBuy, 'f', 8, 32), "")
				time.Sleep(500 * time.Millisecond)
				BalanceRequest()
			}
		}
	}
}

func BalanceRequest() {
	var resp *http.Response
	baseURL := "https://payeer.com/api/trade/"
	endpoint := "account"
	req := Request{
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
		time.Sleep(500 * time.Millisecond)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed to read resp.Body #1 -> ", err)
	}

	json.Unmarshal(bodyBytes, &Balas)

	//fmt.Println(&Balas)
}

func PostOrder(pair string, typee string, action string, amount string, value string, price string) {
	var req Request

	baseURL := "https://payeer.com/api/trade/"
	endpoint := "order_create"

	req = Request{
		Pair:      pair,
		Type:      typee,
		Action:    action,
		Amount:    amount,
		Value:     value,
		Price:     price,
		Timestamp: time.Now().UnixMilli(),
	}

	rBody, err := json.Marshal(req)
	if err != nil {
		fmt.Println("rBody Marshal error #3 -> ", err)
	}

	r, err := http.NewRequest(http.MethodPost, (baseURL + endpoint), bytes.NewBuffer(rBody))
	if err != nil {
		fmt.Println("http.NewRequest error #3 -> ", err)
	}

	client := http.Client{
		Timeout: 10 * time.Hour,
	}

	secret := "plIzgsI8akwDumrU"
	data := endpoint + string(rBody)

	h := hmac.New(sha256.New, []byte(secret)) // Create a new HMAC by defining the hash type and the key (as byte array)
	h.Write([]byte(data))                     // Write Data to it
	sha := hex.EncodeToString(h.Sum(nil))     // Get result and encode as hexadecimal string

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("API-ID", "3dc19f11-b03f-4d38-be50-c33a2585e12d")
	r.Header.Set("API-SIGN", sha)

	resp, err := client.Do(r)
	if err != nil {
		fmt.Println("failed to create request #3 -> ", err)
	}

	if typee == "limit" {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("failed to read resp.Body #1 -> ", err)
		}

		json.Unmarshal(bodyBytes, &LimitOrdId)
	}
}

func CancelLimitOrder(order_id int) {
	baseURL := "https://payeer.com/api/trade/"
	endpoint := "order_cancel"
	req := Request{
		Order_id:  order_id,
		Timestamp: time.Now().UnixMilli(),
	}

	rBody, err := json.Marshal(req)
	if err != nil {
		fmt.Println("rBody Marshal error #5 -> ", err)
	}

	r, err := http.NewRequest(http.MethodPost, (baseURL + endpoint), bytes.NewBuffer(rBody))
	if err != nil {
		fmt.Println("http.NewRequest error #5 -> ", err)
	}

	client := http.Client{
		Timeout: 10 * time.Hour,
	}

	secret := "plIzgsI8akwDumrU"
	data := endpoint + string(rBody)

	h := hmac.New(sha256.New, []byte(secret)) // Create a new HMAC by defining the hash type and the key (as byte array)
	h.Write([]byte(data))                     // Write Data to it
	sha := hex.EncodeToString(h.Sum(nil))     // Get result and encode as hexadecimal string

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("API-ID", "3dc19f11-b03f-4d38-be50-c33a2585e12d")
	r.Header.Set("API-SIGN", sha)

	resp, err := client.Do(r)
	if err != nil {
		fmt.Println("failed to create request #5 -> ", err)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed to read resp.Body #5 -> ", err)
	}
	//fmt.Println(string(bodyBytes))
	json.Unmarshal(bodyBytes, &CancelOrd)

	fmt.Println(CancelOrd)

}

func LimitBuy(pairr pairs, pair string, mx *sync.Mutex, percentOfBalance decimal.Decimal, nonbase currency, EVtoLimitBuy decimal.Decimal) {
	buyorders = 0
	var pricebuy decimal.Decimal
	var pricecansel decimal.Decimal
	var id int

	for {
		if buyorders == 0 {
			time.Sleep(5 * time.Millisecond)

			pricebuy, _ = decimal.NewFromString(binbid)

			pricebuy = pricebuy.Mul(EVtoLimitBuy)
			mx.Lock()
			for _, Bids := range ords.Pairs[pairr].Bids {

				p, err := decimal.NewFromString(Bids.Price)
				if err != nil {
					fmt.Println("p error -> ", err)
				}
				if p.LessThanOrEqual(pricebuy) {
					pricebuy = p.Add(decimal.NewFromFloat(0.01))
					break
				}
			}
			//	fmt.Println(Balas.Balance["ETH_USDT"].Available)
			//сколько денег потрачу /  цена по которой я куплю  = получаю количество баз актива
			a := percentOfBalance.Mul(decimal.NewFromFloat(Balas.Balance[nonbase].Available)) // //сколько денег потрачу

			PostOrder(pair, "limit", "buy", a.Div(pricebuy).String(), "", pricebuy.String()) // пост ордер в лимитке strconv.FormatFloat(pricebuy, 'f', 2, 32)
			id = LimitOrdId.Order_id                                                         // int

			mx.Unlock()

			buyorders = 1
			time.Sleep(1000 * time.Millisecond)
			BalanceRequest()
			time.Sleep(1 * time.Minute)
		}

		if buyorders == 1 {
			time.Sleep(5 * time.Millisecond)

			pricecansel, _ = decimal.NewFromString(binbid)

			pricecansel = pricecansel.Mul(EVtoLimitBuy)
			mx.Lock()
			for _, Bids := range ords.Pairs[pairr].Bids {

				p, err := decimal.NewFromString(Bids.Price)
				if err != nil {
					fmt.Println("p error -> ", err)
				}
				if p.LessThanOrEqual(pricecansel) {
					pricecansel = p
					break
				}
			}
			mx.Unlock()
			//			fmt.Println(pricebuy, pricecansel)
			if !pricecansel.Equal(pricebuy) {
				//				fmt.Println("тут уже не равны")
				for {
					CancelLimitOrder(id)
					if CancelOrd.Errror.ErrrorCode == ACCESS_DENIED || CancelOrd.Success {
						break
					}
					time.Sleep(500 * time.Millisecond)
				}

				time.Sleep(250 * time.Millisecond)
				buyorders = 0
			}
		}
	}
}

func LimitSell(pairr pairs, mx *sync.Mutex) {
	sellorders = 0
	var pricesell float64
	var pricecansel float64
	var id int // 																						нахрена оно?
	for {
		if sellorders == 0 {
			time.Sleep(5 * time.Millisecond)

			pricesell, _ = strconv.ParseFloat(binbid, 64)

			pricesell *= 1.002 // 												% для лимитки
			mx.Lock()
			for _, Bids := range ords.Pairs[pairr].Bids {

				p, err := strconv.ParseFloat(Bids.Price, 64)
				if err != nil {
					fmt.Println("p error -> ", err)
				}
				if p >= pricesell {
					pricesell = p - 0.01
					break
				}
			}

			PostOrder("ETH_USDT", "limit", "sell", "0.01435187", "", strconv.FormatFloat(pricesell, 'f', 2, 32)) // пост ордер в лимитке
			id = LimitOrdId.Order_id
			mx.Unlock()

			sellorders = 1
			time.Sleep(1000 * time.Millisecond)
			BalanceRequest()

		}
		time.Sleep(1 * time.Minute)
		if sellorders == 1 {
			time.Sleep(5 * time.Millisecond)

			pricecansel, _ = strconv.ParseFloat(binbid, 64)

			pricecansel *= 1.002 // 												% для лимитки
			mx.Lock()
			for _, Bids := range ords.Pairs[pairr].Bids {

				p, err := strconv.ParseFloat(Bids.Price, 64)
				if err != nil {
					fmt.Println("p error -> ", err)
				}
				if p >= pricecansel {
					pricecansel = p - 0.01
					break
				}
			}
			mx.Unlock()

			if pricecansel != pricesell {
				CancelLimitOrder(id)
			}
			time.Sleep(250 * time.Millisecond)
			sellorders = 0
		}
	}
}
