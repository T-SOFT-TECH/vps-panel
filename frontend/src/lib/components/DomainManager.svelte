<script lang="ts">
	import { projectsAPI } from '$lib/api/projects';
	import Button from './Button.svelte';
	import Input from './Input.svelte';
	import Alert from './Alert.svelte';
	import Card from './Card.svelte';
	import type { Domain } from '$lib/types';

	interface Props {
		projectId: number;
		domains: Domain[];
		onUpdate?: () => void;
	}

	let { projectId, domains = $bindable([]), onUpdate }: Props = $props();

	let showAddForm = $state(false);
	let editingDomain = $state<Domain | null>(null);
	let loading = $state(false);
	let error = $state('');
	let success = $state('');

	// Form fields
	let newDomain = $state('');
	let newSSLEnabled = $state(true);
	let editDomain = $state('');
	let editSSLEnabled = $state(true);
	let editIsActive = $state(true);

	async function handleAddDomain() {
		if (!newDomain) {
			error = 'Domain cannot be empty';
			return;
		}

		loading = true;
		error = '';
		success = '';

		try {
			const domain = await projectsAPI.addDomain(projectId, {
				domain: newDomain,
				ssl_enabled: newSSLEnabled
			});

			domains = [...domains, domain];
			success = 'Domain added successfully!';
			newDomain = '';
			newSSLEnabled = true;
			showAddForm = false;

			if (onUpdate) onUpdate();
		} catch (err: any) {
			error = err.message || 'Failed to add domain';
		} finally {
			loading = false;
		}
	}

	async function handleUpdateDomain() {
		if (!editingDomain) return;

		loading = true;
		error = '';
		success = '';

		try {
			const updated = await projectsAPI.updateDomain(projectId, editingDomain.id, {
				domain: editDomain,
				ssl_enabled: editSSLEnabled,
				is_active: editIsActive
			});

			domains = domains.map((d) => (d.id === updated.id ? updated : d));
			success = 'Domain updated successfully!';
			editingDomain = null;

			if (onUpdate) onUpdate();
		} catch (err: any) {
			error = err.message || 'Failed to update domain';
		} finally {
			loading = false;
		}
	}

	async function handleDeleteDomain(domainId: number) {
		if (!confirm('Are you sure you want to delete this domain?')) return;

		loading = true;
		error = '';
		success = '';

		try {
			await projectsAPI.deleteDomain(projectId, domainId);
			domains = domains.filter((d) => d.id !== domainId);
			success = 'Domain deleted successfully!';

			if (onUpdate) onUpdate();
		} catch (err: any) {
			error = err.message || 'Failed to delete domain';
		} finally {
			loading = false;
		}
	}

	function startEditing(domain: Domain) {
		editingDomain = domain;
		editDomain = domain.domain;
		editSSLEnabled = domain.ssl_enabled;
		editIsActive = domain.is_active;
		showAddForm = false;
	}

	function cancelEdit() {
		editingDomain = null;
		editDomain = '';
		editSSLEnabled = true;
		editIsActive = true;
	}
</script>

