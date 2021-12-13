package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/labstack/echo/v4"
)

type Auth struct {
	TradingToken string
	Name         string
}

type StockRequest struct {
	Auth
	Stock string
}

type BuyRequest struct {
	Auth
	Stock  string
	Amount uint32
}

type TransactionRequest struct {
	Auth
	TransactionId primitive.ObjectID
}

type TransactionResponse struct {
	Error    string
	Commited bool
}

func getStock(c echo.Context) (err error) {
	r := new(StockRequest)
	if err = c.Bind(r); err != nil {
		return
	}

	if r.Stock == "" {
		return c.JSONBlob(http.StatusOK, MakeRequest("stock"))
	} else {
		if _, ok := stock[r.Stock]; !ok {
			log.Printf("Stock not found: %s", r.Stock)
			return c.JSON(http.StatusForbidden, "Неверный тикер")
		}

		return c.JSONBlob(http.StatusOK, MakeRequest(fmt.Sprintf("stock/%s", r.Stock)))
	}
}

func buyStock(c echo.Context) (err error) {
	r := new(BuyRequest)
	if err = c.Bind(r); err != nil {
		return
	}
	if !check(r.Name, r.TradingToken) {
		return c.JSON(http.StatusForbidden, "Неверный HMAC")
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	id, err := collection("transactions").InsertOne(ctx, r)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Ошибка создания транзакции")
	}

	buyQueue <- BuyTransaction{
		Id:     id.InsertedID.(primitive.ObjectID),
		Name:   r.Name,
		Amount: r.Amount,
		Stock:  r.Stock,
	}

	return c.JSON(http.StatusOK, id)
}

func checkTransactionStatus(c echo.Context) (err error) {
	r := new(TransactionRequest)
	if err = c.Bind(r); err != nil {
		return
	}
	if !check(r.Name, r.TradingToken) {
		return c.JSON(http.StatusForbidden, "Неверный HMAC")
	}

	var transaction BuyTransaction
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	filter := bson.D{{"_id", r.TransactionId}}
	err = collection("transactions").FindOne(ctx, filter).Decode(&transaction)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Ошибка поиска транзакции")
	}

	return c.JSON(http.StatusOK, &TransactionResponse{
		transaction.Error,
		transaction.Commited,
	})
}

func sellStock(c echo.Context) (err error) {
	r := new(BuyRequest)
	if err = c.Bind(r); err != nil {
		return
	}
	if !check(r.Name, r.TradingToken) {
		return c.JSON(http.StatusForbidden, "Неверный HMAC")
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	id, err := collection("transactions").InsertOne(ctx, r)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Ошибка создания транзакции")
	}

	sellQueue <- SellTransaction{
		Id:     id.InsertedID.(primitive.ObjectID),
		Name:   r.Name,
		Amount: r.Amount,
		Stock:  r.Stock,
	}

	return c.JSON(http.StatusOK, id)
}
