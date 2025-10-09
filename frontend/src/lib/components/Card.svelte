<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		children: Snippet;
		class?: string;
		padding?: boolean;
		hover?: boolean;
	}

	let {
		children,
		class: className = '',
		padding = true,
		hover = false
	}: Props = $props();

	const classes = `bg-zinc-900 rounded-lg border border-zinc-800 shadow-sm ${padding ? 'p-6' : ''} ${hover ? 'hover:shadow-md hover:border-zinc-700 transition-all' : ''} ${className}`;
</script>

<div class={classes}>
	{@render children()}
</div>
