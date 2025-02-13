package funcs

import (
	"fmt"
	"os"
	"strconv"

	"github.com/howlrs/buy-the-dips/utils"
	bybit "github.com/wuhewuhe/bybit.go.api"
)

type ArgsForLogic struct {
	IsTest bool

	bybitClient *bybit.Client

	Category  string
	Symbol    string
	IsMarket  bool
	IsExit    bool
	DipsRatio float64
	Size      float64
}

// NewArgsForLogic はロジックに必要な引数を取得する関数です
func NewArgsForLogic() (*ArgsForLogic, error) {
	isTest := os.Getenv("IS_TEST") == "true"

	// ターゲットマーケット
	category := os.Getenv("CATEGORY")
	if category == "" {
		return nil, fmt.Errorf("CATEGORY is required")
	}

	// 対象の通貨ペア
	symbol := os.Getenv("SYMBOL")
	if symbol == "" {
		return nil, fmt.Errorf("SYMBOL is required")
	}
	// マーケット注文かどうか
	// Limit注文の場合はGetTickerを消費し、best_bidに対しポストオンリー注文を行う
	isMarket := os.Getenv("IS_MARKET") == "true"
	// 決済注文を行うかどうか
	// falseで行わない設定の場合は、押し目買い保有となる。
	isExit := os.Getenv("IS_EXIT") == "true"
	// ターゲットの下落率
	temp := os.Getenv("DIPS_RATIO")
	ratio, err := strconv.ParseFloat(temp, 64)
	if err != nil {
		return nil, err
	}
	// 1回のエントリーでの購入金額
	temp = os.Getenv("SIZEUSD")
	size, err := strconv.ParseFloat(temp, 64)
	if err != nil {
		return nil, err
	}

	keys := utils.Split(os.Getenv("BYBIT"))
	if keys[0] == "" || keys[1] == "" {
		return nil, fmt.Errorf("BYBIT_API_KEY and BYBIT_API_SECRET are required")
	}

	client := bybit.NewBybitHttpClient(keys[0], keys[1])
	if client == nil {
		return nil, err
	}

	return &ArgsForLogic{
		IsTest: isTest,

		bybitClient: client,

		Category:  category,
		Symbol:    symbol,
		IsMarket:  isMarket,
		IsExit:    isExit,
		DipsRatio: ratio,
		Size:      size,
	}, nil
}
