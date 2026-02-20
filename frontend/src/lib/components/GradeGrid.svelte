<script lang="ts">
	/**
	 * GradeGrid — students × assignments matrix.
	 * Keyboard-navigable per spec §9.1: Tab between cells, arrow keys to navigate.
	 * Inline editing with auto-save (optimistic UI).
	 */
	import { createEventDispatcher } from 'svelte';
	import GradeInput from './GradeInput.svelte';
	import { formatDate } from '$lib/utils';

	export interface Student {
		id: string;
		first_name: string;
		last_name: string;
		student_number: string;
	}

	export interface AssignmentCol {
		id: string;
		title: string;
		max_points: number;
		category: string;
		due_date: string | null;
	}

	export interface GradeCell {
		assignment_id: string;
		student_id: string;
		points_earned: number | null;
		is_excused: boolean;
		is_missing: boolean;
		is_late: boolean;
		ai_suggested: number | null;
	}

	export let students: Student[] = [];
	export let assignments: AssignmentCol[] = [];
	export let gradeMap: Map<string, GradeCell> = new Map(); // key: `${studentId}:${assignmentId}`
	export let readonly = false;

	const dispatch = createEventDispatcher<{
		gradeChange: GradeCell;
	}>();

	function gradeKey(studentId: string, assignmentId: string) {
		return `${studentId}:${assignmentId}`;
	}

	function handleSave(
		studentId: string,
		assignmentId: string,
		e: CustomEvent<{ value: number | null; isExcused: boolean; isMissing: boolean; isLate: boolean }>
	) {
		const cell: GradeCell = {
			assignment_id: assignmentId,
			student_id: studentId,
			points_earned: e.detail.value,
			is_excused: e.detail.isExcused,
			is_missing: e.detail.isMissing,
			is_late: e.detail.isLate,
			ai_suggested: null
		};
		gradeMap.set(gradeKey(studentId, assignmentId), cell);
		gradeMap = gradeMap; // trigger reactivity
		dispatch('gradeChange', cell);
	}
</script>

<div class="w-full overflow-x-auto">
	<table
		class="min-w-full border-collapse text-sm"
		role="grid"
		aria-label="Grade entry grid"
	>
		<thead>
			<tr class="border-b border-border bg-muted/50">
				<th class="sticky left-0 z-10 min-w-[180px] bg-muted/50 px-3 py-2 text-left font-semibold">
					Student
				</th>
				{#each assignments as a (a.id)}
					<th class="min-w-[100px] px-2 py-2 text-center font-medium">
						<div class="text-xs font-semibold">{a.title}</div>
						<div class="text-[10px] text-muted-foreground">{a.category} · {a.max_points}pts</div>
						{#if a.due_date}
							<div class="text-[10px] text-muted-foreground">{formatDate(a.due_date)}</div>
						{/if}
					</th>
				{/each}
			</tr>
		</thead>
		<tbody>
			{#each students as student (student.id)}
				<tr class="border-b border-border hover:bg-muted/30">
					<td class="sticky left-0 bg-background px-3 py-1.5">
						<div class="font-medium">{student.last_name}, {student.first_name}</div>
						<div class="text-[10px] text-muted-foreground">{student.student_number}</div>
					</td>
					{#each assignments as a (a.id)}
						{@const key = gradeKey(student.id, a.id)}
						{@const cell = gradeMap.get(key)}
						<td class="px-1 py-0.5 text-center">
							<GradeInput
								value={cell?.points_earned ?? null}
								maxPoints={a.max_points}
								isExcused={cell?.is_excused ?? false}
								isMissing={cell?.is_missing ?? false}
								isLate={cell?.is_late ?? false}
								aiSuggested={cell?.ai_suggested ?? null}
								{readonly}
								on:save={(e) => handleSave(student.id, a.id, e)}
							/>
						</td>
					{/each}
				</tr>
			{/each}
		</tbody>
	</table>
</div>
