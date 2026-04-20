import { writable } from 'svelte/store';

// Mining configuration
export const miningSyncInterval = writable(5000); // Default: 5 seconds in milliseconds

// Load config from localStorage on startup
export function loadConfigFromStorage() {
  try {
    const saved = localStorage.getItem('game_config');
    if (saved) {
      const config = JSON.parse(saved);
      if (config.miningSyncInterval) {
        miningSyncInterval.set(config.miningSyncInterval);
      }
    }
  } catch (error) {
    console.error('Failed to load config:', error);
  }
}

// Save config to localStorage whenever it changes
miningSyncInterval.subscribe(function (value) {
  try {
    const config = JSON.parse(localStorage.getItem('game_config') || '{}');
    config.miningSyncInterval = value;
    localStorage.setItem('game_config', JSON.stringify(config));
  } catch (error) {
    console.error('Failed to save config:', error);
  }
});
