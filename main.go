package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"strings"

	"github.com/enzoforreal/mtn-momo-api/momo"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ClientConfig struct {
	SubscriptionKey string
	Environment     string
	ApiKey          string
	ApiUserID       string
}

func NewClient(config ClientConfig) *momo.Client {
	return momo.NewClient(config.SubscriptionKey, config.ApiKey, config.ApiUserID, config.Environment)
}

func main() {
	clientConfig1 := ClientConfig{
		SubscriptionKey: "0285a68a2e9542ae8fb41d6512172362", // Remplacez par votre clé d'abonnement
		Environment:     "sandbox",
		ApiKey:          "b5f50a3e93b64ad4bca4793d4531cc29",     // Remplacez par votre clé d'API
		ApiUserID:       "46680c23-5cb8-4f6e-8f75-aecaa6a7d415", // Remplacez par votre ApiUserID (reference_id)
	}

	router := gin.Default()

	router.POST("/create-api-user", func(c *gin.Context) {
		client := NewClient(clientConfig1)
		var req struct {
			ReferenceID  string `json:"reference_id"`
			CallbackHost string `json:"callback_host"`
		}
		if err := c.BindJSON(&req); err != nil {
			log.Printf("Error binding JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.ReferenceID == "" {
			req.ReferenceID = uuid.New().String()
		}

		log.Printf("Creating API user with reference ID %s and callback host %s", req.ReferenceID, req.CallbackHost)
		err := client.CreateAPIUser(req.ReferenceID, req.CallbackHost)
		if err != nil {
			log.Printf("Error creating API user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Println("API user created successfully")
		c.JSON(http.StatusCreated, gin.H{"message": "API user created successfully", "reference_id": req.ReferenceID})
	})

	router.POST("/create-api-key", func(c *gin.Context) {
		client := NewClient(clientConfig1)
		var req struct {
			ReferenceID string `json:"reference_id"`
		}
		if err := c.BindJSON(&req); err != nil {
			log.Printf("Error binding JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.ReferenceID == "" {
			log.Printf("Reference ID is missing")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Reference ID is required"})
			return
		}

		log.Printf("Creating API key for reference ID %s", req.ReferenceID)
		apiKey, err := client.CreateAPIKey(req.ReferenceID)
		if err != nil {
			log.Printf("Error creating API key: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Println("API key created successfully")
		c.JSON(http.StatusCreated, gin.H{"api_key": apiKey})
	})

	router.POST("/get-auth-token", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("Authorization header missing")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header missing"})
			return
		}

		decodedAuth, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authHeader, "Basic "))
		if err != nil {
			log.Println("Failed to decode authorization header:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header"})
			return
		}

		authParts := strings.SplitN(string(decodedAuth), ":", 2)
		if len(authParts) != 2 {
			log.Println("Invalid authorization format")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization format"})
			return
		}

		apiUserID := authParts[0]
		apiKey := authParts[1]
		log.Printf("Received API User ID: %s, API Key: %s\n", apiUserID, apiKey)

		// Utiliser la fonction GetAuthToken de la bibliothèque momo
		client := NewClient(clientConfig1)
		authToken, err := client.GetAuthToken()
		if err != nil {
			log.Printf("Error getting auth token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Println("Token retrieved successfully")
		c.JSON(http.StatusOK, gin.H{"token": authToken.AccessToken, "expires_in": authToken.ExpiresIn})
	})

	router.GET("/get-account-balance", func(c *gin.Context) {
		client := NewClient(clientConfig1)
		token := c.GetHeader("Authorization")
		if token == "" {
			log.Println("Authorization header missing")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header missing"})
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")
		balance, err := client.GetAccountBalance(token)
		if err != nil {
			log.Printf("Problem parsing balance ( décodage JSON ): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Println("Account balance retrieved successfully")
		c.JSON(http.StatusOK, gin.H{"balance": balance})
	})

	router.Run(":8080")

}
