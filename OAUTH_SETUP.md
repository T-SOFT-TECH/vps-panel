# GitHub OAuth Integration Setup

## What We've Built

### Backend (✅ Completed)
1. **OAuth Configuration** - Added to `internal/config/config.go`
   - `GITHUB_CLIENT_ID`
   - `GITHUB_CLIENT_SECRET`
   - `OAUTH_CALLBACK_URL`

2. **User Model Updates** - `internal/models/user.go`
   - Added `github_connected`, `github_token`, `github_username` fields
   - Tokens are never sent to frontend (`json:"-"`)

3. **GitHub OAuth Service** - `internal/services/oauth/github.go`
   - `GetAuthURL()` - Initiates OAuth flow
   - `ExchangeCode()` - Exchanges code for token
   - `GetUser()` - Fetches GitHub user info
   - `ListRepositories()` - Lists user's repos

4. **OAuth Handlers** - `internal/api/handlers/auth.go`
   - `GitHubOAuthInit` - GET `/api/v1/auth/oauth/github/init`
   - `GitHubOAuthCallback` - GET `/api/v1/auth/oauth/callback/github`
   - `DisconnectGitHub` - GET `/api/v1/auth/oauth/github/disconnect`
   - `ListGitHubRepositories` - GET `/api/v1/auth/oauth/github/repositories`

5. **Database Migrations**
   - `migrations/003_add_git_credentials.sql` - For project git credentials
   - `migrations/004_add_oauth_fields.sql` - For user OAuth fields

### Frontend (🚧 In Progress)
1. **Type Definitions** - Updated `lib/types.ts`
   - Added OAuth fields to User interface
   - Added GitHubRepository interface

2. **OAuth API Client** - Created `lib/api/oauth.ts`
   - Methods to interact with OAuth endpoints

## Setup Instructions

### 1. Create GitHub OAuth App

1. Go to https://github.com/settings/developers
2. Click "New OAuth App"
3. Fill in:
   - **Application name**: VPS Panel (or your app name)
   - **Homepage URL**: `http://localhost:5173` (dev) or your production URL
   - **Authorization callback URL**: `http://localhost:8080/api/v1/auth/oauth/callback/github`
4. Click "Register application"
5. Copy the **Client ID** and generate a **Client Secret**

### 2. Configure Backend Environment

Create a `.env` file in the `backend` directory:

```bash
# OAuth Configuration
GITHUB_CLIENT_ID=your_github_client_id_here
GITHUB_CLIENT_SECRET=your_github_client_secret_here
OAUTH_CALLBACK_URL=http://localhost:8080/api/v1/auth/oauth/callback

# Other existing config...
JWT_SECRET=your-secret-key
PORT=8080
CORS_ORIGINS=http://localhost:5173,http://localhost:4173
```

### 3. Install Required Go Packages

```bash
cd backend
go get golang.org/x/oauth2
go get golang.org/x/oauth2/github
```

### 4. Run Database Migrations

The migrations will be applied automatically on startup, or you can run them manually:

```bash
# Apply migrations
sqlite3 data/vps-panel.db < migrations/003_add_git_credentials.sql
sqlite3 data/vps-panel.db < migrations/004_add_oauth_fields.sql
```

### 5. Rebuild and Restart Backend

```bash
cd backend
go build -o vps-panel.exe ./cmd/server
./vps-panel.exe
```

## Frontend Implementation (Next Steps)

### Create Settings Page for OAuth Connections

Create `frontend/src/routes/(app)/settings/+page.svelte`:

