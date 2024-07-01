package controller

import (
	"net/http"
	"time"
	"cv-extractor/config"
	"cv-extractor/utils"
	"cv-extractor/models"
	"github.com/gin-gonic/gin"
	"log"
)

type CreatePositionInput struct {
	Name          string `json:"name" binding:"required"`
	Education     string `json:"education" binding:"required"`
	Location      string `json:"location" binding:"required"`
	MinWorkExp    int    `json:"minWorkExp" binding:"required"`
	Description   string `json:"description" binding:"required"`
	Qualification string `json:"qualification" binding:"required"`
	DepartmentID  uint   `json:"departmentId" binding:"required"`
}

type EditPositionInput struct {
	Name          string `json:"name" binding:"required"`
	Education     string `json:"education" binding:"required"`
	Location      string `json:"location" binding:"required"`
	MinWorkExp    int    `json:"minWorkExp" binding:"required"`
	Description   string `json:"description" binding:"required"`
	Qualification string `json:"qualification" binding:"required"`
}

type EditPositionCandidatesInput struct {
	ID                  uint   `json:"id" binding:"required"`
	QualifiedCandidates string `json:"qualifiedCandidates" binding:"required"`
}

type TrashPositionInput struct {
	IDs []uint `json:"ids" binding:"required"`
}

type DeletePositionInput struct {
	IDs []uint `json:"ids" binding:"required"`
}

func CreatePosition(c *gin.Context) {
	var input CreatePositionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	var existingPosition models.Position
	if err := config.DB.Where("name = ? AND department_id = ?", input.Name, input.DepartmentID).First(&existingPosition).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Position already exists"})
		return
	}

	var department models.Department
	if err := config.DB.First(&department, input.DepartmentID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Department does not exist"})
		return
	}

	userClaims := c.MustGet("claims").(*utils.Claims)
	if department.CompanyID != userClaims.CompanyID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to create a position for this department"})
		return
	}

	now := time.Now()

	position := models.Position{
		Name:          input.Name,
		Education:     input.Education,
		Location:      input.Location,
		MinWorkExp:    input.MinWorkExp,
		Description:   input.Description,
		Qualification: input.Qualification,
		DepartmentID:  input.DepartmentID,
		CreatedDate:   now,
	}

	if err := config.DB.Create(&position).Error; err != nil {
		log.Printf("Failed to create position: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create position"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Position created successfully",
		"position": position,
	})
}

func GetAllPositions(c *gin.Context) {
	userClaims := c.MustGet("claims").(*utils.Claims)

	var departments []models.Department
	if err := config.DB.Where("company_id = ?", userClaims.CompanyID).Find(&departments).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Departments do not exist"})
		return
	}

	var departmentIDs []uint
	for _, dept := range departments {
		departmentIDs = append(departmentIDs, dept.ID)
	}

	var positions []models.Position
	if err := config.DB.Where("department_id IN (?)", departmentIDs).Find(&positions).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Positions do not exist"})
		return
	}

	c.JSON(http.StatusOK, positions)
}

func GetOnePosition(c *gin.Context) {
	userClaims := c.MustGet("claims").(*utils.Claims)
	id := c.Param("id")
	var position models.Position
	if err := config.DB.First(&position, id).Error; err != nil {
		log.Printf("Position not found: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Position not found"})
		return
	}

	var department models.Department
	if err := config.DB.First(&department, position.DepartmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Department does not exist"})
		return
	}

	if department.CompanyID != userClaims.CompanyID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this position"})
		return
	}

	c.JSON(http.StatusOK, position)
}

func EditPosition(c *gin.Context) {
	userClaims := c.MustGet("claims").(*utils.Claims)
	id := c.Param("id")
	var input EditPositionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	var position models.Position
	if err := config.DB.First(&position, id).Error; err != nil {
		log.Printf("Position not found: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Position not found"})
		return
	}

	var department models.Department
	if err := config.DB.First(&department, position.DepartmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Department does not exist"})
		return
	}

	if department.CompanyID != userClaims.CompanyID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to edit this position"})
		return
	}

	position.Name = input.Name
	position.Education = input.Education
	position.Location = input.Location
	position.MinWorkExp = input.MinWorkExp
	position.Description = input.Description
	position.Qualification = input.Qualification

	if err := config.DB.Save(&position).Error; err != nil {
		log.Printf("Failed to update position: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update position"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Position updated successfully", "position": position})
}

func EditPositionCandidates(c *gin.Context) {
	userClaims := c.MustGet("claims").(*utils.Claims)
	var input EditPositionCandidatesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	var position models.Position
	if err := config.DB.First(&position, input.ID).Error; err != nil {
		log.Printf("Position not found: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Position not found"})
		return
	}

	var department models.Department
	if err := config.DB.First(&department, position.DepartmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Department does not exist"})
		return
	}

	if department.CompanyID != userClaims.CompanyID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to edit this position"})
		return
	}

	position.QualifiedCandidates = input.QualifiedCandidates

	if err := config.DB.Save(&position).Error; err != nil {
		log.Printf("Failed to update position candidates: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update position candidates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Position updated successfully", "position": position})
}

