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
	import type { FrameworkType, BaaSType, GitHubRepository, GitProvider } from '$lib/types';

	let loading = $state(false);
	let detecting = $state(false);
	let error = $state('');
	let success = $state(false);
	let detectionError = $state('');

	// Git Provider and repository selection
	let providers = $state<GitProvider[]>([]);
	let repositories = $state<GitHubRepository[]>([]);
	let loadingRepos = $state(false);
	let selectedRepo = $state<GitHubRepository | null>(null);
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

	function selectRepo(repo: GitHubRepository) {
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
				auto_deploy: autoDeploy
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

<div class="max-w-3xl mx-auto">
	<div class="mb-6" in:fly={{ y: -20, duration: 400, delay: 0 }}>
		<a href="/projects" class="text-sm text-green-500 hover:text-green-400 flex items-center">
			<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
			</svg>
			Back to Projects
		</a>
		<h1 class="text-3xl font-bold text-zinc-100 mt-4">Create New Project</h1>
		<p class="mt-1 text-sm text-zinc-400">Deploy a new application to your VPS</p>
	</div>

	<!-- Git Repository Selector -->
	{#if providers.length > 0 && showRepoSelector}
		<div in:fly={{ y: 20, duration: 400, delay: 100 }} class="mb-6">
			<Card>
				<div class="flex items-center justify-between mb-4">
					<div>
						<h2 class="text-lg font-semibold text-zinc-100">Import Git Repository</h2>
						<p class="text-sm text-zinc-400 mt-1">Select a repository from your connected Git providers</p>
					</div>
					<Button variant="ghost" size="sm" onclick={() => showRepoSelector = false}>
						Or enter manually →
					</Button>
				</div>

				<!-- Provider Selector -->
				{#if providers.length > 1}
					<div class="mb-4">
						<label class="block text-sm font-medium text-zinc-300 mb-2">Git Provider</label>
						<select
							value={selectedProvider?.id}
							onchange={(e) => handleProviderChange(Number(e.currentTarget.value))}
							class="block w-full rounded-lg border border-zinc-800 bg-zinc-900 text-zinc-100 px-4 py-2 focus:outline-none focus:ring-2 focus:ring-green-500"
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
				<div class="mb-4">
					<input
						type="text"
						placeholder="Search repositories..."
						bind:value={repoSearchQuery}
						class="block w-full rounded-lg border border-zinc-800 bg-zinc-900 text-zinc-100 placeholder:text-zinc-500 px-4 py-2 focus:outline-none focus:ring-2 focus:ring-green-500"
					/>
				</div>

				<!-- Repository List -->
				{#if loadingRepos}
					<div class="text-center py-12">
						<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-green-500 mx-auto"></div>
						<p class="mt-4 text-zinc-400">Loading repositories...</p>
					</div>
				{:else if filteredRepos.length === 0}
					<div class="text-center py-12">
						<svg class="mx-auto h-12 w-12 text-zinc-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
						</svg>
						<p class="mt-2 text-sm text-zinc-400">
							{repoSearchQuery ? 'No repositories found' : 'No repositories'}
						</p>
					</div>
				{:else}
					<div class="space-y-2 max-h-96 overflow-y-auto">
						{#each filteredRepos as repo}
							<button
								type="button"
								onclick={() => selectRepo(repo)}
								class="w-full text-left p-4 rounded-lg border border-zinc-800 hover:border-green-500 hover:bg-zinc-800 transition-colors"
							>
								<div class="flex items-start justify-between">
									<div class="flex-1">
										<div class="flex items-center space-x-2">
											<svg class="w-5 h-5 text-zinc-400" fill="currentColor" viewBox="0 0 24 24">
												<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
											</svg>
											<span class="font-medium text-zinc-100">{repo.name}</span>
											{#if repo.private}
												<Badge variant="warning">Private</Badge>
											{/if}
										</div>
										<p class="text-xs text-zinc-400 mt-1">{repo.full_name}</p>
									</div>
									<svg class="w-5 h-5 text-zinc-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
								</div>
							</button>
						{/each}
					</div>
				{/if}
			</Card>
		</div>
	{/if}

	<!-- Project Configuration Form -->
	{#if !showRepoSelector || providers.length === 0}
		<div in:fly={{ y: 20, duration: 400, delay: 100 }}>
			<!-- Show selected repository info if coming from selector -->
			{#if selectedRepo}
				<div class="mb-4">
					<Card>
						<div class="flex items-center justify-between">
							<div class="flex items-center space-x-3">
								<svg class="w-8 h-8 text-green-500" fill="currentColor" viewBox="0 0 24 24">
									<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
								</svg>
								<div>
									<p class="font-medium text-zinc-100">{selectedRepo.full_name}</p>
									<p class="text-xs text-zinc-400">{selectedRepo.default_branch} branch</p>
								</div>
							</div>
							<Button variant="ghost" size="sm" onclick={clearRepoSelection}>
								Change
							</Button>
						</div>
					</Card>
				</div>
			{/if}

			<Card>
			<form onsubmit={handleSubmit} class="space-y-6">
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
				<h3 class="text-lg font-medium text-zinc-100">Basic Information</h3>

				<Input
					label="Project Name"
					bind:value={name}
					placeholder="my-awesome-app"
					required
					disabled={loading}
				/>

				<div>
					<label for="description" class="block text-sm font-medium text-zinc-300 mb-1">
						Description
					</label>
					<textarea
						id="description"
						bind:value={description}
						placeholder="A brief description of your project"
						rows="3"
						disabled={loading}
						class="block w-full rounded-lg border border-zinc-700 bg-zinc-900 text-zinc-100 placeholder:text-zinc-500 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-green-500 disabled:opacity-50"
					></textarea>
				</div>
			</div>

			<!-- Git Configuration -->
			<div class="space-y-4 pt-6 border-t border-zinc-800">
				<div class="flex justify-between items-center">
					<h3 class="text-lg font-medium text-zinc-100">Git Repository</h3>
					<div class="flex space-x-2">
						<Button
							variant="secondary"
							size="sm"
							onclick={loadBranches}
							loading={loadingBranches}
							disabled={loadingBranches || !gitUrl}
						>
							{loadingBranches ? 'Loading...' : 'Load Branches'}
						</Button>
						<Button
							variant="secondary"
							size="sm"
							onclick={detectFramework}
							loading={detecting}
							disabled={detecting || !gitUrl}
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
					<div class="p-4 rounded-lg border border-zinc-800 bg-zinc-900">
						<div class="flex items-center space-x-3">
							<div class="animate-spin rounded-full h-5 w-5 border-b-2 border-green-500"></div>
							<p class="text-sm text-zinc-300">Detecting monorepo structure...</p>
						</div>
					</div>
				{/if}

				{#if showDirectorySelector && directories.length > 0}
					<div class="space-y-3">
						<label class="block text-sm font-medium text-zinc-300">
							Select Directory to Deploy
						</label>
						<p class="text-xs text-zinc-400 -mt-2">
							Multiple directories detected in this repository. Choose which one to deploy:
						</p>
						<div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
							{#each directories as dir}
								<button
									type="button"
									onclick={() => selectDirectory(dir)}
									class="p-3 rounded-lg border border-zinc-800 hover:border-green-500 hover:bg-zinc-800 transition-colors text-left flex items-center justify-between group"
								>
									<div class="flex items-center space-x-2">
										<svg class="w-5 h-5 text-zinc-400 group-hover:text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
										</svg>
										<span class="text-sm font-medium text-zinc-100">{dir}</span>
									</div>
									<svg class="w-5 h-5 text-zinc-500 group-hover:text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
								</button>
							{/each}
						</div>
						<button
							type="button"
							onclick={() => { showDirectorySelector = false; rootDirectory = ''; }}
							class="text-xs text-zinc-400 hover:text-zinc-300 underline"
						>
							Skip - deploy from root directory
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
					<p class="text-xs text-zinc-400 -mt-2">
						For monorepos, specify the subdirectory containing your app (e.g., "frontend"). Leave blank if your app is in the root.
					</p>
				{/if}

				<div class="flex items-center mb-3">
					<input
						id="private-repo"
						type="checkbox"
						bind:checked={showPrivateRepoFields}
						class="h-4 w-4 rounded border-zinc-800 text-green-500 focus:ring-green-500"
					/>
					<label for="private-repo" class="ml-2 text-sm text-zinc-300">
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
					<p class="text-xs text-zinc-400">
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
			<div class="space-y-4 pt-6 border-t border-zinc-800">
				<h3 class="text-lg font-medium text-zinc-100">Framework & Backend</h3>

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
			</div>

			<!-- Build Configuration -->
			<div class="space-y-4 pt-6 border-t border-zinc-800">
				<h3 class="text-lg font-medium text-zinc-100">Build Configuration</h3>

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
			<div class="space-y-4 pt-6 border-t border-zinc-800">
				<h3 class="text-lg font-medium text-zinc-100">Port Configuration</h3>

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

			<!-- Deployment Settings -->
			<div class="pt-6 border-t border-zinc-800">
				<div class="flex items-start">
					<div class="flex items-center h-5">
						<input
							id="auto-deploy"
							type="checkbox"
							bind:checked={autoDeploy}
							disabled={loading}
							class="h-4 w-4 rounded border-zinc-800 text-green-500 focus:ring-green-500"
						/>
					</div>
					<div class="ml-3 text-sm">
						<label for="auto-deploy" class="font-medium text-zinc-300">
							Auto Deploy
						</label>
						<p class="text-zinc-400">
							Automatically deploy when changes are pushed to the repository
						</p>
					</div>
				</div>
			</div>

			<!-- Actions -->
			<div class="flex justify-end space-x-3 pt-6 border-t border-zinc-800">
				<Button variant="ghost" onclick={() => window.history.back()} disabled={loading}>
					Cancel
				</Button>
				<Button type="submit" {loading} disabled={loading}>
					{loading ? 'Creating...' : 'Create Project'}
				</Button>
			</div>
		</form>
	</Card>
		</div>
	{/if}
</div>
