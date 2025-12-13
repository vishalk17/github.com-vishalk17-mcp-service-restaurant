package oauth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vishalk17/mcp-service-restaurant/internal/models"
)

// Storage handles OAuth database operations
type Storage struct {
	db *sql.DB
}

// NewStorage creates a new OAuth storage
func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

// ============================================
// User Operations
// ============================================

// FindUserByEmail finds a user by email
func (s *Storage) FindUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, user_id, email, name, picture, provider, provider_user_id, 
		       status, role, created_at, last_login_at, updated_at
		FROM user_profiles
		WHERE email = $1 AND status = 'active'
	`
	
	user := &models.User{}
	err := s.db.QueryRow(query, email).Scan(
		&user.ID, &user.UserID, &user.Email, &user.Name, &user.Picture,
		&user.Provider, &user.ProviderUserID, &user.Status, &user.Role,
		&user.CreatedAt, &user.LastLoginAt, &user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	
	return user, nil
}

// UpdateUserProvider updates user provider information
func (s *Storage) UpdateUserProvider(userID, provider, providerUserID, name, picture string) error {
	query := `
		UPDATE user_profiles
		SET provider = $1, provider_user_id = $2, name = $3, picture = $4, updated_at = NOW()
		WHERE user_id = $5
	`
	
	_, err := s.db.Exec(query, provider, providerUserID, name, picture, userID)
	if err != nil {
		return fmt.Errorf("failed to update user provider: %w", err)
	}
	
	return nil
}

// UpdateLastLogin updates user's last login timestamp
func (s *Storage) UpdateLastLogin(userID string) error {
	query := `UPDATE user_profiles SET last_login_at = NOW() WHERE user_id = $1`
	_, err := s.db.Exec(query, userID)
	return err
}

// ============================================
// Client Operations
// ============================================

// CreateClient creates a new OAuth client
func (s *Storage) CreateClient(client *models.OAuthClient) error {
	redirectURIsJSON, _ := json.Marshal(client.RedirectURIs)
	grantTypesJSON, _ := json.Marshal(client.GrantTypes)
	responseTypesJSON, _ := json.Marshal(client.ResponseTypes)
	
	query := `
		INSERT INTO oauth_clients (
			client_id, client_secret, client_name, client_uri, logo_uri,
			redirect_uris, grant_types, response_types, scope, application_type,
			token_endpoint_auth_method, client_secret_expires_at, active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`
	
	err := s.db.QueryRow(
		query,
		client.ClientID, client.ClientSecret, client.ClientName, client.ClientURI,
		client.LogoURI, redirectURIsJSON, grantTypesJSON, responseTypesJSON,
		client.Scope, client.ApplicationType, client.TokenEndpointAuthMethod,
		client.ClientSecretExpiresAt, client.Active,
	).Scan(&client.ID, &client.CreatedAt, &client.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	
	return nil
}

// GetClient retrieves a client by client ID
func (s *Storage) GetClient(clientID string) (*models.OAuthClient, error) {
	query := `
		SELECT id, client_id, client_secret, client_name, client_uri, logo_uri,
		       redirect_uris, grant_types, response_types, scope, application_type,
		       token_endpoint_auth_method, created_at, updated_at,
		       client_secret_expires_at, active
		FROM oauth_clients
		WHERE client_id = $1 AND active = true
	`
	
	client := &models.OAuthClient{}
	var redirectURIsJSON, grantTypesJSON, responseTypesJSON []byte
	
	err := s.db.QueryRow(query, clientID).Scan(
		&client.ID, &client.ClientID, &client.ClientSecret, &client.ClientName,
		&client.ClientURI, &client.LogoURI, &redirectURIsJSON, &grantTypesJSON,
		&responseTypesJSON, &client.Scope, &client.ApplicationType,
		&client.TokenEndpointAuthMethod, &client.CreatedAt, &client.UpdatedAt,
		&client.ClientSecretExpiresAt, &client.Active,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}
	
	// Unmarshal JSON arrays
	json.Unmarshal(redirectURIsJSON, &client.RedirectURIs)
	json.Unmarshal(grantTypesJSON, &client.GrantTypes)
	json.Unmarshal(responseTypesJSON, &client.ResponseTypes)
	
	return client, nil
}

// ValidateClientRedirectURI checks if redirect URI is valid for client
func (s *Storage) ValidateClientRedirectURI(clientID, redirectURI string) (bool, error) {
	client, err := s.GetClient(clientID)
	if err != nil {
		return false, err
	}
	if client == nil {
		return false, nil
	}
	
	for _, uri := range client.RedirectURIs {
		if uri == redirectURI {
			return true, nil
		}
	}
	
	return false, nil
}

// ============================================
// Token Operations
// ============================================

// SaveTokenMetadata saves token metadata for revocation
func (s *Storage) SaveTokenMetadata(token *models.OAuthToken) error {
	query := `
		INSERT INTO oauth_tokens (
			token_id, client_id, user_id, token_type, scope, expires_at, active
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	_, err := s.db.Exec(
		query,
		token.TokenID, token.ClientID, token.UserID, token.TokenType,
		token.Scope, token.ExpiresAt, token.Active,
	)
	
	if err != nil {
		return fmt.Errorf("failed to save token metadata: %w", err)
	}
	
	return nil
}

// GetTokenMetadata retrieves token metadata
func (s *Storage) GetTokenMetadata(tokenID string) (*models.OAuthToken, error) {
	query := `
		SELECT id, token_id, client_id, user_id, token_type, scope,
		       expires_at, created_at, active
		FROM oauth_tokens
		WHERE token_id = $1
	`
	
	token := &models.OAuthToken{}
	err := s.db.QueryRow(query, tokenID).Scan(
		&token.ID, &token.TokenID, &token.ClientID, &token.UserID,
		&token.TokenType, &token.Scope, &token.ExpiresAt,
		&token.CreatedAt, &token.Active,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get token metadata: %w", err)
	}
	
	return token, nil
}

// RevokeToken marks a token as inactive
func (s *Storage) RevokeToken(tokenID string) error {
	query := `UPDATE oauth_tokens SET active = false WHERE token_id = $1`
	_, err := s.db.Exec(query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	return nil
}

// CleanupExpiredTokens removes expired token metadata
func (s *Storage) CleanupExpiredTokens() error {
	query := `DELETE FROM oauth_tokens WHERE expires_at < NOW()`
	result, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}
	
	rows, _ := result.RowsAffected()
	if rows > 0 {
		fmt.Printf("Cleaned up %d expired tokens\n", rows)
	}
	
	return nil
}

// IsTokenRevoked checks if a token has been revoked
func (s *Storage) IsTokenRevoked(tokenID string) (bool, error) {
	token, err := s.GetTokenMetadata(tokenID)
	if err != nil {
		return true, err
	}
	if token == nil {
		return false, nil // Token not in database, not revoked
	}
	
	// Check if expired
	if time.Now().After(token.ExpiresAt) {
		return true, nil
	}
	
	return !token.Active, nil
}
