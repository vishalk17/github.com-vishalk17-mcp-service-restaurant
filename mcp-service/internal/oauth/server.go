package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/vishalk17/mcp-service-restaurant/internal/config"
	"github.com/vishalk17/mcp-service-restaurant/internal/models"
)

// AuthorizationCode represents an authorization code
type AuthorizationCode struct {
	Code        string
	ClientID    string
	RedirectURI string
	Scope       string
	UserInfo    *models.UserInfo
	ExpiresAt   time.Time
}

// Server handles OAuth 2.0 operations
type Server struct {
	config         *config.Config
	provider       *Provider
	storage        *Storage
	tokenManager   *TokenManager
	clientRegistry *ClientRegistry
	authCodes      map[string]*AuthorizationCode
	authCodesMux   sync.RWMutex
}

// NewServer creates a new OAuth server
func NewServer(cfg *config.Config, storage *Storage) *Server {
	provider := NewProvider(cfg.OAuth)
	tokenManager := NewTokenManager(
		cfg.Server.JWTSecret,
		cfg.Server.OAuthServerURL,
		cfg.Server.AccessTokenLife,
		cfg.Server.RefreshTokenLife,
		storage,
	)
	clientRegistry := NewClientRegistry(storage)

	return &Server{
		config:         cfg,
		provider:       provider,
		storage:        storage,
		tokenManager:   tokenManager,
		clientRegistry: clientRegistry,
		authCodes:      make(map[string]*AuthorizationCode),
	}
}

// HandleAuthorize handles the authorization endpoint
func (s *Server) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	// Parse parameters
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	responseType := r.URL.Query().Get("response_type")
	scope := r.URL.Query().Get("scope")
	state := r.URL.Query().Get("state")

	if scope == "" {
		scope = "openid profile email"
	}

	// Validate parameters
	if clientID == "" || redirectURI == "" || responseType == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Validate client
	valid, err := s.clientRegistry.ValidateClient(clientID)
	if err != nil || !valid {
		http.Error(w, "Invalid client_id", http.StatusBadRequest)
		return
	}

	// Validate redirect URI
	validURI, err := s.clientRegistry.ValidateClientRedirectURI(clientID, redirectURI)
	if err != nil || !validURI {
		http.Error(w, "Invalid redirect_uri", http.StatusBadRequest)
		return
	}

	// Only support authorization code flow
	if responseType != "code" {
		s.redirectWithError(w, r, redirectURI, "unsupported_response_type", "Only code response type is supported", state)
		return
	}

	// Store OAuth request state
	stateData := map[string]string{
		"client_id":    clientID,
		"redirect_uri": redirectURI,
		"scope":        scope,
		"state":        state,
	}
	stateJSON, _ := json.Marshal(stateData)
	encodedState := base64.URLEncoding.EncodeToString(stateJSON)

	// Redirect to identity provider
	authURL := s.provider.GetAuthorizationURL(encodedState)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// HandleCallback handles OAuth provider callback
func (s *Server) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	if errorParam != "" {
		http.Error(w, fmt.Sprintf("OAuth error: %s", errorParam), http.StatusBadRequest)
		return
	}

	if code == "" || state == "" {
		http.Error(w, "Missing code or state", http.StatusBadRequest)
		return
	}

	// Decode state
	stateData, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	var stateMap map[string]string
	if err := json.Unmarshal(stateData, &stateMap); err != nil {
		http.Error(w, "Invalid state format", http.StatusBadRequest)
		return
	}

	clientID := stateMap["client_id"]
	redirectURI := stateMap["redirect_uri"]
	scope := stateMap["scope"]
	originalState := stateMap["state"]

	// Exchange code for token with provider
	token, err := s.provider.ExchangeCodeForToken(r.Context(), code)
	if err != nil {
		log.Printf("Failed to exchange code: %v", err)
		s.redirectWithError(w, r, redirectURI, "access_denied", "Failed to authenticate", originalState)
		return
	}

	// Get user info from provider
	userInfo, err := s.provider.GetUserInfo(r.Context(), token.AccessToken)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		s.redirectWithError(w, r, redirectURI, "access_denied", "Failed to get user info", originalState)
		return
	}

	// Check if user email exists in database (email whitelist)
	user, err := s.storage.FindUserByEmail(userInfo.Email)
	if err != nil {
		log.Printf("Database error: %v", err)
		s.redirectWithError(w, r, redirectURI, "server_error", "Internal error", originalState)
		return
	}

	if user == nil {
		// User not in database - reject
		log.Printf("Unauthorized user attempt: %s", userInfo.Email)
		s.redirectWithError(w, r, redirectURI, "access_denied", "User not authorized", originalState)
		return
	}

	// Update user provider info if needed
	if user.Provider == nil || user.ProviderUserID == nil || *user.Provider == "" || *user.ProviderUserID == "" {
		s.storage.UpdateUserProvider(user.UserID, s.provider.GetProviderName(), userInfo.Sub, userInfo.Name, userInfo.Picture)
	}

	// Update last login
	s.storage.UpdateLastLogin(user.UserID)

	// Generate authorization code
	authCode := generateAuthCode()
	s.authCodesMux.Lock()
	s.authCodes[authCode] = &AuthorizationCode{
		Code:        authCode,
		ClientID:    clientID,
		RedirectURI: redirectURI,
		Scope:       scope,
		UserInfo:    userInfo,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}
	s.authCodesMux.Unlock()

	// Redirect back to client with code
	redirectURL, _ := url.Parse(redirectURI)
	q := redirectURL.Query()
	q.Set("code", authCode)
	if originalState != "" {
		q.Set("state", originalState)
	}
	redirectURL.RawQuery = q.Encode()

	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}

