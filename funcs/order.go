package funcs

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/howlrs/buy-the-dips/utils"
	"github.com/rs/zerolog/log"
)

// Entry はエントリーする関数です
func (c *ArgsForLogic) Entry(ctx context.Context, bestBid string) error {
	start := time.Now()
	defer func() {
		utils.TestLog(c.IsTest, "order entry elapsed time: %v", time.Since(start))
	}()

	if c.Size <= 0 {
		return fmt.Errorf("size is invalid")
	}

	tokenSize := fmt.Sprintf("%f", c.Size/utils.ToFloat64(bestBid))

	oLinkId := os.Getenv("ORDERLINKID")
	if oLinkId == "" {
		return fmt.Errorf("ORDERLINKID is required")
	}

	orderReq := c.bybitClient.NewPlaceOrderService(
		c.Category,
		c.Symbol,
		"Buy",
		utils.ToString(c.IsMarket),
		tokenSize,
	).Price(bestBid).OrderLinkId(oLinkId)
	// 指値の場合はPostOnlyを設定
	if !c.IsMarket {
		orderReq.TimeInForce("PostOnly")
	}

	if c.IsTest {
		log.Debug().Msgf("[test mode] entry order: %v", orderReq)
		return nil
	}
	res, err := orderReq.Do(ctx)
	if err != nil {
		return err
	}

	if res.RetCode != 0 {
		return fmt.Errorf("failed to entry, %w", err)
	}

	return nil
}

// Exit はエグジットする関数です
func (c *ArgsForLogic) Exit(ctx context.Context, tokenSize float64, bestAsk string) error {
	start := time.Now()
	defer func() {
		utils.TestLog(c.IsTest, "order exit elapsed time: %v", time.Since(start))
	}()

	if !c.IsExit {
		return nil
	} else if tokenSize == 0 {
		// ポジションを持っていない場合はエラーを返す
		// Why: Bybit systemでは、qty: 0を渡し、reduseOnlyをtrueにするオールクローズになるため
		// 上記を回避するためにエラーを返す
		return fmt.Errorf("position is not found")
	}

	// ポジションを決済する
	// 決済の場合はリデュースオンリーをtrueにする
	orderReq := c.bybitClient.NewPlaceOrderService(
		c.Category,
		c.Symbol,
		"Sell",
		utils.ToString(c.IsMarket),
		fmt.Sprintf("%f", tokenSize),
	).Price(bestAsk).ReduceOnly(true)
	// 指値の場合はPostOnlyを設定
	if c.IsMarket {
		orderReq.TimeInForce("PostOnly")
	}

	if c.IsTest {
		log.Debug().Msgf("[test mode] exit order: %v", orderReq)
		return nil
	}
	res, err := orderReq.Do(ctx)
	if err != nil {
		return err
	}

	if res.RetCode != 0 {
		return fmt.Errorf("failed to exit, %w", err)
	}

	return nil
}

func (c *ArgsForLogic) Cancel(ctx context.Context) error {
	start := time.Now()
	defer func() {
		utils.TestLog(c.IsTest, "order cancel elapsed time: %v", time.Since(start))
	}()

	oLinkId := os.Getenv("ORDERLINKID")
	if oLinkId == "" {
		return fmt.Errorf("ORDERLINKID is required")
	}

	params := map[string]interface{}{
		"category":    c.Category,
		"symbol":      c.Symbol,
		"orderLinkId": oLinkId,
	}

	orderReq := c.bybitClient.NewTradeService(
		params,
	)

	// if c.IsTest {
	// 	log.Debug().Msgf("[test mode] cancel order: %v", orderReq)
	// 	return nil
	// }
	res, err := orderReq.CancelOrder(ctx)
	if err != nil {
		log.Log().Msgf("failed to cancel, %+v", err)
		return nil
	}

	if res.RetCode == 110001 {
		utils.TestLog(c.IsTest, "order not exists, orderLinkId: %s, %+v", oLinkId, res)
		return nil
	} else if res.RetCode != 0 {
		utils.TestLog(c.IsTest, "failed to cancel, %+v", res)
		return fmt.Errorf("failed to cancel, %w", err)
	}

	return nil
}
