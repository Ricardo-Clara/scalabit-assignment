package main

import (
	"log"

	"github.com/gin-gonic/gin"

    "github-api-service/internal/api/handlers"
	"github-api-service/internal/api/routes"
)



func main() {
    // Gin router
    r := gin.Default()
    
    // Get github client
    client, err := handlers.GetClient()
    if err != nil {
        log.Fatal(err)
    }

    // Setup routes and handler functions in gin router
    routes.SetupRoutes(r, *client)
    
    log.Fatal(r.Run(":8080"))
}