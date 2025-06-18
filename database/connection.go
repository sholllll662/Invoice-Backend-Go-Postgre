package database

import (
	"fmt"
	"log"
	"os"

	"github.com/sholllll662/invoice-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = db
	fmt.Println("🚀 Connected to PostgreSQL!")

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("❌ Failed to migrate User model:", err)
	}

	// migrate tabel client
	err = db.AutoMigrate(&models.Client{})
	if err != nil {
		log.Fatal("❌ Failed to migrate Client model:", err)
	}

	// migrate tabel client
	err = db.AutoMigrate(&models.Invoice{})
	if err != nil {
		log.Fatal("❌ Failed to migrate Client model:", err)
	}

	// migrate tabel client
	err = db.AutoMigrate(&models.InvoiceItem{})
	if err != nil {
		log.Fatal("❌ Failed to migrate Client model:", err)
	}
}
