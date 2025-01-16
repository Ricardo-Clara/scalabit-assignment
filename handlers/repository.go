package handlers

import (
	"context"
	"net/http"
	"time"

	"github-api-service/models"
	"github-api-service/services"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v45/github"
)

type RepositoryHandler struct {
    githubService *services.GithubService
}

func NewRepositoryHandler() *RepositoryHandler {
    return &RepositoryHandler{
        githubService: services.NewGithubService(),
    }
}

func (h *RepositoryHandler) CreateRepository(c *gin.Context) {
    var req models.RepoRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    token := c.GetString("github_token")
    client := h.githubService.GetClient(token)
    
    repo := &github.Repository{
        Name:        github.String(req.Name),
        Description: github.String(req.Description),
        Private:     github.Bool(req.Private),
    }
    
    ctx := context.Background()
    newRepo, _, err := client.Repositories.Create(ctx, "", repo)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

	response := models.CreateRepoResponse{
        Message: "Repository created successfully",
        Repository: models.RepositoryInfo{
            ID:          newRepo.GetID(),
            Name:        newRepo.GetName(),
            FullName:    newRepo.GetFullName(),
            Description: newRepo.GetDescription(),
            Private:     newRepo.GetPrivate(),
            HTMLURL:     newRepo.GetHTMLURL(),
            CreatedAt:   newRepo.GetCreatedAt().Format(time.RFC3339),
        },
    }
    
    c.JSON(http.StatusCreated, response)
}

func (h *RepositoryHandler) ListRepositories(c *gin.Context) {
    token := c.GetString("github_token")
    client := h.githubService.GetClient(token)
    
    ctx := context.Background()
    opts := &github.RepositoryListOptions{
        ListOptions: github.ListOptions{PerPage: 100},
		Type: "owner",
    }
    
    repos, _, err := client.Repositories.List(ctx, "", opts)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    formattedRepos := make([]models.RepoRequest, 0, len(repos))
    for _, repo := range repos {
        formattedRepos = append(formattedRepos, models.RepoRequest{
            Name:        repo.GetName(),
            Description: repo.GetDescription(),
            Private:     repo.GetPrivate(),
        })
    }
    
    c.JSON(http.StatusOK, formattedRepos)
}

func (h *RepositoryHandler) DeleteRepository(c *gin.Context) {
    owner := c.Param("owner")
    repo := c.Param("repo")
    
    token := c.GetString("github_token")
    client := h.githubService.GetClient(token)
    
    ctx := context.Background()
    _, err := client.Repositories.Delete(ctx, owner, repo)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
	response := models.DeleteRepoResponse{
        Message: "Repository deleted successfully",
        Details: models.RepoDetails{
            Owner: owner,
            Name:  repo,
        },
    }
    
    c.JSON(http.StatusOK, response)
}

func (h *RepositoryHandler) ListOpenPullRequests(c *gin.Context) {
    token := c.GetString("github_token")
    owner := c.Param("owner")  // Get owner from request parameters
    repo := c.Param("repo")    // Get repo name from request parameters

    client := h.githubService.GetClient(token)
    ctx := context.Background()

    opts := &github.PullRequestListOptions{
        State: "open", // Fetch only open pull requests
        ListOptions: github.ListOptions{PerPage: 100},
    }

    pullRequests, _, err := client.PullRequests.List(ctx, owner, repo, opts)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    formattedPRs := make([]models.PullRequestResponse, 0, len(pullRequests))
    for _, pr := range pullRequests {
        formattedPRs = append(formattedPRs, models.PullRequestResponse{
            Title:       pr.GetTitle(),
            Number:      pr.GetNumber(),
            User:        pr.GetUser().GetLogin(),
            CreatedAt:   pr.GetCreatedAt(),
            HtmlURL:     pr.GetHTMLURL(),
        })
    }

    c.JSON(http.StatusOK, formattedPRs)
}

