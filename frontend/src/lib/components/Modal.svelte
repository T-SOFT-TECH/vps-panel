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

	// Prevent scroll when modal is open
	$effect(() => {
		if (open) {
			document.body.style.overflow = 'hidden';
		} else {
			document.body.style.overflow = '';
		}

		return () => {
			document.body.style.overflow = '';
		};
	});
</script>

{#if open}
	<div
		class="fixed inset-0 z-50 overflow-y-auto fade-in"
		role="dialog"
		aria-modal="true"
	>
		<div
			class="flex min-h-screen items-center justify-center p-4"
			onclick={handleBackdropClick}
		>
			<!-- Backdrop with blur -->
			<div class="fixed inset-0 bg-black/60 backdrop-blur-sm transition-opacity"></div>

			<!-- Modal -->
			<div class="relative glass-pro rounded-2xl shadow-2xl {sizes[size]} w-full scale-up border-0">
				<!-- Animated gradient border effect -->
				<div class="absolute inset-0 rounded-2xl opacity-20 pointer-events-none">
					<div class="absolute inset-0 rounded-2xl bg-gradient-to-br from-primary-600 via-transparent to-primary-800 animate-pulse"></div>
				</div>

				<!-- Header -->
				{#if title}
					<div class="relative flex items-center justify-between p-6 border-b" style="border-color: rgb(var(--border-primary));">
						<h3 class="text-xl font-bold bg-gradient-to-r from-primary-700 to-primary-900 bg-clip-text text-transparent">{title}</h3>
						<button
							onclick={handleClose}
							class="p-2 rounded-xl transition-all duration-200 hover:scale-110 hover:rotate-90 group"
							style="color: rgb(var(--text-tertiary)); background-color: rgb(var(--bg-tertiary));"
							aria-label="Close modal"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
							</svg>
						</button>
					</div>
				{:else}
					<!-- Close button without title -->
					<button
						onclick={handleClose}
						class="absolute top-4 right-4 z-10 p-2 rounded-xl transition-all duration-200 hover:scale-110 hover:rotate-90"
						style="color: rgb(var(--text-tertiary)); background-color: rgb(var(--bg-tertiary));"
						aria-label="Close modal"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				{/if}

				<!-- Content -->
				<div class="relative p-6 slide-in-up" style="animation-delay: 0.1s;">
					{@render children()}
				</div>
			</div>
		</div>
	</div>
{/if}
