package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/vishalk17/mcp-service-restaurant/internal/config"
	"github.com/vishalk17/mcp-service-restaurant/internal/models"
	"golang.org/x/oauth2"
)

// Provider handles OAuth provider interactions
type Provider struct {
	config *config.OAuthConfig
	oauth2Config *oauth2.Config
}

// NewProvider creates a new OAuth provider
func NewProvider(cfg *config.OAuthConfig) *Provider {
	oauth2Cfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURI,
		Scopes:       cfg.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthURL,
			TokenURL: cfg.TokenURL,
		},
	}

	return &Provider{
		config:       cfg,
		oauth2Config: oauth2Cfg,
	}
}

// GetAuthorizationURL generates the OAuth authorization URL
func (p *Provider) GetAuthorizationURL(state string) string {
	return p.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// ExchangeCodeForToken exchanges authorization code for access token
func (p *Provider) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := p.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	return token, nil
}

// GetUserInfo retrieves user information from the provider
func (p *Provider) GetUserInfo(ctx context.Context, accessToken string) (*models.UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.config.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user info request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	userInfo := &models.UserInfo{}
	if err := json.Unmarshal(body, userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	// Handle provider-specific field mappings
	p.normalizeUserInfo(userInfo)

	return userInfo, nil
}

// normalizeUserInfo normalizes provider-specific user info fields
func (p *Provider) normalizeUserInfo(userInfo *models.UserInfo) {
	provider := strings.ToLower(p.config.Provider)

	switch provider {
	case "google":
		// Google uses 'sub' for user ID - already standard
		if userInfo.Sub == "" && userInfo.Email != "" {
			userInfo.Sub = userInfo.Email
		}

	case "microsoft":
		// Microsoft uses 'id' instead of 'sub'
		var data map[string]interface{}
		tempJSON, _ := json.Marshal(userInfo)
		json.Unmarshal(tempJSON, &data)
		
		if id, ok := data["id"].(string); ok && id != "" {
			userInfo.Sub = id
		}
		if displayName, ok := data["displayName"].(string); ok && displayName != "" {
			userInfo.Name = displayName
		}
		if mail, ok := data["mail"].(string); ok && mail != "" {
			userInfo.Email = mail
		}
		if userPrincipalName, ok := data["userPrincipalName"].(string); ok && userPrincipalName != "" && userInfo.Email == "" {
			userInfo.Email = userPrincipalName
		}

	case "cognito":
		// Cognito uses standard OpenID Connect fields
		if userInfo.Sub == "" && userInfo.Email != "" {
			userInfo.Sub = userInfo.Email
		}
	}

	// Fallback: use email as sub if sub is empty
	if userInfo.Sub == "" && userInfo.Email != "" {
		userInfo.Sub = userInfo.Email
	}
}

// ValidateState validates the OAuth state parameter (CSRF protection)
func (p *Provider) ValidateState(receivedState, expectedState string) bool {
	return receivedState == expectedState
}

// GetProviderName returns the provider name
func (p *Provider) GetProviderName() string {
	return p.config.Provider
}
