package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/mongo"
)

var db *mongo.Client

type priceInfo struct {
	I    string  `json:"-"`
	O    string  `json:"-"`
	L    string  `json:"-"`
	H    string  `json:"-"`
	T    float32 `json:"-"`
	Buy  float32 `json:"sell"`
	Sell float32 `json:"buy"`
}

var stock = make(map[string]priceInfo)

func main() {
	e := echo.New()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	if os.Getenv("LOG_REQUESTS") == "true" {
		e.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
			fmt.Fprintf(os.Stderr, "%s %s", c.Request().URL, reqBody)
		}))
	}
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.File("/", "./static/index.html")
	e.Static("/static", "static")

	e.POST("/auth", userAuth)
	e.POST("/stockinfo", getStock)
	e.POST("/buy", buyStock)
	e.POST("/sell", sellStock)
	e.POST("/userinfo", userInfo)
	e.POST("/status", checkTransactionStatus)
	e.POST("/dividends", privateSection)

	db = connect()

	err = json.Unmarshal(MakeRequest("stock"), &stock)
	if err != nil {
		log.Fatalf("%s", err)
	}

	for i := 1; i < 10; i++ {
		go processBuyTransactions()
	}
	go processSellTransactions()

	e.Logger.Fatal(e.Start(":80"))
}