<Card>
	<div class="flex items-center justify-between mb-4">
		<div class="flex items-center gap-2">
			<svg class="w-5 h-5" style="color: rgb(var(--text-brand));" fill="currentColor" viewBox="0 0 24 24">
				<path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
			</svg>
			<h3 class="text-lg font-semibold" style="color: rgb(var(--text-primary));">Domains</h3>
		</div>
		<Button variant="primary" size="sm" onclick={() => { showAddForm = !showAddForm; editingDomain = null; }}>
			{showAddForm ? 'Cancel' : '+ Add Domain'}
		</Button>
	</div>

	{#if error}
		<Alert variant="error" dismissible ondismiss={() => error = ''}>
			{error}
		</Alert>
	{/if}

	{#if success}
		<Alert variant="success" dismissible ondismiss={() => success = ''}>
			{success}
		</Alert>
	{/if}

	<!-- Add Domain Form -->
	{#if showAddForm}
		<div class="mb-6 p-4 rounded-lg" style="background-color: rgb(var(--bg-secondary)); border: 1px solid rgb(var(--border-primary));">
			<h4 class="text-sm font-semibold mb-3" style="color: rgb(var(--text-primary));">Add New Domain</h4>
			<div class="space-y-3">
				<Input
					label="Domain"
					bind:value={newDomain}
					placeholder="myapp.example.com"
					disabled={loading}
				/>
				<div class="flex items-center">
					<input
						type="checkbox"
						id="new-ssl"
						bind:checked={newSSLEnabled}
						disabled={loading}
						class="h-4 w-4 rounded text-primary-800 focus:ring-primary-800"
						style="border-color: rgb(var(--border-primary)); background-color: rgb(var(--bg-secondary));"
					/>
					<label for="new-ssl" class="ml-2 text-sm" style="color: rgb(var(--text-primary));">
						Enable SSL/HTTPS (recommended)
					</label>
				</div>
				<div class="flex gap-2">
					<Button variant="primary" size="sm" {loading} onclick={handleAddDomain}>
						Add Domain
					</Button>
					<Button variant="ghost" size="sm" onclick={() => showAddForm = false} disabled={loading}>
						Cancel
					</Button>
				</div>
			</div>
		</div>
	{/if}

	<!-- Domains List -->
	<div class="space-y-3">
		{#each domains as domain (domain.id)}
			<div class="p-4 rounded-lg transition-colors" style="border: 1px solid rgb(var(--border-primary)); background-color: rgb(var(--bg-primary));">
				{#if editingDomain?.id === domain.id}
					<!-- Edit Mode -->
					<div class="space-y-3">
						<Input
							label="Domain"
							bind:value={editDomain}
							placeholder="myapp.example.com"
							disabled={loading}
						/>
						<div class="space-y-2">
							<div class="flex items-center">
								<input
									type="checkbox"
									id="edit-ssl-{domain.id}"
									bind:checked={editSSLEnabled}
									disabled={loading}
									class="h-4 w-4 rounded text-primary-800 focus:ring-primary-800"
									style="border-color: rgb(var(--border-primary)); background-color: rgb(var(--bg-secondary));"
								/>
								<label for="edit-ssl-{domain.id}" class="ml-2 text-sm" style="color: rgb(var(--text-primary));">
									Enable SSL/HTTPS
								</label>
							</div>
							<div class="flex items-center">
								<input
									type="checkbox"
									id="edit-active-{domain.id}"
									bind:checked={editIsActive}
									disabled={loading}
									class="h-4 w-4 rounded text-primary-800 focus:ring-primary-800"
									style="border-color: rgb(var(--border-primary)); background-color: rgb(var(--bg-secondary));"
								/>
								<label for="edit-active-{domain.id}" class="ml-2 text-sm" style="color: rgb(var(--text-primary));">
									Active
								</label>
							</div>
						</div>
						<div class="flex gap-2">
							<Button variant="primary" size="sm" {loading} onclick={handleUpdateDomain}>
								Save Changes
							</Button>
							<Button variant="ghost" size="sm" onclick={cancelEdit} disabled={loading}>
								Cancel
							</Button>
						</div>
					</div>
				{:else}
					<!-- View Mode -->
					<div class="flex items-center justify-between">
						<div class="flex-1">
							<div class="flex items-center gap-2 mb-1">
								<a
									href="{domain.ssl_enabled ? 'https' : 'http'}://{domain.domain}"
									target="_blank"
									rel="noopener noreferrer"
									class="text-base font-medium hover:underline"
									style="color: rgb(var(--text-brand));"
								>
									{domain.domain}
								</a>
								{#if !domain.is_active}
									<span class="px-2 py-0.5 text-xs rounded-full" style="background-color: rgba(239, 68, 68, 0.1); color: rgb(239, 68, 68);">
										Inactive
									</span>
								{/if}
							</div>
							<div class="flex items-center gap-3 text-xs" style="color: rgb(var(--text-tertiary));">
								<span class="flex items-center gap-1">
									{#if domain.ssl_enabled}
										<svg class="w-3.5 h-3.5" style="color: rgb(34, 197, 94);" fill="currentColor" viewBox="0 0 24 24">
											<path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4z"/>
										</svg>
										<span style="color: rgb(34, 197, 94);">SSL Enabled</span>
									{:else}
										<svg class="w-3.5 h-3.5" style="color: rgb(239, 68, 68);" fill="currentColor" viewBox="0 0 24 24">
											<path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zM11 7h2v2h-2V7zm0 4h2v6h-2v-6z"/>
										</svg>
										<span style="color: rgb(239, 68, 68);">SSL Disabled</span>
									{/if}
								</span>
							</div>
						</div>
						<div class="flex gap-2">
							<Button variant="ghost" size="sm" onclick={() => startEditing(domain)} disabled={loading}>
								Edit
							</Button>
							<Button
								variant="ghost"
								size="sm"
								onclick={() => handleDeleteDomain(domain.id)}
								disabled={loading || domains.length <= 1}
							>
								<svg class="w-4 h-4" style="color: rgb(239, 68, 68);" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
								</svg>
							</Button>
						</div>
					</div>
				{/if}
			</div>
		{/each}

		{#if domains.length === 0}
			<div class="text-center py-8">
				<svg class="mx-auto h-12 w-12 mb-3" style="color: rgb(var(--text-tertiary));" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
				</svg>
				<p class="text-sm" style="color: rgb(var(--text-secondary));">No domains configured yet</p>
			</div>
		{/if}
	</div>

	<!-- DNS Configuration Help -->
	<div class="mt-4 p-3 rounded-lg" style="background-color: rgb(var(--bg-secondary)); border-left: 3px solid rgb(10, 101, 34);">
		<p class="text-xs font-medium mb-1" style="color: rgb(var(--text-primary));">DNS Configuration</p>
		<p class="text-xs" style="color: rgb(var(--text-secondary));">
			To use a custom domain, add an <code class="px-1 py-0.5 rounded" style="background-color: rgb(var(--bg-primary)); color: rgb(var(--text-brand));">A</code> record pointing to your VPS IP address in your domain's DNS settings.
		</p>
	</div>
</Card>
