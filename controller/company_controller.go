package controller

import (
    "net/http"
    "cv-extractor/config"
    "cv-extractor/models"
    "cv-extractor/utils"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "log"
)

type CreateCompanyInput struct {
    Name    string `json:"name" binding:"required"`
    Address string `json:"address" binding:"required"`
}

type EditCompanyInput struct {
    Name    string `json:"name" binding:"required"`
    Address string `json:"address" binding:"required"`
}

func CreateCompany(c *gin.Context) {
    var input CreateCompanyInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
        return
    }

    var existingCompany models.Company
    if err := config.DB.Where("name = ?", input.Name).First(&existingCompany).Error; err == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Company already exists"})
        return
    }

    company := models.Company{
        Name:    input.Name,
        Address: input.Address,
    }

    if err := config.DB.Create(&company).Error; err != nil {
        log.Printf("Failed to create company: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Company created successfully",
        "company": company,
    })
}

func GetOneCompany(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    id := c.Param("id")
    var company models.Company
    if err := config.DB.First(&company, id).Error; err != nil {
        log.Printf("Company not found: %v\n", err)
        c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
        return
    }

    if company.ID != userClaims.CompanyID {
        c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this company"})
        return
    }

    c.JSON(http.StatusOK, company)
}

func GetAllCompanies(c *gin.Context) {
    var companies []models.Company
    if err := config.DB.Find(&companies).Error; err != nil {
        log.Printf("Failed to retrieve companies: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve companies"})
        return
    }
    c.JSON(http.StatusOK, companies)
}

func EditCompany(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    id := c.Param("id")
    var input EditCompanyInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
        return
    }

    var company models.Company
    if err := config.DB.First(&company, id).Error; err != nil {
        log.Printf("Company not found: %v\n", err)
        c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
        return
    }

    if company.ID != userClaims.CompanyID {
        c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to edit this company"})
        return
    }

    company.Name = input.Name
    company.Address = input.Address

    if err := config.DB.Save(&company).Error; err != nil {
        log.Printf("Failed to update company: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Company updated successfully", "company": company})
}

func DeleteCompany(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    id := c.Param("id")
    var company models.Company
    if err := config.DB.First(&company, id).Error; err != nil {
        log.Printf("Company not found: %v\n", err)
        c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
        return
    }

    if company.ID != userClaims.CompanyID {
        c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to delete this company"})
        return
    }

    tx := config.DB.Begin()
    if tx.Error != nil {
        log.Printf("Failed to start transaction: %v\n", tx.Error)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
        return
    }

    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            log.Printf("Panic occurred: %v\n", r)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete company"})
        }
    }()

    if err := deleteRelatedData(tx, company.ID); err != nil {
        tx.Rollback()
        log.Printf("Failed to delete related data: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    if err := tx.Delete(&company).Error; err != nil {
        tx.Rollback()
        log.Printf("Failed to delete company: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete company"})
        return
    }

    if err := tx.Commit().Error; err != nil {
        log.Printf("Failed to commit transaction: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Company deleted successfully"})
}

func deleteRelatedData(tx *gorm.DB, companyID uint) error {
    var departments []models.Department
    if err := tx.Where("company_id = ?", companyID).Find(&departments).Error; err != nil {
        return err
    }

    for _, department := range departments {
        var positions []models.Position
        if err := tx.Where("department_id = ?", department.ID).Find(&positions).Error; err != nil {
            return err
        }

        for _, position := range positions {
            var candidates []models.Candidate
            if err := tx.Where("position_id = ?", position.ID).Find(&candidates).Error; err != nil {
                return err
            }

            for _, candidate := range candidates {
                if err := utils.DeleteFileFromFirebase(candidate.CVFile); err != nil {
                    return err
                }

                if err := tx.Delete(&candidate).Error; err != nil {
                    return err
                }
            }

            if err := tx.Delete(&position).Error; err != nil {
                return err
            }
        }

        if err := tx.Delete(&department).Error; err != nil {
            return err
        }
    }

    return nil
}
