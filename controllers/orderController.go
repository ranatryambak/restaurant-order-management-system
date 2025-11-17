package controllers

import (
	"context"
	"log"
	"net/http"
	"restaurant-order-management/database"
	"restaurant-order-management/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var orderCollection *mongo.Collection = database.OpenCollection(database.Client,"order")


func GetOrder() gin.HandlerFunc{
	return func (c *gin.Context)  {
		ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)
		result,err := orderCollection.Find(context.TODO(),bson.M{})
		defer cancel()
		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"error occured while listening to the orders"})
			return
		}
		var allModel []bson.M
		if err = result.All(ctx,&allModel); err!=nil{
			log.Fatal(err)
		}
		c.JSON(http.StatusOK,allModel)
	}
}

func GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(),100*time.Second)
		OrderId := c.Param("order_id")
		var order models.Order
		err := orderCollection.FindOne(ctx,bson.M{"order_id":OrderId}).Decode(&order)
		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"error occured while fetching the order"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK,order)
	}
}

func CreateOrders() gin.HandlerFunc{
	return func(c *gin.Context) {
		ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)
		var order models.Order
		var table models.Table

		err:= c.BindJSON(&order)
		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return
		}

		validationErr := validate.Struct(order)

		if validationErr!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":validationErr.Error()})
			return
		}
		if order.Table_id != nil{
			err = tableCollection.FindOne(ctx,bson.M{"table_id":order.Table_id}).Decode(&table)
			defer cancel()
			if err != nil{
				c.JSON(http.StatusInternalServerError,gin.H{"error":"table was not found"})
				return
			}
		}

		order.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		order.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))

		order.ID = primitive.NewObjectID()
		order.Order_id = order.ID.Hex()
		 
		result,err:= orderCollection.InsertOne(ctx,order)

		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"order was not created"})
			return 
		}
		defer cancel()

		c.JSON(http.StatusOK,result)
	}
}

func UpdateOrder() gin.HandlerFunc{
	return func(c *gin.Context) {
		ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)

		var table models.Table
		var order models.Order

		orderId := c.Param("order_id")

		var updateObj primitive.D
		if err:=c.BindJSON(&order);err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return
		}

		if order.Table_id != nil{
			err := tableCollection.FindOne(ctx, bson.M{"table_id":order.Table_id}).Decode(&table)
			defer cancel()
			if err!=nil{
				c.JSON(http.StatusInternalServerError,gin.H{"error":"table was not found"})
				return
			}
			updateObj = append(updateObj, bson.E{"table_id",order.Table_id})
		}

		order.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))

		updateObj = append(updateObj, bson.E{"updated_at",order.Updated_at})

		upsert := true

		filter := bson.M{"order_id":orderId}

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result,err:= orderCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set",updateObj},
			},
			&opt,
		)

		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"order updation failed"})
			return
		}

		defer cancel()

		c.JSON(http.StatusOK,result)


	}
}

func OrderItemOrderCreator(order models.Order)string {

	ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)

	order.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
	order.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
	order.ID = primitive.NewObjectID()
	order.Order_id = order.ID.Hex()

	orderCollection.InsertOne(ctx,order)
	defer cancel()

	return order.Order_id
	
}