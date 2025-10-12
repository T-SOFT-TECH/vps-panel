<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { projectsAPI } from '$lib/api/projects';
	import { gitProvidersAPI } from '$lib/api/git-providers';
	import { authStore } from '$lib/stores/auth.svelte';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';
	import Card from '$lib/components/Card.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import type { FrameworkType, BaaSType, GitHubRepository, GiteaRepository, GitProvider } from '$lib/types';

	type Repository = GitHubRepository | GiteaRepository;

	let loading = $state(false);
	let detecting = $state(false);
	let error = $state('');
	let success = $state(false);
	let detectionError = $state('');

	// Git Provider and repository selection
	let providers = $state<GitProvider[]>([]);
	let repositories = $state<Repository[]>([]);
	let loadingRepos = $state(false);
	let selectedRepo = $state<Repository | null>(null);
	let repoSearchQuery = $state('');
	let showRepoSelector = $state(true);
	let selectedProvider = $state<GitProvider | null>(null);

	// Form fields
	let name = $state('');
	let description = $state('');
	let gitUrl = $state('');
	let gitBranch = $state('main');
	let gitUsername = $state('');
	let gitToken = $state('');
	let rootDirectory = $state('');
	let framework = $state<FrameworkType>('sveltekit');
	let baasType = $state<BaaSType>('');
	let buildCommand = $state('npm run build');
	let outputDir = $state('build');
	let installCommand = $state('npm install');
	let nodeVersion = $state('20');
	let frontendPort = $state(3000);
	let backendPort = $state(8090);
	let autoDeploy = $state(true);
	let customDomain = $state('');

	// Branch listing
	let branches = $state<string[]>([]);
	let loadingBranches = $state(false);
	let showPrivateRepoFields = $state(false);

	// Directory listing (for monorepos)
	let directories = $state<string[]>([]);
	let loadingDirectories = $state(false);
	let showDirectorySelector = $state(false);

	// Load Git Providers on mount
	onMount(async () => {
		await loadProviders();
	});

	async function loadProviders() {
		try {
			const data = await gitProvidersAPI.getAll();
			providers = data.providers.filter(p => p.connected);

			// Auto-select first connected provider and load its repos
			if (providers.length > 0) {
				selectedProvider = providers[0];
				await loadRepositories(selectedProvider.id);
			}
		} catch (err) {
			console.error('Failed to load providers:', err);
		}
	}

	async function loadRepositories(providerId: number) {
		loadingRepos = true;
		repositories = [];
		try {
			const data = await gitProvidersAPI.listRepositories(providerId);
			repositories = data.repositories;
		} catch (err) {
			console.error('Failed to load repositories:', err);
		} finally {
			loadingRepos = false;
		}
	}

	async function handleProviderChange(providerId: number) {
		const provider = providers.find(p => p.id === providerId);
		if (provider) {
			selectedProvider = provider;
			await loadRepositories(providerId);
		}
	}

	function selectRepo(repo: Repository) {
		selectedRepo = repo;
		gitUrl = repo.clone_url;
		gitBranch = repo.default_branch;
		name = repo.name;
		showRepoSelector = false;

		// For private repos, automatically use provider OAuth credentials
		if (repo.private && selectedProvider) {
			showPrivateRepoFields = true;
			gitUsername = selectedProvider.username || '';
			// Use provider type as marker (e.g., 'github_oauth', 'gitea_oauth')
			gitToken = `${selectedProvider.type}_oauth`;
		}

		// Trigger directory detection, framework detection, and load branches
		setTimeout(() => {
			loadDirectories();
			loadBranches();
		}, 500);
	}

	function clearRepoSelection() {
		selectedRepo = null;
		showRepoSelector = true;
		gitUrl = '';
		name = '';
	}

	const filteredRepos = $derived(
		repositories.filter(repo =>
			repo.name.toLowerCase().includes(repoSearchQuery.toLowerCase()) ||
			repo.full_name.toLowerCase().includes(repoSearchQuery.toLowerCase())
		)
	);

	const frameworkOptions = [
		{ value: 'sveltekit', label: 'SvelteKit' },
		{ value: 'react', label: 'React' },
		{ value: 'vue', label: 'Vue 3' },
		{ value: 'angular', label: 'Angular' },
		{ value: 'nextjs', label: 'Next.js' },
		{ value: 'nuxt', label: 'Nuxt' }
	];

	const baasOptions = [
		{ value: '', label: 'None' },
		{ value: 'pocketbase', label: 'PocketBase' },
		{ value: 'supabase', label: 'Supabase' },
		{ value: 'firebase', label: 'Firebase' },
		{ value: 'appwrite', label: 'Appwrite' }
	];

	const nodeVersionOptions = [
		{ value: '20', label: 'Node.js 20' },
		{ value: '18', label: 'Node.js 18' },
		{ value: '16', label: 'Node.js 16' }
	];

	// Load branches from repository
	async function loadBranches() {
		if (!gitUrl) {
			return;
		}

		loadingBranches = true;
		try {
			const result = await projectsAPI.listBranches(gitUrl, gitUsername, gitToken);
			branches = result.branches;
			if (branches.length > 0 && !branches.includes(gitBranch)) {
				gitBranch = branches[0];
			}
		} catch (err) {
			console.error('Failed to load branches:', err);
			branches = [];
		} finally {
			loadingBranches = false;
		}
	}

	// Load directories from repository (for monorepo detection)
	async function loadDirectories() {
		if (!gitUrl || !gitBranch) {
			return;
		}

		loadingDirectories = true;
		directories = [];
		showDirectorySelector = false;

		try {
			const result = await projectsAPI.listDirectories(
				gitUrl,
				gitBranch,
				gitUsername || undefined,
				gitToken || undefined
			);
			directories = result.directories;

			// Show directory selector if multiple directories found
			if (directories.length > 1) {
				showDirectorySelector = true;
			} else if (directories.length === 1) {
				// Auto-select if only one directory
				rootDirectory = directories[0];
				// Re-trigger framework detection with the root directory
				setTimeout(() => detectFramework(), 300);
			}
		} catch (err) {
			console.error('Failed to load directories:', err);
		} finally {
			loadingDirectories = false;
		}
	}

	function selectDirectory(dir: string) {
		rootDirectory = dir;
		showDirectorySelector = false;
		// Re-trigger framework detection with the selected directory
		detectFramework();
	}

	// Auto-detect framework and BaaS
	async function detectFramework() {
		if (!gitUrl) {
			detectionError = 'Please enter a repository URL first';
			return;
		}

		detecting = true;
		detectionError = '';

		try {
			const result = await projectsAPI.detectFramework(
				gitUrl,
				gitBranch,
				gitUsername || undefined,
				gitToken || undefined,
				rootDirectory || undefined
			);

			if (result.detected) {
				// Apply detected framework
				framework = result.framework;

				// Apply BaaS if detected
				if (result.baas_type) {
					baasType = result.baas_type;
				}

				// Apply framework-specific configurations
				buildCommand = result.build_command;
				installCommand = result.install_command;
				outputDir = result.output_dir;
				nodeVersion = result.node_version;
				frontendPort = result.frontend_port;

				// Apply BaaS port if BaaS detected
				if (result.baas_type && result.backend_port) {
					backendPort = result.backend_port;
				}
			} else {
				detectionError = 'Could not auto-detect framework. Please select manually.';
			}
		} catch (err) {
			detectionError = 'Failed to analyze repository. Please check the URL and try again.';
			console.error(err);
		} finally {
			detecting = false;
		}
	}

	// Update defaults when framework is manually changed
	$effect(() => {
		switch (framework) {
			case 'sveltekit':
				buildCommand = 'npm run build';
				installCommand = 'npm install';
				outputDir = 'build';
				frontendPort = 3000;
				nodeVersion = '20';
				break;
			case 'nextjs':
				buildCommand = 'npm run build';
				installCommand = 'npm install';
				outputDir = '.next';
				frontendPort = 3000;
				nodeVersion = '20';
				break;
			case 'nuxt':
				buildCommand = 'npm run build';
				installCommand = 'npm install';
				outputDir = '.output';
				frontendPort = 3000;
				nodeVersion = '20';
				break;
			case 'react':
				buildCommand = 'npm run build';
				installCommand = 'npm install';
				outputDir = 'dist';
				frontendPort = 5173; // Vite default
				nodeVersion = '20';
				break;
			case 'vue':
				buildCommand = 'npm run build';
				installCommand = 'npm install';
				outputDir = 'dist';
				frontendPort = 5173; // Vite default
				nodeVersion = '20';
				break;
			case 'angular':
				buildCommand = 'ng build';
				installCommand = 'npm install';
				outputDir = 'dist';
				frontendPort = 4200;
				nodeVersion = '20';
				break;
			default:
				buildCommand = 'npm run build';
				installCommand = 'npm install';
				outputDir = 'dist';
				frontendPort = 3000;
				nodeVersion = '20';
				break;
		}
	});

	// Update backend port when BaaS is manually changed
	$effect(() => {
		switch (baasType) {
			case 'pocketbase':
				backendPort = 8090;
				break;
			case 'supabase':
				backendPort = 54321;
				break;
			case 'firebase':
				backendPort = 9099;
				break;
			case 'appwrite':
				backendPort = 80;
				break;
			default:
				backendPort = 8080;
				break;
		}
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();

		if (!name || !gitUrl) {
			error = 'Please fill in all required fields';
			return;
		}

		loading = true;
		error = '';

		try {
			const project = await projectsAPI.create({
				name,
				description,
				git_url: gitUrl,
				git_branch: gitBranch,
				git_username: gitUsername || undefined,
				git_token: gitToken || undefined,
				root_directory: rootDirectory || undefined,
				framework,
				baas_type: baasType,
				build_command: buildCommand,
				output_dir: outputDir,
				install_command: installCommand,
				node_version: nodeVersion,
				frontend_port: frontendPort,
				backend_port: backendPort,
				auto_deploy: autoDeploy,
				custom_domain: customDomain || undefined
			});

			success = true;
			setTimeout(() => {
				goto(`/projects/${project.id}`);
			}, 1500);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create project';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>New Project - VPS Panel</title>
</svelte:head>

<div class="max-w-3xl mx-auto space-y-6 pb-8">
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
			<div class="flex items-center gap-4">
				<div class="w-16 h-16 rounded-2xl bg-gradient-brand flex items-center justify-center shadow-xl glow-green float">
					<svg class="w-9 h-9 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
				</div>
				<div>
					<h1 class="text-4xl font-bold bg-gradient-to-r from-primary-700 to-primary-900 bg-clip-text text-transparent mb-1">
						Create New Project
					</h1>
					<p class="text-base" style="color: rgb(var(--text-secondary));">
						Deploy a new application to your VPS in minutes
					</p>
				</div>
			</div>
		</div>
	</div>

	<!-- Git Repository Selector -->
	{#if providers.length > 0 && showRepoSelector}
		<div class="slide-in-up stagger-1">
			<div class="relative overflow-hidden rounded-2xl">
				<div class="absolute inset-0 bg-gradient-to-br from-primary-600/10 to-primary-800/10 rounded-2xl"></div>
				<div class="relative modern-card p-6 border-0">
					<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
						<div class="flex items-center gap-3">
							<div class="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center shadow-lg">
								<svg class="w-6 h-6 text-white" fill="currentColor" viewBox="0 0 24 24">
									<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
								</svg>
							</div>
							<div>
								<h2 class="text-xl font-bold" style="color: rgb(var(--text-primary));">Import Git Repository</h2>
								<p class="text-sm mt-1" style="color: rgb(var(--text-secondary));">Select from your connected providers</p>
							</div>
						</div>
						<Button variant="ghost" size="sm" onclick={() => showRepoSelector = false} class="hover:scale-105 transition-transform">
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
							</svg>
							Enter manually
						</Button>
					</div>

				<!-- Provider Selector -->
				{#if providers.length > 1}
					<div class="mb-4">
						<label class="block text-sm font-semibold mb-2" style="color: rgb(var(--text-primary));">
							<svg class="w-4 h-4 inline mr-1 -mt-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
							</svg>
							Git Provider
						</label>
						<select
							value={selectedProvider?.id}
							onchange={(e) => handleProviderChange(Number(e.currentTarget.value))}
							class="modern-input block w-full"
						>
							{#each providers as provider}
								<option value={provider.id}>
									{provider.name} ({provider.type}) - @{provider.username}
								</option>
							{/each}
						</select>
					</div>
				{/if}

				<!-- Search -->
				<div class="mb-4 relative group">
					<div class="absolute inset-0 bg-gradient-brand rounded-xl opacity-0 group-focus-within:opacity-10 blur-xl transition-opacity"></div>
					<div class="relative">
						<svg class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 transition-colors group-focus-within:text-primary-600" style="color: rgb(var(--text-tertiary));" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
						</svg>
						<input
							type="text"
							placeholder="Search repositories by name..."
							bind:value={repoSearchQuery}
							class="modern-input block w-full pl-12"
						/>
					</div>
				</div>

				<!-- Repository List -->
				{#if loadingRepos}
					<div class="text-center py-12">
						<div class="relative w-12 h-12 mx-auto">
							<div class="absolute inset-0 rounded-full border-4 border-primary-200"></div>
							<div class="absolute inset-0 rounded-full border-4 border-primary-800 border-t-transparent animate-spin"></div>
						</div>
						<p class="mt-4 font-medium" style="color: rgb(var(--text-secondary));">Loading repositories...</p>
					</div>
				{:else if filteredRepos.length === 0}
					<div class="text-center py-12">
						<div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-primary-800/10 to-primary-600/10 mb-4">
							<svg class="w-8 h-8 text-primary-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
							</svg>
						</div>
						<p class="text-sm font-medium" style="color: rgb(var(--text-secondary));">
							{repoSearchQuery ? 'No repositories found' : 'No repositories available'}
						</p>
					</div>
				{:else}
					<div class="space-y-2 max-h-96 overflow-y-auto pr-2">
						{#each filteredRepos as repo, i}
							<button
								type="button"
								onclick={() => selectRepo(repo)}
								class="w-full text-left p-4 rounded-xl transition-all duration-200 group hover-lift fade-in"
								style="border: 1px solid rgb(var(--border-primary)); background-color: rgb(var(--bg-secondary)); animation-delay: {i * 0.05}s;"
								onmouseenter={(e) => { e.currentTarget.style.borderColor = 'rgb(10, 101, 34)'; e.currentTarget.style.transform = 'translateY(-2px)'; }}
								onmouseleave={(e) => { e.currentTarget.style.borderColor = 'rgb(var(--border-primary))'; e.currentTarget.style.transform = 'translateY(0)'; }}
							>
								<div class="flex items-start justify-between gap-3">
									<div class="flex-1 min-w-0">
										<div class="flex items-center gap-2 mb-1">
											<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-zinc-700 to-zinc-800 flex items-center justify-center flex-shrink-0 group-hover:scale-110 transition-transform">
												<svg class="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 24 24">
													<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
												</svg>
											</div>
											<span class="font-semibold truncate group-hover:text-primary-700 transition-colors" style="color: rgb(var(--text-primary));">{repo.name}</span>
											{#if repo.private}
												<Badge variant="warning">Private</Badge>
											{/if}
										</div>
										<p class="text-xs truncate" style="color: rgb(var(--text-secondary));">{repo.full_name}</p>
									</div>
									<svg class="w-5 h-5 flex-shrink-0 group-hover:translate-x-1 transition-transform" style="color: rgb(var(--text-tertiary));" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
								</div>
							</button>
						{/each}
					</div>
				{/if}
				</div>
			</div>
		</div>
	{/if}

	<!-- Project Configuration Form -->
	{#if !showRepoSelector || providers.length === 0}
		<div class="slide-in-up stagger-2">
			<!-- Show selected repository info if coming from selector -->
			{#if selectedRepo}
				<div class="mb-4">
					<div class="modern-card p-4 border-l-4 border-primary-700 hover-lift">
						<div class="flex items-center justify-between">
							<div class="flex items-center gap-3">
								<div class="w-10 h-10 rounded-lg bg-gradient-to-br from-zinc-700 to-zinc-800 flex items-center justify-center shadow-lg">
									<svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 24 24">
										<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
									</svg>
								</div>
								<div>
									<p class="font-semibold flex items-center gap-2" style="color: rgb(var(--text-primary));">
										{selectedRepo.full_name}
										<Badge variant="success">Selected</Badge>
									</p>
									<p class="text-xs flex items-center gap-1 mt-1" style="color: rgb(var(--text-secondary));">
										<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
										</svg>
										{selectedRepo.default_branch} branch
									</p>
								</div>
							</div>
							<Button variant="ghost" size="sm" onclick={clearRepoSelection} class="hover:scale-105 transition-transform">
								Change
							</Button>
						</div>
					</div>
				</div>
			{/if}

			<div class="relative overflow-hidden rounded-2xl">
				<div class="absolute inset-0 bg-gradient-to-br from-primary-600/5 to-primary-800/5"></div>
				<div class="relative modern-card p-6 border-0">
			<form onsubmit={handleSubmit} class="space-y-8">
				{#if error}
					<Alert variant="error" dismissible ondismiss={() => error = ''}>
						{error}
					</Alert>
				{/if}

				{#if success}
					<Alert variant="success">
						Project created successfully! Redirecting...
					</Alert>
				{/if}

			<!-- Basic Info -->
			<div class="space-y-4">
				<div class="flex items-center gap-3 mb-4">
					<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center shadow-lg">
						<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<h3 class="text-xl font-bold" style="color: rgb(var(--text-primary));">Basic Information</h3>
				</div>

				<Input
					label="Project Name"
					bind:value={name}
					placeholder="my-awesome-app"
					required
					disabled={loading}
				/>

				<div>
					<label for="description" class="block text-sm font-semibold mb-2" style="color: rgb(var(--text-primary));">
						Description
					</label>
					<textarea
						id="description"
						bind:value={description}
						placeholder="A brief description of your project"
						rows="3"
						disabled={loading}
						class="modern-input block w-full"
					></textarea>
				</div>
			</div>

			<!-- Git Configuration -->
			<div class="space-y-4 pt-6" style="border-top: 1px solid rgb(var(--border-primary));">
				<div class="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4 mb-4">
					<div class="flex items-center gap-3">
						<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-purple-600 flex items-center justify-center shadow-lg">
							<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
							</svg>
						</div>
						<h3 class="text-xl font-bold" style="color: rgb(var(--text-primary));">Git Repository</h3>
					</div>
					<div class="flex gap-2">
						<Button
							variant="secondary"
							size="sm"
							onclick={loadBranches}
							loading={loadingBranches}
							disabled={loadingBranches || !gitUrl}
							class="hover:scale-105 transition-transform"
						>
							{loadingBranches ? 'Loading...' : 'Load Branches'}
						</Button>
						<Button
							variant="secondary"
							size="sm"
							onclick={detectFramework}
							loading={detecting}
							disabled={detecting || !gitUrl}
							class="hover:scale-105 transition-transform"
						>
							{detecting ? 'Detecting...' : 'Auto-Detect'}
						</Button>
					</div>
				</div>

				{#if detectionError}
					<Alert variant="warning" dismissible ondismiss={() => detectionError = ''}>
						{detectionError}
					</Alert>
				{/if}

				<Input
					label="Repository URL"
					bind:value={gitUrl}
					placeholder="https://github.com/username/repo.git"
					required
					disabled={loading}
				/>

				{#if loadingDirectories}
					<div class="p-4 rounded-lg" style="border: 1px solid rgb(var(--border-primary)); background-color: rgb(var(--bg-secondary));">
						<div class="flex items-center space-x-3">
							<div class="animate-spin rounded-full h-5 w-5 border-b-2 border-primary-800"></div>
							<p class="text-sm" style="color: rgb(var(--text-primary));">Detecting monorepo structure...</p>
						</div>
					</div>
				{/if}

				{#if showDirectorySelector && directories.length > 0}
					<div class="space-y-3 p-4 rounded-xl" style="background-color: rgb(var(--bg-secondary)); border: 2px dashed rgb(var(--border-primary));">
						<div class="flex items-center gap-2 mb-2">
							<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-amber-500 to-amber-600 flex items-center justify-center">
								<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
								</svg>
							</div>
							<label class="block text-sm font-semibold" style="color: rgb(var(--text-primary));">
								Select Directory to Deploy
							</label>
						</div>
						<p class="text-xs" style="color: rgb(var(--text-secondary));">
							Monorepo detected! Choose which directory contains your application:
						</p>
						<div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
							{#each directories as dir, i}
								<button
									type="button"
									onclick={() => selectDirectory(dir)}
									class="p-3 rounded-xl transition-all duration-200 text-left flex items-center justify-between group hover-lift fade-in"
									style="border: 2px solid rgb(var(--border-primary)); background-color: rgb(var(--bg-primary)); animation-delay: {i * 0.1}s;"
									onmouseenter={(e) => { e.currentTarget.style.borderColor = 'rgb(10, 101, 34)'; e.currentTarget.style.transform = 'translateY(-2px)'; }}
									onmouseleave={(e) => { e.currentTarget.style.borderColor = 'rgb(var(--border-primary))'; e.currentTarget.style.transform = 'translateY(0)'; }}
								>
									<div class="flex items-center gap-2">
										<svg class="w-5 h-5 group-hover:text-primary-700 transition-colors" style="color: rgb(var(--text-tertiary));" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
										</svg>
										<span class="text-sm font-semibold group-hover:text-primary-700 transition-colors" style="color: rgb(var(--text-primary));">{dir}</span>
									</div>
									<svg class="w-5 h-5 group-hover:translate-x-1 transition-transform" style="color: rgb(var(--text-tertiary));" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
								</button>
							{/each}
						</div>
						<button
							type="button"
							onclick={() => { showDirectorySelector = false; rootDirectory = ''; }}
							class="text-xs font-medium underline hover:text-primary-700 transition-colors"
							style="color: rgb(var(--text-secondary));"
						>
							Skip - deploy from root directory instead
						</button>
					</div>
				{/if}

				{#if !showDirectorySelector || directories.length === 0}
					<Input
						label="Root Directory (Optional)"
						bind:value={rootDirectory}
						placeholder="e.g., frontend, client, packages/web"
						disabled={loading}
					/>
					<p class="text-xs -mt-2" style="color: rgb(var(--text-secondary));">
						For monorepos, specify the subdirectory containing your app (e.g., "frontend"). Leave blank if your app is in the root.
					</p>
				{/if}

				<div class="flex items-center mb-3">
					<input
						id="private-repo"
						type="checkbox"
						bind:checked={showPrivateRepoFields}
						class="h-4 w-4 rounded text-primary-800 focus:ring-primary-800"
						style="border-color: rgb(var(--border-primary)); background-color: rgb(var(--bg-secondary));"
					/>
					<label for="private-repo" class="ml-2 text-sm" style="color: rgb(var(--text-primary));">
						Private Repository (requires authentication)
					</label>
				</div>

				{#if showPrivateRepoFields}
					<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
						<Input
							label="Git Username"
							bind:value={gitUsername}
							placeholder="username or token name"
							disabled={loading}
						/>
						<Input
							label="Access Token"
							type="password"
							bind:value={gitToken}
							placeholder="ghp_xxxxxxxxxxxx"
							disabled={loading}
						/>
					</div>
					<p class="text-xs" style="color: rgb(var(--text-secondary));">
						For GitHub, use a personal access token. For GitLab, use a project/personal access token.
					</p>
				{/if}

				{#if branches.length > 0}
					<Select
						label="Branch"
						bind:value={gitBranch}
						options={branches.map(b => ({ value: b, label: b }))}
						disabled={loading}
					/>
				{:else}
					<Input
						label="Branch"
						bind:value={gitBranch}
						placeholder="main"
						disabled={loading}
					/>
				{/if}
			</div>

			<!-- Framework & Backend -->
			<div class="space-y-4 pt-6" style="border-top: 1px solid rgb(var(--border-primary));">
				<div class="flex items-center gap-3 mb-4">
					<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500 to-emerald-600 flex items-center justify-center shadow-lg">
						<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
						</svg>
					</div>
					<h3 class="text-xl font-bold" style="color: rgb(var(--text-primary));">Framework & Backend</h3>
				</div>

				<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
					<Select
						label="Framework"
						bind:value={framework}
						options={frameworkOptions}
						required
						disabled={loading}
					/>

					<Select
						label="Backend/BaaS"
						bind:value={baasType}
						options={baasOptions}
						disabled={loading}
					/>
				</div>

				{#if baasType === 'pocketbase'}
					<div class="relative overflow-hidden p-5 rounded-xl fade-in" style="background: linear-gradient(135deg, rgba(10, 101, 34, 0.05) 0%, rgba(10, 101, 34, 0.1) 100%); border: 2px solid rgba(10, 101, 34, 0.3);">
						<div class="absolute top-0 right-0 w-32 h-32 bg-primary-600 rounded-full blur-3xl opacity-10"></div>
						<div class="relative flex items-start gap-4">
							<div class="w-12 h-12 rounded-xl bg-gradient-brand flex items-center justify-center shadow-lg flex-shrink-0 glow-green">
								<svg class="w-6 h-6 text-white" fill="currentColor" viewBox="0 0 24 24">
									<path d="M13 9h-2V7h2m0 10h-2v-6h2m-1-9A10 10 0 0 0 2 12a10 10 0 0 0 10 10 10 10 0 0 0 10-10A10 10 0 0 0 12 2z"/>
								</svg>
							</div>
							<div class="flex-1">
								<p class="text-base font-bold mb-2" style="color: rgb(var(--text-primary));">PocketBase Backend Included</p>
								<p class="text-sm mb-3" style="color: rgb(var(--text-secondary));">
									PocketBase will be automatically deployed alongside your frontend using the <strong class="text-primary-700">official binary from GitHub</strong>.
									Your deployment will include:
								</p>
								<ul class="text-sm space-y-2">
									<li class="flex items-start gap-2" style="color: rgb(var(--text-secondary));">
										<svg class="w-5 h-5 mt-0.5 flex-shrink-0 text-primary-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
										</svg>
										<span>SQLite database with realtime subscriptions</span>
									</li>
									<li class="flex items-start gap-2" style="color: rgb(var(--text-secondary));">
										<svg class="w-5 h-5 mt-0.5 flex-shrink-0 text-primary-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
										</svg>
										<span>Built-in authentication and file storage</span>
									</li>
									<li class="flex items-start gap-2" style="color: rgb(var(--text-secondary));">
										<svg class="w-5 h-5 mt-0.5 flex-shrink-0 text-primary-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
										</svg>
										<span>Admin dashboard at <code class="px-2 py-0.5 rounded font-mono text-xs" style="background-color: rgb(var(--bg-primary)); color: rgb(var(--text-brand)); border: 1px solid rgb(var(--border-primary));">/_</code></span>
									</li>
									<li class="flex items-start gap-2" style="color: rgb(var(--text-secondary));">
										<svg class="w-5 h-5 mt-0.5 flex-shrink-0 text-primary-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
										</svg>
										<span>REST and Realtime APIs at <code class="px-2 py-0.5 rounded font-mono text-xs" style="background-color: rgb(var(--bg-primary)); color: rgb(var(--text-brand)); border: 1px solid rgb(var(--border-primary));">/api/*</code></span>
									</li>
								</ul>
							</div>
						</div>
					</div>
				{/if}
			</div>

			<!-- Build Configuration -->
			<div class="space-y-4 pt-6" style="border-top: 1px solid rgb(var(--border-primary));">
				<div class="flex items-center gap-3 mb-4">
					<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-orange-500 to-orange-600 flex items-center justify-center shadow-lg">
						<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
						</svg>
					</div>
					<h3 class="text-xl font-bold" style="color: rgb(var(--text-primary));">Build Configuration</h3>
				</div>

				<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
					<Input
						label="Install Command"
						bind:value={installCommand}
						placeholder="npm install"
						disabled={loading}
					/>

					<Input
						label="Build Command"
						bind:value={buildCommand}
						placeholder="npm run build"
						disabled={loading}
					/>

					<Input
						label="Output Directory"
						bind:value={outputDir}
						placeholder="build"
						disabled={loading}
					/>

					<Select
						label="Node Version"
						bind:value={nodeVersion}
						options={nodeVersionOptions}
						disabled={loading}
					/>
				</div>
			</div>

			<!-- Port Configuration -->
			<div class="space-y-4 pt-6" style="border-top: 1px solid rgb(var(--border-primary));">
				<div class="flex items-center gap-3 mb-4">
					<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-500 to-indigo-600 flex items-center justify-center shadow-lg">
						<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
						</svg>
					</div>
					<h3 class="text-xl font-bold" style="color: rgb(var(--text-primary));">Port Configuration</h3>
				</div>

				<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
					<Input
						label="Frontend Port"
						type="number"
						bind:value={frontendPort}
						placeholder="3000"
						disabled={loading}
					/>

					{#if baasType}
						<Input
							label="Backend Port"
							type="number"
							bind:value={backendPort}
							placeholder="8090"
							disabled={loading}
						/>
					{/if}
				</div>
			</div>

			<!-- Domain Configuration -->
			<div class="space-y-4 pt-6" style="border-top: 1px solid rgb(var(--border-primary));">
				<div class="flex items-center gap-3 mb-4">
					<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-cyan-500 to-cyan-600 flex items-center justify-center shadow-lg">
						<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
						</svg>
					</div>
					<h3 class="text-xl font-bold" style="color: rgb(var(--text-primary));">Domain (Optional)</h3>
				</div>

				<Input
					label="Custom Domain"
					bind:value={customDomain}
					placeholder="myapp.example.com"
					disabled={loading}
				/>
				<p class="text-xs -mt-2" style="color: rgb(var(--text-secondary));">
					Leave blank to auto-generate a subdomain (e.g., myapp-1.panel.yourdomain.com)
				</p>
			</div>

			<!-- Deployment Settings -->
			<div class="pt-6" style="border-top: 1px solid rgb(var(--border-primary));">
				<div class="flex items-center gap-3 mb-4">
					<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-pink-500 to-pink-600 flex items-center justify-center shadow-lg">
						<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
						</svg>
					</div>
					<h3 class="text-xl font-bold" style="color: rgb(var(--text-primary));">Deployment Settings</h3>
				</div>
				<div class="flex items-start p-4 rounded-xl hover-lift transition-all" style="background-color: rgb(var(--bg-secondary)); border: 1px solid rgb(var(--border-primary));">
					<div class="flex items-center h-5">
						<input
							id="auto-deploy"
							type="checkbox"
							bind:checked={autoDeploy}
							disabled={loading}
							class="h-5 w-5 rounded text-primary-800 focus:ring-2 focus:ring-primary-600 transition-all cursor-pointer"
							style="border-color: rgb(var(--border-primary)); background-color: rgb(var(--bg-primary));"
						/>
					</div>
					<div class="ml-3 flex-1">
						<label for="auto-deploy" class="text-base font-semibold cursor-pointer" style="color: rgb(var(--text-primary));">
							Auto Deploy
						</label>
						<p class="text-sm mt-1" style="color: rgb(var(--text-secondary));">
							Automatically deploy when changes are pushed to the repository
						</p>
					</div>
				</div>
			</div>

			<!-- Actions -->
			<div class="flex justify-end gap-3 pt-6" style="border-top: 1px solid rgb(var(--border-primary));">
				<Button variant="ghost" onclick={() => window.history.back()} disabled={loading} class="hover:scale-105 transition-transform">
					<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
					Cancel
				</Button>
				<Button type="submit" {loading} disabled={loading} class="btn-primary glow-green-hover hover:scale-105 transition-transform">
					<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					{loading ? 'Creating Project...' : 'Create Project'}
				</Button>
			</div>
		</form>
				</div>
			</div>
		</div>
	{/if}
</div>
