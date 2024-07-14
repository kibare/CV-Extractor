package routes

import (
    "cv-extractor/controller"
    "cv-extractor/middleware"
    "github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
    r := gin.Default()

    // Global middleware
    r.Use(middleware.RequestLogger())
    r.Use(middleware.RecoveryMiddleware())
    r.Use(middleware.CORSMiddleware())

    // Public routes
    publicRoutes(r)

    // Authenticated routes
    authRoutes(r)

    return r
}

func publicRoutes(r *gin.Engine) {
    r.POST("/api/auth/login", controller.Login)
    r.POST("/api/auth/register", controller.Register)
    r.GET("/api/company/get-all-company", controller.GetAllCompanies)
}

func authRoutes(r *gin.Engine) {
    auth := r.Group("/")
    auth.Use(middleware.AuthMiddleware())
    {
        companyRoutes(auth)
        positionRoutes(auth)
        userRoutes(auth)
        candidateRoutes(auth)
        departmentRoutes(auth)
    }
}

func companyRoutes(r *gin.RouterGroup) {
    r.POST("/api/company/create-company", controller.CreateCompany)
    r.GET("/api/company/get-one-company/:id", controller.GetOneCompany)
    r.PUT("/api/company/edit-company/:id", controller.EditCompany)
    r.DELETE("/api/company/delete-company/:id", controller.DeleteCompany)
}

func positionRoutes(r *gin.RouterGroup) {
    r.POST("/api/position/create-position", controller.CreatePosition)
    r.GET("/api/position/get-all-positions", controller.GetAllPositions)
    r.GET("/api/position/get-one-position/:id", controller.GetOnePosition)
    r.PUT("/api/position/edit-position/:id", controller.EditPosition)
    r.DELETE("/api/position/delete-position/:id", controller.DeletePosition)
    r.PUT("/api/position/archive-position/:id", controller.ArchivePosition)
    r.GET("/api/position/get-archived-positions", controller.GetArchivedPositions)
    r.PUT("/api/position/trash-position/:id", controller.TrashPosition)

}

func userRoutes(r *gin.RouterGroup) {
    r.GET("/api/user/get-user", controller.GetUser)
    r.PUT("/api/user/edit-user", controller.EditUser)
    r.PUT("/api/user/change-password", controller.ChangePassword)
    r.DELETE("/api/user/delete-user", controller.DeleteUser)
    r.GET("/api/user/get-all-users", controller.GetAllUsers)
}

func candidateRoutes(r *gin.RouterGroup) {
    r.POST("/api/candidate/create-candidate", controller.CreateCandidate)
    r.GET("/api/candidate/get-all-candidates", controller.GetAllCandidates)
    r.GET("/api/candidate/get-one-candidate/:id", controller.GetOneCandidate)
    r.PUT("/api/candidate/edit-candidate/:id", controller.EditCandidate)
    r.PUT("/api/candidate/score-candidate/:id", controller.ScoreCandidate)
    r.PUT("/api/candidate/qualify-candidate/:id", controller.QualifyCandidate)
    r.DELETE("/api/candidate/delete-candidate/:id", controller.DeleteCandidate)
    r.POST("/api/candidate/get-candidates-by-filters", controller.GetCandidatesByFilters)
    r.POST("/api/candidate/get-archived-candidates-by-filters", controller.GetArchivedCandidatesByFilters)

}

func departmentRoutes(r *gin.RouterGroup) {
    r.POST("/api/department/create-department", controller.CreateDepartment)
    r.GET("/api/department/get-all-departments", controller.GetAllDepartments)
    r.GET("/api/department/get-one-department/:id", controller.GetOneDepartment)
    r.PUT("/api/department/edit-department/:id", controller.EditDepartment)
    r.DELETE("/api/department/delete-department/:id", controller.DeleteDepartment)
}
