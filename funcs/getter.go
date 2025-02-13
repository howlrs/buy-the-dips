package funcs

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/howlrs/buy-the-dips/utils"
)

type Ohlcv struct {
	Start    int64
	Open     string
	High     string
	Low      string
	Close    string
	Volume   string
	Turnover string
}

// GetOHLCV はOHLCVデータを取得する関数です
func (c *ArgsForLogic) GetOHLCV(ctx context.Context, targetNumber int) ([]Ohlcv, error) {
	start := time.Now()
	defer func() {
		utils.TestLog(c.IsTest, "get ohlcv elapsed time: %v", time.Since(start))
	}()

	res, err := c.bybitClient.NewMarketKlineService(
		"kline",
		c.Category,
		c.Symbol,
		"1",
	).Do(ctx)
	if err != nil {
		return nil, err
	}

	if res.RetCode != 0 {
		return nil, fmt.Errorf("failed to get ohlcv")
	}

	result := utils.GetInnerData(res.Result)
	if result == nil {
		return nil, fmt.Errorf("failed to get ohlcv")
	}

	ohlcv, ok := result["list"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("convert failed, changed to list")
	}

	// 1分足のデータを取得する
	// 配列0が新しく、降順となっているため、配列0,1を取得する
	ohlcvs := make([]Ohlcv, targetNumber)
	if len(ohlcv) < targetNumber {
		return nil, fmt.Errorf("failed to get ohlcv")
	}
	for i := 0; i < targetNumber; i++ {
		temp, ok := ohlcv[i].([]interface{})
		if !ok {
			return nil, fmt.Errorf("convert failed, changed to string")
		}

		if len(temp) < 7 {
			continue
		}

		ms, err := strconv.ParseInt(temp[0].(string), 10, 64)
		if err != nil {
			return nil, err
		}

		ohlcvs[i] = Ohlcv{
			Start:    ms,
			Open:     temp[1].(string),
			High:     temp[2].(string),
			Low:      temp[3].(string),
			Close:    temp[4].(string),
			Volume:   temp[5].(string),
			Turnover: temp[6].(string),
		}
	}

	if len(ohlcvs) < targetNumber {
		return nil, fmt.Errorf("failed to get ohlcv")
	}

	utils.TestLog(c.IsTest, "get ohlcv: %v", ohlcvs)

	if ohlcvs[0].Start < ohlcvs[targetNumber-1].Start {
		// 1分足のデータが降順になっていることを確認
		return nil, fmt.Errorf("failed to get ohlcv, not descending order")
	}

	return ohlcvs, nil
}

// GetPositions はポジションを取得する関数です
func (c *ArgsForLogic) GetPositions(ctx context.Context) ([]map[string]any, error) {
	start := time.Now()
	defer func() {
		utils.TestLog(c.IsTest, "get positions elapsed time: %v", time.Since(start))
	}()

	params := map[string]interface{}{
		"category": c.Category,
		"symbol":   c.Symbol,
	}
	res, err := c.bybitClient.NewPositionService(
		params,
	).GetPositionList(ctx)
	if err != nil {
		return nil, err
	}

	if res.RetCode != 0 {
		return nil, fmt.Errorf("failed to get positions")
	}

	pos := utils.GetInnerList(res.Result)
	if pos == nil {
		return nil, fmt.Errorf("failed to get positions")
	}

	// 1分+α以内のポジションを取得する
	var positions = make([]map[string]interface{}, 0)
	durationForTerm := time.Duration(1*time.Minute + 1*time.Second)
	term := time.Now().UTC().Add(-durationForTerm).UnixMilli()
	for _, p := range pos {
		timeint, err := strconv.ParseInt(p["createdTime"].(string), 10, 64)
		if err != nil {
			continue
		}
		if term < int64(timeint) {
			// 指定時間内の新規ポジションではないため、スキップ
			continue
		}

		positions = append(positions, p)
	}

	utils.TestLog(c.IsTest, "get positions: %v", positions)

	return positions, nil
}

type Ticker struct {
	Symbol     string
	Ltp        string
	IndexPrice string
	MarkPrice  string
	BestBid    string
	BestAsk    string
	Timestamp  int64
}

// GetTicker はティッカーを取得する関数です
func (c *ArgsForLogic) GetTicker(ctx context.Context) (*Ticker, error) {
	start := time.Now()
	defer func() {
		utils.TestLog(c.IsTest, "get ticker elapsed time: %v", time.Since(start))
	}()

	params := map[string]interface{}{
		"category": c.Category,
		"symbol":   c.Symbol,
	}
	res, err := c.bybitClient.NewMarketInfoService(
		params,
	).GetMarketTickers(ctx)
	if err != nil {
		return nil, err
	}

	if res.RetCode != 0 {
		return nil, fmt.Errorf("failed to get ticker")
	}

	term := time.Now().UTC().Add(-10 * time.Second).UnixMilli()
	if res.RetCode != 0 {
		return nil, fmt.Errorf("failed to get ticker")
	} else if res.Time < term {
		return nil, fmt.Errorf("failed to get ticker, caused by term over")
	}

	list := utils.GetInnerList(res.Result)
	if list == nil {
		return nil, fmt.Errorf("failed to get ticker")
	}

	for _, l := range list {
		if l["symbol"].(string) == c.Symbol {
			utils.TestLog(c.IsTest, "get ticker, %s bid: %s, ask: %s", c.Symbol, l["bid1Price"].(string), l["ask1Price"].(string))
			return &Ticker{
				Symbol:     l["symbol"].(string),
				Ltp:        l["lastPrice"].(string),
				IndexPrice: l["indexPrice"].(string),
				MarkPrice:  l["markPrice"].(string),
				BestBid:    l["bid1Price"].(string),
				BestAsk:    l["ask1Price"].(string),
				Timestamp:  res.Time,
			}, nil
		}
	}

	return nil, fmt.Errorf("failed to get ticker")
}
