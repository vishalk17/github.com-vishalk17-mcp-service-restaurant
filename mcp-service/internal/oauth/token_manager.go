package oauth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/vishalk17/mcp-service-restaurant/internal/models"
)

// TokenManager handles JWT token operations
type TokenManager struct {
	jwtSecret        []byte
	issuer           string
	accessTokenLife  int64
	refreshTokenLife int64
	storage          *Storage
}

// NewTokenManager creates a new token manager
func NewTokenManager(jwtSecret string, issuer string, accessLife, refreshLife int64, storage *Storage) *TokenManager {
	return &TokenManager{
		jwtSecret:        []byte(jwtSecret),
		issuer:           issuer,
		accessTokenLife:  accessLife,
		refreshTokenLife: refreshLife,
		storage:          storage,
	}
}

// CreateTokens creates access and refresh tokens for a user
func (tm *TokenManager) CreateTokens(user *models.User, clientID, scope string) (*models.TokenResponse, error) {
	now := time.Now()
	accessTokenID := uuid.New().String()
	refreshTokenID := uuid.New().String()

	// Create access token
	accessClaims := jwt.MapClaims{
		"sub":        user.UserID,
		"email":      user.Email,
		"name":       user.Name,
		"picture":    user.Picture,
		"client_id":  clientID,
		"scope":      scope,
		"token_type": "access_token",
		"token_id":   accessTokenID,
		"iat":        now.Unix(),
		"exp":        now.Add(time.Duration(tm.accessTokenLife) * time.Second).Unix(),
		"iss":        tm.issuer,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(tm.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Create refresh token
	refreshClaims := jwt.MapClaims{
		"sub":        user.UserID,
		"email":      user.Email,
		"client_id":  clientID,
		"scope":      scope,
		"token_type": "refresh_token",
		"token_id":   refreshTokenID,
		"iat":        now.Unix(),
		"exp":        now.Add(time.Duration(tm.refreshTokenLife) * time.Second).Unix(),
		"iss":        tm.issuer,
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(tm.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Save token metadata for revocation tracking
	accessTokenMeta := &models.OAuthToken{
		TokenID:   accessTokenID,
		ClientID:  clientID,
		UserID:    user.UserID,
		TokenType: "access_token",
		Scope:     scope,
		ExpiresAt: time.Unix(accessClaims["exp"].(int64), 0),
		Active:    true,
	}
	if err := tm.storage.SaveTokenMetadata(accessTokenMeta); err != nil {
		// Log but don't fail - token is still valid
		fmt.Printf("Warning: failed to save access token metadata: %v\n", err)
	}

	refreshTokenMeta := &models.OAuthToken{
		TokenID:   refreshTokenID,
		ClientID:  clientID,
		UserID:    user.UserID,
		TokenType: "refresh_token",
		Scope:     scope,
		ExpiresAt: time.Unix(refreshClaims["exp"].(int64), 0),
		Active:    true,
	}
	if err := tm.storage.SaveTokenMetadata(refreshTokenMeta); err != nil {
		fmt.Printf("Warning: failed to save refresh token metadata: %v\n", err)
	}

	return &models.TokenResponse{
		AccessToken:  accessTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    tm.accessTokenLife,
		RefreshToken: refreshTokenString,
		Scope:        scope,
	}, nil
}

// ValidateToken validates a JWT token and returns claims
func (tm *TokenManager) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("token expired")
		}
	}

	// Check issuer
	if iss, ok := claims["iss"].(string); ok {
		if iss != tm.issuer {
			return nil, fmt.Errorf("invalid issuer")
		}
	}

	// Check if token is revoked
	if tokenID, ok := claims["token_id"].(string); ok {
		revoked, err := tm.storage.IsTokenRevoked(tokenID)
		if err != nil {
			fmt.Printf("Warning: failed to check token revocation: %v\n", err)
		} else if revoked {
			return nil, fmt.Errorf("token has been revoked")
		}
	}

	return claims, nil
}

// RefreshToken creates new tokens from a refresh token
func (tm *TokenManager) RefreshToken(refreshTokenString string, storage *Storage) (*models.TokenResponse, error) {
	claims, err := tm.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	tokenType, _ := claims["token_type"].(string)
	if tokenType != "refresh_token" {
		return nil, fmt.Errorf("not a refresh token")
	}

	email, _ := claims["email"].(string)
	clientID, _ := claims["client_id"].(string)
	scope, _ := claims["scope"].(string)

	// Find user
	user, err := storage.FindUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Revoke old refresh token
	if tokenID, ok := claims["token_id"].(string); ok {
		tm.storage.RevokeToken(tokenID)
	}

	// Create new tokens
	return tm.CreateTokens(user, clientID, scope)
}

// RevokeToken revokes a token
func (tm *TokenManager) RevokeToken(tokenString string) error {
	claims, err := tm.ValidateToken(tokenString)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	if tokenID, ok := claims["token_id"].(string); ok {
		return tm.storage.RevokeToken(tokenID)
	}

	return fmt.Errorf("token does not have token_id")
}

// IntrospectToken introspects a token and returns its metadata
func (tm *TokenManager) IntrospectToken(tokenString string) (map[string]interface{}, error) {
	claims, err := tm.ValidateToken(tokenString)
	if err != nil {
		return map[string]interface{}{"active": false}, nil
	}

	return map[string]interface{}{
		"active":     true,
		"sub":        claims["sub"],
		"client_id":  claims["client_id"],
		"scope":      claims["scope"],
		"token_type": claims["token_type"],
		"exp":        claims["exp"],
		"iat":        claims["iat"],
		"iss":        claims["iss"],
	}, nil
}

// GetJWKS returns JSON Web Key Set for token verification
func (tm *TokenManager) GetJWKS() map[string]interface{} {
	return map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "oct",
				"use": "sig",
				"kid": "main",
				"alg": "HS256",
			},
		},
	}
}
