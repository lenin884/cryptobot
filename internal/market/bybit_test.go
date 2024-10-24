package market

import (
	"testing"
	"context"
	"fmt"

	bybit "github.com/wuhewuhe/bybit.go.api"
)

func TestMain(t *testing.T) {
	client := bybit.NewBybitHttpClient("API KEY", "SECRET KEY", bybit.WithBaseURL(bybit.MAINNET))
	params := map[string]interface{}{"category": "spot", "symbol": "BTCUSDT", "interval": "1"}
	orderResult, err := client.NewUtaBybitServiceWithParams(params).GetMarketKline(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(bybit.PrettyPrint(orderResult))
}
