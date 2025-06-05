const DEFAULT_LIGHT_THEME = 'nord';
const DEFAULT_DARK_THEME = 'night';

// --- theme management ---

const THEME_KEY = 'IM_THEME';
(function () {
	const themeChangeCallbacks = [];

	window.getTheme = function () {
		return localStorage.getItem(THEME_KEY) ||
			(window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? DEFAULT_DARK_THEME : DEFAULT_LIGHT_THEME);
	}

	window.setTheme = function (theme) {
		localStorage.setItem(THEME_KEY, theme);
		document.documentElement.setAttribute('data-theme', theme);
		themeChangeCallbacks.forEach(cb => cb(theme));
	};

	window.onThemeChange = function (cb) {
		themeChangeCallbacks.push(cb);
	};

	loadThem = getTheme();
	document.documentElement.setAttribute('data-theme', loadThem);
	localStorage.setItem(THEME_KEY, loadThem);
})();
// setup navbar theme switcher
document.addEventListener('DOMContentLoaded', function () {
	const themes = document.getElementById('themes')
	if (!themes) return;
	const themeButtons = themes.querySelectorAll('[data-set-theme]')
	const loadTheme = getTheme()
	for (const button of themeButtons) {
		const theme = button.getAttribute('data-set-theme')
		if (theme === loadTheme) {
			button.querySelector('.theme-checkmark').classList.remove('invisible')
		}
		button.addEventListener('click', () => {
			setTheme(theme)
			localStorage.setItem(THEME_KEY, theme)
			for (const btn of themeButtons) {
				if (btn !== button) {
					btn.querySelector('.theme-checkmark').classList.add('invisible')
				}
			}
			button.querySelector('.theme-checkmark').classList.remove('invisible')
		})
	}
});

// --- click blocker ---

(function () {
	window.blockClicks = function () {
		const blocker = document.getElementById('click-blocker');
		if (blocker) blocker.classList.remove('hidden');
	};

	window.unblockClicks = function () {
		const blocker = document.getElementById('click-blocker');
		if (blocker) blocker.classList.add('hidden');
	};
})();
