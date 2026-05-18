import { writable } from 'svelte/store';
import { miningAPI } from '../services/mining.js';

export const miningSkill = writable({
  level: 1,
  xp: 0,
  xp_required: 0,
  xp_progress: 0,
});

export async function fetchMiningSkill() {
  try {
    const response = await miningAPI.getMiningSkill();
    miningSkill.set(response.data);
    return response.data;
  } catch (error) {
    console.error('Failed to fetch mining skill:', error);
    return null;
  }
}

export async function syncMiningSkill() {
  return await fetchMiningSkill();
}
