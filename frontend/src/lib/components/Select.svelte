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
	const selectClasses = `block w-full rounded-lg border ${error ? 'border-red-500 focus:ring-red-500' : 'border-zinc-700 focus:ring-green-500'} bg-zinc-900 text-zinc-100 px-3 py-2 focus:outline-none focus:ring-2 disabled:opacity-50 disabled:cursor-not-allowed ${className}`;
</script>

<div class="w-full">
	{#if label}
		<label for={selectId} class="block text-sm font-medium text-zinc-300 mb-1">
			{label}
			{#if required}
				<span class="text-red-500">*</span>
			{/if}
		</label>
	{/if}
	<select
		{required}
		{disabled}
		{name}
		id={selectId}
		bind:value
		class={selectClasses}
	>
		<option value="" disabled>{placeholder}</option>
		{#each options as option}
			<option value={option.value}>{option.label}</option>
		{/each}
	</select>
	{#if error}
		<p class="mt-1 text-sm text-red-600">{error}</p>
	{/if}
</div>
