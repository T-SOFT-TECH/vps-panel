<script lang="ts">
	import { themeStore } from '$lib/stores/theme.svelte';
	import { fly, scale } from 'svelte/transition';

	let isToggling = $state(false);

	function toggle() {
		if (isToggling) return;
		isToggling = true;
		themeStore.toggle();
		setTimeout(() => {
			isToggling = false;
		}, 300);
	}
</script>

<button
	onclick={toggle}
	class="relative inline-flex items-center justify-center w-12 h-12 rounded-xl bg-gradient-to-br from-zinc-800 to-zinc-900 hover:from-zinc-700 hover:to-zinc-800
	       dark:from-zinc-800 dark:to-zinc-900
	       light:from-white light:to-zinc-50
	       border border-zinc-700 dark:border-zinc-700 light:border-zinc-200
	       shadow-lg hover:shadow-xl transform hover:scale-105 transition-all duration-200
	       focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 focus:ring-offset-zinc-950 dark:focus:ring-offset-zinc-950 light:focus:ring-offset-white"
	aria-label="Toggle theme"
	disabled={isToggling}
>
	<!-- Sun Icon (Light Mode) -->
	{#if themeStore.current === 'light'}
		<svg
			in:scale={{ duration: 200, start: 0.5 }}
			out:scale={{ duration: 200, start: 0.5 }}
			class="w-6 h-6 text-yellow-500"
			fill="none"
			stroke="currentColor"
			viewBox="0 0 24 24"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
			/>
		</svg>
	{:else}
		<!-- Moon Icon (Dark Mode) -->
		<svg
			in:scale={{ duration: 200, start: 0.5 }}
			out:scale={{ duration: 200, start: 0.5 }}
			class="w-6 h-6 text-primary-400"
			fill="none"
			stroke="currentColor"
			viewBox="0 0 24 24"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
			/>
		</svg>
	{/if}

	<!-- Glow Effect -->
	<div
		class="absolute inset-0 rounded-xl bg-gradient-to-br from-primary-500/20 to-primary-600/20 opacity-0 group-hover:opacity-100 blur transition-opacity duration-300"
	></div>
</button>
