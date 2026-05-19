import { writable } from 'svelte/store';
import axios from 'axios';
import { API_BASE_URL } from '../services/api.js';

export const blacksmithSkill = writable({
  level: 1,
  xp: 0,
  xp_required: 0,
  xp_progress: 0,
});

export async function fetchBlacksmithSkill() {
  try {
    const response = await axios.get(`${API_BASE_URL}/api/blacksmith/skill`, {
      withCredentials: true,
    });
    blacksmithSkill.set(response.data);
    return response.data;
  } catch (error) {
    console.error('Failed to fetch blacksmith skill:', error);
  }
}

export async function syncBlacksmithSkill() {
  return fetchBlacksmithSkill();
}
