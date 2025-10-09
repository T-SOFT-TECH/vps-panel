<script lang="ts">
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { authStore } from '$lib/stores/auth.svelte';
	import { oauthAPI } from '$lib/api/oauth';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/Card.svelte';
	import Alert from '$lib/components/Alert.svelte';

	let connecting = $state(false);
	let disconnecting = $state(false);
	let message = $state('');

	// Check for OAuth callback success
	onMount(() => {
		const params = new URLSearchParams(window.location.search);
		if (params.get('github') === 'connected') {
			message = 'GitHub connected successfully!';
			authStore.fetchCurrentUser();
			// Clean up URL
			window.history.replaceState({}, '', '/settings');
		}
	});

	async function connectGitHub() {
		connecting = true;
		try {
			const { url } = await oauthAPI.getGitHubAuthURL();
			// Redirect to GitHub OAuth
			window.location.href = url;
		} catch (error) {
			console.error('Failed to connect GitHub:', error);
			message = 'Failed to connect GitHub. Please try again.';
			connecting = false;
		}
	}

	async function disconnectGitHub() {
		if (!confirm('Are you sure you want to disconnect GitHub?')) {
			return;
		}

		disconnecting = true;
		try {
			await oauthAPI.disconnectGitHub();
			await authStore.fetchCurrentUser();
			message = 'GitHub disconnected successfully';
		} catch (error) {
			console.error('Failed to disconnect GitHub:', error);
			message = 'Failed to disconnect GitHub. Please try again.';
		} finally {
			disconnecting = false;
		}
	}
</script>

<svelte:head>
	<title>Settings - VPS Panel</title>
</svelte:head>

<div class="max-w-4xl mx-auto space-y-6">
	<div in:fly={{ y: -20, duration: 400, delay: 0 }}>
		<h1 class="text-3xl font-bold text-zinc-100">Settings</h1>
		<p class="mt-1 text-sm text-zinc-400">Manage your account and integrations</p>
	</div>

	{#if message}
		<div in:fly={{ y: 20, duration: 400, delay: 50 }}>
			<Alert variant="success" dismissible ondismiss={() => message = ''}>
				{message}
			</Alert>
		</div>
	{/if}

	<!-- Profile Section -->
	<div in:fly={{ y: 20, duration: 400, delay: 100 }}>
		<Card>
			<h2 class="text-lg font-semibold text-zinc-100 mb-4">Profile</h2>

			<div class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-zinc-300 mb-1">Name</label>
					<p class="text-zinc-100">{authStore.user?.name || 'Not set'}</p>
				</div>

				<div>
					<label class="block text-sm font-medium text-zinc-300 mb-1">Email</label>
					<p class="text-zinc-100">{authStore.user?.email}</p>
				</div>

				<div>
					<label class="block text-sm font-medium text-zinc-300 mb-1">Role</label>
					<p class="text-zinc-100 capitalize">{authStore.user?.role}</p>
				</div>
			</div>
		</Card>
	</div>

	<!-- Git Providers -->
	<div in:fly={{ y: 20, duration: 400, delay: 150 }}>
		<Card>
			<div class="flex items-center justify-between mb-4">
				<div>
					<h2 class="text-lg font-semibold text-zinc-100">Git Providers</h2>
					<p class="text-sm text-zinc-400 mt-1">
						Manage OAuth connections for GitHub, Gitea, and more
					</p>
				</div>
				<a href="/settings/git-providers">
					<Button>Manage Providers</Button>
				</a>
			</div>

			<p class="text-sm text-zinc-400">
				Configure your Git providers to seamlessly import and deploy repositories. You can add multiple accounts for GitHub, self-hosted Gitea instances, and more.
			</p>
		</Card>
	</div>

	<!-- Legacy Connected Accounts (Deprecated - kept for backward compatibility) -->
	<div in:fly={{ y: 20, duration: 400, delay: 150 }}>
		<Card>
			<h2 class="text-lg font-semibold text-zinc-100 mb-4">Connected Accounts (Legacy)</h2>
			<p class="text-sm text-zinc-400 mb-6">
				⚠️ This section is deprecated. Please use Git Providers above for better management.
			</p>

			<!-- GitHub -->
			<div class="flex items-center justify-between p-4 rounded-lg border border-zinc-800">
				<div class="flex items-center space-x-4">
					<div class="flex-shrink-0">
						<svg class="w-10 h-10 text-zinc-100" fill="currentColor" viewBox="0 0 24 24">
							<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
						</svg>
					</div>
					<div>
						<h3 class="text-base font-medium text-zinc-100">GitHub</h3>
						{#if authStore.user?.github_connected}
							<p class="text-sm text-green-500 flex items-center mt-1">
								<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
								</svg>
								Connected as @{authStore.user.github_username}
							</p>
						{:else}
							<p class="text-sm text-zinc-400 mt-1">Not connected</p>
						{/if}
					</div>
				</div>

				<div>
					{#if authStore.user?.github_connected}
						<Button
							variant="secondary"
							onclick={disconnectGitHub}
							loading={disconnecting}
							disabled={disconnecting}
						>
							{disconnecting ? 'Disconnecting...' : 'Disconnect'}
						</Button>
					{:else}
						<Button
							onclick={connectGitHub}
							loading={connecting}
							disabled={connecting}
						>
							{connecting ? 'Connecting...' : 'Connect GitHub'}
						</Button>
					{/if}
				</div>
			</div>

			<!-- GitLab (Coming Soon) -->
			<div class="flex items-center justify-between p-4 rounded-lg border border-zinc-800 mt-4 opacity-50">
				<div class="flex items-center space-x-4">
					<div class="flex-shrink-0">
						<svg class="w-10 h-10 text-orange-500" viewBox="0 0 24 24" fill="currentColor">
							<path d="M23.955 13.587l-1.342-4.135-2.664-8.189a.455.455 0 0 0-.867 0L16.418 9.45H7.582L4.919 1.263a.455.455 0 0 0-.867 0L1.388 9.452.046 13.587a.924.924 0 0 0 .331 1.023L12 23.054l11.623-8.443a.92.92 0 0 0 .332-1.024"/>
						</svg>
					</div>
					<div>
						<h3 class="text-base font-medium text-zinc-100">GitLab</h3>
						<p class="text-sm text-zinc-400 mt-1">Coming soon</p>
					</div>
				</div>
				<Button variant="secondary" disabled>
					Coming Soon
				</Button>
			</div>
		</Card>
	</div>

	<!-- Danger Zone -->
	<div in:fly={{ y: 20, duration: 400, delay: 200 }}>
		<Card>
			<h2 class="text-lg font-semibold text-red-500 mb-4">Danger Zone</h2>

			<div class="space-y-4">
				<div class="flex items-center justify-between p-4 rounded-lg border border-red-900/50 bg-red-950/20">
					<div>
						<h3 class="text-base font-medium text-zinc-100">Delete Account</h3>
						<p class="text-sm text-zinc-400 mt-1">
							Permanently delete your account and all associated data
						</p>
					</div>
					<Button variant="secondary" disabled>
						Delete Account
					</Button>
				</div>
			</div>
		</Card>
	</div>
</div>
