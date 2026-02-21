<script lang="ts">
	import type { PageData, ActionData } from './$types';
	import { enhance } from '$app/forms';
	import { notifications } from '$lib/stores/notifications';

	export let data: PageData;
	export let form: ActionData;

	let showUserForm = false;
	let showEditForm = false;
	let confirmDeactivate = false;

	$: if (form?.success && form?.action === 'createUser') {
		showUserForm = false;
		notifications.add('success', 'User created successfully.');
	}
	$: if (form?.success && form?.action === 'update') {
		showEditForm = false;
		notifications.add('success', 'School updated.');
	}
	$: if (form?.success && form?.action === 'deactivate') {
		notifications.add('info', `School deactivated. ${form.users_deactivated} users affected.`);
	}

	const roleBadgeColors: Record<string, string> = {
		super_admin: 'bg-purple-100 text-purple-800 dark:bg-purple-900/40 dark:text-purple-300',
		admin: 'bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-300',
		teacher: 'bg-green-100 text-green-800 dark:bg-green-900/40 dark:text-green-300',
		parent: 'bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300',
		student: 'bg-slate-100 text-slate-800 dark:bg-slate-800 dark:text-slate-300'
	};
</script>

<svelte:head><title>{data.school.name} — Pragma</title></svelte:head>

