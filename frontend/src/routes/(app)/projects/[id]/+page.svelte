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

	onMount(async () => {
		await loadProject();
		await loadDeployments();
		await loadEnvironments();
	});

	async function loadProject() {
		try {
			project = await projectsAPI.getById(projectId);
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
</script>

<svelte:head>
	<title>{project?.name || 'Project'} - VPS Panel</title>
</svelte:head>

{#if loading}
	<div class="text-center py-12">
		<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-green-500 mx-auto"></div>
		<p class="mt-4 text-zinc-400">Loading project...</p>
	</div>
{:else if !project}
	<Card>
		<div class="text-center py-12">
			<h3 class="text-lg font-medium text-zinc-100">Project not found</h3>
			<p class="mt-1 text-sm text-zinc-400">The project you're looking for doesn't exist.</p>
			<div class="mt-6">
				<Button onclick={() => window.location.href = '/projects'}>
					Back to Projects
				</Button>
			</div>
		</div>
	</Card>
{:else}
	<div class="space-y-6">
		<!-- Header -->
		<div>
			<a href="/projects" class="text-sm text-green-500 hover:text-green-400 flex items-center mb-4">
				<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
				</svg>
				Back to Projects
			</a>

			<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
				<div>
					<div class="flex items-center gap-3">
						<h1 class="text-3xl font-bold text-zinc-100">{project.name}</h1>
						<Badge variant={getStatusVariant(project.status)}>
							{project.status}
						</Badge>
					</div>
					{#if project.description}
						<p class="mt-1 text-zinc-400">{project.description}</p>
					{/if}
				</div>

				<div class="flex gap-2">
					<Button onclick={handleDeploy} loading={deploying} disabled={deploying}>
						<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
						</svg>
						Deploy
					</Button>
					<Button variant="secondary" onclick={() => goto(`/projects/${projectId}/edit`)}>
						<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
						</svg>
						Edit
					</Button>
					<Button variant="danger" onclick={() => deleteModalOpen = true}>
						Delete
					</Button>
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
			<div class="lg:col-span-2 space-y-6">
				<!-- Deployments -->
				<Card>
					<h2 class="text-lg font-semibold text-zinc-100 mb-4">Recent Deployments</h2>

					{#if deployments.length === 0}
						<div class="text-center py-8">
							<svg class="mx-auto h-12 w-12 text-zinc-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
							</svg>
							<h3 class="mt-2 text-sm font-medium text-zinc-100">No deployments yet</h3>
							<p class="mt-1 text-sm text-zinc-400">Get started by deploying your project.</p>
							<div class="mt-6">
								<Button onclick={handleDeploy} loading={deploying}>
									Deploy Now
								</Button>
							</div>
						</div>
					{:else}
						<div class="space-y-3">
							{#each deployments as deployment}
								<a
									href="/projects/{projectId}/deployments/{deployment.id}"
									class="block hover:bg-zinc-800 -mx-6 px-6 py-4 transition-colors border-b border-zinc-800 last:border-0"
								>
									<div class="flex items-start justify-between">
										<div class="flex-1 min-w-0">
											<div class="flex items-center gap-2 mb-1">
												<Badge variant={getStatusVariant(deployment.status)}>
													{deployment.status}
												</Badge>
												<span class="text-xs text-zinc-400">
													{formatRelativeTime(deployment.created_at)}
												</span>
											</div>
											<p class="text-sm font-medium text-zinc-100 truncate">
												{deployment.commit_message || 'No commit message'}
											</p>
											<div class="flex items-center gap-4 mt-2 text-xs text-zinc-400">
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
										<svg class="w-5 h-5 text-zinc-500 ml-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
										</svg>
									</div>
								</a>
							{/each}
						</div>
					{/if}
				</Card>
			</div>

			<!-- Sidebar -->
			<div class="space-y-6">
				<!-- Project Info -->
				<Card>
					<h3 class="text-sm font-semibold text-zinc-100 mb-3">Project Information</h3>
					<dl class="space-y-3 text-sm">
						<div>
							<dt class="text-zinc-400">Framework</dt>
							<dd class="font-medium text-zinc-100 capitalize">{project.framework}</dd>
						</div>
						{#if project.baas_type}
							<div>
								<dt class="text-zinc-400">Backend</dt>
								<dd class="font-medium text-zinc-100 capitalize">{project.baas_type}</dd>
							</div>
						{/if}
						<div>
							<dt class="text-zinc-400">Branch</dt>
							<dd class="font-medium text-zinc-100">{project.git_branch}</dd>
						</div>
						<div>
							<dt class="text-zinc-400">Auto Deploy</dt>
							<dd class="font-medium text-zinc-100">{project.auto_deploy ? 'Enabled' : 'Disabled'}</dd>
						</div>
						{#if project.last_deployed}
							<div>
								<dt class="text-zinc-400">Last Deployed</dt>
								<dd class="font-medium text-zinc-100">{formatRelativeTime(project.last_deployed)}</dd>
							</div>
						{/if}
					</dl>
				</Card>

				<!-- Domains -->
				{#if project.domains && project.domains.length > 0}
					<Card>
						<h3 class="text-sm font-semibold text-zinc-100 mb-3">Domains</h3>
						<div class="space-y-2">
							{#each project.domains as domain}
								<div class="flex items-center justify-between text-sm">
									<a
										href="https://{domain.domain}"
										target="_blank"
										rel="noopener noreferrer"
										class="text-green-500 hover:text-green-400 truncate"
									>
										{domain.domain}
									</a>
									{#if domain.ssl_enabled}
										<svg class="w-4 h-4 text-green-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
										</svg>
									{/if}
								</div>
							{/each}
						</div>
					</Card>
				{/if}

				<!-- Environment Variables -->
				<Card>
					<div class="flex items-center justify-between mb-3">
						<h3 class="text-sm font-semibold text-zinc-100">Environment Variables</h3>
						<Button variant="ghost" size="sm" onclick={openAddEnvModal}>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
							</svg>
						</Button>
					</div>

					{#if environments.length === 0}
						<p class="text-sm text-zinc-400 text-center py-4">
							No environment variables configured
						</p>
					{:else}
						<div class="space-y-2">
							{#each environments as env}
								<div class="bg-zinc-800 rounded-lg p-3 border border-zinc-700">
									<div class="flex items-start justify-between gap-2">
										<div class="flex-1 min-w-0">
											<div class="flex items-center gap-2 mb-1">
												<code class="text-xs font-mono text-green-500">{env.key}</code>
												{#if env.is_secret}
													<Badge variant="warning">Secret</Badge>
												{/if}
											</div>
											<p class="text-xs text-zinc-400 font-mono break-all">
												{env.is_secret ? '••••••••' : env.value}
											</p>
										</div>
										<div class="flex gap-1 flex-shrink-0">
											<button
												onclick={() => openEditEnvModal(env)}
												class="p-1 hover:bg-zinc-700 rounded text-zinc-400 hover:text-zinc-100 transition-colors"
												title="Edit"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
												</svg>
											</button>
											<button
												onclick={() => handleDeleteEnvironment(env.id)}
												disabled={envDeleting === env.id}
												class="p-1 hover:bg-red-500/10 rounded text-zinc-400 hover:text-red-500 transition-colors disabled:opacity-50"
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
				</Card>

				<!-- Repository -->
				<Card>
					<h3 class="text-sm font-semibold text-zinc-100 mb-3">Repository</h3>
					<a
						href={project.git_url}
						target="_blank"
						rel="noopener noreferrer"
						class="text-sm text-green-500 hover:text-green-400 break-all"
					>
						{project.git_url}
					</a>
				</Card>
			</div>
		</div>
	</div>

	<!-- Delete Confirmation Modal -->
	<Modal bind:open={deleteModalOpen} title="Delete Project" size="sm">
		<div class="space-y-4">
			<p class="text-sm text-zinc-400">
				Are you sure you want to delete <strong>{project.name}</strong>? This action cannot be undone.
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
				<label for="env-key" class="block text-sm font-medium text-zinc-300 mb-2">
					Key
				</label>
				<input
					id="env-key"
					type="text"
					bind:value={envForm.key}
					disabled={editingEnv !== null}
					placeholder="e.g., PUBLIC_API_URL, DATABASE_URL"
					class="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-zinc-100 placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-green-500 focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed"
					required
				/>
				<p class="mt-1 text-xs text-zinc-400">
					Variable name (cannot be changed after creation)
				</p>
			</div>

			<div>
				<label for="env-value" class="block text-sm font-medium text-zinc-300 mb-2">
					Value
				</label>
				<textarea
					id="env-value"
					bind:value={envForm.value}
					placeholder="Enter the value for this environment variable"
					rows="3"
					class="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-zinc-100 placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-green-500 focus:border-transparent resize-none"
					required
				></textarea>
			</div>

			{#if !editingEnv}
				<div class="flex items-center gap-2">
					<input
						id="env-secret"
						type="checkbox"
						bind:checked={envForm.is_secret}
						class="w-4 h-4 bg-zinc-800 border-zinc-700 rounded text-green-500 focus:ring-green-500 focus:ring-offset-zinc-900"
					/>
					<label for="env-secret" class="text-sm text-zinc-300">
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
