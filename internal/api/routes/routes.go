package routes

import (
    "github.com/gin-gonic/gin"
    "github-api-service/internal/api/handlers"
)

func SetupRoutes(r *gin.Engine, client handlers.Client) {
    r.POST("/repositories",  client.App.CreateRepository)
	r.GET("/repositories/:repo/pull-requests", client.App.ListOpenPullRequests)
    r.GET("/repositories", client.App.ListRepositories)
    r.DELETE("/repositories/:repo", client.App.DeleteRepository)
}