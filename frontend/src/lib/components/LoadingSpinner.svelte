<script lang="ts">
	interface Props {
		size?: 'sm' | 'md' | 'lg' | 'xl';
		variant?: 'primary' | 'white' | 'secondary';
		text?: string;
		fullscreen?: boolean;
	}

	let {
		size = 'md',
		variant = 'primary',
		text,
		fullscreen = false
	}: Props = $props();

	const sizes = {
		sm: 'w-4 h-4',
		md: 'w-8 h-8',
		lg: 'w-12 h-12',
		xl: 'w-16 h-16'
	};

	const variants = {
		primary: 'border-primary-800',
		white: 'border-white',
		secondary: 'border-zinc-500'
	};
</script>

{#if fullscreen}
	<div class="fixed inset-0 z-50 flex items-center justify-center" style="background-color: rgb(var(--bg-primary));">
		<div class="text-center scale-up">
			<!-- Animated gradient circle -->
			<div class="relative {sizes[size === 'sm' ? 'lg' : 'xl']} mx-auto mb-6">
				<div class="absolute inset-0 rounded-full border-4 border-primary-200"></div>
				<div class="absolute inset-0 rounded-full border-4 border-primary-800 border-t-transparent animate-spin"></div>
				<div class="absolute inset-0 flex items-center justify-center">
					<div class="w-1/2 h-1/2 bg-gradient-brand rounded-full pulse"></div>
				</div>
			</div>

			{#if text}
				<p class="text-lg font-semibold mb-2" style="color: rgb(var(--text-primary));">{text}</p>
				<p class="text-sm" style="color: rgb(var(--text-secondary));">Please wait a moment</p>
			{:else}
				<p class="text-lg font-semibold mb-2" style="color: rgb(var(--text-primary));">Loading...</p>
				<p class="text-sm" style="color: rgb(var(--text-secondary));">Please wait a moment</p>
			{/if}

			<!-- Animated dots -->
			<div class="flex items-center justify-center gap-2 mt-6">
				<div class="w-2 h-2 bg-primary-800 rounded-full animate-bounce" style="animation-delay: 0s;"></div>
				<div class="w-2 h-2 bg-primary-700 rounded-full animate-bounce" style="animation-delay: 0.2s;"></div>
				<div class="w-2 h-2 bg-primary-600 rounded-full animate-bounce" style="animation-delay: 0.4s;"></div>
			</div>
		</div>
	</div>
{:else}
	<div class="inline-flex flex-col items-center gap-3">
		<div class="relative {sizes[size]}">
			<div class="absolute inset-0 rounded-full border-4" style="border-color: rgb(var(--border-primary));"></div>
			<div class="absolute inset-0 rounded-full border-4 {variants[variant]} border-t-transparent animate-spin"></div>
		</div>
		{#if text}
			<p class="text-sm font-medium" style="color: rgb(var(--text-secondary));">{text}</p>
		{/if}
	</div>
{/if}
