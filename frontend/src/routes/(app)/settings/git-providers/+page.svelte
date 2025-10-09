<script lang="ts">
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { gitProvidersAPI } from '$lib/api/git-providers';
	import type { GitProvider, ProviderType } from '$lib/types';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/Card.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';

	let providers = $state<GitProvider[]>([]);
	let loading = $state(false);
	let error = $state('');
	let success = $state('');

	// Add/Edit Modal
	let showModal = $state(false);
	let editingProvider = $state<GitProvider | null>(null);
	let modalLoading = $state(false);

	// Form fields
	let providerType = $state<ProviderType>('github');
	let providerName = $state('');
	let providerUrl = $state('');
	let clientId = $state('');
	let clientSecret = $state('');
	let isDefault = $state(false);

	// OAuth callback URL - computed from current origin
	let callbackUrl = $state('');

	const providerTypeOptions = [
		{ value: 'github', label: 'GitHub' },
		{ value: 'gitea', label: 'Gitea (Self-hosted)' },
		{ value: 'gitlab', label: 'GitLab (Coming Soon)', disabled: true }
	];

	onMount(() => {
		loadProviders();

		// Set OAuth callback URL from current origin
		callbackUrl = `${window.location.origin}/api/v1/auth/oauth/callback`;

		// Check if redirected after OAuth connection
		const params = new URLSearchParams(window.location.search);
		if (params.get('connected') === 'true') {
			success = 'Git provider connected successfully!';
			window.history.replaceState({}, '', '/settings/git-providers');
			loadProviders();
		}
	});

	async function loadProviders() {
		loading = true;
		try {
			const data = await gitProvidersAPI.getAll();
			providers = data.providers || [];
		} catch (err) {
			error = 'Failed to load providers';
			console.error(err);
		} finally {
			loading = false;
		}
	}

	function openAddModal() {
		editingProvider = null;
		providerType = 'github';
		providerName = '';
		providerUrl = '';
		clientId = '';
		clientSecret = '';
		isDefault = false;
		showModal = true;
	}

	function openEditModal(provider: GitProvider) {
		editingProvider = provider;
		providerType = provider.type;
		providerName = provider.name;
		providerUrl = provider.url || '';
		clientId = '';
		clientSecret = '';
		isDefault = provider.is_default;
		showModal = true;
	}

	async function handleSubmit() {
		if (!providerName || !clientId || !clientSecret) {
			error = 'Please fill in all required fields';
			return;
		}

		if ((providerType === 'gitea' || providerType === 'gitlab') && !providerUrl) {
			error = 'URL is required for self-hosted providers';
			return;
		}

		modalLoading = true;
		error = '';

		try {
			if (editingProvider) {
				await gitProvidersAPI.update(editingProvider.id, {
					name: providerName,
					url: providerUrl,
					client_id: clientId,
					client_secret: clientSecret,
					is_default: isDefault
				});
				success = 'Provider updated successfully';
			} else {
				await gitProvidersAPI.create({
					type: providerType,
					name: providerName,
					url: providerUrl,
					client_id: clientId,
					client_secret: clientSecret,
					is_default: isDefault
				});
				success = 'Provider added successfully';
			}

			showModal = false;
			await loadProviders();
		} catch (err: any) {
			error = err.message || 'Failed to save provider';
		} finally {
			modalLoading = false;
		}
	}

	async function handleDelete(provider: GitProvider) {
		if (!confirm(`Are you sure you want to delete "${provider.name}"?`)) {
			return;
		}

		try {
			await gitProvidersAPI.delete(provider.id);
			success = 'Provider deleted successfully';
			await loadProviders();
		} catch (err: any) {
			error = err.message || 'Failed to delete provider';
		}
	}

	async function handleConnect(provider: GitProvider) {
		try {
			const { url } = await gitProvidersAPI.initiateOAuth(provider.id);
			window.location.href = url;
		} catch (err: any) {
			error = err.message || 'Failed to initiate OAuth';
		}
	}

	async function handleDisconnect(provider: GitProvider) {
		if (!confirm(`Disconnect ${provider.name}?`)) {
			return;
		}

		try {
			await gitProvidersAPI.disconnect(provider.id);
			success = 'Provider disconnected successfully';
			await loadProviders();
		} catch (err: any) {
			error = err.message || 'Failed to disconnect provider';
		}
	}

	function getProviderIcon(type: ProviderType) {
		switch (type) {
			case 'github':
				return `<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>`;
			case 'gitea':
				return `<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>`;
			case 'gitlab':
				return `<path d="M23.955 13.587l-1.342-4.135-2.664-8.189a.455.455 0 0 0-.867 0L16.418 9.45H7.582L4.919 1.263a.455.455 0 0 0-.867 0L1.388 9.452.046 13.587a.924.924 0 0 0 .331 1.023L12 23.054l11.623-8.443a.92.92 0 0 0 .332-1.024"/>`;
			default:
				return '';
		}
	}

	function getProviderColor(type: ProviderType) {
		switch (type) {
			case 'github':
				return 'text-zinc-100';
			case 'gitea':
				return 'text-green-500';
			case 'gitlab':
				return 'text-orange-500';
			default:
				return 'text-zinc-400';
		}
	}
