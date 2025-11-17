package routes

import (
	"github.com/gin-gonic/gin"
	"restaurant-order-management/controllers"
)

func UserRouter(c *gin.Engine) {
	c.GET("/users",controllers.GetUsers())
	c.GET("/users/:user_id",controllers.GetUser())
	c.POST("/users/signup",controllers.Signup())
	c.POST("/users/login", controllers.Login())
}