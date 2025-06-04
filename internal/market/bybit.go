package market

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://api.bybit.com"

type Bybit struct {
	apiKey    string
	apiSecret string
	client    *http.Client
}

// Trade represents a trade from Bybit history.
type Trade struct {
	Symbol    string
	Category  string
	Side      string
	Qty       float64
	Price     float64
	Timestamp int64
}

func NewBybit(apiKey, apiSecret string) *Bybit {
	return &Bybit{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (b *Bybit) signedRequest(ctx context.Context, method, endpoint string, params url.Values) (*http.Response, error) {
	recvWindow := "5000"
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	query := params.Encode()
	var payload string
	if method == http.MethodPost {
		payload = query
	} else {
		payload = query
	}
	signBase := ts + b.apiKey + recvWindow + payload
	mac := hmac.New(sha256.New, []byte(b.apiSecret))
	mac.Write([]byte(signBase))
	sign := hex.EncodeToString(mac.Sum(nil))

	reqURL := fmt.Sprintf("%s%s", baseURL, endpoint)
	if method == http.MethodGet && query != "" {
		reqURL += "?" + query
	}

	var body io.Reader
	if method == http.MethodPost {
		body = strings.NewReader(query)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-BAPI-API-KEY", b.apiKey)
	req.Header.Set("X-BAPI-SIGN", sign)
	req.Header.Set("X-BAPI-TIMESTAMP", ts)
	req.Header.Set("X-BAPI-RECV-WINDOW", recvWindow)
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return b.client.Do(req)
}

func (b *Bybit) GetAssets(ctx context.Context) (map[string]float64, error) {
	params := url.Values{}
	params.Set("accountType", "UNIFIED")
	resp, err := b.signedRequest(ctx, http.MethodGet, "/v5/asset/transfer/query-account-coins-balance", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result struct {
			List []struct {
				Coin string `json:"coin"`
				Free string `json:"transferBalance"`
			} `json:"list"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	assets := make(map[string]float64)
	for _, a := range result.Result.List {
		qty, _ := strconv.ParseFloat(a.Free, 64)
		assets[a.Coin] = qty
	}
	return assets, nil
}

func (b *Bybit) GetTradeHistory(ctx context.Context, category string) ([]Trade, error) {
	params := url.Values{}
	params.Set("category", category)
	params.Set("limit", "50")
	resp, err := b.signedRequest(ctx, http.MethodGet, "/v5/execution/list", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result struct {
			List []struct {
				Symbol string `json:"symbol"`
				Side   string `json:"side"`
				Qty    string `json:"qty"`
				Price  string `json:"price"`
				Time   int64  `json:"execTime"`
			} `json:"list"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	var trades []Trade
	for _, t := range result.Result.List {
		qty, _ := strconv.ParseFloat(t.Qty, 64)
		price, _ := strconv.ParseFloat(t.Price, 64)
		trades = append(trades, Trade{
			Symbol:    t.Symbol,
			Category:  category,
			Side:      strings.Title(strings.ToLower(t.Side)),
			Qty:       qty,
			Price:     price,
			Timestamp: t.Time,
		})
	}
	return trades, nil
}

func (b *Bybit) GetCurrentPrice(ctx context.Context, symbol string) (float64, error) {
	endpoint := fmt.Sprintf("%s/v5/market/tickers", baseURL)
	params := url.Values{}
	params.Set("category", "spot")
	params.Set("symbol", symbol)
	reqURL := endpoint + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Result struct {
			List []struct {
				LastPrice string `json:"lastPrice"`
			} `json:"list"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	if len(result.Result.List) == 0 {
		return 0, fmt.Errorf("no price")
	}
	return strconv.ParseFloat(result.Result.List[0].LastPrice, 64)
}
