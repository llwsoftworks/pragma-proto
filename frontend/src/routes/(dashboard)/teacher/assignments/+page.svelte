<script lang="ts">
	/**
	 * Teacher — Assignment list grouped by course (spec §5.1).
	 */
	import type { PageData } from './$types';
	import { formatDate, encodeId } from '$lib/utils';
	import type { AssignmentListItem } from '$lib/api';

	export let data: PageData;

	// Group assignments by course name for display.
	$: grouped = data.assignments.reduce<Record<string, AssignmentListItem[]>>((acc, a) => {
		(acc[a.course_name] ??= []).push(a);
		return acc;
	}, {});

	$: courseNames = Object.keys(grouped).sort();
</script>

<svelte:head>
	<title>Assignments — Pragma</title>
</svelte:head>

<div class="mx-auto max-w-7xl px-4 py-6">
	<div class="mb-4 flex items-center justify-between">
		<h1 class="text-xl font-bold">Assignments</h1>
		<a
			href="/teacher/assignments/new"
			class="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90"
		>
			+ New Assignment
		</a>
	</div>

	{#if data.assignments.length === 0}
		<div class="rounded-lg border border-dashed border-border p-8 text-center text-muted-foreground">
			<p>No assignments yet.</p>
			<a href="/teacher/assignments/new" class="mt-2 inline-block text-sm text-primary hover:underline">
				Create your first one →
			</a>
		</div>
	{:else}
		<div class="space-y-6">
			{#each courseNames as courseName (courseName)}
				<section class="rounded-lg border border-border bg-card">
					<h2 class="border-b border-border px-4 py-3 text-sm font-semibold">
						{courseName}
					</h2>
					<ul class="divide-y divide-border">
						{#each grouped[courseName] as a (a.id)}
							<li class="flex items-center gap-4 px-4 py-3 hover:bg-muted/50">
								<div class="min-w-0 flex-1">
									<a
										href="/teacher/assignments/{encodeId(a.id)}"
										class="text-sm font-medium hover:underline"
									>
										{a.title}
									</a>
									<div class="mt-0.5 text-xs text-muted-foreground">
										{a.category} · {a.max_points} pts
										{#if a.due_date}
											· Due {formatDate(a.due_date)}
										{/if}
									</div>
								</div>
								<span
									class="shrink-0 rounded-full px-2 py-0.5 text-xs font-medium"
									class:bg-green-100={a.is_published}
									class:text-green-700={a.is_published}
									class:bg-muted={!a.is_published}
									class:text-muted-foreground={!a.is_published}
								>
									{a.is_published ? 'Published' : 'Draft'}
								</span>
							</li>
						{/each}
					</ul>
				</section>
			{/each}
		</div>
	{/if}
</div>
