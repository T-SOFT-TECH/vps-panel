<script lang="ts">
	import { authStore } from '$lib/stores/auth.svelte';
	import { page } from '$app/stores';
	import Button from './Button.svelte';

	let mobileMenuOpen = $state(false);

	const navigation = [
		{ name: 'Dashboard', href: '/dashboard' },
		{ name: 'Projects', href: '/projects' },
		{ name: 'Settings', href: '/settings' }
	];

	function isActive(path: string) {
		return $page.url.pathname === path;
	}
</script>

<nav class="bg-zinc-900 border-b border-zinc-800">
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
		<div class="flex justify-between h-16">
			<!-- Logo and navigation -->
			<div class="flex">
				<div class="flex-shrink-0 flex items-center">
					<a href="/" class="text-2xl font-bold text-green-500">
						VPS Panel
					</a>
				</div>
				<div class="hidden sm:ml-6 sm:flex sm:space-x-8">
					{#each navigation as item}
						<a
							href={item.href}
							class="inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium {isActive(item.href)
								? 'border-green-500 text-zinc-100'
								: 'border-transparent text-zinc-400 hover:border-zinc-600 hover:text-zinc-300'}"
						>
							{item.name}
						</a>
					{/each}
				</div>
			</div>

			<!-- User menu -->
			<div class="hidden sm:ml-6 sm:flex sm:items-center space-x-4">
				{#if authStore.user}
					<span class="text-sm text-zinc-300">
						{authStore.user.name}
					</span>
					<Button variant="ghost" size="sm" onclick={() => authStore.logout()}>
						Sign out
					</Button>
				{:else}
					<Button variant="ghost" size="sm" onclick={() => window.location.href = '/login'}>
						Sign in
					</Button>
				{/if}
			</div>

			<!-- Mobile menu button -->
			<div class="flex items-center sm:hidden">
				<button
					type="button"
					onclick={() => mobileMenuOpen = !mobileMenuOpen}
					class="inline-flex items-center justify-center p-2 rounded-md text-zinc-400 hover:text-zinc-300 hover:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-green-500"
				>
					<span class="sr-only">Open main menu</span>
					<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
					</svg>
				</button>
			</div>
		</div>
	</div>

	<!-- Mobile menu -->
	{#if mobileMenuOpen}
		<div class="sm:hidden">
			<div class="pt-2 pb-3 space-y-1">
				{#each navigation as item}
					<a
						href={item.href}
						class="block pl-3 pr-4 py-2 border-l-4 text-base font-medium {isActive(item.href)
							? 'bg-zinc-800 border-green-500 text-green-400'
							: 'border-transparent text-zinc-400 hover:bg-zinc-800 hover:border-zinc-600 hover:text-zinc-300'}"
					>
						{item.name}
					</a>
				{/each}
			</div>
			<div class="pt-4 pb-3 border-t border-zinc-800">
				{#if authStore.user}
					<div class="flex items-center px-4">
						<div class="flex-shrink-0">
							<div class="h-10 w-10 rounded-full bg-green-600 flex items-center justify-center text-white font-bold">
								{authStore.user.name.charAt(0).toUpperCase()}
							</div>
						</div>
						<div class="ml-3">
							<div class="text-base font-medium text-zinc-100">{authStore.user.name}</div>
							<div class="text-sm font-medium text-zinc-400">{authStore.user.email}</div>
						</div>
					</div>
					<div class="mt-3 space-y-1">
						<button
							onclick={() => authStore.logout()}
							class="block w-full text-left px-4 py-2 text-base font-medium text-zinc-400 hover:text-zinc-300 hover:bg-zinc-800"
						>
							Sign out
						</button>
					</div>
				{:else}
					<div class="px-4">
						<Button variant="primary" class="w-full" onclick={() => window.location.href = '/login'}>
							Sign in
						</Button>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</nav>
