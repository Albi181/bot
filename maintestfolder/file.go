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
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

var InfoForPairs PairsInfo

type PairsInfo struct {
	Status bool                  `json:"success"`
	Pairs  map[pairs]Information `json:"pairs"`
}

type Information struct {
	MinAmount float64 `json:"min_amount"`
	MinValue  float64 `json:"min_value"`
}

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

var binask string // цена бинанс аск (стримится функцией Bintic2)
var binbid string
var ords OrdersResponce

var buyorders int64
var sellorders int64

func main() {
	InfoRequest()

	// var buy, sell bool
	var mx sync.Mutex

	const binpair binance.Symbol = "ETHUSDT"
	pair := "ETH_USDT"
	const pairr pairs = "ETH_USDT"
	const base currency = "ETH"
	// const nonbase currency = "USDT"
	// buy = true
	// sell = true

	var percentOfBalance decimal.Decimal = decimal.NewFromFloat(0.05) // процент от баланса идущий на лимит ордер
	//	var evToLimitBuy decimal.Decimal = decimal.NewFromFloat(0.99)    	 // процент дельты от цены бинанса для лимит покупки
	var evToLimitSell decimal.Decimal = decimal.NewFromFloat(1.05) // процент дельты от цены бинанса для лимит продажи

	// var evToMarketBuy decimal.Decimal = decimal.NewFromFloat(0.98)  // процент дельты от цены бинанса для маркет покупки
	// var evToMarketSell decimal.Decimal = decimal.NewFromFloat(1.02) // процент дельты от цены бинанса для маркет продажи

	BalanceRequest()

	go Bintic2(binpair)
	go OrdersRequest(pair, &mx)

	time.Sleep(10 * time.Second)

	go LimitSell(pairr, pair, &mx, percentOfBalance, base, evToLimitSell)
	// go LimitBuy(pairr, pair, &mx, percentOfBalance, nonbase, evToLimitBuy)

	// go MarketBuy(pairr, &mx, pair, buy, nonbase, evToMarketBuy)
	// go MarketSell(pairr, &mx, pair, sell, base, evToMarketSell)

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

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed to read resp.Body #1 -> ", err)
	}

	fmt.Println("bodyBytes ->", string(bodyBytes))

	if typee == "limit" {
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

func MarketBuy(pairr pairs, mx *sync.Mutex, pair string, buy bool, nonbase currency, evToMarketBuy decimal.Decimal) {

	if buy {
		for {
			time.Sleep(5 * time.Millisecond)

			ss := decimal.NewFromFloat(0) // sum of Asks.Values  (велью ту бай)
			dd := decimal.NewFromFloat(0) // dd sum of Asks.Amounts (эмаунт ту бай)

			valueFUCK := decimal.NewFromFloat(Balas.Balance[nonbase].Available)
			mx.Lock()
			for _, Asks := range ords.Pairs[pairr].Asks {
				p, err := decimal.NewFromString(Asks.Price)
				if err != nil {
					fmt.Println("p error -> ", err)
				}
				b, err := decimal.NewFromString(binbid)
				if err != nil {
					fmt.Println("b error -> ", err)
				}
				c, err := decimal.NewFromString(Asks.Value)
				if err != nil {
					fmt.Println("c error -> ", err)
				}
				d, err := decimal.NewFromString(Asks.Amount)
				if err != nil {
					fmt.Println("c error -> ", err)
				}

				if p.Div(b).LessThan(evToMarketBuy) { //		ев для маркет бая

					if valueFUCK.LessThanOrEqual(c) {
						ss = ss.Add(valueFUCK)
						dd = dd.Add(valueFUCK.Div(p))
						break
					} else {
						ss = ss.Add(c)
						dd = dd.Add(d)
						valueFUCK = valueFUCK.Sub(c)
					}
				}
			}
			mx.Unlock()
			//	fmt.Println("sum of Asks.Values ->", ss, "sum of Asks.Amounts ->", dd)

			if ss.GreaterThanOrEqual(decimal.NewFromFloat(InfoForPairs.Pairs[pairr].MinValue)) && dd.GreaterThanOrEqual(decimal.NewFromFloat(InfoForPairs.Pairs[pairr].MinAmount)) {
				PostOrder(pair, "market", "buy", "", ss.String(), "")
				time.Sleep(500 * time.Millisecond)
				mx.Lock()
				BalanceRequest()
				mx.Unlock()
			}
		}
	}
}

func MarketSell(pairr pairs, mx *sync.Mutex, pair string, sell bool, base currency, evToMarketSell decimal.Decimal) {

	if sell {
		for {
			time.Sleep(5 * time.Millisecond)

			ss := decimal.NewFromFloat(0) // sum of Bids.Values  (велью ту селл)
			dd := decimal.NewFromFloat(0) // dd sum of Bids.Amounts (эмаунт ту селл)

			amountFUCK := decimal.NewFromFloat(300)
			mx.Lock()
			for _, Bids := range ords.Pairs[pairr].Bids {
				p, err := decimal.NewFromString(Bids.Price)
				if err != nil {
					fmt.Println("p error -> ", err)
				}
				b, err := decimal.NewFromString(binask)
				if err != nil {
					fmt.Println("b error -> ", err)
				}
				c, err := decimal.NewFromString(Bids.Value)
				if err != nil {
					fmt.Println("c error -> ", err)
				}
				d, err := decimal.NewFromString(Bids.Amount)
				if err != nil {
					fmt.Println("c error -> ", err)
				}

				if p.Div(b).GreaterThan(evToMarketSell) {

					if amountFUCK.LessThanOrEqual(d) {
						dd = dd.Add(amountFUCK)
						ss = ss.Add(p.Mul(amountFUCK))
						break
					} else {
						dd = dd.Add(d)
						ss = ss.Add(c)
						amountFUCK = amountFUCK.Sub(d)
					}
				}
			}
			mx.Unlock()
			//	fmt.Println("sum of Bids.Values ->", ss, "sum of Bids.Amounts ->", dd)
			if ss.GreaterThanOrEqual(decimal.NewFromFloat(InfoForPairs.Pairs[pairr].MinValue)) && dd.GreaterThanOrEqual(decimal.NewFromFloat(InfoForPairs.Pairs[pairr].MinAmount)) {
				PostOrder(pair, "market", "sell", dd.String(), "", "")
				time.Sleep(500 * time.Millisecond)
				mx.Lock()
				BalanceRequest()
				mx.Unlock()
			}
		}
	}
}

func LimitBuy(pairr pairs, pair string, mx *sync.Mutex, percentOfBalance decimal.Decimal, nonbase currency, evToLimitBuy decimal.Decimal) {
	buyorders = 0
	var pricebuy decimal.Decimal
	var pricecansel decimal.Decimal
	var id int

	for {
		if buyorders == 0 {
			time.Sleep(5 * time.Millisecond)

			pricebuy, _ = decimal.NewFromString(binbid)

			pricebuy = pricebuy.Mul(evToLimitBuy)
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
			mx.Unlock()
			//	fmt.Println(Balas.Balance["ETH_USDT"].Available)
			//сколько денег потрачу /  цена по которой я куплю  = получаю количество баз актива
			a := percentOfBalance.Mul(decimal.NewFromFloat(Balas.Balance[nonbase].Total))                      // размер заявки
			if a.Div(pricebuy).GreaterThanOrEqual(decimal.NewFromFloat(InfoForPairs.Pairs[pairr].MinAmount)) { // 		больше минималки
				PostOrder(pair, "limit", "buy", a.Div(pricebuy).String(), "", pricebuy.String()) // пост ордер в лимитке
				id = LimitOrdId.Order_id                                                         // int
			} else {
				continue
			}

			buyorders = 1
			time.Sleep(1000 * time.Millisecond)
			BalanceRequest()
			time.Sleep(1 * time.Minute)
		}

		if buyorders == 1 {
			time.Sleep(5 * time.Millisecond)

			pricecansel, _ = decimal.NewFromString(binbid)

			pricecansel = pricecansel.Mul(evToLimitBuy)
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

			if !pricecansel.Equal(pricebuy) {
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

func LimitSell(pairr pairs, pair string, mx *sync.Mutex, percentOfBalance decimal.Decimal, base currency, evToLimitSell decimal.Decimal) {
	sellorders = 0
	var pricesell decimal.Decimal
	var pricecansel decimal.Decimal
	var id int
	for {
		if sellorders == 0 {
			time.Sleep(5 * time.Millisecond)

			pricesell, _ = decimal.NewFromString(binask)

			pricesell = pricesell.Mul(evToLimitSell) // 												% для лимитки
			mx.Lock()
			for _, Asks := range ords.Pairs[pairr].Asks {

				p, err := decimal.NewFromString(Asks.Price)
				if err != nil {
					fmt.Println("p error -> ", err)
				}
				if p.GreaterThanOrEqual(pricesell) {
					pricesell = p.Sub(decimal.NewFromFloat(0.01))
					break
				}
			}
			mx.Unlock()

			a := percentOfBalance.Mul(decimal.NewFromFloat(Balas.Balance[base].Total))

			if a.GreaterThanOrEqual(decimal.NewFromFloat(InfoForPairs.Pairs[pairr].MinAmount)) { // 		больше минималки
				PostOrder(pair, "limit", "buy", a.String(), "", pricesell.String()) // пост ордер в лимитке
				id = LimitOrdId.Order_id                                            // int
			} else {
				continue
			}

			sellorders = 1
			time.Sleep(1000 * time.Millisecond)
			BalanceRequest()
			time.Sleep(1 * time.Minute)
		}

		if sellorders == 1 {
			time.Sleep(5 * time.Millisecond)

			pricecansel, _ = decimal.NewFromString(binask)

			pricecansel = pricecansel.Mul(evToLimitSell)
			mx.Lock()
			for _, Asks := range ords.Pairs[pairr].Asks {

				p, err := decimal.NewFromString(Asks.Price)
				if err != nil {
					fmt.Println("p error -> ", err)
				}
				if p.GreaterThanOrEqual(pricecansel) {
					pricecansel = p
					break
				}
			}
			mx.Unlock()

			if !pricecansel.Equal(pricesell) {
				for {
					CancelLimitOrder(id)
					if CancelOrd.Errror.ErrrorCode == ACCESS_DENIED || CancelOrd.Success {
						break
					}
					time.Sleep(500 * time.Millisecond)
				}

				time.Sleep(250 * time.Millisecond)
				sellorders = 0
			}
		}
	}
}

func InfoRequest() {

	var r *http.Response
	var err error
	baseURL := "https://payeer.com/api/trade/"
	endpoint := "info"

	for {
		r, err = http.Get(baseURL + endpoint)
		if err == nil {
			break
		}
		fmt.Println("request error #2 -> ", err)
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("responce error #2 -> ", err)
	}
	//	fmt.Println(string(bodyBytes))
	json.Unmarshal(bodyBytes, &InfoForPairs)

	//	fmt.Println(InfoForPairs)
}
