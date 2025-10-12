<script lang="ts">
	import { onMount } from 'svelte';
	import Button from '$lib/components/Button.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import { projectsAPI } from '$lib/api/projects';

	interface WebhookInfo {
		enabled: boolean;
		webhook?: {
			secret: string;
			urls: {
				github: string;
				gitlab: string;
				gitea: string;
			};
			branch: string;
		};
	}

	interface Props {
		projectId: number;
	}

	let { projectId }: Props = $props();

	let webhookInfo = $state<WebhookInfo | null>(null);
	let loading = $state(true);
	let enabling = $state(false);
	let disabling = $state(false);
	let error = $state('');
	let success = $state('');
	let setupModalOpen = $state(false);
	let selectedProvider = $state<'github' | 'gitlab' | 'gitea'>('github');

	// Validate projectId
	$effect(() => {
		if (!projectId || isNaN(projectId)) {
			loading = false;
			error = 'Invalid project ID';
		}
	});

	onMount(async () => {
		if (projectId && !isNaN(projectId)) {
			await loadWebhookInfo();
		}
	});

	async function loadWebhookInfo() {
		if (!projectId || isNaN(projectId)) {
			loading = false;
			return;
		}

		loading = true;
		try {
			webhookInfo = await projectsAPI.getWebhookInfo(projectId);
		} catch (err) {
			console.error('Failed to load webhook info:', err);
			error = 'Failed to load webhook configuration';
		} finally {
			loading = false;
		}
	}

	async function enableWebhook() {
		if (!projectId || isNaN(projectId)) return;

		enabling = true;
		error = '';
		success = '';

		try {
			const data = await projectsAPI.enableWebhook(projectId);

			await loadWebhookInfo();

			// Show appropriate message based on auto-creation status
			if (data.auto_created) {
				success = '✓ Auto-deploy enabled! Webhook automatically configured in your Git provider.';
			} else if (data.manual_setup_required) {
				success = 'Auto-deploy enabled. Opening setup instructions...';
				setupModalOpen = true;
			} else {
				success = 'Webhook enabled successfully!';
				setupModalOpen = true;
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to enable webhook';
		} finally {
			enabling = false;
		}
	}

	async function disableWebhook() {
		if (!projectId || isNaN(projectId)) return;

		if (!confirm('Are you sure you want to disable auto-deploy? You can re-enable it anytime.')) {
			return;
		}

		disabling = true;
		error = '';
		success = '';

		try {
			const data = await projectsAPI.disableWebhook(projectId);

			await loadWebhookInfo();

			// Show appropriate message based on auto-deletion status
			if (data.auto_deleted) {
				success = '✓ Auto-deploy disabled and webhook automatically removed from your Git provider.';
			} else {
				success = 'Auto-deploy disabled. You may want to manually remove the webhook from your Git provider.';
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to disable webhook';
		} finally {
			disabling = false;
		}
	}

	function copyToClipboard(text: string) {
		navigator.clipboard.writeText(text);
		success = 'Copied to clipboard!';
		setTimeout(() => (success = ''), 2000);
	}

	function getProviderInstructions(provider: 'github' | 'gitlab' | 'gitea') {
		switch (provider) {
			case 'github':
				return [
					'Go to your repository settings',
					'Click on "Webhooks" in the left sidebar',
					'Click "Add webhook"',
					'Paste the Payload URL below',
					'Set Content type to "application/json"',
					'Paste the Secret below',
					'Select "Just the push event"',
					'Click "Add webhook"'
				];
			case 'gitlab':
				return [
					'Go to your repository settings',
					'Click on "Webhooks" in the left sidebar',
					'Paste the URL below',
					'Paste the Secret Token below',
					'Check "Push events"',
					'Uncheck all other events',
					'Click "Add webhook"'
				];
			case 'gitea':
				return [
					'Go to your repository settings',
					'Click on "Webhooks"',
					'Click "Add Webhook" → "Gitea"',
					'Paste the Target URL below',
					'Set HTTP Method to "POST"',
					'Set Content Type to "application/json"',
					'Paste the Secret below',
					'Select "Push" event',
					'Click "Add Webhook"'
				];
		}
	}
</script>

<div class="modern-card p-5 hover-lift transition-all">
	<div class="flex items-center justify-between mb-4">
		<div class="flex items-center gap-2">
			<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center">
				<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
				</svg>
			</div>
			<h3 class="text-base font-bold" style="color: rgb(var(--text-primary));">Auto-Deploy</h3>
		</div>
		{#if webhookInfo?.enabled}
			<Badge variant="success">Enabled</Badge>
		{:else}
			<Badge variant="info">Disabled</Badge>
		{/if}
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-4">
			<div class="w-6 h-6 rounded-full border-2 border-primary-800 border-t-transparent animate-spin"></div>
		</div>
	{:else if webhookInfo?.enabled && webhookInfo?.webhook}
		<div class="space-y-3">
			<div class="text-sm" style="color: rgb(var(--text-secondary));">
				<p class="mb-2">Auto-deploy is enabled for <strong style="color: rgb(var(--text-primary));">{webhookInfo.webhook.branch}</strong> branch.</p>
				<p>Pushes to this branch will automatically trigger deployments.</p>
			</div>

			<Button variant="secondary" size="sm" onclick={() => setupModalOpen = true} class="w-full">
				<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
				</svg>
				View Webhook URLs
			</Button>

			<Button variant="ghost" size="sm" onclick={disableWebhook} loading={disabling} disabled={disabling} class="w-full text-red-600 hover:bg-red-50">
				<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
				</svg>
				Disable Auto-Deploy
			</Button>
		</div>
	{:else}
		<div class="space-y-3">
			<p class="text-sm" style="color: rgb(var(--text-secondary));">
				Enable auto-deploy to automatically redeploy your project when you push code to your repository.
			</p>

			<Button variant="primary" size="sm" onclick={enableWebhook} loading={enabling} disabled={enabling} class="w-full glow-green-hover">
				<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
				</svg>
				Enable Auto-Deploy
			</Button>
		</div>
	{/if}

	{#if error}
		<Alert variant="error" dismissible ondismiss={() => error = ''} class="mt-3">
			{error}
		</Alert>
	{/if}

	{#if success}
		<Alert variant="success" dismissible ondismiss={() => success = ''} class="mt-3">
			{success}
		</Alert>
	{/if}
</div>

<!-- Webhook Setup Modal -->
{#if webhookInfo?.enabled && webhookInfo?.webhook}
	<Modal bind:open={setupModalOpen} title="Webhook Configuration" size="lg">
		<div class="space-y-6">
			<!-- Provider Selection -->
			<div>
				<label class="block text-sm font-medium mb-3" style="color: rgb(var(--text-primary));">
					Select your Git provider:
				</label>
				<div class="flex gap-2">
					<button
						onclick={() => selectedProvider = 'github'}
						class="flex-1 py-3 px-4 rounded-lg border-2 transition-all"
						class:border-primary-800={selectedProvider === 'github'}
						class:bg-primary-50={selectedProvider === 'github'}
						style:border-color={selectedProvider === 'github' ? 'rgb(var(--border-brand))' : 'rgb(var(--border-primary))'}
						style:background-color={selectedProvider === 'github' ? 'rgb(var(--bg-secondary))' : 'transparent'}
					>
						<div class="flex items-center justify-center gap-2">
							<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
								<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
							</svg>
							<span class="font-medium text-sm">GitHub</span>
						</div>
					</button>
					<button
						onclick={() => selectedProvider = 'gitlab'}
						class="flex-1 py-3 px-4 rounded-lg border-2 transition-all"
						class:border-primary-800={selectedProvider === 'gitlab'}
						class:bg-primary-50={selectedProvider === 'gitlab'}
						style:border-color={selectedProvider === 'gitlab' ? 'rgb(var(--border-brand))' : 'rgb(var(--border-primary))'}
						style:background-color={selectedProvider === 'gitlab' ? 'rgb(var(--bg-secondary))' : 'transparent'}
					>
						<div class="flex items-center justify-center gap-2">
							<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
								<path d="M23.955 13.587l-1.342-4.135-2.664-8.189a.455.455 0 00-.867 0L16.418 9.45H7.582L4.918 1.263a.455.455 0 00-.867 0L1.387 9.452.045 13.587a.924.924 0 00.331 1.03L12 23.054l11.624-8.437a.924.924 0 00.331-1.03"/>
							</svg>
							<span class="font-medium text-sm">GitLab</span>
						</div>
					</button>
					<button
						onclick={() => selectedProvider = 'gitea'}
						class="flex-1 py-3 px-4 rounded-lg border-2 transition-all"
						class:border-primary-800={selectedProvider === 'gitea'}
						class:bg-primary-50={selectedProvider === 'gitea'}
						style:border-color={selectedProvider === 'gitea' ? 'rgb(var(--border-brand))' : 'rgb(var(--border-primary))'}
						style:background-color={selectedProvider === 'gitea' ? 'rgb(var(--bg-secondary))' : 'transparent'}
					>
						<div class="flex items-center justify-center gap-2">
							<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
								<path d="M4.186 3.815c-.613 0-1.162.24-1.572.651-.41.41-.65.96-.65 1.572v11.924c0 .613.24 1.162.65 1.572.41.41.96.651 1.572.651h15.628c.613 0 1.162-.24 1.572-.651.41-.41.651-.96.651-1.572V6.038c0-.613-.24-1.162-.651-1.572-.41-.41-.96-.651-1.572-.651H4.186zm7.814 2.952c2.297 0 4.162 1.865 4.162 4.162s-1.865 4.162-4.162 4.162-4.162-1.865-4.162-4.162S9.703 6.767 12 6.767z"/>
							</svg>
							<span class="font-medium text-sm">Gitea</span>
						</div>
					</button>
				</div>
			</div>

			<!-- Instructions -->
			<div class="rounded-lg p-4" style="background-color: rgb(var(--bg-secondary)); border: 1px solid rgb(var(--border-primary));">
				<h4 class="font-semibold mb-3" style="color: rgb(var(--text-primary));">
					Setup Instructions for {selectedProvider.charAt(0).toUpperCase() + selectedProvider.slice(1)}:
				</h4>
				<ol class="space-y-2 text-sm" style="color: rgb(var(--text-secondary));">
					{#each getProviderInstructions(selectedProvider) as instruction, index}
						<li class="flex items-start gap-2">
							<span class="flex-shrink-0 w-5 h-5 rounded-full bg-primary-800 text-white text-xs flex items-center justify-center">
								{index + 1}
							</span>
							<span>{instruction}</span>
						</li>
					{/each}
				</ol>
			</div>

			<!-- Webhook URL -->
			<div>
				<label class="block text-sm font-medium mb-2" style="color: rgb(var(--text-primary));">
					{selectedProvider === 'github' ? 'Payload URL' : selectedProvider === 'gitlab' ? 'URL' : 'Target URL'}:
				</label>
				<div class="flex gap-2">
					<input
						type="text"
						value={webhookInfo.webhook.urls[selectedProvider]}
						readonly
						class="modern-input flex-1 font-mono text-sm"
					/>
					<Button
						variant="secondary"
						size="sm"
						onclick={() => copyToClipboard(webhookInfo!.webhook!.urls[selectedProvider])}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
						</svg>
					</Button>
				</div>
			</div>

			<!-- Secret -->
			<div>
				<label class="block text-sm font-medium mb-2" style="color: rgb(var(--text-primary));">
					{selectedProvider === 'gitlab' ? 'Secret Token' : 'Secret'}:
				</label>
				<div class="flex gap-2">
					<input
						type="text"
						value={webhookInfo.webhook.secret}
						readonly
						class="modern-input flex-1 font-mono text-sm"
					/>
					<Button
						variant="secondary"
						size="sm"
						onclick={() => copyToClipboard(webhookInfo!.webhook!.secret)}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
						</svg>
					</Button>
				</div>
			</div>

			<!-- Branch Info -->
			<Alert variant="info">
				Auto-deploy is configured for the <strong>{webhookInfo.webhook.branch}</strong> branch. Pushes to other branches will be ignored.
			</Alert>

			<div class="flex justify-end">
				<Button variant="ghost" onclick={() => setupModalOpen = false}>
					Close
				</Button>
			</div>
		</div>
	</Modal>
{/if}
