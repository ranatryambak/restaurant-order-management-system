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

var tableCollection *mongo.Collection = database.OpenCollection(database.Client,"table")

func GetTables() gin.HandlerFunc{
	return func (c *gin.Context)  {
		ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)

		result,err := tableCollection.Find(ctx,bson.M{})
		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"unable to fetch the tables"})
			return
		}
		var allTable []bson.M

		err = result.All(ctx,&allTable)
		if err!=nil{
			log.Fatal(err)
		}

		defer cancel()
		c.JSON(http.StatusOK,allTable)
	}
}

func GetTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx,cancel:= context.WithTimeout(context.Background(),100*time.Second)

		var Table models.Table

		tableId := c.Param("table_id")

		err := tableCollection.FindOne(ctx,bson.M{"table_id":tableId}).Decode(&Table)
		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"unable to fetch this table details"})
			return 
		}

		defer cancel()

		c.JSON(http.StatusInternalServerError,Table)
	}
}

func CreateTables() gin.HandlerFunc{
	return func(c *gin.Context) {
		ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)

		var Table models.Table

		err := c.BindJSON(&Table)
		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return 
		}
		
		validationErr := validate.Struct(Table)
		if validationErr != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":validationErr.Error()})
			return
		}

		Table.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		Table.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))

		Table.ID = primitive.NewObjectID()
		Table.Table_id = Table.ID.Hex()

		result,err := tableCollection.InsertOne(ctx,Table)

		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"unable to create table"})
			return
		}

		defer cancel()

		c.JSON(http.StatusInternalServerError,result)


	}
}

func UpdateTable() gin.HandlerFunc{
	return func(c *gin.Context) {
		ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)

		var Table models.Table

		err := c.BindJSON(&Table)
		if err != nil{
			log.Fatal(err)
		}

		tableId := c.Param("table_id")

		filter := bson.M{"table_id":tableId}

		var updateObj primitive.D

		if Table.Number_of_guests != nil{
			updateObj = append(updateObj, bson.E{"number_of_guests",Table.Number_of_guests})
		}

		if Table.Table_number != nil{
			updateObj = append(updateObj, bson.E{"table_number",Table.Table_number})
		}

		Table.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))

		updateObj = append(updateObj, bson.E{"updated_at",Table.Updated_at})

		upsert := true

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := tableCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set",updateObj},
			},
			&opt,
		)

		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"unable to update the table section"})
			return
		}

		defer cancel()

		c.JSON(http.StatusInternalServerError,result)

	}
}