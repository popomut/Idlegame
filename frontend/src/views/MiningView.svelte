<script>
  import { onMount, onDestroy } from 'svelte';
  import { ores, addLogEntry } from '../stores/game.js';
  import { activeMining, miningPopups, miningProgress, startMining, stopMining, showMiningPopup, syncMiningProgress, syncOreInventoryDuringMining } from '../stores/mining.js';
  import { miningSyncInterval } from '../stores/config.js';

  const oreTypes = [
    { id: 1, storeKey: 'copperOre', name: 'Copper Ore',   icon: '🪨', color: '#b87333', difficulty: 'Easy' },
    { id: 2, storeKey: 'ironOre',   name: 'Iron Ore',     icon: '⚫', color: '#5a5a5a', difficulty: 'Normal' },
    { id: 3, storeKey: 'goldOre',   name: 'Gold Ore',     icon: '✨', color: '#ffd700', difficulty: 'Hard' },
    { id: 4, storeKey: 'mithrilOre', name: 'Mithril Ore', icon: '💎', color: '#00bfff', difficulty: 'Very Hard' },
    { id: 5, storeKey: 'diamondOre', name: 'Diamond Ore', icon: '💠', color: '#00ffff', difficulty: 'Impossible' },
  ];

  let miningInterval = null;
  let syncInterval = null;

  function startMiningPopups(ore) {
    // Clear any existing intervals
    if (miningInterval) {
      clearInterval(miningInterval);
    }
    if (syncInterval) {
      clearInterval(syncInterval);
    }
    
    // Show popup every 2 seconds (matching mining interval)
    // SECURITY FIX: Only show popup animation, don't increment ores locally
    // Actual ore count comes from backend on stopMining()
    miningInterval = setInterval(function () {
      showMiningPopup(1);
    }, 2000);

    // Get the current sync interval value
    let syncIntervalMs;
    const unsubscribe = miningSyncInterval.subscribe(function (value) {
      syncIntervalMs = value;
    });
    unsubscribe();

    // Sync inventory at configurable interval to show live updates
    syncInterval = setInterval(async function () {
      try {
        await syncOreInventoryDuringMining();
      } catch (error) {
        console.error('Failed to sync inventory during mining:', error);
      }
    }, syncIntervalMs);
  }

  async function handleOreClick(ore) {
    if ($activeMining) {
      await stopMining();
      clearInterval(miningInterval);
      clearInterval(syncInterval);
      addLogEntry(`Stopped mining ${$activeMining.oreName}.`);
    } else {
      await startMining(ore.storeKey, ore.name, ore.id);
      startMiningPopups(ore);
    }
  }

  onMount(async function () {
    // Sync any mining progress that happened while away from this view
    if ($activeMining) {
      // Use the mining status endpoint to get pending earnings
      await syncOreInventoryDuringMining();
      
      const ore = oreTypes.find(o => o.storeKey === $activeMining.oreType);
      if (ore) {
        console.log('Resuming mining popups for:', ore.name);
        addLogEntry(`Resumed mining ${ore.name}...`);
        startMiningPopups(ore);
      }
    } else {
      // If not actively mining, still sync inventory to show latest state
      await syncOreInventoryDuringMining();
    }
  });

  // Watch for changes to activeMining and start/stop popups
  $: if ($activeMining) {
    const ore = oreTypes.find(o => o.storeKey === $activeMining.oreType);
    if (ore && !miningInterval) {
      console.log('Mining active, starting popups for:', ore.name);
      startMiningPopups(ore);
    }
  } else {
    // Mining stopped, clear intervals
    if (miningInterval) {
      clearInterval(miningInterval);
      miningInterval = null;
    }
    if (syncInterval) {
      clearInterval(syncInterval);
      syncInterval = null;
    }
  }

  onDestroy(function () {
    if (miningInterval) {
      clearInterval(miningInterval);
    }
    if (syncInterval) {
      clearInterval(syncInterval);
    }
  });
</script>

