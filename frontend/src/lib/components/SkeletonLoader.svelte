<script lang="ts">
	interface Props {
		variant?: 'text' | 'card' | 'stat' | 'table' | 'list' | 'avatar' | 'image';
		count?: number;
		height?: string;
		width?: string;
		class?: string;
	}

	let {
		variant = 'text',
		count = 1,
		height,
		width,
		class: className = ''
	}: Props = $props();
</script>

<div class={className}>
	{#if variant === 'text'}
		{#each Array(count) as _, i}
			<div class="skeleton mb-3 {height ? `h-[${height}]` : 'h-4'} {width ? `w-[${width}]` : 'w-full'} fade-in" style="animation-delay: {i * 0.1}s;"></div>
		{/each}
	{:else if variant === 'card'}
		{#each Array(count) as _, i}
			<div class="modern-card p-6 mb-4 fade-in" style="animation-delay: {i * 0.1}s;">
				<div class="flex items-start gap-4">
					<div class="skeleton w-12 h-12 rounded-xl flex-shrink-0"></div>
					<div class="flex-1 space-y-3">
						<div class="skeleton h-5 w-3/4"></div>
						<div class="skeleton h-4 w-full"></div>
						<div class="skeleton h-4 w-5/6"></div>
					</div>
				</div>
			</div>
		{/each}
	{:else if variant === 'stat'}
		{#each Array(count) as _, i}
			<div class="modern-card p-6 fade-in" style="animation-delay: {i * 0.1}s;">
				<div class="flex items-center justify-between mb-4">
					<div class="skeleton w-14 h-14 rounded-xl"></div>
				</div>
				<div class="skeleton h-4 w-1/2 mb-2"></div>
				<div class="skeleton h-8 w-20"></div>
			</div>
		{/each}
	{:else if variant === 'table'}
		<div class="modern-card overflow-hidden">
			<!-- Header -->
			<div class="flex items-center gap-4 p-4 border-b" style="border-color: rgb(var(--border-primary));">
				{#each Array(4) as _}
					<div class="skeleton h-4 flex-1"></div>
				{/each}
			</div>
			<!-- Rows -->
			{#each Array(count) as _, i}
				<div class="flex items-center gap-4 p-4 border-b fade-in" style="border-color: rgb(var(--border-primary)); animation-delay: {i * 0.1}s;">
					{#each Array(4) as _}
						<div class="skeleton h-4 flex-1"></div>
					{/each}
				</div>
			{/each}
		</div>
	{:else if variant === 'list'}
		{#each Array(count) as _, i}
			<div class="flex items-center gap-3 p-4 rounded-xl border mb-3 fade-in" style="background-color: rgb(var(--bg-secondary)); border-color: rgb(var(--border-primary)); animation-delay: {i * 0.1}s;">
				<div class="skeleton w-10 h-10 rounded-lg flex-shrink-0"></div>
				<div class="flex-1 space-y-2">
					<div class="skeleton h-4 w-3/4"></div>
					<div class="skeleton h-3 w-1/2"></div>
				</div>
				<div class="skeleton w-16 h-6 rounded-full"></div>
			</div>
		{/each}
	{:else if variant === 'avatar'}
		<div class="skeleton {height ? `h-[${height}]` : 'h-12'} {width ? `w-[${width}]` : 'w-12'} rounded-full"></div>
	{:else if variant === 'image'}
		<div class="skeleton {height ? `h-[${height}]` : 'h-48'} {width ? `w-[${width}]` : 'w-full'} rounded-xl"></div>
	{/if}
</div>
