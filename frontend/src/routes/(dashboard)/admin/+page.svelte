<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;
	$: d = data.dashboard;
</script>

<svelte:head><title>Admin Dashboard — Pragma</title></svelte:head>

<div class="mx-auto max-w-7xl px-4 py-6">
	<h1 class="mb-6 text-xl font-bold">School Overview</h1>

	<div class="grid gap-4 sm:grid-cols-3">
		<div class="rounded-lg border border-border bg-card p-4">
			<div class="text-3xl font-bold text-primary">{d.total_students ?? 0}</div>
			<div class="mt-1 text-sm text-muted-foreground">Active Students</div>
		</div>
		<div class="rounded-lg border border-border bg-card p-4">
			<div class="text-3xl font-bold">{d.total_teachers ?? 0}</div>
			<div class="mt-1 text-sm text-muted-foreground">Teachers</div>
		</div>
		<div class="rounded-lg border border-border bg-card p-4">
			<div class="text-3xl font-bold text-amber-600">{d.locked_students ?? 0}</div>
			<div class="mt-1 text-sm text-muted-foreground">Grade-Locked Students</div>
		</div>
	</div>

	<div class="mt-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
		<a href="/admin/students" class="rounded-lg border border-border bg-card p-4 hover:border-primary transition-colors">
			<div class="font-semibold">Students</div>
			<div class="text-sm text-muted-foreground">Manage roster</div>
		</a>
		<a href="/admin/grade-locks" class="rounded-lg border border-border bg-card p-4 hover:border-primary transition-colors">
			<div class="font-semibold">Grade Locks</div>
			<div class="text-sm text-muted-foreground">Lock / unlock access</div>
		</a>
		<a href="/admin/reports" class="rounded-lg border border-border bg-card p-4 hover:border-primary transition-colors">
			<div class="font-semibold">Reports</div>
			<div class="text-sm text-muted-foreground">Generate report cards</div>
		</a>
		<a href="/admin/settings" class="rounded-lg border border-border bg-card p-4 hover:border-primary transition-colors">
			<div class="font-semibold">Settings</div>
			<div class="text-sm text-muted-foreground">School configuration</div>
		</a>
	</div>
</div>
