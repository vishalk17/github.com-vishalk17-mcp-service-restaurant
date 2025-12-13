# OAuth 2.0 Implementation Guide for MCP Servers

**Transform your non-authenticated MCP server into a production-ready OAuth 2.0 secured service in 60 minutes.**

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [What You'll Get](#what-youll-get)
- [Prerequisites](#prerequisites)
- [Migration Path: Non-Auth → OAuth](#migration-path-non-auth--oauth)
- [Step-by-Step Implementation](#step-by-step-implementation)
- [Compliance & Standards](#compliance--standards)
- [Security Checklist](#security-checklist)
- [Configuration Reference](#configuration-reference)
- [Deployment](#deployment)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

**Already have an MCP server without authentication? Add OAuth in these steps:**

1. ✅ Add database tables (5 min)
2. ✅ Configure OAuth provider (10 min)
3. ✅ Add OAuth server code (20 min)
4. ✅ Wrap MCP endpoint with auth middleware (10 min)
5. ✅ Deploy and test (15 min)

**What changes**: Your MCP server gains enterprise authentication while keeping all existing tools intact.

---

## What You'll Get

### Before (Non-Auth MCP Server)
```
GET /mcp → Anyone can access your tools ❌
```

### After (OAuth-Secured MCP Server)
```
GET /mcp + Bearer Token → Only authorized users ✅
  ↳ Email whitelist enforcement
  ↳ Token expiration (7 days)
  ↳ Audit trail (who accessed what)
  ↳ Multi-provider support
  ↳ ChatGPT/Claude Desktop compatible
```

### Why OAuth for MCP?

- 🔒 **Access Control**: Only whitelisted users can access your tools
- 🌐 **Enterprise SSO**: Use your existing Google/Microsoft/AWS accounts
- 🎟️ **Token-Based**: JWT tokens with automatic expiration
- 📊 **Audit Trail**: Track who uses which tools and when
- 🏢 **Production Ready**: Follows OAuth 2.0 & OpenID Connect RFCs
- 🤖 **AI Compatible**: Works seamlessly with ChatGPT & Claude Desktop

---

## Prerequisites

### What You Need

**Already Running**:
- ✅ MCP server with tools (e.g., restaurant CRUD, file operations, etc.)
- ✅ HTTP server (Go, Python, Node.js, any language)
- ✅ Database (PostgreSQL/MySQL recommended)

**Need to Add**:
- 🔧 OAuth provider account (choose one):
  - Google Cloud (free tier available)
  - Microsoft Azure AD (free tier available)
  - AWS Cognito (free tier: 50K MAU)
- 🔧 HTTPS domain (required for production)
- 🔧 JWT library for your language

**Time Required**: ~60 minutes total

---

## Migration Path: Non-Auth → OAuth

### Your Current Architecture

```
┌─────────────┐
│   ChatGPT   │
└──────┬──────┘
       │ Direct access (no auth)
       ↓
┌─────────────────────┐
│   MCP Server        │
│                     │
│  POST /mcp          │ ← Your existing endpoint
│  • tools/list       │
│  • tools/call       │
└─────────────────────┘
```

### Target Architecture (After OAuth)

```
┌─────────────┐
│   ChatGPT   │
└──────┬──────┘
       │ 1. Register client (auto)
       ↓
┌─────────────────────────────────────┐
│   OAuth Layer (NEW)                 │ ← ADD THIS
│  ┌──────────────────────────────┐  │
│  │ /oauth/register (DCR)        │  │
│  │ /oauth/authorize             │  │
│  │ /oauth/callback              │  │
│  │ /oauth/token                 │  │
│  │ /.well-known/*               │  │
│  └──────────────────────────────┘  │
│                                     │
│  ┌──────────────────────────────┐  │
│  │ Auth Middleware (NEW)        │  │ ← ADD THIS
│  │ • Validate Bearer token      │  │
│  │ • Check email whitelist      │  │
│  └──────────┬───────────────────┘  │
│             │                       │
│  ┌──────────▼───────────────────┐  │
│  │ POST /mcp (EXISTING)         │  │ ← KEEP AS IS
│  │ • tools/list                 │  │
│  │ • tools/call                 │  │
│  └──────────────────────────────┘  │
└─────────────────────────────────────┘
       │
       ↓
┌─────────────────┐      ┌─────────────┐
│  Google/MS/AWS  │      │  Database   │ ← ADD TABLES
│  OAuth Provider │      │  (3 tables) │
└─────────────────┘      └─────────────┘
```

**Key Point**: Your existing MCP tools **don't change**. You only add an OAuth wrapper around them.

---

## Step-by-Step Implementation

### Phase 1: Database Setup (5 minutes)

**What to add**: 3 OAuth tables to your existing database

**Why**: Store authorized users, OAuth clients, and token metadata

**Run this SQL**:

```sql
-- Table 1: User whitelist (email-based authorization)
CREATE TABLE user_profiles (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,        -- THIS IS YOUR WHITELIST
    name VARCHAR(255),
    picture TEXT,
    provider VARCHAR(50),                       -- google/microsoft/cognito
    provider_user_id VARCHAR(255),
    status VARCHAR(20) DEFAULT 'active',        -- active/suspended
    role VARCHAR(20) DEFAULT 'user',            -- user/admin
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_login_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Table 2: OAuth clients (ChatGPT/Claude auto-register)
CREATE TABLE oauth_clients (
    id SERIAL PRIMARY KEY,
    client_id VARCHAR(255) UNIQUE NOT NULL,
    client_secret VARCHAR(255) NOT NULL,
    client_name VARCHAR(255) NOT NULL,          -- e.g., "ChatGPT Desktop"
    redirect_uris JSONB NOT NULL,               -- OAuth redirect URLs
    grant_types JSONB DEFAULT '["authorization_code"]'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    active BOOLEAN DEFAULT true
);

-- Table 3: Token tracking (revocation & audit)
CREATE TABLE oauth_tokens (
    id SERIAL PRIMARY KEY,
    token_id VARCHAR(255) UNIQUE NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    token_type VARCHAR(50) NOT NULL,            -- access_token/refresh_token
    expires_at TIMESTAMPTZ NOT NULL,
    active BOOLEAN DEFAULT true,                -- for revocation
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (client_id) REFERENCES oauth_clients(client_id)
);

-- Indexes (IMPORTANT for performance)
CREATE INDEX idx_user_email ON user_profiles(email);
CREATE INDEX idx_oauth_tokens_active ON oauth_tokens(active, expires_at);
CREATE INDEX idx_oauth_clients_active ON oauth_clients(client_id, active);

-- Seed first authorized user (REPLACE WITH YOUR EMAIL)
INSERT INTO user_profiles (user_id, email, name, status, role)
VALUES (
    gen_random_uuid()::text,
    'your-email@example.com',  -- ← CHANGE THIS
    'Your Name',
    'active',
    'admin'
);
```

**Validate**:
```sql
SELECT * FROM user_profiles;  -- Should show 1 user
```

---

### Phase 2: Configure OAuth Provider (10 minutes)

**Choose one provider and complete its setup**:

#### Option A: AWS Cognito (Recommended)

**Why**: Easiest to configure, generous free tier (50K users/month)

**Steps**:
1. Go to AWS Cognito Console
2. Create User Pool:
   - Name: `mcp-users`
   - Sign-in: Email only
   - MFA: Optional (recommend disabled for testing)
3. Create App Client:
   - Name: `mcp-oauth-server`
   - Client secret: Generate
   - Auth flows: Check "Authorization code grant"
   - Scopes: Check `openid`, `email`, `profile`
   - Callback URL: `https://your-domain.com/oauth/callback`
4. Configure Domain:
   - Choose: Cognito domain
   - Prefix: `your-app-name` (e.g., `mcp-restaurant-app`)
5. Create first user:
   - Email: Same as your whitelist email
   - Temporary password: Will be sent via email

**Save these values**:
```bash
OAUTH_PROVIDER=cognito
OAUTH_CLIENT_ID=<from App Client>
OAUTH_CLIENT_SECRET=<from App Client>
OAUTH_AUTH_URL=https://<your-domain>.auth.<region>.amazoncognito.com/oauth2/authorize
OAUTH_TOKEN_URL=https://<your-domain>.auth.<region>.amazoncognito.com/oauth2/token
OAUTH_USERINFO_URL=https://<your-domain>.auth.<region>.amazoncognito.com/oauth2/userInfo
OAUTH_SCOPES=openid,email,profile
```

#### Option B: Google OAuth

**Steps**:
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create Project → Enable "Google+ API"
3. Credentials → Create OAuth 2.0 Client ID
4. Application type: Web application
5. Authorized redirect URIs: `https://your-domain.com/oauth/callback`

**Save these values**:
```bash
OAUTH_PROVIDER=google
OAUTH_CLIENT_ID=<your-client-id>.apps.googleusercontent.com
OAUTH_CLIENT_SECRET=<your-client-secret>
OAUTH_AUTH_URL=https://accounts.google.com/o/oauth2/v2/auth
OAUTH_TOKEN_URL=https://oauth2.googleapis.com/token
OAUTH_USERINFO_URL=https://www.googleapis.com/oauth2/v2/userinfo
OAUTH_SCOPES=openid,profile,email
```

#### Option C: Microsoft Azure AD

**Steps**:
1. Go to [Azure Portal](https://portal.azure.com/)
2. Azure Active Directory → App registrations → New
3. Redirect URI: `https://your-domain.com/oauth/callback`
4. Certificates & secrets → New client secret
5. API permissions → Add `User.Read`

**Save these values**:
```bash
OAUTH_PROVIDER=microsoft
OAUTH_CLIENT_ID=<application-id>
OAUTH_CLIENT_SECRET=<client-secret>
OAUTH_AUTH_URL=https://login.microsoftonline.com/common/oauth2/v2.0/authorize
OAUTH_TOKEN_URL=https://login.microsoftonline.com/common/oauth2/v2.0/token
OAUTH_USERINFO_URL=https://graph.microsoft.com/v1.0/me
OAUTH_SCOPES=openid,profile,email
```

---

### Phase 3: Add OAuth Server (20 minutes)

**What to add**: OAuth authorization endpoints to your server

**Files to create** (adapt to your language):

#### 3.1: Environment Configuration

**Create `.env` file**:
```bash
# Server
OAUTH_SERVER_URL=https://your-domain.com
PORT=8080

# Database (your existing connection)
DATABASE_URL=postgresql://user:pass@host:5432/dbname

# JWT Tokens (GENERATE SECURE SECRET)
JWT_SECRET=<run: openssl rand -base64 32>
ACCESS_TOKEN_LIFETIME=604800     # 7 days in seconds
REFRESH_TOKEN_LIFETIME=2592000   # 30 days in seconds

# OAuth Provider (from Phase 2)
OAUTH_PROVIDER=cognito
OAUTH_CLIENT_ID=...
OAUTH_CLIENT_SECRET=...
OAUTH_AUTH_URL=...
OAUTH_TOKEN_URL=...
OAUTH_USERINFO_URL=...
OAUTH_SCOPES=openid,email,profile

# Logging
LOG_LEVEL=info  # debug for troubleshooting
```

**🔒 SECURITY REQUIREMENT**: JWT_SECRET MUST be:
- At least 32 characters
- Cryptographically random
- Different for each environment (dev/staging/prod)
- Never committed to git

**Generate secure secret**:
```bash
# Linux/Mac
openssl rand -base64 32

# Or use this Go code
package main
import ("crypto/rand"; "encoding/base64"; "fmt")
func main() {
    b := make([]byte, 32)
    rand.Read(b)
    fmt.Println(base64.StdEncoding.EncodeToString(b))
}
```

#### 3.2: Add Required Dependencies

**Go**:
```bash
go get github.com/golang-jwt/jwt/v5
go get github.com/google/uuid
go get golang.org/x/oauth2
go get github.com/joho/godotenv
```

**Python**:
```bash
pip install PyJWT cryptography authlib httpx
```

**Node.js**:
```bash
npm install jsonwebtoken uuid axios dotenv
```

#### 3.3: Core OAuth Components

**What to implement**:

1. **Token Manager** (JWT creation & validation)
   - Create access tokens (7 day TTL)
   - Create refresh tokens (30 day TTL)
   - Validate token signature
   - Check expiration
   - Extract user claims

2. **OAuth Server** (Authorization flow)
   - `POST /oauth/register` - Dynamic Client Registration
   - `GET /oauth/authorize` - Start OAuth flow
   - `GET /oauth/callback` - Handle provider callback
   - `POST /oauth/token` - Exchange code for tokens
   - `POST /oauth/revoke` - Revoke tokens
   - `GET /.well-known/oauth-authorization-server` - Metadata

3. **Auth Middleware** (Request authentication)
   - Extract Bearer token from header
   - Validate JWT signature
   - Check expiration
   - Verify user in whitelist
   - Inject user into request context

**See complete code examples in the [Reference Implementation](#reference-implementation) section below.**

---

### Phase 4: Secure Your MCP Endpoint (10 minutes)

**What to change**: Add auth middleware to your existing MCP endpoint

**Before** (no auth):
```go
router.POST("/mcp", handleMCP)  // Anyone can access
```

**After** (with auth):
```go
// Public endpoints (no auth required)
router.POST("/oauth/register", oauthServer.HandleRegister)
router.GET("/oauth/authorize", oauthServer.HandleAuthorize)
router.GET("/oauth/callback", oauthServer.HandleCallback)
router.POST("/oauth/token", oauthServer.HandleToken)
router.GET("/.well-known/oauth-authorization-server", oauthServer.HandleMetadata)

// Protected endpoints (auth required)
protected := router.Group("/")
protected.Use(AuthMiddleware(tokenManager))  // ← ADD THIS LINE
{
    protected.POST("/mcp", handleMCP)  // Now requires Bearer token
}
```

**Your MCP handler stays the same**, just add user context:

```go
func handleMCP(w http.ResponseWriter, r *http.Request) {
    // Get authenticated user from context
    user := r.Context().Value("user").(jwt.MapClaims)
    userEmail := user["email"].(string)
    
    // Log access for audit trail
    log.Printf("MCP request from user: %s", userEmail)
    
    // Your existing MCP logic here (unchanged)
    var req MCPRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    switch req.Method {
    case "tools/list":
        // Your existing code
    case "tools/call":
        // Your existing code
    }
}
```

**That's it!** Your MCP tools now require authentication.

---

### Phase 5: Deploy & Test (15 minutes)

#### 5.1: Deployment

**Update your Kubernetes deployment** (or Docker Compose):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mcp-oauth-service
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: mcp-service
        image: your-registry/mcp-service:oauth-v1  # ← New image with OAuth
        env:
        - name: OAUTH_SERVER_URL
          value: "https://api.yourdomain.com"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: oauth-secrets
              key: jwt-secret
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: connection-string
        # ... rest of env vars from .env file
```

**Deploy**:
```bash
kubectl apply -f k8s/deployment.yaml
kubectl rollout status deployment/mcp-oauth-service
```

#### 5.2: Manual Testing

**Test 1: Metadata endpoint (should work without auth)**
```bash
curl https://api.yourdomain.com/.well-known/oauth-authorization-server | jq

# Should return:
{
  "issuer": "https://api.yourdomain.com",
  "authorization_endpoint": "https://api.yourdomain.com/oauth/authorize",
  "token_endpoint": "https://api.yourdomain.com/oauth/token",
  "registration_endpoint": "https://api.yourdomain.com/oauth/register",
  ...
}
```

**Test 2: MCP endpoint without auth (should fail)**
```bash
curl -X POST https://api.yourdomain.com/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'

# Should return: 401 Unauthorized
```

**Test 3: Register OAuth client**
```bash
curl -X POST https://api.yourdomain.com/oauth/register \
  -H "Content-Type: application/json" \
  -d '{
    "client_name": "Test Client",
    "redirect_uris": ["http://localhost:3000/callback"]
  }' | jq

# Save the client_id and client_secret from response
```

**Test 4: Get authorization URL**
```bash
CLIENT_ID=<from-test-3>
REDIRECT_URI=http://localhost:3000/callback

echo "https://api.yourdomain.com/oauth/authorize?client_id=$CLIENT_ID&redirect_uri=$REDIRECT_URI&response_type=code&state=test123&scope=openid%20email%20profile"

# Open this URL in browser → Should redirect to Google/Cognito login
```

**Test 5: Complete OAuth flow**
- Login with your whitelisted email
- Get redirected with `?code=xxx&state=test123`
- Exchange code for tokens:

```bash
AUTH_CODE=<from-redirect>

curl -X POST https://api.yourdomain.com/oauth/token \
  -d "grant_type=authorization_code" \
  -d "code=$AUTH_CODE" \
  -d "client_id=$CLIENT_ID" \
  -d "redirect_uri=$REDIRECT_URI" | jq

# Save the access_token from response
```

**Test 6: MCP endpoint with auth (should work)**
```bash
ACCESS_TOKEN=<from-test-5>

curl -X POST https://api.yourdomain.com/mcp \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}' | jq

# Should return your MCP tools
```

#### 5.3: ChatGPT Desktop Integration

**Configure ChatGPT** to use your OAuth server:

1. Open ChatGPT Settings → Custom GPTs/Connectors
2. Add new connector:
   - Name: Your Service Name
   - Base URL: `https://api.yourdomain.com`
   - Auth Type: OAuth 2.0
   - Authorization URL: `https://api.yourdomain.com/oauth/authorize`
   - Token URL: `https://api.yourdomain.com/oauth/token`
   - Scope: `openid email profile`
3. Click "Connect"
4. Login with your whitelisted email
5. Verify connection ✅

**Test in ChatGPT**:
```
You: "List all restaurants"
ChatGPT: [calls your MCP server with OAuth token]
```

---

## Compliance & Standards

### OAuth 2.0 Compliance

This implementation is **fully compliant** with industry standards:

| RFC/Standard | Feature | Status | Why It Matters |
|--------------|---------|--------|----------------|
| **RFC 6749** | OAuth 2.0 Authorization Framework | ✅ Complete | Core OAuth protocol |
| **RFC 6750** | Bearer Token Usage | ✅ Complete | Secure token transmission |
| **RFC 7519** | JSON Web Token (JWT) | ✅ Complete | Stateless token validation |
| **RFC 7591** | Dynamic Client Registration (DCR) | ✅ Complete | Auto-registration for ChatGPT/Claude |
| **RFC 7009** | Token Revocation | ✅ Complete | Logout & security |
| **RFC 7662** | Token Introspection | ✅ Complete | Token validation API |
| **RFC 8414** | Authorization Server Metadata | ✅ Complete | Auto-discovery |
| **OpenID Connect 1.0** | Identity layer on OAuth | ✅ Complete | User info & authentication |

### OAuth Flows Supported

✅ **Authorization Code Flow** (recommended)
- Used by ChatGPT Desktop & Claude Desktop
- Most secure flow for web applications
- State parameter for CSRF protection

✅ **Refresh Token Flow**
- Long-lived sessions (30 days)
- Automatic token renewal
- Reduces re-authentication friction

❌ **Implicit Flow** (deprecated)
- Not implemented (security concerns)
- Use authorization code flow instead

❌ **Password Flow** (not recommended)
- Not implemented (less secure)
- OAuth providers handle passwords

### Token Types

**Access Token** (JWT):
- Lifetime: 7 days (configurable)
- Used for: API authentication
- Contains: user ID, email, name, scopes
- Algorithm: HS256 (HMAC with SHA-256)

**Refresh Token** (JWT):
- Lifetime: 30 days (configurable)
- Used for: Getting new access tokens
- Contains: user ID, token ID
- Can be revoked individually

### Endpoints Required

**Public Endpoints** (no auth):
```
POST /oauth/register              → Dynamic Client Registration (DCR)
GET  /oauth/authorize             → Start OAuth flow
GET  /oauth/callback              → Provider callback
POST /oauth/token                 → Token exchange
GET  /.well-known/oauth-authorization-server  → Metadata
```

**Protected Endpoints** (requires Bearer token):
```
POST /oauth/revoke                → Revoke tokens
POST /oauth/introspect            → Check token validity
GET  /oauth/userinfo              → Get user info
POST /mcp                         → Your MCP endpoint
```

**Optional Endpoints**:
```
GET  /.well-known/jwks.json       → Public keys (for RS256)
GET  /.well-known/openid-configuration  → OpenID metadata
```

---

## Architecture

```
┌─────────────┐
│   ChatGPT   │
│   Desktop   │
└──────┬──────┘
       │ 1. Request tools
       ↓
┌─────────────────────────────────────┐
│     MCP Server with OAuth           │
│                                     │
│  ┌──────────────────────────────┐  │
│  │   OAuth Authorization        │  │
│  │   Server                     │  │
│  ├──────────────────────────────┤  │
│  │ • /oauth/authorize           │  │
│  │ • /oauth/token               │  │
│  │ • /oauth/register (DCR)      │  │
│  │ • /.well-known/*             │  │
│  └──────────┬───────────────────┘  │
│             │                       │
│  ┌──────────▼───────────────────┐  │
│  │   Auth Middleware            │  │
│  │   (JWT Validation)           │  │
│  └──────────┬───────────────────┘  │
│             │                       │
│  ┌──────────▼───────────────────┐  │
│  │   MCP JSON-RPC Endpoint      │  │
│  │   • initialize               │  │
│  │   • tools/list               │  │
│  │   • tools/call               │  │
│  └──────────────────────────────┘  │
└─────────────────────────────────────┘
       │
       ↓
┌─────────────────┐      ┌─────────────┐
│  OAuth Provider │      │  Database   │
│  (Google/MS/    │      │  (Users,    │
│   Cognito)      │      │   Tokens)   │
└─────────────────┘      └─────────────┘
```

---

## Prerequisites

### Required

1. **OAuth Provider** (choose one):
   - Google OAuth 2.0
   - Microsoft Azure AD
   - AWS Cognito
   - Any OpenID Connect provider

2. **Database**:
   - PostgreSQL 12+ (recommended)
   - MySQL 8.0+
   - Any SQL database with JSON support

3. **Dependencies**:
   - JWT library (e.g., `golang-jwt/jwt`)
   - OAuth 2.0 client library
   - HTTP router
   - Database driver

### Environment Requirements

- HTTPS endpoint (required for production)
- Domain name for OAuth callbacks
- SSL/TLS certificate

---

## Implementation Steps

### Step 1: Database Schema

Create tables for OAuth entities:

```sql
-- User profiles (email whitelist)
CREATE TABLE user_profiles (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    picture TEXT,
    provider VARCHAR(50),
    provider_user_id VARCHAR(255),
    status VARCHAR(20) DEFAULT 'active',
    role VARCHAR(20) DEFAULT 'user',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_login_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- OAuth clients (Dynamic Client Registration)
CREATE TABLE oauth_clients (
    id SERIAL PRIMARY KEY,
    client_id VARCHAR(255) UNIQUE NOT NULL,
    client_secret VARCHAR(255) NOT NULL,
    client_name VARCHAR(255) NOT NULL,
    redirect_uris JSONB NOT NULL,
    grant_types JSONB DEFAULT '["authorization_code"]'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    active BOOLEAN DEFAULT true
);

-- Token metadata (for revocation)
CREATE TABLE oauth_tokens (
    id SERIAL PRIMARY KEY,
    token_id VARCHAR(255) UNIQUE NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    token_type VARCHAR(50) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    active BOOLEAN DEFAULT true,
    FOREIGN KEY (client_id) REFERENCES oauth_clients(client_id)
);

-- Indexes
CREATE INDEX idx_user_email ON user_profiles(email);
CREATE INDEX idx_oauth_tokens_active ON oauth_tokens(active);
```

### Step 2: Configuration

Environment variables needed:

```bash
# OAuth Server
OAUTH_SERVER_URL=https://your-domain.com
JWT_SECRET=generate-a-secure-random-string-at-least-32-chars

# Token Lifetimes (in seconds)
ACCESS_TOKEN_LIFETIME=604800    # 7 days
REFRESH_TOKEN_LIFETIME=2592000  # 30 days

# OAuth Provider (choose one: google, microsoft, cognito)
OAUTH_PROVIDER=google
OAUTH_CLIENT_ID=your-provider-client-id
OAUTH_CLIENT_SECRET=your-provider-client-secret

# Provider URLs
OAUTH_AUTH_URL=https://accounts.google.com/o/oauth2/v2/auth
OAUTH_TOKEN_URL=https://oauth2.googleapis.com/token
OAUTH_USERINFO_URL=https://www.googleapis.com/oauth2/v2/userinfo
OAUTH_SCOPES=openid,profile,email

# Database
DATABASE_URL=postgresql://user:pass@localhost/dbname
```

### Step 3: Core Components

#### 3.1 OAuth Provider Integration

```go
type Provider struct {
    config *oauth2.Config
}

func NewProvider(cfg *OAuthConfig) *Provider {
    return &Provider{
        config: &oauth2.Config{
            ClientID:     cfg.ClientID,
            ClientSecret: cfg.ClientSecret,
            RedirectURL:  cfg.ServerURL + "/oauth/callback",
            Scopes:       cfg.Scopes,
            Endpoint: oauth2.Endpoint{
                AuthURL:  cfg.AuthURL,
                TokenURL: cfg.TokenURL,
            },
        },
    }
}

func (p *Provider) GetAuthorizationURL(state string) string {
    return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *Provider) GetUserInfo(ctx context.Context, token string) (*UserInfo, error) {
    // Fetch user info from provider's userinfo endpoint
}
```

#### 3.2 JWT Token Manager

```go
type TokenManager struct {
    secret     []byte
    issuer     string
    accessTTL  time.Duration
    refreshTTL time.Duration
}

func (tm *TokenManager) CreateTokens(user *User, clientID, scope string) (*TokenResponse, error) {
    // Generate access token
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub":        user.UserID,
        "email":      user.Email,
        "name":       user.Name,
        "client_id":  clientID,
        "scope":      scope,
        "token_type": "access_token",
        "exp":        time.Now().Add(tm.accessTTL).Unix(),
        "iat":        time.Now().Unix(),
        "iss":        tm.issuer,
    })
    
    accessTokenString, _ := accessToken.SignedString(tm.secret)
    
    // Generate refresh token (similar structure)
    // Save token metadata to database
    
    return &TokenResponse{
        AccessToken:  accessTokenString,
        TokenType:    "Bearer",
        ExpiresIn:    int64(tm.accessTTL.Seconds()),
        RefreshToken: refreshTokenString,
        Scope:        scope,
    }, nil
}

func (tm *TokenManager) ValidateToken(tokenString string) (jwt.MapClaims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return tm.secret, nil
    })
    
    if err != nil || !token.Valid {
        return nil, err
    }
    
    return token.Claims.(jwt.MapClaims), nil
}
```

#### 3.3 Auth Middleware

```go
func AuthMiddleware(tokenManager *TokenManager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Skip auth for public endpoints
            if isPublicPath(r.URL.Path) {
                next.ServeHTTP(w, r)
                return
            }
            
            // Extract Bearer token
            authHeader := r.Header.Get("Authorization")
            if !strings.HasPrefix(authHeader, "Bearer ") {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            
            token := strings.TrimPrefix(authHeader, "Bearer ")
            
            // Validate token
            claims, err := tokenManager.ValidateToken(token)
            if err != nil {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }
            
            // Inject user context
            ctx := context.WithValue(r.Context(), "user", claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func isPublicPath(path string) bool {
    publicPaths := []string{
        "/oauth/",
        "/.well-known/",
        "/health",
    }
    
    for _, public := range publicPaths {
        if strings.HasPrefix(path, public) {
            return true
        }
    }
    return false
}
```

### Step 4: OAuth Endpoints

#### 4.1 Authorization Endpoint

```go
func (s *OAuthServer) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
    clientID := r.URL.Query().Get("client_id")
    redirectURI := r.URL.Query().Get("redirect_uri")
    state := r.URL.Query().Get("state")
    scope := r.URL.Query().Get("scope")
    
    // Validate client and redirect URI
    if !s.validateClient(clientID, redirectURI) {
        http.Error(w, "Invalid client", http.StatusBadRequest)
        return
    }
    
    // Store request state and redirect to OAuth provider
    oauthState := s.generateState(clientID, redirectURI, state, scope)
    providerURL := s.provider.GetAuthorizationURL(oauthState)
    
    http.Redirect(w, r, providerURL, http.StatusFound)
}
```

#### 4.2 Callback Endpoint

```go
func (s *OAuthServer) HandleCallback(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    state := r.URL.Query().Get("state")
    
    // Decode state
    requestData := s.decodeState(state)
    
    // Exchange code with provider
    token, err := s.provider.ExchangeCode(r.Context(), code)
    if err != nil {
        http.Error(w, "Failed to authenticate", http.StatusUnauthorized)
        return
    }
    
    // Get user info
    userInfo, err := s.provider.GetUserInfo(r.Context(), token.AccessToken)
    if err != nil {
        http.Error(w, "Failed to get user info", http.StatusUnauthorized)
        return
    }
    
    // Check if user is authorized (email whitelist)
    user, err := s.storage.FindUserByEmail(userInfo.Email)
    if err != nil || user == nil {
        http.Error(w, "User not authorized", http.StatusForbidden)
        return
    }
    
    // Generate authorization code
    authCode := s.generateAuthCode(requestData.ClientID, user, requestData.Scope)
    
    // Redirect back to client with code
    redirectURL := fmt.Sprintf("%s?code=%s&state=%s", 
        requestData.RedirectURI, authCode, requestData.State)
    http.Redirect(w, r, redirectURL, http.StatusFound)
}
```

#### 4.3 Token Endpoint

```go
func (s *OAuthServer) HandleToken(w http.ResponseWriter, r *http.Request) {
    grantType := r.FormValue("grant_type")
    
    switch grantType {
    case "authorization_code":
        s.handleAuthorizationCode(w, r)
    case "refresh_token":
        s.handleRefreshToken(w, r)
    default:
        s.jsonError(w, "unsupported_grant_type", http.StatusBadRequest)
    }
}

func (s *OAuthServer) handleAuthorizationCode(w http.ResponseWriter, r *http.Request) {
    code := r.FormValue("code")
    clientID := r.FormValue("client_id")
    
    // Validate authorization code
    authCode := s.validateAndConsumeAuthCode(code, clientID)
    if authCode == nil {
        s.jsonError(w, "invalid_grant", http.StatusBadRequest)
        return
    }
    
    // Create tokens
    tokens, err := s.tokenManager.CreateTokens(
        authCode.User, 
        clientID, 
        authCode.Scope,
    )
    if err != nil {
        s.jsonError(w, "server_error", http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(tokens)
}
```

#### 4.4 Dynamic Client Registration (RFC 7591)

```go
func (s *OAuthServer) HandleRegister(w http.ResponseWriter, r *http.Request) {
    var req struct {
        ClientName   string   `json:"client_name"`
        RedirectURIs []string `json:"redirect_uris"`
    }
    
    json.NewDecoder(r.Body).Decode(&req)
    
    // Generate credentials
    clientID := "mcp-" + uuid.New().String()
    clientSecret := generateSecureSecret()
    
    // Default redirect URIs for ChatGPT/Claude
    if len(req.RedirectURIs) == 0 {
        req.RedirectURIs = []string{
            "https://chatgpt.com/connector_platform_oauth_redirect",
        }
    }
    
    // Save to database
    client := &OAuthClient{
        ClientID:       clientID,
        ClientSecret:   clientSecret,
        ClientName:     req.ClientName,
        RedirectURIs:   req.RedirectURIs,
        ClientIDIssuedAt: time.Now().Unix(),
    }
    
    s.storage.CreateClient(client)
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(client)
}
```

#### 4.5 Well-Known Metadata (RFC 8414)

```go
func (s *OAuthServer) HandleMetadata(w http.ResponseWriter, r *http.Request) {
    metadata := map[string]interface{}{
        "issuer":                   s.config.ServerURL,
        "authorization_endpoint":   s.config.ServerURL + "/oauth/authorize",
        "token_endpoint":           s.config.ServerURL + "/oauth/token",
        "registration_endpoint":    s.config.ServerURL + "/oauth/register",
        "userinfo_endpoint":        s.config.ServerURL + "/oauth/userinfo",
        "jwks_uri":                 s.config.ServerURL + "/.well-known/jwks.json",
        "revocation_endpoint":      s.config.ServerURL + "/oauth/revoke",
        "introspection_endpoint":   s.config.ServerURL + "/oauth/introspect",
        "response_types_supported": []string{"code"},
        "grant_types_supported":    []string{"authorization_code", "refresh_token"},
        "token_endpoint_auth_methods_supported": []string{"none", "client_secret_post"},
        "scopes_supported":         []string{"openid", "profile", "email"},
        "subject_types_supported":  []string{"public"},
    }
    
    json.NewEncoder(w).Encode(metadata)
}
```

### Step 5: MCP Integration

#### 5.1 MCP Endpoint with Auth

```go
func (s *Server) HandleMCP(w http.ResponseWriter, r *http.Request) {
    // Auth middleware has already validated token
    // User info is in context
    user := r.Context().Value("user").(jwt.MapClaims)
    
    var req MCPRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    var response MCPResponse
    
    switch req.Method {
    case "initialize":
        response = s.handleInitialize(req.ID)
    case "tools/list":
        response = s.handleToolsList(req.ID)
    case "tools/call":
        response = s.handleToolsCall(req, user)
    default:
        response = s.errorResponse(req.ID, -32601, "Method not found")
    }
    
    json.NewEncoder(w).Encode(response)
}
```

---

## Security Considerations

### Must-Have Security Features

1. **HTTPS Only**: Never use HTTP in production
2. **JWT Secret**: Use cryptographically secure random string (≥32 chars)
3. **State Parameter**: Prevent CSRF attacks in OAuth flow
4. **Email Whitelist**: Only allow pre-registered users
5. **Token Expiration**: Set reasonable TTLs (7 days access, 30 days refresh)
6. **Token Revocation**: Implement token blacklisting
7. **Rate Limiting**: Prevent brute force attacks
8. **CORS Configuration**: Restrict allowed origins

### Best Practices

```go
// Generate secure secrets
func generateSecureSecret() string {
    b := make([]byte, 32)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)
}

// Constant-time comparison for secrets
func compareSecrets(a, b string) bool {
    return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// Rate limiting middleware
func RateLimitMiddleware(next http.Handler) http.Handler {
    // Implement token bucket or sliding window
}
```

---

## Testing

### Unit Tests

```go
func TestTokenValidation(t *testing.T) {
    tm := NewTokenManager("test-secret", "test-issuer", 3600, 86400, nil)
    
    user := &User{UserID: "123", Email: "test@example.com"}
    tokens, err := tm.CreateTokens(user, "client-1", "openid email")
    
    assert.NoError(t, err)
    assert.NotEmpty(t, tokens.AccessToken)
    
    claims, err := tm.ValidateToken(tokens.AccessToken)
    assert.NoError(t, err)
    assert.Equal(t, "123", claims["sub"])
}
```

### Integration Tests

```bash
# Test OAuth flow
curl -X POST http://localhost:8080/oauth/register \
  -d '{"client_name":"Test","redirect_uris":["http://localhost:3000/callback"]}'

# Test authorization
curl "http://localhost:8080/oauth/authorize?client_id=CLIENT_ID&redirect_uri=REDIRECT_URI&response_type=code&state=test"

# Test token endpoint
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=authorization_code&code=AUTH_CODE&client_id=CLIENT_ID"

# Test MCP with token
curl http://localhost:8080/mcp \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

---

## Deployment

### Docker

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o oauth-mcp ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/oauth-mcp .
EXPOSE 8080
CMD ["./oauth-mcp"]
```

### Kubernetes

#### Step 1: Create ConfigMap (Non-Sensitive)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mcp-oauth-config
  namespace: default
data:
  # Server Configuration
  HOST: "0.0.0.0"
  PORT: "8080"
  
  # Token Lifetimes
  ACCESS_TOKEN_LIFETIME: "604800"    # 7 days
  REFRESH_TOKEN_LIFETIME: "2592000"  # 30 days
  
  # AWS Cognito OAuth Provider URLs
  OAUTH_PROVIDER: "cognito"
  OAUTH_AUTH_URL: "https://your-domain.auth.region.amazoncognito.com/oauth2/authorize"
  OAUTH_TOKEN_URL: "https://your-domain.auth.region.amazoncognito.com/oauth2/token"
  OAUTH_USERINFO_URL: "https://your-domain.auth.region.amazoncognito.com/oauth2/userInfo"
  OAUTH_SCOPES: "openid,email,profile"
  
  # Logging
  LOG_LEVEL: "info"
```

**Apply**:
```bash
kubectl apply -f mcp-oauth-config.yaml
```

#### Step 2: Create Secret (Sensitive - Encrypted)

**🔒 NEVER commit secrets to git in plain text!**

Use `kubectl create secret` to generate base64-encoded secrets:

```bash
kubectl create secret generic mcp-oauth-secrets \
  --from-literal=OAUTH_SERVER_URL='https://api.yourdomain.com' \
  --from-literal=JWT_SECRET='your-secure-jwt-secret-32-chars-minimum' \
  --from-literal=OAUTH_CLIENT_ID='your-cognito-client-id' \
  --from-literal=OAUTH_CLIENT_SECRET='your-cognito-client-secret' \
  --from-literal=DEFAULT_ADMIN_EMAIL='admin@example.com' \
  --from-literal=DEFAULT_ADMIN_NAME='Admin User' \
  --namespace=default \
  --dry-run=client -o yaml | kubectl apply -f -
```

This generates:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mcp-oauth-secrets
  namespace: default
type: Opaque
data:
  OAUTH_SERVER_URL: <base64-encoded>
  JWT_SECRET: <base64-encoded>
  OAUTH_CLIENT_ID: <base64-encoded>
  OAUTH_CLIENT_SECRET: <base64-encoded>
  DEFAULT_ADMIN_EMAIL: <base64-encoded>
  DEFAULT_ADMIN_NAME: <base64-encoded>
```

**What goes where?**

| Type | Goes In | Examples |
|------|---------|----------|
| **Sensitive** | Secret (encrypted) | JWT_SECRET, OAUTH_CLIENT_SECRET, OAUTH_SERVER_URL, DATABASE_URL |
| **Non-Sensitive** | ConfigMap (plain) | Provider URLs, timeouts, scopes, log levels |

#### Step 3: Create Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mcp-service-restaurant
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mcp-service-restaurant
  template:
    metadata:
      labels:
        app: mcp-service-restaurant
    spec:
      containers:
      - name: mcp-restaurant-api
        image: ghcr.io/yourname/mcp-service-restaurant:v0.55
        ports:
        - containerPort: 8080
        env:
        # Database (Sensitive)
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: url
        
        # Sensitive Secrets (from Secret)
        - name: OAUTH_SERVER_URL
          valueFrom:
            secretKeyRef:
              name: mcp-oauth-secrets
              key: OAUTH_SERVER_URL
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: mcp-oauth-secrets
              key: JWT_SECRET
        - name: OAUTH_CLIENT_ID
          valueFrom:
            secretKeyRef:
              name: mcp-oauth-secrets
              key: OAUTH_CLIENT_ID
        - name: OAUTH_CLIENT_SECRET
          valueFrom:
            secretKeyRef:
              name: mcp-oauth-secrets
              key: OAUTH_CLIENT_SECRET
        - name: DEFAULT_ADMIN_EMAIL
          valueFrom:
            secretKeyRef:
              name: mcp-oauth-secrets
              key: DEFAULT_ADMIN_EMAIL
        - name: DEFAULT_ADMIN_NAME
          valueFrom:
            secretKeyRef:
              name: mcp-oauth-secrets
              key: DEFAULT_ADMIN_NAME
        
        # Non-Sensitive Config (from ConfigMap)
        - name: HOST
          valueFrom:
            configMapKeyRef:
              name: mcp-oauth-config
              key: HOST
        - name: PORT
          valueFrom:
            configMapKeyRef:
              name: mcp-oauth-config
              key: PORT
        - name: ACCESS_TOKEN_LIFETIME
          valueFrom:
            configMapKeyRef:
              name: mcp-oauth-config
              key: ACCESS_TOKEN_LIFETIME
        - name: REFRESH_TOKEN_LIFETIME
          valueFrom:
            configMapKeyRef:
              name: mcp-oauth-config
              key: REFRESH_TOKEN_LIFETIME
        - name: OAUTH_PROVIDER
          valueFrom:
            configMapKeyRef:
              name: mcp-oauth-config
              key: OAUTH_PROVIDER
        - name: OAUTH_AUTH_URL
          valueFrom:
            configMapKeyRef:
              name: mcp-oauth-config
              key: OAUTH_AUTH_URL
        - name: OAUTH_TOKEN_URL
          valueFrom:
            configMapKeyRef:
              name: mcp-oauth-config
              key: OAUTH_TOKEN_URL
        - name: OAUTH_USERINFO_URL
          valueFrom:
            configMapKeyRef:
              name: mcp-oauth-config
              key: OAUTH_USERINFO_URL
        - name: OAUTH_SCOPES
          valueFrom:
            configMapKeyRef:
              name: mcp-oauth-config
              key: OAUTH_SCOPES
        - name: LOG_LEVEL
          valueFrom:
            configMapKeyRef:
              name: mcp-oauth-config
              key: LOG_LEVEL
        
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 15
          periodSeconds: 20
---
apiVersion: v1
kind: Service
metadata:
  name: mcp-service-restaurant
  namespace: default
spec:
  selector:
    app: mcp-service-restaurant
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
```

**Deploy**:
```bash
kubectl apply -f deployment.yaml
kubectl rollout status deployment/mcp-service-restaurant
```

**Verify**:
```bash
# Check pods are running
kubectl get pods -l app=mcp-service-restaurant

# Check logs
kubectl logs -l app=mcp-service-restaurant --tail=50

# Test health endpoint
kubectl port-forward svc/mcp-service-restaurant 8080:80
curl http://localhost:8080/health
```

### Monitoring

```go
// Prometheus metrics
var (
    oauthRequests = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "oauth_requests_total",
            Help: "Total OAuth requests",
        },
        []string{"endpoint", "status"},
    )
    
    tokenValidations = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name: "token_validation_duration_seconds",
            Help: "Token validation duration",
        },
    )
)
```

---

## Troubleshooting

### Common Issues

#### 1. "redirect_uri_mismatch"

**Problem**: OAuth provider rejects redirect URI

**Solution**: 
- Add `https://your-domain.com/oauth/callback` to provider's allowed URIs
- Ensure exact match (no trailing slashes, correct protocol)

#### 2. "invalid_scope"

**Problem**: Provider doesn't support requested scopes

**Solution**:
- Check provider documentation for supported scopes
- For Cognito: Enable OAuth flows and scopes in app client settings

#### 3. "User not authorized"

**Problem**: User email not in database

**Solution**:
```sql
INSERT INTO user_profiles (user_id, email, name, status, role)
VALUES (gen_random_uuid(), 'user@example.com', 'User Name', 'active', 'user');
```

#### 4. Token validation fails

**Problem**: JWT signature invalid

**Solution**:
- Verify JWT_SECRET matches between token creation and validation
- Check token hasn't expired
- Ensure token format is correct (3 parts separated by dots)

### Debug Logging

```go
// Enable debug mode
LOG_LEVEL=debug

// Logs will show:
// - All HTTP requests with headers
// - OAuth flow steps
// - Token validation details
// - Database queries
```

---

## ChatGPT/Claude Configuration

### ChatGPT Desktop

```json
{
  "mcpServers": {
    "your-service": {
      "url": "https://api.yourdomain.com",
      "oauth": {
        "authorization_endpoint": "https://api.yourdomain.com/oauth/authorize",
        "token_endpoint": "https://api.yourdomain.com/oauth/token",
        "redirect_uri": "https://chatgpt.com/connector_platform_oauth_redirect"
      }
    }
  }
}
```

### Default Redirect URIs

Always include these in your OAuth client:

```go
defaultRedirectURIs := []string{
    "https://chatgpt.com/connector_platform_oauth_redirect",
    "https://chatgpt.com/aip/c/o/redirect",
    "http://localhost:3000/callback", // For local testing
}
```

---

## Checklist

Before going to production:

- [ ] HTTPS enabled with valid SSL certificate
- [ ] JWT_SECRET is cryptographically secure (≥32 chars)
- [ ] OAuth provider configured with correct redirect URIs
- [ ] Database tables created with indexes
- [ ] Email whitelist populated with authorized users
- [ ] Token lifetimes configured appropriately
- [ ] CORS configured correctly
- [ ] Rate limiting enabled
- [ ] Logging and monitoring set up
- [ ] Backup and disaster recovery plan
- [ ] Security audit completed
- [ ] Load testing performed

---

## Reference Implementation

See this complete working implementation:
- Repository: `/home/vishalk17/mcp-service/mcp-service`
- Key files:
  - `internal/oauth/server.go` - OAuth server
  - `internal/oauth/token_manager.go` - JWT tokens
  - `internal/oauth/middleware.go` - Auth middleware
  - `internal/oauth/client_registry.go` - DCR
  - `internal/handlers/mcp.go` - MCP + OAuth integration

---

## Support & Resources

### Standards & RFCs
- [RFC 6749](https://tools.ietf.org/html/rfc6749) - OAuth 2.0
- [RFC 7519](https://tools.ietf.org/html/rfc7519) - JWT
- [RFC 7591](https://tools.ietf.org/html/rfc7591) - DCR
- [RFC 8414](https://tools.ietf.org/html/rfc8414) - Metadata
- [OpenID Connect](https://openid.net/connect/)

### Tools
- [JWT.io](https://jwt.io/) - Debug JWT tokens
- [OAuth.tools](https://oauth.tools/) - Test OAuth flows
- [Postman](https://www.postman.com/) - API testing

---

**License**: MIT

**Author**: Vishal Kapadi (vishalkapadi17@hotmail.com)

**Version**: 1.0.0 (December 2025)
