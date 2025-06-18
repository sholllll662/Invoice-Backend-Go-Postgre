package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sholllll662/invoice-backend/config"
	"github.com/sholllll662/invoice-backend/database"
	"github.com/sholllll662/invoice-backend/middlewares"
	"github.com/sholllll662/invoice-backend/routes"
)

func main() {
	config.LoadEnv()
	database.ConnectDB()

	r := gin.Default()

	r.Use(middlewares.CORSMiddleware())
	routes.RegisterRoutes(r)

	// Test route
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Invoice Tracker Backend is running!",
		})
	})

	port := ":8080"
	fmt.Println("Server running on http://localhost" + port)
	r.Run(port) // listen and serve on 0.0.0.0:8080

}