func ResolvePosition(c *gin.Context) {
	userClaims := c.MustGet("claims").(*utils.Claims)
	id := c.Param("id")
	var position models.Position
	if err := config.DB.First(&position, id).Error; err != nil {
		log.Printf("Position not found: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Position not found"})
		return
	}

	var department models.Department
	if err := config.DB.First(&department, position.DepartmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Department does not exist"})
		return
	}

	if department.CompanyID != userClaims.CompanyID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to resolve this position"})
		return
	}

	position.IsResolved = !position.IsResolved

	if err := config.DB.Save(&position).Error; err != nil {
		log.Printf("Failed to update position resolved status: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update position resolved status"})
		return
	}

	statusMessage := "Position resolved successfully"
	if !position.IsResolved {
		statusMessage = "Position restored successfully"
	}

	c.JSON(http.StatusOK, gin.H{"message": statusMessage})
}

func TrashPosition(c *gin.Context) {
	userClaims := c.MustGet("claims").(*utils.Claims)
	var input TrashPositionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	now := time.Now()

	for _, id := range input.IDs {
		var position models.Position
		if err := config.DB.First(&position, id).Error; err != nil {
			log.Printf("Position not found: %v\n", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Position not found"})
			return
		}

		var department models.Department
		if err := config.DB.First(&department, position.DepartmentID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Department does not exist"})
			return
		}

		if department.CompanyID != userClaims.CompanyID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to trash this position"})
			return
		}

		position.IsTrash = !position.IsTrash
		position.RemovedDate = now

		if err := config.DB.Save(&position).Error; err != nil {
			log.Printf("Failed to update position trash status: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update position trash status"})
			return
		}
	}

	statusMessage := "Positions removed successfully"
	if !now.IsZero() {
		statusMessage = "Positions restored successfully"
	}

	c.JSON(http.StatusOK, gin.H{"message": statusMessage})
}

func ArchivePosition(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    id := c.Param("id")
    var position models.Position
    if err := config.DB.First(&position, id).Error; err != nil {
        log.Printf("Position not found: %v\n", err)
        c.JSON(http.StatusNotFound, gin.H{"error": "Position not found"})
        return
    }

    var department models.Department
    if err := config.DB.First(&department, position.DepartmentID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Department does not exist"})
        return
    }

    if department.CompanyID != userClaims.CompanyID {
        c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to archive this position"})
        return
    }

    position.IsArchive = !position.IsArchive

    if err := config.DB.Save(&position).Error; err != nil {
        log.Printf("Failed to archive position: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive position"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Position archived successfully", "position": position})
}

func GetArchivedPositions(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)

    var departments []models.Department
    if err := config.DB.Where("company_id = ?", userClaims.CompanyID).Find(&departments).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Departments do not exist"})
        return
    }

    var departmentIDs []uint
    for _, dept := range departments {
        departmentIDs = append(departmentIDs, dept.ID)
    }

    var positions []models.Position
    if err := config.DB.Where("department_id IN (?) AND is_archive = ?", departmentIDs, true).Find(&positions).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Archived positions do not exist"})
        return
    }

    c.JSON(http.StatusOK, positions)
}

func DeletePosition(c *gin.Context) {
	userClaims := c.MustGet("claims").(*utils.Claims)
	var input DeletePositionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Start a database transaction
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete positions"})
		}
	}()

	for _, id := range input.IDs {
		var position models.Position
		if err := tx.First(&position, id).Error; err != nil {
			tx.Rollback()
			log.Printf("Position not found: %v\n", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Position not found"})
			return
		}

		var department models.Department
		if err := config.DB.First(&department, position.DepartmentID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Department does not exist"})
			return
		}

		if department.CompanyID != userClaims.CompanyID {
			tx.Rollback()
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to delete this position"})
			return
		}

		var candidates []models.Candidate
		if err := tx.Where("position_id = ?", position.ID).Find(&candidates).Error; err != nil {
			tx.Rollback()
			log.Printf("Failed to retrieve candidates: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve candidates"})
			return
		}

		for _, candidate := range candidates {
			if err := utils.DeleteFileFromFirebase(candidate.CVFile); err != nil {
				tx.Rollback()
				log.Printf("Failed to delete file from S3: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file from S3"})
				return
			}

			if err := tx.Delete(&candidate).Error; err != nil {
				tx.Rollback()
				log.Printf("Failed to delete candidate: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete candidate"})
				return
			}
		}

		if err := tx.Delete(&position).Error; err != nil {
			tx.Rollback()
			log.Printf("Failed to delete position: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete position"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Failed to commit transaction: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Positions deleted successfully"})
}
