package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// OAuthConfig holds OAuth provider configuration
type OAuthConfig struct {
	Provider     string
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
	RedirectURI  string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host              string
	Port              string
	OAuthServerURL    string
	JWTSecret         string
	AccessTokenLife   int64 // in seconds
	RefreshTokenLife  int64 // in seconds
	DefaultAdminEmail string
	DefaultAdminName  string
}

// Config holds all application configuration
type Config struct {
	Database     string
	OAuth        *OAuthConfig
	Server       *ServerConfig
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (optional for local dev)
	_ = godotenv.Load()

	config := &Config{}

	// Database configuration
	config.Database = os.Getenv("DATABASE_URL")
	if config.Database == "" {
		config.Database = "host=localhost port=5432 user=postgres password=postgres dbname=mcp_restaurant sslmode=disable"
	}

	// Server configuration
	config.Server = &ServerConfig{
		Host:              os.Getenv("HOST"),
		Port:              os.Getenv("PORT"),
		OAuthServerURL:    os.Getenv("OAUTH_SERVER_URL"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		DefaultAdminEmail: os.Getenv("DEFAULT_ADMIN_EMAIL"),
		DefaultAdminName:  os.Getenv("DEFAULT_ADMIN_NAME"),
	}

	// Set defaults
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.Port == "" {
		config.Server.Port = "8080"
	}
	if config.Server.OAuthServerURL == "" {
		config.Server.OAuthServerURL = fmt.Sprintf("http://%s:%s", config.Server.Host, config.Server.Port)
	}
	if config.Server.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET environment variable is required")
	}
	if len(config.Server.JWTSecret) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 characters long")
	}
	if config.Server.DefaultAdminEmail == "" {
		config.Server.DefaultAdminEmail = "vishalkapadi17@hotmail.com"
	}
	if config.Server.DefaultAdminName == "" {
		config.Server.DefaultAdminName = "Vishal Kapadi"
	}

	// Token lifetimes
	accessTokenLife := os.Getenv("ACCESS_TOKEN_LIFETIME")
	if accessTokenLife == "" {
		config.Server.AccessTokenLife = 604800 // 7 days
	} else {
		lifetime, err := strconv.ParseInt(accessTokenLife, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid ACCESS_TOKEN_LIFETIME: %w", err)
		}
		config.Server.AccessTokenLife = lifetime
	}

	refreshTokenLife := os.Getenv("REFRESH_TOKEN_LIFETIME")
	if refreshTokenLife == "" {
		config.Server.RefreshTokenLife = 2592000 // 30 days
	} else {
		lifetime, err := strconv.ParseInt(refreshTokenLife, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid REFRESH_TOKEN_LIFETIME: %w", err)
		}
		config.Server.RefreshTokenLife = lifetime
	}

	// OAuth configuration
	oauthConfig, err := loadOAuthConfig(config.Server.OAuthServerURL)
	if err != nil {
		return nil, fmt.Errorf("OAuth configuration error: %w", err)
	}
	config.OAuth = oauthConfig

	return config, nil
}

// loadOAuthConfig loads OAuth provider configuration from environment
func loadOAuthConfig(serverURL string) (*OAuthConfig, error) {
	provider := os.Getenv("OAUTH_PROVIDER")
	if provider == "" {
		return nil, errors.New("OAUTH_PROVIDER environment variable is required")
	}

	clientID := os.Getenv("OAUTH_CLIENT_ID")
	if clientID == "" {
		return nil, errors.New("OAUTH_CLIENT_ID environment variable is required")
	}

	clientSecret := os.Getenv("OAUTH_CLIENT_SECRET")
	if clientSecret == "" {
		return nil, errors.New("OAUTH_CLIENT_SECRET environment variable is required")
	}

	authURL := os.Getenv("OAUTH_AUTH_URL")
	if authURL == "" {
		return nil, errors.New("OAUTH_AUTH_URL environment variable is required")
	}

	tokenURL := os.Getenv("OAUTH_TOKEN_URL")
	if tokenURL == "" {
		return nil, errors.New("OAUTH_TOKEN_URL environment variable is required")
	}

	userInfoURL := os.Getenv("OAUTH_USERINFO_URL")
	if userInfoURL == "" {
		return nil, errors.New("OAUTH_USERINFO_URL environment variable is required")
	}

	scopesStr := os.Getenv("OAUTH_SCOPES")
	if scopesStr == "" {
		scopesStr = "openid,profile,email"
	}
	scopes := strings.Split(scopesStr, ",")
	for i := range scopes {
		scopes[i] = strings.TrimSpace(scopes[i])
	}

	redirectURI := serverURL + "/oauth/callback"

	return &OAuthConfig{
		Provider:     provider,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthURL:      authURL,
		TokenURL:     tokenURL,
		UserInfoURL:  userInfoURL,
		Scopes:       scopes,
		RedirectURI:  redirectURI,
	}, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database == "" {
		return errors.New("database configuration is required")
	}
	if c.Server.JWTSecret == "" {
		return errors.New("JWT_SECRET is required")
	}
	if c.OAuth == nil {
		return errors.New("OAuth configuration is required")
	}
	if c.OAuth.ClientID == "" || c.OAuth.ClientSecret == "" {
		return errors.New("OAuth client credentials are required")
	}
	return nil
}
