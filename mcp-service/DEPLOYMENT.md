# Kubernetes Deployment - v0.8

## Deployment Summary

✅ **Successfully deployed to Kubernetes**

### Image Details
- **Registry**: GitHub Container Registry (ghcr.io)
- **Image**: `ghcr.io/vishalk17/mcp-service-restaurant:v0.8`
- **Digest**: `sha256:b120164cae3c38ca2e205f367dd71eb9925fe51de41561e450605e5e3bc97f0e`
- **Public**: Yes (no credentials needed to pull)

### What's Deployed

1. **Database (PostgreSQL)**
   - StatefulSet with persistent storage
   - Service: `postgres-service:5432`
   - Secret: `postgres-secret` (credentials)

2. **Application (HTTP API)**
   - Deployment: `mcp-service-restaurant`
   - Image: v0.8
   - Service: `mcp-service-restaurant:80` (ClusterIP)
   - Health checks: `/health` endpoint
   - Ready and running ✅

### Kubernetes Resources

```bash
# Check deployment
kubectl get deployment mcp-service-restaurant

# Check pods
kubectl get pods -l app=mcp-service-restaurant

# Check logs
kubectl logs -l app=mcp-service-restaurant

# Check service
kubectl get svc mcp-service-restaurant

# Check database
kubectl get statefulset postgres
kubectl get svc postgres-service
```

### Build & Push Commands

```bash
# Build image
docker build -t ghcr.io/vishalk17/mcp-service-restaurant:v0.8 .

# Push to registry
docker push ghcr.io/vishalk17/mcp-service-restaurant:v0.8

# Apply to Kubernetes
kubectl apply -f k8s/database.yaml
kubectl apply -f k8s/deployment.yaml
```

### Deployment Status

```
NAME                     READY   UP-TO-DATE   AVAILABLE   IMAGES
mcp-service-restaurant   1/1     1            1           ghcr.io/vishalk17/mcp-service-restaurant:v0.8
```

### Service Endpoints

- **Internal**: `http://mcp-service-restaurant.default.svc.cluster.local`
- **Port**: 80 (maps to container port 8080)
- **Health Check**: `http://mcp-service-restaurant/health`

### API Endpoints (HTTP)

All endpoints are prefixed with `/api`:

- `GET /api/restaurants` - List all restaurants
- `GET /api/restaurants/{id}` - Get restaurant by ID
- `GET /api/restaurants/{id}/menu` - Get restaurant menu
- `POST /api/restaurants` - Create restaurant
- `GET /api/orders` - List all orders
- `GET /api/orders/{id}` - Get order by ID
- `POST /api/orders` - Create order
- `GET /health` - Health check

### Environment Variables

The application uses:
- `PORT=8080` - HTTP server port
- `DATABASE_URL` - PostgreSQL connection (from secret)

### Rolling Updates

To update to a new version:

```bash
# Build new version
docker build -t ghcr.io/vishalk17/mcp-service-restaurant:v0.9 .
docker push ghcr.io/vishalk17/mcp-service-restaurant:v0.9

# Update deployment.yaml image tag
# Then apply
kubectl apply -f k8s/deployment.yaml

# Kubernetes will automatically perform rolling update
```

### Notes

- This deploys the **HTTP API** (`cmd/api/main.go`), not the MCP server
- For MCP server usage with desktop apps, use the `mcp-server` binary locally
- The HTTP API is suitable for web access, Kubernetes ingress, etc.
- Database credentials are stored in `postgres-secret`
- Image is public and doesn't require authentication to pull
