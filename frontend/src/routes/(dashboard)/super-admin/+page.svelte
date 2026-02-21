<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;
	$: d = data.dashboard;
</script>

<svelte:head><title>Platform Overview — Pragma</title></svelte:head>

<div class="mx-auto max-w-7xl px-4 py-6">
	<h1 class="mb-6 text-xl font-bold">Platform Overview</h1>

	<!-- Aggregate stats -->
	<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
		<div class="rounded-lg border border-border bg-card p-4">
			<div class="text-3xl font-bold text-primary">{d.total_schools ?? 0}</div>
			<div class="mt-1 text-sm text-muted-foreground">Schools</div>
		</div>
		<div class="rounded-lg border border-border bg-card p-4">
			<div class="text-3xl font-bold">{d.total_users ?? 0}</div>
			<div class="mt-1 text-sm text-muted-foreground">Total Users</div>
		</div>
		<div class="rounded-lg border border-border bg-card p-4">
			<div class="text-3xl font-bold">{d.total_students ?? 0}</div>
			<div class="mt-1 text-sm text-muted-foreground">Active Students</div>
		</div>
		<div class="rounded-lg border border-border bg-card p-4">
			<div class="text-3xl font-bold">{d.total_teachers ?? 0}</div>
			<div class="mt-1 text-sm text-muted-foreground">Teachers</div>
		</div>
		<div class="rounded-lg border border-border bg-card p-4">
			<div class="text-3xl font-bold text-amber-600">{d.total_locked_students ?? 0}</div>
			<div class="mt-1 text-sm text-muted-foreground">Grade-Locked</div>
		</div>
	</div>

	<!-- Quick actions -->
	<div class="mt-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
		<a href="/super-admin/schools" class="rounded-lg border border-border bg-card p-4 hover:border-primary transition-colors">
			<div class="font-semibold">Schools</div>
			<div class="text-sm text-muted-foreground">Manage all schools</div>
		</a>
		<a href="/super-admin/audit-logs" class="rounded-lg border border-border bg-card p-4 hover:border-primary transition-colors">
			<div class="font-semibold">Audit Logs</div>
			<div class="text-sm text-muted-foreground">Platform-wide activity</div>
		</a>
		<a href="/admin/students" class="rounded-lg border border-border bg-card p-4 hover:border-primary transition-colors">
			<div class="font-semibold">Students</div>
			<div class="text-sm text-muted-foreground">Manage student roster</div>
		</a>
		<a href="/admin/settings" class="rounded-lg border border-border bg-card p-4 hover:border-primary transition-colors">
			<div class="font-semibold">Settings</div>
			<div class="text-sm text-muted-foreground">School configuration</div>
		</a>
	</div>

	<!-- Schools overview table -->
	<div class="mt-8">
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-lg font-semibold">Schools</h2>
			<a href="/super-admin/schools" class="text-sm text-primary hover:underline">View all</a>
		</div>
		{#if data.schools && data.schools.length > 0}
			<div class="overflow-x-auto rounded-lg border border-border">
				<table class="w-full text-sm">
					<thead>
						<tr class="border-b border-border bg-muted/50">
							<th class="px-4 py-2 text-left font-medium text-muted-foreground">School</th>
							<th class="px-4 py-2 text-right font-medium text-muted-foreground">Users</th>
							<th class="px-4 py-2 text-right font-medium text-muted-foreground">Students</th>
							<th class="px-4 py-2 text-right font-medium text-muted-foreground">Created</th>
						</tr>
					</thead>
					<tbody>
						{#each data.schools as school (school.id)}
							<tr class="border-b border-border last:border-0 hover:bg-muted/30">
								<td class="px-4 py-2">
									<a href="/super-admin/schools/{school.id}" class="font-medium text-primary hover:underline">
										{school.name}
									</a>
								</td>
								<td class="px-4 py-2 text-right">{school.user_count}</td>
								<td class="px-4 py-2 text-right">{school.student_count}</td>
								<td class="px-4 py-2 text-right text-muted-foreground">
									{new Date(school.created_at).toLocaleDateString()}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{:else}
			<div class="rounded-lg border border-dashed border-border p-8 text-center">
				<p class="text-muted-foreground">No schools yet.</p>
				<a href="/super-admin/schools" class="mt-2 inline-block text-sm text-primary hover:underline">Create your first school</a>
			</div>
		{/if}
	</div>

	<!-- Recent activity -->
	{#if d.recent_activity && d.recent_activity.length > 0}
		<div class="mt-8">
			<h2 class="mb-4 text-lg font-semibold">Recent Activity</h2>
			<div class="space-y-2">
				{#each d.recent_activity as entry}
					<div class="flex items-center justify-between rounded-md border border-border px-4 py-2 text-sm">
						<div>
							<span class="font-mono text-xs bg-muted px-1.5 py-0.5 rounded">{entry.action}</span>
							<span class="ml-2 text-muted-foreground">{entry.user_email}</span>
						</div>
						<div class="text-xs text-muted-foreground">
							{entry.school_name} &middot; {new Date(entry.created_at).toLocaleString()}
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>
