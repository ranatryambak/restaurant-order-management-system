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

var menuCollection *mongo.Collection = database.OpenCollection(database.Client,"menu")




func GetMenus() gin.HandlerFunc{
	return func (c *gin.Context)  {
		ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)
		result,err := menuCollection.Find(context.TODO(),bson.M{})

		defer cancel()
		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"error occured while fetching the data"})
			return 
		}

		var allMenu []bson.M
		if err = result.All(ctx,&allMenu); err != nil{
			log.Fatal(err)
		}
		c.JSON(http.StatusOK,allMenu)
	}
}

func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()
		menuId := c.Param("Menu_id")
		var menu models.Menu
		err := menuCollection.FindOne(ctx, bson.M{"menu_id":menuId}).Decode(&menu)
		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"error occured while fetching the food"})
		}
		c.JSON(http.StatusOK,menu)
	}
}

func CreateMenus() gin.HandlerFunc{
	return func(c *gin.Context) {
		ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)
		
		var menu models.Menu
		if err := c.BindJSON(&menu); err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return
		}
		if validateErr := validate.Struct(menu); validateErr!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":validateErr.Error()})
			return
		}
		menu.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		menu.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))

		menu.ID = primitive.NewObjectID()
		menu.Menu_id = menu.ID.Hex()

		result,InsertErr := menuCollection.InsertOne(ctx,menu)
		if InsertErr!=nil{
			msg := "Menu item was not created"
			c.JSON(http.StatusInternalServerError,gin.H{"error":msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK,result)
	}
}

func UpdateMenu() gin.HandlerFunc{
	return func(c *gin.Context) {
		ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)
		var menu models.Menu
		if err:=c.BindJSON(&menu); err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return
		}
		menuId := c.Param("menu_id")
		filter := bson.M{"menu_id":menuId}
		

		var updateObj primitive.D

		if menu.Start_Date!=nil && menu.End_Date!=nil{
			if !inTimeSpan(*menu.Start_Date,*menu.End_Date,time.Now()){
				msg := "kindly retype the time"
				c.JSON(http.StatusInternalServerError,gin.H{"error":msg})
				defer cancel()
				return 
			}
		}
		updateObj = append(updateObj, bson.E{"start_date",menu.Start_Date})
		updateObj = append(updateObj, bson.E{"end_date", menu.End_Date})

		if menu.Name != ""{
			updateObj = append(updateObj, bson.E{"name", menu.Name})
		}

		if menu.Category != ""{
			updateObj = append(updateObj, bson.E{"category", menu.Category})
		}

		menu.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at",menu.Updated_at})

		upsert := true

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result,err :=menuCollection.UpdateOne(
				ctx,
				filter,
				bson.D{
					{"$set", updateObj},
				},
				&opt,
			)
		if err!=nil{
			msg := "Menu update failed"
			c.JSON(http.StatusInternalServerError,gin.H{"error":msg})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK,result)
	}

}

func inTimeSpan(start,end,check time.Time) bool {
	return start.After(time.Now()) && end.After(start)
}