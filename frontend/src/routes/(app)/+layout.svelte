<script lang="ts">
	import { authStore } from '$lib/stores/auth.svelte';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import Navbar from '$lib/components/Navbar.svelte';

	let { children } = $props();

	onMount(() => {
		// Check if user is authenticated
		if (!authStore.isAuthenticated && !authStore.loading) {
			goto('/login');
		}
	});
</script>

{#if authStore.isAuthenticated}
	<div class="min-h-screen bg-zinc-950 dark:bg-zinc-950 light:bg-zinc-50 transition-colors duration-300">
		<Navbar />
		<main class="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
			{@render children()}
		</main>
	</div>
{:else}
	<div class="min-h-screen bg-zinc-950 dark:bg-zinc-950 light:bg-zinc-50 flex items-center justify-center transition-colors duration-300">
		<div class="text-center">
			<div class="relative w-16 h-16 mx-auto mb-4">
				<div class="absolute inset-0 rounded-full border-4 border-zinc-800 dark:border-zinc-800 light:border-zinc-200"></div>
				<div class="absolute inset-0 rounded-full border-4 border-primary-500 border-t-transparent animate-spin"></div>
			</div>
			<p class="text-zinc-400 dark:text-zinc-400 light:text-zinc-600 font-medium">Loading...</p>
		</div>
	</div>
{/if}
