package models

import "time"

type RepoRequest struct {
    Name        string `json:"name" binding:"required"`
    Description string `json:"description"`
    Private     bool   `json:"private"`
}

type CreateRepoResponse struct {
    Message    string         `json:"message"`
    Repository RepositoryInfo `json:"repository"`
}

type RepositoryInfo struct {
    ID          int64  `json:"id"`
    Name        string `json:"name"`
    FullName    string `json:"full_name"`
    Description string `json:"description"`
    Private     bool   `json:"private"`
    HTMLURL     string `json:"html_url"`
    CreatedAt   string `json:"created_at"`
}

type DeleteRepoResponse struct {
    Message  string `json:"message"`
    Details  RepoDetails `json:"details"`
}

type RepoDetails struct {
    Owner string `json:"owner"`
    Name  string `json:"name"`
}

type PullRequestResponse struct {
    Title     string    `json:"title"`
    Number    int       `json:"number"`
    User      string    `json:"user"`
    CreatedAt time.Time `json:"created_at"`
    HtmlURL   string    `json:"html_url"`
}

