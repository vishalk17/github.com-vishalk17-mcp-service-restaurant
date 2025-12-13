package oauth

import (
	"context"
	"net/http"
	"strings"
)

// UserContextKey is the key for user context
type contextKey string

const UserContextKey = contextKey("user")

// AuthMiddleware validates Bearer tokens and injects user context
type AuthMiddleware struct {
	tokenManager *TokenManager
	publicPaths  []string
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(tokenManager *TokenManager, publicPaths []string) *AuthMiddleware {
	if publicPaths == nil {
		publicPaths = []string{
			"/health",
			"/.well-known/oauth-authorization-server",
			"/.well-known/openid-configuration",
			"/.well-known/jwks.json",
			"/oauth/authorize",
			"/oauth/callback",
			"/oauth/register",
			"/oauth/token",
		}
	}
	return &AuthMiddleware{
		tokenManager: tokenManager,
		publicPaths:  publicPaths,
	}
}

// Middleware wraps an HTTP handler with authentication
func (am *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path is public
		if am.isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			am.unauthorized(w, "Missing Authorization header")
			return
		}

		// Check Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			am.unauthorized(w, "Invalid Authorization header format")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		claims, err := am.tokenManager.ValidateToken(token)
		if err != nil {
			am.unauthorized(w, "Invalid or expired token")
			return
		}

		// Inject user context
		userCtx := map[string]interface{}{
			"sub":       claims["sub"],
			"email":     claims["email"],
			"name":      claims["name"],
			"picture":   claims["picture"],
			"client_id": claims["client_id"],
			"scope":     claims["scope"],
		}

		ctx := context.WithValue(r.Context(), UserContextKey, userCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// isPublicPath checks if the path is public
func (am *AuthMiddleware) isPublicPath(path string) bool {
	for _, publicPath := range am.publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}

// unauthorized sends an unauthorized response
func (am *AuthMiddleware) unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("WWW-Authenticate", "Bearer realm=\"MCP OAuth\"")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"unauthorized","error_description":"` + message + `"}`))
}

// GetUserFromContext retrieves user from request context
func GetUserFromContext(ctx context.Context) map[string]interface{} {
	user, ok := ctx.Value(UserContextKey).(map[string]interface{})
	if !ok {
		return nil
	}
	return user
}
