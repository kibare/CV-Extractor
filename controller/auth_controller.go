package controller

import (
    "net/http"
    "time"
    "cv-extractor/config"
    "cv-extractor/models"
    "cv-extractor/utils"

    "github.com/gin-gonic/gin"
)

type LoginInput struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type RegisterInput struct {
    Name      string `json:"name" binding:"required"`
    Email     string `json:"email" binding:"required,email"`
    Password  string `json:"password" binding:"required,min=8"`
    Phone     string `json:"phone" binding:"required"`
    CompanyID uint   `json:"company_id" binding:"required"`
}

func Login(c *gin.Context) {
    var input LoginInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
        return
    }

    var user models.User
    if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    if !utils.CheckPasswordHash(input.Password, user.Password) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    token, err := utils.GenerateJWT(user.ID, *user.CompanyID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "token":   token,
        "user":    user.ID,
        "company": user.CompanyID,
    })
}

func Register(c *gin.Context) {
    var input RegisterInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
        return
    }

    if userExists(input.Email) {
        c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
        return
    }

    if !companyExists(input.CompanyID) {
        c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
        return
    }

    hashedPassword, err := utils.HashPassword(input.Password)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    user := models.User{
        Name:        input.Name,
        Email:       input.Email,
        Password:    hashedPassword,
        Phone:       input.Phone,
        CompanyID:   &input.CompanyID,
        CreatedDate: time.Now(),
    }

    if err := config.DB.Create(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "User created successfully",
        "user":    user,
    })
}

func userExists(email string) bool {
    var user models.User
    if err := config.DB.Where("email = ?", email).First(&user).Error; err == nil {
        return true
    }
    return false
}

func companyExists(companyID uint) bool {
    var company models.Company
    if err := config.DB.First(&company, companyID).Error; err == nil {
        return true
    }
    return false
}
