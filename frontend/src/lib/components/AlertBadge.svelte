<script lang="ts">
	export let type: 'warning' | 'info' | 'success' | 'error' = 'info';
	export let message: string;

	const styleMap: Record<string, string> = {
		warning: 'bg-amber-50 text-amber-800 dark:bg-amber-950/40 dark:text-amber-400',
		info: 'bg-blue-50 text-blue-800 dark:bg-blue-950/40 dark:text-blue-400',
		success: 'bg-green-50 text-green-800 dark:bg-green-950/40 dark:text-green-400',
		error: 'bg-red-50 text-red-800 dark:bg-red-950/40 dark:text-red-400'
	};

	$: colorClasses = styleMap[type] ?? styleMap.info;
</script>

<div
	class="flex items-start gap-2 rounded-md px-3 py-2 text-sm {colorClasses}"
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