</script>

<svelte:head>
	<title>Git Providers - VPS Panel</title>
</svelte:head>

<div class="max-w-6xl mx-auto">
	<div class="mb-6 flex items-center justify-between" in:fly={{ y: -20, duration: 400 }}>
		<div>
			<h1 class="text-3xl font-bold text-zinc-100">Git Providers</h1>
			<p class="mt-1 text-sm text-zinc-400">
				Manage your Git OAuth providers for seamless repository integration
			</p>
		</div>
		<Button onclick={openAddModal}>Add Provider</Button>
	</div>

	{#if error}
		<div in:fly={{ y: 20, duration: 400 }} class="mb-4">
			<Alert variant="error" dismissible ondismiss={() => (error = '')}>
				{error}
			</Alert>
		</div>
	{/if}

	{#if success}
		<div in:fly={{ y: 20, duration: 400 }} class="mb-4">
			<Alert variant="success" dismissible ondismiss={() => (success = '')}>
				{success}
			</Alert>
		</div>
	{/if}

	{#if loading}
		<div class="text-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-green-500 mx-auto"></div>
			<p class="mt-4 text-zinc-400">Loading providers...</p>
		</div>
	{:else if providers.length === 0}
		<Card>
			<div class="text-center py-12">
				<svg
					class="mx-auto h-12 w-12 text-zinc-500"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M12 6v6m0 0v6m0-6h6m-6 0H6"
					/>
				</svg>
				<h3 class="mt-2 text-sm font-medium text-zinc-100">No Git providers</h3>
				<p class="mt-1 text-sm text-zinc-400">Get started by adding your first Git provider.</p>
				<div class="mt-6">
					<Button onclick={openAddModal}>Add Provider</Button>
				</div>
			</div>
		</Card>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
			{#each providers as provider (provider.id)}
				<div in:fly={{ y: 20, duration: 400 }}>
					<Card>
						<div class="flex items-start justify-between">
							<div class="flex items-start space-x-4">
								<svg class="w-10 h-10 {getProviderColor(provider.type)}" fill="currentColor" viewBox="0 0 24 24">
									{@html getProviderIcon(provider.type)}
								</svg>
								<div>
									<h3 class="text-lg font-semibold text-zinc-100">{provider.name}</h3>
									<p class="text-sm text-zinc-400 capitalize">{provider.type}</p>
									{#if provider.url}
										<p class="text-xs text-zinc-500 mt-1">{provider.url}</p>
									{/if}
									{#if provider.connected}
										<p class="text-sm text-green-500 flex items-center mt-2">
											<svg
												class="w-4 h-4 mr-1"
												fill="none"
												stroke="currentColor"
												viewBox="0 0 24 24"
											>
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M5 13l4 4L19 7"
												/>
											</svg>
											Connected as @{provider.username}
										</p>
									{:else}
										<p class="text-sm text-zinc-500 mt-2">Not connected</p>
									{/if}
									{#if provider.is_default}
										<span
											class="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-green-500/10 text-green-500 mt-2"
										>
											Default
										</span>
									{/if}
								</div>
							</div>

							<div class="flex flex-col space-y-2">
								{#if provider.connected}
									<Button variant="secondary" size="sm" onclick={() => handleDisconnect(provider)}>
										Disconnect
									</Button>
								{:else}
									<Button size="sm" onclick={() => handleConnect(provider)}>Connect</Button>
								{/if}
								<Button variant="ghost" size="sm" onclick={() => openEditModal(provider)}>
									Edit
								</Button>
								<Button variant="ghost" size="sm" onclick={() => handleDelete(provider)}>
									Delete
								</Button>
							</div>
						</div>
					</Card>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Add/Edit Modal -->
{#if showModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
		<div class="bg-zinc-900 rounded-lg border border-zinc-800 max-w-md w-full max-h-[90vh] overflow-y-auto">
			<div class="p-6">
				<h3 class="text-xl font-semibold text-zinc-100 mb-4">
					{editingProvider ? 'Edit Provider' : 'Add Git Provider'}
				</h3>

				<form
					onsubmit={(e) => {
						e.preventDefault();
						handleSubmit();
					}}
					class="space-y-4"
				>
					{#if !editingProvider}
						<Select
							label="Provider Type"
							bind:value={providerType}
							options={providerTypeOptions}
							required
						/>
					{/if}

					<Input label="Name" bind:value={providerName} placeholder="e.g., My Company GitHub" required />

					{#if providerType === 'gitea' || providerType === 'gitlab'}
						<Input
							label="Instance URL"
							bind:value={providerUrl}
							placeholder="https://git.example.com"
							required
						/>
					{/if}

					<Input
						label="Client ID"
						bind:value={clientId}
						placeholder="OAuth Application Client ID"
						required
					/>

					<Input
						label="Client Secret"
						type="password"
						bind:value={clientSecret}
						placeholder="OAuth Application Client Secret"
						required
					/>

					<!-- OAuth Callback URL -->
					<div class="p-3 bg-zinc-800/50 rounded-lg border border-zinc-700">
						<label class="block text-xs font-medium text-zinc-400 mb-2">
							OAuth Callback URL
						</label>
						<div class="flex items-center space-x-2">
							<code class="flex-1 text-xs text-green-400 break-all">
								{callbackUrl}
							</code>
							<button
								type="button"
								onclick={() => {
									navigator.clipboard.writeText(callbackUrl);
									success = 'Callback URL copied to clipboard!';
									setTimeout(() => { success = ''; }, 2000);
								}}
								class="px-2 py-1 text-xs bg-zinc-700 hover:bg-zinc-600 text-zinc-300 rounded transition-colors"
								title="Copy to clipboard"
							>
								Copy
							</button>
						</div>
						<p class="mt-2 text-xs text-zinc-500">
							Use this URL when configuring the OAuth application in your {providerType} settings.
						</p>
					</div>

					<div class="flex items-center">
						<input
							id="is-default"
							type="checkbox"
							bind:checked={isDefault}
							class="h-4 w-4 rounded border-zinc-800 text-green-500 focus:ring-green-500"
						/>
						<label for="is-default" class="ml-2 text-sm text-zinc-300">
							Set as default provider for {providerType}
						</label>
					</div>

					<div class="flex space-x-3 pt-4">
						<Button
							variant="ghost"
							onclick={() => (showModal = false)}
							disabled={modalLoading}
							type="button"
						>
							Cancel
						</Button>
						<Button type="submit" loading={modalLoading} disabled={modalLoading}>
							{modalLoading ? 'Saving...' : editingProvider ? 'Update' : 'Add'}
						</Button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}
