<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		variant?: 'primary' | 'secondary' | 'danger' | 'ghost' | 'outline';
		size?: 'sm' | 'md' | 'lg';
		type?: 'button' | 'submit' | 'reset';
		disabled?: boolean;
		loading?: boolean;
		onclick?: () => void;
		children: Snippet;
		class?: string;
	}

	let {
		variant = 'primary',
		size = 'md',
		type = 'button',
		disabled = false,
		loading = false,
		onclick,
		children,
		class: className = ''
	}: Props = $props();

	const baseStyles = 'inline-flex items-center justify-center font-semibold rounded-xl transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed relative overflow-hidden group';

	const variants = {
		primary: 'bg-gradient-brand text-white hover:shadow-lg hover:scale-105 focus:ring-primary-600 shadow-md glow-green-hover',
		secondary: 'text-white hover:scale-105 focus:ring-zinc-500 border shadow-md',
		danger: 'bg-gradient-to-br from-red-600 to-red-700 text-white hover:shadow-lg hover:scale-105 focus:ring-red-500 shadow-md',
		ghost: 'hover:scale-105 focus:ring-zinc-500',
		outline: 'border hover:scale-105 focus:ring-primary-600 hover:border-primary-600'
	};

	const sizes = {
		sm: 'px-3 py-1.5 text-sm',
		md: 'px-4 py-2 text-base',
		lg: 'px-6 py-3 text-lg'
	};

	const classes = `${baseStyles} ${variants[variant]} ${sizes[size]} ${className}`;
</script>

<button
	{type}
	class={classes}
	disabled={disabled || loading}
	onclick={onclick}
	style="{variant === 'secondary' ? 'background-color: rgb(var(--bg-tertiary)); color: rgb(var(--text-primary)); border-color: rgb(var(--border-primary));' : ''}{variant === 'ghost' ? 'color: rgb(var(--text-secondary)); background-color: rgb(var(--bg-tertiary));' : ''}{variant === 'outline' ? 'background-color: transparent; color: rgb(var(--text-primary)); border-color: rgb(var(--border-primary));' : ''}"
>
	{#if !disabled && variant === 'primary'}
		<div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover:translate-x-full transition-transform duration-1000"></div>
	{/if}

	<span class="relative flex items-center gap-2">
		{#if loading}
			<svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
				<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
			</svg>
		{/if}
		{@render children()}
	</span>
</button>
