<script lang="ts">
	import type { PageData, ActionData } from './$types';
	import { enhance } from '$app/forms';

	export let data: PageData;
	export let form: ActionData;

	let lockReason = '';
	let selectedStudent = '';

	$: locked = data.students.filter((s) => s.is_grade_locked);
	$: unlocked = data.students.filter((s) => !s.is_grade_locked);
</script>

<svelte:head><title>Grade Locks — Pragma</title></svelte:head>

<div class="mx-auto max-w-4xl px-4 py-6">
	<h1 class="mb-6 text-xl font-bold">Grade Access Locks</h1>

	{#if form?.error}
		<div class="mb-4 rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{form.error}</div>
	{/if}
	{#if form?.success}
		<div class="mb-4 rounded-md bg-green-50 px-4 py-3 text-sm text-green-800 dark:bg-green-950/40 dark:text-green-400">
			Grade access {form.action} successfully.
		</div>
	{/if}

	<!-- Lock a student -->
	<section class="mb-8 rounded-lg border border-border bg-card p-4">
		<h2 class="mb-3 font-semibold">Lock Grade Access</h2>
		<form method="POST" action="?/lock" use:enhance class="flex flex-col gap-3 sm:flex-row sm:items-end">
			<div class="flex-1">
				<label for="student_id" class="mb-1 block text-sm font-medium">Student</label>
				<select
					id="student_id"
					name="student_id"
					required
					class="w-full rounded-sm border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
				>
					<option value="">Select a student…</option>
					{#each unlocked as student (student.id)}
						<option value={student.id}>
							{student.last_name}, {student.first_name} ({student.student_number})
						</option>
					{/each}
				</select>
			</div>
			<div class="flex-1">
				<label for="reason" class="mb-1 block text-sm font-medium">Reason (internal only)</label>
				<input
					id="reason"
					name="reason"
					type="text"
					required
					maxlength="500"
					placeholder="e.g. Outstanding tuition — January 2026"
					class="w-full rounded-sm border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
				/>
			</div>
			<button
				type="submit"
				class="rounded-md bg-destructive px-4 py-2 text-sm font-semibold text-destructive-foreground hover:opacity-90"
			>
				Lock Access
			</button>
		</form>
	</section>

	<!-- Currently locked students -->
	{#if locked.length > 0}
		<section>
			<h2 class="mb-3 font-semibold">Currently Locked ({locked.length})</h2>
			<ul class="space-y-2">
				{#each locked as student (student.id)}
					<li class="flex items-center justify-between rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 dark:border-amber-800 dark:bg-amber-950/20">
						<div>
							<span class="font-medium">{student.last_name}, {student.first_name}</span>
							<span class="ml-2 text-sm text-muted-foreground">#{student.student_number} · {student.grade_level}</span>
							{#if student.lock_reason}
								<div class="mt-0.5 text-xs text-muted-foreground">Reason: {student.lock_reason}</div>
							{/if}
						</div>
						<form method="POST" action="?/unlock" use:enhance>
							<input type="hidden" name="student_id" value={student.id} />
							<button
								type="submit"
								class="rounded-md border border-border bg-background px-3 py-1.5 text-sm hover:bg-muted"
							>
								Unlock
							</button>
						</form>
					</li>
				{/each}
			</ul>
		</section>
	{:else}
		<p class="text-sm text-muted-foreground">No students currently have locked grade access.</p>
	{/if}
</div>
