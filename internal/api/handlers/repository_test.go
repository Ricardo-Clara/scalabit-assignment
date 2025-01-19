package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github-api-service/internal/models"
	"github-api-service/internal/api/handlers"
	"github-api-service/internal/api/routes"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
)

var (
	errJSONMarshal   = errors.New("failed to marshal repository request")
    errJSONUnmarshal = errors.New("failed to unmarshal repository response")
    errRequestCreate = errors.New("failed to create HTTP request")
)


func TestCreateRepository(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create new Gin engine
	r := gin.Default()

	mockClient := &handlers.GitHubMock{}
	ghClient := handlers.GetClientForTest(mockClient)
	routes.SetupRoutes(r, *ghClient)

	t.Run("Successful repository creation", func(t *testing.T) {
		request := models.RepoRequest{
			Name: "hello-world",
			Description: "test repository",
			Private: false,
		}

		requestBody, err := json.Marshal(request)
		assert.NoError(t, err, errJSONMarshal)

		req, _ := http.NewRequest("POST", "/repositories", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var response models.RepoResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, errJSONUnmarshal)

		assert.Equal(t, http.StatusCreated, w.Code, `Code should be 201 Created`)
		assert.Equal(t, request.Name, response.Name, "Repository name should match")
    	assert.Equal(t, request.Description, response.Description, "Repository description should match")
    	assert.Equal(t, request.Private, response.Private, "Repository private status should match")
	})

	
	t.Run("Invalid JSON payload", func(t *testing.T) {
		request := `{name: "invalid"}` // Invalid JSON
		req, _ := http.NewRequest("POST", "/repositories", bytes.NewBufferString(request))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code, "Code should be 400 BadRequest")
	})

	t.Run("Missing name in JSON", func(t *testing.T) {
		request := `{"description": "test repo"}`
		req, _ := http.NewRequest("POST", "/repositories", bytes.NewBufferString(request))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code, "Code should be 400 BadRequest")
	})

	t.Run("Github error", func(t *testing.T) {
		r := gin.Default()
		mockClientError := &handlers.GitHubMock{MockError: errors.New("github api error")}
		ghClient := handlers.GetClientForTest(mockClientError)
		routes.SetupRoutes(r, *ghClient)
		
		request := models.RepoRequest{
			Name: "hello-world",
			Description: "test repository",
			Private: false,
		}

		requestBody, err := json.Marshal(request)
		assert.NoError(t, err, errJSONMarshal)

		req, _ := http.NewRequest("POST", "/repositories", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var response models.RepoResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, errJSONUnmarshal)
	
		assert.Equal(t, http.StatusBadRequest, w.Code, "Code should be 400 BadRequest")
	})
}

