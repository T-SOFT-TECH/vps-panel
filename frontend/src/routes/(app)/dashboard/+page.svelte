<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly, scale } from 'svelte/transition';
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

<div class="space-y-8 pb-8">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4" in:fly={{ y: -20, duration: 500, delay: 0 }}>
		<div>
			<h1 class="text-4xl font-bold bg-gradient-to-r from-primary-400 via-primary-500 to-emerald-400 bg-clip-text text-transparent">
				Dashboard
			</h1>
			<p class="mt-2 text-base text-zinc-400 dark:text-zinc-400 light:text-zinc-600">
				Overview of your projects and deployments
			</p>
		</div>
		<Button onclick={() => window.location.href = '/projects/new'} class="btn-primary">
			<svg class="w-5 h-5 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
			</svg>
			New Project
		</Button>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20" in:scale={{ duration: 300 }}>
			<div class="text-center">
				<div class="relative w-16 h-16 mx-auto mb-4">
					<div class="absolute inset-0 rounded-full border-4 border-zinc-800 dark:border-zinc-800 light:border-zinc-200"></div>
					<div class="absolute inset-0 rounded-full border-4 border-primary-500 border-t-transparent animate-spin"></div>
				</div>
				<p class="text-zinc-400 dark:text-zinc-400 light:text-zinc-600 font-medium">Loading dashboard...</p>
			</div>
		</div>
	{:else}
		<!-- Stats Grid -->
		<div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
			<!-- Total Projects Card -->
			<div in:fly={{ y: 20, duration: 500, delay: 100 }}>
				<div class="modern-card p-6 group hover:scale-105 transition-transform duration-200">
					<div class="flex items-center justify-between">
						<div class="flex-1">
							<p class="text-sm font-medium text-zinc-400 dark:text-zinc-400 light:text-zinc-600 mb-1">
								Total Projects
							</p>
							<p class="text-3xl font-bold text-zinc-100 dark:text-zinc-100 light:text-zinc-900">
								{stats.totalProjects}
							</p>
						</div>
						<div class="w-14 h-14 rounded-xl bg-gradient-to-br from-blue-500/20 to-blue-600/20 border border-blue-500/30 flex items-center justify-center">
							<svg class="w-7 h-7 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
							</svg>
						</div>
					</div>
				</div>
			</div>

			<!-- Active Projects Card -->
			<div in:fly={{ y: 20, duration: 500, delay: 150 }}>
				<div class="modern-card p-6 group hover:scale-105 transition-transform duration-200">
					<div class="flex items-center justify-between">
						<div class="flex-1">
							<p class="text-sm font-medium text-zinc-400 dark:text-zinc-400 light:text-zinc-600 mb-1">
								Active Projects
							</p>
							<p class="text-3xl font-bold text-zinc-100 dark:text-zinc-100 light:text-zinc-900">
								{stats.activeProjects}
							</p>
						</div>
						<div class="w-14 h-14 rounded-xl bg-gradient-to-br from-primary-500/20 to-primary-600/20 border border-primary-500/30 flex items-center justify-center">
							<svg class="w-7 h-7 text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
						</div>
					</div>
				</div>
			</div>

			<!-- Total Deployments Card -->
			<div in:fly={{ y: 20, duration: 500, delay: 200 }}>
				<div class="modern-card p-6 group hover:scale-105 transition-transform duration-200">
					<div class="flex items-center justify-between">
						<div class="flex-1">
							<p class="text-sm font-medium text-zinc-400 dark:text-zinc-400 light:text-zinc-600 mb-1">
								Deployments
							</p>
							<p class="text-3xl font-bold text-zinc-100 dark:text-zinc-100 light:text-zinc-900">
								{stats.totalDeployments}
							</p>
						</div>
						<div class="w-14 h-14 rounded-xl bg-gradient-to-br from-purple-500/20 to-purple-600/20 border border-purple-500/30 flex items-center justify-center">
							<svg class="w-7 h-7 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
							</svg>
						</div>
					</div>
				</div>
			</div>

			<!-- Success Rate Card -->
			<div in:fly={{ y: 20, duration: 500, delay: 250 }}>
				<div class="modern-card p-6 group hover:scale-105 transition-transform duration-200">
					<div class="flex items-center justify-between">
						<div class="flex-1">
							<p class="text-sm font-medium text-zinc-400 dark:text-zinc-400 light:text-zinc-600 mb-1">
								Success Rate
							</p>
							<p class="text-3xl font-bold text-zinc-100 dark:text-zinc-100 light:text-zinc-900">
								{stats.successRate}%
							</p>
						</div>
						<div class="w-14 h-14 rounded-xl bg-gradient-to-br from-emerald-500/20 to-emerald-600/20 border border-emerald-500/30 flex items-center justify-center">
							<svg class="w-7 h-7 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
							</svg>
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Content Grid -->
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<!-- Recent Projects -->
			<div in:fly={{ y: 20, duration: 500, delay: 300 }}>
				<div class="modern-card p-6 h-full">
					<div class="flex justify-between items-center mb-6">
						<h2 class="text-xl font-bold text-zinc-100 dark:text-zinc-100 light:text-zinc-900 flex items-center gap-2">
							<svg class="w-5 h-5 text-primary-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
							</svg>
							Recent Projects
						</h2>
						<a
							href="/projects"
							class="text-sm font-medium text-primary-400 hover:text-primary-300 dark:text-primary-400 light:text-primary-600 transition-colors flex items-center gap-1"
						>
							View all
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
							</svg>
						</a>
					</div>

					{#if projects.length === 0}
						<div class="text-center py-12">
							<div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-zinc-800/50 dark:bg-zinc-800/50 light:bg-zinc-100 border border-zinc-700/50 dark:border-zinc-700/50 light:border-zinc-200 mb-4">
								<svg class="w-8 h-8 text-zinc-500 dark:text-zinc-500 light:text-zinc-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
								</svg>
							</div>
							<h3 class="text-base font-semibold text-zinc-100 dark:text-zinc-100 light:text-zinc-900 mb-1">No projects yet</h3>
							<p class="text-sm text-zinc-400 dark:text-zinc-400 light:text-zinc-600 mb-6">
								Get started by creating your first project
							</p>
							<Button onclick={() => window.location.href = '/projects/new'} class="btn-primary">
								<svg class="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
								</svg>
								Create Project
							</Button>
						</div>
					{:else}
						<div class="space-y-3">
							{#each projects as project}
								<a
									href="/projects/{project.id}"
									class="block rounded-xl p-4 transition-all duration-200
									       hover:bg-zinc-800/50 dark:hover:bg-zinc-800/50 light:hover:bg-zinc-50
									       border border-transparent hover:border-zinc-700/50 dark:hover:border-zinc-700/50 light:hover:border-zinc-200
									       group"
								>
									<div class="flex items-center justify-between">
										<div class="flex-1 min-w-0">
											<div class="flex items-center gap-3 mb-2">
												<div class="w-10 h-10 rounded-lg bg-gradient-to-br from-primary-500/20 to-primary-600/20 border border-primary-500/30 flex items-center justify-center flex-shrink-0">
													<svg class="w-5 h-5 text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
													</svg>
												</div>
												<div class="min-w-0 flex-1">
													<h3 class="text-sm font-semibold text-zinc-100 dark:text-zinc-100 light:text-zinc-900 truncate group-hover:text-primary-400 transition-colors">
														{project.name}
													</h3>
													<p class="text-xs text-zinc-400 dark:text-zinc-400 light:text-zinc-600 flex items-center gap-2 mt-1">
														<span class="inline-flex items-center">
															<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
																<path d="M2 6a2 2 0 012-2h12a2 2 0 012 2v2a2 2 0 01-2 2H4a2 2 0 01-2-2V6z"/>
															</svg>
															{project.framework}
														</span>
														<span class="text-zinc-600 dark:text-zinc-600 light:text-zinc-400">•</span>
														<span class="inline-flex items-center">
															<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
																<path fill-rule="evenodd" d="M12.316 3.051a1 1 0 01.633 1.265l-4 12a1 1 0 11-1.898-.632l4-12a1 1 0 011.265-.633zM5.707 6.293a1 1 0 010 1.414L3.414 10l2.293 2.293a1 1 0 11-1.414 1.414l-3-3a1 1 0 010-1.414l3-3a1 1 0 011.414 0zm8.586 0a1 1 0 011.414 0l3 3a1 1 0 010 1.414l-3 3a1 1 0 11-1.414-1.414L16.586 10l-2.293-2.293a1 1 0 010-1.414z" clip-rule="evenodd"/>
															</svg>
															{project.git_branch}
														</span>
													</p>
												</div>
											</div>
										</div>
										<div class="ml-4 flex-shrink-0">
											<Badge variant={getStatusVariant(project.status)}>
												{project.status}
											</Badge>
										</div>
									</div>
								</a>
							{/each}
						</div>
					{/if}
				</div>
			</div>

			<!-- Recent Deployments -->
			<div in:fly={{ y: 20, duration: 500, delay: 350 }}>
				<div class="modern-card p-6 h-full">
					<div class="flex justify-between items-center mb-6">
						<h2 class="text-xl font-bold text-zinc-100 dark:text-zinc-100 light:text-zinc-900 flex items-center gap-2">
							<svg class="w-5 h-5 text-primary-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
							</svg>
							Recent Deployments
						</h2>
					</div>

					{#if recentDeployments.length === 0}
						<div class="text-center py-12">
							<div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-zinc-800/50 dark:bg-zinc-800/50 light:bg-zinc-100 border border-zinc-700/50 dark:border-zinc-700/50 light:border-zinc-200 mb-4">
								<svg class="w-8 h-8 text-zinc-500 dark:text-zinc-500 light:text-zinc-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
								</svg>
							</div>
							<h3 class="text-base font-semibold text-zinc-100 dark:text-zinc-100 light:text-zinc-900 mb-1">No deployments yet</h3>
							<p class="text-sm text-zinc-400 dark:text-zinc-400 light:text-zinc-600">
								Deployments will appear here once you create and deploy a project
							</p>
						</div>
					{:else}
						<div class="space-y-4">
							{#each recentDeployments as deployment}
								<div class="relative pl-6 pb-4 last:pb-0">
									<!-- Timeline line -->
									<div class="absolute left-[7px] top-2 bottom-0 w-px bg-zinc-700/50 dark:bg-zinc-700/50 light:bg-zinc-300 last:hidden"></div>

									<!-- Timeline dot -->
									<div class="absolute left-0 top-2 w-4 h-4 rounded-full border-2 border-zinc-700 dark:border-zinc-700 light:border-zinc-300
									            {getStatusVariant(deployment.status) === 'success' ? 'bg-primary-500' :
									             getStatusVariant(deployment.status) === 'error' ? 'bg-red-500' :
									             getStatusVariant(deployment.status) === 'warning' ? 'bg-yellow-500' : 'bg-zinc-500'}">
									</div>

									<div class="flex items-start justify-between gap-3">
										<div class="flex-1 min-w-0">
											<p class="text-sm font-medium text-zinc-100 dark:text-zinc-100 light:text-zinc-900 truncate">
												{deployment.commit_message || 'No commit message'}
											</p>
											<div class="flex items-center gap-2 mt-1 text-xs text-zinc-400 dark:text-zinc-400 light:text-zinc-600">
												<span class="inline-flex items-center">
													<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
														<path fill-rule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clip-rule="evenodd"/>
													</svg>
													{deployment.commit_author}
												</span>
												<span class="text-zinc-600 dark:text-zinc-600 light:text-zinc-400">•</span>
												<span class="inline-flex items-center">
													<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
														<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd"/>
													</svg>
													{formatRelativeTime(deployment.created_at)}
												</span>
											</div>
										</div>
										<div class="flex-shrink-0">
											<Badge variant={getStatusVariant(deployment.status)}>
												{deployment.status}
											</Badge>
										</div>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>
		</div>

		<!-- Quick Actions -->
		<div in:fly={{ y: 20, duration: 500, delay: 400 }}>
			<div class="modern-card p-6">
				<h2 class="text-xl font-bold text-zinc-100 dark:text-zinc-100 light:text-zinc-900 mb-4 flex items-center gap-2">
					<svg class="w-5 h-5 text-primary-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
					</svg>
					Quick Actions
				</h2>
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
					<a
						href="/projects/new"
						class="flex items-center gap-3 p-4 rounded-xl border border-zinc-700/50 dark:border-zinc-700/50 light:border-zinc-200
						       hover:border-primary-500/50 hover:bg-zinc-800/30 dark:hover:bg-zinc-800/30 light:hover:bg-zinc-50
						       transition-all duration-200 group"
					>
						<div class="w-10 h-10 rounded-lg bg-gradient-to-br from-primary-500/20 to-primary-600/20 border border-primary-500/30 flex items-center justify-center flex-shrink-0">
							<svg class="w-5 h-5 text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
							</svg>
						</div>
						<div>
							<p class="text-sm font-semibold text-zinc-100 dark:text-zinc-100 light:text-zinc-900 group-hover:text-primary-400 transition-colors">
								New Project
							</p>
							<p class="text-xs text-zinc-400 dark:text-zinc-400 light:text-zinc-600">
								Create project
							</p>
						</div>
					</a>

					<a
						href="/projects"
						class="flex items-center gap-3 p-4 rounded-xl border border-zinc-700/50 dark:border-zinc-700/50 light:border-zinc-200
						       hover:border-primary-500/50 hover:bg-zinc-800/30 dark:hover:bg-zinc-800/30 light:hover:bg-zinc-50
						       transition-all duration-200 group"
					>
						<div class="w-10 h-10 rounded-lg bg-gradient-to-br from-blue-500/20 to-blue-600/20 border border-blue-500/30 flex items-center justify-center flex-shrink-0">
							<svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
							</svg>
						</div>
						<div>
							<p class="text-sm font-semibold text-zinc-100 dark:text-zinc-100 light:text-zinc-900 group-hover:text-primary-400 transition-colors">
								All Projects
							</p>
							<p class="text-xs text-zinc-400 dark:text-zinc-400 light:text-zinc-600">
								View all projects
							</p>
						</div>
					</a>

					<a
						href="/settings"
						class="flex items-center gap-3 p-4 rounded-xl border border-zinc-700/50 dark:border-zinc-700/50 light:border-zinc-200
						       hover:border-primary-500/50 hover:bg-zinc-800/30 dark:hover:bg-zinc-800/30 light:hover:bg-zinc-50
						       transition-all duration-200 group"
					>
						<div class="w-10 h-10 rounded-lg bg-gradient-to-br from-purple-500/20 to-purple-600/20 border border-purple-500/30 flex items-center justify-center flex-shrink-0">
							<svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
							</svg>
						</div>
						<div>
							<p class="text-sm font-semibold text-zinc-100 dark:text-zinc-100 light:text-zinc-900 group-hover:text-primary-400 transition-colors">
								Settings
							</p>
							<p class="text-xs text-zinc-400 dark:text-zinc-400 light:text-zinc-600">
								Configure panel
							</p>
						</div>
					</a>

					<a
						href="https://github.com/yourusername/vps-panel"
						target="_blank"
						rel="noopener noreferrer"
						class="flex items-center gap-3 p-4 rounded-xl border border-zinc-700/50 dark:border-zinc-700/50 light:border-zinc-200
						       hover:border-primary-500/50 hover:bg-zinc-800/30 dark:hover:bg-zinc-800/30 light:hover:bg-zinc-50
						       transition-all duration-200 group"
					>
						<div class="w-10 h-10 rounded-lg bg-gradient-to-br from-zinc-600/20 to-zinc-700/20 border border-zinc-600/30 flex items-center justify-center flex-shrink-0">
							<svg class="w-5 h-5 text-zinc-400" fill="currentColor" viewBox="0 0 24 24">
								<path d="M12 0C5.374 0 0 5.373 0 12c0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.3 24 12c0-6.627-5.373-12-12-12z"/>
							</svg>
						</div>
						<div>
							<p class="text-sm font-semibold text-zinc-100 dark:text-zinc-100 light:text-zinc-900 group-hover:text-primary-400 transition-colors">
								Documentation
							</p>
							<p class="text-xs text-zinc-400 dark:text-zinc-400 light:text-zinc-600">
								View on GitHub
							</p>
						</div>
					</a>
				</div>
			</div>
		</div>
	{/if}
</div>
