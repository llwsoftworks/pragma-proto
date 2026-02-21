<script lang="ts">
	import type { LayoutData } from './$types';
	import { onMount } from 'svelte';
	import { currentUser } from '$lib/stores/auth';
	import { theme } from '$lib/stores/theme';
	import { notifications } from '$lib/stores/notifications';

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
			{ href: '/super-admin', label: 'Platform' },
			{ href: '/super-admin/schools', label: 'Schools' },
			{ href: '/super-admin/audit-logs', label: 'Audit Logs' },
			{ href: '/admin/students', label: 'Students' },
			{ href: '/admin/teachers', label: 'Teachers' },
			{ href: '/admin/courses', label: 'Courses' },
			{ href: '/admin/grade-locks', label: 'Grade Locks' },
			{ href: '/admin/reports', label: 'Reports' },
			{ href: '/admin/settings', label: 'Settings' }
		]
	};

	$: navItems = navByRole[data.user.role] ?? [];

	// Mobile nav drawer state.
	let mobileMenuOpen = false;

	// Toggle between light and dark; spec §9.1: "system-preference-aware toggle".
	function toggleTheme() {
		theme.set($theme === 'dark' ? 'light' : 'dark');
	}

	$: themeIcon = $theme === 'dark' ? '☀' : '◐';
	$: themeLabel = $theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode';

	const toastStyles: Record<string, string> = {
		success: 'bg-green-50 text-green-800 border-green-200 dark:bg-green-950/60 dark:text-green-300 dark:border-green-800/40',
		error: 'bg-red-50 text-red-800 border-red-200 dark:bg-red-950/60 dark:text-red-300 dark:border-red-800/40',
		info: 'bg-blue-50 text-blue-800 border-blue-200 dark:bg-blue-950/60 dark:text-blue-300 dark:border-blue-800/40',
		warning: 'bg-amber-50 text-amber-800 border-amber-200 dark:bg-amber-950/60 dark:text-amber-300 dark:border-amber-800/40'
	};
</script>

<div class="flex min-h-screen flex-col">
	<!-- Top navigation bar -->
	<header class="border-b border-border bg-background">
		<div class="mx-auto flex max-w-7xl items-center gap-3 px-4 py-3">
			<a href={data.user.role === 'super_admin' ? '/super-admin' : `/${data.user.role}`} class="shrink-0 text-lg font-bold text-primary">Pragma</a>

			<!-- Desktop nav (hidden on mobile) -->
			<nav class="hidden flex-1 gap-1 md:flex" aria-label="Main navigation">
				{#each navItems as item (item.href)}
					<a
						href={item.href}
						class="rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
					>
						{item.label}
					</a>
				{/each}
			</nav>

			<!-- Desktop: theme toggle + sign-out -->
			<div class="ml-auto hidden items-center gap-2 md:flex">
				<button
					class="rounded-md px-2 py-1 text-sm text-muted-foreground hover:bg-muted"
					on:click={toggleTheme}
					aria-label={themeLabel}
					title={themeLabel}
				>
					{themeIcon}
				</button>
				<form method="POST" action="/logout">
					<button type="submit" class="text-xs text-muted-foreground hover:text-foreground">
						Sign out
					</button>
				</form>
			</div>

			<!-- Mobile: theme toggle + hamburger button -->
			<div class="ml-auto flex items-center gap-1 md:hidden">
				<button
					class="rounded-md px-2 py-1 text-sm text-muted-foreground hover:bg-muted"
					on:click={toggleTheme}
					aria-label={themeLabel}
					title={themeLabel}
				>
					{themeIcon}
				</button>
				<button
					class="rounded-md p-1.5 text-muted-foreground hover:bg-muted"
					on:click={() => (mobileMenuOpen = !mobileMenuOpen)}
					aria-expanded={mobileMenuOpen}
					aria-label={mobileMenuOpen ? 'Close menu' : 'Open menu'}
				>
					{#if mobileMenuOpen}
						<!-- Close (X) icon -->
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
							<path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
						</svg>
					{:else}
						<!-- Hamburger icon -->
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
							<path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
						</svg>
					{/if}
				</button>
			</div>
		</div>

		<!-- Mobile nav drawer (shown when hamburger is open) -->
		{#if mobileMenuOpen}
			<nav
				class="border-t border-border bg-background px-4 pb-3 md:hidden"
				aria-label="Mobile navigation"
			>
				{#each navItems as item (item.href)}
					<a
						href={item.href}
						class="block rounded-md px-3 py-2 text-sm font-medium text-muted-foreground hover:bg-muted hover:text-foreground"
						on:click={() => (mobileMenuOpen = false)}
					>
						{item.label}
					</a>
				{/each}
				<div class="mt-2 border-t border-border pt-2">
					<form method="POST" action="/logout">
						<button
							type="submit"
							class="block w-full rounded-md px-3 py-2 text-left text-sm text-muted-foreground hover:bg-muted hover:text-foreground"
						>
							Sign out
						</button>
					</form>
				</div>
			</nav>
		{/if}
	</header>

	<!-- Page content -->
	<main class="flex-1">
		<slot />
	</main>

	<!-- Toast notifications -->
	<div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2" aria-live="polite">
		{#each $notifications as toast (toast.id)}
			<div
				class="flex items-center gap-3 rounded-md border px-4 py-3 text-sm shadow-lg {toastStyles[toast.type] ?? toastStyles.info}"
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
