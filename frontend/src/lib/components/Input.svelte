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
	const inputClasses = `block w-full rounded-lg border ${error ? 'border-red-500 focus:ring-red-500' : 'border-zinc-700 focus:ring-green-500'} bg-zinc-900 text-zinc-100 px-3 py-2 focus:outline-none focus:ring-2 disabled:opacity-50 disabled:cursor-not-allowed placeholder:text-zinc-500 ${className}`;
</script>

<div class="w-full">
	{#if label}
		<label for={inputId} class="block text-sm font-medium text-zinc-300 mb-1">
			{label}
			{#if required}
				<span class="text-red-500">*</span>
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
		class={inputClasses}
	/>
	{#if error}
		<p class="mt-1 text-sm text-red-600">{error}</p>
	{/if}
</div>
