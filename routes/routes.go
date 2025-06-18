package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/sholllll662/invoice-backend/controllers"
	"github.com/sholllll662/invoice-backend/middlewares"
)

func RegisterRoutes(router *gin.Engine) {

	api := router.Group("/api")
	{
		api.POST("/register", controller.Register)
		api.POST("/login", controller.Login)
	}

	protected := router.Group("/api")
	protected.Use(middlewares.AuthMiddleware())
	{
		protected.GET("/protected", controller.ProtectedEndpoint)
		protected.POST("/clients", controller.CreateClient)
		protected.GET("/clients", controller.GetClients)
		protected.GET("/profile", middlewares.AuthMiddleware(), controller.GetProfile)
		protected.PUT("/clients/:id", controller.UpdateClient)
		protected.DELETE("/clients/:id", controller.DeleteClient)
		protected.POST("/invoices", controller.CreateInvoice)
		protected.GET("/invoices", middlewares.AuthMiddleware(), controller.GetInvoices)
		protected.GET("/invoices/:id", middlewares.AuthMiddleware(), controller.GetInvoiceByID)
		protected.PUT("/invoices/:id", middlewares.AuthMiddleware(), controller.UpdateInvoiceByID)
		protected.DELETE("/invoices/:id", middlewares.AuthMiddleware(), controller.DeleteInvoiceByID)
		protected.GET("/invoices/:id/pdf", middlewares.AuthMiddleware(), controller.ExportInvoicePDF)
	}
}
