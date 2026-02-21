<script lang="ts">
	import type { PageData } from './$types';
	import DigitalId from '$lib/components/DigitalId.svelte';

	export let data: PageData;

	function printId() {
		window.print();
	}
</script>

<svelte:head>
	<title>My ID Card — Pragma</title>
</svelte:head>

<div class="mx-auto max-w-lg px-4 py-6">
	<div class="mb-4 flex items-center justify-between no-print">
		<h1 class="text-xl font-bold">My Student ID</h1>
		{#if data.digitalId}
			<button
				on:click={printId}
				class="rounded-md border border-border px-3 py-1.5 text-sm hover:bg-muted"
				aria-label="Print ID card"
			>
				Print
			</button>
		{/if}
	</div>

	{#if data.error}
		<div class="rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800 dark:border-amber-800 dark:bg-amber-950/40 dark:text-amber-400">
			{data.error}
		</div>
	{:else if data.digitalId}
		<div class="flex justify-center">
			<DigitalId
				idNumber={data.digitalId.id_number}
				studentName="{data.digitalId.first_name} {data.digitalId.last_name}"
				gradeLevel={data.digitalId.grade_level}
				schoolName={data.digitalId.school_name}
				schoolLogoUrl={data.digitalId.school_logo_url}
				photoUrl={data.digitalId.photo_url}
				qrCodeData={data.digitalId.qr_code_data}
				issuedAt={data.digitalId.issued_at}
				expiresAt={data.digitalId.expires_at}
				isValid={data.digitalId.is_valid}
			/>
		</div>
		<p class="mt-4 text-center text-xs text-muted-foreground no-print">
			Scan the QR code to verify this ID at <span class="font-mono">pragma.school/verify</span>
		</p>
	{:else}
		<div class="skeleton h-48 w-full rounded-lg" aria-label="Loading…" />
	{/if}
</div>
