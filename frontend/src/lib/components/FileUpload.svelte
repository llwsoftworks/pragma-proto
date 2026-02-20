<script lang="ts">
	/**
	 * FileUpload — drag-and-drop file upload using R2 presigned URLs.
	 * Files go directly from browser to R2; they never pass through the Go server.
	 */
	import { createEventDispatcher } from 'svelte';
	import { formatFileSize } from '$lib/utils';
	import { notifications } from '$lib/stores/notifications';

	export let assignmentId: string;
	export let maxFileSizeMB = 25;
	export let accept = '.pdf,.docx,.pptx,.xlsx,.jpg,.jpeg,.png,.gif,.mp3,.mp4';

	const dispatch = createEventDispatcher<{ uploaded: { attachmentId: string; fileName: string } }>();

	let dragging = false;
	let uploading = false;
	let fileInput: HTMLInputElement;

	const MAX_BYTES = maxFileSizeMB * 1024 * 1024;

	async function handleFiles(files: FileList | null) {
		if (!files || files.length === 0) return;

		for (const file of Array.from(files)) {
			if (file.size > MAX_BYTES) {
				notifications.error(`${file.name} exceeds the ${maxFileSizeMB}MB limit`);
				continue;
			}
			await uploadFile(file);
		}
	}

	async function uploadFile(file: File) {
		uploading = true;
		try {
			// Step 1: Get a presigned upload URL from the Go API via the server action.
			const res = await fetch(`/api/assignments/${assignmentId}/upload-url`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					file_name: file.name,
					mime_type: file.type,
					file_size_bytes: file.size
				})
			});

			if (!res.ok) {
				const err = await res.json().catch(() => ({}));
				throw new Error(err.message ?? 'Failed to get upload URL');
			}

			const { upload_url, attachment_id } = await res.json();

			// Step 2: PUT directly to R2.
			const r2Res = await fetch(upload_url, {
				method: 'PUT',
				body: file,
				headers: { 'Content-Type': file.type }
			});

			if (!r2Res.ok) {
				throw new Error('Upload to storage failed');
			}

			notifications.success(`${file.name} uploaded`);
			dispatch('uploaded', { attachmentId: attachment_id, fileName: file.name });
		} catch (err) {
			notifications.error(`Failed to upload ${file.name}: ${err instanceof Error ? err.message : 'Unknown error'}`);
		} finally {
			uploading = false;
		}
	}

	function onDrop(e: DragEvent) {
		e.preventDefault();
		dragging = false;
		handleFiles(e.dataTransfer?.files ?? null);
	}
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<div
	class="relative rounded-lg border-2 border-dashed p-8 text-center transition-colors {dragging ? 'border-primary bg-primary/5' : 'border-border'}"
	role="button"
	tabindex="0"
	aria-label="Drop files here or click to browse"
	on:dragenter|preventDefault={() => (dragging = true)}
	on:dragleave|preventDefault={() => (dragging = false)}
	on:dragover|preventDefault
	on:drop={onDrop}
	on:click={() => fileInput.click()}
	on:keydown={(e) => e.key === 'Enter' && fileInput.click()}
>
	<input
		bind:this={fileInput}
		type="file"
		{accept}
		multiple
		class="sr-only"
		on:change={(e) => handleFiles(e.currentTarget.files)}
	/>

	{#if uploading}
		<div class="text-muted-foreground">
			<div class="mb-2 h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent mx-auto" />
			Uploading…
		</div>
	{:else}
		<div class="text-muted-foreground">
			<svg class="mx-auto mb-2 h-8 w-8 text-muted-foreground/50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
					d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5m-13.5-9L12 3m0 0l4.5 4.5M12 3v13.5" />
			</svg>
			<p class="text-sm font-medium">Drop files here, or <span class="text-primary">browse</span></p>
			<p class="mt-1 text-xs text-muted-foreground">Max {maxFileSizeMB}MB per file</p>
		</div>
	{/if}
</div>
