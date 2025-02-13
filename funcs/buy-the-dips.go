package funcs

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/howlrs/buy-the-dips/utils"
	"github.com/labstack/echo/v4"
)

// BuyTheDips は指定値以上の下落を検知した場合にエントリーする関数です
func BuyTheDips(c echo.Context) error {
	targetNumber := 2
	q, err := strconv.Atoi(c.QueryParam("targetN"))
	if err == nil {
		targetNumber = q
	}

	// ロジックに必要な引数の取得
	config, err := NewArgsForLogic()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": fmt.Sprintf("failed to get args, %s", err.Error())})
	}

	ctx := context.Background()
	// 既存の注文のキャンセル
	// 既存の注文の判別は固有のOrderLinkIDを使用
	if err := config.Cancel(ctx); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": fmt.Sprintf("failed to cancel, %s", err.Error())})
	}

	// 必要なデータの取得
	ohlcv, err := config.GetOHLCV(ctx, targetNumber)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": fmt.Sprintf("failed to get ohlcv, %s", err.Error())})
	}

	// ポジションの取得
	positions, err := config.GetPositions(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": fmt.Sprintf("failed to get positions, %s", err.Error())})
	}
	// ポジションを持っている場合の処理
	if tokenSize, isThere := config.HasPosition(positions); isThere {
		// ティッカーの取得
		ticker, err := config.GetTicker(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": fmt.Sprintf("failed to get ticker, %s", err.Error())})
		}

		// エントリー済みのポジションを決済する処理
		if err := config.Exit(ctx, tokenSize, ticker.BestAsk); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": fmt.Sprintf("failed to exit, %s", err.Error())})
		} else {
			// 当else処理をコメントアウトすると、ポジションを決済時にも価格変動検出を行います。
			// 決済後再度エントリーを行う際には、価格変動検出を行います。
			return c.JSON(http.StatusOK, echo.Map{"message": "exit the position"})
		}
	}

	// 下落率の計算
	if !config.IsDip(targetNumber, ohlcv) {
		return c.JSON(http.StatusNoContent, echo.Map{"message": fmt.Sprintf("not dip")})
	}

	// ティッカーの取得
	ticker, err := config.GetTicker(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": fmt.Sprintf("failed to get ticker, %s", err.Error())})
	}
	// 指定値以上の下落を検知した場合の処理
	if err := config.Entry(ctx, ticker.BestBid); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": fmt.Sprintf("failed to entry, %s", err.Error())})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "buy the dips"})
}

// HasPosition はポジションを持っているかどうかを判定する関数です
func (c *ArgsForLogic) HasPosition(positions []map[string]any) (float64, bool) {
	if len(positions) < 1 {
		return 0, false
	}

	// aggregateの計算
	var aggregate float64
	for _, p := range positions {
		// allways positive number
		size, ok := p["size"].(float64)
		if !ok {
			continue
		}

		aggregate += size
	}

	utils.TestLog(c.IsTest, "%s aggregate position size: %f", c.Symbol, aggregate)

	return aggregate, false
}

// IsDip は指定値以上の下落を検知する関数です
func (c *ArgsForLogic) IsDip(targetNumber int, ohlcv []Ohlcv) bool {
	if len(ohlcv) != targetNumber {
		return false
	}

	// string to float64
	numbers := make([]float64, targetNumber)
	for i, o := range ohlcv {
		prev, err := strconv.ParseFloat(o.Close, 64)
		if err != nil {
			return false
		}
		numbers[i] = prev
	}

	dipRatio := utils.DipsRatio(numbers)

	if dipRatio > 0 {
		utils.TestLog(c.IsTest, "raise ratio: %f", dipRatio)
	} else if dipRatio < c.DipsRatio {
		// 下落率が指定値（negative）以下の場合はエントリーする
		return true
	}

	return false
}
