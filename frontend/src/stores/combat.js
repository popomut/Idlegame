import { writable, get } from 'svelte/store';
import { combatAPI } from '../services/api.js';

const POLL_INTERVAL_MS = 1500;

// ── Store shape ──────────────────────────────────────────────────────────────
// status: 'none' | 'active' | 'fled' | 'dead'
const defaultState = {
  status: 'none',
  isActive: false,
  zoneKey: '',
  playerHPCurrent: 0,
  playerMaxHP: 100,
  currentEnemy: null,     // { key, name, icon, hp_current, hp_max, attack_value }
  enemiesDefeated: 0,
  totalXPGained: 0,
  totalMoneyGained: 0,
  recentLogs: [],         // newest-first (reversed for display)
  sessionStartedAt: 0,
  wasOffline: false,
  offlineTimeMS: 0,
  offlineEnemies: 0,
  error: null,
};

export const combatState = writable({ ...defaultState });

let pollTimer = null;
let visibilityListener = null;
let lastKnownPageVisible = true;

// ── Actions ──────────────────────────────────────────────────────────────────

export async function startCombatSession(zoneKey) {
  combatState.update(s => ({ ...s, error: null }));
  try {
    await combatAPI.startCombat(zoneKey);
    await fetchCombatStatus();
    startPolling();
  } catch (e) {
    const msg = e?.response?.data?.error || 'Failed to start combat.';
    combatState.update(s => ({ ...s, error: msg }));
    throw e;
  }
}

export async function fleeCombatSession() {
  stopPolling();
  try {
    await combatAPI.flee();
    await fetchCombatStatus(); // get final state
  } catch (e) {
    // Still try to get status
    await fetchCombatStatus();
  }
}

export async function fetchCombatStatus() {
  try {
    const res = await combatAPI.getStatus();
    const d = res.data;
    combatState.set({
      status: d.status || 'none',
      isActive: d.is_active,
      zoneKey: d.zone_key || '',
      playerHPCurrent: d.player_hp_current || 0,
      playerMaxHP: d.player_max_hp || 100,
      currentEnemy: d.current_enemy || null,
      enemiesDefeated: d.enemies_defeated || 0,
      totalXPGained: d.total_xp_gained || 0,
      totalMoneyGained: d.total_money_gained || 0,
      recentLogs: (d.recent_logs || []).slice().reverse(), // show newest first
      sessionStartedAt: d.session_started_at || 0,
      wasOffline: d.was_offline || false,
      offlineTimeMS: d.offline_time_ms || 0,
      offlineEnemies: d.offline_enemies || 0,
      error: null,
    });

    // Stop polling when combat ends
    if (d.status === 'dead' || d.status === 'fled') {
      stopPolling();
    }
  } catch (e) {
    if (e?.response?.status !== 401) {
      combatState.update(s => ({ ...s, error: 'Failed to get combat status.' }));
    }
  }
}

export function startPolling() {
  stopPolling();
  
  // Set up visibility change listener
  if (typeof document !== 'undefined' && !visibilityListener) {
    visibilityListener = () => {
      if (document.hidden) {
        // Page is hidden (user switched tabs)
        lastKnownPageVisible = false;
        console.log('[Combat] Page hidden - stopping polling');
        // Stop polling while hidden to save resources
        if (pollTimer) {
          clearInterval(pollTimer);
          pollTimer = null;
        }
      } else {
        // Page is visible again (user returned to tab)
        lastKnownPageVisible = true;
        console.log('[Combat] Page visible - fetching status immediately');
        // Fetch status immediately to catch up with offline progress
        const state = get(combatState);
        if (state.isActive) {
          console.log('[Combat] Combat is active, calling fetchCombatStatus()');
          fetchCombatStatus().then(() => {
            console.log('[Combat] fetchCombatStatus() completed');
          });
          // Restart polling
          resumePolling();
        }
      }
    };
    document.addEventListener('visibilitychange', visibilityListener);
    console.log('[Combat] Visibility listener attached');
  }
  
  // Start the polling interval
  resumePolling();
}

function resumePolling() {
  // Don't start if already running
  if (pollTimer) {
    console.log('[Combat] Polling already running, skipping resume');
    return;
  }
  
  console.log('[Combat] Starting polling interval');
  pollTimer = setInterval(() => {
    const state = get(combatState);
    if (state.isActive) {
      console.log('[Combat] Polling: fetching status');
      fetchCombatStatus();
    } else {
      console.log('[Combat] Combat not active, stopping polling');
      stopPolling();
    }
  }, POLL_INTERVAL_MS);
}

export function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
  // Remove visibility listener when stopping polling
  if (visibilityListener && typeof document !== 'undefined') {
    document.removeEventListener('visibilitychange', visibilityListener);
    visibilityListener = null;
  }
}

export function resetCombatState() {
  stopPolling();
  combatState.set({ ...defaultState });
}

export function dismissOfflineGains() {
  combatState.update(s => ({ ...s, wasOffline: false, offlineEnemies: 0, offlineTimeMS: 0 }));
}
