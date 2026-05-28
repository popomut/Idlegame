import { writable } from 'svelte/store';
import axios from 'axios';
import { API_BASE_URL } from '../services/api.js';
import { addLogEntry } from './game.js';

export const activeCrafting = writable(null);
export const craftingPopups = writable([]);
export const offlineCraftingGains = writable(null);
export const tabSwitchCraftingGains = writable(null); // gains while tab was hidden (session continues)
export const isLoadingCrafting = writable(false);
export const craftingProgress = writable(0); // 0-100 progress bar (per ingot cycle)
export const ingotInventory = writable({});

let globalProgressInterval = null;
let isStoppingCrafting = false;

function startGlobalProgressUpdater(craftingTimeMS) {
  if (globalProgressInterval) {
    clearInterval(globalProgressInterval);
  }
  const interval = craftingTimeMS || 5000;

  globalProgressInterval = setInterval(function () {
    activeCrafting.subscribe(function (crafting) {
      if (crafting && crafting.startedAt) {
        const elapsed = Date.now() - new Date(crafting.startedAt).getTime();
        const cycle = elapsed % interval;
        craftingProgress.set((cycle / interval) * 100);
      }
    })();
  }, 50);
}

function stopGlobalProgressUpdater() {
  if (globalProgressInterval) {
    clearInterval(globalProgressInterval);
    globalProgressInterval = null;
  }
  craftingProgress.set(0);
}

activeCrafting.subscribe(function (crafting) {
  if (crafting) {
    startGlobalProgressUpdater(crafting.craftingTimeMS);
  } else {
    stopGlobalProgressUpdater();
  }
});

export async function startCrafting(craftableItemId, recipeName, craftingTimeMS) {
  try {
    isLoadingCrafting.set(true);
    const response = await axios.post(`${API_BASE_URL}/api/blacksmith/start`, 
      { craftable_item_id: craftableItemId },
      { withCredentials: true }
    );

    activeCrafting.set({
      craftableItemId,
      recipeName,
      craftingTimeMS: craftingTimeMS || 5000,
      sessionId: response.data.session_id,
      startedAt: new Date(response.data.started_at),
    });

    addLogEntry(`Started crafting ${recipeName}...`);
  } catch (error) {
    console.error('Failed to start crafting:', error);
    addLogEntry('Failed to start crafting. Try again.');
  } finally {
    isLoadingCrafting.set(false);
  }
}

export async function stopCrafting() {
  try {
    isLoadingCrafting.set(true);
    const response = await axios.post(`${API_BASE_URL}/api/blacksmith/stop`, {}, 
      { withCredentials: true }
    );
    const ingotsProduced = response.data.ingots_produced || 0;
    const xpEarned = response.data.xp_earned || 0;

    activeCrafting.set(null);

    if (ingotsProduced > 0) {
      showCraftingPopup(ingotsProduced);
      addLogEntry(`Crafting complete - produced ${ingotsProduced} ingot(s), earned ${xpEarned} XP!`);
      await syncIngotInventory();
    }
  } catch (error) {
    console.error('Failed to stop crafting:', error);
    addLogEntry('Failed to stop crafting. Try again.');
  } finally {
    isLoadingCrafting.set(false);
  }
}

export async function checkCraftingStatus() {
  try {
    const response = await axios.get(`${API_BASE_URL}/api/blacksmith/status`, 
      { withCredentials: true }
    );
    const status = response.data;

    // Handle offline crafting gains
    if (status.offline_gains && status.offline_gains.was_offline) {
      if (isStoppingCrafting) return;
      isStoppingCrafting = true;
      const gains = status.offline_gains;
      activeCrafting.set(null); // clear first to prevent re-entrant calls
      offlineCraftingGains.set({
        wasOffline: true,
        timeMs: gains.offline_time_ms,
        ingotsGained: gains.ingots_gained,
        recipeName: gains.recipe_name,
      });
      addLogEntry(`You crafted ${gains.ingots_gained} ${gains.recipe_name} while away!`);
      await stopCrafting();
      isStoppingCrafting = false;
      return;
    }

    if (status.is_active) {
      activeCrafting.set({
        craftableItemId: status.craftable_item_id || 0,
        recipeName: status.recipe_name,
        craftingTimeMS: status.crafting_time_ms || 5000,
        startedAt: new Date(status.started_at),
      });
    } else {
      activeCrafting.set(null);
    }

    // Use current_ingots from status (includes pending) not a separate call
    if (status.current_ingots) {
      ingotInventory.set({ ...status.current_ingots });
    } else {
      await syncIngotInventory();
    }
  } catch (error) {
    console.error('Failed to check crafting status:', error);
  }
}

export async function syncIngotInventory() {
  try {
    const response = await axios.get(`${API_BASE_URL}/api/blacksmith/inventory`, 
      { withCredentials: true }
    );
    ingotInventory.set(response.data || {});
  } catch (error) {
    console.error('Failed to sync ingot inventory:', error);
  }
}

// Poll crafting status to update counts with pending server-side estimate (mirrors syncOreInventoryDuringMining)
export async function syncIngotInventoryDuringCrafting() {
  try {
    const response = await axios.get(`${API_BASE_URL}/api/blacksmith/status`, 
      { withCredentials: true }
    );
    const data = response.data;

    if (data.current_ingots) {
      ingotInventory.set({ ...data.current_ingots });
    }
  } catch (error) {
    console.error('Failed to sync ingot inventory during crafting:', error);
  }
}

export function showCraftingPopup(count = 1, itemName = 'Ingot') {
  const id = Date.now();
  craftingPopups.update(function (popups) {
    return [...popups, { id, count, itemName }];
  });
  setTimeout(function () {
    craftingPopups.update(function (popups) {
      return popups.filter(function (p) { return p.id !== id; });
    });
  }, 1500);
}

export function initCraftingStatus() {
  checkCraftingStatus();
}
