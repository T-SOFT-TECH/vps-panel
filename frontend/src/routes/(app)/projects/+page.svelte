<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly, scale } from 'svelte/transition';
	import { projectsAPI } from '$lib/api/projects';
	import Card from '$lib/components/Card.svelte';
	import Button from '$lib/components/Button.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import { formatRelativeTime } from '$lib/utils/format';
	import type { Project } from '$lib/types';

	let projects = $state<Project[]>([]);
	let loading = $state(true);
	let searchQuery = $state('');

	onMount(async () => {
		await loadProjects();
	});

	async function loadProjects() {
		try {
			const { projects: projectList } = await projectsAPI.getAll();
			projects = projectList;
		} catch (error) {
			console.error('Failed to load projects:', error);
		} finally {
			loading = false;
		}
	}

	function getStatusVariant(status: string): 'success' | 'warning' | 'error' | 'info' {
		switch (status) {
			case 'active':
				return 'success';
			case 'failed':
				return 'error';
			case 'deploying':
				return 'warning';
			default:
				return 'info';
		}
	}

	const filteredProjects = $derived(
		projects.filter(project =>
			project.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
			project.description?.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);
</script>

<svelte:head>
	<title>Projects - VPS Panel</title>
</svelte:head>

<div class="space-y-8 pb-8">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4" in:fly={{ y: -20, duration: 500, delay: 0 }}>
		<div>
			<h1 class="text-4xl font-bold bg-gradient-to-r from-primary-400 via-primary-500 to-emerald-400 bg-clip-text text-transparent">
				Projects
			</h1>
			<p class="mt-2 text-base text-zinc-400 dark:text-zinc-400 light:text-zinc-600">
				Manage your deployed applications
			</p>
		</div>
		<Button onclick={() => window.location.href = '/projects/new'} class="btn-primary">
			<svg class="w-5 h-5 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
			</svg>
			New Project
		</Button>
	</div>

	<!-- Search -->
	<div class="max-w-md" in:fly={{ y: 20, duration: 500, delay: 50 }}>
		<div class="relative">
			<svg class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-zinc-500 dark:text-zinc-500 light:text-zinc-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
			</svg>
			<input
				type="text"
				placeholder="Search projects..."
				bind:value={searchQuery}
				class="modern-input w-full pl-12"
			/>
		</div>
	</div>

	<!-- Projects List -->
	{#if loading}
		<div class="flex items-center justify-center py-20" in:scale={{ duration: 300 }}>
			<div class="text-center">
				<div class="relative w-16 h-16 mx-auto mb-4">
					<div class="absolute inset-0 rounded-full border-4 border-zinc-800 dark:border-zinc-800 light:border-zinc-200"></div>
					<div class="absolute inset-0 rounded-full border-4 border-primary-500 border-t-transparent animate-spin"></div>
				</div>
				<p class="text-zinc-400 dark:text-zinc-400 light:text-zinc-600 font-medium">Loading projects...</p>
			</div>
		</div>
	{:else if filteredProjects.length === 0}
		<div in:fly={{ y: 20, duration: 500, delay: 100 }}>
			<div class="modern-card p-12 text-center">
				<div class="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-zinc-800/50 dark:bg-zinc-800/50 light:bg-zinc-100 border border-zinc-700/50 dark:border-zinc-700/50 light:border-zinc-200 mb-6">
					<svg class="w-10 h-10 text-zinc-500 dark:text-zinc-500 light:text-zinc-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
					</svg>
				</div>
				<h3 class="text-xl font-bold text-zinc-100 dark:text-zinc-100 light:text-zinc-900 mb-2">
					{searchQuery ? 'No projects found' : 'No projects yet'}
				</h3>
				<p class="text-base text-zinc-400 dark:text-zinc-400 light:text-zinc-600 mb-8">
					{searchQuery ? 'Try adjusting your search criteria' : 'Get started by creating your first project'}
				</p>
				{#if !searchQuery}
					<Button onclick={() => window.location.href = '/projects/new'} class="btn-primary">
						<svg class="w-5 h-5 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
						</svg>
						Create Your First Project
					</Button>
				{/if}
			</div>
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
			{#each filteredProjects as project, i}
				<a href="/projects/{project.id}" in:fly={{ y: 20, duration: 500, delay: 100 + (i * 50) }}>
					<div class="modern-card p-6 h-full hover:scale-105 transition-all duration-200 group">
						<!-- Header -->
						<div class="flex items-start justify-between mb-4">
							<div class="flex items-center gap-3 flex-1 min-w-0">
								<div class="w-12 h-12 rounded-xl bg-gradient-to-br from-primary-500/20 to-primary-600/20 border border-primary-500/30 flex items-center justify-center flex-shrink-0">
									<svg class="w-6 h-6 text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
									</svg>
								</div>
								<div class="min-w-0 flex-1">
									<h3 class="text-lg font-bold text-zinc-100 dark:text-zinc-100 light:text-zinc-900 truncate group-hover:text-primary-400 transition-colors">
										{project.name}
									</h3>
								</div>
							</div>
							<Badge variant={getStatusVariant(project.status)} class="ml-2 flex-shrink-0">
								{project.status}
							</Badge>
						</div>

						{#if project.description}
							<p class="text-sm text-zinc-400 dark:text-zinc-400 light:text-zinc-600 mb-4 line-clamp-2 min-h-[2.5rem]">
								{project.description}
							</p>
						{:else}
							<div class="mb-4 min-h-[2.5rem]"></div>
						{/if}

						<!-- Info Grid -->
						<div class="space-y-3 text-sm mb-4">
							<div class="flex items-center text-zinc-400 dark:text-zinc-400 light:text-zinc-600">
								<div class="w-8 h-8 rounded-lg bg-blue-500/10 border border-blue-500/20 flex items-center justify-center mr-3">
									<svg class="w-4 h-4 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
									</svg>
								</div>
								<span class="capitalize">{project.framework}</span>
							</div>

							<div class="flex items-center text-zinc-400 dark:text-zinc-400 light:text-zinc-600">
								<div class="w-8 h-8 rounded-lg bg-purple-500/10 border border-purple-500/20 flex items-center justify-center mr-3">
									<svg class="w-4 h-4 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
									</svg>
								</div>
								<span class="truncate">{project.git_branch}</span>
							</div>

							{#if project.last_deployed}
								<div class="flex items-center text-zinc-400 dark:text-zinc-400 light:text-zinc-600">
									<div class="w-8 h-8 rounded-lg bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center mr-3">
										<svg class="w-4 h-4 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
										</svg>
									</div>
									<span>{formatRelativeTime(project.last_deployed)}</span>
								</div>
							{/if}
						</div>

						<!-- Domain -->
						{#if project.domains && project.domains.length > 0}
							<div class="pt-4 border-t border-zinc-700/50 dark:border-zinc-700/50 light:border-zinc-200">
								<div class="flex items-center text-sm">
									<svg class="w-4 h-4 mr-2 text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
									</svg>
									<span class="text-primary-400 dark:text-primary-400 light:text-primary-600 truncate font-medium">
										{project.domains[0].domain}
									</span>
								</div>
							</div>
						{/if}
					</div>
				</a>
			{/each}
		</div>
	{/if}
</div>
