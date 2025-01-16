package middleware

import (
    "github.com/gin-gonic/gin"
    "net/http"
	"encoding/json"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "No authorization token provided"})
            c.Abort()
            return
        }
        
        // Remove "Bearer " prefix if present
        if len(token) > 7 && token[:7] == "Bearer " {
            token = token[7:]
        }
        
        // Validate the token with GitHub API
        if !isValidGitHubToken(token) {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or badly formed GitHub token"})
            c.Abort()
            return
        }
        
        c.Set("github_token", token)
        c.Next()
    }
}

func isValidGitHubToken(token string) bool {
    url := "https://api.github.com/user"
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return false
    }
    
    req.Header.Set("Authorization", "token "+token)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return false
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return false
    }
    
    var result map[string]interface{}
    if json.NewDecoder(resp.Body).Decode(&result) != nil {
        return false
    }
    
    // Check if the response contains a "login" field, which indicates a valid user
    _, ok := result["login"]
    return ok
}
