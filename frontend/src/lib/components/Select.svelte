<script lang="ts">
	interface Option {
		value: string | number;
		label: string;
	}

	interface Props {
		label?: string;
		options: Option[];
		value?: string | number;
		error?: string;
		required?: boolean;
		disabled?: boolean;
		class?: string;
		id?: string;
		name?: string;
		placeholder?: string;
	}

	let {
		label,
		options,
		value = $bindable(''),
		error,
		required = false,
		disabled = false,
		class: className = '',
		id,
		name,
		placeholder = 'Select an option'
	}: Props = $props();

	const selectId = id || name || `select-${Math.random().toString(36).substr(2, 9)}`;
</script>

<div class="w-full">
	{#if label}
		<label for={selectId} class="block text-sm font-semibold mb-2" style="color: rgb(var(--text-primary));">
			{label}
			{#if required}
				<span class="text-red-500 ml-1">*</span>
			{/if}
		</label>
	{/if}
	<div class="relative">
		<select
			{required}
			{disabled}
			{name}
			id={selectId}
			bind:value
			class="w-full px-4 py-3 rounded-xl font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-primary-600 disabled:opacity-50 disabled:cursor-not-allowed appearance-none {error ? 'border-2 border-red-500 focus:ring-red-500' : 'border'} {className}"
			style="background-color: rgb(var(--bg-tertiary)); border-color: {error ? '#ef4444' : 'rgb(var(--border-primary))'}; color: rgb(var(--text-primary));"
		>
			<option value="" disabled>{placeholder}</option>
			{#each options as option}
				<option value={option.value}>{option.label}</option>
			{/each}
		</select>
		<div class="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none">
			<svg class="w-5 h-5" style="color: rgb(var(--text-tertiary));" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
			</svg>
		</div>
	</div>
	{#if error}
		<div class="mt-2 flex items-center gap-2 text-sm text-red-500 font-medium fade-in">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
			</svg>
			{error}
		</div>
	{/if}
</div>
