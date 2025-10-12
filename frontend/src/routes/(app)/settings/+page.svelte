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

<div class="max-w-7xl mx-auto space-y-6 pb-8">
	<!-- Header with Gradient Background -->
	<div class="relative overflow-hidden rounded-2xl slide-in-down">
		<div class="absolute inset-0 mesh-gradient opacity-50"></div>
		<div class="relative glass-pro p-6 border-0">
			<div class="flex items-center gap-4">
				<div class="w-16 h-16 rounded-2xl bg-gradient-brand flex items-center justify-center shadow-xl glow-green float">
					<svg class="w-9 h-9 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
					</svg>
				</div>
				<div>
					<h1 class="text-4xl font-bold bg-gradient-to-r from-primary-700 to-primary-900 bg-clip-text text-transparent mb-1">
						Settings
					</h1>
					<p class="text-base" style="color: rgb(var(--text-secondary));">
						Manage your account and integrations
					</p>
				</div>
			</div>
		</div>
	</div>

	{#if message}
		<div in:fly={{ y: 20, duration: 400, delay: 50 }}>
			<Alert variant="success" dismissible ondismiss={() => message = ''}>
				{message}
			</Alert>
		</div>
	{/if}

	<!-- Profile Section -->
	<div class="slide-in-up stagger-1">
		<div class="relative overflow-hidden rounded-2xl">
			<div class="absolute inset-0 bg-gradient-to-br from-primary-600/5 to-primary-800/5"></div>
			<div class="relative modern-card p-6 border-0 hover-lift transition-all">
				<div class="flex items-center gap-3 mb-6">
					<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center shadow-lg">
						<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
						</svg>
					</div>
					<h2 class="text-xl font-bold" style="color: rgb(var(--text-primary));">Profile</h2>
				</div>

				<div class="grid grid-cols-1 sm:grid-cols-3 gap-6">
					<div class="p-4 rounded-xl" style="background-color: rgb(var(--bg-secondary)); border: 1px solid rgb(var(--border-primary));">
						<label class="block text-xs font-semibold mb-2" style="color: rgb(var(--text-tertiary));">Name</label>
						<p class="font-semibold" style="color: rgb(var(--text-primary));">{authStore.user?.name || 'Not set'}</p>
					</div>

					<div class="p-4 rounded-xl" style="background-color: rgb(var(--bg-secondary)); border: 1px solid rgb(var(--border-primary));">
						<label class="block text-xs font-semibold mb-2" style="color: rgb(var(--text-tertiary));">Email</label>
						<p class="font-semibold truncate" style="color: rgb(var(--text-primary));">{authStore.user?.email}</p>
					</div>

					<div class="p-4 rounded-xl" style="background-color: rgb(var(--bg-secondary)); border: 1px solid rgb(var(--border-primary));">
						<label class="block text-xs font-semibold mb-2" style="color: rgb(var(--text-tertiary));">Role</label>
						<p class="font-semibold capitalize" style="color: rgb(var(--text-primary));">{authStore.user?.role}</p>
					</div>
				</div>
			</div>
		</div>
	</div>

	<!-- Git Providers -->
	<div class="slide-in-up stagger-2">
		<div class="relative overflow-hidden rounded-2xl">
			<div class="absolute inset-0 bg-gradient-to-br from-primary-600/5 to-primary-800/5"></div>
			<div class="relative modern-card p-6 border-0 hover-lift transition-all">
				<div class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4 mb-4">
					<div class="flex items-center gap-3">
						<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500 to-emerald-600 flex items-center justify-center shadow-lg">
							<svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 24 24">
								<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
							</svg>
						</div>
						<div>
							<h2 class="text-xl font-bold" style="color: rgb(var(--text-primary));">Git Providers</h2>
							<p class="text-sm mt-1" style="color: rgb(var(--text-secondary));">
								Manage OAuth connections for GitHub, Gitea, and more
							</p>
						</div>
					</div>
					<a href="/settings/git-providers">
						<Button class="btn-primary glow-green-hover hover:scale-105 transition-transform whitespace-nowrap">
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
							</svg>
							Manage Providers
						</Button>
					</a>
				</div>

				<p class="text-sm p-4 rounded-xl" style="color: rgb(var(--text-secondary)); background-color: rgb(var(--bg-secondary)); border: 1px solid rgb(var(--border-primary));">
					Configure your Git providers to seamlessly import and deploy repositories. You can add multiple accounts for GitHub, self-hosted Gitea instances, and more.
				</p>
			</div>
		</div>
	</div>

	<!-- Legacy Connected Accounts (Deprecated - kept for backward compatibility) -->
	<div in:fly={{ y: 20, duration: 400, delay: 150 }}>
		<Card>
			<h2 class="text-lg font-semibold mb-4" style="color: rgb(var(--text-primary));">Connected Accounts (Legacy)</h2>
			<p class="text-sm mb-6" style="color: rgb(var(--text-tertiary));">
				⚠️ This section is deprecated. Please use Git Providers above for better management.
			</p>

			<!-- GitHub -->
			<div class="flex items-center justify-between p-4 rounded-lg border" style="border-color: rgb(var(--border-primary));">
				<div class="flex items-center space-x-4">
					<div class="flex-shrink-0">
						<svg class="w-10 h-10" style="color: rgb(var(--text-primary));" fill="currentColor" viewBox="0 0 24 24">
							<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
						</svg>
					</div>
					<div>
						<h3 class="text-base font-medium" style="color: rgb(var(--text-primary));">GitHub</h3>
						{#if authStore.user?.github_connected}
							<p class="text-sm flex items-center mt-1 bg-primary-800" style="color: rgb(var(--text-primary));">
								<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
								</svg>
								Connected as @{authStore.user.github_username}
							</p>
						{:else}
							<p class="text-sm mt-1" style="color: rgb(var(--text-tertiary));">Not connected</p>
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
			<div class="flex items-center justify-between p-4 rounded-lg border mt-4 opacity-50" style="border-color: rgb(var(--border-primary));">
				<div class="flex items-center space-x-4">
					<div class="flex-shrink-0">
						<svg class="w-10 h-10 text-orange-500" viewBox="0 0 24 24" fill="currentColor">
							<path d="M23.955 13.587l-1.342-4.135-2.664-8.189a.455.455 0 0 0-.867 0L16.418 9.45H7.582L4.919 1.263a.455.455 0 0 0-.867 0L1.388 9.452.046 13.587a.924.924 0 0 0 .331 1.023L12 23.054l11.623-8.443a.92.92 0 0 0 .332-1.024"/>
						</svg>
					</div>
					<div>
						<h3 class="text-base font-medium" style="color: rgb(var(--text-primary));">GitLab</h3>
						<p class="text-sm mt-1" style="color: rgb(var(--text-tertiary));">Coming soon</p>
					</div>
				</div>
				<Button variant="secondary" disabled>
					Coming Soon
				</Button>
			</div>
		</Card>
	</div>

	<!-- Danger Zone -->
	<div class="slide-in-up stagger-4">
		<div class="relative overflow-hidden rounded-2xl border-2 border-red-900/50">
			<div class="absolute inset-0 bg-gradient-to-br from-red-500/5 to-red-600/10"></div>
			<div class="relative modern-card p-6 border-0">
				<div class="flex items-center gap-3 mb-6">
					<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-red-500 to-red-600 flex items-center justify-center shadow-lg">
						<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
						</svg>
					</div>
					<h2 class="text-xl font-bold text-red-500">Danger Zone</h2>
				</div>

				<div class="p-5 rounded-xl border-2 border-red-900/50" style="background: linear-gradient(135deg, rgba(239, 68, 68, 0.05) 0%, rgba(220, 38, 38, 0.1) 100%);">
					<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
						<div>
							<h3 class="text-base font-bold flex items-center gap-2" style="color: rgb(var(--text-primary));">
								<svg class="w-5 h-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
								</svg>
								Delete Account
							</h3>
							<p class="text-sm mt-1" style="color: rgb(var(--text-secondary));">
								Permanently delete your account and all associated data
							</p>
						</div>
						<Button variant="secondary" disabled class="whitespace-nowrap">
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
							</svg>
							Locked
						</Button>
					</div>
				</div>
			</div>
		</div>
	</div>
</div>
