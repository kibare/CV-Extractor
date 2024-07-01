package controller

import (
    "cv-extractor/config"
    "cv-extractor/models"
    "cv-extractor/utils"
    "net/http"

    "github.com/gin-gonic/gin"
)

type CreateDepartmentInput struct {
    Name string `json:"name" binding:"required"`
}

func CreateDepartment(c *gin.Context) {
    var input CreateDepartmentInput

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "All fields must be required", "details": err.Error()})
        return
    }

    userClaims := c.MustGet("claims").(*utils.Claims)
    companyID := userClaims.CompanyID

    var company models.Company
    if err := config.DB.First(&company, companyID).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "Company doesn't exist"})
        return
    }

    department := models.Department{
        Name:      input.Name,
        CompanyID: companyID,
    }

    if err := config.DB.Create(&department).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Server Error", "error": err.Error()})
        return
    }

    if err := config.DB.Preload("Company").First(&department, department.ID).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Server Error", "error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message":    "Department created successfully",
        "department": department,
    })
}

func GetAllDepartments(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)

    var departments []models.Department
    if err := config.DB.Where("company_id = ?", userClaims.CompanyID).Preload("Positions").Find(&departments).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Server Error", "error": err.Error()})
        return
    }

    if len(departments) == 0 {
        c.JSON(http.StatusNotFound, gin.H{"message": "Department is empty"})
        return
    }

    c.JSON(http.StatusOK, departments)
}

func GetOneDepartment(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest, gin.H{"message": "ID must be provided"})
        return
    }

    var department models.Department
    if err := config.DB.Preload("Company").First(&department, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Department does not exist"})
        return
    }

    if department.CompanyID != userClaims.CompanyID {
        c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this department"})
        return
    }

    c.JSON(http.StatusOK, department)
}

func EditDepartment(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    id := c.Param("id")
    var input struct {
      Name string `json:"name" binding:"required"`
    }
  
    if err := c.ShouldBindJSON(&input); err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"message": "All fields must be required", "details": err.Error()})
      return
    }
  
    var department models.Department
    if err := config.DB.First(&department, id).Error; err != nil {
      c.JSON(http.StatusNotFound, gin.H{"message": "Department does not exist"})
      return
    }
  
    if department.CompanyID != userClaims.CompanyID {
        c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to edit this department"})
        return
    }

    var existingDepartment models.Department
    if err := config.DB.Where("name = ? AND company_id = ?", input.Name, department.CompanyID).First(&existingDepartment).Error; err == nil {
      c.JSON(http.StatusConflict, gin.H{"message": "Department name already exists"})
      return
    }
  
    department.Name = input.Name
    if err := config.DB.Save(&department).Error; err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"message": "Server Error", "error": err.Error()})
      return
    }
  
    if err := config.DB.Preload("Company").First(&department, id).Error; err != nil {
      c.JSON(http.StatusNotFound, gin.H{"message": "Department does not exist"})
      return
    }
  
    c.JSON(http.StatusOK, gin.H{
      "message": "Department updated successfully",
      "department": department,
    })
}

func DeleteDepartment(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    id := c.Param("id")

    var department models.Department
    if err := config.DB.First(&department, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Department not found"})
        return
    }

    if department.CompanyID != userClaims.CompanyID {
        c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to delete this department"})
        return
    }

    var positions []models.Position
    if err := config.DB.Where("department_id = ?", department.ID).Find(&positions).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Server Error", "error": err.Error()})
        return
    }

    for _, position := range positions {
        var candidates []models.Candidate
        if err := config.DB.Where("position_id = ?", position.ID).Find(&candidates).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"message": "Server Error", "error": err.Error()})
            return
        }

        for _, candidate := range candidates {
            if err := utils.DeleteFileFromFirebase(candidate.CVFile); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"message": "Server Error", "error": err.Error()})
                return
            }

            if err := config.DB.Delete(&candidate).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"message": "Server Error", "error": err.Error()})
                return
            }
        }

        if err := config.DB.Delete(&position).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"message": "Server Error", "error": err.Error()})
            return
        }
    }

    if err := config.DB.Delete(&department).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Server Error", "error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Department deleted successfully"})
}
