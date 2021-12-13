package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/labstack/echo/v4"
)

func userAuth(c echo.Context) (err error) {
	u := new(User)
	if err = c.Bind(u); err != nil {
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	userCollection := collection("users")

	cur, err := userCollection.Find(ctx, bson.M{"name": u.Name})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Ошибка: пользователь не найден")
	}
	if cur.Next(ctx) {
		var existingUser User
		err := cur.Decode(&existingUser)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Ошибка: пользователь неверный")
		}
		if existingUser.Password == u.Password {
			return c.JSON(http.StatusOK, bson.M{
				"tradingToken": sign(existingUser.Name),
			})
		} else {
			return c.JSON(http.StatusForbidden, "Ошибка: неверный пароль")
		}
	} else {
		u.Id = rand.Uint32()
		u.Money = 10000
		u.Portfolio = make(map[string]uint32)

		if len(u.Name) < 10 || len(u.Password) < 10 {
			return c.JSON(http.StatusBadRequest, "Ошибка: логин и пароль должны быть не короче 10 символов")
		}

		ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
		_, err := collection("users").InsertOne(ctx, u)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Ошибка: пользователь не создан")
		}

		return c.JSON(http.StatusOK, bson.M{
			"tradingToken": sign(u.Name),
		})
	}
}

func userInfo(c echo.Context) (err error) {
	a := new(Auth)
	if err = c.Bind(a); err != nil {
		return
	}
	if !check(a.Name, a.TradingToken) {
		return c.JSON(http.StatusForbidden, "Неверный HMAC")
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	userCollection := collection("users")

	var user User
	err = userCollection.FindOne(ctx, bson.M{"name": a.Name}).Decode(&user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Ошибка: пользователь неверный")
	}

	return c.JSON(http.StatusOK, user)
}

func privateSection(c echo.Context) (err error) {
	a := new(Auth)
	if err = c.Bind(a); err != nil {
		return
	}
	if !check(a.Name, a.TradingToken) {
		return c.JSON(http.StatusForbidden, "Неверный HMAC")
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	userCollection := collection("users")

	var user User
	err = userCollection.FindOne(ctx, bson.M{"name": a.Name}).Decode(&user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Ошибка: пользователь неверный")
	}

	if _, ok := user.Portfolio["BIEN"]; ok {
		if user.Portfolio["BIEN"] > 0 {
			return c.Blob(http.StatusOK, "text/html", MakeRequest(fmt.Sprintf("dividends/%s", a.TradingToken)))
		}
	}

	return c.JSON(http.StatusForbidden, user)
}

func sign(data string) string {
	h := hmac.New(sha256.New, []byte("secret"))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func check(name string, messageMAC string) bool {
	h := hmac.New(sha256.New, []byte("secret"))
	h.Write([]byte(name))
	expectedMAC := hex.EncodeToString(h.Sum(nil))
	return messageMAC == expectedMAC
}
