package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sholllll662/invoice-backend/database"
	"github.com/sholllll662/invoice-backend/models"
)

func CreateClient(c *gin.Context) {
	// amnil user ID dari context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"Error": "Unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	// bind input json
	var input struct {
		Nama  string `json:"nama" binding:"required"`
		Email string `json:"email" binding:"required,email"`
		NoTlp string `json:"no_tlp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// buat client baru
	client := models.Client{
		UserID: userID,
		Nama:   input.Nama,
		Email:  input.Email,
		NoTlp:  input.NoTlp,
	}

	if err := database.DB.Create(&client).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan data client"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Client berhasil ditambahkan",
		"client":  client,
	})
}

func GetClients(c *gin.Context) {
	// Ambil user ID dari context (hasil middleware)
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	// Ambil query pencarian jika ada
	search := c.DefaultQuery("search", "")

	var clients []models.Client
	query := database.DB.Where("user_id = ?", userID)

	if search != "" {
		query = query.Where("LOWER(nama) LIKE ?", "%"+strings.ToLower(search)+"%")
	}

	if err := query.Order("created_at desc").Find(&clients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data client"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"clients": clients})
}

func UpdateClient(c *gin.Context) {
	userID := c.GetUint("userID")
	clientID := c.Param("id")

	var client models.Client
	if err := database.DB.Where("id = ? AND user_id = ?", clientID, userID).First(&client).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client tidak ditemukan"})
		return
	}

	var input struct {
		Nama  string `json:"nama" binding:"required"`
		Email string `json:"email" binding:"required,email"`
		NoTlp string `json:"no_tlp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client.Nama = input.Nama
	client.Email = input.Email
	client.NoTlp = input.NoTlp

	if err := database.DB.Save(&client).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update client"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "client berhasil diupdate", "client": client})
}

func DeleteClient(c *gin.Context) {
	userID := c.GetUint("userID")
	clientID := c.Param("id")

	var client models.Client
	if err := database.DB.Where("id = ? AND user_id = ?", clientID, userID).First(&client).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client tidak ditemukan"})
		return
	}

	if err := database.DB.Delete(&client).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus client"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Client berhasil dihapus"})
}
