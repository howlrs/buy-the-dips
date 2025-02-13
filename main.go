package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/howlrs/buy-the-dips/funcs"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/joho/godotenv"
)

var (
	PORT   string
	isTest bool
)

func init() {
	// 環境別の処理
	if runtime.GOOS == "linux" {
		godotenv.Load(".env.production")
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		PORT = fmt.Sprintf("0.0.0.0:%s", os.Getenv("PORT"))
		log.Debug().Msgf("Linuxでの処理, PORT: %s", PORT)

		isTest = os.Getenv("IS_TEST") == "true"
		if isTest {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			log.Debug().Msg("テストモードで起動します")
		}
	} else {
		godotenv.Load(".env.local")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		PORT = fmt.Sprintf("localhost:%s", os.Getenv("PORT"))
		log.Debug().Msgf("その他のOSでの処理, PORT: %s", PORT)
		isTest = true
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.CORS())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"message": "OK"})
	})

	e.GET("/buy-the-dips", funcs.BuyTheDips)

	log.Fatal().Err(e.Start(PORT))
}
