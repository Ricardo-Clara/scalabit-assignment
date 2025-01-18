package models

import "time"

type RepoRequest struct {
    Name        string `json:"name" binding:"required"`
    Description string `json:"description"`
    Private     bool   `json:"private"`
}

type RepoResponse struct {
    Message     string `json:"message"`
    Name        string `json:"name" binding:"required"`
    Description string `json:"description"`
    Private     bool   `json:"private"`
}

type DeleteRepoResponse struct {
    Message  string `json:"message"`
}

type PullRequestResponse struct {
    Title     string    `json:"title"`
    Number    int       `json:"number"`
    User      string    `json:"login"`
    CreatedAt time.Time `json:"created_at"`
    HtmlURL   string    `json:"html_url"`
}

