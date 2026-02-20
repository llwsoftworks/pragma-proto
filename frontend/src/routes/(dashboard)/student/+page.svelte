<script lang="ts">
	import type { PageData } from './$types';
	import AlertBadge from '$lib/components/AlertBadge.svelte';

	export let data: PageData;
	$: d = data.dashboard;
</script>

<svelte:head><title>Dashboard — Pragma</title></svelte:head>

<div class="mx-auto max-w-4xl px-4 py-6">
	<h1 class="mb-4 text-xl font-bold">My Dashboard</h1>

	{#if d.is_grade_locked}
		<AlertBadge
			type="warning"
			message="Your grade access has been temporarily restricted. Please contact your school administration."
		/>
	{:else}
		<div class="grid gap-3 sm:grid-cols-3">
			<a
				href="/student/grades"
				class="rounded-lg border border-border bg-card p-4 text-center hover:border-primary transition-colors"
			>
				<div class="text-2xl font-bold text-primary">→</div>
				<div class="mt-1 text-sm font-medium">View Grades</div>
			</a>
			<a
				href="/student/schedule"
				class="rounded-lg border border-border bg-card p-4 text-center hover:border-primary transition-colors"
			>
				<div class="text-2xl font-bold text-primary">📅</div>
				<div class="mt-1 text-sm font-medium">My Schedule</div>
			</a>
			<a
				href="/student/id-card"
				class="rounded-lg border border-border bg-card p-4 text-center hover:border-primary transition-colors"
			>
				<div class="text-2xl font-bold text-primary">🪪</div>
				<div class="mt-1 text-sm font-medium">My ID Card</div>
			</a>
		</div>
	{/if}
</div>
