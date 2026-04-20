import { writable } from 'svelte/store';

// On desktop: true = expanded (icon + text), false = collapsed (icon only)
// On mobile:  true = overlay visible,        false = hidden off-screen
const isDesktop = typeof window !== 'undefined' && window.innerWidth >= 768;
export const sidebarOpen = writable(isDesktop);

export function toggleSidebar() {
  sidebarOpen.update(function (open) {
    return !open;
  });
}

export function closeSidebar() {
  sidebarOpen.set(false);
}
