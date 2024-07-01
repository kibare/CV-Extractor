package main

import (
    "cv-extractor/config"
    "cv-extractor/routes"
    "cv-extractor/utils"
    "log"
    "os"
)

func main() {
    config.InitDB()
    if err := utils.InitFirebase(); err != nil {
        log.Fatalf("Failed to initialize Firebase: %v", err)
    }

    r := routes.SetupRouter()

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    if err := r.Run(":" + port); err != nil {
        log.Fatalf("Failed to run server: %v", err)
    } else {
        log.Printf("Server is running on port %s", port)
    }
}
