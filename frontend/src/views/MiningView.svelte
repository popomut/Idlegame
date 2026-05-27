<script>
  import { onMount, onDestroy } from 'svelte';
  import { ores, herbs, addLogEntry } from '../stores/game.js';
  import { miningSkill } from '../stores/mining_skill.js';
  import {
    activeMining, miningPopups, miningProgress,
    startMining, stopMining, showMiningPopup,
    syncMiningProgress, syncOreInventoryDuringMining, syncOreInventory, syncHerbInventory
  } from '../stores/mining.js';
  import { miningSyncInterval } from '../stores/config.js';
  import { inventoryAPI } from '../services/api.js';

  let extractionTypes = [];
  let selectedExtractionTypeId = null;
  let oreTypes = [];         // current tab's resources (for extraction buttons)
  let allOreTypes = [];      // all ores — for material cache
  let allHerbTypes = [];     // all herbs — for material cache
  let inventoryOpen = true;
  let miningInterval = null;
  let syncInterval = null;
  // Tracks local ore/herb gains between server syncs
  let pendingOres = {};

  onMount(async function () {
    // Load extraction types from master table
    try {
      const resp = await inventoryAPI.getExtractableTypes();
      extractionTypes = resp.data;
      
      // If actively gathering herb, select herb type in dropdown
      if ($activeMining?.resourceType === 'herb') {
        const herbType = extractionTypes.find(t => t.TypeKey === 'herb');
        if (herbType) {
          selectedExtractionTypeId = herbType.ID;
          await loadResourcesByExtractionType(herbType.ID);
        }
      } else if (extractionTypes.length > 0) {
        selectedExtractionTypeId = extractionTypes[0].ID;
        await loadResourcesByExtractionType(selectedExtractionTypeId);
      }
    } catch (e) {
      console.error('Failed to load extraction types:', e);
      try {
        const resp = await inventoryAPI.getOreTypes();
        oreTypes = resp.data;
      } catch (e2) {
        console.error('Failed to load ore types:', e2);
      }
    }

    // Load ALL resource types for material cache (independent of dropdown selection)
    try {
      const [oreResp, herbResp] = await Promise.all([
        inventoryAPI.getOreTypes(),
        inventoryAPI.getHerbTypes(),
      ]);
      allOreTypes = oreResp.data;
      allHerbTypes = herbResp.data;
    } catch (e) {
      console.error('Failed to load all resource types for cache:', e);
    }

    // Always sync inventory first to populate Material Cache
    await syncOreInventory();
    await syncHerbInventory();

    if ($activeMining) {
      if ($activeMining.resourceType === 'herb') {
        await syncHerbInventory();
      } else {
        await syncOreInventoryDuringMining();
      }
      // Find the active resource — check oreTypes first, then allHerbTypes fallback
      const resource = oreTypes.find(o => o.ID === $activeMining.oreId)
        ?? allHerbTypes.find(h => h.ID === $activeMining.oreId)
        ?? allOreTypes.find(o => o.ID === $activeMining.oreId);
      if (resource) {
        startMiningPopups(resource); // resets pendingOres internally
        // Re-apply estimated pending after startMiningPopups resets it
        if ($activeMining.resourceType === 'herb') {
          const elapsed = Date.now() - new Date($activeMining.startedAt).getTime();
          const interval = $activeMining.extractionTimeMS || 3000;
          const estimated = Math.floor(elapsed / interval);
          if (estimated > 0) {
            pendingOres = { [$activeMining.oreKey]: estimated };
          }
        }
      }
    }
  });

  async function loadResourcesByExtractionType(typeId) {
    try {
      const extractionType = extractionTypes.find(t => t.ID === typeId);
      if (!extractionType) {
        console.error('Extraction type not found:', typeId);
        return;
      }

      let resp;
      if (extractionType.TypeKey === 'ore') {
        resp = await inventoryAPI.getOreTypes(typeId);
      } else if (extractionType.TypeKey === 'herb') {
        resp = await inventoryAPI.getHerbTypes(typeId);
      } else {
        console.error('Unknown extraction type:', extractionType.TypeKey);
        return;
      }
      
      oreTypes = resp.data;
    } catch (e) {
      console.error('Failed to load resources for extraction type:', e);
    }
  }

  function handleExtractionTypeChange(e) {
    selectedExtractionTypeId = parseInt(e.target.value, 10);
    loadResourcesByExtractionType(selectedExtractionTypeId);
  }

  function startMiningPopups(ore) {
    if (miningInterval) clearInterval(miningInterval);
    if (syncInterval) clearInterval(syncInterval);

    const interval = ore.MiningTimeMS ?? ore.GatherTimeMS ?? 3000;
    const resourceKey = ore.OreKey ?? ore.HerbKey;
    const isHerb = !!ore.HerbKey;
    const maxQty = ore.MaxQuantity || 0;

    // Reset pending on new session so display starts clean
    pendingOres = {};

    miningInterval = setInterval(function () {
      // Check cap against base (server) + pending — use correct store
      const base = isHerb ? ($herbs[resourceKey] || 0) : ($ores[resourceKey] || 0);
      const pending = (pendingOres[resourceKey] || 0);
      if (maxQty > 0 && base + pending >= maxQty) return;

      pendingOres = { ...pendingOres, [resourceKey]: pending + 1 };
      showMiningPopup(1);
    }, interval);

    let syncIntervalMs;
    const unsub = miningSyncInterval.subscribe(v => { syncIntervalMs = v; });
    unsub();

    // Sync with server every 15s
    syncInterval = setInterval(async function () {
      if (isHerb) {
        await syncHerbInventory();
      } else {
        await syncOreInventoryDuringMining();
      }
      pendingOres = {};
    }, syncIntervalMs);
  }

  async function handleOreClick(ore) {
    // Check if ore is unlocked by level (only if ore exists and not aborting)
    if (ore && ore.LevelRequired && ore.LevelRequired > $miningSkill.level) {
      addLogEntry(`🔒 Level ${ore.LevelRequired} required to mine ${ore.OreName}`);
      return;
    }

    if ($activeMining) {
      // Save name before clearing activeMining
      const oreName = $activeMining.oreName;
      clearInterval(miningInterval);
      clearInterval(syncInterval);
      miningInterval = null;
      syncInterval = null;
      pendingOres = {};
      await stopMining();
      addLogEntry(`Stopped extracting ${oreName}.`);
    } else if (ore) {
      pendingOres = {};
      const extractionTime = ore.MiningTimeMS ?? ore.GatherTimeMS ?? 3000;
      const resourceKey = ore.OreKey ?? ore.HerbKey;
      const resourceName = ore.OreName ?? ore.HerbName;
      const resourceType = ore.HerbKey ? 'herb' : 'ore';
      await startMining(ore.ID, resourceName, resourceKey, extractionTime, resourceType);
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

  <!-- Extraction type selector -->
  {#if extractionTypes.length > 0}
    <div class="extraction-controls">
      <label for="extraction-type-select" class="control-label">Extraction Type:</label>
      <select
        id="extraction-type-select"
        value={selectedExtractionTypeId}
        on:change={handleExtractionTypeChange}
        class="type-select"
      >
        {#each extractionTypes as type (type.ID)}
          <option value={type.ID}>{type.TypeName}</option>
        {/each}
      </select>
    </div>
  {/if}

  <!-- Ore inventory - collapsible -->
  <div class="card inventory-card">
    <button class="card-header collapse-toggle" on:click={() => inventoryOpen = !inventoryOpen} aria-expanded={inventoryOpen}>
      <span class="card-icon">&#x1F4E6;</span>
      <h2 class="card-title">Material Cache</h2>
      <span class="collapse-arrow" class:open={inventoryOpen}>&#x25B8;</span>
    </button>

    {#if inventoryOpen}
      <div class="ore-summary">
        {#if allOreTypes.length === 0 && allHerbTypes.length === 0}
          <p class="loading-text">Loading cache...</p>
        {:else}
          {@const allResources = [
            ...allOreTypes.map(o => ({ ...o, _isHerb: false, _key: o.OreKey })),
            ...allHerbTypes.map(h => ({ ...h, _isHerb: true, _key: h.HerbKey })),
          ]}
          {@const collectedResources = allResources.filter(r => {
            const storeVal = r._isHerb ? ($herbs[r._key] ?? 0) : ($ores[r._key] ?? 0);
            return storeVal + (pendingOres[r._key] ?? 0) > 0;
          })}
          {#if collectedResources.length === 0}
            <p class="loading-text">No materials collected yet.</p>
          {:else}
            {#each collectedResources as r}
              <div class="ore-count">
                <div class="ore-icon">
                  {#if r.SVG}
                    <div class="ore-svg-small">{@html r.SVG}</div>
                  {:else}
                    <span>{r.Icon}</span>
                  {/if}
                </div>
                <span class="ore-label">{r.OreName ?? r.HerbName}</span>
                <span class="ore-qty">{((r._isHerb ? $herbs[r._key] : $ores[r._key]) ?? 0) + (pendingOres[r._key] ?? 0)}</span>
                {#if r.MaxQuantity > 0}
                  <span class="ore-max">/ {r.MaxQuantity}</span>
                {/if}
              </div>
            {/each}
          {/if}
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
          {@const resourceKey = ore.OreKey ?? ore.HerbKey}
          {@const isActive = $activeMining?.oreKey === resourceKey}
          {@const pickaxe = pickaxeLabel(ore.PickaxeRequired)}
          {@const isLocked = ore.LevelRequired > $miningSkill.level}
          <button
            class="ore-btn"
            class:active={isActive}
            class:locked={isLocked}
            disabled={isLocked}
            on:click={() => handleOreClick(ore)}
          >
            <div class="ore-btn-icon">
              {#if isLocked}
                🔒
              {:else if ore.SVG}
                <div class="ore-svg-large">{@html ore.SVG}</div>
              {:else}
                {ore.Icon}
              {/if}
            </div>
            <div class="ore-btn-info">
              <div class="ore-btn-name">{ore.OreName ?? ore.HerbName}</div>
              <div class="ore-btn-meta">
                {#if isLocked}
                  <span class="badge locked-badge">🔒 Level {ore.LevelRequired}</span>
                {:else}
                  <span class="badge difficulty">{ore.Difficulty}</span>
                  <span class="badge interval">&#x23F3; {formatInterval(ore.MiningTimeMS ?? ore.GatherTimeMS)}</span>
                  {#if pickaxe}
                    <span class="badge pickaxe">&#x26CF; {pickaxe}</span>
                  {/if}
                {/if}
              </div>
            </div>
            <div class="ore-btn-status">
              {#if isActive}
                <span class="mining-indicator">&#x23F1;&#xFE0F;</span>
              {:else if !isLocked}
                <span class="ore-btn-qty">{((ore.OreKey ? $ores[resourceKey] : $herbs[resourceKey]) ?? 0) + (pendingOres[resourceKey] ?? 0)}</span>
              {/if}
            </div>
          </button>
        {/each}
      {/if}
    </div>

    {#if $activeMining}
      <div class="mining-info">
        <p class="mining-text">Extracting 1 unit every {formatInterval($activeMining.extractionTimeMS)}...</p>

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

  .extraction-controls {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    background: var(--color-bg-panel);
    border: 1px solid var(--color-border);
    border-radius: 8px;
  }

  .control-label {
    font-size: 14px;
    font-weight: 600;
    color: var(--color-text-heading);
    white-space: nowrap;
  }

  .type-select {
    flex: 1;
    padding: 8px 12px;
    background: var(--color-bg);
    color: var(--color-text);
    border: 1px solid var(--color-border);
    border-radius: 6px;
    font-size: 14px;
    cursor: pointer;
    transition: border-color 0.2s, background 0.2s;
  }

  .type-select:hover {
    border-color: var(--color-gold-dim);
  }

  .type-select:focus {
    outline: none;
    border-color: var(--color-gold-bright);
    background: var(--color-bg-panel);
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

  .ore-icon { font-size: 18px; width: 24px; height: 24px; text-align: center; display: flex; align-items: center; justify-content: center; }

  .ore-svg-small {
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .ore-svg-small :global(svg) {
    width: 100%;
    height: 100%;
  }

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

  .ore-btn.locked {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .ore-btn.locked:hover {
    background-color: var(--color-bg-elevated);
    border-color: var(--color-border-subtle);
  }

  .ore-btn-icon { font-size: 28px; width: 48px; height: 48px; flex-shrink: 0; display: flex; align-items: center; justify-content: center; }

  .ore-svg-large {
    width: 48px;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .ore-svg-large :global(svg) {
    width: 100%;
    height: 100%;
  }

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

  .badge.locked-badge {
    background-color: rgba(255, 100, 0, 0.15);
    color: #ff6400;
    border: 1px solid rgba(255, 100, 0, 0.3);
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
