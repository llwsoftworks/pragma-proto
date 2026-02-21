<script lang="ts">
	/**
	 * GradeInput — single inline grade entry cell.
	 * Keyboard-first: Tab to move, Enter to save, Escape to cancel.
	 * Auto-saves with optimistic UI per spec §9.1.
	 */
	import { createEventDispatcher } from 'svelte';
	import { clamp } from '$lib/utils';

	export let value: number | null = null;
	export let maxPoints: number;
	export let isExcused = false;
	export let isMissing = false;
	export let isLate = false;
	export let readonly = false;
	export let aiSuggested: number | null = null;

	const dispatch = createEventDispatcher<{
		save: { value: number | null; isExcused: boolean; isMissing: boolean; isLate: boolean };
	}>();

	let editing = false;
	let inputValue = value?.toString() ?? '';
	let inputEl: HTMLInputElement;
	let saved = false;

	function startEdit() {
		if (readonly) return;
		editing = true;
		inputValue = value?.toString() ?? '';
		requestAnimationFrame(() => {
			inputEl?.select();
		});
	}

	function save() {
		editing = false;
		const parsed = inputValue.trim() === '' ? null : Number(inputValue);
		if (parsed !== null && (isNaN(parsed) || parsed < 0 || parsed > maxPoints)) {
			// Revert to previous value.
			inputValue = value?.toString() ?? '';
			return;
		}
		value = parsed;

		// Optimistic save flash.
		saved = true;
		setTimeout(() => (saved = false), 600);

		dispatch('save', { value, isExcused, isMissing, isLate });
	}

	function cancel() {
		editing = false;
		inputValue = value?.toString() ?? '';
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			save();
		} else if (e.key === 'Escape') {
			cancel();
		}
		// Tab is handled by the browser naturally.
	}

	$: displayValue = isExcused ? 'EX' : isMissing ? 'MIS' : value != null ? value.toString() : '—';
	$: percentStr = value != null && maxPoints > 0
		? `${((value / maxPoints) * 100).toFixed(1)}%`
		: '';
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<div
	class="grade-cell relative"
	class:grade-saved={saved}
	role="gridcell"
>
	{#if editing}
		<input
			bind:this={inputEl}
			bind:value={inputValue}
			type="number"
			min="0"
			max={maxPoints}
			step="0.5"
			class="w-16 rounded-sm border border-primary bg-background px-1 py-0.5 text-sm font-mono tabular-nums focus:outline-none"
			on:blur={save}
			on:keydown={handleKeydown}
			aria-label="Grade entry (max {maxPoints})"
		/>
	{:else}
		<button
			class="min-w-[3rem] rounded-sm px-2 py-1 text-sm font-mono tabular-nums hover:bg-muted focus:bg-muted focus:outline-none disabled:cursor-default"
			class:text-muted-foreground={isExcused || (value == null && !isMissing)}
			class:text-amber-600={!isExcused && (isMissing || isLate)}
			disabled={readonly}
			on:click={startEdit}
			title={percentStr}
			aria-label="Grade: {displayValue} of {maxPoints}"
		>
			{displayValue}
		</button>
	{/if}

	{#if aiSuggested != null && value == null && !editing}
		<span
			class="absolute -top-1 right-0 rounded bg-blue-100 px-1 text-[9px] text-blue-700 dark:bg-blue-900 dark:text-blue-200"
			title="AI suggested: {aiSuggested}"
		>
			AI
		</span>
	{/if}
</div>
