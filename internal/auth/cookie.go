package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

// UpdateAuthCookie updates the auth cookie with new token data
func UpdateAuthCookie(w http.ResponseWriter, newTokenData map[string]interface{}, authTokenName, cookieDomain string) error {
	// Encode the new token data to JSON and then base64
	jsonData, err := json.Marshal(newTokenData)
	if err != nil {
		return fmt.Errorf("failed to marshal new token data: %v", err)
	}

	encodedData := base64.StdEncoding.EncodeToString(jsonData)
	cookieValue := "base64-" + encodedData

	// Set the new cookie
	cookie := &http.Cookie{
		Name:     authTokenName,
		Value:    cookieValue,
		Path:     "/",
		Domain:   cookieDomain,
		HttpOnly: true,
		Secure:   true, // Set to false for development if using HTTP
		SameSite: http.SameSiteLaxMode,
	}

	w.Header().Set("Set-Cookie", cookie.String())
	return nil
}
