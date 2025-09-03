package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const userContextKey contextKey = "user"

func authMiddleware(next HTTPHandlerFunc) HTTPHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// Load configuration
		config := LoadConfig()
		if err := config.Validate(); err != nil {
			log.Fatal(err)
		}

		// Define excluded path patterns that should bypass authentication
		excludedPatterns := []string{
			// `^/api/health$`,   // Health check endpoint
			`^/favicon\.ico$`, // Favicon
			`^/robots\.txt$`,  // Robots.txt (optional)
		}

		p := r.URL.Path

		// Check if path matches any excluded pattern
		for _, pattern := range excludedPatterns {
			matched, err := regexp.MatchString(pattern, p)
			if err != nil {
				log.Printf("Error matching pattern %s: %v", pattern, err)
				continue
			}
			if matched {
				log.Printf("Skipping auth for excluded path: %s (pattern: %s)", p, pattern)
				return next(w, r)
			}
		}

		// Fetch and cache JWKS keys
		jwks, err := fetchJWKS(config.JWKSUrl)
		if err != nil {
			log.Printf("Failed to fetch JWKS: %v", err)
			return InternalServerError("Failed to fetch JWKS")
		}

		// Read auth-token cookie
		authCookie, err := r.Cookie(config.AuthTokenName)
		if err != nil {
			log.Printf("Auth token cookie not found: %v", err)
			return Unauthorized("Authentication required")
		}
		cookieValue := authCookie.Value
		cookieValue = strings.TrimPrefix(cookieValue, "base64-")

		// Decode base64 to bytes
		decodedBytes, err := base64.StdEncoding.DecodeString(cookieValue)
		if err != nil {
			log.Printf("Failed to decode base64: %v", err)
			return BadRequest("Invalid token format")
		}

		// Parse JSON
		var tokenData map[string]interface{}
		err = json.Unmarshal(decodedBytes, &tokenData)
		if err != nil {
			log.Printf("Failed to parse JSON: %v", err)
			return BadRequest("Invalid token JSON")
		}

		accessToken, ok := tokenData["access_token"].(string)
		if !ok {
			log.Printf("Access token not found or invalid type")
			return BadRequest("Invalid access token")
		}

		// Verify the JWT token
		token, err := verifyJWT(accessToken, jwks)
		if err != nil {
			log.Printf("Failed to verify JWT: %v", err)

			// Check if we have a refresh token to try refreshing
			if refreshToken, ok := tokenData["refresh_token"].(string); ok && refreshToken != "" {
				log.Printf("Attempting to refresh access token...")

				refreshResp, refreshErr := refreshAccessToken(refreshToken, config.TokenRefreshUrl, config.TokenRefreshKey)
				if refreshErr != nil {
					log.Printf("Failed to refresh token: %v", refreshErr)
					return Unauthorized("Token refresh failed")
				}

				// Update token data with new values
				tokenData["access_token"] = refreshResp.AccessToken
				tokenData["refresh_token"] = refreshResp.RefreshToken
				tokenData["expires_at"] = refreshResp.ExpiresAt
				tokenData["user"] = refreshResp.User

				// Update the cookie with new token data
				if err := updateAuthCookie(w, tokenData, config.AuthTokenName, config.CookieDomain); err != nil {
					log.Printf("Failed to update auth cookie: %v", err)
				}

				// Try verifying the new access token
				token, err = verifyJWT(refreshResp.AccessToken, jwks)
				if err != nil {
					log.Printf("Failed to verify refreshed JWT: %v", err)
					return Unauthorized("Invalid refreshed token")
				}

				log.Printf("Successfully refreshed and verified token")
			} else {
				return Unauthorized("Invalid token and no refresh token available")
			}
		}

		// Extract claims from verified token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Printf("Failed to extract claims from token")
			return Unauthorized("Invalid token claims")
		}

		log.Printf("Authenticated request to %s with user ID: %v", p, claims["sub"])

		// Inject User id into request context
		ctx := context.WithValue(r.Context(), userContextKey, claims["sub"])
		return next(w, r.WithContext(ctx))
	}
}
