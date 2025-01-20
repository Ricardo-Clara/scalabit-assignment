# GitHub API Service

A REST API service that interacts with GitHub to manage repositories and pull requests.

## Environment Variables

Create a `config.env` file in the root directory with the following content:

```env
TOKEN=your_github_personal_access_token
OWNER=your_github_username
```

## Installation

1. Clone the repository
2. Install dependencies:

```
go mod tidy
```

## Running Locally

To run the service locally you can uncomment the code in handlers/repository.go in GetClient() to load the token and owner from config.env and then run:
```
go run cmd/main.go
```
The server will start on port 8080.

## API Endpoints

- Create Repository 
```
POST /repositories
Content-Type: application/json

{
    "name": "repo-name",
    "description": "Repository description", // Optional
    "private": false // Optional, defaults to false
}
```
- List Repositories
```
GET /repositories
```
- Delete Repository
```
DELETE /repositories/:repo
```
- List Open Pull Requests
```
GET /repositories/:repo/pull-requests?limit=0 // limit is an optional parameter
```

## Minikube Deployment

1. Create the secrets:
```
./scripts/create_secret.sh
```
2. Build the image
```
docker build -t <user>/github-api-service:latest .
```
3. Apply kubernetes manifests:
```
kubectl apply -f kubernetes
```

## CI/CD Pipeline

The project includes a Github Actions workflows that:

- Runs tests
- Performs linting
- Runs security scans
- Deploys locally to Minikube