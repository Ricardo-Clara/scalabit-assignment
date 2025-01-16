package routes

import (
    "github.com/gin-gonic/gin"
    "github-api-service/handlers"
)

func SetupRoutes(r *gin.Engine) {
    repoHandler := handlers.NewRepositoryHandler()
    
    // Repository routes
    r.POST("/repositories", repoHandler.CreateRepository)
	r.GET("/repositories/:owner/:repo/pull-requests", repoHandler.ListOpenPullRequests)
    r.GET("/repositories", repoHandler.ListRepositories)
    r.DELETE("/repositories/:owner/:repo", repoHandler.DeleteRepository)
}