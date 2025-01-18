package handlers

import (
	"context"
	"errors"
	_"fmt"
	"os"

	"net/http"

	"strconv"

	"github-api-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v68/github"
	_"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

type ApplicationInterface interface {
    CreateRepository(c *gin.Context)
    DeleteRepository(c *gin.Context)
    ListRepositories(c *gin.Context)
    ListOpenPullRequests(c *gin.Context)
}

type Application struct {
    githubClient *github.Client
    owner string
}

type Client struct {
    App ApplicationInterface
}

func GetClientForTest(mockClient ApplicationInterface) *Client {
    return &Client{ App: mockClient }
}

func GetClient() (*Client, error) {
	// err := godotenv.Load("config.env")
	// if err != nil {
	// 	fmt.Println("Warning: Could not load .env file. Using system environment variables.")
	// }
	
	token := os.Getenv("TOKEN")
	owner := os.Getenv("OWNER")

	if token == "" {
		return nil, errors.New("missing token")
	}
	if owner == "" {
		return nil, errors.New("missing owner")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	application := &Application{
        githubClient: github.NewClient(tc),
        owner: owner,
    }

	return &Client{ App: application }, nil
}

func (a *Application) CreateRepository(c *gin.Context) {
    var req models.RepoRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    repo := &github.Repository{
        Name:        github.Ptr(req.Name),
        Description: github.Ptr(req.Description),
        Private:     github.Ptr(req.Private),
    }
    
    ctx := context.Background()
    newRepo, _, err := a.githubClient.Repositories.Create(ctx, a.owner, repo)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

	response := models.RepoResponse{
        Message: "Repository created successfully",
        Name: newRepo.GetName(),
        Description: newRepo.GetDescription(),
        Private: newRepo.GetPrivate(),
    }
    
    c.JSON(http.StatusCreated, response)
}

func (a *Application) ListRepositories(c *gin.Context) {
    ctx := context.Background()
    opts := &github.RepositoryListByAuthenticatedUserOptions{
        ListOptions: github.ListOptions{PerPage: 100},
		Type: "owner",
    }
    
    repos, _, err := a.githubClient.Repositories.ListByAuthenticatedUser(ctx, opts)
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

func (a *Application) DeleteRepository(c *gin.Context) {
    repo := c.Param("repo")
    
    ctx := context.Background()
    _, err := a.githubClient.Repositories.Delete(ctx, a.owner, repo)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
	response := models.DeleteRepoResponse{
        Message: "Repository deleted successfully",
    }
    
    c.JSON(http.StatusOK, response)
}

func (a *Application) ListOpenPullRequests(c *gin.Context) {
    repo := c.Param("repo")    // Get repo name from request parameters

    ctx := context.Background()
    opts := &github.PullRequestListOptions{
        State: "open", // Fetch only open pull requests
        ListOptions: github.ListOptions{PerPage: 100},
    }

    pullRequests, _, err := a.githubClient.PullRequests.List(ctx, a.owner, repo, opts)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Get the 'limit' query parameter if provided
    limit, err := strconv.Atoi(c.DefaultQuery("limit", "0"))
    if err != nil || limit < 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
        return
    }

    // Apply limit if it's greater than 0 and less than the total number of PRs
    if limit > 0 && limit < len(pullRequests) {
        pullRequests = pullRequests[:limit]
    }

    formattedPRs := make([]models.PullRequestResponse, 0, len(pullRequests))
    for _, pr := range pullRequests {
        formattedPRs = append(formattedPRs, models.PullRequestResponse{
            Title:       pr.GetTitle(),
            Number:      pr.GetNumber(),
            User:        pr.GetUser().GetLogin(),
            CreatedAt:   pr.GetCreatedAt().Time,
            HtmlURL:     pr.GetHTMLURL(),
        })
    }

    c.JSON(http.StatusOK, formattedPRs)
}

