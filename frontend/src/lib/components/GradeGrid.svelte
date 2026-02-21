<script lang="ts">
	/**
	 * GradeGrid — students x assignments matrix.
	 *
	 * Desktop/tablet (md+): spreadsheet table with keyboard navigation.
	 *   Tab between cells, arrow keys to navigate, Enter to save.
	 *   Inline editing with auto-save (optimistic UI) per spec section 9.1.
	 *
	 * Mobile (<md): read-only card list per spec section 9.1:
	 *   "Grade entry is desktop/tablet only to prevent accidental edits."
	 *   Each student is a card with their assignments listed vertically.
	 *   No horizontal overflow. Touch-friendly.
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
	export let gradeMap: Map<string, GradeCell> = new Map();
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
		gradeMap = gradeMap;
		dispatch('gradeChange', cell);
	}

	/** Mobile display value for a grade cell. */
	function mobileDisplay(cell: GradeCell | undefined, maxPoints: number): string {
		if (!cell) return '—';
		if (cell.is_excused) return 'EX';
		if (cell.is_missing) return 'MIS';
		if (cell.points_earned != null) return `${cell.points_earned}/${maxPoints}`;
		return '—';
	}

	/** Percentage string for mobile display. */
	function mobilePercent(cell: GradeCell | undefined, maxPoints: number): string {
		if (!cell || cell.points_earned == null || maxPoints <= 0) return '';
		return `${((cell.points_earned / maxPoints) * 100).toFixed(0)}%`;
	}

	/** CSS class for mobile grade color coding. */
	function mobileGradeClass(cell: GradeCell | undefined, maxPoints: number): string {
		if (!cell) return 'text-muted-foreground';
		if (cell.is_excused) return 'text-muted-foreground';
		if (cell.is_missing) return 'text-amber-600 dark:text-amber-400';
		if (cell.is_late) return 'text-amber-600 dark:text-amber-400';
		if (cell.points_earned == null) return 'text-muted-foreground';
		const pct = maxPoints > 0 ? (cell.points_earned / maxPoints) * 100 : 0;
		if (pct >= 90) return 'text-green-600 dark:text-green-400';
		if (pct >= 80) return 'text-blue-600 dark:text-blue-400';
		if (pct >= 70) return 'text-yellow-600 dark:text-yellow-400';
		if (pct >= 60) return 'text-orange-600 dark:text-orange-400';
		return 'text-red-600 dark:text-red-400';
	}

	// Track which student cards are expanded on mobile.
	let expandedStudents: Set<string> = new Set();

	function toggleStudent(id: string) {
		if (expandedStudents.has(id)) {
			expandedStudents.delete(id);
		} else {
			expandedStudents.add(id);
		}
		expandedStudents = expandedStudents;
	}
</script>

<!-- ============================================================
     MOBILE: card-based read-only list (<md)
     Per spec section 9.1: "Grade entry is desktop/tablet only."
     ============================================================ -->
<div class="md:hidden">
	{#if students.length === 0}
		<p class="py-4 text-center text-sm text-muted-foreground">No students.</p>
	{:else}
		<p class="mb-3 text-xs text-muted-foreground">
			Viewing grades (read-only). Use a tablet or desktop to edit.
		</p>
		<div class="flex flex-col gap-2">
			{#each students as student (student.id)}
				{@const isExpanded = expandedStudents.has(student.id)}
				<div class="rounded-lg border border-border bg-card">
					<!-- Student header — always visible, tap to expand -->
					<button
						class="flex w-full items-center justify-between px-3 py-2.5 text-left"
						on:click={() => toggleStudent(student.id)}
						aria-expanded={isExpanded}
						aria-controls="mobile-grades-{student.id}"
					>
						<div class="min-w-0 flex-1">
							<div class="truncate text-sm font-semibold">
								{student.last_name}, {student.first_name}
							</div>
							<div class="text-xs text-muted-foreground">{student.student_number}</div>
						</div>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="ml-2 h-4 w-4 shrink-0 text-muted-foreground transition-transform duration-150"
							class:rotate-180={isExpanded}
							viewBox="0 0 20 20"
							fill="currentColor"
							aria-hidden="true"
						>
							<path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clip-rule="evenodd" />
						</svg>
					</button>

					<!-- Assignment grades — shown when expanded -->
					{#if isExpanded}
						<div
							id="mobile-grades-{student.id}"
							class="border-t border-border"
						>
							{#if assignments.length === 0}
								<p class="px-3 py-3 text-xs text-muted-foreground">No assignments.</p>
							{:else}
								{#each assignments as a, idx (a.id)}
									{@const key = gradeKey(student.id, a.id)}
									{@const cell = gradeMap.get(key)}
									<div
										class="flex items-center justify-between gap-2 px-3 py-2"
										class:border-t={idx > 0}
										class:border-border={idx > 0}
									>
										<div class="min-w-0 flex-1">
											<div class="truncate text-sm">{a.title}</div>
											<div class="text-xs text-muted-foreground">
												{a.category} · {a.max_points}pts{#if a.due_date}&nbsp;· {formatDate(a.due_date)}{/if}
											</div>
										</div>
										<div class="flex shrink-0 items-center gap-1.5 text-right">
											<span class="text-sm font-mono tabular-nums font-medium {mobileGradeClass(cell, a.max_points)}">
												{mobileDisplay(cell, a.max_points)}
											</span>
											{#if mobilePercent(cell, a.max_points)}
												<span class="text-xs text-muted-foreground tabular-nums">
													{mobilePercent(cell, a.max_points)}
												</span>
											{/if}
											{#if cell?.ai_suggested != null && cell.points_earned == null}
												<span class="rounded bg-blue-100 px-1 text-[9px] text-blue-700 dark:bg-blue-900 dark:text-blue-200">
													AI
												</span>
											{/if}
										</div>
									</div>
								{/each}
							{/if}
						</div>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- ============================================================
     DESKTOP / TABLET: spreadsheet table (md+)
     Full inline editing, keyboard navigation.
     ============================================================ -->
<div class="hidden md:block">
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
						<td class="sticky left-0 z-10 bg-background px-3 py-1.5">
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
</div>
