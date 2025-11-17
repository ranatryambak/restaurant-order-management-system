package routes

import (
	"github.com/gin-gonic/gin"
	"restaurant-order-management/controllers"
)

func OrderItemRoutes(c *gin.Engine)  {
	c.GET("/orderitems",controllers.GetOrderItems())
	c.GET("/orderitems/:orderitems_id",controllers.GetOrderItem())
	c.GET("/orderitems-order/:orderitems",controllers.GetOrderItemsByOrder())
	c.POST("/orderitems",controllers.CreateOrderItems())
	c.PATCH("/orderitems:orderitem_id",controllers.UpdateOrderItem())
}