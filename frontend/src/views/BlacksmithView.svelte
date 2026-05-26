<script>
  import { onMount, onDestroy } from 'svelte';
  import { blacksmithSkill } from '../stores/blacksmith_skill.js';
  import { 
    activeCrafting, craftingPopups, craftingProgress,
    startCrafting, stopCrafting,
    syncIngotInventory, syncIngotInventoryDuringCrafting, ingotInventory, checkCraftingStatus, showCraftingPopup
  } from '../stores/blacksmith.js';
  import { addLogEntry } from '../stores/game.js';
  import axios from 'axios';
  import { API_BASE_URL } from '../services/api.js';

  let recipes = [];
  let inventoryOpen = true;
  let craftingInterval = null;
  let syncInterval = null;
  let pendingIngots = {}; // Tracks ingots produced in current cycle before sync
  let oreInventory = {}; // ore_key → quantity for ingredient availability display

  async function fetchOreInventory() {
    try {
      const resp = await axios.get(`${API_BASE_URL}/api/inventory/ores`, { withCredentials: true });
      if (Array.isArray(resp.data)) {
        const map = {};
        resp.data.forEach(item => { map[item.ore_key] = item.quantity; });
        oreInventory = map;
      }
    } catch (e) {
      console.error('Failed to load ore inventory:', e);
    }
  }

  onMount(async function () {
    try {
      const resp = await axios.get(`${API_BASE_URL}/api/blacksmith/recipes`, {
        withCredentials: true,
      });
      recipes = resp.data;
    } catch (e) {
      console.error('Failed to load recipes:', e);
    }

    // Always sync inventory first to populate Material Cache + ingredient availability
    await Promise.all([syncIngotInventory(), fetchOreInventory()]);

    if ($activeCrafting) {
      // If resuming an active session, sync with server first to get pending ingots
      await syncIngotInventoryDuringCrafting();
      const recipe = recipes.find(r => r.id === $activeCrafting.craftableItemId);
      if (recipe) {
        const maxCrafts = await computeMaxCrafts(recipe);
        startCraftingUI(recipe, maxCrafts);
      } else {
        // Recipe not found - stop crafting to prevent ghost sessions
        await stopCrafting();
        $activeCrafting = null;
        addLogEntry(`⚠️ Recipe no longer available - stopped crafting.`);
      }
    }
  });

  onDestroy(function () {
    clearInterval(craftingInterval);
    clearInterval(syncInterval);
  });

  // Returns how many times recipe can be crafted based on current inventory.
  // Infinity if recipe has no ingredients (no natural limit).
  async function computeMaxCrafts(recipe) {
    if (!recipe.ingredients || recipe.ingredients.length === 0) return Infinity;

    let oreMap = {};
    const hasOreIngredient = recipe.ingredients.some(i => i.ingredient_type === 'ore');
    if (hasOreIngredient) {
      try {
        const oreResp = await axios.get(`${API_BASE_URL}/api/inventory/ores`, { withCredentials: true });
        if (Array.isArray(oreResp.data)) {
          oreResp.data.forEach(item => { oreMap[item.ore_key] = item.quantity; });
        }
      } catch (e) {
        console.error('Failed to fetch ore inventory for maxCrafts:', e);
      }
    }

    let maxCrafts = Infinity;
    for (const ing of recipe.ingredients) {
      let available = 0;
      if (ing.ingredient_type === 'ore') {
        available = oreMap[ing.ingredient_key] || 0;
      } else if (ing.ingredient_type === 'ingot') {
        available = $ingotInventory[ing.ingredient_key] || 0;
      }
      maxCrafts = Math.min(maxCrafts, Math.floor(available / (ing.quantity_required || 1)));
    }
    return maxCrafts;
  }

  async function handleRecipeClick(recipe) {
    if (!recipe) {
      if ($activeCrafting) {
        const recipeName = $activeCrafting.recipeName;
        clearInterval(craftingInterval);
        clearInterval(syncInterval);
        craftingInterval = null;
        syncInterval = null;
        pendingIngots = {};
        await stopCrafting();
        addLogEntry(`Stopped crafting ${recipeName}.`);
      }
      return;
    }

    if (recipe.level_required && recipe.level_required > $blacksmithSkill.level) {
      addLogEntry(`🔒 Level ${recipe.level_required} required to craft ${recipe.name}`);
      return;
    }

    if ($activeCrafting) {
      const currentName = $activeCrafting.recipeName;
      clearInterval(craftingInterval);
      clearInterval(syncInterval);
      craftingInterval = null;
      syncInterval = null;
      pendingIngots = {};
      await stopCrafting();
      addLogEntry(`Stopped crafting ${currentName}.`);
    } else {
      const maxCrafts = await computeMaxCrafts(recipe);

      if (maxCrafts <= 0) {
        // Find the first missing ingredient for the error message
        let missingIngredient = 'unknown ingredient';
        for (const ing of recipe.ingredients) {
          let available = 0;
          if (ing.ingredient_type === 'ore') {
            // oreMap already fetched inside computeMaxCrafts — re-check via ingotInventory approach not possible here
            // Just report the ingredient key
            available = 0;
          } else if (ing.ingredient_type === 'ingot') {
            available = $ingotInventory[ing.ingredient_key] || 0;
          }
          if (available < (ing.quantity_required || 1)) {
            missingIngredient = `${ing.ingredient_key} (have: ${available}, need: ${ing.quantity_required})`;
            break;
          }
        }
        addLogEntry(`❌ Not enough ingredients: ${missingIngredient}`);
        return;
      }

      pendingIngots = {};
      await startCrafting(recipe.id, recipe.name, recipe.crafting_time_ms);
      startCraftingUI(recipe, maxCrafts);
    }
  }

  function startCraftingUI(recipe, maxCrafts = Infinity) {
    if (craftingInterval) clearInterval(craftingInterval);
    if (syncInterval) clearInterval(syncInterval);

    const itemKey = recipe.item_key || recipe.ItemKey;
    const interval = recipe.crafting_time_ms || 3000;

    // Reset pending on new session so display starts clean
    pendingIngots = {};

    let remainingCrafts = maxCrafts;

    // Fast loop: Update pendingIngots and show popup every crafting cycle
    craftingInterval = setInterval(async function () {
      // Safety check: ensure recipe still exists and has ingredients
      if (!recipe || !recipe.ingredients || recipe.ingredients.length === 0) {
        clearInterval(craftingInterval);
        clearInterval(syncInterval);
        craftingInterval = null;
        syncInterval = null;
        stopCrafting().catch(err => console.error('Failed to auto-stop crafting:', err));
        addLogEntry(`⚠️ Recipe no longer available - stopped crafting.`);
        return;
      }

      // Stop when we've exhausted all craftable cycles
      if (remainingCrafts <= 0) {
        clearInterval(craftingInterval);
        clearInterval(syncInterval);
        craftingInterval = null;
        syncInterval = null;
        pendingIngots = {};
        await stopCrafting();
        addLogEntry(`⚠️ Stopped crafting - not enough ingredients.`);
        return;
      }

      remainingCrafts--;

      // Svelte tracks let assignments - this reliably triggers re-render
      const currentPending = (pendingIngots[itemKey] || 0);
      pendingIngots = { ...pendingIngots, [itemKey]: currentPending + 1 };
      showCraftingPopup(1, itemKey);

      // Check server state every cycle — catches ingredient depletion immediately
      await syncCraftingProgressDuringCrafting(recipe.name);
    }, interval);

    // Keep a fallback sync in case the cycle check misses a tick
    syncInterval = setInterval(async function () {
      await syncCraftingProgressDuringCrafting(recipe.name);
    }, 15000);
  }

  async function syncCraftingProgressDuringCrafting(recipeName) {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/blacksmith/status`, 
        { withCredentials: true }
      );
      const status = response.data;

      // Update ingot inventory with server values and reset pending
      if (status.current_ingots) {
        $ingotInventory = { ...status.current_ingots };
        // Server now has the true count - reset pending
        pendingIngots = {};
        // Also refresh ore inventory since ores may have been consumed
        fetchOreInventory();
      }

      // If status shows inactive, server auto-stopped us (no ingredients)
      if (!status.is_active) {
        clearInterval(craftingInterval);
        clearInterval(syncInterval);
        craftingInterval = null;
        syncInterval = null;
        pendingIngots = {};
        $activeCrafting = null;
        addLogEntry(`⚠️ Crafting stopped - not enough ingredients.`);
      }
    } catch (error) {
      console.error('Failed to sync crafting progress:', error);
    }
  }

  // React to activeCrafting changes (e.g., resumed session from another tab)
  $: if ($activeCrafting && !craftingInterval) {
    const recipe = recipes.find(r => r.id === $activeCrafting.craftableItemId);
    if (recipe) {
      computeMaxCrafts(recipe).then(maxCrafts => startCraftingUI(recipe, maxCrafts));
    }
  } else if (!$activeCrafting) {
    clearInterval(craftingInterval);
    clearInterval(syncInterval);
    craftingInterval = null;
    syncInterval = null;
  }

  function formatIngredient(ing) {
    return `${ing.quantity_required}x ${ing.ingredient_key}`;
  }

  // Map from item_key → max_quantity for Material Cache display
  $: maxQtyMap = Object.fromEntries(
    recipes.filter(r => r.max_quantity > 0).map(r => [r.item_key, r.max_quantity])
  );

  function formatInterval(ms) {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(0)}s`;
  }
