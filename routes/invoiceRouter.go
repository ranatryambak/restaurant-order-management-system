package routes

import (
	"github.com/gin-gonic/gin"
	"restaurant-order-management/controllers"
)

func InvoiceRoutes(c *gin.Engine)  {
	c.GET("/invoices",controllers.GetInvoices())
	c.GET("/invoices/:invoice_id",controllers.GetInvoice())
	c.POST("/invoices",controllers.CreateInvoices())
	c.PATCH("/invoices:invoice_id",controllers.UpdateInvoice())
}