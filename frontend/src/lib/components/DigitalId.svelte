<script lang="ts">
	/**
	 * DigitalId — student ID card display.
	 * Printable via CSS print stylesheet to 3.375" × 2.125" (standard ID size).
	 * Offline-capable (cache via PWA service worker).
	 */
	import { formatDate } from '$lib/utils';

	export let idNumber: string;
	export let studentName: string;
	export let gradeLevel: string;
	export let schoolName: string;
	export let schoolLogoUrl: string | null = null;
	export let photoUrl: string | null = null;
	export let qrCodeData: string;
	export let issuedAt: string;
	export let expiresAt: string;
	export let isValid = true;
</script>

<div
	class="id-card relative flex flex-col rounded-lg border border-border bg-white shadow-md dark:bg-card"
	class:opacity-60={!isValid}
	style="width: 3.375in; min-height: 2.125in; font-family: Inter, sans-serif;"
>
	<!-- Header strip -->
	<div class="flex items-center gap-2 rounded-t-lg bg-primary px-3 py-2">
		{#if schoolLogoUrl}
			<img src={schoolLogoUrl} alt="{schoolName} logo" class="h-6 w-6 rounded object-contain" />
		{/if}
		<span class="text-xs font-bold uppercase tracking-wide text-white">{schoolName}</span>
		{#if !isValid}
			<span class="ml-auto rounded bg-red-600 px-1 py-0.5 text-[9px] font-bold uppercase text-white">
				REVOKED
			</span>
		{/if}
	</div>

	<!-- Body -->
	<div class="flex flex-1 gap-3 px-3 py-2">
		<!-- Photo -->
		<div class="flex-shrink-0">
			{#if photoUrl}
				<img src={photoUrl} alt="Student photo" class="h-16 w-12 rounded object-cover" />
			{:else}
				<div class="flex h-16 w-12 items-center justify-center rounded bg-muted text-muted-foreground">
					<svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
							d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" />
					</svg>
				</div>
			{/if}
		</div>

		<!-- Info -->
		<div class="flex flex-1 flex-col justify-between">
			<div>
				<div class="text-sm font-bold text-foreground leading-tight">{studentName}</div>
				<div class="text-xs text-muted-foreground">Grade {gradeLevel}</div>
				<div class="mt-1 font-mono text-xs font-semibold tracking-wider">{idNumber}</div>
			</div>
			<div class="text-[9px] text-muted-foreground">
				<div>Issued: {formatDate(issuedAt)}</div>
				<div>Expires: {formatDate(expiresAt)}</div>
			</div>
		</div>

		<!-- QR Code -->
		<div class="flex-shrink-0">
			<img
				src="/api/qr/{encodeURIComponent(qrCodeData)}"
				alt="QR code for ID verification"
				class="h-16 w-16"
				loading="lazy"
			/>
		</div>
	</div>
</div>

<style>
	@media print {
		.id-card {
			width: 3.375in !important;
			min-height: 2.125in !important;
			box-shadow: none !important;
			border: 1px solid #000 !important;
			break-inside: avoid;
		}
	}
</style>