<div class="view-mining">
  <div class="page-header">
    <h1 class="page-title">⛏️ Mining</h1>
    <p class="page-subtitle">Extract ores from the earth</p>
  </div>

  <!-- Ore inventory summary -->
  <div class="card inventory-card">
    <div class="card-header">
      <span class="card-icon">📦</span>
      <h2 class="card-title">Ore Inventory</h2>
    </div>
    <div class="ore-summary">
      {#each oreTypes as ore}
        <div class="ore-count">
          <span class="ore-icon">{ore.icon}</span>
          <span class="ore-label">{ore.name}</span>
          <span class="ore-qty">{$ores[ore.storeKey]}</span>
        </div>
      {/each}
    </div>
  </div>

  <!-- Mining area -->
  <div class="card mining-card">
    <div class="card-header">
      <span class="card-icon">⛏️</span>
      <h2 class="card-title">
        {#if $activeMining}
          Mining: {$activeMining.oreName}
        {:else}
          Select an Ore to Mine
        {/if}
      </h2>
    </div>

    <div class="ore-selection">
      {#each oreTypes as ore}
        <button
          class="ore-btn"
          class:active={$activeMining?.oreType === ore.storeKey}
          on:click={() => handleOreClick(ore)}
        >
          <div class="ore-btn-icon">{ore.icon}</div>
          <div class="ore-btn-info">
            <div class="ore-btn-name">{ore.name}</div>
            <div class="ore-btn-difficulty">{ore.difficulty}</div>
          </div>
          <div class="ore-btn-status">
            {#if $activeMining?.oreType === ore.storeKey}
              <span class="mining-indicator">⏱️</span>
            {:else}
              <span class="ore-btn-qty">{$ores[ore.storeKey]}</span>
            {/if}
          </div>
        </button>
      {/each}
    </div>

    {#if $activeMining}
      <div class="mining-info">
        <p class="mining-text">Mining 1 ore every 2 seconds...</p>
        
        <div class="progress-bar-container">
          <div class="progress-bar" style="width: {$miningProgress}%"></div>
        </div>
        <p class="progress-text">{Math.round($miningProgress)}%</p>
        
        <button class="stop-btn" on:click={() => handleOreClick(null)}>
          Stop Mining
        </button>
      </div>
    {/if}
  </div>
</div>

<!-- Mining popups -->
<div class="mining-popups">
  {#each $miningPopups as popup (popup.id)}
    <div class="mining-popup">+{popup.count} Ore</div>
  {/each}
</div>

<style>
  .view-mining {
    padding: 20px 16px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    max-width: 760px;
  }

  .page-header {
    padding: 8px 0 4px;
  }

  .page-title {
    font-family: var(--font-heading);
    font-size: 22px;
    color: var(--color-text-heading);
    margin: 0 0 4px;
    font-weight: 600;
  }

  .page-subtitle {
    font-size: 13px;
    color: var(--color-gold-dim);
    margin: 0;
    letter-spacing: 0.5px;
    text-transform: uppercase;
  }

  .card {
    background-color: var(--color-bg-panel);
    border: 1px solid var(--color-border);
    border-radius: 10px;
    padding: 16px;
  }

  .card-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 14px;
  }

  .card-icon {
    font-size: 18px;
  }

  .card-title {
    font-family: var(--font-heading);
    font-size: 15px;
    color: var(--color-text-heading);
    margin: 0;
    font-weight: 600;
    flex: 1;
  }

  /* Inventory card */
  .inventory-card {
    border-color: var(--color-gold-dim);
  }

  .ore-summary {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }

  .ore-count {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    background-color: var(--color-bg-elevated);
    border: 1px solid var(--color-border-subtle);
    border-radius: 8px;
  }

  .ore-icon {
    font-size: 18px;
    width: 24px;
    text-align: center;
  }

  .ore-label {
    font-size: 13px;
    color: var(--color-text-muted);
    flex: 1;
  }

  .ore-qty {
    font-size: 14px;
    font-weight: 600;
    color: var(--color-gold);
  }

  /* Mining card */
  .mining-card {
    border-color: var(--color-border);
  }

  .ore-selection {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .ore-btn {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    background-color: var(--color-bg-elevated);
    border: 1px solid var(--color-border-subtle);
    border-radius: 8px;
    color: var(--color-text);
    cursor: pointer;
    font-family: var(--font-body);
    transition: all var(--transition-fast);
    text-align: left;
  }

  .ore-btn:hover {
    background-color: var(--color-bg-hover);
    border-color: var(--color-border);
  }

  .ore-btn.active {
    background-color: rgba(200, 168, 75, 0.15);
    border-color: var(--color-gold-dim);
  }

  .ore-btn-icon {
    font-size: 28px;
    width: 32px;
    flex-shrink: 0;
  }

  .ore-btn-info {
    flex: 1;
    min-width: 0;
  }

  .ore-btn-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--color-text-heading);
    margin-bottom: 2px;
  }

  .ore-btn-difficulty {
    font-size: 12px;
    color: var(--color-text-muted);
  }

  .ore-btn-status {
    flex-shrink: 0;
  }

  .ore-btn-qty {
    font-size: 14px;
    font-weight: 600;
    color: var(--color-gold);
  }

  .mining-indicator {
    font-size: 16px;
    animation: pulse 1s infinite;
  }

  @keyframes pulse {
    0%, 100% {
      opacity: 1;
    }
    50% {
      opacity: 0.5;
    }
  }

  .mining-info {
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid var(--color-border-subtle);
    text-align: center;
  }

  .mining-text {
    font-size: 14px;
    color: var(--color-magic-bright);
    margin-bottom: 12px;
  }

  .progress-bar-container {
    width: 100%;
    height: 24px;
    background-color: var(--color-bg-elevated);
    border: 1px solid var(--color-border-subtle);
    border-radius: 12px;
    overflow: hidden;
    margin-bottom: 8px;
  }

  .progress-bar {
    height: 100%;
    background: linear-gradient(90deg, var(--color-magic-bright), var(--color-gold-bright));
    transition: width 0.05s linear;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .progress-text {
    font-size: 12px;
    color: var(--color-text-muted);
    margin-bottom: 12px;
  }

  .stop-btn {
    padding: 8px 20px;
    background-color: rgba(220, 38, 38, 0.15);
    border: 1px solid var(--color-danger);
    border-radius: 6px;
    color: var(--color-danger-bright);
    cursor: pointer;
    font-size: 14px;
    font-family: var(--font-body);
    font-weight: 500;
    transition: background-color var(--transition-fast);
  }

  .stop-btn:hover {
    background-color: rgba(220, 38, 38, 0.25);
  }

  /* Mining popups */
  .mining-popups {
    position: fixed;
    bottom: calc(var(--bottombar-height) + 30px);
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    flex-direction: column;
    gap: 8px;
    z-index: 300;
    pointer-events: none;
  }

  .mining-popup {
    padding: 8px 16px;
    background-color: var(--color-gold);
    color: #000;
    border-radius: 6px;
    font-weight: 600;
    font-size: 14px;
    animation: popup-float 1s ease-out forwards;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
  }

  @keyframes popup-float {
    0% {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
    100% {
      opacity: 0;
      transform: translateY(-30px) scale(1.1);
    }
  }
</style>
