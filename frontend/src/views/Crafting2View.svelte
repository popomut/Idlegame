<script>
  import { onMount, onDestroy } from 'svelte';
  import { blacksmithSkill } from '../stores/blacksmith_skill.js';
  import { 
    activeCrafting, craftingPopups, craftingProgress,
    startCrafting, stopCrafting,
    syncIngotInventory, syncPotionInventory, syncIngotInventoryDuringCrafting, syncPotionInventoryDuringCrafting,
    ingotInventory, potionInventory, checkCraftingStatus, showCraftingPopup
  } from '../stores/blacksmith.js';
  import { addLogEntry } from '../stores/game.js';
  import axios from 'axios';
  import { API_BASE_URL } from '../services/api.js';

  let recipes = [];
  let selectedOutputType = 'ingot'; // Default to ingot recipes
  let filteredRecipes = [];
  let inventoryOpen = true;
  let craftingInterval = null;
  let syncInterval = null;
  let pendingItems = {}; // Tracks items produced in current cycle before sync
  let ingredientInventory = {}; // ore_key/herb_key → quantity for ingredient availability display
  let allOreTypes = [];      // all ores — for material cache with icons
  let allHerbTypes = [];     // all herbs — for material cache with icons
  let pendingIngredientConsumption = {}; // ingredient_key → quantity being consumed across all crafting cycles

  async function fetchAllIngredients() {
    try {
      const [oreResp, herbResp] = await Promise.all([
        axios.get(`${API_BASE_URL}/api/inventory/ores`, { withCredentials: true }),
        axios.get(`${API_BASE_URL}/api/inventory/herbs`, { withCredentials: true }),
      ]);
      
      const combined = {};
      if (Array.isArray(oreResp.data)) {
        oreResp.data.forEach(item => { combined[item.ore_key] = item.quantity; });
      }
      if (Array.isArray(herbResp.data)) {
        herbResp.data.forEach(item => { combined[item.herb_key] = item.quantity; });
      }
      ingredientInventory = combined;
    } catch (e) {
      console.error('Failed to load ingredient inventory:', e);
    }
  }

  function filterRecipesByOutputType(outputType) {
    filteredRecipes = recipes.filter(r => r.output_type === outputType);
  }

  function selectOutputType(outputType) {
    selectedOutputType = outputType;
    filterRecipesByOutputType(outputType);
  }

  onMount(async function () {
    try {
      const resp = await axios.get(`${API_BASE_URL}/api/blacksmith/recipes`, {
        withCredentials: true,
      });
      recipes = resp.data;
      filterRecipesByOutputType(selectedOutputType);
    } catch (e) {
      console.error('Failed to load recipes:', e);
    }

    // Load ALL ore and herb types for material cache icons
    try {
      const [oreResp, herbResp] = await Promise.all([
        axios.get(`${API_BASE_URL}/api/ore-types`, { withCredentials: true }),
        axios.get(`${API_BASE_URL}/api/herb-types`, { withCredentials: true }),
      ]);
      allOreTypes = oreResp.data || [];
      allHerbTypes = herbResp.data || [];
    } catch (e) {
      console.error('Failed to load ore/herb types for cache:', e);
    }

    // Always sync inventory first to populate Material Cache
    await Promise.all([syncIngotInventory(), syncPotionInventory(), fetchAllIngredients()]);

    if ($activeCrafting) {
      // If resuming an active session, sync with server first
      const recipe = recipes.find(r => r.id === $activeCrafting.craftableItemId);
      if (recipe) {
        selectOutputType(recipe.output_type);
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
    pendingItems = {};
    pendingIngredientConsumption = {};
  });

  // Returns how many times recipe can be crafted based on current inventory
  async function computeMaxCrafts(recipe) {
    if (!recipe.ingredients || recipe.ingredients.length === 0) return Infinity;

    let maxCrafts = Infinity;
    for (const ing of recipe.ingredients) {
      let available = 0;
      if (ing.ingredient_type === 'ore') {
        available = ingredientInventory[ing.ingredient_key] || 0;
      } else if (ing.ingredient_type === 'herb') {
        available = ingredientInventory[ing.ingredient_key] || 0;
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
        pendingItems = {};
        pendingIngredientConsumption = {};
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
      pendingItems = {};
      pendingIngredientConsumption = {};
      await stopCrafting();
      addLogEntry(`Stopped crafting ${currentName}.`);
    } else {
      const maxCrafts = await computeMaxCrafts(recipe);

      if (maxCrafts <= 0) {
        // Find the first missing ingredient for the error message
        let missingIngredient = 'unknown ingredient';
        for (const ing of recipe.ingredients) {
          let available = 0;
          if (ing.ingredient_type === 'ore' || ing.ingredient_type === 'herb') {
            available = ingredientInventory[ing.ingredient_key] || 0;
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

      pendingItems = {};
      await startCrafting(recipe.id, recipe.name, recipe.crafting_time_ms);
      startCraftingUI(recipe, maxCrafts);
    }
  }

  function startCraftingUI(recipe, maxCrafts = Infinity) {
    if (craftingInterval) clearInterval(craftingInterval);
    if (syncInterval) clearInterval(syncInterval);

    const itemKey = recipe.item_key || recipe.ItemKey;
    const interval = recipe.crafting_time_ms || 5000;

    // Reset pending on new session so display starts clean
    pendingItems = {};

    // Start with empty consumption tracking
    // We'll INCREMENT this as each cycle completes (tracking what's been consumed so far)
    pendingIngredientConsumption = {};

    let remainingCrafts = maxCrafts;

    // Fast loop: Update pendingItems and show popup every crafting cycle
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
        pendingItems = {};
        pendingIngredientConsumption = {};
        await stopCrafting();
        addLogEntry(`⚠️ Stopped crafting - not enough ingredients.`);
        return;
      }

      remainingCrafts--;

      // Check server state FIRST — if session stopped (no ingredients), don't show popup
      const stillActive = await syncCraftingProgressDuringCrafting(recipe.name);
      if (!stillActive) return;

      // Only show popup and increment pending if server confirms still active
      const currentPending = (pendingItems[itemKey] || 0);
      pendingItems = { ...pendingItems, [itemKey]: currentPending + 1 };
      showCraftingPopup(1, itemKey);

      // Increment pending ingredient consumption (track what's been consumed so far in this session)
      recipe.ingredients.forEach(ing => {
        const current = pendingIngredientConsumption[ing.ingredient_key] || 0;
        pendingIngredientConsumption[ing.ingredient_key] = current + ing.quantity_required;
      });
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

      // Refresh ingredient inventory (to reflect consumed ingredients)
      // but do NOT overwrite $ingotInventory/$potionInventory here —
      // we let pendingItems accumulate locally (mirrors MiningView pattern).
      // The confirmed inventory is only synced on stopCrafting().
      fetchAllIngredients();

      // If status shows inactive, server auto-stopped us (no ingredients)
      if (!status.is_active) {
        clearInterval(craftingInterval);
        clearInterval(syncInterval);
        craftingInterval = null;
        syncInterval = null;
        pendingItems = {};
        pendingIngredientConsumption = {};
        $activeCrafting = null;
        addLogEntry(`⚠️ Crafting stopped - not enough ingredients.`);
        return false; // signal to caller that crafting stopped
      }

      return true; // still active
    } catch (error) {
      console.error('Failed to sync crafting progress:', error);
      return true; // assume still active on error
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

  // Combine all inventories for display
  $: allInventories = {
    ...$ingotInventory,
    ...$potionInventory,
  };

  // Map from item_key → max_quantity for Material Cache display
  $: maxQtyMap = Object.fromEntries(
    recipes.filter(r => r.max_quantity > 0).map(r => [r.item_key, r.max_quantity])
  );

  // Build map from ingredient key to metadata (name, icon, type)
  function buildIngredientMetadataMap() {
    const map = {};
    
    // Add ores
    allOreTypes.forEach(ore => {
      map[ore.ore_key] = {
        name: ore.ore_name,
        icon: ore.icon,
        type: 'ore'
      };
    });
    
    // Add herbs
    allHerbTypes.forEach(herb => {
      map[herb.herb_key] = {
        name: herb.herb_name,
        icon: herb.icon,
        type: 'herb'
      };
    });
    
    // Add ingots (from recipes with ingot output type)
    recipes.forEach(recipe => {
      if (recipe.output_type === 'ingot') {
        map[recipe.item_key] = {
          name: recipe.name,
          icon: recipe.icon,
          type: 'ingot'
        };
      }
    });
    
    return map;
  }

  function formatInterval(ms) {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(0)}s`;
  }
</script>

<div class="view-crafting2">
  <div class="page-header">
    <h1 class="page-title">🔮 Crafting Menu</h1>
    <p class="page-subtitle">Craft potions and enhance items</p>
  </div>

  <!-- Crafting type selector - buttons -->
  <div class="crafting-type-buttons">
    <button
      class="type-btn"
      class:active={selectedOutputType === 'ingot'}
      on:click={() => selectOutputType('ingot')}
    >
      ⚒️ Ingot
    </button>
    <button
      class="type-btn"
      class:active={selectedOutputType === 'potion'}
      on:click={() => selectOutputType('potion')}
    >
      🧪 Potion
    </button>
  </div>

  <!-- Material cache inventory - collapsible -->
  <div class="card inventory-card">
    <button class="card-header collapse-toggle" on:click={() => inventoryOpen = !inventoryOpen} aria-expanded={inventoryOpen}>
      <span class="card-icon">📦</span>
      <h2 class="card-title">Material Cache</h2>
      <span class="collapse-arrow" class:open={inventoryOpen}>▸</span>
    </button>

    {#if inventoryOpen}
      {#if recipes.filter(r => r.output_type === selectedOutputType).length === 0}
        <div class="inventory-summary">
          <p class="loading-text">No items crafted yet</p>
        </div>
      {:else}
        {@const activeInventory = selectedOutputType === 'potion' ? $potionInventory : $ingotInventory}
        {@const craftedItems = recipes
          .filter(r => r.output_type === selectedOutputType)
          .map(r => ({
            key: r.item_key,
            name: r.name,
            icon: r.icon,
            max_quantity: r.max_quantity,
            quantity: activeInventory[r.item_key] ?? 0,
            pending: pendingItems[r.item_key] ?? 0,
          }))
          .filter(item => item.quantity + item.pending > 0)
        }
        <div class="inventory-summary">
          {#if craftedItems.length === 0}
            <p class="loading-text">No items crafted yet</p>
          {:else}
            {#each craftedItems as item}
              <div class="item-count">
                <span class="item-icon">{item.icon}</span>
                <span class="item-label">{item.name}</span>
                <span class="item-qty">{item.quantity + item.pending}</span>
                {#if item.max_quantity > 0}
                  <span class="item-max">/ {item.max_quantity}</span>
                {/if}
              </div>
            {/each}
          {/if}
        </div>
      {/if}
    {/if}
  </div>

  <!-- Crafting area -->
  <div class="card crafting-card">
    <div class="card-header">
      <span class="card-icon">{selectedOutputType === 'potion' ? '🧪' : '⚒️'}</span>
      <h2 class="card-title">
        {#if $activeCrafting}
          Crafting: {$activeCrafting.recipeName}
        {:else}
          Select Recipe
        {/if}
      </h2>
    </div>

    <div class="recipe-selection">
      {#if filteredRecipes.length === 0}
        <p class="loading-text">
          {recipes.length === 0 ? 'Loading recipes...' : 'No recipes available'}
        </p>
      {:else}
        {#each filteredRecipes as recipe}
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
                  {@const confirmed = ing.ingredient_type === 'ore' || ing.ingredient_type === 'herb'
                    ? (ingredientInventory[ing.ingredient_key] || 0)
                    : ($ingotInventory[ing.ingredient_key] || 0)}
                  {@const pending = pendingIngredientConsumption[ing.ingredient_key] || 0}
                  {@const have = Math.max(0, confirmed - pending)}
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
  .view-crafting2 {
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

  .crafting-type-buttons {
    display: flex;
    gap: 8px;
    margin-bottom: 8px;
  }

  .type-btn {
    flex: 1;
    padding: 12px 16px;
    background-color: var(--color-bg-panel);
    border: 1px solid var(--color-border);
    border-radius: 8px;
    color: var(--color-text);
    font-family: var(--font-body);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .type-btn:hover {
    background-color: var(--color-bg-elevated);
    border-color: var(--color-gold-dim);
  }

  .type-btn.active {
    background-color: var(--color-gold-dim);
    border-color: var(--color-gold);
    color: var(--color-bg);
    font-weight: 600;
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

  .inventory-summary {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }

  .item-count {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 10px;
    background-color: var(--color-bg);
    border-radius: 6px;
    font-size: 13px;
  }

  .item-icon {
    font-size: 16px;
  }

  .item-label {
    flex: 1;
    color: var(--color-text);
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .item-qty {
    font-weight: 600;
    color: var(--color-gold);
  }

  .item-max {
    color: var(--color-text-muted);
    font-size: 12px;
  }

  .loading-text {
    text-align: center;
    color: var(--color-text-muted);
    padding: 20px 0;
    font-size: 13px;
    margin: 0;
  }

  .crafting-card {
    display: flex;
    flex-direction: column;
  }

  .recipe-selection {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .recipe-btn {
    display: grid;
    grid-template-columns: 40px 1fr auto;
    gap: 12px;
    align-items: start;
    padding: 12px;
    background-color: var(--color-bg);
    border: 1px solid var(--color-border);
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;
    font-family: var(--font-body);
    color: var(--color-text);
  }

  .recipe-btn:hover:not(.locked) {
    background-color: var(--color-bg-elevated);
    border-color: var(--color-gold-dim);
  }

  .recipe-btn.active {
    background-color: var(--color-gold-dim);
    border-color: var(--color-gold);
    color: var(--color-bg);
  }

  .recipe-btn.locked {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .recipe-btn-icon {
    font-size: 24px;
    text-align: center;
  }

  .recipe-btn-info {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .recipe-btn-name {
    font-weight: 600;
    font-size: 14px;
  }

  .recipe-btn-ingredients {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }

  .ingredient-badge {
    display: inline-block;
    padding: 2px 6px;
    background-color: var(--color-gold-dim);
    color: var(--color-bg);
    border-radius: 3px;
    font-size: 11px;
    font-weight: 500;
  }

  .recipe-btn.active .ingredient-badge {
    background-color: var(--color-bg);
    color: var(--color-gold-dim);
  }

  .recipe-btn-meta {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }

  .badge {
    display: inline-block;
    padding: 3px 8px;
    background-color: var(--color-bg-elevated);
    border-radius: 3px;
    font-size: 11px;
    font-weight: 500;
  }

  .locked-badge {
    background-color: #4a4a4a;
    color: #ccc;
  }

  .time, .xp {
    color: var(--color-text-muted);
  }

  .recipe-btn-status {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .crafting-indicator {
    font-size: 18px;
    animation: pulse 1s infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }

  .recipe-ing-status {
    grid-column: 1 / -1;
    display: flex;
    gap: 8px;
    padding-top: 8px;
    border-top: 1px solid var(--color-border);
    flex-wrap: wrap;
  }

  .ing-avail {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 3px;
    padding: 4px 8px;
    background-color: var(--color-bg-elevated);
    border-radius: 4px;
    font-size: 11px;
  }

  .ing-avail-low {
    background-color: #3d3d1f;
    border: 1px solid #665522;
  }

  .ing-avail-count {
    font-weight: 600;
    color: var(--color-gold);
  }

  .ing-avail-low .ing-avail-count {
    color: #ff6b6b;
  }

  .ing-avail-name {
    color: var(--color-text-muted);
  }

  .crafting-info {
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid var(--color-border);
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .crafting-text {
    margin: 0;
    text-align: center;
    color: var(--color-gold);
    font-weight: 600;
    font-size: 13px;
  }

  .progress-bar-container {
    width: 100%;
    height: 24px;
    background-color: var(--color-bg);
    border: 1px solid var(--color-border);
    border-radius: 4px;
    overflow: hidden;
  }

  .progress-bar {
    height: 100%;
    background: linear-gradient(90deg, var(--color-gold-dim), var(--color-gold));
    transition: width 0.1s linear;
  }

  .progress-text {
    margin: 0;
    text-align: center;
    font-size: 12px;
    color: var(--color-text-muted);
  }

  .stop-btn {
    padding: 10px 16px;
    background-color: #4a3333;
    border: 1px solid #8b4444;
    border-radius: 6px;
    color: #ff6b6b;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    font-family: var(--font-body);
  }

  .stop-btn:hover {
    background-color: #5a3333;
    border-color: #ff6b6b;
  }

  .crafting-popups {
    position: fixed;
    top: 80px;
    right: 20px;
    display: flex;
    flex-direction: column;
    gap: 10px;
    z-index: 1000;
    pointer-events: none;
  }

  .crafting-popup {
    padding: 8px 16px;
    background-color: var(--color-gold-dim);
    color: var(--color-bg);
    border-radius: 6px;
    font-weight: 600;
    font-size: 14px;
    animation: popupFloat 1.5s ease-out forwards;
  }

  @keyframes popupFloat {
    0% {
      opacity: 1;
      transform: translateY(0);
    }
    100% {
      opacity: 0;
      transform: translateY(-40px);
    }
  }

  @media (max-width: 640px) {
    .inventory-summary {
      grid-template-columns: 1fr;
    }

    .recipe-selection {
      gap: 8px;
    }
  }
</style>
