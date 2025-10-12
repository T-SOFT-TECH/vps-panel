<script lang="ts">
	import { onMount } from 'svelte';
	import { projectsAPI } from '$lib/api/projects';
	import Card from '$lib/components/Card.svelte';
	import Button from '$lib/components/Button.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
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
	<!-- Header with Gradient Background -->
	<div class="relative overflow-hidden rounded-2xl">
		<div class="absolute inset-0 mesh-gradient opacity-50"></div>
		<div class="relative glass-pro p-8 border-0">
			<div class="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
				<div class="slide-in-left">
					<h1 class="text-5xl font-bold bg-gradient-to-r from-primary-700 to-primary-900 bg-clip-text text-transparent mb-2">
						Projects
					</h1>
					<p class="text-lg" style="color: rgb(var(--text-secondary));">
						Manage and deploy your applications with ease
					</p>
				</div>
				<div class="slide-in-right">
					<Button onclick={() => window.location.href = '/projects/new'} class="btn-primary glow-green-hover px-6 py-4 text-base hover:scale-105 transition-transform">
						<svg class="w-5 h-5 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
						</svg>
						New Project
					</Button>
				</div>
			</div>
		</div>
	</div>

	<!-- Search & Stats -->
	<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
		<!-- Search -->
		<div class="md:col-span-2 slide-in-left">
			<div class="relative group">
				<div class="absolute inset-0 bg-gradient-brand rounded-xl opacity-0 group-focus-within:opacity-20 blur-xl transition-opacity"></div>
				<div class="relative">
					<svg class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 transition-colors group-focus-within:text-primary-600" style="color: rgb(var(--text-tertiary));" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
					<input
						type="text"
						placeholder="Search projects by name or description..."
						bind:value={searchQuery}
						class="modern-input w-full pl-12 hover-lift"
					/>
				</div>
			</div>
		</div>

		<!-- Quick Stat -->
		<div class="slide-in-right">
			<div class="modern-card p-4 h-full flex items-center justify-between card-highlight">
				<div>
					<p class="text-sm font-medium mb-1" style="color: rgb(var(--text-secondary));">Total Projects</p>
					<p class="text-3xl font-bold" style="color: rgb(var(--text-primary));">{projects.length}</p>
				</div>
				<div class="w-12 h-12 rounded-xl bg-gradient-brand flex items-center justify-center shadow-lg glow-green">
					<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
					</svg>
				</div>
			</div>
		</div>
	</div>

	<!-- Projects Grid -->
	{#if loading}
		<LoadingSpinner fullscreen={false} text="Loading your projects..." size="lg" />
	{:else if filteredProjects.length === 0}
		<div class="modern-card p-16 text-center fade-in">
			<div class="inline-flex items-center justify-center w-24 h-24 rounded-2xl bg-gradient-to-br from-primary-800/10 to-primary-600/10 mb-6">
				<svg class="w-12 h-12 text-primary-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
				</svg>
			</div>
			<h3 class="text-2xl font-bold mb-3" style="color: rgb(var(--text-primary));">
				{searchQuery ? 'No projects found' : 'No projects yet'}
			</h3>
			<p class="text-base mb-8 max-w-md mx-auto" style="color: rgb(var(--text-secondary));">
				{searchQuery ? 'Try adjusting your search criteria or create a new project' : 'Get started by deploying your first application in just a few clicks'}
			</p>
			{#if !searchQuery}
				<Button onclick={() => window.location.href = '/projects/new'} class="btn-primary glow-green-hover">
					<svg class="w-5 h-5 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					Create Your First Project
				</Button>
			{/if}
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
			{#each filteredProjects as project, i}
				<a href="/projects/{project.id}" class="group fade-in" style="animation-delay: {i * 0.1}s;">
					<div class="relative h-full">
						<!-- Gradient glow on hover -->
						<div class="absolute inset-0 bg-gradient-to-br from-primary-600/20 to-primary-800/20 rounded-2xl blur-xl opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>

						<div class="relative modern-card p-6 h-full hover-lift card-highlight">
							<!-- Header -->
							<div class="flex items-start justify-between mb-4">
								<div class="flex items-center gap-3 flex-1 min-w-0">
									<div class="w-14 h-14 rounded-xl bg-gradient-brand flex items-center justify-center flex-shrink-0 shadow-lg group-hover:scale-110 transition-transform glow-green">
										<svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
										</svg>
									</div>
									<div class="min-w-0 flex-1">
										<h3 class="text-lg font-bold truncate group-hover:text-primary-700 transition-colors" style="color: rgb(var(--text-primary));">
											{project.name}
										</h3>
										<Badge variant={getStatusVariant(project.status)} class="mt-1">
											{project.status}
										</Badge>
									</div>
								</div>
							</div>

							{#if project.description}
								<p class="text-sm mb-4 line-clamp-2 min-h-[2.5rem]" style="color: rgb(var(--text-secondary));">
									{project.description}
								</p>
							{:else}
								<p class="text-sm mb-4 min-h-[2.5rem] italic" style="color: rgb(var(--text-tertiary));">
									No description provided
								</p>
							{/if}

							<!-- Info Grid -->
							<div class="space-y-3 mb-4">
								<div class="flex items-center gap-3 px-3 py-2 rounded-lg" style="background-color: rgb(var(--bg-tertiary));">
									<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center shadow-sm">
										<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
										</svg>
									</div>
									<div class="flex-1 min-w-0">
										<p class="text-xs font-medium" style="color: rgb(var(--text-tertiary));">Framework</p>
										<p class="text-sm font-semibold truncate capitalize" style="color: rgb(var(--text-primary));">{project.framework}</p>
									</div>
								</div>

								<div class="flex items-center gap-3 px-3 py-2 rounded-lg" style="background-color: rgb(var(--bg-tertiary));">
									<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-purple-500 to-purple-600 flex items-center justify-center shadow-sm">
										<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
										</svg>
									</div>
									<div class="flex-1 min-w-0">
										<p class="text-xs font-medium" style="color: rgb(var(--text-tertiary));">Branch</p>
										<p class="text-sm font-semibold truncate" style="color: rgb(var(--text-primary));">{project.git_branch}</p>
									</div>
								</div>

								{#if project.last_deployed}
									<div class="flex items-center gap-3 px-3 py-2 rounded-lg" style="background-color: rgb(var(--bg-tertiary));">
										<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-emerald-500 to-emerald-600 flex items-center justify-center shadow-sm">
											<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
											</svg>
										</div>
										<div class="flex-1 min-w-0">
											<p class="text-xs font-medium" style="color: rgb(var(--text-tertiary));">Last Deploy</p>
											<p class="text-sm font-semibold truncate" style="color: rgb(var(--text-primary));">{formatRelativeTime(project.last_deployed)}</p>
										</div>
									</div>
								{/if}
							</div>

							<!-- Domain -->
							{#if project.domains && project.domains.length > 0}
								<div class="pt-4 border-t" style="border-color: rgb(var(--border-primary));">
									<div class="flex items-center gap-2">
										<div class="w-6 h-6 rounded-lg bg-primary-800/20 flex items-center justify-center">
											<svg class="w-4 h-4 text-primary-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
											</svg>
										</div>
										<span class="text-sm text-primary-700 truncate font-semibold">
											{project.domains[0].domain}
										</span>
										{#if project.domains.length > 1}
											<span class="text-xs px-2 py-0.5 rounded-full bg-primary-800/20 text-primary-700 font-medium">
												+{project.domains.length - 1}
											</span>
										{/if}
									</div>
								</div>
							{/if}

							<!-- View Project Arrow -->
							<div class="absolute bottom-6 right-6 opacity-0 group-hover:opacity-100 transition-opacity">
								<div class="w-8 h-8 rounded-lg bg-gradient-brand flex items-center justify-center shadow-md">
									<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
								</div>
							</div>
						</div>
					</div>
				</a>
			{/each}
		</div>
	{/if}
</div>
