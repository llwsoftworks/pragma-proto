<script lang="ts">
	import type { LayoutData } from './$types';
	import { onMount } from 'svelte';
	import { currentUser } from '$lib/stores/auth';
	import { theme } from '$lib/stores/theme';
	import { notifications } from '$lib/stores/notifications';
	import '../../app.css';

	export let data: LayoutData;

	onMount(() => {
		currentUser.set({
			id: data.user.id,
			email: data.user.email,
			role: data.user.role,
			schoolId: data.user.schoolId,
			mfaDone: true
		});
		theme.init();
	});

	// Role-based navigation items.
	const navByRole: Record<string, Array<{ href: string; label: string }>> = {
		teacher: [
			{ href: '/teacher', label: 'Dashboard' },
			{ href: '/teacher/grades', label: 'Grades' },
			{ href: '/teacher/assignments', label: 'Assignments' },
			{ href: '/teacher/schedule', label: 'Schedule' },
			{ href: '/teacher/reports', label: 'Reports' },
			{ href: '/teacher/ai', label: 'AI Assistant' }
		],
		student: [
			{ href: '/student', label: 'Dashboard' },
			{ href: '/student/grades', label: 'Grades' },
			{ href: '/student/schedule', label: 'Schedule' },
			{ href: '/student/id-card', label: 'My ID' },
			{ href: '/student/documents', label: 'Documents' }
		],
		parent: [
			{ href: '/parent', label: 'Dashboard' },
			{ href: '/parent/grades', label: 'Grades' },
			{ href: '/parent/reports', label: 'Reports' },
			{ href: '/parent/documents', label: 'Documents' },
			{ href: '/parent/messages', label: 'Messages' }
		],
		admin: [
			{ href: '/admin', label: 'Dashboard' },
			{ href: '/admin/students', label: 'Students' },
			{ href: '/admin/teachers', label: 'Teachers' },
			{ href: '/admin/parents', label: 'Parents' },
			{ href: '/admin/courses', label: 'Courses' },
			{ href: '/admin/grade-locks', label: 'Grade Locks' },
			{ href: '/admin/reports', label: 'Reports' },
			{ href: '/admin/settings', label: 'Settings' }
		],
		super_admin: [
			{ href: '/admin', label: 'Dashboard' },
			{ href: '/admin/students', label: 'Students' },
			{ href: '/admin/settings', label: 'Settings' }
		]
	};

	$: navItems = navByRole[data.user.role] ?? [];
</script>

<div class="flex min-h-screen flex-col">
	<!-- Top navigation bar -->
	<header class="border-b border-border bg-background">
		<div class="mx-auto flex max-w-7xl items-center gap-4 px-4 py-3">
			<a href="/{data.user.role}" class="text-lg font-bold text-primary">Pragma</a>
			<nav class="flex gap-1" aria-label="Main navigation">
				{#each navItems as item (item.href)}
					<a
						href={item.href}
						class="rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
					>
						{item.label}
					</a>
				{/each}
			</nav>
			<div class="ml-auto flex items-center gap-2">
				<button
					class="rounded-md px-2 py-1 text-xs text-muted-foreground hover:bg-muted"
					on:click={() => theme.set('dark')}
					aria-label="Toggle dark mode"
				>
					◐
				</button>
				<form method="POST" action="/auth/logout">
					<button type="submit" class="text-xs text-muted-foreground hover:text-foreground">
						Sign out
					</button>
				</form>
			</div>
		</div>
	</header>

	<!-- Page content -->
	<main class="flex-1">
		<slot />
	</main>

	<!-- Toast notifications -->
	<div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2" aria-live="polite">
		{#each $notifications as toast (toast.id)}
			<div
				class="flex items-center gap-3 rounded-md px-4 py-3 text-sm shadow-lg"
				class:bg-green-50={toast.type === 'success'}
				class:text-green-800={toast.type === 'success'}
				class:bg-red-50={toast.type === 'error'}
				class:text-red-800={toast.type === 'error'}
				class:bg-blue-50={toast.type === 'info'}
				class:text-blue-800={toast.type === 'info'}
				class:bg-amber-50={toast.type === 'warning'}
				class:text-amber-800={toast.type === 'warning'}
				role="status"
			>
				<span>{toast.message}</span>
				{#if toast.onUndo}
					<button
						class="ml-2 rounded px-2 py-0.5 text-xs font-semibold underline"
						on:click={toast.onUndo}
					>
						Undo
					</button>
				{/if}
				<button
					class="ml-auto opacity-60 hover:opacity-100"
					on:click={() => notifications.dismiss(toast.id)}
					aria-label="Dismiss notification"
				>
					✕
				</button>
			</div>
		{/each}
	</div>
</div>
