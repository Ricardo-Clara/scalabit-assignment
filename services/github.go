package services

import (
    "github.com/google/go-github/v45/github"
    "golang.org/x/oauth2"
    "context"
)

type GithubService struct{}

func NewGithubService() *GithubService {
    return &GithubService{}
}

func (s *GithubService) GetClient(token string) *github.Client {
    ctx := context.Background()
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: token},
    )
    tc := oauth2.NewClient(ctx, ts)
    return github.NewClient(tc)
}