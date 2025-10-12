<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		children: Snippet;
		class?: string;
		padding?: boolean;
		hover?: boolean;
		variant?: 'default' | 'glass' | 'elevated';
	}

	let {
		children,
		class: className = '',
		padding = true,
		hover = false,
		variant = 'default'
	}: Props = $props();

	const variants = {
		default: 'modern-card',
		glass: 'glass-card',
		elevated: 'elevated-card'
	};

	const classes = `${variants[variant]} ${padding ? 'p-6' : ''} ${hover ? 'hover-lift card-highlight cursor-pointer' : ''} ${className}`;
</script>

<div class={classes}>
	{@render children()}
</div>
