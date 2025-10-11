<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { deploymentsAPI } from '$lib/api/deployments';
	import Card from '$lib/components/Card.svelte';
	import Button from '$lib/components/Button.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import { formatDate, formatDuration } from '$lib/utils/format';
	import type { Deployment, BuildLog } from '$lib/types';

	const projectId = Number($page.params.id);
	const deploymentId = Number($page.params.deploymentId);

	let deployment = $state<Deployment | null>(null);
	let logs = $state<BuildLog[]>([]);
	let loading = $state(true);
	let autoRefresh: ReturnType<typeof setInterval> | null = null;

	onMount(async () => {
		await loadDeployment();
		await loadLogs();

		// Auto-refresh for active deployments
		autoRefresh = setInterval(async () => {
			if (deployment && ['pending', 'building', 'deploying'].includes(deployment.status)) {
				await loadDeployment();
				await loadLogs();
			}
		}, 5000);
	});

	onDestroy(() => {
		if (autoRefresh) {
			clearInterval(autoRefresh);
		}
	});

	async function loadDeployment() {
		try {
			deployment = await deploymentsAPI.getById(projectId, deploymentId);
		} catch (err) {
			console.error('Failed to load deployment:', err);
		} finally {
			loading = false;
		}
	}

	async function loadLogs() {
		try {
			const { logs: logList } = await deploymentsAPI.getLogs(projectId, deploymentId);
			logs = logList;
		} catch (err) {
			console.error('Failed to load logs:', err);
		}
	}

	async function handleCancel() {
		if (!deployment) return;

		try {
			await deploymentsAPI.cancel(projectId, deploymentId);
			await loadDeployment();
		} catch (err) {
			console.error('Failed to cancel deployment:', err);
		}
	}

	function getStatusVariant(status: string): 'success' | 'warning' | 'error' | 'info' {
		switch (status) {
			case 'success':
				return 'success';
			case 'failed':
			case 'cancelled':
				return 'error';
			case 'building':
			case 'deploying':
				return 'warning';
			default:
				return 'info';
		}
	}

	function getLogColor(logType: string): string {
		switch (logType) {
			case 'error':
				return 'text-red-600';
			case 'warning':
				return 'text-yellow-600';
			default:
				return 'text-zinc-300';
		}
	}
</script>

<svelte:head>
	<title>Deployment #{deploymentId} - VPS Panel</title>
</svelte:head>

