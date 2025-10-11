import { browser } from '$app/environment';

type Theme = 'light' | 'dark';

function createThemeStore() {
	let theme = $state<Theme>('dark');

	// Initialize theme from localStorage on mount
	if (browser) {
		const stored = localStorage.getItem('theme') as Theme | null;
		const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;

		theme = stored || (prefersDark ? 'dark' : 'light');
		applyTheme(theme);
	}

	function applyTheme(newTheme: Theme) {
		if (browser) {
			document.documentElement.setAttribute('data-theme', newTheme);
		}
	}

	return {
		get current() {
			return theme;
		},
		toggle() {
			const newTheme = theme === 'dark' ? 'light' : 'dark';
			theme = newTheme;
			applyTheme(newTheme);
			if (browser) {
				localStorage.setItem('theme', newTheme);
			}
		},
		set(newTheme: Theme) {
			theme = newTheme;
			applyTheme(newTheme);
			if (browser) {
				localStorage.setItem('theme', newTheme);
			}
		}
	};
}

export const themeStore = createThemeStore();
