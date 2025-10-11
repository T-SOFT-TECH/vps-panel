<script lang="ts">
	import { authStore } from '$lib/stores/auth.svelte';
	import { page } from '$app/stores';
	import Button from './Button.svelte';
	import ThemeToggle from './ThemeToggle.svelte';
	import { fade, fly } from 'svelte/transition';

	let mobileMenuOpen = $state(false);

	const navigation = [
		{ name: 'Dashboard', href: '/dashboard', icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6' },
		{ name: 'Projects', href: '/projects', icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z' },
		{ name: 'Settings', href: '/settings', icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z' }
	];

	function isActive(path: string) {
		return $page.url.pathname === path || $page.url.pathname.startsWith(path + '/');
	}
</script>

<nav class="sticky top-0 z-50 bg-zinc-900/80 dark:bg-zinc-900/80 light:bg-white/80 backdrop-blur-xl border-b border-zinc-800/50 dark:border-zinc-800/50 light:border-zinc-200 shadow-lg">
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
		<div class="flex justify-between h-16">
			<!-- Logo and navigation -->
			<div class="flex">
				<!-- Logo -->
				<div class="flex-shrink-0 flex items-center">
					<a href="/dashboard" class="flex items-center gap-3 group">
						<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-primary-500 to-primary-600 flex items-center justify-center shadow-lg transform group-hover:scale-110 transition-transform duration-200">
							<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
							</svg>
						</div>
						<span class="text-xl font-bold bg-gradient-to-r from-primary-400 to-emerald-400 bg-clip-text text-transparent">
							VPS Panel
						</span>
					</a>
				</div>

				<!-- Desktop Navigation -->
				<div class="hidden sm:ml-8 sm:flex sm:space-x-2">
					{#each navigation as item}
						<a
							href={item.href}
							class="inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200
							       {isActive(item.href)
								? 'bg-primary-500/10 text-primary-400 dark:text-primary-400 light:text-primary-600 border border-primary-500/20'
								: 'text-zinc-400 dark:text-zinc-400 light:text-zinc-600 hover:bg-zinc-800/50 dark:hover:bg-zinc-800/50 light:hover:bg-zinc-100 hover:text-zinc-100 dark:hover:text-zinc-100 light:hover:text-zinc-900'}"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={item.icon} />
							</svg>
							{item.name}
						</a>
					{/each}
				</div>
			</div>

			<!-- Right section: Theme toggle + User menu -->
			<div class="hidden sm:ml-6 sm:flex sm:items-center sm:gap-3">
				<!-- Theme Toggle -->
				<ThemeToggle />

				<!-- User Menu -->
				{#if authStore.user}
					<div class="flex items-center gap-3 pl-3 border-l border-zinc-700/50 dark:border-zinc-700/50 light:border-zinc-200">
						<!-- User Avatar & Name -->
						<div class="flex items-center gap-2">
							<div class="w-9 h-9 rounded-lg bg-gradient-to-br from-primary-500 to-primary-600 flex items-center justify-center text-white font-semibold text-sm shadow-lg">
								{authStore.user.name.charAt(0).toUpperCase()}
							</div>
							<span class="text-sm font-medium text-zinc-100 dark:text-zinc-100 light:text-zinc-900">
								{authStore.user.name}
							</span>
						</div>
						<!-- Sign Out Button -->
						<Button variant="ghost" size="sm" onclick={() => authStore.logout()}>
							<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
							</svg>
							Sign out
						</Button>
					</div>
				{:else}
					<Button variant="primary" size="sm" onclick={() => window.location.href = '/login'}>
						Sign in
					</Button>
				{/if}
			</div>

			<!-- Mobile menu button -->
			<div class="flex items-center sm:hidden gap-2">
				<ThemeToggle />
				<button
					type="button"
					onclick={() => mobileMenuOpen = !mobileMenuOpen}
					class="inline-flex items-center justify-center p-2 rounded-lg text-zinc-400 dark:text-zinc-400 light:text-zinc-600
					       hover:text-zinc-100 dark:hover:text-zinc-100 light:hover:text-zinc-900
					       hover:bg-zinc-800/50 dark:hover:bg-zinc-800/50 light:hover:bg-zinc-100
					       focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 focus:ring-offset-zinc-900 dark:focus:ring-offset-zinc-900 light:focus:ring-offset-white
					       transition-all duration-200"
					aria-label="Toggle mobile menu"
				>
					{#if mobileMenuOpen}
						<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					{:else}
						<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
						</svg>
					{/if}
				</button>
			</div>
		</div>
	</div>

	<!-- Mobile menu -->
	{#if mobileMenuOpen}
		<div class="sm:hidden border-t border-zinc-800/50 dark:border-zinc-800/50 light:border-zinc-200" transition:fly={{ y: -10, duration: 200 }}>
			<!-- Navigation Links -->
			<div class="px-2 pt-2 pb-3 space-y-1">
				{#each navigation as item}
					<a
						href={item.href}
						class="flex items-center gap-3 px-3 py-3 rounded-lg text-base font-medium transition-all duration-200
						       {isActive(item.href)
							? 'bg-primary-500/10 text-primary-400 dark:text-primary-400 light:text-primary-600 border border-primary-500/20'
							: 'text-zinc-400 dark:text-zinc-400 light:text-zinc-600 hover:bg-zinc-800/50 dark:hover:bg-zinc-800/50 light:hover:bg-zinc-100 hover:text-zinc-100 dark:hover:text-zinc-100 light:hover:text-zinc-900'}"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={item.icon} />
						</svg>
						{item.name}
					</a>
				{/each}
			</div>

			<!-- User Section -->
			<div class="pt-4 pb-3 border-t border-zinc-800/50 dark:border-zinc-800/50 light:border-zinc-200">
				{#if authStore.user}
					<div class="px-4">
						<!-- User Info -->
						<div class="flex items-center mb-3">
							<div class="flex-shrink-0">
								<div class="h-12 w-12 rounded-xl bg-gradient-to-br from-primary-500 to-primary-600 flex items-center justify-center text-white font-bold text-lg shadow-lg">
									{authStore.user.name.charAt(0).toUpperCase()}
								</div>
							</div>
							<div class="ml-3">
								<div class="text-base font-semibold text-zinc-100 dark:text-zinc-100 light:text-zinc-900">{authStore.user.name}</div>
								<div class="text-sm text-zinc-400 dark:text-zinc-400 light:text-zinc-600">{authStore.user.email}</div>
							</div>
						</div>
						<!-- Sign Out Button -->
						<Button
							variant="outline"
							class="w-full"
							onclick={() => authStore.logout()}
						>
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
							</svg>
							Sign out
						</Button>
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
