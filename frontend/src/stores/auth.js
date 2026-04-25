import { writable } from 'svelte/store';
import { API_BASE_URL } from '../services/api.js';

export const authToken = writable(null);
export const username = writable(null);
export const isAuthenticated = writable(false);

export async function loadAuthFromStorage() {
  // Check if authenticated by attempting to fetch user info
  // Token is in httpOnly cookie, so we don't need to store it
  // Just check if we have a valid session
  return await fetchUserInfo();
}

async function fetchUserInfo() {
  try {
    const response = await fetch(`${API_BASE_URL}/api/user`, {
      credentials: 'include', // Send cookies
    });
    
    if (response.ok) {
      const data = await response.json();
      username.set(data.username);
      isAuthenticated.set(true);
      return true;
    } else {
      clearAuth();
      return false;
    }
  } catch (error) {
    console.error('Failed to load user info:', error);
    clearAuth();
    return false;
  }
}

export function clearAuth() {
  // Token is in httpOnly cookie - browser will handle deletion
  // Just clear local state
  username.set(null);
  isAuthenticated.set(false);
  
  // Optionally call logout endpoint to clear server session
  fetch(`${API_BASE_URL}/api/auth/logout`, {
    method: 'POST',
    credentials: 'include',
  }).catch(() => {});
}

export async function refreshToken() {
  // No need to manually refresh - cookies handled automatically
  // Just verify session is still valid
  return await fetchUserInfo();
}
