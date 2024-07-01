package controller

import (
    "net/http"
    "cv-extractor/config"
    "cv-extractor/models"
    "cv-extractor/utils"
    "github.com/gin-gonic/gin"
)

type EditUserInput struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
    Phone string `json:"phone" binding:"required"`
}

type ChangePasswordInput struct {
    Password string `json:"password" binding:"required,min=8"`
}

// GetUser retrieves a user by ID
func GetUser(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    var user models.User
    if err := config.DB.First(&user, userClaims.UserID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found", "details": err.Error()})
        return
    }
    c.JSON(http.StatusOK, user)
}

// EditUser updates a user's details
func EditUser(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    var input EditUserInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
        return
    }

    var user models.User
    if err := config.DB.First(&user, userClaims.UserID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found", "details": err.Error()})
        return
    }

    var existingUser models.User
    if err := config.DB.Where("email = ? AND id != ?", input.Email, userClaims.UserID).First(&existingUser).Error; err == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Email already in use"})
        return
    }

    user.Name = input.Name
    user.Email = input.Email
    user.Phone = input.Phone

    if err := config.DB.Save(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user", "details": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User updated successfully", "user": user})
}

// ChangePassword updates a user's password
func ChangePassword(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    var input ChangePasswordInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
        return
    }

    var user models.User
    if err := config.DB.First(&user, userClaims.UserID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found", "details": err.Error()})
        return
    }

    hashedPassword, err := utils.HashPassword(input.Password)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password", "details": err.Error()})
        return
    }

    user.Password = hashedPassword

    if err := config.DB.Save(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password", "details": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// DeleteUser deletes a user by ID
func DeleteUser(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    var user models.User
    if err := config.DB.First(&user, userClaims.UserID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found", "details": err.Error()})
        return
    }

    if err := config.DB.Delete(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user", "details": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// GetAllUsers retrieves all users
func GetAllUsers(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    var users []models.User
    if err := config.DB.Where("company_id = ?", userClaims.CompanyID).Find(&users).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users", "details": err.Error()})
        return
    }
    c.JSON(http.StatusOK, users)
}
