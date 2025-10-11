<script lang="ts">
	import { authStore } from '$lib/stores/auth.svelte';
	import { page } from '$app/stores';
	import Button from './Button.svelte';
	import ThemeToggle from './ThemeToggle.svelte';

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

<nav class="sticky top-0 z-50 modern-card border-b shadow-lg">
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
		<div class="flex justify-between h-16">
			<!-- Logo -->
			<div class="flex items-center">
				<a href="/dashboard" class="flex items-center gap-3 group">
					<img
						src="/img/My Icon.png"
						alt="TSOFT Logo"
						class="w-10 h-10 rounded-lg shadow-md group-hover:scale-110 transition-transform duration-200"
					/>
					<span class="text-lg font-bold" style="color: rgb(var(--text-primary));">
						VPS Panel
					</span>
				</a>

				<!-- Desktop Navigation -->
				<div class="hidden sm:ml-8 sm:flex sm:space-x-2">
					{#each navigation as item}
						<a
							href={item.href}
							class="inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200
							       {isActive(item.href)
								? 'bg-gradient-brand text-white shadow-md'
								: 'hover:bg-opacity-50'}"
							style="color: {isActive(item.href) ? 'white' : 'rgb(var(--text-secondary))'}"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={item.icon} />
							</svg>
							{item.name}
						</a>
					{/each}
				</div>
			</div>

			<!-- Right section -->
			<div class="hidden sm:flex sm:items-center sm:gap-3">
				<ThemeToggle />

				{#if authStore.user}
					<div class="flex items-center gap-3 pl-3 border-l" style="border-color: rgb(var(--border-primary));">
						<div class="flex items-center gap-2">
							<div class="w-9 h-9 rounded-lg bg-gradient-brand flex items-center justify-center text-white font-semibold text-sm shadow-md">
								{authStore.user.name.charAt(0).toUpperCase()}
							</div>
							<span class="text-sm font-medium" style="color: rgb(var(--text-primary));">
								{authStore.user.name}
							</span>
						</div>
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
					class="inline-flex items-center justify-center p-2 rounded-lg transition-all duration-200
					       hover:scale-105"
					style="color: rgb(var(--text-secondary));"
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
		<div class="sm:hidden border-t" style="border-color: rgb(var(--border-primary));">
			<div class="px-2 pt-2 pb-3 space-y-1">
				{#each navigation as item}
					<a
						href={item.href}
						class="flex items-center gap-3 px-3 py-3 rounded-lg text-base font-medium transition-all duration-200
						       {isActive(item.href)
							? 'bg-gradient-brand text-white'
							: ''}"
						style="color: {isActive(item.href) ? 'white' : 'rgb(var(--text-secondary))'}"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={item.icon} />
						</svg>
						{item.name}
					</a>
				{/each}
			</div>

			<div class="pt-4 pb-3 border-t" style="border-color: rgb(var(--border-primary));">
				{#if authStore.user}
					<div class="px-4">
						<div class="flex items-center mb-3">
							<div class="h-12 w-12 rounded-xl bg-gradient-brand flex items-center justify-center text-white font-bold text-lg shadow-lg">
								{authStore.user.name.charAt(0).toUpperCase()}
							</div>
							<div class="ml-3">
								<div class="text-base font-semibold" style="color: rgb(var(--text-primary));">{authStore.user.name}</div>
								<div class="text-sm" style="color: rgb(var(--text-secondary));">{authStore.user.email}</div>
							</div>
						</div>
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
