# Auto-Deploy with Webhooks

Automatically deploy your projects when you push code to your repository using webhooks.

## Features

- **Automatic Deployments**: Push code and your project deploys automatically
- **Secure**: HMAC SHA-256 signature verification for GitHub and Gitea, token-based for GitLab
- **Per-Project Configuration**: Each project has its own webhook secret
- **Branch Filtering**: Only deploy pushes to specified branches
- **Multi-Provider Support**: Works with GitHub, GitLab, and Gitea
- **Async Deployment**: Webhook returns immediately while deployment runs in background

## How It Works

1. **Enable Auto-Deploy**: Go to your project settings and enable auto-deploy
2. **Get Webhook URL**: Copy the webhook URL and secret for your Git provider
3. **Configure Webhook**: Add the webhook to your repository settings
4. **Push Code**: When you push to the configured branch, deployment triggers automatically

## Enabling Auto-Deploy

### Via Web UI

1. Navigate to your project page
2. Find the "Auto-Deploy" card in the sidebar
3. Click "Enable Auto-Deploy"
4. Click "View Webhook URLs" to get configuration details
5. Follow the provider-specific instructions

### Via API

```bash
# Enable auto-deploy
curl -X POST https://your-panel.com/api/v1/projects/{project_id}/webhook/enable \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json"

# Get webhook info
curl https://your-panel.com/api/v1/projects/{project_id}/webhook \
  -H "Authorization: Bearer YOUR_TOKEN"

# Disable auto-deploy
curl -X POST https://your-panel.com/api/v1/projects/{project_id}/webhook/disable \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Setting Up Webhooks

### GitHub

1. Go to your repository → Settings → Webhooks
2. Click "Add webhook"
3. **Payload URL**: `https://your-panel.com/api/v1/webhooks/github/{project_id}`
4. **Content type**: `application/json`
5. **Secret**: Your webhook secret from the VPS Panel
6. **Events**: Select "Just the push event"
7. **Active**: Check this box
8. Click "Add webhook"

#### Verification

GitHub uses HMAC SHA-256 with the `X-Hub-Signature-256` header.

### GitLab

1. Go to your repository → Settings → Webhooks
2. **URL**: `https://your-panel.com/api/v1/webhooks/gitlab/{project_id}`
3. **Secret Token**: Your webhook secret from the VPS Panel
4. **Trigger**: Check only "Push events"
5. **SSL verification**: Enable (recommended)
6. Click "Add webhook"

#### Verification

GitLab uses a simple token comparison with the `X-Gitlab-Token` header.

### Gitea

1. Go to your repository → Settings → Webhooks
2. Click "Add Webhook" → "Gitea"
3. **Target URL**: `https://your-panel.com/api/v1/webhooks/gitea/{project_id}`
4. **HTTP Method**: POST
5. **POST Content Type**: `application/json`
6. **Secret**: Your webhook secret from the VPS Panel
7. **Trigger**: Select "Push" event only
8. **Active**: Check this box
9. Click "Add Webhook"

#### Verification

Gitea uses HMAC SHA-256 with the `X-Gitea-Signature` header.

## Webhook Payloads

### GitHub Payload

```json
{
  "ref": "refs/heads/main",
  "repository": {
    "clone_url": "https://github.com/user/repo.git",
    "html_url": "https://github.com/user/repo"
  },
  "head_commit": {
    "id": "abc123...",
    "message": "Update feature",
    "author": {
      "name": "John Doe",
      "email": "john@example.com"
    }
  }
}
```

### GitLab Payload

```json
{
  "ref": "refs/heads/main",
  "project": {
    "http_url": "https://gitlab.com/user/repo.git",
    "ssh_url": "git@gitlab.com:user/repo.git"
  },
  "commits": [
    {
      "id": "abc123...",
      "message": "Update feature",
      "author": {
        "name": "John Doe",
        "email": "john@example.com"
      }
    }
  ]
}
```

### Gitea Payload

```json
{
  "ref": "refs/heads/main",
  "repository": {
    "clone_url": "https://gitea.example.com/user/repo.git",
    "html_url": "https://gitea.example.com/user/repo"
  },
  "head_commit": {
    "id": "abc123...",
    "message": "Update feature",
    "author": {
      "name": "John Doe",
      "email": "john@example.com"
    }
  }
}
```

## Webhook Response

### Success Response (201 Created)

```json
{
  "message": "Deployment triggered successfully",
  "deployment_id": 123,
  "project_id": 45,
  "branch": "main",
  "commit": "abc123"
}
```

### Error Responses

#### 400 Bad Request
```json
{
  "error": "Invalid project ID"
}
```

#### 401 Unauthorized
```json
{
  "error": "Invalid signature"
}
```

#### 404 Not Found
```json
{
  "error": "Project not found"
}
```

#### Auto-Deploy Disabled
```json
{
  "message": "Auto-deploy is disabled for this project"
}
```

#### Wrong Branch
```json
{
  "message": "Push to develop ignored. Auto-deploy configured for main"
}
```

## Configuration

### Branch Configuration

By default, auto-deploy is configured for your project's main branch (`git_branch`). You can customize this:

```sql
-- Set custom auto-deploy branch
UPDATE projects
SET auto_deploy_branch = 'production'
WHERE id = 1;
```

### Webhook Secret Rotation

To rotate your webhook secret:

1. Disable auto-deploy
2. Enable auto-deploy again (generates new secret)
3. Update the secret in your Git provider

Or via SQL:

