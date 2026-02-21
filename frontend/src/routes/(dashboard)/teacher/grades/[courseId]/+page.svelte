<script lang="ts">
	/**
	 * Full grade grid for one course — spec §5.1 teacher/grades/[courseId].
	 * Inline editing, keyboard navigation (Tab, Enter, arrow keys).
	 * Auto-saves with optimistic UI per spec §9.1.
	 */
	import type { PageData } from './$types';
	import GradeGrid from '$lib/components/GradeGrid.svelte';
	import { notifications } from '$lib/stores/notifications';

	export let data: PageData;

	interface GradeRecord {
		assignment_id: string;
		student_id: string;
		points_earned: number | null;
		is_excused: boolean;
		is_missing: boolean;
		is_late: boolean;
		ai_suggested: number | null;
	}

	// Build a Map keyed by "studentId:assignmentId" for the GradeGrid component.
	$: gradeMap = new Map(
		(data.grades as GradeRecord[]).map((g) => [
			`${g.student_id}:${g.assignment_id}`,
			{
				assignment_id: g.assignment_id,
				student_id: g.student_id,
				points_earned: g.points_earned,
				is_excused: g.is_excused ?? false,
				is_missing: g.is_missing ?? false,
				is_late: g.is_late ?? false,
				ai_suggested: g.ai_suggested ?? null
			}
		])
	);

	async function handleGradeChange(e: CustomEvent<GradeRecord>) {
		const cell = e.detail;
		try {
			const form = new FormData();
			form.set('assignment_id', cell.assignment_id);
			form.set('student_id', cell.student_id);
			form.set('points_earned', cell.points_earned?.toString() ?? '');
			form.set('is_excused', String(cell.is_excused));
			form.set('is_missing', String(cell.is_missing));
			form.set('is_late', String(cell.is_late));

			const res = await fetch('?/saveGrade', {
				method: 'POST',
				body: form
			});

			if (!res.ok) {
				notifications.error('Failed to save grade. Please try again.');
			}
		} catch {
			notifications.error('Failed to save grade. Please try again.');
		}
	}
</script>

<svelte:head>
	<title>{data.course?.name ?? 'Course'} — Grades — Pragma</title>
</svelte:head>

<div class="mx-auto max-w-7xl px-4 py-6">
	<div class="mb-4 flex items-center gap-3">
		<a
			href="/teacher/grades"
			class="rounded-md px-2 py-1 text-sm text-muted-foreground hover:bg-muted"
			aria-label="Back to courses"
		>
			← Courses
		</a>
		<div>
			<h1 class="text-xl font-bold">{data.course?.name ?? 'Course Grades'}</h1>
			{#if data.course?.subject}
				<p class="text-sm text-muted-foreground">{data.course.subject}{data.course.period ? ` · ${data.course.period}` : ''}</p>
			{/if}
		</div>
	</div>

	{#if data.students.length === 0}
		<div class="rounded-lg border border-dashed border-border p-8 text-center text-muted-foreground">
			<p>No students enrolled in this course.</p>
		</div>
	{:else if data.assignments.length === 0}
		<div class="rounded-lg border border-dashed border-border p-8 text-center text-muted-foreground">
			<p>No assignments yet. <a href="/teacher/assignments/new" class="text-primary hover:underline">Create your first one →</a></p>
		</div>
	{:else}
		<GradeGrid
			students={data.students}
			assignments={data.assignments}
			{gradeMap}
			on:gradeChange={handleGradeChange}
		/>
	{/if}
</div>