func TestListRepository(t *testing.T) {
	t.Run("List repositories successfully", func(t *testing.T) {
		// Set up mock client and mock repositories
		mockClient := &handlers.GitHubMock{
			RepositoryList: []*github.Repository{
				&github.Repository{Name: github.Ptr("test-repo"), Description: github.Ptr("test repo"), Private: github.Ptr(false)},
				&github.Repository{Name: github.Ptr("hello-world"), Description: github.Ptr("test repo"), Private: github.Ptr(true)},
			},
		}

		gin.SetMode(gin.TestMode)

		r := gin.Default()
		ghClient := handlers.GetClientForTest(mockClient)
		routes.SetupRoutes(r, *ghClient)
	
		req, err := http.NewRequest("GET", "/repositories", nil)
		assert.NoError(t, err, errRequestCreate)
	
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	
		assert.Equal(t, http.StatusOK, w.Code, "Code should be 200 OK")
	
		var repos []*github.Repository
		err = json.Unmarshal(w.Body.Bytes(), &repos)
		assert.NoError(t, err, errJSONUnmarshal)
	
		assert.Len(t, repos, 2, "There should be 2 repositories listed")
		assert.Equal(t, "test-repo", *repos[0].Name, "First repo name should be 'test-repo'")
		assert.Equal(t, "hello-world", *repos[1].Name, "Second repo name should be 'hello-world'")
	})

	t.Run("Error while listing repositories", func(t *testing.T) {
		// Set up mock client with an error message
		mockClient := &handlers.GitHubMock{MockError: errors.New("mock error")}

		gin.SetMode(gin.TestMode)

		r := gin.Default()
		ghClient := handlers.GetClientForTest(mockClient)
		routes.SetupRoutes(r, *ghClient)
	
		req, err := http.NewRequest("GET", "/repositories", nil)
		assert.NoError(t, err, errRequestCreate)
	
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Code should be 400 BadRequest")
	
		var response map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, errJSONUnmarshal)
	
		assert.Equal(t, "mock error", response["error"], "Error message should be 'mock error'")
	})

	t.Run("Empty repository list", func(t *testing.T) {
		mockClient := &handlers.GitHubMock{
			RepositoryList: []*github.Repository{}, // No repositories
		}

		gin.SetMode(gin.TestMode)

		r := gin.Default()
		ghClient := handlers.GetClientForTest(mockClient)
		routes.SetupRoutes(r, *ghClient)
	
		req, err := http.NewRequest("GET", "/repositories", nil)
		assert.NoError(t, err, errRequestCreate)
	
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	
		assert.Equal(t, http.StatusOK, w.Code, "Code should be 200 OK")
	
		var repos []*github.Repository
		err = json.Unmarshal(w.Body.Bytes(), &repos)
		assert.NoError(t, err, errJSONUnmarshal)
	
		assert.Len(t, repos, 0, "There should be no repositories listed")
	})
	
}

func TestDeleteRepository(t *testing.T) {
	t.Run("Successfully delete repository", func(t *testing.T) {
		mockClient := &handlers.GitHubMock{
			RepositoryList: []*github.Repository{
				{Name: github.Ptr("test-repo")},
				{Name: github.Ptr("hello-world")},
			},
		}

		gin.SetMode(gin.TestMode)
		
		r := gin.Default()
		ghClient := handlers.GetClientForTest(mockClient)
		routes.SetupRoutes(r, *ghClient)
	
		req, err := http.NewRequest("DELETE", "/repositories/test-repo", nil)
		assert.NoError(t, err, errRequestCreate)
	
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	
		assert.Equal(t, http.StatusOK, w.Code, "Code should be 200 OK")
	
		var response models.DeleteRepoResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, errJSONUnmarshal)

		assert.Equal(t, "Repository successfully deleted", response.Message)
		assert.Len(t, mockClient.RepositoryList, 1, "There should only be 1 repository")
		assert.Equal(t, "hello-world", *mockClient.RepositoryList[0].Name, "The remaining repository should be 'hello-world'")
	})

	t.Run("Delete non-existent repository", func(t *testing.T) {
		mockClient := &handlers.GitHubMock{
			RepositoryList: []*github.Repository{
				&github.Repository{Name: github.Ptr("test-repo")},
			},
		}

		gin.SetMode(gin.TestMode)

		r := gin.Default()
		ghClient := handlers.GetClientForTest(mockClient)
		routes.SetupRoutes(r, *ghClient)
	
		req, err := http.NewRequest("DELETE", "/repositories/hello-world", nil)
		assert.NoError(t, err, errRequestCreate)
	
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	
		assert.Equal(t, http.StatusNotFound, w.Code, "Code should be 404 NotFound")
	
		var response models.DeleteRepoResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, errJSONUnmarshal)
	
		assert.Equal(t, "Repository not found", response.Message, "Error message should be 'Repository not found'")
		assert.Len(t, mockClient.RepositoryList, 1, "There should still be 1 repository")
		assert.Equal(t, "test-repo", *mockClient.RepositoryList[0].Name, "Remaining repository should be 'test-repo'")
	})
	
	t.Run("Error while deleting repository", func(t *testing.T) {
		mockClient := &handlers.GitHubMock{MockError: errors.New("mock error")}

		gin.SetMode(gin.TestMode)

		r := gin.Default()
		ghClient := handlers.GetClientForTest(mockClient)
		routes.SetupRoutes(r, *ghClient)
	
		req, err := http.NewRequest("DELETE", "/repositories/test-repo", nil)
		assert.NoError(t, err, errRequestCreate)
	
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Code should be 400 BadRequest")
	
		var response map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, errJSONUnmarshal)

		assert.Equal(t, "mock error", response["error"], "Error message should be 'mock error'")
	})	
}