{#if loading}
	<div class="text-center py-12">
		<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-800 mx-auto"></div>
		<p class="mt-4" style="color: rgb(var(--text-secondary));">Loading deployment...</p>
	</div>
{:else if !deployment}
	<Card>
		<div class="text-center py-12">
			<h3 class="text-lg font-medium" style="color: rgb(var(--text-primary));">Deployment not found</h3>
			<p class="mt-1 text-sm" style="color: rgb(var(--text-secondary));">The deployment you're looking for doesn't exist.</p>
			<div class="mt-6">
				<Button onclick={() => window.history.back()}>
					Go Back
				</Button>
			</div>
		</div>
	</Card>
{:else}
	<div class="space-y-6">
		<!-- Header -->
		<div>
			<a href="/projects/{projectId}" class="flex items-center mb-4 text-sm" style="color: rgb(var(--text-brand));">
				<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
				</svg>
				Back to Project
			</a>

			<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
				<div>
					<div class="flex items-center gap-3">
						<h1 class="text-3xl font-bold" style="color: rgb(var(--text-primary));">Deployment #{deployment.id}</h1>
						<Badge variant={getStatusVariant(deployment.status)}>
							{deployment.status}
						</Badge>
					</div>
					{#if deployment.commit_message}
						<p class="mt-1" style="color: rgb(var(--text-secondary));">{deployment.commit_message}</p>
					{/if}
				</div>

				{#if deployment.status === 'building' || deployment.status === 'deploying'}
					<Button variant="danger" onclick={handleCancel}>
						Cancel Deployment
					</Button>
				{/if}
			</div>
		</div>

		<!-- Deployment Info -->
		<div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
			<Card>
				<h3 class="text-sm font-semibold mb-3" style="color: rgb(var(--text-primary));">Commit Information</h3>
				<dl class="space-y-3 text-sm">
					{#if deployment.commit_hash}
						<div>
							<dt style="color: rgb(var(--text-secondary));">Commit Hash</dt>
							<dd class="font-mono" style="color: rgb(var(--text-primary));">{deployment.commit_hash.substring(0, 7)}</dd>
						</div>
					{/if}
					{#if deployment.commit_author}
						<div>
							<dt style="color: rgb(var(--text-secondary));">Author</dt>
							<dd style="color: rgb(var(--text-primary));">{deployment.commit_author}</dd>
						</div>
					{/if}
					<div>
						<dt style="color: rgb(var(--text-secondary));">Branch</dt>
						<dd style="color: rgb(var(--text-primary));">{deployment.branch}</dd>
					</div>
				</dl>
			</Card>

			<Card>
				<h3 class="text-sm font-semibold mb-3" style="color: rgb(var(--text-primary));">Deployment Details</h3>
				<dl class="space-y-3 text-sm">
					<div>
						<dt style="color: rgb(var(--text-secondary));">Triggered By</dt>
						<dd class="capitalize" style="color: rgb(var(--text-primary));">{deployment.triggered_by}</dd>
					</div>
					{#if deployment.started_at}
						<div>
							<dt style="color: rgb(var(--text-secondary));">Started</dt>
							<dd style="color: rgb(var(--text-primary));">{formatDate(deployment.started_at)}</dd>
						</div>
					{/if}
					{#if deployment.completed_at}
						<div>
							<dt style="color: rgb(var(--text-secondary));">Completed</dt>
							<dd style="color: rgb(var(--text-primary));">{formatDate(deployment.completed_at)}</dd>
						</div>
					{/if}
				</dl>
			</Card>

			<Card>
				<h3 class="text-sm font-semibold mb-3" style="color: rgb(var(--text-primary));">Performance</h3>
				<dl class="space-y-3 text-sm">
					{#if deployment.duration > 0}
						<div>
							<dt style="color: rgb(var(--text-secondary));">Duration</dt>
							<dd style="color: rgb(var(--text-primary));">{formatDuration(deployment.duration)}</dd>
						</div>
					{/if}
					<div>
						<dt style="color: rgb(var(--text-secondary));">Status</dt>
						<dd>
							<Badge variant={getStatusVariant(deployment.status)}>
								{deployment.status}
							</Badge>
						</dd>
					</div>
				</dl>
			</Card>
		</div>

		<!-- Build Logs -->
		<Card>
			<div class="flex items-center justify-between mb-4">
				<h2 class="text-lg font-semibold" style="color: rgb(var(--text-primary));">Build Logs</h2>
				{#if ['building', 'deploying'].includes(deployment.status)}
					<div class="flex items-center text-sm" style="color: rgb(var(--text-secondary));">
						<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-primary-800 mr-2"></div>
						Live
					</div>
				{/if}
			</div>

			{#if logs.length === 0}
				<div class="text-center py-8" style="color: rgb(var(--text-secondary));">
					No logs available yet
				</div>
			{:else}
				<div class="rounded-lg p-4 overflow-x-auto" style="background-color: rgb(var(--bg-secondary));">
					<div class="font-mono text-sm space-y-1">
						{#each logs as log}
							<div class={getLogColor(log.log_type)}>
								<span style="color: rgb(var(--text-secondary));">[{formatDate(log.created_at)}]</span>
								<span style="color: rgb(var(--text-primary));">{log.log}</span>
							</div>
						{/each}
					</div>
				</div>
			{/if}

			{#if deployment.error_message}
				<div class="mt-4 p-4 bg-red-50 rounded-lg" style="border: 1px solid #fecaca;">
					<h3 class="text-sm font-semibold text-red-800 mb-2">Error</h3>
					<p class="text-sm text-red-700 font-mono">{deployment.error_message}</p>
				</div>
			{/if}
		</Card>
	</div>
{/if}
