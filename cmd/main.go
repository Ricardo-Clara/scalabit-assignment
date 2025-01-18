package main

import (
	"log"

	"github.com/gin-gonic/gin"

    "github-api-service/internal/api/handlers"
	"github-api-service/internal/api/routes"
)



func main() {
    r := gin.Default()
    
    client, err := handlers.GetClient()
    if err != nil {
        log.Fatal(err)
    }

    routes.SetupRoutes(r, *client)
    
    log.Fatal(r.Run(":8080"))
}