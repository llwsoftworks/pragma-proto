<script lang="ts">
	/**
	 * Parent Dashboard — per spec §9.4.
	 * Shows all linked children's grades at a glance.
	 * Tabbed or card layout for multiple children.
	 */
	import type { PageData } from './$types';
	import ChildSelector from '$lib/components/ChildSelector.svelte';
	import AlertBadge from '$lib/components/AlertBadge.svelte';

	export let data: PageData;

	$: children = data.dashboard.children ?? [];
	// Use data.dashboard directly so this is safe during SSR before the reactive
	// `children` declaration has been evaluated (accessing `children[0]` when
	// `children` is still undefined would throw a TypeError and cause a 500).
	let selectedChildId = data.dashboard.children?.[0]?.student_id ?? '';

	$: selectedChild = children.find((c) => c.student_id === selectedChildId);
</script>

<svelte:head>
	<title>Dashboard — Pragma</title>
</svelte:head>

<div class="mx-auto max-w-7xl px-4 py-6">
	{#if children.length === 0}
		<div class="rounded-lg border border-dashed border-border p-8 text-center text-muted-foreground">
			<p>No students linked to your account. Contact your school administrator.</p>
		</div>
	{:else}
		<ChildSelector
			{children}
			bind:selectedId={selectedChildId}
			on:select={(e) => (selectedChildId = e.detail.child.student_id)}
		/>

		{#if selectedChild}
			<div class="mt-4 rounded-lg border border-border bg-card p-4">
				<div class="mb-3">
					<h2 class="text-lg font-semibold">
						{selectedChild.first_name} {selectedChild.last_name}
						<span class="text-sm font-normal text-muted-foreground">· {selectedChild.grade_level}</span>
					</h2>
				</div>

				{#if selectedChild.is_grade_locked || !selectedChild.can_view_grades}
					<AlertBadge
						type="warning"
						message="Your grade access has been temporarily restricted. Please contact your school administration."
					/>
				{:else}
					<a
						href="/parent/grades?student={selectedChildId}"
						class="mt-2 inline-block text-sm text-primary hover:underline"
					>
						View detailed grades →
					</a>
				{/if}
			</div>
		{/if}

		<!-- Quick Actions -->
		<div class="mt-4 rounded-lg border border-border bg-card p-4">
			<h2 class="mb-3 text-sm font-semibold uppercase tracking-wide text-muted-foreground">
				Quick Actions
			</h2>
			<div class="flex flex-wrap gap-2">
				<a
					href="/parent/reports?student={selectedChildId}"
					class="rounded-md border border-border bg-background px-4 py-2 text-sm font-medium hover:bg-muted"
				>
					View Report Card
				</a>
				<a
					href="/parent/documents?student={selectedChildId}"
					class="rounded-md border border-border bg-background px-4 py-2 text-sm font-medium hover:bg-muted"
				>
					Download Enrollment Cert
				</a>
			</div>
		</div>
	{/if}
</div>
