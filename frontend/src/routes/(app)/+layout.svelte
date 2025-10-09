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
	<div class="min-h-screen bg-zinc-950">
		<Navbar />
		<main class="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
			{@render children()}
		</main>
	</div>
{:else}
	<div class="min-h-screen bg-zinc-950 flex items-center justify-center">
		<div class="text-center">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-green-500 mx-auto"></div>
			<p class="mt-4 text-zinc-400">Loading...</p>
		</div>
	</div>
{/if}
