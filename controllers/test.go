package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ProtectedEndpoint(c *gin.Context) {
	// ambil data user dari context (dari middleware)
	userID := c.GetUint("userID")
	userEmail := c.GetString("userEmail")

	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to the protected Route!",
		"userID":  userID,
		"email":   userEmail,
	})
}
