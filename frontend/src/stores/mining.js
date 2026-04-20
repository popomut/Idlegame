import { writable } from 'svelte/store';
import { miningAPI, inventoryAPI } from '../services/api.js';
import { ores, addLogEntry } from './game.js';

export const activeMining = writable(null);
export const miningPopups = writable([]);
export const offlineGains = writable(null);
export const isLoadingMining = writable(false);
export const miningProgress = writable(0); // Progress bar (0-100)

// Global progress update loop
let globalProgressInterval = null;

function startGlobalProgressUpdater() {
  // Clear any existing interval
  if (globalProgressInterval) {
    clearInterval(globalProgressInterval);
  }

  // Update progress every 50ms based on elapsed time
  globalProgressInterval = setInterval(function () {
    activeMining.subscribe(function (mining) {
      if (mining && mining.startedAt) {
        // Calculate progress based on time elapsed
        const now = new Date();
        const startedAt = new Date(mining.startedAt);
        const elapsedMs = now.getTime() - startedAt.getTime();
        
        // Mining takes 2 seconds per ore (2000ms)
        // Progress should cycle: 0% to 100% every 2 seconds
        const cyclePosition = elapsedMs % 2000;
        const progress = (cyclePosition / 2000) * 100;
        
        miningProgress.set(Math.min(progress, 100));
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

// Watch for mining state changes
activeMining.subscribe(function (mining) {
  if (mining) {
    startGlobalProgressUpdater();
  } else {
    stopGlobalProgressUpdater();
  }
});

let lastOreInventory = {
  copperOre: 0,
  ironOre: 0,
  goldOre: 0,
  mithrilOre: 0,
  diamondOre: 0,
};

export async function startMining(oreType, oreName, oreId) {
  try {
    isLoadingMining.set(true);
    const response = await miningAPI.startMining(oreId);
    
    activeMining.set({ 
      oreType, 
      oreName,
      sessionId: response.data.session_id,
      startedAt: new Date(response.data.started_at),
    });
    
    addLogEntry(`Started mining ${oreName}...`);
  } catch (error) {
    console.error('Failed to start mining:', error);
    addLogEntry('❌ Failed to start mining. Try again.');
  } finally {
    isLoadingMining.set(false);
  }
}

export async function stopMining() {
  try {
    isLoadingMining.set(true);
    const response = await miningAPI.stopMining();
    const oresGained = response.data.ores_gained;
    
    activeMining.set(null);
    
    if (oresGained > 0) {
      showMiningPopup(oresGained);
      addLogEntry(`⛏️  Earned ${oresGained} ores!`);
      
      // Sync inventory with backend
      await syncOreInventory();
    }
  } catch (error) {
    console.error('Failed to stop mining:', error);
    addLogEntry('❌ Failed to stop mining. Try again.');
  } finally {
    isLoadingMining.set(false);
  }
}

export async function checkMiningStatus() {
  try {
    const response = await miningAPI.getMiningStatus();
    const status = response.data;
    
    // Check for offline gains
    if (status.offline_gains && status.offline_gains.was_offline) {
      const gains = status.offline_gains;
      offlineGains.set({
        wasOffline: true,
        timeMs: gains.offline_time_ms,
        oresGained: gains.ores_gained,
        oreName: gains.ore_name,
      });
      
      addLogEntry(
        `⚡ You earned ${gains.ores_gained} ${gains.ore_name} while away!`
      );
      
      // IMPORTANT: Stop mining to actually save the ores to inventory!
      await stopMining();
      return;
    }
    
    // Update mining status
    if (status.is_active && status.current_ore) {
      activeMining.set({
        oreType: status.current_ore.ore_key,
        oreName: status.current_ore.ore_name,
        startedAt: new Date(status.started_at),
      });
    } else {
      activeMining.set(null);
    }
    
    // Sync ores from backend
    await syncOreInventory();
  } catch (error) {
    console.error('Failed to check mining status:', error);
  }
}

export async function syncOreInventoryDuringMining() {
  try {
    // Get mining status which includes pending ore earnings
    const response = await miningAPI.getMiningStatus();
    const data = response.data;
    
    ores.set({
      copperOre: data.current_ores.copper_ore || 0,
      ironOre: data.current_ores.iron_ore || 0,
      goldOre: data.current_ores.gold_ore || 0,
      mithrilOre: data.current_ores.mithril_ore || 0,
      diamondOre: data.current_ores.diamond_ore || 0,
    });
  } catch (error) {
    console.error('Failed to sync ore inventory during mining:', error);
  }
}

export async function syncOreInventory() {
  try {
    const response = await inventoryAPI.getOreInventory();
    const data = response.data;
    
    ores.set({
      copperOre: data.copper_ore || 0,
      ironOre: data.iron_ore || 0,
      goldOre: data.gold_ore || 0,
      mithrilOre: data.mithril_ore || 0,
      diamondOre: data.diamond_ore || 0,
    });
  } catch (error) {
    console.error('Failed to sync ore inventory:', error);
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

// Sync mining gains while away from Mining view
export async function syncMiningProgress() {
  try {
    // Get current inventory from backend
    const response = await inventoryAPI.getOreInventory();
    const data = response.data;
    
    const currentInventory = {
      copperOre: data.copper_ore || 0,
      ironOre: data.iron_ore || 0,
      goldOre: data.gold_ore || 0,
      mithrilOre: data.mithril_ore || 0,
      diamondOre: data.diamond_ore || 0,
    };
    
    // Calculate differences
    const gains = {
      copperOre: currentInventory.copperOre - lastOreInventory.copperOre,
      ironOre: currentInventory.ironOre - lastOreInventory.ironOre,
      goldOre: currentInventory.goldOre - lastOreInventory.goldOre,
      mithrilOre: currentInventory.mithrilOre - lastOreInventory.mithrilOre,
      diamondOre: currentInventory.diamondOre - lastOreInventory.diamondOre,
    };
    
    // Update local store with current backend values
    ores.set(currentInventory);
    lastOreInventory = currentInventory;
    
    // Show popups for any gains
    const totalGains = gains.copperOre + gains.ironOre + gains.goldOre + gains.mithrilOre + gains.diamondOre;
    if (totalGains > 0) {
      showMiningPopup(totalGains);
      addLogEntry(`⛏️  Earned ${totalGains} ores while away!`);
    }
  } catch (error) {
    console.error('Failed to sync mining progress:', error);
  }
}

// Call this on app startup to check offline gains
export function initMiningStatus() {
  checkMiningStatus();
}
