package handlers

import (
	"context"
	"errors"
	_"fmt"
	"net/http"
	"os"
	"strconv"

	"github-api-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v68/github"
	_"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

// Use an interface for ease of testing with mocking
type ApplicationInterface interface {
    CreateRepository(c *gin.Context)
    DeleteRepository(c *gin.Context)
    ListRepositories(c *gin.Context)
    ListOpenPullRequests(c *gin.Context)
}

// Github service wrapper
type Application struct {
    githubClient *github.Client
    owner string
}

// ApplicationInterface wrapper for dependency injection
type Client struct {
    App ApplicationInterface
}

// GetClientForTest returns a mock client to facilitate testing
func GetClientForTest(mockClient ApplicationInterface) *Client {
    return &Client{ App: mockClient }
}

// GetClient initializes a GitHub client using OAuth authentication
func GetClient() (*Client, error) {
    // Use this if running without minikube
	// err := godotenv.Load("config.env")
	// if err != nil {
	// 	fmt.Println("Warning: Could not load .env file. Using system environment variables.")
	// }
	
    // Load authentication details
	token := os.Getenv("TOKEN")
	owner := os.Getenv("OWNER")

	if token == "" {
		return nil, errors.New("missing token")
	}
	if owner == "" {
		return nil, errors.New("missing owner")
	}

    // Create a client with the access token
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	application := &Application{
        githubClient: github.NewClient(tc),
        owner: owner,
    }

	return &Client{ App: application }, nil
}

// CreateRepository handles the creation of a new GitHub repository
func (a *Application) CreateRepository(c *gin.Context) {
    var req models.RepoRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Construct a GitHub repository object from the request
    repo := &github.Repository{
        Name:        github.Ptr(req.Name),
        Description: github.Ptr(req.Description),
        Private:     github.Ptr(req.Private),
    }
    
    ctx := context.Background()
    newRepo, _, err := a.githubClient.Repositories.Create(ctx, "", repo)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Return the created repository details
	response := models.RepoResponse{
        Message: "Repository created successfully",
        Name: newRepo.GetName(),
        Description: newRepo.GetDescription(),
        Private: newRepo.GetPrivate(),
    }
    
    c.JSON(http.StatusCreated, response)
}

// ListRepositories retrieves all repositories owned by the authenticated user
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
    
    // Convert the GitHub response into a simplified format
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

// DeleteRepository removes a repository from the authenticated user's GitHub
func (a *Application) DeleteRepository(c *gin.Context) {
    repo := c.Param("repo")
    
    ctx := context.Background()
    _, err := a.githubClient.Repositories.Delete(ctx, a.owner, repo)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Return success message after deletion
	response := models.DeleteRepoResponse{
        Message: "Repository deleted successfully",
    }
    
    c.JSON(http.StatusOK, response)
}

// ListOpenPullRequests fetches open PRs for a given repository
func (a *Application) ListOpenPullRequests(c *gin.Context) {
    repo := c.Param("repo") // Get repository name from URL parameter

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

    // Convert PR data into a simplified format
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

