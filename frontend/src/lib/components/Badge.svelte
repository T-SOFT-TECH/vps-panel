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

	const variants = {
		success: 'bg-green-950/50 text-green-400 border-green-800',
		warning: 'bg-yellow-950/50 text-yellow-400 border-yellow-800',
		error: 'bg-red-950/50 text-red-400 border-red-800',
		info: 'bg-green-950/50 text-green-400 border-green-800',
		default: 'bg-zinc-800 text-zinc-300 border-zinc-700'
	};

	const classes = `inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${variants[variant]} ${className}`;
</script>

<span class={classes}>
	{@render children()}
</span>
