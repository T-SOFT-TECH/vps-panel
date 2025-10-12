<script lang="ts">
	interface Props {
		label?: string;
		type?: string;
		placeholder?: string;
		value?: string | number;
		error?: string;
		required?: boolean;
		disabled?: boolean;
		class?: string;
		id?: string;
		name?: string;
	}

	let {
		label,
		type = 'text',
		placeholder = '',
		value = $bindable(''),
		error,
		required = false,
		disabled = false,
		class: className = '',
		id,
		name
	}: Props = $props();

	const inputId = id || name || `input-${Math.random().toString(36).substr(2, 9)}`;
</script>

<div class="w-full">
	{#if label}
		<label for={inputId} class="block text-sm font-semibold mb-2" style="color: rgb(var(--text-primary));">
			{label}
			{#if required}
				<span class="text-red-500 ml-1">*</span>
			{/if}
		</label>
	{/if}
	<input
		{type}
		{placeholder}
		{required}
		{disabled}
		{name}
		id={inputId}
		bind:value
		class="w-full px-4 py-3 rounded-xl font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-primary-600 disabled:opacity-50 disabled:cursor-not-allowed {error ? 'border-2 border-red-500 focus:ring-red-500' : 'border'} {className}"
		style="background-color: rgb(var(--bg-tertiary)); border-color: {error ? '#ef4444' : 'rgb(var(--border-primary))'}; color: rgb(var(--text-primary));"
	/>
	{#if error}
		<div class="mt-2 flex items-center gap-2 text-sm text-red-500 font-medium fade-in">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
			</svg>
			{error}
		</div>
	{/if}
</div>
