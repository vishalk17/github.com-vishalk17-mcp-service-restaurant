# MCP Service with OAuth 2.0

Complete OAuth 2.0 server implementation for MCP (Model Context Protocol) compatible with ChatGPT Desktop and Claude Desktop.

## 🚀 Features

- ✅ **OAuth 2.0 Authorization Code Flow**
- ✅ **Dynamic Client Registration (DCR)** - RFC 7591 compliant
- ✅ **Multi-Provider Support** - Google, Microsoft, AWS Cognito
- ✅ **JWT Access & Refresh Tokens**
- ✅ **Token Introspection & Revocation**
- ✅ **Email Whitelist** - Only registered users can access
- ✅ **Default Admin User** - Pre-seeded admin account
- ✅ **Well-known Discovery Endpoints**
- ✅ **CORS Support** - Works with web clients
- ✅ **OpenID Connect Compatible**
- ✅ **ChatGPT & Claude Desktop Compatible**

## 📋 Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- OAuth credentials from Google/Microsoft/Cognito

## 🛠️ Quick Start

### 1. Setup Database

```bash
# Create database
createdb mcp_restaurant

# Run schema (automatically done on first start)
psql mcp_restaurant < database/schema.sql
```

### 2. Configure Environment

Create `.env` file:

```bash
# Database
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/mcp_restaurant

# OAuth Server
OAUTH_SERVER_URL=https://api-vishalk17.kavish.world
JWT_SECRET=your-super-secret-jwt-key-at-least-32-characters-long

# Token Lifetimes
ACCESS_TOKEN_LIFETIME=604800    # 7 days
REFRESH_TOKEN_LIFETIME=2592000  # 30 days

# Default Admin (pre-seeded)
DEFAULT_ADMIN_EMAIL=vishalkapadi17@hotmail.com
DEFAULT_ADMIN_NAME=Vishal Kapadi

# Provider: google, microsoft, or cognito
OAUTH_PROVIDER=google
OAUTH_CLIENT_ID=your-client-id
OAUTH_CLIENT_SECRET=your-client-secret

# Provider URLs
OAUTH_AUTH_URL=https://accounts.google.com/o/oauth2/v2/auth
OAUTH_TOKEN_URL=https://oauth2.googleapis.com/token
OAUTH_USERINFO_URL=https://www.googleapis.com/oauth2/v2/userinfo
OAUTH_SCOPES=openid,profile,email

# Server
HOST=0.0.0.0
PORT=8080
```

### 3. Build and Run

```bash
# Build
go build -o oauth-mcp cmd/api/main.go

# Run
./oauth-mcp
```

## 🔐 OAuth Provider Setup

### Google OAuth

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create OAuth credentials
3. Add redirect URI: `https://your-domain.com/oauth/callback`
4. Use these URLs:
   - Auth: `https://accounts.google.com/o/oauth2/v2/auth`
   - Token: `https://oauth2.googleapis.com/token`
   - UserInfo: `https://www.googleapis.com/oauth2/v2/userinfo`

### Microsoft OAuth

