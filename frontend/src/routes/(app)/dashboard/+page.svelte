<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { projectsAPI } from '$lib/api/projects';
	import { deploymentsAPI } from '$lib/api/deployments';
	import Card from '$lib/components/Card.svelte';
	import Button from '$lib/components/Button.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import { formatRelativeTime } from '$lib/utils/format';
	import type { Project, Deployment } from '$lib/types';

	let projects = $state<Project[]>([]);
	let recentDeployments = $state<Deployment[]>([]);
	let loading = $state(true);
	let stats = $state({
		totalProjects: 0,
		activeProjects: 0,
		totalDeployments: 0,
		successRate: 0
	});

	onMount(async () => {
		try {
			const { projects: projectList } = await projectsAPI.getAll();
			projects = projectList.slice(0, 5);

			// Calculate stats
			stats.totalProjects = projectList.length;
			stats.activeProjects = projectList.filter(p => p.status === 'active').length;

			// Get recent deployments from all projects
			const allDeployments: Deployment[] = [];
			for (const project of projectList.slice(0, 3)) {
				try {
					const { deployments } = await deploymentsAPI.getAll(project.id);
					allDeployments.push(...deployments);
				} catch (err) {
					console.error('Failed to fetch deployments:', err);
				}
			}

			recentDeployments = allDeployments
				.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
				.slice(0, 5);

			stats.totalDeployments = allDeployments.length;
			const successfulDeployments = allDeployments.filter(d => d.status === 'success').length;
			stats.successRate = allDeployments.length > 0
				? Math.round((successfulDeployments / allDeployments.length) * 100)
				: 0;

		} catch (error) {
			console.error('Failed to load dashboard data:', error);
		} finally {
			loading = false;
		}
	});

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
</script>

<svelte:head>
	<title>Dashboard - VPS Panel</title>
</svelte:head>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex justify-between items-center" in:fly={{ y: -20, duration: 400, delay: 0 }}>
		<div>
			<h1 class="text-3xl font-bold text-zinc-100">Dashboard</h1>
			<p class="mt-1 text-sm text-zinc-400">Overview of your projects and deployments</p>
		</div>
		<Button onclick={() => window.location.href = '/projects/new'}>
			New Project
		</Button>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-green-500 mx-auto"></div>
			<p class="mt-4 text-zinc-400">Loading dashboard...</p>
		</div>
	{:else}
		<!-- Stats -->
		<div class="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
			<div in:fly={{ y: 20, duration: 400, delay: 100 }}>
				<Card padding={false} class="overflow-hidden">
					<div class="p-5">
						<div class="flex items-center">
							<div class="flex-shrink-0">
								<svg class="h-6 w-6 text-zinc-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
								</svg>
							</div>
							<div class="ml-5 w-0 flex-1">
								<dl>
									<dt class="text-sm font-medium text-zinc-400 truncate">Total Projects</dt>
									<dd class="text-2xl font-semibold text-zinc-100">{stats.totalProjects}</dd>
								</dl>
							</div>
						</div>
					</div>
				</Card>
			</div>

			<div in:fly={{ y: 20, duration: 400, delay: 150 }}>
				<Card padding={false} class="overflow-hidden">
					<div class="p-5">
						<div class="flex items-center">
							<div class="flex-shrink-0">
								<svg class="h-6 w-6 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
								</svg>
							</div>
							<div class="ml-5 w-0 flex-1">
								<dl>
									<dt class="text-sm font-medium text-zinc-400 truncate">Active Projects</dt>
									<dd class="text-2xl font-semibold text-zinc-100">{stats.activeProjects}</dd>
								</dl>
							</div>
						</div>
					</div>
				</Card>
			</div>

			<div in:fly={{ y: 20, duration: 400, delay: 200 }}>
				<Card padding={false} class="overflow-hidden">
					<div class="p-5">
						<div class="flex items-center">
							<div class="flex-shrink-0">
								<svg class="h-6 w-6 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
								</svg>
							</div>
							<div class="ml-5 w-0 flex-1">
								<dl>
									<dt class="text-sm font-medium text-zinc-400 truncate">Deployments</dt>
									<dd class="text-2xl font-semibold text-zinc-100">{stats.totalDeployments}</dd>
								</dl>
							</div>
						</div>
					</div>
				</Card>
			</div>

			<div in:fly={{ y: 20, duration: 400, delay: 250 }}>
				<Card padding={false} class="overflow-hidden">
					<div class="p-5">
						<div class="flex items-center">
							<div class="flex-shrink-0">
								<svg class="h-6 w-6 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
								</svg>
							</div>
							<div class="ml-5 w-0 flex-1">
								<dl>
									<dt class="text-sm font-medium text-zinc-400 truncate">Success Rate</dt>
									<dd class="text-2xl font-semibold text-zinc-100">{stats.successRate}%</dd>
								</dl>
							</div>
						</div>
					</div>
				</Card>
			</div>
		</div>

		<!-- Recent Projects -->
		<div in:fly={{ y: 20, duration: 400, delay: 300 }}>
			<Card>
			<div class="flex justify-between items-center mb-4">
				<h2 class="text-lg font-semibold text-zinc-100">Recent Projects</h2>
				<a href="/projects" class="text-sm text-green-500 hover:text-green-400">
					View all
				</a>
			</div>

			{#if projects.length === 0}
				<div class="text-center py-8">
					<svg class="mx-auto h-12 w-12 text-zinc-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
					</svg>
					<h3 class="mt-2 text-sm font-medium text-zinc-100">No projects</h3>
					<p class="mt-1 text-sm text-zinc-400">Get started by creating a new project.</p>
					<div class="mt-6">
						<Button onclick={() => window.location.href = '/projects/new'}>
							New Project
						</Button>
					</div>
				</div>
			{:else}
				<div class="space-y-3">
					{#each projects as project}
						<a href="/projects/{project.id}" class="block hover:bg-zinc-800 -mx-6 px-6 py-3 transition-colors rounded">
							<div class="flex items-center justify-between">
								<div class="flex-1">
									<h3 class="text-sm font-medium text-zinc-100">{project.name}</h3>
									<p class="text-sm text-zinc-400 mt-1">
										{project.framework} • {project.git_branch}
									</p>
								</div>
								<div class="ml-4">
									<Badge variant={getStatusVariant(project.status)}>
										{project.status}
									</Badge>
								</div>
							</div>
						</a>
					{/each}
				</div>
			{/if}
			</Card>
		</div>

		<!-- Recent Deployments -->
		<div in:fly={{ y: 20, duration: 400, delay: 350 }}>
			<Card>
			<div class="flex justify-between items-center mb-4">
				<h2 class="text-lg font-semibold text-zinc-100">Recent Deployments</h2>
			</div>

			{#if recentDeployments.length === 0}
				<div class="text-center py-8 text-zinc-400">
					No deployments yet
				</div>
			{:else}
				<div class="space-y-3">
					{#each recentDeployments as deployment}
						<div class="flex items-center justify-between py-3 border-b border-zinc-800 last:border-0">
							<div class="flex-1">
								<p class="text-sm font-medium text-zinc-100">
									{deployment.commit_message || 'No commit message'}
								</p>
								<p class="text-xs text-zinc-400 mt-1">
									{deployment.commit_author} • {formatRelativeTime(deployment.created_at)}
								</p>
							</div>
							<div class="ml-4">
								<Badge variant={getStatusVariant(deployment.status)}>
									{deployment.status}
								</Badge>
							</div>
						</div>
					{/each}
				</div>
			{/if}
			</Card>
		</div>
	{/if}
</div>