func TestListOpenPullRequests(t *testing.T) {
	t.Run("Successfully list open pull requests with no limit", func(t *testing.T) {
		mockPRs := []*github.PullRequest{
            {
                Number:    github.Ptr(1),
                Title:     github.Ptr("First PR"),
                State:     github.Ptr("open"),
                HTMLURL:   github.Ptr("https://github.com/test/test-repo/pull/1"),
                CreatedAt: &github.Timestamp{Time: time.Now()},
                User:      &github.User{Login: github.Ptr("testuser")},
                Base: &github.PullRequestBranch{
                    Repo: &github.Repository{
                        Name: github.Ptr("test-repo"),
                    },
                },
            },
            {
                Number:    github.Ptr(2),
                Title:     github.Ptr("Second PR"),
                State:     github.Ptr("open"),
                HTMLURL:   github.Ptr("https://github.com/test/test-repo/pull/2"),
                CreatedAt: &github.Timestamp{Time: time.Now()},
                User:      &github.User{Login: github.Ptr("testuser")},
                Base: &github.PullRequestBranch{
                    Repo: &github.Repository{
                        Name: github.Ptr("test-repo"),
                    },
                },
            },
        }
		
		mockClient := &handlers.GitHubMock{
            RepositoryList: []*github.Repository{
                {Name: github.Ptr("test-repo")}, // Add the repository to RepositoryList
            },
            PRList: mockPRs,
        }

		gin.SetMode(gin.TestMode)

		r := gin.Default()
        ghClient := handlers.GetClientForTest(mockClient)
        routes.SetupRoutes(r, *ghClient)

        req, err := http.NewRequest("GET", "/repositories/test-repo/pull-requests", nil)
        assert.NoError(t, err, errRequestCreate)

        w := httptest.NewRecorder()
        r.ServeHTTP(w, req)

        assert.Equal(t, http.StatusOK, w.Code, "Code should be 200 OK")

        var response []models.PullRequestResponse
        err = json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err, errJSONUnmarshal)

        assert.Len(t, response, 2, "There should be 2 pull requests")

        // Assert first PR fields
        assert.Equal(t, 1, response[0].Number, "First PR number should match")
        assert.Equal(t, "First PR", response[0].Title, "First PR title should match")
        assert.Equal(t, "testuser", response[0].User, "First PR user should match")
        assert.Equal(t, "https://github.com/test/test-repo/pull/1", response[0].HtmlURL, "First PR URL should match")
        assert.Equal(t, mockPRs[0].GetCreatedAt().Time.UTC(), response[0].CreatedAt, "First PR creation time should match")

        // Assert second PR fields
        assert.Equal(t, 2, response[1].Number, "Second PR number should match")
        assert.Equal(t, "Second PR", response[1].Title, "Second PR title should match")
        assert.Equal(t, "testuser", response[1].User, "Second PR user should match")
        assert.Equal(t, "https://github.com/test/test-repo/pull/2", response[1].HtmlURL, "Second PR URL should match")
        assert.Equal(t, mockPRs[1].GetCreatedAt().Time.UTC(), response[1].CreatedAt, "Second PR creation time should match")
    })

	t.Run("Successfully list open pull requests with limit", func(t *testing.T) {
		mockPRs := []*github.PullRequest{
            {
                Number:    github.Ptr(1),
                Title:     github.Ptr("First PR"),
                State:     github.Ptr("open"),
                HTMLURL:   github.Ptr("https://github.com/test/test-repo/pull/1"),
                CreatedAt: &github.Timestamp{Time: time.Now()},
                User:      &github.User{Login: github.Ptr("testuser")},
                Base: &github.PullRequestBranch{
                    Repo: &github.Repository{
                        Name: github.Ptr("test-repo"),
                    },
                },
            },
            {
                Number:    github.Ptr(2),
                Title:     github.Ptr("Second PR"),
                State:     github.Ptr("open"),
                HTMLURL:   github.Ptr("https://github.com/test/test-repo/pull/2"),
                CreatedAt: &github.Timestamp{Time: time.Now()},
                User:      &github.User{Login: github.Ptr("testuser")},
                Base: &github.PullRequestBranch{
                    Repo: &github.Repository{
                        Name: github.Ptr("test-repo"),
                    },
                },
            },
			{
                Number:    github.Ptr(3),
                Title:     github.Ptr("Third PR"),
                State:     github.Ptr("open"),
                HTMLURL:   github.Ptr("https://github.com/test/test-repo/pull/3"),
                CreatedAt: &github.Timestamp{Time: time.Now()},
                User:      &github.User{Login: github.Ptr("testuser")},
                Base: &github.PullRequestBranch{
                    Repo: &github.Repository{
                        Name: github.Ptr("test-repo"),
                    },
                },
            },
        }

		mockClient := &handlers.GitHubMock{
            RepositoryList: []*github.Repository{
                {Name: github.Ptr("test-repo")}, 
            },
            PRList: mockPRs,
        }

		gin.SetMode(gin.TestMode)

		r := gin.Default()
		ghClient := handlers.GetClientForTest(mockClient)
		routes.SetupRoutes(r, *ghClient)
	
		req, err := http.NewRequest("GET", "/repositories/test-repo/pull-requests?limit=2", nil)
		assert.NoError(t, err, errRequestCreate)
	
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	
		assert.Equal(t, http.StatusOK, w.Code, "Code should be 200 OK")
	
		var response []models.PullRequestResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, errJSONUnmarshal)
	
		assert.Len(t, response, 2, "There should be 2 pull requests")

		assert.Equal(t, 1, response[0].Number, "First PR number should match")
		assert.Equal(t, "First PR", response[0].Title, "First PR title should match")
		assert.Equal(t, "testuser", response[0].User, "First PR user should match")
		assert.Equal(t, "https://github.com/test/test-repo/pull/1", response[0].HtmlURL, "First PR URL should match")
		assert.Equal(t, mockPRs[0].GetCreatedAt().Time.UTC().Truncate(time.Second), 
					 response[0].CreatedAt.UTC().Truncate(time.Second), 
					 "First PR creation time should match")
	
		assert.Equal(t, 2, response[1].Number, "Second PR number should match")
		assert.Equal(t, "Second PR", response[1].Title, "Second PR title should match")
		assert.Equal(t, "testuser", response[1].User, "Second PR user should match")
		assert.Equal(t, "https://github.com/test/test-repo/pull/2", response[1].HtmlURL, "Second PR URL should match")
		assert.Equal(t, mockPRs[1].GetCreatedAt().Time.UTC().Truncate(time.Second), 
					 response[1].CreatedAt.UTC().Truncate(time.Second), 
					 "Second PR creation time should match")
	})
	
	t.Run("Invalid limit parameter", func(t *testing.T) {
		mockPRs := []*github.PullRequest{
            {
                Number:    github.Ptr(1),
                Title:     github.Ptr("First PR"),
                State:     github.Ptr("open"),
                HTMLURL:   github.Ptr("https://github.com/test/test-repo/pull/1"),
                CreatedAt: &github.Timestamp{Time: time.Now()},
                User:      &github.User{Login: github.Ptr("testuser")},
                Base: &github.PullRequestBranch{
                    Repo: &github.Repository{
                        Name: github.Ptr("test-repo"),
                    },
                },
            },
		}

		mockClient := &handlers.GitHubMock{
            PRList: mockPRs,
        }

		gin.SetMode(gin.TestMode)

		r := gin.Default()
		ghClient := handlers.GetClientForTest(mockClient)
		routes.SetupRoutes(r, *ghClient)
	
		req, err := http.NewRequest("GET", "/repositories/test-repo/pull-requests?limit=invalid", nil)
		assert.NoError(t, err, errRequestCreate)
	
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	
		assert.Equal(t, http.StatusBadRequest, w.Code, "Code should be 400 BadRequest")
	
		var response map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, errJSONUnmarshal)
	
		assert.Equal(t, "Invalid limit value", response["error"], "Error message should be 'Invalid limit value'")
	})
	
	t.Run("Repository does not exist", func(t *testing.T) {
		mockPRs := []*github.PullRequest{
            {
                Number:    github.Ptr(1),
                Title:     github.Ptr("First PR"),
                State:     github.Ptr("open"),
                HTMLURL:   github.Ptr("https://github.com/test/test-repo/pull/1"),
                CreatedAt: &github.Timestamp{Time: time.Now()},
                User:      &github.User{Login: github.Ptr("testuser")},
                Base: &github.PullRequestBranch{
                    Repo: &github.Repository{
                        Name: github.Ptr("test-repo"),
                    },
                },
            },
		}

		mockClient := &handlers.GitHubMock{
            PRList: mockPRs,
        }

		gin.SetMode(gin.TestMode)

		r := gin.Default()
		ghClient := handlers.GetClientForTest(mockClient)
		routes.SetupRoutes(r, *ghClient)
	
		req, err := http.NewRequest("GET", "/repositories/hello-world/pull-requests", nil)
		assert.NoError(t, err, errRequestCreate)
	
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	
		assert.Equal(t, http.StatusNotFound, w.Code, "Code should be 404 NotFound")
	
		var response map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, errJSONUnmarshal)
	
		assert.Equal(t, "Repository 'hello-world' does not exist", response["error"], "Error message should be: 'Repository 'hello-world' does not exist'")
	})
	
	t.Run("No pull requests for repository", func(t *testing.T) {
		mockPRs := []*github.PullRequest{}
		mockRepo := []*github.Repository{
			&github.Repository{Name: github.Ptr("test-repo")},
		}

		mockClient := &handlers.GitHubMock{
            PRList: mockPRs,
			RepositoryList: mockRepo,
        }

		gin.SetMode(gin.TestMode)

		r := gin.Default()
		ghClient := handlers.GetClientForTest(mockClient)
		routes.SetupRoutes(r, *ghClient)
	
		req, err := http.NewRequest("GET", "/repositories/test-repo/pull-requests", nil)
		assert.NoError(t, err, errRequestCreate)
	
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	
		assert.Equal(t, http.StatusOK, w.Code, "Code should be 200 OK")
	
		var response []models.PullRequestResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, errJSONUnmarshal)
	
		assert.Len(t, response, 0, "There should be 0 pull requests")
	})

	t.Run("Error listing pull requests", func(t *testing.T) {
        mockClient := &handlers.GitHubMock{
            MockError: errors.New("failed to fetch pull requests"),
        }

		gin.SetMode(gin.TestMode)
        
        r := gin.Default()
        ghClient := handlers.GetClientForTest(mockClient)
        routes.SetupRoutes(r, *ghClient)

        req, err := http.NewRequest("GET", "/repositories/test-repo/pull-requests", nil)
        assert.NoError(t, err, errRequestCreate)

        w := httptest.NewRecorder()
        r.ServeHTTP(w, req)

        assert.Equal(t, http.StatusBadRequest, w.Code, "Code should be 400 Bad Request")

        var response map[string]string
        err = json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err, errJSONUnmarshal)

        assert.Equal(t, "failed to fetch pull requests", response["error"], "Error message should match mock error")
    })
}
