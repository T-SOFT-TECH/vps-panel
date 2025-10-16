<script lang="ts">
	import { authStore } from '$lib/stores/auth.svelte';
	import { page } from '$app/stores';
	import Button from './Button.svelte';
	import ThemeToggle from './ThemeToggle.svelte';

	let mobileMenuOpen = $state(false);
	let scrolled = $state(false);

	// Handle scroll effect
	if (typeof window !== 'undefined') {
		window.addEventListener('scroll', () => {
			scrolled = window.scrollY > 10;
		});
	}

	const navigation = [
		{ name: 'Dashboard', href: '/dashboard', icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6' },
		{ name: 'Projects', href: '/projects', icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z' },
		{ name: 'Settings', href: '/settings', icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z' }
	];

	function isActive(path: string) {
		return $page.url.pathname === path || $page.url.pathname.startsWith(path + '/');
	}
</script>

<nav class="sticky top-0 z-50 transition-all duration-300 {scrolled ? 'shadow-2xl' : 'glass-card shadow-lg'} border-b"
     style="{scrolled ? 'background-color: rgb(var(--bg-secondary)); backdrop-filter: blur(12px);' : ''}">
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
		<div class="flex justify-between h-16">
			<!-- Logo -->
			<div class="flex items-center">
				<a href="/dashboard" class="flex items-center gap-3 group">
					<div class="relative">
						<div class="absolute inset-0 bg-gradient-brand rounded-xl blur-md opacity-50 group-hover:opacity-75 transition-opacity"></div>
						<div class="relative w-10 h-10 rounded-xl bg-gradient-brand shadow-lg group-hover:scale-110 transition-transform duration-200 glow-green flex items-center justify-center">
							<img
								src="/img/My-Icon.webp"
								alt="TSOFT Logo"
								class="w-8 h-8 rounded-lg"
							/>
						</div>
					</div>
					<div class="flex flex-col">
						<span class="text-lg font-bold bg-gradient-to-r from-primary-700 to-primary-900 bg-clip-text text-transparent">
							VPS Panel
						</span>
						<span class="text-xs font-medium" style="color: rgb(var(--text-tertiary));">
							by TSOFT
						</span>
					</div>
				</a>

				<!-- Desktop Navigation -->
				<div class="hidden sm:ml-10 sm:flex sm:space-x-2">
					{#each navigation as item}
						<a
							href={item.href}
							class="relative inline-flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-semibold transition-all duration-200 group
							       {isActive(item.href)
								? 'text-white'
								: 'hover:scale-105'}"
							style="color: {isActive(item.href) ? 'white' : 'rgb(var(--text-secondary))'}"
						>
							{#if isActive(item.href)}
								<div class="absolute inset-0 bg-gradient-brand rounded-xl shadow-lg glow-green"></div>
							{/if}
							<div class="relative flex items-center gap-2">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={item.icon} />
								</svg>
								{item.name}
							</div>
						</a>
					{/each}
				</div>
			</div>

			<!-- Right section -->
			<div class="hidden sm:flex sm:items-center sm:gap-3">
				<ThemeToggle />

				{#if authStore.user}
					<div class="flex items-center gap-3 pl-3 ml-3 border-l" style="border-color: rgb(var(--border-primary));">
						<div class="flex items-center gap-3 group cursor-pointer">
							<div class="relative">
								<div class="absolute inset-0 bg-gradient-brand rounded-xl blur opacity-50 group-hover:opacity-75 transition-opacity"></div>
								<div class="relative w-10 h-10 rounded-xl bg-gradient-brand flex items-center justify-center text-white font-bold text-sm shadow-lg group-hover:scale-110 transition-transform">
									{authStore.user.name.charAt(0).toUpperCase()}
								</div>
							</div>
							<div class="flex flex-col">
								<span class="text-sm font-bold" style="color: rgb(var(--text-primary));">
									{authStore.user.name}
								</span>
								<span class="text-xs" style="color: rgb(var(--text-tertiary));">
									Administrator
								</span>
							</div>
						</div>
						<button
							onclick={() => authStore.logout()}
							class="inline-flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-medium transition-all duration-200 hover:scale-105 border group"
							style="background-color: rgb(var(--bg-tertiary)); color: rgb(var(--text-primary)); border-color: rgb(var(--border-primary));"
						>
							<svg class="w-4 h-4 group-hover:translate-x-0.5 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
							</svg>
							Sign out
						</button>
					</div>
				{:else}
					<Button variant="primary" size="sm" onclick={() => window.location.href = '/login'} class="glow-green-hover">
						<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
						</svg>
						Sign in
					</Button>
				{/if}
			</div>

			<!-- Mobile menu button -->
			<div class="flex items-center sm:hidden gap-3">
				<ThemeToggle />
				<button
					type="button"
					onclick={() => mobileMenuOpen = !mobileMenuOpen}
					class="inline-flex items-center justify-center p-2 rounded-xl transition-all duration-200
					       hover:scale-110 border"
					style="color: rgb(var(--text-secondary)); border-color: rgb(var(--border-primary)); background-color: rgb(var(--bg-tertiary));"
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
		<div class="sm:hidden border-t slide-in-up" style="border-color: rgb(var(--border-primary)); background-color: rgb(var(--bg-secondary) / 0.95);">
			<div class="px-2 pt-2 pb-3 space-y-1">
				{#each navigation as item, i}
					<a
						href={item.href}
						class="relative flex items-center gap-3 px-4 py-3 rounded-xl text-base font-semibold transition-all duration-200 group fade-in
						       {isActive(item.href)
							? 'text-white'
							: ''}"
						style="color: {isActive(item.href) ? 'white' : 'rgb(var(--text-secondary))'}; animation-delay: {i * 0.1}s;"
					>
						{#if isActive(item.href)}
							<div class="absolute inset-0 bg-gradient-brand rounded-xl shadow-lg"></div>
						{/if}
						<div class="relative flex items-center gap-3 w-full">
							<div class="w-10 h-10 rounded-lg flex items-center justify-center {isActive(item.href) ? 'bg-white/20' : ''}" style="{!isActive(item.href) ? 'background-color: rgb(var(--bg-tertiary));' : ''}">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={item.icon} />
								</svg>
							</div>
							{item.name}
						</div>
					</a>
				{/each}
			</div>

			<div class="pt-4 pb-3 border-t" style="border-color: rgb(var(--border-primary));">
				{#if authStore.user}
					<div class="px-4 fade-in">
						<div class="flex items-center mb-4 p-3 rounded-xl" style="background-color: rgb(var(--bg-tertiary));">
							<div class="h-14 w-14 rounded-xl bg-gradient-brand flex items-center justify-center text-white font-bold text-xl shadow-lg glow-green">
								{authStore.user.name.charAt(0).toUpperCase()}
							</div>
							<div class="ml-3 flex-1">
								<div class="text-base font-bold" style="color: rgb(var(--text-primary));">{authStore.user.name}</div>
								<div class="text-sm" style="color: rgb(var(--text-secondary));">{authStore.user.email}</div>
								<div class="text-xs mt-1 inline-flex items-center px-2 py-0.5 rounded-md bg-gradient-brand text-white">
									<svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
									</svg>
									Administrator
								</div>
							</div>
						</div>
						<button
							onclick={() => authStore.logout()}
							class="w-full inline-flex items-center justify-center gap-2 px-4 py-3 rounded-xl text-sm font-semibold transition-all duration-200 hover:scale-105 border group"
							style="background-color: rgb(var(--bg-tertiary)); color: rgb(var(--text-primary)); border-color: rgb(var(--border-primary));"
						>
							<svg class="w-5 h-5 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
							</svg>
							Sign out
						</button>
					</div>
				{:else}
					<div class="px-4">
						<Button variant="primary" class="w-full glow-green-hover" onclick={() => window.location.href = '/login'}>
							<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
							</svg>
							Sign in
						</Button>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</nav>