```sql
-- Generate new secret
UPDATE projects
SET webhook_secret = 'your-new-random-secret-32-chars'
WHERE id = 1;
```

## Monitoring Deployments

### Via Web UI

1. Go to your project page
2. View "Recent Deployments" section
3. Deployments triggered by webhooks show `triggered_by: "webhook-{provider}"`

### Via API

```bash
# Get all deployments
curl https://your-panel.com/api/v1/projects/{project_id}/deployments \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get specific deployment
curl https://your-panel.com/api/v1/projects/{project_id}/deployments/{deployment_id} \
  -H "Authorization: Bearer YOUR_TOKEN"

# View deployment logs
curl https://your-panel.com/api/v1/projects/{project_id}/deployments/{deployment_id}/logs \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Security Best Practices

1. **Use HTTPS**: Always use HTTPS for webhook URLs in production
2. **Verify Signatures**: The VPS Panel automatically verifies webhook signatures
3. **Rotate Secrets**: Rotate webhook secrets periodically
4. **Monitor Activity**: Check deployment logs regularly
5. **Branch Protection**: Configure branch protection rules in your Git provider

## Troubleshooting

### Webhook Not Triggering

1. **Check Webhook Configuration**:
   - Verify URL is correct and includes project ID
   - Verify secret matches
   - Check webhook is active

2. **Check Branch**:
   - Ensure you're pushing to the configured branch
   - Check `auto_deploy_branch` or defaults to `git_branch`

3. **Check Logs**:
   - View webhook delivery logs in your Git provider
   - Check VPS Panel logs for errors

### Deployment Fails

1. **Check Build Logs**:
   - View deployment logs in the UI
   - Look for build errors or missing dependencies

2. **Check Credentials**:
   - Verify Git credentials are valid
   - Check private repository access

3. **Check Resources**:
   - Ensure enough disk space
   - Check Docker is running
   - Verify ports are available

### Signature Verification Fails

1. **GitHub/Gitea**:
   - Verify secret matches exactly
   - Check `X-Hub-Signature-256` or `X-Gitea-Signature` header is present

2. **GitLab**:
   - Verify secret token matches exactly
   - Check `X-Gitlab-Token` header is present

## Database Schema

```sql
-- Projects table webhook fields
ALTER TABLE projects ADD COLUMN webhook_secret TEXT;
ALTER TABLE projects ADD COLUMN auto_deploy_branch TEXT;
ALTER TABLE projects ADD COLUMN auto_deploy BOOLEAN DEFAULT false;

-- Index for faster lookups
CREATE INDEX idx_projects_auto_deploy ON projects(auto_deploy)
WHERE auto_deploy = true;
```

## API Reference

### Enable Webhook
- **Endpoint**: `POST /api/v1/projects/:id/webhook/enable`
- **Auth**: Required (JWT)
- **Response**: Webhook configuration with URLs and secret

### Disable Webhook
- **Endpoint**: `POST /api/v1/projects/:id/webhook/disable`
- **Auth**: Required (JWT)
- **Response**: Success message

### Get Webhook Info
- **Endpoint**: `GET /api/v1/projects/:id/webhook`
- **Auth**: Required (JWT)
- **Response**: Current webhook configuration

### Receive Webhook (GitHub)
- **Endpoint**: `POST /api/v1/webhooks/github/:project_id`
- **Auth**: Signature verification (X-Hub-Signature-256)
- **Response**: Deployment created

### Receive Webhook (GitLab)
- **Endpoint**: `POST /api/v1/webhooks/gitlab/:project_id`
- **Auth**: Token verification (X-Gitlab-Token)
- **Response**: Deployment created

### Receive Webhook (Gitea)
- **Endpoint**: `POST /api/v1/webhooks/gitea/:project_id`
- **Auth**: Signature verification (X-Gitea-Signature)
- **Response**: Deployment created

## Environment Variables

Add these to your `.env` file:

```env
# Panel URL for webhook URL generation
PANEL_URL=https://panel.yourdomain.com

# Optional: Panel domain for auto-subdomains
PANEL_DOMAIN=yourdomain.com
```

## Examples

### Testing Webhooks Locally

For local testing, use a tool like ngrok:

```bash
# Start ngrok
ngrok http 8080

# Use the ngrok URL in webhook configuration
# Example: https://abc123.ngrok.io/api/v1/webhooks/github/1
```

### Curl Testing

```bash
# Test GitHub webhook
curl -X POST https://your-panel.com/api/v1/webhooks/github/1 \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=YOUR_SIGNATURE" \
  -d '{
    "ref": "refs/heads/main",
    "repository": {
      "clone_url": "https://github.com/user/repo.git"
    },
    "head_commit": {
      "id": "abc123",
      "message": "Test commit",
      "author": {
        "name": "Test User",
        "email": "test@example.com"
      }
    }
  }'
```

## Migration Guide

### From Manual Deployments

1. Ensure project deploys successfully manually
2. Enable auto-deploy in project settings
3. Configure webhook in Git provider
4. Test with a small commit
5. Monitor first automatic deployment

### Updating Existing Projects

Run migration:

```bash
# Apply migration
sqlite3 data/vps-panel.db < backend/migrations/008_add_webhook_fields.sql
```

Or let the application auto-migrate on startup.

## Support

For issues or questions:
- Check the [GitHub Issues](https://github.com/yourusername/vps-panel/issues)
- Review deployment logs
- Check Git provider webhook delivery logs
- Verify signature/token configuration
