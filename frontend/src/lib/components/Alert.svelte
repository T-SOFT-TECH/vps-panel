<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		variant?: 'success' | 'warning' | 'error' | 'info';
		children: Snippet;
		dismissible?: boolean;
		ondismiss?: () => void;
		class?: string;
	}

	let {
		variant = 'info',
		children,
		dismissible = false,
		ondismiss,
		class: className = ''
	}: Props = $props();

	let visible = $state(true);

	const variants = {
		success: 'bg-primary-950/30 text-primary-400 border-primary-800/50',
		warning: 'bg-yellow-950/30 text-yellow-400 border-yellow-800/50',
		error: 'bg-red-950/30 text-red-400 border-red-800/50',
		info: 'bg-blue-950/30 text-blue-400 border-blue-800/50'
	};

	const icons = {
		success: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
		warning: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z',
		error: 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z',
		info: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z'
	};

	function handleDismiss() {
		visible = false;
		ondismiss?.();
	}
</script>

{#if visible}
	<div class="rounded-xl border p-4 backdrop-blur-sm shadow-lg slide-in-up {variants[variant]} {className}" role="alert">
		<div class="flex items-start gap-3">
			<div class="flex-shrink-0 w-10 h-10 rounded-xl flex items-center justify-center {variant === 'success' ? 'bg-primary-800/30' : variant === 'warning' ? 'bg-yellow-800/30' : variant === 'error' ? 'bg-red-800/30' : 'bg-blue-800/30'}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={icons[variant]} />
				</svg>
			</div>
			<div class="flex-1 pt-1.5">
				{@render children()}
			</div>
			{#if dismissible}
				<button
					onclick={handleDismiss}
					class="flex-shrink-0 p-1.5 rounded-lg transition-all duration-200 hover:scale-110 hover:rotate-90"
					style="color: rgb(var(--text-tertiary)); background-color: rgb(var(--bg-tertiary) / 0.5);"
					aria-label="Dismiss alert"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			{/if}
		</div>
	</div>
{/if}
