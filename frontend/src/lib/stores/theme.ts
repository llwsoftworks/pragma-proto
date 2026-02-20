import { writable } from 'svelte/store';
import { browser } from '$app/environment';

type Theme = 'light' | 'dark' | 'system';

function createThemeStore() {
	const stored = browser ? (localStorage.getItem('theme') as Theme | null) : null;
	const { subscribe, set } = writable<Theme>(stored ?? 'system');

	return {
		subscribe,
		set: (theme: Theme) => {
			if (browser) {
				localStorage.setItem('theme', theme);
				applyTheme(theme);
			}
			set(theme);
		},
		init: () => {
			if (!browser) return;
			const saved = (localStorage.getItem('theme') as Theme | null) ?? 'system';
			set(saved);
			applyTheme(saved);

			// Listen for system preference changes.
			window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
				// Re-apply only if the user hasn't made an explicit choice.
				const current = localStorage.getItem('theme') as Theme | null;
				if (!current || current === 'system') {
					applyTheme('system');
				}
			});
		}
	};
}

function applyTheme(theme: Theme) {
	const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
	const isDark = theme === 'dark' || (theme === 'system' && prefersDark);
	document.documentElement.classList.toggle('dark', isDark);
}

export const theme = createThemeStore();
