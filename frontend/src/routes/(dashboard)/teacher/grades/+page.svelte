<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;
</script>

<svelte:head><title>Grades — Pragma</title></svelte:head>

<div class="mx-auto max-w-7xl px-4 py-6">
	<h1 class="mb-4 text-xl font-bold">Grades</h1>
	{#if data.courses && data.courses.length > 0}
		<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
			{#each data.courses as course (course.id)}
				<a
					href="/teacher/grades/{course.short_id}"
					class="rounded-lg border border-border bg-card p-4 hover:border-primary hover:bg-card/80 transition-colors"
				>
					<div class="font-semibold">{course.name}</div>
					<div class="mt-1 text-sm text-muted-foreground">{course.subject}</div>
					{#if course.period}<div class="text-sm text-muted-foreground">{course.period}</div>{/if}
					<div class="mt-2 text-xs text-muted-foreground">{course.enrollment_count ?? 0} students</div>
				</a>
			{/each}
		</div>
	{:else}
		<div class="rounded-lg border border-dashed border-border p-8 text-center text-muted-foreground">
			<p>No courses found. Contact your administrator to be assigned to courses.</p>
		</div>
	{/if}
</div>
