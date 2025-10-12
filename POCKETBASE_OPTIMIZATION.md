# PocketBase Performance Optimization Guide

## Overview

This document explains the performance optimization implemented for SvelteKit + PocketBase projects to ensure server-side requests use localhost communication instead of going through the public internet.

## The Problem

When deploying SvelteKit + PocketBase projects, there are two types of API requests:

### 1. Client-Side Requests (Browser)
```
User's Browser → Internet → Your VPS → Caddy → PocketBase Container
```
**Status**: Unavoidable - must go through internet

### 2. Server-Side Requests (SSR)
**Before Optimization**:
```
SvelteKit Container → Public Domain URL → Internet → DNS → Your VPS → Caddy → PocketBase Container
```
**Problem**: Server making network request to itself through the internet

**After Optimization**:
```
SvelteKit Container → Docker Network → PocketBase Container
```
**Solution**: Direct localhost communication via Docker network

## The Solution

### Environment-Aware URL Selection

The optimization uses SvelteKit's `browser` context to select different URLs:

```typescript
import { browser } from '$app/environment';

const pbUrl = browser
  ? (import.meta.env.PUBLIC_POCKETBASE_URL || 'http://localhost:8090')  // Client-side
  : (import.meta.env.POCKETBASE_URL || 'http://localhost:8090');         // Server-side
```

### Environment Variables

Two separate environment variables are configured:

| Variable | Usage | Value (Production) | Value (Development) |
|----------|-------|-------------------|-------------------|
| `PUBLIC_POCKETBASE_URL` | Client-side (browser) | `https://yourdomain.com` | `http://localhost:8090` |
| `POCKETBASE_URL` | Server-side (SSR) | `http://pocketbase:8090` | `http://localhost:8090` |

### Docker Network Communication

In production, containers communicate via Docker's internal DNS:

```yaml
# docker-compose.yml
services:
  pocketbase:
    container_name: vps-panel-yourapp-pocketbase-1
    networks:
      - app-network

  frontend:
    container_name: vps-panel-yourapp-frontend-1
    environment:
      - POCKETBASE_URL=http://pocketbase:8090  # Uses Docker service name
    networks:
      - app-network

networks:
  app-network:
    driver: bridge
```

When the frontend makes a request to `http://pocketbase:8090`:
1. Docker's internal DNS resolves `pocketbase` to the container's IP (e.g., 172.18.0.2)
2. Request goes through Docker's network bridge
3. **No external network, no DNS lookup, no public IP routing**

## Implementation

### 1. PocketBase Initialization Template

Location: `backend/templates/pocketbase.ts.template`

```typescript
import PocketBase from 'pocketbase';
import { browser } from '$app/environment';

// Environment-aware PocketBase URL selection
const pbUrl = browser
  ? (import.meta.env.PUBLIC_POCKETBASE_URL || 'http://localhost:8090')
  : (import.meta.env.POCKETBASE_URL || 'http://localhost:8090');

export const pb = new PocketBase(pbUrl);
```

### 2. Dockerfile Configuration

The Dockerfile accepts both environment variables as build args:

```dockerfile
# Build stage
ARG PUBLIC_POCKETBASE_URL
ARG POCKETBASE_URL

ENV PUBLIC_POCKETBASE_URL=$PUBLIC_POCKETBASE_URL
ENV POCKETBASE_URL=$POCKETBASE_URL

RUN npm run build
```

### 3. Docker Compose Configuration

Location: `backend/internal/services/deployment/pocketbase.go`

```yaml
frontend:
  build:
    args:
      PUBLIC_POCKETBASE_URL: https://yourdomain.com
      POCKETBASE_URL: http://pocketbase:8090
  environment:
    - PUBLIC_POCKETBASE_URL=https://yourdomain.com
    - POCKETBASE_URL=http://pocketbase:8090
```

## Performance Comparison

### Before Optimization
```
Server-Side Request Time:
- DNS lookup: ~20-50ms
- Network round trip: ~10-100ms
- SSL handshake: ~50-200ms
- Total: ~80-350ms per request
```

### After Optimization
```
Server-Side Request Time:
- Docker network: ~0.1-1ms
- Total: <1ms per request
```

**Result**: 80-350x faster for server-side requests!

## Migration Guide

### For Existing Projects

If you have existing SvelteKit + PocketBase projects, update your `src/lib/pocketbase.ts`:

```typescript
import PocketBase from 'pocketbase';
import { browser } from '$app/environment';

// OLD (uses same URL for both)
// const pbUrl = import.meta.env.PUBLIC_POCKETBASE_URL || 'http://localhost:8090';

// NEW (environment-aware)
const pbUrl = browser
  ? (import.meta.env.PUBLIC_POCKETBASE_URL || 'http://localhost:8090')
  : (import.meta.env.POCKETBASE_URL || 'http://localhost:8090');

export const pb = new PocketBase(pbUrl);
```

### For New Projects

New projects deployed through the VPS Panel will automatically use this optimization.

## Verification

To verify the optimization is working:

### 1. Check Container Logs
```bash
docker logs vps-panel-yourapp-frontend-1
```

Look for any PocketBase API calls in the logs. Server-side requests will be nearly instant.

### 2. Network Inspection

In development, add logging:
```typescript
const pbUrl = browser
  ? (import.meta.env.PUBLIC_POCKETBASE_URL || 'http://localhost:8090')
  : (import.meta.env.POCKETBASE_URL || 'http://localhost:8090');

if (import.meta.env.DEV) {
  console.log(`PocketBase URL: ${pbUrl} (${browser ? 'browser' : 'server'})`);
}
```

### 3. Performance Monitoring

Server-side requests should complete in <5ms:

```typescript
// In your +page.server.ts
export async function load() {
  const start = Date.now();
  const records = await pb.collection('users').getFullList();
  const duration = Date.now() - start;

  console.log(`Fetched ${records.length} users in ${duration}ms`);
  return { records };
}
```

## Troubleshooting

### Issue: Server-side requests still going through public URL

**Check**:
1. Verify `POCKETBASE_URL` is set in docker-compose environment
2. Check that your code uses `browser` to select the URL
3. Ensure containers are on the same Docker network

**Solution**:
```bash
# Check environment variables in container
docker exec vps-panel-yourapp-frontend-1 env | grep POCKETBASE
```

### Issue: Client-side requests failing

**Check**:
1. Verify `PUBLIC_POCKETBASE_URL` is set correctly
2. Ensure Caddy is routing `/api/*` to PocketBase

**Solution**:
```bash
# Check Caddy configuration
cat /var/lib/vps-panel/caddy/yourproject.caddy
```

## Additional Resources

- SvelteKit Environment Variables: https://kit.svelte.dev/docs/modules#$env-static-public
- Docker Networking: https://docs.docker.com/network/
- PocketBase Documentation: https://pocketbase.io/docs/

## Summary

This optimization ensures that:
- ✅ Client-side requests work normally (through public domain)
- ✅ Server-side requests use Docker internal network (80-350x faster)
- ✅ Development and production environments work seamlessly
- ✅ No changes needed to your application logic
- ✅ Automatic configuration through VPS Panel deployment
