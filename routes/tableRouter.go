package routes

import (
	"github.com/gin-gonic/gin"
	"restaurant-order-management/controllers"
)

func TableRoutes(c *gin.Engine)  {
	c.GET("/tables",controllers.GetTables())
	c.GET("/tables/:table_id",controllers.GetTable())
	c.POST("/tables",controllers.CreateTables())
	c.PATCH("/tables:table_id",controllers.UpdateTable())
}