1. Go to [Azure Portal](https://portal.azure.com/)
2. Register application
3. Add redirect URI: `https://your-domain.com/oauth/callback`
4. Use these URLs:
   - Auth: `https://login.microsoftonline.com/common/oauth2/v2.0/authorize`
   - Token: `https://login.microsoftonline.com/common/oauth2/v2.0/token`
   - UserInfo: `https://graph.microsoft.com/v1.0/me`

### AWS Cognito

1. Create Cognito User Pool
2. Create App Client
3. Add redirect URI: `https://your-domain.com/oauth/callback`
4. Use these URLs:
   - Auth: `https://your-domain.auth.region.amazoncognito.com/oauth2/authorize`
   - Token: `https://your-domain.auth.region.amazoncognito.com/oauth2/token`
   - UserInfo: `https://your-domain.auth.region.amazoncognito.com/oauth2/userInfo`

## 📡 API Endpoints

### OAuth Endpoints

- `GET /oauth/authorize` - Authorization endpoint
- `GET /oauth/callback` - Provider callback
- `POST /oauth/token` - Token endpoint
- `POST /oauth/register` - Dynamic Client Registration
- `GET /oauth/userinfo` - User information
- `POST /oauth/introspect` - Token introspection
- `POST /oauth/revoke` - Token revocation

### Well-known Endpoints

- `GET /.well-known/oauth-authorization-server` - OAuth metadata
- `GET /.well-known/openid-configuration` - OpenID configuration
- `GET /.well-known/jwks.json` - JSON Web Key Set

### Management

- `GET /health` - Health check

## 🎯 Using with ChatGPT / Claude Desktop

### 1. Register OAuth Client

```bash
curl -X POST http://localhost:8080/oauth/register \
  -H "Content-Type: application/json" \
  -d '{
    "client_name": "ChatGPT Desktop",
    "redirect_uris": ["https://chatgpt.com/aip/c/o/redirect"]
  }'
```

Save the `client_id` and `client_secret` from the response.

### 2. Configure in ChatGPT/Claude

Add to your MCP settings:

```json
{
  "mcpServers": {
    "restaurant-service": {
      "url": "https://api-vishalk17.kavish.world",
      "oauth": {
        "authorization_endpoint": "https://api-vishalk17.kavish.world/oauth/authorize",
        "token_endpoint": "https://api-vishalk17.kavish.world/oauth/token",
        "client_id": "mcp-xxxx-xxxx",
        "client_secret": "xxxxxx"
      }
    }
  }
}
```

## 👥 User Management

### Default Admin

The system automatically creates an admin user:
- Email: `vishalkapadi17@hotmail.com`
- Status: Active
- Role: Admin

This user can login immediately after OAuth setup.

### Adding New Users

Users must be pre-registered in the database:

```sql
-- Add a new user
INSERT INTO user_profiles (user_id, email, name, status, role) 
VALUES (
    gen_random_uuid(), 
    'newuser@example.com',
    'New User',
    'active',
    'user'
);
```

### Email Whitelist

Only users with emails in the `user_profiles` table can login via OAuth. This provides security by preventing unauthorized access.

## 🔒 Security Features

- **Email Whitelist** - Only pre-registered users can access
- **JWT Tokens** - Cryptographically signed tokens
- **Token Expiration** - Configurable token lifetimes
- **Token Revocation** - Ability to revoke tokens
- **CORS Protection** - Configurable origin restrictions
- **State Parameter** - CSRF protection for OAuth flow
- **HTTPS Required** - For production use

## 🧪 Testing

### Test OAuth Flow

```bash
# 1. Health check
curl http://localhost:8080/health

# 2. Get OAuth metadata
curl http://localhost:8080/.well-known/oauth-authorization-server

# 3. Register client
curl -X POST http://localhost:8080/oauth/register \
  -H "Content-Type: application/json" \
  -d '{"client_name":"Test Client","redirect_uris":["http://localhost:3000/callback"]}'

# 4. Start OAuth flow
# Open in browser:
# http://localhost:8080/oauth/authorize?client_id=YOUR_CLIENT_ID&redirect_uri=http://localhost:3000/callback&response_type=code
```

## 📦 Project Structure

```
mcp-service/
├── cmd/
│   └── api/
│       └── main.go              # Main entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── models/
│   │   └── oauth.go             # Data models
│   ├── database/
│   │   └── db.go                # Database connection
│   ├── oauth/
│   │   ├── server.go            # OAuth server
│   │   ├── provider.go          # Generic OAuth provider
│   │   ├── token_manager.go     # JWT token management
│   │   ├── client_registry.go   # Dynamic Client Registration
│   │   ├── storage.go           # Database operations
│   │   └── middleware.go        # Auth middleware
│   └── middleware/
│       └── cors.go              # CORS middleware
├── database/
│   └── schema.sql               # Database schema
├── .env.example                 # Example environment variables
└── README.md                    # This file
```

## 🐛 Troubleshooting

### "User not authorized" error

The user's email must exist in the `user_profiles` table. Add them:

```sql
INSERT INTO user_profiles (user_id, email, name, status, role) 
VALUES (gen_random_uuid(), 'user@example.com', 'User Name', 'active', 'user');
```

### "Invalid client_id" error

Register your OAuth client first using `/oauth/register`.

### "Invalid redirect_uri" error

Ensure the redirect URI exactly matches what you registered.

### Database connection errors

Check your `DATABASE_URL` is correct and PostgreSQL is running.

## 📝 License

MIT License

## 🤝 Contributing

Contributions welcome! Please open an issue or PR.

## 📧 Support

For issues or questions, contact: vishalkapadi17@hotmail.com
