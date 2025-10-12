<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		variant?: 'success' | 'warning' | 'error' | 'info' | 'default';
		children: Snippet;
		class?: string;
	}

	let {
		variant = 'default',
		children,
		class: className = ''
	}: Props = $props();

	const classes = `badge badge-${variant} ${className}`;
</script>

<span class={classes}>
	{@render children()}
</span>
