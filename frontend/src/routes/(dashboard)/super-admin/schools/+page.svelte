<script lang="ts">
	import type { PageData, ActionData } from './$types';
	import { enhance } from '$app/forms';
	import { notifications } from '$lib/stores/notifications';

	export let data: PageData;
	export let form: ActionData;

	let showCreateForm = false;

	$: if (form?.success) {
		showCreateForm = false;
		notifications.add('success', 'School created successfully.');
	}
</script>

<svelte:head><title>Schools — Pragma</title></svelte:head>

<div class="mx-auto max-w-7xl px-4 py-6">
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-xl font-bold">Schools ({data.total})</h1>
		<button
			class="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
			on:click={() => (showCreateForm = !showCreateForm)}
		>
			{showCreateForm ? 'Cancel' : '+ New School'}
		</button>
	</div>

	<!-- Create school form -->
	{#if showCreateForm}
		<div class="mb-6 rounded-lg border border-border bg-card p-4">
			<h2 class="mb-3 font-semibold">Create New School</h2>
			<form method="POST" action="?/create" use:enhance>
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="name" class="mb-1 block text-sm font-medium">School Name *</label>
						<input
							id="name"
							name="name"
							type="text"
							required
							class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
							placeholder="Springfield Elementary"
						/>
					</div>
					<div>
						<label for="address" class="mb-1 block text-sm font-medium">Address</label>
						<input
							id="address"
							name="address"
							type="text"
							class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
							placeholder="123 Main St, Springfield"
						/>
					</div>
				</div>
				{#if form?.error}
					<p class="mt-2 text-sm text-red-600">{form.error}</p>
				{/if}
				<button
					type="submit"
					class="mt-3 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
				>
					Create School
				</button>
			</form>
		</div>
	{/if}

	<!-- Schools list -->
	{#if data.schools && data.schools.length > 0}
		<div class="overflow-x-auto rounded-lg border border-border">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b border-border bg-muted/50">
						<th class="px-4 py-2.5 text-left font-medium text-muted-foreground">School</th>
						<th class="px-4 py-2.5 text-left font-medium text-muted-foreground">Address</th>
						<th class="px-4 py-2.5 text-right font-medium text-muted-foreground">Users</th>
						<th class="px-4 py-2.5 text-right font-medium text-muted-foreground">Students</th>
						<th class="px-4 py-2.5 text-right font-medium text-muted-foreground">Created</th>
						<th class="px-4 py-2.5 text-right font-medium text-muted-foreground">Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each data.schools as school (school.id)}
						<tr class="border-b border-border last:border-0 hover:bg-muted/30">
							<td class="px-4 py-2.5">
								<a href="/super-admin/schools/{school.id}" class="font-medium text-primary hover:underline">
									{school.name}
								</a>
							</td>
							<td class="px-4 py-2.5 text-muted-foreground">{school.address ?? '—'}</td>
							<td class="px-4 py-2.5 text-right">{school.user_count}</td>
							<td class="px-4 py-2.5 text-right">{school.student_count}</td>
							<td class="px-4 py-2.5 text-right text-muted-foreground">
								{new Date(school.created_at).toLocaleDateString()}
							</td>
							<td class="px-4 py-2.5 text-right">
								<a
									href="/super-admin/schools/{school.id}"
									class="text-sm text-primary hover:underline"
								>
									Manage
								</a>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{:else}
		<div class="rounded-lg border border-dashed border-border p-8 text-center">
			<p class="text-muted-foreground">No schools on the platform yet.</p>
			<button
				class="mt-2 text-sm text-primary hover:underline"
				on:click={() => (showCreateForm = true)}
			>
				Create your first school
			</button>
		</div>
	{/if}
</div>
