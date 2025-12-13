package models

import "time"

// User represents a user profile from OAuth providers
type User struct {
	ID             int        `json:"id"`
	UserID         string     `json:"user_id"`           // UUID
	Email          string     `json:"email"`             // Primary identifier
	Name           string     `json:"name"`
	Picture        *string    `json:"picture,omitempty"` // Nullable
	Provider       *string    `json:"provider,omitempty"` // Nullable
	ProviderUserID *string    `json:"provider_user_id,omitempty"` // Nullable
	Status         string     `json:"status"`            // active, inactive, suspended
	Role           string     `json:"role"`              // admin, user
	CreatedAt      time.Time  `json:"created_at"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// OAuthClient represents a registered OAuth client (DCR)
type OAuthClient struct {
	ID                      int       `json:"id"`
	ClientID                string    `json:"client_id"`
	ClientSecret            string    `json:"client_secret"`
	ClientName              string    `json:"client_name"`
	ClientURI               string    `json:"client_uri,omitempty"`
	LogoURI                 string    `json:"logo_uri,omitempty"`
	RedirectURIs            []string  `json:"redirect_uris"`
	GrantTypes              []string  `json:"grant_types"`
	ResponseTypes           []string  `json:"response_types"`
	Scope                   string    `json:"scope"`
	ApplicationType         string    `json:"application_type"`
	TokenEndpointAuthMethod string    `json:"token_endpoint_auth_method"`
	CreatedAt               time.Time `json:"-"`
	UpdatedAt               time.Time `json:"-"`
	ClientIDIssuedAt        int64     `json:"client_id_issued_at"`
	ClientSecretExpiresAt   int64     `json:"client_secret_expires_at"`
	Active                  bool      `json:"active"`
}

// OAuthToken represents token metadata for revocation
type OAuthToken struct {
	ID        int       `json:"id"`
	TokenID   string    `json:"token_id"`
	ClientID  string    `json:"client_id"`
	UserID    string    `json:"user_id"`
	TokenType string    `json:"token_type"` // access_token, refresh_token
	Scope     string    `json:"scope"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Active    bool      `json:"active"`
}

// TokenResponse represents OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// UserInfo represents user information from OAuth provider
type UserInfo struct {
	Sub           string      `json:"sub"`
	Email         string      `json:"email"`
	EmailVerified interface{} `json:"email_verified,omitempty"` // Can be bool or string
	Name          string      `json:"name"`
	Picture       string      `json:"picture,omitempty"`
	GivenName     string      `json:"given_name,omitempty"`
	FamilyName    string      `json:"family_name,omitempty"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	Sub       string `json:"sub"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Picture   string `json:"picture,omitempty"`
	ClientID  string `json:"client_id"`
	Scope     string `json:"scope"`
	TokenType string `json:"token_type"` // access_token, refresh_token
	TokenID   string `json:"token_id"`
	Iat       int64  `json:"iat"`
	Exp       int64  `json:"exp"`
	Issuer    string `json:"iss"`
}
