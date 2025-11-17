package routes

import (
	"github.com/gin-gonic/gin"
	"restaurant-order-management/controllers"
)

func FoodRoutes(c *gin.Engine) {
	c.GET("/foods",controllers.GetFoods())
	c.GET("/foods/:food_id",controllers.GetFood())
	c.POST("/foods",controllers.CreateFood())
	c.PATCH("foods:Food_id",controllers.UpdateFood())
}