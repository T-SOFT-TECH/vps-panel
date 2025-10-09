<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		open?: boolean;
		title?: string;
		children: Snippet;
		onclose?: () => void;
		size?: 'sm' | 'md' | 'lg' | 'xl';
	}

	let {
		open = $bindable(false),
		title,
		children,
		onclose,
		size = 'md'
	}: Props = $props();

	const sizes = {
		sm: 'max-w-sm',
		md: 'max-w-md',
		lg: 'max-w-lg',
		xl: 'max-w-xl'
	};

	function handleClose() {
		open = false;
		onclose?.();
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			handleClose();
		}
	}
</script>

{#if open}
	<div
		class="fixed inset-0 z-50 overflow-y-auto"
		role="dialog"
		aria-modal="true"
	>
		<div
			class="flex min-h-screen items-center justify-center p-4"
			onclick={handleBackdropClick}
		>
			<!-- Backdrop -->
			<div class="fixed inset-0 bg-black bg-opacity-75 transition-opacity"></div>

			<!-- Modal -->
			<div class="relative bg-zinc-900 border border-zinc-800 rounded-lg shadow-xl {sizes[size]} w-full">
				<!-- Header -->
				{#if title}
					<div class="flex items-center justify-between p-6 border-b border-zinc-800">
						<h3 class="text-lg font-semibold text-zinc-100">{title}</h3>
						<button
							onclick={handleClose}
							class="text-zinc-400 hover:text-zinc-300 focus:outline-none"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
							</svg>
						</button>
					</div>
				{/if}

				<!-- Content -->
				<div class="p-6">
					{@render children()}
				</div>
			</div>
		</div>
	</div>
{/if}
