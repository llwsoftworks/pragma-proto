<script lang="ts">
	export let type: 'warning' | 'info' | 'success' | 'error' = 'info';
	export let message: string;
</script>

<div
	class="flex items-start gap-2 rounded-md px-3 py-2 text-sm"
	class:bg-amber-50={type === 'warning'}
	class:text-amber-800={type === 'warning'}
	class:bg-blue-50={type === 'info'}
	class:text-blue-800={type === 'info'}
	class:bg-green-50={type === 'success'}
	class:text-green-800={type === 'success'}
	class:bg-red-50={type === 'error'}
	class:text-red-800={type === 'error'}
	role="alert"
>
	{#if type === 'warning'}
		<span aria-hidden="true">⚠</span>
	{:else if type === 'info'}
		<span aria-hidden="true">🔔</span>
	{:else if type === 'success'}
		<span aria-hidden="true">✓</span>
	{:else}
		<span aria-hidden="true">✕</span>
	{/if}
	<span>{message}</span>
</div>
