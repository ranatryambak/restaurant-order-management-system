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

type OrderItemPack struct {
	Table_id    *string
	Order_items []models.OrderItems
}

var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client,"orderItem")

func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)

		result,err := orderItemCollection.Find(context.TODO(),bson.M{})

		defer cancel()

		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"order items were not found"})
			return 
		}

		var allOrderItems []bson.M

		err = result.All(ctx,&allOrderItems)
		if err!=nil{
			log.Fatal(err)
			return
		}

		c.JSON(http.StatusOK,result)
	}
}

func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)

		orderItemId := c.Param("orderitem_id")

		var orderItem models.OrderItems

		err:= orderItemCollection.FindOne(ctx,bson.M{"orderitem_id":orderItemId}).Decode(&orderItem)
		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"error occured while fetching the orderitems"})
			return
		}

		defer cancel()

		c.JSON(http.StatusOK,orderItem)
	}
}

func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		
		orderId := c.Param("order_id")

		allOrderItems,err := ItemsByOrder(orderId) 

		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"unable to fetch the order items"})
			return
		}

		c.JSON(http.StatusInternalServerError,allOrderItems)
	}
}

func ItemsByOrder(id string) (OrderItems []primitive.M, err error) {
	ctx,cancel:= context.WithTimeout(context.Background(),100*time.Second)

	matchStage := bson.D{{"$match",bson.D{{"order_id",id}}}}

	lookupStage := bson.D{{"$lookup",bson.D{{"from","food"},{"localField","food_id"},{"foreignField","food_id"},{"as","food"}}}}
	unwindStage := bson.D{{"$unwind",bson.D{{"path","$food"},{"preserveNullAndEmptyArrays",true}}}}

	lookupOrderStage := bson.D{{"$lookup",bson.D{{"from","order"},{"localField","order_id"},{"foreignField","order_id"},{"as","order"}}}}
	unwindOrderStage := bson.D{{"$unwind",bson.D{{"path","$order"},{"preserveNullAndEmptyArrays",true}}}}

	lookupTableStage := bson.D{{"$lookup",bson.D{{"from","table"},{"localField","order.table_id"},{"foreignField","table_id"},{"as","table"}}}}
	unwindTableStage := bson.D{{"$unwind",bson.D{{"path","$table"},{"preserveNullAndEmptyArrays",true}}}}

	projectStage := bson.D{{"$project",bson.D{
		{"_id",0},
		{"amount","$food.price"},
		{"total_count",1},
		{"food_name","$food.name"},
		{"food_image","$food.food_image"},
		{"table_number","$food.table_number"},
		{"table_id","$table.table_id"},
		{"order_id","$table.order_id"},{"price","$food.price"},
		{"Quantity",1},
	}}}

	groupStage := bson.D{{"$group",bson.D{{"_id",bson.D{{"order_id","$order_id"},{"table_id","$table_id"},{"table_number","$table_number"}}},{"payment_due",bson.D{{"$sum","$amount"}}},{"total_count",bson.D{{"$sum",1}}},{"order_items",bson.D{{"$push","$$ROOT"}}}}}}

	projectStage2 := bson.D{{"$project",bson.D{
		{"_id",0},
		{"payment_due",1},
		{"total_count",1},
		{"table_number","$_id.table_number"},
		{"order_items",1},
	}}}

	result,err := orderItemCollection.Aggregate(ctx,mongo.Pipeline{
					matchStage,lookupStage,unwindStage,lookupOrderStage,unwindOrderStage,lookupTableStage,unwindTableStage,projectStage,groupStage,projectStage2,
				})
	if err!=nil{
		panic(err)
	}

	

	err = result.All(ctx,&OrderItems)
	if err !=nil{
		panic(err)
	}

	defer cancel()

	return OrderItems,err

}

func CreateOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx,cancel:= context.WithTimeout(context.Background(),100*time.Second)

		var orderItemPack OrderItemPack
		var order models.Order

		err:= c.BindJSON(&orderItemPack)
		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return
		}

		order.Order_Date,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		orderItemsToBeInserted := []interface{}{}
		order.Table_id = orderItemPack.Table_id
		order_id := OrderItemOrderCreator(order)

		for _,orderItem := range orderItemPack.Order_items{
			orderItem.Order_id = order_id

			validationErr := validate.Struct(orderItem)
			if validationErr != nil{
				c.JSON(http.StatusBadRequest,gin.H{"error":validationErr.Error()})
				return
			}
			orderItem.ID = primitive.NewObjectID()
			orderItem.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
			orderItem.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
			orderItem.Order_items_id = orderItem.ID.Hex()
			var num = tobefixed(*orderItem.Unit_Price,2)
			orderItem.Unit_Price = &num
			orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)
			 
		}

		inserted , err := orderCollection.InsertMany(ctx,orderItemsToBeInserted)
		if err!=nil{
			log.Fatal(err)
		}
		defer cancel()
		c.JSON(http.StatusOK,inserted)
	}
}

func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)

		var orderItem models.OrderItems
		OrderItemId := c.Param("order_items_id")
		filter := bson.M{"order_items_id":OrderItemId}

		var updateObj primitive.D

		if orderItem.Unit_Price != nil{
			updateObj = append(updateObj, bson.E{"unit_price",orderItem.Unit_Price})
		}
		if orderItem.Quantity != nil{
			updateObj = append(updateObj, bson.E{"quantity",orderItem.Quantity})
		}
		if orderItem.Food_id != nil{
			updateObj = append(updateObj, bson.E{"food_id",orderItem.Food_id})
		}


		orderItem.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))

		updateObj = append(updateObj, bson.E{"updated_at",orderItem.Updated_at})


		upsert:= true

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result,err := orderItemCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set",updateObj},
			},
			&opt,
		)
		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"unable to update the order items"})
			return
		}

		defer cancel()

		c.JSON(http.StatusOK,result)

	}
}
