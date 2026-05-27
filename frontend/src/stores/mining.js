import { writable } from 'svelte/store';
import { miningAPI, inventoryAPI } from '../services/api.js';
import { ores, herbs, addLogEntry } from './game.js';

export const activeMining = writable(null);
export const miningPopups = writable([]);
export const offlineGains = writable(null);
export const isLoadingMining = writable(false);
export const miningProgress = writable(0); // 0-100 progress bar

let globalProgressInterval = null;
let isStoppingMining = false;

function startGlobalProgressUpdater(extractionTimeMS) {
  if (globalProgressInterval) {
    clearInterval(globalProgressInterval);
  }
  const interval = extractionTimeMS || 3000;

  globalProgressInterval = setInterval(function () {
    activeMining.subscribe(function (mining) {
      if (mining && mining.startedAt) {
        const elapsed = Date.now() - new Date(mining.startedAt).getTime();
        const cycle = elapsed % interval;
        miningProgress.set((cycle / interval) * 100);
      }
    })();
  }, 50);
}

function stopGlobalProgressUpdater() {
  if (globalProgressInterval) {
    clearInterval(globalProgressInterval);
    globalProgressInterval = null;
  }
  miningProgress.set(0);
}

activeMining.subscribe(function (mining) {
  if (mining) {
    startGlobalProgressUpdater(mining.extractionTimeMS);
  } else {
    stopGlobalProgressUpdater();
  }
});

// Internal: last synced inventory map (ore_key -> quantity)
let lastOreInventory = {};

export async function startMining(oreId, oreName, oreKey, extractionTimeMS, resourceType = 'ore') {
  try {
    isLoadingMining.set(true);
    const response = await miningAPI.startMining(oreId, resourceType);

    activeMining.set({
      oreId,
      oreKey,
      oreName,
      resourceType,
      extractionTimeMS: extractionTimeMS || 3000,
      sessionId: response.data.session_id,
      startedAt: new Date(response.data.started_at),
    });

    addLogEntry(`Started extracting ${oreName}...`);
  } catch (error) {
    console.error('Failed to start mining:', error);
    addLogEntry('Failed to start extraction. Try again.');
  } finally {
    isLoadingMining.set(false);
  }
}

export async function stopMining() {
  try {
    isLoadingMining.set(true);
    const response = await miningAPI.stopMining();
    const oresGained = response.data.ores_gained;
    const wasHerb = response.data.resource_type === 'herb';

    activeMining.set(null);

    // Always sync both inventories on stop
    await syncOreInventory();
    try { await syncHerbInventory(); } catch (e) { /* ignore */ }

    if (oresGained > 0) {
      showMiningPopup(oresGained);
      addLogEntry(`Extraction complete - gained ${oresGained} units.`);
    }
  } catch (error) {
    console.error('Failed to stop mining:', error);
    addLogEntry('Failed to stop extraction. Try again.');
  } finally {
    isLoadingMining.set(false);
  }
}

export async function checkMiningStatus() {
  try {
    const response = await miningAPI.getMiningStatus();
    const status = response.data;

    if (status.offline_gains && status.offline_gains.was_offline) {
      if (isStoppingMining) return;
      isStoppingMining = true;
      const gains = status.offline_gains;
      activeMining.set(null);
      offlineGains.set({
        wasOffline: true,
        timeMs: gains.offline_time_ms,
        oresGained: gains.ores_gained,
        oreName: gains.ore_name,
      });
      addLogEntry(`You gained ${gains.ores_gained} ${gains.ore_name} while away!`);
      await stopMining();
      isStoppingMining = false;
      return;
    }

    if (status.is_active && status.current_ore) {
      activeMining.set({
        oreId: status.current_ore.id,
        oreKey: status.current_ore.ore_key,
        oreName: status.current_ore.ore_name,
        resourceType: status.current_ore.resource_type || 'ore',
        extractionTimeMS: status.current_ore.mining_time_ms || 3000,
        startedAt: new Date(status.started_at),
      });
    } else {
      activeMining.set(null);
    }

    // Sync inventory — ores always, herbs if present
    if (status.current_ores) {
      ores.set({ ...status.current_ores });
    } else {
      await syncOreInventory();
    }
    if (status.current_herbs) {
      herbs.set({ ...status.current_herbs });
    } else {
      await syncHerbInventory();
    }
  } catch (error) {
    console.error('Failed to check mining status:', error);
  }
}

// Sync ore inventory from backend (returns array of {ore_key, quantity, ...})
export async function syncOreInventory() {
  try {
    const response = await inventoryAPI.getOreInventory();
    const data = response.data;

    const oreMap = {};
    if (Array.isArray(data)) {
      data.forEach(function (item) {
        oreMap[item.ore_key] = item.quantity;
      });
    }

    ores.set(oreMap);
    lastOreInventory = { ...oreMap };
  } catch (error) {
    console.error('Failed to sync ore inventory:', error);
  }
}

// Poll mining status to update counts with pending server-side estimate
export async function syncOreInventoryDuringMining() {
  try {
    const response = await miningAPI.getMiningStatus();
    const data = response.data;

    if (data.current_ores) {
      ores.set({ ...data.current_ores });
    }
  } catch (error) {
    console.error('Failed to sync ore inventory during mining:', error);
  }
}

export function showMiningPopup(count = 1) {
  const id = Date.now();
  miningPopups.update(function (popups) {
    return [...popups, { id, count }];
  });
  setTimeout(function () {
    miningPopups.update(function (popups) {
      return popups.filter(function (p) { return p.id !== id; });
    });
  }, 1000);
}

// Sync gains while away from Mining view
export async function syncMiningProgress() {
  try {
    const response = await inventoryAPI.getOreInventory();
    const data = response.data;

    const currentInventory = {};
    if (Array.isArray(data)) {
      data.forEach(function (item) {
        currentInventory[item.ore_key] = item.quantity;
      });
    }

    let totalGains = 0;
    Object.keys(currentInventory).forEach(function (key) {
      const prev = lastOreInventory[key] || 0;
      const gain = currentInventory[key] - prev;
      if (gain > 0) totalGains += gain;
    });

    ores.set(currentInventory);
    lastOreInventory = { ...currentInventory };

    if (totalGains > 0) {
      showMiningPopup(totalGains);
      addLogEntry(`Gained ${totalGains} ores while away!`);
    }
  } catch (error) {
    console.error('Failed to sync mining progress:', error);
  }
}

export function initMiningStatus() {
  checkMiningStatus();
}

// Sync herb inventory from backend
export async function syncHerbInventory() {
  try {
    const response = await inventoryAPI.getHerbInventory();
    const data = response.data;

    const herbMap = {};
    if (Array.isArray(data)) {
      data.forEach(function (item) {
        herbMap[item.herb_key] = item.quantity;
      });
    }

    herbs.set(herbMap);
  } catch (error) {
    console.error('Failed to sync herb inventory:', error);
  }
}
