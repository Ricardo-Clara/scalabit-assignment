package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github-api-service/middleware"
	"github-api-service/routes"
)

func main() {
    r := gin.Default()
    
    // Add middleware to all routes
    r.Use(middleware.AuthMiddleware())
    
    // Setup routes
    routes.SetupRoutes(r)
    
    log.Fatal(r.Run(":8080"))
}