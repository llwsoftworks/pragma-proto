<script lang="ts">
	import type { PageData } from './$types';
	import { goto } from '$app/navigation';

	export let data: PageData;

	let selectedSchool = data.filters.school_id ?? '';
	let selectedAction = data.filters.action ?? '';

	function applyFilters() {
		const params = new URLSearchParams();
		if (selectedSchool) params.set('school_id', selectedSchool);
		if (selectedAction) params.set('action', selectedAction);
		const qs = params.toString();
		goto(`/super-admin/audit-logs${qs ? '?' + qs : ''}`, { invalidateAll: true });
	}

	function clearFilters() {
		selectedSchool = '';
		selectedAction = '';
		goto('/super-admin/audit-logs', { invalidateAll: true });
	}

	function formatTimestamp(ts: string): string {
		return new Date(ts).toLocaleString();
	}
</script>

<svelte:head><title>Audit Logs — Pragma</title></svelte:head>

<div class="mx-auto max-w-7xl px-4 py-6">
	<h1 class="mb-6 text-xl font-bold">Platform Audit Logs</h1>

	<!-- Filters -->
	<div class="mb-6 flex flex-wrap items-end gap-3 rounded-lg border border-border bg-card p-4">
		<div>
			<label for="school-filter" class="mb-1 block text-xs font-medium text-muted-foreground">School</label>
			<select
				id="school-filter"
				bind:value={selectedSchool}
				class="rounded-md border border-input bg-background px-3 py-1.5 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
			>
				<option value="">All Schools</option>
				{#each data.schools as school (school.id)}
					<option value={school.id}>{school.name}</option>
				{/each}
			</select>
		</div>
		<div>
			<label for="action-filter" class="mb-1 block text-xs font-medium text-muted-foreground">Action</label>
			<select
				id="action-filter"
				bind:value={selectedAction}
				class="rounded-md border border-input bg-background px-3 py-1.5 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
			>
				<option value="">All Actions</option>
				<option value="user.login">Login</option>
				<option value="user.login_failed">Login Failed</option>
				<option value="user.register">Register</option>
				<option value="user.create">User Create</option>
				<option value="grade.">Grades</option>
				<option value="grade_lock.">Grade Locks</option>
				<option value="school.">School Changes</option>
				<option value="settings.">Settings</option>
				<option value="document.">Documents</option>
				<option value="report_card.">Reports</option>
			</select>
		</div>
		<button
			class="rounded-md bg-primary px-4 py-1.5 text-sm font-medium text-primary-foreground hover:bg-primary/90"
			on:click={applyFilters}
		>
			Apply
		</button>
		{#if selectedSchool || selectedAction}
			<button
				class="text-sm text-muted-foreground hover:text-foreground"
				on:click={clearFilters}
			>
				Clear
			</button>
		{/if}
	</div>

	<!-- Audit log table -->
	{#if data.auditLogs && data.auditLogs.length > 0}
		<div class="overflow-x-auto rounded-lg border border-border">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b border-border bg-muted/50">
						<th class="px-4 py-2.5 text-left font-medium text-muted-foreground">Timestamp</th>
						<th class="px-4 py-2.5 text-left font-medium text-muted-foreground">Action</th>
						<th class="px-4 py-2.5 text-left font-medium text-muted-foreground">User</th>
						<th class="px-4 py-2.5 text-left font-medium text-muted-foreground">School</th>
						<th class="px-4 py-2.5 text-left font-medium text-muted-foreground">Entity</th>
						<th class="px-4 py-2.5 text-left font-medium text-muted-foreground">IP</th>
					</tr>
				</thead>
				<tbody>
					{#each data.auditLogs as log (log.id)}
						<tr class="border-b border-border last:border-0 hover:bg-muted/30">
							<td class="px-4 py-2 text-xs text-muted-foreground whitespace-nowrap">
								{formatTimestamp(log.created_at)}
							</td>
							<td class="px-4 py-2">
								<span class="font-mono text-xs bg-muted px-1.5 py-0.5 rounded">{log.action}</span>
							</td>
							<td class="px-4 py-2 text-muted-foreground">{log.user_email || '—'}</td>
							<td class="px-4 py-2 text-muted-foreground">{log.school_name || '—'}</td>
							<td class="px-4 py-2 text-xs text-muted-foreground">
								{log.entity_type}{log.entity_id ? ` (${log.entity_id.slice(0, 8)}...)` : ''}
							</td>
							<td class="px-4 py-2 text-xs text-muted-foreground">{log.ip_address ?? '—'}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{:else}
		<div class="rounded-lg border border-dashed border-border p-8 text-center">
			<p class="text-muted-foreground">No audit log entries found.</p>
		</div>
	{/if}
</div>
