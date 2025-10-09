<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';
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

<div class="space-y-6">
	<!-- Header -->
	<div class="flex justify-between items-center" in:fly={{ y: -20, duration: 400, delay: 0 }}>
		<div>
			<h1 class="text-3xl font-bold text-zinc-100">Projects</h1>
			<p class="mt-1 text-sm text-zinc-400">Manage your deployed applications</p>
		</div>
		<Button onclick={() => window.location.href = '/projects/new'}>
			<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
			</svg>
			New Project
		</Button>
	</div>

	<!-- Search -->
	<div class="max-w-md" in:fly={{ y: 20, duration: 400, delay: 50 }}>
		<input
			type="text"
			placeholder="Search projects..."
			bind:value={searchQuery}
			class="block w-full rounded-lg border border-zinc-800 bg-zinc-900 text-zinc-100 placeholder:text-zinc-500 px-4 py-2 focus:outline-none focus:ring-2 focus:ring-green-500"
		/>
	</div>

	<!-- Projects List -->
	{#if loading}
		<div class="text-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-green-500 mx-auto"></div>
			<p class="mt-4 text-zinc-400">Loading projects...</p>
		</div>
	{:else if filteredProjects.length === 0}
		<div in:fly={{ y: 20, duration: 400, delay: 100 }}>
			<Card>
				<div class="text-center py-12">
				<svg class="mx-auto h-12 w-12 text-zinc-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
				</svg>
				<h3 class="mt-2 text-sm font-medium text-zinc-100">
					{searchQuery ? 'No projects found' : 'No projects'}
				</h3>
				<p class="mt-1 text-sm text-zinc-400">
					{searchQuery ? 'Try adjusting your search' : 'Get started by creating a new project'}
				</p>
				{#if !searchQuery}
					<div class="mt-6">
						<Button onclick={() => window.location.href = '/projects/new'}>
							<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
							</svg>
							New Project
						</Button>
					</div>
				{/if}
				</div>
			</Card>
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
			{#each filteredProjects as project, i}
				<div in:fly={{ y: 20, duration: 400, delay: 100 + (i * 50) }}>
					<Card hover class="cursor-pointer" padding={false}>
						<a href="/projects/{project.id}" class="block p-6">
						<div class="flex items-center justify-between mb-4">
							<h3 class="text-lg font-semibold text-zinc-100 truncate">
								{project.name}
							</h3>
							<Badge variant={getStatusVariant(project.status)}>
								{project.status}
							</Badge>
						</div>

						{#if project.description}
							<p class="text-sm text-zinc-400 mb-4 line-clamp-2">
								{project.description}
							</p>
						{/if}

						<div class="space-y-2 text-sm">
							<div class="flex items-center text-zinc-400">
								<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
								</svg>
								<span class="capitalize">{project.framework}</span>
							</div>

							<div class="flex items-center text-zinc-400">
								<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
								</svg>
								<span>{project.git_branch}</span>
							</div>

							{#if project.last_deployed}
								<div class="flex items-center text-zinc-400">
									<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
									<span>Deployed {formatRelativeTime(project.last_deployed)}</span>
								</div>
							{/if}
						</div>

						{#if project.domains && project.domains.length > 0}
							<div class="mt-4 pt-4 border-t border-zinc-800">
								<div class="flex items-center text-sm text-green-500">
									<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
									</svg>
									<span class="truncate">{project.domains[0].domain}</span>
								</div>
							</div>
						{/if}
					</a>
					</Card>
				</div>
			{/each}
		</div>
	{/if}
</div>
