<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import { enhance } from '$app/forms';

	export let data: PageData;
	export let form: ActionData;

	let loading = false;
</script>

<svelte:head>
	<title>Create Account — Pragma</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center bg-background px-4 py-10">
	<div class="w-full max-w-md space-y-6">
		<div class="text-center">
			<h1 class="text-2xl font-bold tracking-tight">Create an account</h1>
			<p class="mt-1 text-sm text-muted-foreground">
				Already have an account?
				<a href="/login" class="text-primary hover:underline">Sign in</a>
			</p>
		</div>

		{#if form?.error}
			<div class="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">
				{form.error}
			</div>
		{/if}

		<form
			method="POST"
			class="space-y-4"
			use:enhance={() => {
				loading = true;
				return async ({ update }) => {
					await update();
					loading = false;
				};
			}}
		>
			<!-- Name -->
			<div class="grid grid-cols-2 gap-3">
				<div>
					<label for="first_name" class="mb-1 block text-sm font-medium">First name</label>
					<input
						id="first_name"
						name="first_name"
						type="text"
						autocomplete="given-name"
						required
						value={form?.firstName ?? ''}
						class="w-full rounded-sm border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
					/>
				</div>
				<div>
					<label for="last_name" class="mb-1 block text-sm font-medium">Last name</label>
					<input
						id="last_name"
						name="last_name"
						type="text"
						autocomplete="family-name"
						required
						value={form?.lastName ?? ''}
						class="w-full rounded-sm border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
					/>
				</div>
			</div>

			<!-- Email -->
			<div>
				<label for="email" class="mb-1 block text-sm font-medium">Email</label>
				<input
					id="email"
					name="email"
					type="email"
					autocomplete="email"
					required
					value={form?.email ?? ''}
					class="w-full rounded-sm border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
					placeholder="you@school.edu"
				/>
			</div>

			<!-- Password -->
			<div>
				<label for="password" class="mb-1 block text-sm font-medium">Password</label>
				<input
					id="password"
					name="password"
					type="password"
					autocomplete="new-password"
					required
					minlength="12"
					class="w-full rounded-sm border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
				/>
				<p class="mt-1 text-xs text-muted-foreground">Minimum 12 characters</p>
			</div>

			<!-- Confirm password -->
			<div>
				<label for="confirm_password" class="mb-1 block text-sm font-medium"
					>Confirm password</label
				>
				<input
					id="confirm_password"
					name="confirm_password"
					type="password"
					autocomplete="new-password"
					required
					class="w-full rounded-sm border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
				/>
			</div>

			<!-- Role -->
			<div>
				<label for="role" class="mb-1 block text-sm font-medium">Role</label>
				<select
					id="role"
					name="role"
					required
					class="w-full rounded-sm border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
				>
					<option value="" selected={!form?.role}>Select a role…</option>
					<option value="student" selected={form?.role === 'student'}>Student</option>
					<option value="parent" selected={form?.role === 'parent'}>Parent</option>
					<option value="teacher" selected={form?.role === 'teacher'}>Teacher</option>
					<option value="admin" selected={form?.role === 'admin'}>Admin</option>
				</select>
			</div>

			<!-- School ID -->
			<div>
				<label for="school_id" class="mb-1 block text-sm font-medium">School ID</label>
				<input
					id="school_id"
					name="school_id"
					type="text"
					required
					value={form?.schoolId ?? ''}
					spellcheck="false"
					class="w-full rounded-sm border border-input bg-background px-3 py-2 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-primary"
					placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
				/>
				<p class="mt-1 text-xs text-muted-foreground">UUID provided by your school administrator</p>
			</div>

			<!-- Phone (optional) -->
			<div>
				<label for="phone" class="mb-1 block text-sm font-medium">
					Phone <span class="font-normal text-muted-foreground">(optional)</span>
				</label>
				<input
					id="phone"
					name="phone"
					type="tel"
					autocomplete="tel"
					value={form?.phone ?? ''}
					class="w-full rounded-sm border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
				/>
			</div>

			<button
				type="submit"
				disabled={loading}
				class="w-full rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-50"
			>
				{loading ? 'Creating account…' : 'Create account'}
			</button>
		</form>
	</div>
</div>