</script>

<div class="view-blacksmith">
  <div class="page-header">
    <h1 class="page-title">⚒️ Blacksmith Forge</h1>
    <p class="page-subtitle">Craft equipment and materials</p>
  </div>

  <!-- Ingot inventory - collapsible -->
  <div class="card inventory-card">
    <button class="card-header collapse-toggle" on:click={() => inventoryOpen = !inventoryOpen} aria-expanded={inventoryOpen}>
      <span class="card-icon">📦</span>
      <h2 class="card-title">Material Cache</h2>
      <span class="collapse-arrow" class:open={inventoryOpen}>▸</span>
    </button>

    {#if inventoryOpen}
      <div class="ingot-summary">
        {#if Object.keys($ingotInventory).length === 0 && Object.keys(pendingIngots).length === 0}
          <p class="loading-text">No ingots crafted yet</p>
        {:else}
          {#each Object.entries($ingotInventory) as [key, qty]}
            <div class="ingot-count">
              <span class="ingot-icon">⚒️</span>
              <span class="ingot-label">{key}</span>
              <span class="ingot-qty">{qty + (pendingIngots[key] ?? 0)}</span>
              {#if maxQtyMap[key]}
                <span class="ingot-max">/ {maxQtyMap[key]}</span>
              {/if}
            </div>
          {/each}
          {#each Object.entries(pendingIngots) as [key, qty]}
            {#if !$ingotInventory[key]}
              <div class="ingot-count">
                <span class="ingot-icon">⚒️</span>
                <span class="ingot-label">{key}</span>
                <span class="ingot-qty">{qty}</span>
                {#if maxQtyMap[key]}
                  <span class="ingot-max">/ {maxQtyMap[key]}</span>
                {/if}
              </div>
            {/if}
          {/each}
        {/if}
      </div>
    {/if}
  </div>

  <!-- Crafting area -->
  <div class="card crafting-card">
    <div class="card-header">
      <span class="card-icon">⚒️</span>
      <h2 class="card-title">
        {#if $activeCrafting}
          Crafting: {$activeCrafting.recipeName}
        {:else}
          Select Recipe
        {/if}
      </h2>
    </div>

    <div class="recipe-selection">
      {#if recipes.length === 0}
        <p class="loading-text">Loading recipes...</p>
      {:else}
        {#each recipes as recipe}
          {@const isActive = $activeCrafting?.craftableItemId === recipe.id}
          {@const isLocked = recipe.level_required > $blacksmithSkill.level}
          <button
            class="recipe-btn"
            class:active={isActive}
            class:locked={isLocked}
            disabled={isLocked}
            on:click={() => handleRecipeClick(recipe)}
          >
            <div class="recipe-btn-icon">{isLocked ? '🔒' : recipe.icon}</div>
            <div class="recipe-btn-info">
              <div class="recipe-btn-name">{recipe.name}</div>
              <div class="recipe-btn-ingredients">
                {#each recipe.ingredients as ing}
                  <span class="ingredient-badge">{formatIngredient(ing)}</span>
                {/each}
              </div>
              <div class="recipe-btn-meta">
                {#if isLocked}
                  <span class="badge locked-badge">🔒 Level {recipe.level_required}</span>
                {:else}
                  <span class="badge time">⏱️ {formatInterval(recipe.crafting_time_ms)}</span>
                  <span class="badge xp">⭐ {recipe.xp_per_craft} XP</span>
                {/if}
              </div>
            </div>
            <div class="recipe-btn-status">
              {#if isActive}
                <span class="crafting-indicator">⏳</span>
              {/if}
            </div>
            {#if !isLocked && recipe.ingredients && recipe.ingredients.length > 0}
              <div class="recipe-ing-status">
                {#each recipe.ingredients as ing}
                  {@const have = ing.ingredient_type === 'ore'
                    ? (oreInventory[ing.ingredient_key] || 0)
                    : ($ingotInventory[ing.ingredient_key] || 0)}
                  {@const enough = have >= ing.quantity_required}
                  <div class="ing-avail" class:ing-avail-low={!enough}>
                    <span class="ing-avail-count">{have}/{ing.quantity_required}</span>
                    <span class="ing-avail-name">{ing.ingredient_key}</span>
                  </div>
                {/each}
              </div>
            {/if}
          </button>
        {/each}
      {/if}
    </div>

    {#if $activeCrafting}
      <div class="crafting-info">
        <p class="crafting-text">Crafting in progress...</p>

        <div class="progress-bar-container">
          <div class="progress-bar" style="width: {$craftingProgress}%"></div>
        </div>
        <p class="progress-text">{Math.round($craftingProgress)}%</p>

        <button class="stop-btn" on:click={() => handleRecipeClick(null)}>
          Stop Crafting
        </button>
      </div>
    {/if}
  </div>
</div>

<!-- Crafting popups -->
<div class="crafting-popups">
  {#each $craftingPopups as popup (popup.id)}
    <div class="crafting-popup">+{popup.count} {popup.itemName}</div>
  {/each}
</div>

<style>
  .view-blacksmith {
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

  .inventory-card {
    border-color: var(--color-gold-dim);
  }

  .ingot-summary {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }

  .ingot-count {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    background-color: var(--color-bg-elevated);
    border: 1px solid var(--color-border-subtle);
    border-radius: 8px;
  }

  .ingot-icon {
    font-size: 18px;
    width: 24px;
    text-align: center;
  }

  .ingot-label {
    font-size: 13px;
    color: var(--color-text-muted);
    flex: 1;
  }

  .ingot-qty {
    font-size: 14px;
    font-weight: 600;
    color: var(--color-gold);
  }

  .ingot-max {
    font-size: 11px;
    color: var(--color-text-dim);
  }

  .loading-text {
    font-size: 13px;
    color: var(--color-text-muted);
    margin: 0;
    text-align: center;
    padding: 20px 0;
  }

  .recipe-selection {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .recipe-btn {
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

  .recipe-btn:hover {
    background-color: var(--color-bg-hover);
    border-color: var(--color-border);
  }

  .recipe-btn.active {
    background-color: rgba(204, 120, 30, 0.15);
    border-color: var(--color-danger-bright);
  }

  .recipe-btn.locked {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .recipe-btn.locked:hover {
    background-color: var(--color-bg-elevated);
    border-color: var(--color-border-subtle);
  }

  .recipe-btn-icon {
    font-size: 28px;
    width: 32px;
    flex-shrink: 0;
  }

  .recipe-btn-info {
    flex: 1;
    min-width: 0;
  }

  .recipe-btn-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--color-text-heading);
    margin-bottom: 4px;
  }

  .recipe-btn-ingredients {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-bottom: 4px;
  }

  .ingredient-badge {
    font-size: 11px;
    padding: 2px 7px;
    border-radius: 4px;
    background-color: rgba(204, 120, 30, 0.15);
    color: var(--color-danger-bright);
    border: 1px solid rgba(204, 120, 30, 0.3);
    font-weight: 500;
  }

  .recipe-btn-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .badge {
    font-size: 11px;
    padding: 2px 7px;
    border-radius: 4px;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.3px;
  }

  .badge.time {
    background-color: rgba(200, 168, 75, 0.1);
    color: var(--color-gold-dim);
    border: 1px solid rgba(200, 168, 75, 0.2);
  }

  .badge.xp {
    background-color: rgba(42, 158, 42, 0.15);
    color: var(--color-hazard);
    border: 1px solid rgba(42, 158, 42, 0.3);
  }

  .badge.locked-badge {
    background-color: rgba(255, 100, 0, 0.15);
    color: #ff6400;
    border: 1px solid rgba(255, 100, 0, 0.3);
  }

  .recipe-btn-status {
    flex-shrink: 0;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .recipe-ing-status {
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
    align-items: flex-end;
    min-width: 80px;
  }

  .ing-avail {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 1px;
  }

  .ing-avail-count {
    font-size: 12px;
    font-weight: 700;
    color: var(--color-hazard);
    line-height: 1.2;
  }

  .ing-avail-name {
    font-size: 10px;
    color: var(--color-text-muted);
    line-height: 1.2;
    max-width: 90px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-align: right;
  }

  .ing-avail.ing-avail-low .ing-avail-count {
    color: var(--color-text-muted);
    opacity: 0.6;
  }

  .ing-avail.ing-avail-low .ing-avail-name {
    opacity: 0.5;
  }

  .crafting-indicator {
    font-size: 20px;
  }

  .crafting-card {
    border-color: var(--color-danger-bright);
  }

  .crafting-info {
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid var(--color-border-subtle);
  }

  .crafting-text {
    font-size: 13px;
    color: var(--color-text-muted);
    margin: 0 0 12px;
  }

  .progress-bar-container {
    height: 24px;
    background-color: var(--color-bg-elevated);
    border: 1px solid var(--color-border-subtle);
    border-radius: 6px;
    overflow: hidden;
    margin-bottom: 8px;
  }

  .progress-bar {
    height: 100%;
    background: linear-gradient(90deg, var(--color-danger), var(--color-danger-bright));
    transition: width 0.1s linear;
  }

  .progress-text {
    font-size: 12px;
    color: var(--color-text-muted);
    text-align: center;
    margin: 0 0 12px;
  }

  .stop-btn {
    width: 100%;
    padding: 10px 16px;
    background-color: var(--color-danger-bright);
    color: white;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-weight: 600;
    font-family: var(--font-body);
    transition: all var(--transition-fast);
  }

  .stop-btn:hover {
    background-color: #ff4444;
    transform: scale(1.02);
  }

  .crafting-popups {
    position: fixed;
    bottom: calc(var(--bottombar-height, 60px) + 30px);
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    flex-direction: column;
    gap: 8px;
    z-index: 300;
    pointer-events: none;
  }

  .crafting-popup {
    padding: 8px 16px;
    background-color: var(--color-gold);
    color: #000;
    border-radius: 6px;
    font-weight: 600;
    font-size: 13px;
    animation: popupSlide 0.3s ease-out;
  }

  @keyframes popupSlide {
    from {
      transform: translateY(10px);
      opacity: 0;
    }
    to {
      transform: translateY(0);
      opacity: 1;
    }
  }

  @media (max-width: 600px) {
    .view-blacksmith {
      padding: 16px 12px;
      gap: 12px;
    }

    .page-title {
      font-size: 18px;
    }

    .card {
      padding: 12px;
    }

    .ingot-summary {
      grid-template-columns: 1fr;
    }

    .recipe-btn {
      padding: 10px;
      gap: 10px;
    }

    .recipe-btn-icon {
      font-size: 24px;
      width: 28px;
    }

    .recipe-btn-info {
      font-size: 12px;
    }
  }
</style>
