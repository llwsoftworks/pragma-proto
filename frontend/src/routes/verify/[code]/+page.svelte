<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;
	$: v = data.verification;
</script>

<svelte:head>
	<title>Document Verification — Pragma</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center bg-background px-4">
	<div class="w-full max-w-sm rounded-lg border border-border bg-card p-6 text-center shadow">
		<div class="mb-4 text-5xl" aria-hidden="true">
			{v.valid ? '✅' : '❌'}
		</div>
		<h1 class="text-xl font-bold">
			{v.valid ? 'Verified' : 'Invalid or Expired'}
		</h1>

		{#if v.valid}
			<div class="mt-4 space-y-2 text-sm text-muted-foreground">
				{#if v.student_name}
					<p><span class="font-medium text-foreground">Name:</span> {v.student_name}</p>
				{/if}
				{#if v.document_type}
					<p><span class="font-medium text-foreground">Type:</span> {v.document_type.replace(/_/g, ' ')}</p>
				{/if}
				{#if v.issued_at}
					<p><span class="font-medium text-foreground">Issued:</span> {v.issued_at}</p>
				{/if}
				{#if v.expires_at}
					<p><span class="font-medium text-foreground">Expires:</span> {v.expires_at}</p>
				{/if}
			</div>
		{:else}
			<p class="mt-2 text-sm text-muted-foreground">
				This document could not be verified. It may be expired, revoked, or the code may be incorrect.
			</p>
		{/if}

		<div class="mt-6 text-xs text-muted-foreground">
			Verification code: <span class="font-mono">{data.code.slice(0, 12)}…</span>
		</div>
	</div>
</div>