```svelte
<script lang="ts">
	import { authStore } from '$lib/stores/auth.svelte';
	import { oauthAPI } from '$lib/api/oauth';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/Card.svelte';

	let connecting = $state(false);

	async function connectGitHub() {
		connecting = true;
		try {
			const { url } = await oauthAPI.getGitHubAuthURL();
			window.location.href = url;
		} catch (error) {
			console.error('Failed to initiate GitHub OAuth:', error);
			connecting = false;
		}
	}

	async function disconnectGitHub() {
		try {
			await oauthAPI.disconnectGitHub();
			await authStore.fetchCurrentUser();
		} catch (error) {
			console.error('Failed to disconnect GitHub:', error);
		}
	}
</script>

<div class="max-w-4xl mx-auto space-y-6">
	<h1 class="text-3xl font-bold text-zinc-100">Settings</h1>

	<Card>
		<h2 class="text-lg font-semibold text-zinc-100 mb-4">Connected Accounts</h2>

		<div class="flex items-center justify-between py-4 border-b border-zinc-800">
			<div class="flex items-center space-x-3">
				<svg class="w-8 h-8" viewBox="0 0 24 24" fill="currentColor">
					<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
				</svg>
				<div>
					<p class="font-medium text-zinc-100">GitHub</p>
					{#if authStore.user?.github_connected}
						<p class="text-sm text-zinc-400">Connected as @{authStore.user.github_username}</p>
					{:else}
						<p class="text-sm text-zinc-400">Not connected</p>
					{/if}
				</div>
			</div>

			{#if authStore.user?.github_connected}
				<Button variant="secondary" onclick={disconnectGitHub}>
					Disconnect
				</Button>
			{:else}
				<Button loading={connecting} onclick={connectGitHub}>
					Connect GitHub
				</Button>
			{/if}
		</div>
	</Card>
</div>
```

### Update New Project Page to Use GitHub Repos

In `frontend/src/routes/(app)/projects/new/+page.svelte`, add repository selection:

```typescript
import { oauthAPI } from '$lib/api/oauth';
import { authStore } from '$lib/stores/auth.svelte';

let githubRepos = $state<GitHubRepository[]>([]);
let loadingRepos = $state(false);
let selectedRepo = $state<GitHubRepository | null>(null);

async function loadGitHubRepos() {
	if (!authStore.user?.github_connected) return;

	loadingRepos = true;
	try {
		const { repositories } = await oauthAPI.listGitHubRepositories();
		githubRepos = repositories;
	} catch (error) {
		console.error('Failed to load repositories:', error);
	} finally {
		loadingRepos = false;
	}
}

// Auto-load repos when component mounts if GitHub is connected
$effect(() => {
	if (authStore.user?.github_connected) {
		loadGitHubRepos();
	}
});

function selectRepo(repo: GitHubRepository) {
	selectedRepo = repo;
	gitUrl = repo.clone_url;
	gitBranch = repo.default_branch;
	name = repo.name;
}
```

## How It Works

### OAuth Flow

1. **User clicks "Connect GitHub"** → Frontend calls `/auth/oauth/github/init`
2. **Backend generates auth URL** → Returns GitHub OAuth URL with state
3. **User redirects to GitHub** → Authorizes the application
4. **GitHub redirects back** → To `/api/v1/auth/oauth/callback/github?code=...&state=...`
5. **Backend exchanges code** → Gets access token from GitHub
6. **Backend fetches user info** → Gets GitHub username
7. **Backend saves to database** → Updates user record with token
8. **Redirect to frontend** → `/settings?github=connected`

### Using Connected Repos

1. User goes to "New Project" page
2. If GitHub connected, automatically fetches repositories
3. User selects a repository from the list
4. Repository details auto-fill the form
5. Backend uses stored OAuth token for git operations (no manual credentials needed)

## Benefits

✅ **No manual token entry** - Users just click "Connect GitHub"
✅ **Auto repository discovery** - Fetches all user's repos
✅ **Seamless experience** - Like Vercel, Netlify, etc.
✅ **Secure** - Tokens never exposed to frontend
✅ **Works with private repos** - OAuth token has repo access
✅ **Auto-fill project details** - Repository name, URL, branch

## Testing

1. Start backend: `./vps-panel.exe`
2. Start frontend: `npm run dev`
3. Login to your account
4. Go to Settings → Connect GitHub
5. Authorize the app
6. Go to New Project → See your repositories listed

## Production Deployment

When deploying to production:

1. Update GitHub OAuth App callback URL to production domain
2. Update `OAUTH_CALLBACK_URL` env var
3. Update `CORS_ORIGINS` to include production frontend URL
4. Ensure HTTPS is enabled (required for OAuth in production)
