package controller

import (
    "cv-extractor/config"
    "cv-extractor/models"
    "cv-extractor/utils"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "net/http"
    "time"
    "mime/multipart"
)

type EditCandidateInput struct {
    Name     string `json:"name" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Domicile string `json:"domicile" binding:"required"`
}

type ScoreCandidateInput struct {
    ID     uint   `json:"id" binding:"required"`
    Score  int    `json:"score" binding:"required"`
    Skills string `json:"skills" binding:"required"`
}

type QualifyCandidateInput struct {
    IDs []uint `json:"ids" binding:"required"`
}

type DeleteCandidateInput struct {
    IDs []uint `json:"ids" binding:"required"`
}

type CreateCandidateInput struct {
    Name       string `form:"name" binding:"required"`
    Email      string `form:"email" binding:"required,email"`
    Domicile   string `form:"domicile" binding:"required"`
    PositionID uint   `form:"positionId" binding:"required"`
    CVFile     *multipart.FileHeader `form:"cv_file" binding:"required"`
}

type CandidateFilterInput struct {
    DepartmentID uint `json:"departmentId"`
    PositionID   uint `json:"positionId"`
}


func CreateCandidate(c *gin.Context) {
    var input CreateCandidateInput

    if err := c.ShouldBind(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid candidate details", "details": err.Error()})
        return
    }

    file, err := c.FormFile("cv_file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "CV file is required", "details": err.Error()})
        return
    }

    fileURL, err := utils.UploadFileToFirebase(file)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to upload CV file", "details": err.Error()})
        return
    }

    var existingCandidate models.Candidate
    if err := config.DB.Where("email = ? AND position_id = ?", input.Email, input.PositionID).First(&existingCandidate).Error; err == nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "Candidate already exists"})
        return
    }

    var position models.Position
    if err := config.DB.First(&position, input.PositionID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Position does not exist"})
        return
    }

    userClaims := c.MustGet("claims").(*utils.Claims)
    var department models.Department
    if err := config.DB.First(&department, position.DepartmentID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Department does not exist"})
        return
    }

    if department.CompanyID != userClaims.CompanyID {
        c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to create a candidate for this position"})
        return
    }

    newCandidate := models.Candidate{
        Name:        input.Name,
        Email:       input.Email,
        Domicile:    input.Domicile,
        PositionID:  input.PositionID,
        CVFile:      fileURL,
        CreatedDate: time.Now(),
    }

    if err := config.DB.Create(&newCandidate).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create candidate"})
        return
    }

    if err := config.DB.Preload("Position").First(&newCandidate, newCandidate.ID).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to load position data", "error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Candidate created successfully", "candidate": newCandidate})
}

func GetAllCandidates(c *gin.Context) {
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

    var positionIDs []uint
    for _, pos := range positions {
        positionIDs = append(positionIDs, pos.ID)
    }

    var candidates []models.Candidate
    if err := config.DB.Preload("Position").Where("position_id IN (?)", positionIDs).Find(&candidates).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Candidates do not exist"})
        return
    }

    c.JSON(http.StatusOK, candidates)
}

func GetOneCandidate(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    id := c.Param("id")
    var candidate models.Candidate
    if err := config.DB.Preload("Position").First(&candidate, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Candidate does not exist"})
        return
    }

    var position models.Position
    if err := config.DB.First(&position, candidate.PositionID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Position does not exist"})
        return
    }

    var department models.Department
    if err := config.DB.First(&department, position.DepartmentID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Department does not exist"})
        return
    }

    if department.CompanyID != userClaims.CompanyID {
        c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this candidate"})
        return
    }

    c.JSON(http.StatusOK, candidate)
}

func GetCandidatesByPosition(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    positionID := c.Param("positionId")

    var position models.Position
    if err := config.DB.First(&position, positionID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Position does not exist"})
        return
    }

    var department models.Department
    if err := config.DB.First(&department, position.DepartmentID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Department does not exist"})
        return
    }

    if department.CompanyID != userClaims.CompanyID {
        c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to view candidates for this position"})
        return
    }

    var candidates []models.Candidate
    if err := config.DB.Preload("Position").Preload("Position.Department").Where("position_id = ?", positionID).Find(&candidates).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "No candidates found for this position"})
        return
    }

    c.JSON(http.StatusOK, candidates)
}

func GetCandidatesByFilters(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    var input CandidateFilterInput

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
        return
    }

    var query = config.DB.Preload("Position").Joins("JOIN positions ON candidates.position_id = positions.id").Joins("JOIN departments ON positions.department_id = departments.id").Where("departments.company_id = ?", userClaims.CompanyID).Where("positions.is_archive = ?", false)

    if input.DepartmentID != 0 {
        query = query.Where("positions.department_id = ?", input.DepartmentID)
    }

    if input.PositionID != 0 {
        query = query.Where("position_id = ?", input.PositionID)
    }

    var candidates []models.Candidate
    if err := query.Find(&candidates).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Candidates do not exist"})
        return
    }

    c.JSON(http.StatusOK, candidates)
}

func GetArchivedCandidatesByFilters(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    var input CandidateFilterInput

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
        return
    }

    var query = config.DB.Preload("Position").Joins("JOIN positions ON candidates.position_id = positions.id").Joins("JOIN departments ON positions.department_id = departments.id").Where("departments.company_id = ?", userClaims.CompanyID).Where("positions.is_archive = ?", true)

    if input.DepartmentID != 0 {
        query = query.Where("positions.department_id = ?", input.DepartmentID)
    }

    if input.PositionID != 0 {
        query = query.Where("position_id = ?", input.PositionID)
    }

    var candidates []models.Candidate
    if err := query.Find(&candidates).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Candidates do not exist"})
        return
    }

    c.JSON(http.StatusOK, candidates)
}

func EditCandidate(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    id := c.Param("id")
    var input EditCandidateInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
        return
    }

    var candidate models.Candidate
    if err := config.DB.First(&candidate, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Candidate does not exist"})
        return
    }

    var position models.Position
    if err := config.DB.First(&position, candidate.PositionID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Position does not exist"})
        return
    }

    var department models.Department
    if err := config.DB.First(&department, position.DepartmentID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Department does not exist"})
        return
    }

    if department.CompanyID != userClaims.CompanyID {
        c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to edit this candidate"})
        return
    }

    candidate.Name = input.Name
    candidate.Email = input.Email
    candidate.Domicile = input.Domicile

    if err := config.DB.Save(&candidate).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update candidate"})
        return
    }

    if err := config.DB.Preload("Position").First(&candidate, id).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to load position data", "error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Candidate updated successfully", "candidate": candidate})
}

func ScoreCandidate(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    var scores []ScoreCandidateInput
    if err := c.ShouldBindJSON(&scores); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
        return
    }

    for _, scoreData := range scores {
        var candidate models.Candidate
        if err := config.DB.First(&candidate, scoreData.ID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"message": "Candidate does not exist"})
            return
        }

        var position models.Position
        if err := config.DB.First(&position, candidate.PositionID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"message": "Position does not exist"})
            return
        }

        var department models.Department
        if err := config.DB.First(&department, position.DepartmentID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"message": "Department does not exist"})
            return
        }

        if department.CompanyID != userClaims.CompanyID {
            c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to score this candidate"})
            return
        }

        isQualified := scoreData.Score > 0

        if (candidate.Score == 0 && isQualified) || (candidate.Score > 0 && !isQualified) {
            var updateValue int
            if isQualified {
                updateValue = 1
            } else {
                updateValue = -1
            }

            if err := config.DB.Model(&models.Position{}).Where("id = ?", candidate.PositionID).Update("filtered_cv", gorm.Expr("filtered_cv + ?", updateValue)).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update position"})
                return
            }
        }

        candidate.Score = scoreData.Score
        candidate.Skills = scoreData.Skills

        if err := config.DB.Save(&candidate).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update candidate"})
            return
        }
    }

    c.JSON(http.StatusOK, gin.H{"message": "Scores updated successfully"})
}

func QualifyCandidate(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    var input QualifyCandidateInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
        return
    }

    for _, id := range input.IDs {
        var candidate models.Candidate
        if err := config.DB.First(&candidate, id).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"message": "Candidate does not exist"})
            return
        }

        var position models.Position
        if err := config.DB.First(&position, candidate.PositionID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"message": "Position does not exist"})
            return
        }

        var department models.Department
        if err := config.DB.First(&department, position.DepartmentID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"message": "Department does not exist"})
            return
        }

        if department.CompanyID != userClaims.CompanyID {
            c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to qualify this candidate"})
            return
        }

        candidate.IsQualified = !candidate.IsQualified

        if err := config.DB.Save(&candidate).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update candidate"})
            return
        }
    }

    c.JSON(http.StatusOK, gin.H{"message": "Qualified status changed successfully"})
}

func DeleteCandidate(c *gin.Context) {
    userClaims := c.MustGet("claims").(*utils.Claims)
    var input DeleteCandidateInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
        return
    }

    tx := config.DB.Begin()
    if tx.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
        return
    }

    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete candidates"})
        }
    }()

    for _, id := range input.IDs {
        var candidate models.Candidate
        if err := tx.First(&candidate, id).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusNotFound, gin.H{"message": "Candidate does not exist"})
            return
        }

        var position models.Position
        if err := config.DB.First(&position, candidate.PositionID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"message": "Position does not exist"})
            return
        }

        var department models.Department
        if err := config.DB.First(&department, position.DepartmentID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"message": "Department does not exist"})
            return
        }

        if department.CompanyID != userClaims.CompanyID {
            tx.Rollback()
            c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to delete this candidate"})
            return
        }

        if err := tx.Delete(&candidate).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete candidate"})
            return
        }
    }

    if err := tx.Commit().Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Candidates deleted successfully"})
}

func GetAll(c *gin.Context) {
    var candidates []models.Candidate
    if err := config.DB.Find(&candidates).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve candidates"})
        return
    }
    c.JSON(http.StatusOK, candidates)
}

func DeleteAll(c *gin.Context) {
    if err := config.DB.Where("1 = 1").Delete(&models.Candidate{}).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete all candidates"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "All candidates deleted successfully"})
}
