<script lang="ts">
	import { themeStore } from '$lib/stores/theme.svelte';

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
	class="relative inline-flex items-center justify-center w-10 h-10 rounded-lg
	       transition-all duration-200 hover:scale-105
	       bg-gradient-brand shadow-md hover:shadow-lg"
	aria-label="Toggle theme"
	disabled={isToggling}
	title={themeStore.current === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
>
	{#if themeStore.current === 'light'}
		<!-- Sun Icon -->
		<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
			/>
		</svg>
	{:else}
		<!-- Moon Icon -->
		<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
			/>
		</svg>
	{/if}
</button>
