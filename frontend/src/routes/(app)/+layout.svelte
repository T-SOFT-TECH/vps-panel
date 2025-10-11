<script lang="ts">
	import { authStore } from '$lib/stores/auth.svelte';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import Navbar from '$lib/components/Navbar.svelte';

	let { children } = $props();

	onMount(() => {
		if (!authStore.isAuthenticated && !authStore.loading) {
			goto('/login');
		}
	});
</script>

{#if authStore.isAuthenticated}
	<div class="min-h-screen transition-colors duration-300" style="background-color: rgb(var(--bg-primary));">
		<Navbar />
		<main class="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
			{@render children()}
		</main>
	</div>
{:else}
	<div class="min-h-screen flex items-center justify-center transition-colors duration-300" style="background-color: rgb(var(--bg-primary));">
		<div class="text-center">
			<div class="relative w-16 h-16 mx-auto mb-4">
				<div class="absolute inset-0 rounded-full border-4" style="border-color: rgb(var(--border-primary));"></div>
				<div class="absolute inset-0 rounded-full border-4 border-primary-800 border-t-transparent animate-spin"></div>
			</div>
			<p class="font-medium" style="color: rgb(var(--text-secondary));">Loading...</p>
		</div>
	</div>
{/if}
