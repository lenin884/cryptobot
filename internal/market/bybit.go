package market

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

const (
	baseURL = "https://api.bybit.com"
)

type Bybit struct {
	apiKey    string
	apiSecret string
}

func NewBybit(apiKey, apiSecret string) *Bybit {
	return &Bybit{
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
}

func (b *Bybit) GetAssets() (map[string]float64, error) {
	url := fmt.Sprintf("%s/v2/private/wallet/balance", baseURL)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	assets := make(map[string]float64)
	for asset, data := range result["result"].(map[string]interface{}) {
		assets[asset] = data.(map[string]interface{})["available_balance"].(float64)
	}

	return assets, nil
}

func (b *Bybit) GetTradeHistory() ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v2/private/execution/list", baseURL)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	trades := result["result"].([]interface{})
	tradeHistory := make([]map[string]interface{}, len(trades))
	for i, trade := range trades {
		tradeHistory[i] = trade.(map[string]interface{})
	}

	return tradeHistory, nil
}

func (b *Bybit) GetCurrentPrice(symbol string) (float64, error) {
	url := fmt.Sprintf("%s/v2/public/tickers?symbol=%s", baseURL, symbol)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	price := result["result"].([]interface{})[0].(map[string]interface{})["last_price"].(string)
	return strconv.ParseFloat(price, 64)
}