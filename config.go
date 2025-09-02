package main

import (
	"fmt" // Add this import
	"log"
	"os"

	"github.com/joho/godotenv"
)

// EnvConfig holds all environment variables
type EnvConfig struct {
	JWKSUrl         string
	AuthTokenName   string
	TokenRefreshUrl string
	TokenRefreshKey string
	CookieDomain    string
}

// LoadConfig loads environment variables and returns a Config struct
func LoadConfig() *EnvConfig {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found or could not be loaded")
	}

	return &EnvConfig{
		JWKSUrl:         os.Getenv("JWKS_URL"),
		AuthTokenName:   os.Getenv("AUTH_TOKEN_NAME"),
		TokenRefreshUrl: os.Getenv("TOKEN_REFRESH_URL"),
		TokenRefreshKey: os.Getenv("TOKEN_REFRESH_KEY"),
		CookieDomain:    os.Getenv("COOKIE_DOMAIN"),
	}
}

// Validate checks if all required environment variables are set
func (c *EnvConfig) Validate() error {
	if c.JWKSUrl == "" {
		return fmt.Errorf("JWKS_URL environment variable is required")
	}
	if c.AuthTokenName == "" {
		return fmt.Errorf("AUTH_TOKEN_NAME environment variable is required")
	}
	if c.TokenRefreshUrl == "" {
		return fmt.Errorf("TOKEN_REFRESH_URL environment variable is required")
	}
	if c.TokenRefreshKey == "" {
		return fmt.Errorf("TOKEN_REFRESH_KEY environment variable is required")
	}
	if c.CookieDomain == "" {
		return fmt.Errorf("COOKIE_DOMAIN environment variable is required")
	}
	return nil
}
