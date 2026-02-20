<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import { enhance } from '$app/forms';

	export let data: PageData;
	export let form: ActionData;

	let loading = false;
</script>

<svelte:head>
	<title>Log in — Pragma</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center bg-background px-4">
	<div class="w-full max-w-sm space-y-6">
		<div class="text-center">
			<h1 class="text-2xl font-bold tracking-tight">Welcome back</h1>
			<p class="mt-1 text-sm text-muted-foreground">Sign in to your account</p>
			<p class="mt-1 text-sm text-muted-foreground">
				No account? <a href="/register" class="text-primary hover:underline">Create one</a>
			</p>
		</div>

		{#if data.registered}
			<div class="rounded-md bg-green-500/10 px-4 py-3 text-sm text-green-700 dark:text-green-400" role="status">
				Account created — please sign in.
			</div>
		{/if}

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

			<div>
				<label for="password" class="mb-1 block text-sm font-medium">Password</label>
				<input
					id="password"
					name="password"
					type="password"
					autocomplete="current-password"
					required
					class="w-full rounded-sm border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
				/>
			</div>

			<div class="flex items-center justify-between text-sm">
				<a href="/forgot-password" class="text-primary hover:underline">Forgot password?</a>
			</div>

			<button
				type="submit"
				disabled={loading}
				class="w-full rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-50"
			>
				{loading ? 'Signing in…' : 'Sign in'}
			</button>
		</form>
	</div>
</div>
