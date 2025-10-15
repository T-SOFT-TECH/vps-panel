<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { projectsAPI } from '$lib/api/projects';
	import { deploymentsAPI } from '$lib/api/deployments';
	import Card from '$lib/components/Card.svelte';
	import Button from '$lib/components/Button.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import DomainManager from "$lib/components/DomainManager.svelte";
	import WebhookConfig from '$lib/components/WebhookConfig.svelte';
	import { formatRelativeTime, formatDuration } from '$lib/utils/format';
	import type { Project, Deployment, Environment } from '$lib/types';


	const projectId = Number($page.params.id);

	let project = $state<Project | null>(null);
	let deployments = $state<Deployment[]>([]);
	let environments = $state<Environment[]>([]);
	let loading = $state(true);
	let deploying = $state(false);
	let error = $state('');
	let deleteModalOpen = $state(false);
	let deleting = $state(false);

	// Environment Variables state
	let envModalOpen = $state(false);
	let editingEnv = $state<Environment | null>(null);
	let envForm = $state({ key: '', value: '', is_secret: false });
	let envSaving = $state(false);
	let envDeleting = $state<number | null>(null);

	// PocketBase update state
	let pocketbaseUpdateInfo = $state<{
		current_version: string;
		latest_version: string;
		update_available: boolean;
	} | null>(null);
	let checkingUpdate = $state(false);
	let updating = $state(false);



	onMount(async () => {
		await loadProject();
		await loadDeployments();
		await loadEnvironments();
		// Check for PocketBase updates if project uses PocketBase
		if (project && project.baas_type === 'pocketbase') {
			await checkPocketBaseUpdate();
		}
	});

	async function loadProject() {
		try {
			project = await projectsAPI.getById(projectId);
			// Ensure domains is always an array
			if (project && !project.domains) {
				project.domains = [];
			}
		} catch (err) {
			error = 'Failed to load project';
			console.error(err);
		} finally {
			loading = false;
		}
	}

	async function loadDeployments() {
		try {
			const { deployments: deploymentList } = await deploymentsAPI.getAll(projectId);
			deployments = deploymentList;
		} catch (err) {
			console.error('Failed to load deployments:', err);
		}
	}

	async function handleDeploy() {
		deploying = true;
		error = '';

		try {
			await deploymentsAPI.create(projectId);
			await loadDeployments();
			await loadProject();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create deployment';
		} finally {
			deploying = false;
		}
	}

	async function handleDelete() {
		deleting = true;

		try {
			await projectsAPI.delete(projectId);
			window.location.href = '/projects';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete project';
			deleting = false;
			deleteModalOpen = false;
		}
	}

	function getStatusVariant(status: string): 'success' | 'warning' | 'error' | 'info' {
		switch (status) {
			case 'success':
			case 'active':
				return 'success';
			case 'failed':
				return 'error';
			case 'building':
			case 'deploying':
				return 'warning';
			default:
				return 'info';
		}
	}

	// Environment Variables functions
	async function loadEnvironments() {
		try {
			const { environments: envList } = await projectsAPI.getEnvironments(projectId);
			environments = envList;
		} catch (err) {
			console.error('Failed to load environments:', err);
		}
	}

	function openAddEnvModal() {
		editingEnv = null;
		envForm = { key: '', value: '', is_secret: false };
		envModalOpen = true;
	}

	function openEditEnvModal(env: Environment) {
		editingEnv = env;
		envForm = { key: env.key, value: env.value, is_secret: env.is_secret };
		envModalOpen = true;
	}

	async function handleSaveEnvironment() {
		if (!envForm.key || !envForm.value) {
			error = 'Key and value are required';
			return;
		}

		envSaving = true;
		error = '';

		try {
			if (editingEnv) {
				// Update existing
				await projectsAPI.updateEnvironment(projectId, editingEnv.id, {
					value: envForm.value
				});
			} else {
				// Create new
				await projectsAPI.addEnvironment(projectId, {
					key: envForm.key,
					value: envForm.value,
					is_secret: envForm.is_secret
				});
			}

			await loadEnvironments();
			envModalOpen = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to save environment variable';
		} finally {
			envSaving = false;
		}
	}

	async function handleDeleteEnvironment(envId: number) {
		if (!confirm('Are you sure you want to delete this environment variable?')) {
			return;
		}

		envDeleting = envId;
		error = '';

		try {
			await projectsAPI.deleteEnvironment(projectId, envId);
			await loadEnvironments();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete environment variable';
		} finally {
			envDeleting = null;
		}
	}

	// PocketBase update functions
	async function checkPocketBaseUpdate() {
		if (!project || project.baas_type !== 'pocketbase') {
			return;
		}

		checkingUpdate = true;

		try {
			pocketbaseUpdateInfo = await projectsAPI.checkPocketBaseUpdate(projectId);
		} catch (err) {
			console.error('Failed to check PocketBase update:', err);
			// Silently fail - this is not critical
		} finally {
			checkingUpdate = false;
		}
	}

	async function handlePocketBaseUpdate() {
		if (!confirm('This will redeploy your project with the latest PocketBase version. Continue?')) {
			return;
		}

		updating = true;
		error = '';

		try {
			const result = await projectsAPI.updatePocketBase(projectId);
			// Reload project and deployments to show the new deployment
			await loadProject();
			await loadDeployments();
			await checkPocketBaseUpdate();

			// Show success message
			alert(`PocketBase update initiated! Deployment #${result.deployment_id} created. Updating from ${result.current_version} to ${result.target_version}.`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update PocketBase';
		} finally {
			updating = false;
		}
	}
</script>

<svelte:head>
	<title>{project?.name || 'Project'} - VPS Panel</title>
</svelte:head>

{#if loading}
	<div class="text-center py-16">
		<div class="relative w-16 h-16 mx-auto mb-6">
			<div class="absolute inset-0 rounded-full border-4 border-primary-200"></div>
			<div class="absolute inset-0 rounded-full border-4 border-primary-800 border-t-transparent animate-spin"></div>
			<div class="absolute inset-0 flex items-center justify-center">
				<div class="w-8 h-8 bg-gradient-brand rounded-full pulse"></div>
			</div>
		</div>
		<p class="text-lg font-medium" style="color: rgb(var(--text-secondary));">Loading project...</p>
	</div>
{:else if !project}
	<div class="modern-card p-12">
		<div class="text-center">
			<div class="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-gradient-to-br from-red-500/10 to-red-600/10 mb-6">
				<svg class="w-10 h-10 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
				</svg>
			</div>
			<h3 class="text-2xl font-bold mb-2" style="color: rgb(var(--text-primary));">Project not found</h3>
			<p class="text-base mb-8" style="color: rgb(var(--text-secondary));">The project you're looking for doesn't exist.</p>
			<Button onclick={() => window.location.href = '/projects'} class="btn-primary glow-green-hover">
				<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
				</svg>
				Back to Projects
			</Button>
		</div>
	</div>
{:else}
	<div class="space-y-6 pb-8">
		<!-- Header with Gradient Background -->
		<div class="relative overflow-hidden rounded-2xl slide-in-down">
			<div class="absolute inset-0 mesh-gradient opacity-50"></div>
			<div class="relative glass-pro p-6 border-0">
				<a href="/projects" class="inline-flex items-center text-sm font-medium hover:scale-105 transition-transform mb-4 group" style="color: rgb(var(--text-brand));">
					<svg class="w-4 h-4 mr-2 group-hover:-translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
					</svg>
					Back to Projects
				</a>

				<div class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
					<div class="flex-1">
						<div class="flex items-center gap-3 mb-2">
							<div class="w-14 h-14 rounded-xl bg-gradient-brand flex items-center justify-center shadow-xl glow-green float">
								<svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
								</svg>
							</div>
							<div>
								<h1 class="text-4xl font-bold bg-gradient-to-r from-primary-700 to-primary-900 bg-clip-text text-transparent">
									{project.name}
								</h1>
								<div class="flex items-center gap-2 mt-1">
									<Badge variant={getStatusVariant(project.status)} class="text-xs">
										{project.status}
									</Badge>
									{#if project.framework}
										<Badge variant="info" class="text-xs capitalize">
											{project.framework}
										</Badge>
									{/if}
								</div>
							</div>
						</div>
						{#if project.description}
							<p class="text-base" style="color: rgb(var(--text-secondary));">{project.description}</p>
						{/if}
					</div>

					<div class="flex flex-wrap gap-2">
						<Button onclick={handleDeploy} loading={deploying} disabled={deploying} class="btn-primary glow-green-hover hover:scale-105 transition-transform">
							<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
							</svg>
							Deploy
						</Button>
						<Button variant="secondary" onclick={() => goto(`/projects/${projectId}/edit`)} class="hover:scale-105 transition-transform">
							<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
							</svg>
							Edit
						</Button>
						<Button variant="danger" onclick={() => deleteModalOpen = true} class="hover:scale-105 transition-transform">
							<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
							</svg>
							Delete
						</Button>
					</div>
				</div>
			</div>
		</div>

		{#if error}
			<Alert variant="error" dismissible ondismiss={() => error = ''}>
				{error}
			</Alert>
		{/if}

		<!-- Project Details -->
		<div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
			<div class="lg:col-span-2 space-y-6 slide-in-left">
				<!-- Deployments -->
				<div class="relative overflow-hidden rounded-2xl">
					<div class="absolute inset-0 bg-gradient-to-br from-primary-600/5 to-primary-800/5"></div>
					<div class="relative modern-card p-6 border-0">
						<div class="flex items-center gap-3 mb-6">
							<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center shadow-lg">
								<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
								</svg>
							</div>
							<h2 class="text-xl font-bold" style="color: rgb(var(--text-primary));">Recent Deployments</h2>
						</div>

					{#if deployments.length === 0}
						<div class="text-center py-12">
							<div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-primary-800/10 to-primary-600/10 mb-4">
								<svg class="w-8 h-8 text-primary-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
								</svg>
							</div>
							<h3 class="text-lg font-bold mb-2" style="color: rgb(var(--text-primary));">No deployments yet</h3>
							<p class="text-sm mb-6" style="color: rgb(var(--text-secondary));">Get started by deploying your project.</p>
							<Button onclick={handleDeploy} loading={deploying} class="btn-primary glow-green-hover">
								<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
								</svg>
								Deploy Now
							</Button>
						</div>
					{:else}
						<div class="space-y-3">
							{#each deployments as deployment}
								<a
									href="/projects/{projectId}/deployments/{deployment.id}"
									class="block -mx-6 px-6 py-4 transition-colors last:border-0"
									style="border-bottom: 1px solid rgb(var(--border-primary));"
									onmouseenter={(e) => { e.currentTarget.style.backgroundColor = 'rgb(var(--bg-secondary))'; }}
									onmouseleave={(e) => { e.currentTarget.style.backgroundColor = 'transparent'; }}
								>
									<div class="flex items-start justify-between">
										<div class="flex-1 min-w-0">
											<div class="flex items-center gap-2 mb-1">
												<Badge variant={getStatusVariant(deployment.status)}>
													{deployment.status}
												</Badge>
												<span class="text-xs" style="color: rgb(var(--text-secondary));">
													{formatRelativeTime(deployment.created_at)}
												</span>
											</div>
											<p class="text-sm font-medium truncate" style="color: rgb(var(--text-primary));">
												{deployment.commit_message || 'No commit message'}
											</p>
											<div class="flex items-center gap-4 mt-2 text-xs" style="color: rgb(var(--text-secondary));">
												{#if deployment.commit_hash}
													<span class="font-mono">{deployment.commit_hash.substring(0, 7)}</span>
												{/if}
												{#if deployment.commit_author}
													<span>{deployment.commit_author}</span>
												{/if}
												{#if deployment.duration > 0}
													<span>{formatDuration(deployment.duration)}</span>
												{/if}
											</div>
										</div>
										<svg class="w-5 h-5 ml-4" style="color: rgb(var(--text-tertiary));" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
										</svg>
									</div>
								</a>
							{/each}
						</div>
					{/if}
					</div>
				</div>
			</div>

			<!-- Sidebar -->
			<div class="space-y-6 slide-in-right">
				<!-- Project Info -->
				<div class="modern-card p-5 hover-lift transition-all">
					<div class="flex items-center gap-2 mb-4">
						<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-emerald-500 to-emerald-600 flex items-center justify-center">
							<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
						</div>
						<h3 class="text-base font-bold" style="color: rgb(var(--text-primary));">Project Information</h3>
					</div>
					<dl class="space-y-3 text-sm">
						<div>
							<dt style="color: rgb(var(--text-secondary));">Framework</dt>
							<dd class="font-medium capitalize" style="color: rgb(var(--text-primary));">{project.framework}</dd>
						</div>
						{#if project.baas_type}
							<div>
								<dt style="color: rgb(var(--text-secondary));">Backend</dt>
								<dd class="font-medium capitalize" style="color: rgb(var(--text-primary));">{project.baas_type}</dd>
							</div>
						{/if}
						<div>
							<dt style="color: rgb(var(--text-secondary));">Branch</dt>
							<dd class="font-medium" style="color: rgb(var(--text-primary));">{project.git_branch}</dd>
						</div>
						<div>
							<dt style="color: rgb(var(--text-secondary));">Auto Deploy</dt>
							<dd class="font-medium" style="color: rgb(var(--text-primary));">{project.auto_deploy ? 'Enabled' : 'Disabled'}</dd>
						</div>
						{#if project.last_deployed}
							<div>
								<dt style="color: rgb(var(--text-secondary));">Last Deployed</dt>
								<dd class="font-medium" style="color: rgb(var(--text-primary));">{formatRelativeTime(project.last_deployed)}</dd>
							</div>
						{/if}
					</dl>
				</div>

				<!-- Domain Management -->

				<DomainManager
					{projectId}
					bind:domains={project.domains}
					onUpdate={loadProject}
				/>

				<!-- PocketBase Access URLs -->
				{#if project.baas_type === 'pocketbase' && project.domains && project.domains.length > 0}
					{@const domain = project.domains.find(d => d.is_active)}
					{#if domain}
						<Card>
							<div class="flex items-center gap-2 mb-3">
								<svg class="w-5 h-5 text-primary-800" fill="currentColor" viewBox="0 0 24 24">
									<path d="M20 6h-4V4c0-1.11-.89-2-2-2h-4c-1.11 0-2 .89-2 2v2H4c-1.11 0-1.99.89-1.99 2L2 19c0 1.11.89 2 2 2h16c1.11 0 2-.89 2-2V8c0-1.11-.89-2-2-2zm-6 0h-4V4h4v2z"/>
								</svg>
								<h3 class="text-sm font-semibold" style="color: rgb(var(--text-primary));">PocketBase Access</h3>
							</div>
							<div class="space-y-3">
								<div>
									<dt class="text-xs font-medium mb-1" style="color: rgb(var(--text-tertiary));">🚀 Frontend</dt>
									<dd>
										<a
											href="https://{domain.domain}"
											target="_blank"
											rel="noopener noreferrer"
											class="text-sm break-all"
											style="color: rgb(var(--text-brand));"
										>
											https://{domain.domain}
										</a>
									</dd>
								</div>
								<div>
									<dt class="text-xs font-medium mb-1" style="color: rgb(var(--text-tertiary));">🗄️ API Endpoint</dt>
									<dd>
										<a
											href="https://{domain.domain}/api/"
											target="_blank"
											rel="noopener noreferrer"
											class="text-sm break-all font-mono"
											style="color: rgb(var(--text-brand));"
										>
											https://{domain.domain}/api/*
										</a>
									</dd>
								</div>
								<div>
									<dt class="text-xs font-medium mb-1" style="color: rgb(var(--text-tertiary));">🔧 Admin Dashboard</dt>
									<dd>
										<a
											href="https://{domain.domain}/_/"
											target="_blank"
											rel="noopener noreferrer"
											class="text-sm break-all font-mono"
											style="color: rgb(var(--text-brand));"
										>
											https://{domain.domain}/_
										</a>
									</dd>
								</div>
							</div>
							<div class="mt-4 p-3 rounded-lg" style="background-color: rgb(var(--bg-secondary)); border-left: 3px solid rgb(10, 101, 34);">
								<p class="text-xs" style="color: rgb(var(--text-secondary));">
									<strong style="color: rgb(var(--text-primary));">Official PocketBase binary</strong> downloaded from GitHub releases. Data persists in pb_data directory.
								</p>
							</div>
						</Card>
					{/if}
				{/if}

				<!-- PocketBase Version & Update -->
				{#if project.baas_type === 'pocketbase'}
					<div class="modern-card p-5 hover-lift transition-all">
						<div class="flex items-center gap-2 mb-4">
							<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-indigo-500 to-indigo-600 flex items-center justify-center">
								<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10" />
								</svg>
							</div>
							<h3 class="text-base font-bold" style="color: rgb(var(--text-primary));">PocketBase Version</h3>
						</div>

						{#if checkingUpdate}
							<div class="text-center py-4">
								<div class="inline-block w-6 h-6 border-2 border-primary-800 border-t-transparent rounded-full animate-spin"></div>
								<p class="text-xs mt-2" style="color: rgb(var(--text-secondary));">Checking for updates...</p>
							</div>
						{:else if pocketbaseUpdateInfo}
							<div class="space-y-3">
								<div class="flex items-center justify-between">
									<span class="text-sm" style="color: rgb(var(--text-secondary));">Current Version</span>
									<span class="text-sm font-mono font-semibold" style="color: rgb(var(--text-primary));">
										v{pocketbaseUpdateInfo.current_version}
									</span>
								</div>

								{#if pocketbaseUpdateInfo.update_available}
									<div class="flex items-center justify-between">
										<span class="text-sm" style="color: rgb(var(--text-secondary));">Latest Version</span>
										<div class="flex items-center gap-2">
											<Badge variant="warning" class="text-xs">Update Available</Badge>
											<span class="text-sm font-mono font-semibold" style="color: rgb(var(--text-brand));">
												v{pocketbaseUpdateInfo.latest_version}
											</span>
										</div>
									</div>

									<div class="pt-2">
										<Button
											onclick={handlePocketBaseUpdate}
											loading={updating}
											disabled={updating}
											class="w-full btn-primary glow-green-hover"
											size="sm"
										>
											<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
											</svg>
											{updating ? 'Updating...' : `Update to v${pocketbaseUpdateInfo.latest_version}`}
										</Button>
									</div>

									<div class="mt-3 p-3 rounded-lg" style="background-color: rgb(var(--bg-secondary)); border-left: 3px solid rgb(234, 179, 8);">
										<p class="text-xs" style="color: rgb(var(--text-secondary));">
											A new deployment will be created with the latest PocketBase version. Your data will be preserved.
										</p>
									</div>
								{:else}
									<div class="flex items-center gap-2 p-3 rounded-lg" style="background-color: rgb(var(--bg-secondary)); border-left: 3px solid rgb(10, 101, 34);">
										<svg class="w-4 h-4 text-primary-700 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
										</svg>
										<p class="text-xs" style="color: rgb(var(--text-secondary));">
											You're running the latest version
										</p>
									</div>
								{/if}

								<button
									onclick={checkPocketBaseUpdate}
									disabled={checkingUpdate}
									class="w-full text-xs py-2 rounded-lg transition-colors disabled:opacity-50"
									style="color: rgb(var(--text-brand));"
									onmouseenter={(e) => { if (!e.currentTarget.disabled) e.currentTarget.style.backgroundColor = 'rgb(var(--bg-secondary))'; }}
									onmouseleave={(e) => { e.currentTarget.style.backgroundColor = 'transparent'; }}
								>
									Check for Updates
								</button>
							</div>
						{:else}
							<div class="text-center py-4">
								<p class="text-sm mb-3" style="color: rgb(var(--text-secondary));">Version information unavailable</p>
								<Button variant="ghost" size="sm" onclick={checkPocketBaseUpdate} disabled={checkingUpdate}>
									Check for Updates
								</Button>
							</div>
						{/if}
					</div>
				{/if}

				<!-- Environment Variables -->
				<div class="modern-card p-5 hover-lift transition-all">
					<div class="flex items-center justify-between mb-4">
						<div class="flex items-center gap-2">
							<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-purple-500 to-purple-600 flex items-center justify-center">
								<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4" />
								</svg>
							</div>
							<h3 class="text-base font-bold" style="color: rgb(var(--text-primary));">Environment Variables</h3>
						</div>
						<Button variant="ghost" size="sm" onclick={openAddEnvModal} class="hover:scale-110 transition-transform">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
							</svg>
						</Button>
					</div>

					{#if environments.length === 0}
						<p class="text-sm text-center py-4" style="color: rgb(var(--text-secondary));">
							No environment variables configured
						</p>
					{:else}
						<div class="space-y-2">
							{#each environments as env}
								<div class="rounded-lg p-3" style="background-color: rgb(var(--bg-secondary)); border: 1px solid rgb(var(--border-primary));">
									<div class="flex items-start justify-between gap-2">
										<div class="flex-1 min-w-0">
											<div class="flex items-center gap-2 mb-1">
												<code class="text-xs font-mono text-primary-800">{env.key}</code>
												{#if env.is_secret}
													<Badge variant="warning">Secret</Badge>
												{/if}
											</div>
											<p class="text-xs font-mono break-all" style="color: rgb(var(--text-secondary));">
												{env.is_secret ? '••••••••' : env.value}
											</p>
										</div>
										<div class="flex gap-1 flex-shrink-0">
											<button
												onclick={() => openEditEnvModal(env)}
												class="p-1 rounded transition-colors"
												style="color: rgb(var(--text-tertiary));"
												onmouseenter={(e) => { e.currentTarget.style.backgroundColor = 'rgb(var(--bg-primary))'; e.currentTarget.style.color = 'rgb(var(--text-primary))'; }}
												onmouseleave={(e) => { e.currentTarget.style.backgroundColor = 'transparent'; e.currentTarget.style.color = 'rgb(var(--text-tertiary))'; }}
												title="Edit"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
												</svg>
											</button>
											<button
												onclick={() => handleDeleteEnvironment(env.id)}
												disabled={envDeleting === env.id}
												class="p-1 rounded transition-colors disabled:opacity-50"
												style="color: rgb(var(--text-tertiary));"
												onmouseenter={(e) => { if (!e.currentTarget.disabled) { e.currentTarget.style.backgroundColor = 'rgba(239, 68, 68, 0.1)'; e.currentTarget.style.color = '#ef4444'; } }}
												onmouseleave={(e) => { e.currentTarget.style.backgroundColor = 'transparent'; e.currentTarget.style.color = 'rgb(var(--text-tertiary))'; }}
												title="Delete"
											>
												{#if envDeleting === env.id}
													<div class="w-4 h-4 animate-spin rounded-full border-2 border-red-500 border-t-transparent"></div>
												{:else}
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
													</svg>
												{/if}
											</button>
										</div>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>

				<!-- Webhook Auto-Deploy Configuration -->
				{#if project && projectId && !isNaN(projectId)}
					<WebhookConfig {projectId} />
				{/if}

				<!-- Repository -->
				<div class="modern-card p-5 hover-lift transition-all">
					<div class="flex items-center gap-2 mb-3">
						<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-orange-500 to-orange-600 flex items-center justify-center">
							<svg class="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 24 24">
								<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
							</svg>
						</div>
						<h3 class="text-base font-bold" style="color: rgb(var(--text-primary));">Repository</h3>
					</div>
					<a
						href={project.git_url}
						target="_blank"
						rel="noopener noreferrer"
						class="text-sm break-all font-medium hover:text-primary-700 transition-colors inline-flex items-center gap-1 group"
						style="color: rgb(var(--text-brand));"
					>
						<span>{project.git_url}</span>
						<svg class="w-4 h-4 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
						</svg>
					</a>
				</div>
			</div>
		</div>
	</div>

	<!-- Delete Confirmation Modal -->
	<Modal bind:open={deleteModalOpen} title="Delete Project" size="sm">
		<div class="space-y-4">
			<p class="text-sm" style="color: rgb(var(--text-secondary));">
				Are you sure you want to delete <strong style="color: rgb(var(--text-primary));">{project.name}</strong>? This action cannot be undone.
			</p>

			<Alert variant="warning">
				This will stop the application and remove all deployments.
			</Alert>

			<div class="flex justify-end gap-3">
				<Button variant="ghost" onclick={() => deleteModalOpen = false} disabled={deleting}>
					Cancel
				</Button>
				<Button variant="danger" onclick={handleDelete} loading={deleting} disabled={deleting}>
					{deleting ? 'Deleting...' : 'Delete Project'}
				</Button>
			</div>
		</div>
	</Modal>

	<!-- Environment Variable Add/Edit Modal -->
	<Modal
		bind:open={envModalOpen}
		title={editingEnv ? 'Edit Environment Variable' : 'Add Environment Variable'}
		size="md"
	>
		<form onsubmit={(e) => { e.preventDefault(); handleSaveEnvironment(); }} class="space-y-4">
			<div>
				<label for="env-key" class="block text-sm font-medium mb-2" style="color: rgb(var(--text-primary));">
					Key
				</label>
				<input
					id="env-key"
					type="text"
					bind:value={envForm.key}
					disabled={editingEnv !== null}
					placeholder="e.g., PUBLIC_API_URL, DATABASE_URL"
					class="modern-input w-full disabled:opacity-50 disabled:cursor-not-allowed"
					required
				/>
				<p class="mt-1 text-xs" style="color: rgb(var(--text-secondary));">
					Variable name (cannot be changed after creation)
				</p>
			</div>

			<div>
				<label for="env-value" class="block text-sm font-medium mb-2" style="color: rgb(var(--text-primary));">
					Value
				</label>
				<textarea
					id="env-value"
					bind:value={envForm.value}
					placeholder="Enter the value for this environment variable"
					rows="3"
					class="modern-input w-full resize-none"
					required
				></textarea>
			</div>

			{#if !editingEnv}
				<div class="flex items-center gap-2">
					<input
						id="env-secret"
						type="checkbox"
						bind:checked={envForm.is_secret}
						class="w-4 h-4 rounded text-primary-800 focus:ring-primary-800"
						style="border-color: rgb(var(--border-primary)); background-color: rgb(var(--bg-secondary));"
					/>
					<label for="env-secret" class="text-sm" style="color: rgb(var(--text-primary));">
						Mark as secret (value will be masked in UI)
					</label>
				</div>
			{/if}

			<Alert variant="info">
				Environment variables will be available during build and runtime. Changes require a new deployment to take effect.
			</Alert>

			<div class="flex justify-end gap-3 pt-2">
				<Button
					variant="ghost"
					type="button"
					onclick={() => envModalOpen = false}
					disabled={envSaving}
				>
					Cancel
				</Button>
				<Button
					type="submit"
					loading={envSaving}
					disabled={envSaving}
				>
					{envSaving ? 'Saving...' : (editingEnv ? 'Update' : 'Add')} Variable
				</Button>
			</div>
		</form>
	</Modal>
{/if}
