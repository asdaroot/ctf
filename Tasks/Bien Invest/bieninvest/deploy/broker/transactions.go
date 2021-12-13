package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/bson"
)

type BuyTransaction struct {
	Id       primitive.ObjectID
	Name     string
	Amount   uint32
	Stock    string
	Error    string
	Commited bool
}

type SellTransaction struct {
	Id       primitive.ObjectID
	Name     string
	Amount   uint32
	Stock    string
	Error    string
	Commited bool
}

type PriceResponse struct {
	Bought uint32
	Sold   uint32
	Price  uint32
	Summ   uint32
	Token  uint32
}

var buyQueue = make(chan BuyTransaction, 1000)
var sellQueue = make(chan SellTransaction, 1000)

func processBuyTransactions() {
	for {
		tr := <-buyQueue

		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		userCollection := collection("users")
		var user User
		err := userCollection.FindOne(ctx, bson.M{"name": tr.Name}).Decode(&user)
		if err != nil {
			log.Printf("%s", err)
			continue
		}

		if _, ok := stock[tr.Stock]; !ok {
			log.Printf("Stock not found: %s", tr.Stock)
			decline(tr.Id, "Тикер не найден")
			continue
		}
		var price PriceResponse
		err = json.Unmarshal(MakeRequest(fmt.Sprintf("sell/%s/%d/", tr.Stock, tr.Amount)), &price)
		if err != nil {
			log.Printf("%s", err)
			decline(tr.Id, "Тикер закрыт")
			continue
		}

		if user.Money >= price.Summ {

			data := MakeRequest(fmt.Sprintf("commit/%d/", price.Token))
			if data == nil {
				log.Printf("%s", err)
				decline(tr.Id, "Биржа отклонила операцию")
				continue
			}

			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			filter := bson.D{{"_id", tr.Id}}
			update := bson.D{{"$set", bson.D{{"commited", true}}}}
			_, err = collection("transactions").UpdateOne(ctx, filter, update)
			if err != nil {
				log.Printf("UpdateTr: %s", err)
				continue
			}

			//if _, ok := user.Portfolio[tr.Stock]; !ok {
			//	user.Portfolio[tr.Stock] = tr.Amount
			//} else {
			//	user.Portfolio[tr.Stock] += tr.Amount
			//}

			filter = bson.D{{"name", tr.Name}}
			update = bson.D{{"$set", bson.D{{"money", user.Money - price.Summ}, {fmt.Sprintf("portfolio.%s", tr.Stock), user.Portfolio[tr.Stock] + tr.Amount}}}}
			_, err = collection("users").UpdateOne(ctx, filter, update)
			if err != nil {
				log.Printf("UpdateUser: %s", err)
				continue
			}

			log.Println(fmt.Sprintf("Payed %s\n", price))
		} else {
			log.Println(fmt.Sprintf("not enough money: %s\n", price))
			decline(tr.Id, "Недостаточно средств")
		}
	}
}

func decline(transactionId primitive.ObjectID, cause string) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	filter := bson.D{{"_id", transactionId}}
	update := bson.D{{"$set", bson.D{{"error", cause}}}}
	_, err := collection("transactions").UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("DeclineTr: %s", err)
	}
}

func processSellTransactions() {
	for {
		tr := <-sellQueue

		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		userCollection := collection("users")
		var user User
		err := userCollection.FindOne(ctx, bson.M{"name": tr.Name}).Decode(&user)
		if err != nil {
			log.Printf("%s", err)
			continue
		}

		if _, ok := stock[tr.Stock]; !ok {
			log.Printf("Stock not found: %s", tr.Stock)
			decline(tr.Id, "Тикер не найден")
			continue
		}
		var price PriceResponse
		err = json.Unmarshal(MakeRequest(fmt.Sprintf("buy/%s/%d/", tr.Stock, tr.Amount)), &price)
		if err != nil {
			log.Printf("%s", err)
			decline(tr.Id, "Тикер закрыт")
			continue
		}

		if _, ok := user.Portfolio[tr.Stock]; !ok {
			log.Printf("User %s dont have stock: %s", user.Name, tr.Stock)
			decline(tr.Id, "У вас недостаточно акций")
			continue
		}

		if user.Portfolio[tr.Stock] < tr.Amount {
			log.Printf("User %s dont have enough stock: %s", user.Name, tr.Stock)
			decline(tr.Id, "У вас недостаточно акций")
			continue
		}

		ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
		filter := bson.D{{"_id", tr.Id}}
		update := bson.D{{"$set", bson.D{{"commited", true}}}}
		_, err = collection("transactions").UpdateOne(ctx, filter, update)
		if err != nil {
			log.Printf("UpdateTr: %s", err)
			continue
		}

		user.Portfolio[tr.Stock] -= tr.Amount

		filter = bson.D{{"name", tr.Name}}
		update = bson.D{{"$set", bson.D{{"money", user.Money + price.Summ}, {"portfolio", user.Portfolio}}}}
		_, err = collection("users").UpdateOne(ctx, filter, update)
		if err != nil {
			log.Printf("UpdateUser: %s", err)
			continue
		}

		log.Println(fmt.Sprintf("Sold %d %s for %s\n", tr.Amount, tr.Stock, price.Summ))
	}
}
