package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBinstance() *mongo.Client {
	MongoDB := "mongodb://localhost:27017/go-auth"
	fmt.Println(MongoDB)


	client,err := mongo.NewClient(options.Client().ApplyURI(MongoDB))

	if err!= nil{
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(),10*time.Second)

	defer cancel()

	err = client.Connect(ctx)

	if err!=nil{
		log.Fatal(err)
	}

	fmt.Println("connected to Mongdb")

	return client

}

var Client *mongo.Client = DBinstance()

func OpenCollection(client *mongo.Client,collectionName string) *mongo.Collection {
	var Collection *mongo.Collection = client.Database("restaurant").Collection(collectionName)

	return Collection
}