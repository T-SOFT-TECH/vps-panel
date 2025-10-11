<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { projectsAPI } from '$lib/api/projects';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';
	import Card from '$lib/components/Card.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import type { FrameworkType, BaaSType, Project } from '$lib/types';

	let projectId = parseInt($page.params.id);
	let loading = $state(false);
	let loadingProject = $state(true);
	let detecting = $state(false);
	let error = $state('');
	let success = $state(false);
	let detectionError = $state('');

	// Form fields
	let name = $state('');
	let description = $state('');
	let gitUrl = $state('');
	let gitBranch = $state('main');
	let framework = $state<FrameworkType>('sveltekit');
	let baasType = $state<BaaSType>('');
	let buildCommand = $state('npm run build');
	let outputDir = $state('build');
	let installCommand = $state('npm install');
	let nodeVersion = $state('20');
	let frontendPort = $state(3000);
	let backendPort = $state(8090);
	let autoDeploy = $state(true);

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

	// Load project data
	onMount(async () => {
		try {
			const project = await projectsAPI.getById(projectId);

			// Populate form fields
			name = project.name;
			description = project.description;
			gitUrl = project.git_url;
			gitBranch = project.git_branch;
			framework = project.framework;
			baasType = project.baas_type;
			buildCommand = project.build_command;
			outputDir = project.output_dir;
			installCommand = project.install_command;
			nodeVersion = project.node_version;
			frontendPort = project.frontend_port;
			backendPort = project.backend_port;
			autoDeploy = project.auto_deploy;
		} catch (err) {
			error = 'Failed to load project';
			console.error(err);
		} finally {
			loadingProject = false;
		}
	});

	// Auto-detect framework and BaaS
	async function detectFramework() {
		if (!gitUrl) {
			detectionError = 'Please enter a repository URL first';
			return;
		}

		detecting = true;
		detectionError = '';

		try {
			const result = await projectsAPI.detectFramework(gitUrl, gitBranch);

			if (result.detected) {
				framework = result.framework;
				if (result.baas_type) {
					baasType = result.baas_type;
				}
				// Success message will be shown via the UI
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

	// Update defaults based on framework
	$effect(() => {
		switch (framework) {
			case 'sveltekit':
				buildCommand = 'npm run build';
				outputDir = 'build';
				break;
			case 'nextjs':
				buildCommand = 'npm run build';
				outputDir = '.next';
				break;
			case 'nuxt':
				buildCommand = 'npm run build';
				outputDir = '.output';
				break;
			case 'react':
			case 'vue':
				buildCommand = 'npm run build';
				outputDir = 'dist';
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
			await projectsAPI.update(projectId, {
				name,
				description,
				git_url: gitUrl,
				git_branch: gitBranch,
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
				goto(`/projects/${projectId}`);
			}, 1500);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update project';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Edit Project - VPS Panel</title>
</svelte:head>

{#if loadingProject}
	<div class="text-center py-12">
		<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-800 mx-auto"></div>
		<p class="mt-4" style="color: rgb(var(--text-secondary));">Loading project...</p>
	</div>
{:else}
	<div class="max-w-3xl mx-auto">
		<div class="mb-6">
			<a href="/projects/{projectId}" class="flex items-center text-sm" style="color: rgb(var(--text-brand));">
				<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
				</svg>
				Back to Project
			</a>
			<h1 class="text-3xl font-bold mt-4" style="color: rgb(var(--text-primary));">Edit Project</h1>
			<p class="mt-1 text-sm" style="color: rgb(var(--text-secondary));">Update your project configuration</p>
		</div>

		<Card>
			<form onsubmit={handleSubmit} class="space-y-6">
				{#if error}
					<Alert variant="error" dismissible ondismiss={() => error = ''}>
						{error}
					</Alert>
				{/if}

				{#if success}
					<Alert variant="success">
						Project updated successfully! Redirecting...
					</Alert>
				{/if}

				<!-- Basic Info -->
				<div class="space-y-4">
					<h3 class="text-lg font-medium" style="color: rgb(var(--text-primary));">Basic Information</h3>

					<Input
						label="Project Name"
						bind:value={name}
						placeholder="my-awesome-app"
						required
						disabled={loading}
					/>

					<div>
						<label for="description" class="block text-sm font-medium mb-1" style="color: rgb(var(--text-primary));">
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
					<div class="flex justify-between items-center">
						<h3 class="text-lg font-medium" style="color: rgb(var(--text-primary));">Git Repository</h3>
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

					<Input
						label="Branch"
						bind:value={gitBranch}
						placeholder="main"
						disabled={loading}
					/>
				</div>

				<!-- Framework & Backend -->
				<div class="space-y-4 pt-6" style="border-top: 1px solid rgb(var(--border-primary));">
					<h3 class="text-lg font-medium" style="color: rgb(var(--text-primary));">Framework & Backend</h3>

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
				<div class="space-y-4 pt-6" style="border-top: 1px solid rgb(var(--border-primary));">
					<h3 class="text-lg font-medium" style="color: rgb(var(--text-primary));">Build Configuration</h3>

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
					<h3 class="text-lg font-medium" style="color: rgb(var(--text-primary));">Port Configuration</h3>

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
				<div class="pt-6" style="border-top: 1px solid rgb(var(--border-primary));">
					<div class="flex items-start">
						<div class="flex items-center h-5">
							<input
								id="auto-deploy"
								type="checkbox"
								bind:checked={autoDeploy}
								disabled={loading}
								class="h-4 w-4 rounded text-primary-800 focus:ring-primary-800"
								style="border-color: rgb(var(--border-primary)); background-color: rgb(var(--bg-secondary));"
							/>
						</div>
						<div class="ml-3 text-sm">
							<label for="auto-deploy" class="font-medium" style="color: rgb(var(--text-primary));">
								Auto Deploy
							</label>
							<p style="color: rgb(var(--text-secondary));">
								Automatically deploy when changes are pushed to the repository
							</p>
						</div>
					</div>
				</div>

				<!-- Actions -->
				<div class="flex justify-end space-x-3 pt-6" style="border-top: 1px solid rgb(var(--border-primary));">
					<Button variant="ghost" onclick={() => window.history.back()} disabled={loading}>
						Cancel
					</Button>
					<Button type="submit" {loading} disabled={loading}>
						{loading ? 'Saving...' : 'Save Changes'}
					</Button>
				</div>
			</form>
		</Card>
	</div>
{/if}
