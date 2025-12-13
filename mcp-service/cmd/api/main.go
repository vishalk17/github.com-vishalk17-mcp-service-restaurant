package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/vishalk17/mcp-service-restaurant/internal/config"
	"github.com/vishalk17/mcp-service-restaurant/internal/database"
	"github.com/vishalk17/mcp-service-restaurant/internal/middleware"
	"github.com/vishalk17/mcp-service-restaurant/internal/oauth"
)

func main() {
	log.Println("🚀 Starting MCP Service with OAuth...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal("Invalid configuration:", err)
	}

	log.Printf("✅ Configuration loaded successfully")
	log.Printf("   Provider: %s", cfg.OAuth.Provider)
	log.Printf("   OAuth Server: %s", cfg.Server.OAuthServerURL)
	log.Printf("   Default Admin: %s", cfg.Server.DefaultAdminEmail)

	// Connect to database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize OAuth components
	oauthStorage := oauth.NewStorage(db.DB)
	oauthServer := oauth.NewServer(cfg, oauthStorage)
	authMiddleware := oauth.NewAuthMiddleware(
		oauthServer.GetTokenManager(),
		nil, // Use default public paths
	)

	log.Println("✅ OAuth server initialized")

	// Create main router
	mux := http.NewServeMux()

	// OAuth endpoints (public)
	mux.HandleFunc("/oauth/authorize", oauthServer.HandleAuthorize)
	mux.HandleFunc("/oauth/callback", oauthServer.HandleCallback)
	mux.HandleFunc("/oauth/token", oauthServer.HandleToken)
	mux.HandleFunc("/oauth/register", oauthServer.HandleRegister)
	mux.HandleFunc("/oauth/userinfo", oauthServer.HandleUserInfo)
	mux.HandleFunc("/oauth/introspect", oauthServer.HandleIntrospect)
	mux.HandleFunc("/oauth/revoke", oauthServer.HandleRevoke)

	// Well-known endpoints (public)
	mux.HandleFunc("/.well-known/oauth-authorization-server", oauthServer.HandleOAuthMetadata)
	mux.HandleFunc("/.well-known/openid-configuration", oauthServer.HandleOAuthMetadata)
	mux.HandleFunc("/.well-known/jwks.json", oauthServer.HandleJWKS)

	// Health check (public)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	log.Println("✅ OAuth routes registered")
	log.Println("")
	log.Println("📍 OAuth Endpoints:")
	log.Printf("   Authorization: %s/oauth/authorize", cfg.Server.OAuthServerURL)
	log.Printf("   Token: %s/oauth/token", cfg.Server.OAuthServerURL)
	log.Printf("   Register: %s/oauth/register", cfg.Server.OAuthServerURL)
	log.Printf("   Metadata: %s/.well-known/oauth-authorization-server", cfg.Server.OAuthServerURL)
	log.Println("")

	// Apply middleware (CORS, then Auth)
	handler := middleware.CORSMiddleware(authMiddleware.Middleware(mux))

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🌐 Server listening on %s", addr)
	log.Printf("")
	log.Printf("✨ MCP Service with OAuth is ready!")
	log.Printf("   Compatible with ChatGPT Desktop and Claude Desktop")
	log.Printf("")

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal("Server failed:", err)
	}
}
