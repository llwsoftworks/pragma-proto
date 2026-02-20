<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import { enhance } from '$app/forms';

	export let data: PageData;
	export let form: ActionData;

	let loading = false;
</script>

<svelte:head>
	<title>Sign in — Pragma</title>
</svelte:head>

<div class="flex min-h-screen flex-col items-center justify-center bg-slate-50 px-4 py-16 dark:bg-slate-950">
	<!-- Brand -->
	<div class="mb-8 text-center">
		<span class="text-3xl font-bold tracking-tight text-primary">Pragma</span>
		<p class="mt-1 text-sm text-slate-500 dark:text-slate-400">Student Grading Platform</p>
	</div>

	<!-- Card -->
	<div class="w-full max-w-sm rounded-2xl bg-white p-8 shadow-sm ring-1 ring-slate-200 dark:bg-slate-900 dark:ring-slate-800">
		<h1 class="mb-1 text-xl font-semibold text-slate-900 dark:text-slate-50">Welcome back</h1>
		<p class="mb-6 text-sm text-slate-500 dark:text-slate-400">Sign in to your account to continue.</p>

		{#if data.registered}
			<div
				class="mb-5 rounded-lg bg-emerald-50 px-4 py-3 text-sm text-emerald-700 ring-1 ring-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-400 dark:ring-emerald-800"
				role="status"
			>
				Account created — please sign in.
			</div>
		{/if}

		{#if form?.error}
			<div
				class="mb-5 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600 ring-1 ring-red-200 dark:bg-red-950/40 dark:text-red-400 dark:ring-red-800"
				role="alert"
			>
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
				<label
					for="email"
					class="mb-1.5 block text-sm font-medium text-slate-700 dark:text-slate-300"
				>
					Email
				</label>
				<input
					id="email"
					name="email"
					type="email"
					autocomplete="email"
					required
					value={form?.email ?? ''}
					placeholder="you@school.edu"
					class="block w-full rounded-lg border border-slate-200 bg-white px-3.5 py-2.5 text-sm text-slate-900 placeholder-slate-400 transition-[box-shadow,border-color] duration-150 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-50 dark:placeholder-slate-500 dark:focus:border-primary"
				/>
			</div>

			<div>
				<div class="mb-1.5 flex items-center justify-between">
					<label
						for="password"
						class="block text-sm font-medium text-slate-700 dark:text-slate-300"
					>
						Password
					</label>
					<a href="/forgot-password" class="text-xs text-primary hover:underline">
						Forgot password?
					</a>
				</div>
				<input
					id="password"
					name="password"
					type="password"
					autocomplete="current-password"
					required
					class="block w-full rounded-lg border border-slate-200 bg-white px-3.5 py-2.5 text-sm text-slate-900 transition-[box-shadow,border-color] duration-150 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-50 dark:focus:border-primary"
				/>
			</div>

			<button
				type="submit"
				disabled={loading}
				class="mt-1 flex w-full items-center justify-center rounded-lg bg-primary px-4 py-2.5 text-sm font-semibold text-primary-foreground shadow-sm transition-opacity duration-150 hover:opacity-90 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
			>
				{#if loading}
					<svg
						class="mr-2 h-4 w-4 animate-spin"
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						aria-hidden="true"
					>
						<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
						<path
							class="opacity-75"
							fill="currentColor"
							d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
						/>
					</svg>
					Signing in…
				{:else}
					Sign in
				{/if}
			</button>
		</form>
	</div>

	<!-- Footer -->
	<p class="mt-6 text-sm text-slate-500 dark:text-slate-400">
		No account?
		<a href="/register" class="font-medium text-primary hover:underline">Create one</a>
	</p>
</div>
