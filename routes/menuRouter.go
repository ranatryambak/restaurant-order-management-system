package routes

import (
	"github.com/gin-gonic/gin"
	"restaurant-order-management/controllers"
)

func MenuRoutes(c *gin.Engine)  {
	c.GET("/menus",controllers.GetMenus())
	c.GET("/menus/:menu_id",controllers.GetMenu())
	c.POST("/menus",controllers.CreateMenus())
	c.PATCH("/menus:menu_id",controllers.UpdateMenu())
}