<div class="mx-auto max-w-7xl px-4 py-6">
	<div class="mb-1">
		<a href="/super-admin/schools" class="text-sm text-muted-foreground hover:text-foreground">&larr; All Schools</a>
	</div>

	<div class="flex items-start justify-between mb-6">
		<div>
			<h1 class="text-xl font-bold">{data.school.name}</h1>
			{#if data.school.address}
				<p class="text-sm text-muted-foreground">{data.school.address}</p>
			{/if}
		</div>
		<div class="flex gap-2">
			<button
				class="rounded-md border border-border px-3 py-1.5 text-sm hover:bg-muted transition-colors"
				on:click={() => (showEditForm = !showEditForm)}
			>
				Edit
			</button>
			<button
				class="rounded-md border border-red-200 px-3 py-1.5 text-sm text-red-600 hover:bg-red-50 dark:border-red-800 dark:hover:bg-red-950/30 transition-colors"
				on:click={() => (confirmDeactivate = !confirmDeactivate)}
			>
				Deactivate
			</button>
		</div>
	</div>

	<!-- Deactivate confirmation -->
	{#if confirmDeactivate}
		<div class="mb-6 rounded-lg border border-red-200 bg-red-50 p-4 dark:border-red-800 dark:bg-red-950/30">
			<p class="text-sm text-red-800 dark:text-red-300">
				This will deactivate all users at this school and invalidate their sessions. This action cannot be undone from the UI.
			</p>
			<form method="POST" action="?/deactivateSchool" use:enhance class="mt-2">
				<button type="submit" class="rounded-md bg-red-600 px-4 py-1.5 text-sm font-medium text-white hover:bg-red-700">
					Confirm Deactivation
				</button>
				<button type="button" class="ml-2 text-sm text-muted-foreground hover:text-foreground" on:click={() => (confirmDeactivate = false)}>
					Cancel
				</button>
			</form>
		</div>
	{/if}

	<!-- Edit school form -->
	{#if showEditForm}
		<div class="mb-6 rounded-lg border border-border bg-card p-4">
			<h2 class="mb-3 font-semibold">Edit School</h2>
			<form method="POST" action="?/updateSchool" use:enhance>
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="name" class="mb-1 block text-sm font-medium">Name</label>
						<input id="name" name="name" type="text" value={data.school.name}
							class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary" />
					</div>
					<div>
						<label for="address" class="mb-1 block text-sm font-medium">Address</label>
						<input id="address" name="address" type="text" value={data.school.address ?? ''}
							class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary" />
					</div>
				</div>
				<button type="submit" class="mt-3 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
					Save Changes
				</button>
			</form>
		</div>
	{/if}

	<!-- Stats cards -->
	<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-5 mb-8">
		<div class="rounded-lg border border-border bg-card p-4">
			<div class="text-2xl font-bold">{data.total_users}</div>
			<div class="text-sm text-muted-foreground">Users</div>
		</div>
		<div class="rounded-lg border border-border bg-card p-4">
			<div class="text-2xl font-bold">{data.total_students}</div>
			<div class="text-sm text-muted-foreground">Students</div>
		</div>
		<div class="rounded-lg border border-border bg-card p-4">
			<div class="text-2xl font-bold">{data.total_teachers}</div>
			<div class="text-sm text-muted-foreground">Teachers</div>
		</div>
		<div class="rounded-lg border border-border bg-card p-4">
			<div class="text-2xl font-bold">{data.total_courses}</div>
			<div class="text-sm text-muted-foreground">Courses</div>
		</div>
		<div class="rounded-lg border border-border bg-card p-4">
			<div class="text-2xl font-bold text-amber-600">{data.locked_students}</div>
			<div class="text-sm text-muted-foreground">Grade-Locked</div>
		</div>
	</div>

	<!-- Users section -->
	<div class="flex items-center justify-between mb-4">
		<h2 class="text-lg font-semibold">Users</h2>
		<button
			class="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
			on:click={() => (showUserForm = !showUserForm)}
		>
			{showUserForm ? 'Cancel' : '+ Add User'}
		</button>
	</div>

	<!-- Create user form -->
	{#if showUserForm}
		<div class="mb-6 rounded-lg border border-border bg-card p-4">
			<h3 class="mb-3 font-semibold">Add User to School</h3>
			<form method="POST" action="?/createUser" use:enhance>
				<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
					<div>
						<label for="role" class="mb-1 block text-sm font-medium">Role *</label>
						<select id="role" name="role" required
							class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary">
							<option value="admin">Admin</option>
							<option value="teacher">Teacher</option>
							<option value="parent">Parent</option>
							<option value="student">Student</option>
							<option value="super_admin">Super Admin</option>
						</select>
					</div>
					<div>
						<label for="email" class="mb-1 block text-sm font-medium">Email *</label>
						<input id="email" name="email" type="email" required
							class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary" />
					</div>
					<div>
						<label for="password" class="mb-1 block text-sm font-medium">Password *</label>
						<input id="password" name="password" type="password" required minlength="12"
							class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary" />
					</div>
					<div>
						<label for="first_name" class="mb-1 block text-sm font-medium">First Name *</label>
						<input id="first_name" name="first_name" type="text" required
							class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary" />
					</div>
					<div>
						<label for="last_name" class="mb-1 block text-sm font-medium">Last Name *</label>
						<input id="last_name" name="last_name" type="text" required
							class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary" />
					</div>
				</div>
				{#if form?.error}
					<p class="mt-2 text-sm text-red-600">{form.error}</p>
				{/if}
				<button type="submit" class="mt-3 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
					Create User
				</button>
			</form>
		</div>
	{/if}

	<!-- Users table -->
	{#if data.users && data.users.length > 0}
		<div class="overflow-x-auto rounded-lg border border-border">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b border-border bg-muted/50">
						<th class="px-4 py-2.5 text-left font-medium text-muted-foreground">Name</th>
						<th class="px-4 py-2.5 text-left font-medium text-muted-foreground">Email</th>
						<th class="px-4 py-2.5 text-left font-medium text-muted-foreground">Role</th>
						<th class="px-4 py-2.5 text-center font-medium text-muted-foreground">MFA</th>
						<th class="px-4 py-2.5 text-center font-medium text-muted-foreground">Status</th>
						<th class="px-4 py-2.5 text-right font-medium text-muted-foreground">Last Login</th>
					</tr>
				</thead>
				<tbody>
					{#each data.users as user (user.id)}
						<tr class="border-b border-border last:border-0 hover:bg-muted/30">
							<td class="px-4 py-2.5 font-medium">{user.first_name} {user.last_name}</td>
							<td class="px-4 py-2.5 text-muted-foreground">{user.email}</td>
							<td class="px-4 py-2.5">
								<span class="inline-block rounded-full px-2 py-0.5 text-xs font-medium {roleBadgeColors[user.role] ?? ''}">
									{user.role}
								</span>
							</td>
							<td class="px-4 py-2.5 text-center">
								{#if user.mfa_enabled}
									<span class="text-green-600">On</span>
								{:else}
									<span class="text-muted-foreground">Off</span>
								{/if}
							</td>
							<td class="px-4 py-2.5 text-center">
								{#if user.is_active}
									<span class="text-green-600">Active</span>
								{:else}
									<span class="text-red-600">Inactive</span>
								{/if}
							</td>
							<td class="px-4 py-2.5 text-right text-muted-foreground">
								{user.last_login_at ? new Date(user.last_login_at).toLocaleDateString() : 'Never'}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{:else}
		<div class="rounded-lg border border-dashed border-border p-8 text-center">
			<p class="text-muted-foreground">No users at this school yet.</p>
			<button class="mt-2 text-sm text-primary hover:underline" on:click={() => (showUserForm = true)}>
				Add the first user
			</button>
		</div>
	{/if}
</div>
