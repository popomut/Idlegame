import { writable } from 'svelte/store';
import { API_BASE_URL } from '../services/api.js';
import { addLogEntry } from './game.js';

export const quests = writable([]);
export const isLoadingQuests = writable(false);

// Fetch all quests with current user status
export async function syncQuests() {
  try {
    isLoadingQuests.set(true);
    const response = await fetch(`${API_BASE_URL}/api/quests`, {
      credentials: 'include',
    });
    if (response.ok) {
      const data = await response.json();
      quests.set(data || []);
    }
  } catch (err) {
    console.error('Failed to sync quests:', err);
  } finally {
    isLoadingQuests.set(false);
  }
}

// Attempt to complete a quest; returns { ok, data, error }
export async function completeQuest(questKey) {
  try {
    const response = await fetch(`${API_BASE_URL}/api/quests/${questKey}/complete`, {
      method: 'POST',
      credentials: 'include',
    });
    const data = await response.json();
    if (response.ok) {
      addLogEntry(`Quest complete: ${data.quest}!`);
      // Refresh quest list to reflect new statuses
      await syncQuests();
      return { ok: true, data };
    }
    return { ok: false, error: data.error, unmet: data.unmet };
  } catch (err) {
    console.error('Failed to complete quest:', err);
    return { ok: false, error: 'Network error' };
  }
}
