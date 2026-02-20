<script lang="ts">
	import type { PageData, ActionData } from './$types';
	import { enhance } from '$app/forms';
	import FileUpload from '$lib/components/FileUpload.svelte';

	export let data: PageData;
	export let form: ActionData;

	let loading = false;
	let createdAssignmentId: string | null = null;

	const categories = [
		{ value: 'homework', label: 'Homework' },
		{ value: 'quiz', label: 'Quiz' },
		{ value: 'test', label: 'Test' },
		{ value: 'exam', label: 'Exam' },
		{ value: 'project', label: 'Project' },
		{ value: 'classwork', label: 'Classwork' },
		{ value: 'participation', label: 'Participation' },
		{ value: 'other', label: 'Other' }
	];
</script>

<svelte:head>
	<title>New Assignment — Pragma</title>
</svelte:head>

<div class="mx-auto max-w-2xl px-4 py-6">
	<h1 class="mb-6 text-xl font-bold">Create Assignment</h1>

	{#if form?.error}
		<div class="mb-4 rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{form.error}</div>
	{/if}

	<form
		method="POST"
		class="space-y-4"
		use:enhance={() => {
			loading = true;
			return async ({ update }) => {
				await update();
				loading = false;
			};
		}}
	>
		<div>
			<label for="course_id" class="mb-1 block text-sm font-medium">Course *</label>
			<select
				id="course_id"
				name="course_id"
				required
				class="w-full rounded-sm border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
			>
				<option value="">Select a course…</option>
				{#each data.courses as course (course.id)}
					<option value={course.id}>{course.name}</option>
				{/each}
			</select>
		</div>

		<div>
			<label for="title" class="mb-1 block text-sm font-medium">Title *</label>
			<input
				id="title"
				name="title"
				type="text"
				required
				maxlength="300"
				class="w-full rounded-sm border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
				placeholder="e.g. Chapter 5 Quiz"
			/>
		</div>

		<div>
			<label for="description" class="mb-1 block text-sm font-medium">Description</label>
			<textarea
				id="description"
				name="description"
				rows="3"
				class="w-full rounded-sm border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
				placeholder="Optional instructions…"
			/>
		</div>

		<div class="grid grid-cols-2 gap-4">
			<div>
				<label for="due_date" class="mb-1 block text-sm font-medium">Due Date</label>
				<input
					id="due_date"
					name="due_date"
					type="datetime-local"
					class="w-full rounded-sm border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
				/>
			</div>
			<div>
				<label for="max_points" class="mb-1 block text-sm font-medium">Max Points *</label>
				<input
					id="max_points"
					name="max_points"
					type="number"
					required
					min="0"
					step="0.5"
					class="w-full rounded-sm border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
					placeholder="100"
				/>
			</div>
		</div>

		<div class="grid grid-cols-2 gap-4">
			<div>
				<label for="category" class="mb-1 block text-sm font-medium">Category *</label>
				<select
					id="category"
					name="category"
					required
					class="w-full rounded-sm border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
				>
					{#each categories as cat}
						<option value={cat.value}>{cat.label}</option>
					{/each}
				</select>
			</div>
			<div>
				<label for="weight" class="mb-1 block text-sm font-medium">Weight</label>
				<input
					id="weight"
					name="weight"
					type="number"
					min="0"
					max="1"
					step="0.01"
					value="1.0"
					class="w-full rounded-sm border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
				/>
			</div>
		</div>

		<div class="flex items-center gap-2">
			<input id="is_published" name="is_published" type="checkbox" class="rounded" />
			<label for="is_published" class="text-sm">Publish immediately (students can see this assignment)</label>
		</div>

		<div class="flex gap-3 pt-2">
			<button
				type="submit"
				disabled={loading}
				class="rounded-md bg-primary px-6 py-2 text-sm font-semibold text-primary-foreground disabled:opacity-50 hover:opacity-90"
			>
				{loading ? 'Creating…' : 'Create Assignment'}
			</button>
			<a
				href="/teacher/assignments"
				class="rounded-md border border-border px-6 py-2 text-sm font-medium hover:bg-muted"
			>
				Cancel
			</a>
		</div>
	</form>

	{#if createdAssignmentId}
		<div class="mt-8">
			<h2 class="mb-3 font-semibold">Attachments</h2>
			<FileUpload
				assignmentId={createdAssignmentId}
				on:uploaded={(e) => console.log('Uploaded:', e.detail)}
			/>
		</div>
	{/if}
</div>
