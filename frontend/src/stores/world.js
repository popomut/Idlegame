import { writable } from 'svelte/store';
import { worldAPI } from '../services/api.js';

export const worldActivities = writable([]);
export const worldLastSync = writable(0);
export const worldLoading = writable(false);
export const worldError = writable(null);

export async function fetchWorldActivity() {
  worldLoading.set(true);
  try {
    const { data } = await worldAPI.getActivity();
    worldActivities.set(Array.isArray(data?.activities) ? data.activities : []);
    worldLastSync.set(data?.now || Date.now());
    worldError.set(null);
  } catch (err) {
    worldError.set(err?.response?.data?.error || err.message || 'failed to load');
  } finally {
    worldLoading.set(false);
  }
}
