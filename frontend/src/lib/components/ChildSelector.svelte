<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	export interface Child {
		student_id: string;
		first_name: string;
		last_name: string;
		grade_level: string;
		is_grade_locked: boolean;
	}

	export let children: Child[] = [];
	export let selectedId: string = children[0]?.student_id ?? '';

	const dispatch = createEventDispatcher<{ select: { child: Child } }>();

	function select(child: Child) {
		selectedId = child.student_id;
		dispatch('select', { child });
	}
</script>

{#if children.length > 1}
	<div class="flex gap-2 border-b border-border pb-2" role="tablist" aria-label="Select child">
		{#each children as child (child.student_id)}
			<button
				role="tab"
				aria-selected={selectedId === child.student_id}
				class="rounded-md px-3 py-1.5 text-sm font-medium transition-colors focus:outline-none"
				class:bg-primary={selectedId === child.student_id}
				class:text-primary-foreground={selectedId === child.student_id}
				class:text-muted-foreground={selectedId !== child.student_id}
				class:hover:bg-muted={selectedId !== child.student_id}
				on:click={() => select(child)}
			>
				{child.first_name} {child.last_name}
				<span class="ml-1 text-xs opacity-70">({child.grade_level})</span>
				{#if child.is_grade_locked}
					<span class="ml-1 text-amber-500" title="Grade access restricted">🔒</span>
				{/if}
			</button>
		{/each}
	</div>
{/if}
