<script lang="ts">
	import { onMount } from 'svelte';
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
	<div class="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
		<div>
			<h1 class="text-4xl font-bold" style="color: rgb(var(--text-primary));">
				Projects
			</h1>
			<p class="mt-2 text-base" style="color: rgb(var(--text-secondary));">
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
	<div class="max-w-md">
		<div class="relative">
			<svg class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5" style="color: rgb(var(--text-tertiary));" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
		<div class="flex items-center justify-center py-20">
			<div class="text-center">
				<div class="relative w-16 h-16 mx-auto mb-4">
					<div class="absolute inset-0 rounded-full border-4" style="border-color: rgb(var(--border-primary));"></div>
					<div class="absolute inset-0 rounded-full border-4 border-primary-800 border-t-transparent animate-spin"></div>
				</div>
				<p class="font-medium" style="color: rgb(var(--text-secondary));">Loading projects...</p>
			</div>
		</div>
	{:else if filteredProjects.length === 0}
		<div class="modern-card p-12 text-center">
			<div class="inline-flex items-center justify-center w-20 h-20 rounded-2xl mb-6" style="background-color: rgb(var(--bg-secondary)); border: 1px solid rgb(var(--border-primary));">
				<svg class="w-10 h-10" style="color: rgb(var(--text-secondary));" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
				</svg>
			</div>
			<h3 class="text-xl font-bold mb-2" style="color: rgb(var(--text-primary));">
				{searchQuery ? 'No projects found' : 'No projects yet'}
			</h3>
			<p class="text-base mb-8" style="color: rgb(var(--text-secondary));">
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
	{:else}
		<div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
			{#each filteredProjects as project}
				<a href="/projects/{project.id}">
					<div class="modern-card p-6 h-full hover:scale-105 transition-all duration-200 group">
						<!-- Header -->
						<div class="flex items-start justify-between mb-4">
							<div class="flex items-center gap-3 flex-1 min-w-0">
								<div class="w-12 h-12 rounded-xl bg-primary-800 flex items-center justify-center flex-shrink-0 shadow-md">
									<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
									</svg>
								</div>
								<div class="min-w-0 flex-1">
									<h3 class="text-lg font-bold truncate group-hover:text-primary-800 transition-colors" style="color: rgb(var(--text-primary));">
										{project.name}
									</h3>
								</div>
							</div>
							<Badge variant={getStatusVariant(project.status)} class="ml-2 flex-shrink-0">
								{project.status}
							</Badge>
						</div>

						{#if project.description}
							<p class="text-sm mb-4 line-clamp-2 min-h-[2.5rem]" style="color: rgb(var(--text-secondary));">
								{project.description}
							</p>
						{:else}
							<div class="mb-4 min-h-[2.5rem]"></div>
						{/if}

						<!-- Info Grid -->
						<div class="space-y-3 text-sm mb-4">
							<div class="flex items-center" style="color: rgb(var(--text-secondary));">
								<div class="w-8 h-8 rounded-lg bg-blue-500 flex items-center justify-center mr-3 shadow-sm">
									<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
									</svg>
								</div>
								<span class="capitalize">{project.framework}</span>
							</div>

							<div class="flex items-center" style="color: rgb(var(--text-secondary));">
								<div class="w-8 h-8 rounded-lg bg-purple-500 flex items-center justify-center mr-3 shadow-sm">
									<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
									</svg>
								</div>
								<span class="truncate">{project.git_branch}</span>
							</div>

							{#if project.last_deployed}
								<div class="flex items-center" style="color: rgb(var(--text-secondary));">
									<div class="w-8 h-8 rounded-lg bg-emerald-500 flex items-center justify-center mr-3 shadow-sm">
										<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
										</svg>
									</div>
									<span>{formatRelativeTime(project.last_deployed)}</span>
								</div>
							{/if}
						</div>

						<!-- Domain -->
						{#if project.domains && project.domains.length > 0}
							<div class="pt-4 border-t" style="border-color: rgb(var(--border-primary));">
								<div class="flex items-center text-sm">
									<svg class="w-4 h-4 mr-2 text-primary-800" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
									</svg>
									<span class="text-primary-800 truncate font-medium">
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
