<script>
  import { onMount, onDestroy } from 'svelte';
  import { ores, addLogEntry } from '../stores/game.js';
  import {
    activeMining, miningPopups, miningProgress,
    startMining, stopMining, showMiningPopup,
    syncMiningProgress, syncOreInventoryDuringMining, syncOreInventory
  } from '../stores/mining.js';
  import { miningSyncInterval } from '../stores/config.js';
  import { inventoryAPI } from '../services/api.js';

  let oreTypes = [];
  let inventoryOpen = true;
  let miningInterval = null;
  let syncInterval = null;

  onMount(async function () {
    // Load ore types from master table
    try {
      const resp = await inventoryAPI.getOreTypes();
      oreTypes = resp.data;
    } catch (e) {
      console.error('Failed to load ore types:', e);
    }

    // Always sync inventory first to populate Material Cache
    await syncOreInventory();

    if ($activeMining) {
      await syncOreInventoryDuringMining();
      const ore = oreTypes.find(o => o.ID === $activeMining.oreId);
      if (ore) {
        startMiningPopups(ore);
      }
    }
  });

  function startMiningPopups(ore) {
    if (miningInterval) clearInterval(miningInterval);
    if (syncInterval) clearInterval(syncInterval);

    const interval = ore.MiningTimeMS || 3000;
    miningInterval = setInterval(function () {
      showMiningPopup(1);
    }, interval);

    let syncIntervalMs;
    const unsub = miningSyncInterval.subscribe(v => { syncIntervalMs = v; });
    unsub();

    syncInterval = setInterval(async function () {
      await syncOreInventoryDuringMining();
    }, syncIntervalMs);
  }

  async function handleOreClick(ore) {
    if ($activeMining) {
      // Save name before clearing activeMining
      const oreName = $activeMining.oreName;
      clearInterval(miningInterval);
      clearInterval(syncInterval);
      miningInterval = null;
      syncInterval = null;
      await stopMining();
      addLogEntry(`Stopped extracting ${oreName}.`);
    } else {
      await startMining(ore.ID, ore.OreName, ore.OreKey, ore.MiningTimeMS);
      startMiningPopups(ore);
    }
  }

  // React to activeMining changes (e.g., resumed session from another tab)
  $: if ($activeMining && !miningInterval) {
    const ore = oreTypes.find(o => o.ID === $activeMining.oreId);
    if (ore) startMiningPopups(ore);
  } else if (!$activeMining) {
    clearInterval(miningInterval);
    clearInterval(syncInterval);
    miningInterval = null;
    syncInterval = null;
  }

  onDestroy(function () {
    clearInterval(miningInterval);
    clearInterval(syncInterval);
  });

  function formatInterval(ms) {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(0)}s`;
  }

  function pickaxeLabel(key) {
    if (!key || key === 'none') return null;
    return key.replace('_pickaxe', '').replace('_', ' ') + ' pickaxe';
  }
</script>

<div class="view-mining">
  <div class="page-header">
    <h1 class="page-title">&#x26CF;&#xFE0F; Resource Extraction</h1>
    <p class="page-subtitle">Salvage materials from contaminated zones</p>
  </div>

  <!-- Ore inventory - collapsible -->
  <div class="card inventory-card">
    <button class="card-header collapse-toggle" on:click={() => inventoryOpen = !inventoryOpen} aria-expanded={inventoryOpen}>
      <span class="card-icon">&#x1F4E6;</span>
      <h2 class="card-title">Material Cache</h2>
      <span class="collapse-arrow" class:open={inventoryOpen}>&#x25B8;</span>
    </button>

    {#if inventoryOpen}
      <div class="ore-summary">
        {#if oreTypes.length === 0}
          <p class="loading-text">Loading cache...</p>
        {:else}
          {#each oreTypes as ore}
            <div class="ore-count">
              <span class="ore-icon">{ore.Icon}</span>
              <span class="ore-label">{ore.OreName}</span>
              <span class="ore-qty">{$ores[ore.OreKey] ?? 0}</span>
              {#if ore.MaxQuantity > 0}
                <span class="ore-max">/ {ore.MaxQuantity}</span>
              {/if}
            </div>
          {/each}
        {/if}
      </div>
    {/if}
  </div>

  <!-- Mining area -->
  <div class="card mining-card">
    <div class="card-header">
      <span class="card-icon">&#x26CF;&#xFE0F;</span>
      <h2 class="card-title">
        {#if $activeMining}
          Extracting: {$activeMining.oreName}
        {:else}
          Select Extraction Target
        {/if}
      </h2>
    </div>

    <div class="ore-selection">
      {#if oreTypes.length === 0}
        <p class="loading-text">Loading extraction targets...</p>
      {:else}
        {#each oreTypes as ore}
          {@const isActive = $activeMining?.oreKey === ore.OreKey}
          {@const pickaxe = pickaxeLabel(ore.PickaxeRequired)}
          <button
            class="ore-btn"
            class:active={isActive}
            on:click={() => handleOreClick(ore)}
          >
            <div class="ore-btn-icon">{ore.Icon}</div>
            <div class="ore-btn-info">
              <div class="ore-btn-name">{ore.OreName}</div>
              <div class="ore-btn-meta">
                <span class="badge difficulty">{ore.Difficulty}</span>
                <span class="badge interval">&#x23F3; {formatInterval(ore.MiningTimeMS)}</span>
                {#if pickaxe}
                  <span class="badge pickaxe">&#x26CF; {pickaxe}</span>
                {/if}
              </div>
            </div>
            <div class="ore-btn-status">
              {#if isActive}
                <span class="mining-indicator">&#x23F1;&#xFE0F;</span>
              {:else}
                <span class="ore-btn-qty">{$ores[ore.OreKey] ?? 0}</span>
              {/if}
            </div>
          </button>
        {/each}
      {/if}
    </div>

    {#if $activeMining}
      <div class="mining-info">
        <p class="mining-text">Extracting 1 unit every {formatInterval($activeMining.miningTimeMS)}...</p>

        <div class="progress-bar-container">
          <div class="progress-bar" style="width: {$miningProgress}%"></div>
        </div>
        <p class="progress-text">{Math.round($miningProgress)}%</p>

        <button class="stop-btn" on:click={() => handleOreClick(null)}>
          Abort Extraction
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

  .page-header { padding: 8px 0 4px; }

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

  .collapse-toggle {
    width: 100%;
    background: none;
    border: none;
    color: inherit;
    cursor: pointer;
    padding: 0;
    margin-bottom: 14px;
    font-family: var(--font-body);
  }

  .card-icon { font-size: 18px; }

  .card-title {
    font-family: var(--font-heading);
    font-size: 15px;
    color: var(--color-text-heading);
    margin: 0;
    font-weight: 600;
    flex: 1;
    text-align: left;
  }

  .collapse-arrow {
    font-size: 14px;
    color: var(--color-text-muted);
    display: inline-block;
    transform: rotate(0deg);
    transition: transform 0.2s;
  }

  .collapse-arrow.open {
    transform: rotate(90deg);
  }

  .inventory-card { border-color: var(--color-gold-dim); }

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

  .ore-icon { font-size: 18px; width: 24px; text-align: center; }

  .ore-label { font-size: 13px; color: var(--color-text-muted); flex: 1; }

  .ore-qty { font-size: 14px; font-weight: 600; color: var(--color-gold); }

  .ore-max { font-size: 11px; color: var(--color-text-dim); }

  .ore-selection { display: flex; flex-direction: column; gap: 8px; }

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

  .ore-btn-icon { font-size: 28px; width: 32px; flex-shrink: 0; }

  .ore-btn-info { flex: 1; min-width: 0; }

  .ore-btn-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--color-text-heading);
    margin-bottom: 4px;
  }

  .ore-btn-meta { display: flex; flex-wrap: wrap; gap: 6px; }

  .badge {
    font-size: 11px;
    padding: 2px 7px;
    border-radius: 4px;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.3px;
  }

  .badge.difficulty {
    background-color: rgba(42, 158, 42, 0.15);
    color: var(--color-hazard);
    border: 1px solid rgba(42, 158, 42, 0.3);
  }

  .badge.interval {
    background-color: rgba(200, 168, 75, 0.1);
    color: var(--color-gold-dim);
    border: 1px solid rgba(200, 168, 75, 0.2);
  }

  .badge.pickaxe {
    background-color: rgba(204, 74, 0, 0.12);
    color: var(--color-danger-bright);
    border: 1px solid rgba(204, 74, 0, 0.25);
  }

  .ore-btn-status { flex-shrink: 0; }

  .ore-btn-qty { font-size: 14px; font-weight: 600; color: var(--color-gold); }

  .mining-indicator { font-size: 16px; animation: pulse 1s infinite; }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
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
    letter-spacing: 0.5px;
    text-transform: uppercase;
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
  }

  .progress-text { font-size: 12px; color: var(--color-text-muted); margin-bottom: 12px; }

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

  .stop-btn:hover { background-color: rgba(220, 38, 38, 0.25); }

  .loading-text { color: var(--color-text-muted); font-size: 13px; }

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
    0% { opacity: 1; transform: translateY(0) scale(1); }
    100% { opacity: 0; transform: translateY(-30px) scale(1.1); }
  }
</style>
