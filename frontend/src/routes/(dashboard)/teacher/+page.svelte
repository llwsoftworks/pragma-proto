<script lang="ts">
	/**
	 * Teacher Dashboard — glanceable per spec §9.3.
	 * Shows: today's schedule, needs-attention alerts, quick actions, recent grade activity.
	 */
	import type { PageData } from './$types';
	import AlertBadge from '$lib/components/AlertBadge.svelte';
	import { formatTime, formatPercent } from '$lib/utils';

	export let data: PageData;

	$: d = data.dashboard;
</script>

<svelte:head>
	<title>Dashboard — Pragma</title>
</svelte:head>

<div class="mx-auto max-w-7xl px-4 py-6">
	<div class="grid gap-4 lg:grid-cols-2">
		<!-- Today's Schedule -->
		<section class="rounded-lg border border-border bg-card p-4">
			<h2 class="mb-3 text-sm font-semibold uppercase tracking-wide text-muted-foreground">
				Today's Schedule
			</h2>
			{#if d.today_schedule && d.today_schedule.length > 0}
				<ul class="space-y-2">
					{#each d.today_schedule as block (block.start_time)}
						<li class="flex items-center gap-3">
							<span
								class="h-3 w-3 rounded-full flex-shrink-0"
								style="background-color: {block.color ?? '#3b82f6'}"
								aria-hidden="true"
							/>
							<div>
								<div class="text-sm font-medium">
									{block.course_name || block.label || 'Free'}
								</div>
								<div class="text-xs text-muted-foreground">
									{formatTime(block.start_time)} – {formatTime(block.end_time)}
									{#if block.room} · {block.room}{/if}
								</div>
							</div>
						</li>
					{/each}
				</ul>
			{:else}
				<p class="text-sm text-muted-foreground">No classes scheduled for today.</p>
			{/if}
		</section>

		<!-- Needs Attention -->
		<section class="rounded-lg border border-border bg-card p-4">
			<h2 class="mb-3 text-sm font-semibold uppercase tracking-wide text-muted-foreground">
				Needs Attention
			</h2>
			<div class="space-y-2">
				{#if d.ungraded_assignments && d.ungraded_assignments > 0}
					<AlertBadge type="warning" message="{d.ungraded_assignments} ungraded assignments" />
				{/if}
				{#if !d.ungraded_assignments || d.ungraded_assignments === 0}
					<p class="text-sm text-muted-foreground">Nothing needs your attention right now.</p>
				{/if}
			</div>
		</section>
	</div>

	<!-- Quick Actions -->
	<section class="mt-4 rounded-lg border border-border bg-card p-4">
		<h2 class="mb-3 text-sm font-semibold uppercase tracking-wide text-muted-foreground">
			Quick Actions
		</h2>
		<div class="flex flex-wrap gap-2">
			<a
				href="/teacher/assignments/new"
				class="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90"
			>
				+ New Assignment
			</a>
			<a
				href="/teacher/grades"
				class="rounded-md border border-border bg-background px-4 py-2 text-sm font-medium hover:bg-muted"
			>
				Enter Grades
			</a>
			<a
				href="/teacher/reports"
				class="rounded-md border border-border bg-background px-4 py-2 text-sm font-medium hover:bg-muted"
			>
				Generate Report
			</a>
		</div>
	</section>

	<!-- Recent Grade Activity -->
	{#if d.recent_grade_activity && d.recent_grade_activity.length > 0}
		<section class="mt-4 rounded-lg border border-border bg-card p-4">
			<h2 class="mb-3 text-sm font-semibold uppercase tracking-wide text-muted-foreground">
				Recent Grade Activity
			</h2>
			<ul class="space-y-3">
				{#each d.recent_grade_activity as row (row.assignment_title)}
					<li class="flex items-center gap-3">
						<div class="min-w-0 flex-1">
							<div class="text-sm font-medium">{row.course_name} — {row.assignment_title}</div>
							<div class="mt-1 flex items-center gap-2">
								<!-- Progress bar -->
								<div class="h-2 w-32 overflow-hidden rounded-full bg-muted">
									<div
										class="h-full rounded-full bg-primary"
										style="width: {row.average_percent.toFixed(0)}%"
									/>
								</div>
								<span class="text-xs text-muted-foreground">
									Avg {formatPercent(row.average_percent)} · {row.graded_count}/{row.total_count}
								</span>
							</div>
						</div>
					</li>
				{/each}
			</ul>
		</section>
	{/if}
</div>
