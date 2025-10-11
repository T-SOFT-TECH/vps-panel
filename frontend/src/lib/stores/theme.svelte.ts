import { browser } from '$app/environment';

type Theme = 'light' | 'dark';

function createThemeStore() {
	let theme = $state<Theme>('dark');

	// Initialize theme from localStorage on mount
	if (browser) {
		const stored = localStorage.getItem('theme') as Theme | null;
		if (stored) {
			theme = stored;
		} else {
			// Check system preference
			const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
			theme = prefersDark ? 'dark' : 'light';
		}
		applyTheme(theme);
	}

	function applyTheme(newTheme: Theme) {
		if (browser) {
			document.documentElement.classList.remove('light', 'dark');
			document.documentElement.classList.add(newTheme);
			document.documentElement.setAttribute('data-theme', newTheme);
		}
	}

	return {
		get current() {
			return theme;
		},
		toggle() {
			theme = theme === 'dark' ? 'light' : 'dark';
			applyTheme(theme);
			if (browser) {
				localStorage.setItem('theme', theme);
			}
		},
		set(newTheme: Theme) {
			theme = newTheme;
			applyTheme(theme);
			if (browser) {
				localStorage.setItem('theme', newTheme);
			}
		}
	};
}

export const themeStore = createThemeStore();
