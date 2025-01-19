package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github-api-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v68/github"
)

// GitHubMock represents a mock implementation of a GitHub client
// MockError allows us to mock an api failure
type GitHubMock struct {
	MockError     error
	RepositoryList []*github.Repository  
	PRList         []*github.PullRequest 
}

// Mock of CreateRepository handler function
func (g *GitHubMock) CreateRepository(c *gin.Context) {
    if g.MockError != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": g.MockError.Error()})
        return
    }

	var repoRequest models.RepoRequest
	if err := c.ShouldBindJSON(&repoRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

    newRepo := &github.Repository{
		Name:        github.Ptr(repoRequest.Name),
		Description: github.Ptr(repoRequest.Description),
		Private:     github.Ptr(repoRequest.Private),
	}
	g.RepositoryList = append(g.RepositoryList, newRepo)

	response := models.RepoResponse{
		Message:    "Successfully created repository",
		Name:       newRepo.GetName(),
		Description: newRepo.GetDescription(),
		Private:    newRepo.GetPrivate(),
	}

	c.JSON(http.StatusCreated, response)
}

// Mock of ListRepositories handler function
func (g *GitHubMock) ListRepositories(c *gin.Context) {
    if g.MockError != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": g.MockError.Error()})
		return
	}

	c.JSON(http.StatusOK, g.RepositoryList)
}

// Mock of DeleteRepository handler function
func (g *GitHubMock) DeleteRepository(c *gin.Context) {
    if g.MockError != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": g.MockError.Error()})
		return
	}

    repoName := c.Param("repo")
	for i, repo := range g.RepositoryList {
		if repo.GetName() == repoName {
			g.RepositoryList = append(g.RepositoryList[:i], g.RepositoryList[i+1:]...)
			response := models.DeleteRepoResponse{Message: "Repository successfully deleted"}
			c.JSON(http.StatusOK, response)
			return
		}
	}

	response := models.DeleteRepoResponse{Message: "Repository not found"}
	c.JSON(http.StatusNotFound, response)
}

// Mock of ListOpenPullRequests handler function
func (g *GitHubMock) ListOpenPullRequests(c *gin.Context) {
    if g.MockError != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": g.MockError.Error()})
		return
	}

	repoName := c.Param("repo")
	limitParam := c.DefaultQuery("limit", "0")
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value"})
		return
	}

	// Check if repository exists
	repoExists := false
    for _, repo := range g.RepositoryList {
        if repo.GetName() == repoName {
            repoExists = true
            break
        }
    }

	if !repoExists {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Repository '%s' does not exist", repoName)})
		return
	}

	var formattedPRs []models.PullRequestResponse
    for _, pr := range g.PRList {
        if pr.GetBase().GetRepo().GetName() == repoName {
            formattedPRs = append(formattedPRs, models.PullRequestResponse{
                Title:     pr.GetTitle(),
                Number:    pr.GetNumber(),
                User:      pr.GetUser().GetLogin(),
                CreatedAt: pr.GetCreatedAt().Time,
                HtmlURL:   pr.GetHTMLURL(),
            })
        }
    }

	if limit > 0 && limit < len(formattedPRs) {
		formattedPRs = formattedPRs[:limit]
	}

	c.JSON(http.StatusOK, formattedPRs)
}

