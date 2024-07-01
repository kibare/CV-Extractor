package config

import (
    "fmt"
    "log"
    "os"

    "github.com/joho/godotenv"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "cv-extractor/models"
)

var DB *gorm.DB

func InitDB() {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using environment variables")
    }

    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        log.Fatalf("DATABASE_URL not set in environment")
    }

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Error connecting to database: %v", err)
    }

    sqlDB, err := db.DB()
    if err != nil {
        log.Fatalf("Error getting DB from gorm DB: %v", err)
    }

    if err := sqlDB.Ping(); err != nil {
        log.Fatalf("Error pinging database: %v", err)
    }

    if err := db.AutoMigrate(&models.User{}, &models.Company{}, &models.Department{}, &models.Position{}, &models.Candidate{}); err != nil {
        log.Fatalf("Error during AutoMigrate: %v", err)
    }

    DB = db
    fmt.Println("Database connected successfully!")
}
