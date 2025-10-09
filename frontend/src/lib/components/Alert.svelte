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
		success: 'bg-green-950/50 text-green-400 border-green-800',
		warning: 'bg-yellow-950/50 text-yellow-400 border-yellow-800',
		error: 'bg-red-950/50 text-red-400 border-red-800',
		info: 'bg-green-950/50 text-green-400 border-green-800'
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
	<div class="rounded-lg border p-4 {variants[variant]} {className}" role="alert">
		<div class="flex items-start">
			<svg class="w-5 h-5 mr-3 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={icons[variant]} />
			</svg>
			<div class="flex-1">
				{@render children()}
			</div>
			{#if dismissible}
				<button
					onclick={handleDismiss}
					class="ml-3 inline-flex flex-shrink-0 focus:outline-none"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			{/if}
		</div>
	</div>
{/if}
