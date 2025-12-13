package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vishalk17/mcp-service-restaurant/internal/models"
)

// ClientRegistry handles OAuth client registration (RFC 7591)
type ClientRegistry struct {
	storage *Storage
}

// NewClientRegistry creates a new client registry
func NewClientRegistry(storage *Storage) *ClientRegistry {
	return &ClientRegistry{storage: storage}
}

// RegisterClient registers a new OAuth client
func (cr *ClientRegistry) RegisterClient(req map[string]interface{}) (*models.OAuthClient, error) {
	now := time.Now()
	client := &models.OAuthClient{
		ClientID:                "mcp-" + uuid.New().String(),
		ClientSecret:            generateSecureSecret(),
		ClientName:              getStringOrDefault(req, "client_name", "MCP Client"),
		ClientURI:               getStringOrDefault(req, "client_uri", ""),
		LogoURI:                 getStringOrDefault(req, "logo_uri", ""),
		RedirectURIs:            getStringArrayOrDefault(req, "redirect_uris", getDefaultRedirectURIs()),
		GrantTypes:              getStringArrayOrDefault(req, "grant_types", []string{"authorization_code", "refresh_token"}),
		ResponseTypes:           getStringArrayOrDefault(req, "response_types", []string{"code"}),
		Scope:                   getStringOrDefault(req, "scope", "openid profile email"),
		ApplicationType:         getStringOrDefault(req, "application_type", "web"),
		TokenEndpointAuthMethod: getStringOrDefault(req, "token_endpoint_auth_method", "none"),
		CreatedAt:               now,
		UpdatedAt:               now,
		ClientIDIssuedAt:        now.Unix(),
		ClientSecretExpiresAt:   0, // Never expires
		Active:                  true,
	}

	// Validate
	if err := cr.validateClient(client); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save to database
	if err := cr.storage.CreateClient(client); err != nil {
		return nil, fmt.Errorf("failed to save client: %w", err)
	}

	return client, nil
}

// GetClient retrieves a client
func (cr *ClientRegistry) GetClient(clientID string) (*models.OAuthClient, error) {
	return cr.storage.GetClient(clientID)
}

// ValidateClient checks if client exists and is active
func (cr *ClientRegistry) ValidateClient(clientID string) (bool, error) {
	client, err := cr.GetClient(clientID)
	if err != nil {
		return false, err
	}
	return client != nil && client.Active, nil
}

// ValidateClientRedirectURI checks if redirect URI is valid for client
func (cr *ClientRegistry) ValidateClientRedirectURI(clientID, redirectURI string) (bool, error) {
	return cr.storage.ValidateClientRedirectURI(clientID, redirectURI)
}

// validateClient validates client data
func (cr *ClientRegistry) validateClient(client *models.OAuthClient) error {
	if client.ClientName == "" {
		return fmt.Errorf("client_name is required")
	}
	if len(client.RedirectURIs) == 0 {
		return fmt.Errorf("at least one redirect_uri is required")
	}
	if len(client.GrantTypes) == 0 {
		return fmt.Errorf("at least one grant_type is required")
	}
	if len(client.ResponseTypes) == 0 {
		return fmt.Errorf("at least one response_type is required")
	}
	return nil
}

// generateSecureSecret generates a secure client secret
func generateSecureSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// getDefaultRedirectURIs returns default redirect URIs for ChatGPT and Claude
func getDefaultRedirectURIs() []string {
	return []string{
		"https://chatgpt.com/connector_platform_oauth_redirect",
		"https://chatgpt.com/aip/c/o/redirect",
		"http://localhost:3000/callback",
	}
}

// Helper functions for parsing request data
func getStringOrDefault(req map[string]interface{}, key, defaultValue string) string {
	if val, ok := req[key].(string); ok && val != "" {
		return val
	}
	return defaultValue
}

func getStringArrayOrDefault(req map[string]interface{}, key string, defaultValue []string) []string {
	if val, ok := req[key].([]interface{}); ok {
		result := make([]string, 0, len(val))
		for _, v := range val {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}
