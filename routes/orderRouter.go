package routes

import (
	"github.com/gin-gonic/gin"
	"restaurant-order-management/controllers"
)

func OrderRoutes(c *gin.Engine)  {
	c.GET("/orders",controllers.GetOrders())
	c.GET("/orders/:order_id",controllers.GetOrder())
	c.POST("/orders",controllers.CreateOrders())
	c.PATCH("/orders:order_id",controllers.UpdateOrder())
}