// HandleToken handles token endpoint
func (s *Server) HandleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	grantType := r.FormValue("grant_type")

	switch grantType {
	case "authorization_code":
		s.handleAuthorizationCodeGrant(w, r)
	case "refresh_token":
		s.handleRefreshTokenGrant(w, r)
	default:
		s.jsonError(w, "unsupported_grant_type", "Grant type not supported", http.StatusBadRequest)
	}
}

// handleAuthorizationCodeGrant handles authorization code grant
func (s *Server) handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	clientID := r.FormValue("client_id")

	if code == "" || clientID == "" {
		s.jsonError(w, "invalid_request", "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Get and validate authorization code
	s.authCodesMux.Lock()
	authCode, exists := s.authCodes[code]
	if exists {
		delete(s.authCodes, code) // Single use
	}
	s.authCodesMux.Unlock()

	if !exists {
		s.jsonError(w, "invalid_grant", "Invalid authorization code", http.StatusBadRequest)
		return
	}

	if time.Now().After(authCode.ExpiresAt) {
		s.jsonError(w, "invalid_grant", "Authorization code expired", http.StatusBadRequest)
		return
	}

	if authCode.ClientID != clientID {
		s.jsonError(w, "invalid_grant", "Client mismatch", http.StatusBadRequest)
		return
	}

	// Find user by email
	user, err := s.storage.FindUserByEmail(authCode.UserInfo.Email)
	if err != nil || user == nil {
		s.jsonError(w, "invalid_grant", "User not found", http.StatusBadRequest)
		return
	}

	// Create tokens
	tokens, err := s.tokenManager.CreateTokens(user, clientID, authCode.Scope)
	if err != nil {
		log.Printf("Failed to create tokens: %v", err)
		s.jsonError(w, "server_error", "Failed to create tokens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

// handleRefreshTokenGrant handles refresh token grant
func (s *Server) handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.FormValue("refresh_token")

	if refreshToken == "" {
		s.jsonError(w, "invalid_request", "Missing refresh_token", http.StatusBadRequest)
		return
	}

	// Refresh tokens
	tokens, err := s.tokenManager.RefreshToken(refreshToken, s.storage)
	if err != nil {
		log.Printf("Failed to refresh token: %v", err)
		s.jsonError(w, "invalid_grant", "Invalid refresh token", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

// HandleRegister handles client registration
func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "invalid_request", "Invalid JSON", http.StatusBadRequest)
		return
	}

	client, err := s.clientRegistry.RegisterClient(req)
	if err != nil {
		log.Printf("Failed to register client: %v", err)
		s.jsonError(w, "invalid_request", err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(client)
}

// HandleUserInfo handles userinfo endpoint
func (s *Server) HandleUserInfo(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		s.jsonError(w, "invalid_token", "Missing or invalid authorization header", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := s.tokenManager.ValidateToken(token)
	if err != nil {
		s.jsonError(w, "invalid_token", "Invalid access token", http.StatusUnauthorized)
		return
	}

	userInfo := map[string]interface{}{
		"sub":     claims["sub"],
		"email":   claims["email"],
		"name":    claims["name"],
		"picture": claims["picture"],
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

// HandleIntrospect handles token introspection
func (s *Server) HandleIntrospect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	token := r.FormValue("token")

	result, _ := s.tokenManager.IntrospectToken(token)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleRevoke handles token revocation
func (s *Server) HandleRevoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	token := r.FormValue("token")

	s.tokenManager.RevokeToken(token)

	w.WriteHeader(http.StatusOK)
}

// HandleOAuthMetadata handles OAuth authorization server metadata
func (s *Server) HandleOAuthMetadata(w http.ResponseWriter, r *http.Request) {
	baseURL := s.config.Server.OAuthServerURL

	metadata := map[string]interface{}{
		"issuer":                            baseURL,
		"authorization_endpoint":            baseURL + "/oauth/authorize",
		"token_endpoint":                    baseURL + "/oauth/token",
		"registration_endpoint":             baseURL + "/oauth/register",
		"jwks_uri":                          baseURL + "/.well-known/jwks.json",
		"userinfo_endpoint":                 baseURL + "/oauth/userinfo",
		"revocation_endpoint":               baseURL + "/oauth/revoke",
		"introspection_endpoint":            baseURL + "/oauth/introspect",
		"response_types_supported":          []string{"code"},
		"grant_types_supported":             []string{"authorization_code", "refresh_token"},
		"token_endpoint_auth_methods_supported": []string{"none", "client_secret_post"},
		"scopes_supported":                  []string{"openid", "profile", "email"},
		"subject_types_supported":           []string{"public"},
		"id_token_signing_alg_values_supported": []string{"HS256"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}

// HandleJWKS handles JWKS endpoint
func (s *Server) HandleJWKS(w http.ResponseWriter, r *http.Request) {
	jwks := s.tokenManager.GetJWKS()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jwks)
}

// Helper functions

func (s *Server) redirectWithError(w http.ResponseWriter, r *http.Request, redirectURI, errorCode, errorDesc, state string) {
	u, _ := url.Parse(redirectURI)
	q := u.Query()
	q.Set("error", errorCode)
	q.Set("error_description", errorDesc)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
}

func (s *Server) jsonError(w http.ResponseWriter, errorCode, errorDesc string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":             errorCode,
		"error_description": errorDesc,
	})
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func generateAuthCode() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// GetTokenManager returns the token manager
func (s *Server) GetTokenManager() *TokenManager {
	return s.tokenManager
}
