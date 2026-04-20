import { writable } from 'svelte/store';

const THEME_KEY = 'theme-preference';

function createThemeStore() {
  function loadThemeFromStorage() {
    if (typeof window === 'undefined') return 'dark';
    const saved = localStorage.getItem(THEME_KEY);
    return saved || 'dark';
  }

  const initialTheme = loadThemeFromStorage();
  const { subscribe, set } = writable(initialTheme);

  // Apply initial theme to DOM
  if (typeof document !== 'undefined') {
    document.documentElement.setAttribute('data-theme', initialTheme);
  }

  return {
    subscribe,
    setTheme: function (theme) {
      set(theme);
      localStorage.setItem(THEME_KEY, theme);
      if (typeof document !== 'undefined') {
        document.documentElement.setAttribute('data-theme', theme);
      }
    }
  };
}

export const theme = createThemeStore();

