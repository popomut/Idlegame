<script>
  import { onMount } from 'svelte';
  import { equipmentAPI } from '../services/api.js';

  // ── State ──────────────────────────────────────────────────────────────
  let slots = [];       // EquippedSlotResponse[]
  let bag = [];         // UserEquipmentResponse[]
  let bagSort = 'latest'; // 'alpha' | 'latest'
  let selectedItem = null; // UserEquipmentResponse — drives popup
  let ringSlotChoice = ''; // 'ring1' | 'ring2' — shown when equipping a ring
  let loading = true;
  let errorMsg = '';

  // ── Layout: body positions for 8 slots ────────────────────────────────
  // Grid: 3 columns × 4 rows. Empty cells use slot id 'empty'.
  const bodyLayout = [
    ['empty',  'head',   'empty' ],
    ['ring1',  'chest',  'amulet'],
    ['weapon', 'empty',  'shield'],
    ['ring2',  'legs',   'empty' ],
  ];

  const slotMeta = {
    head:   { label: 'Head',    icon: '⛑️' },
    chest:  { label: 'Chest',   icon: '🦺' },
    legs:   { label: 'Legs',    icon: '👖' },
    weapon: { label: 'Weapon',  icon: '🗡️' },
    shield: { label: 'Shield',  icon: '🛡️' },
    ring1:  { label: 'Ring L',  icon: '💍' },
    ring2:  { label: 'Ring R',  icon: '💍' },
    amulet: { label: 'Amulet',  icon: '📿' },
  };

  const rarityColor = {
    common:    'var(--color-text-muted)',
    uncommon:  '#3a9a3a',
    rare:      'var(--color-magic-bright)',
    epic:      'var(--color-danger-bright)',
    legendary: 'var(--color-gold)',
  };

  // ── Data loading ───────────────────────────────────────────────────────
  onMount(async function () {
    await loadAll();
  });

  async function loadAll() {
    loading = true;
    errorMsg = '';
    try {
      const [slotsRes, bagRes] = await Promise.all([
        equipmentAPI.getSlots(),
        equipmentAPI.getBag(),
      ]);
      slots = slotsRes.data;
      bag = bagRes.data;
    } catch (err) {
      errorMsg = 'Failed to load equipment data.';
    } finally {
      loading = false;
    }
  }

  // ── Derived helpers ────────────────────────────────────────────────────
  function getSlot(slotName) {
    return slots.find(s => s.slot === slotName) || { slot: slotName, user_equipment_id: 0, equipment: null };
  }

  $: sortedBag = bagSort === 'alpha'
    ? [...bag].sort((a, b) => a.equipment.name.localeCompare(b.equipment.name))
    : [...bag].sort((a, b) => new Date(b.obtained_at) - new Date(a.obtained_at));

  // ── Popup actions ──────────────────────────────────────────────────────
  function openPopup(item) {
    selectedItem = item;
    ringSlotChoice = '';
  }

  function closePopup() {
    selectedItem = null;
    ringSlotChoice = '';
  }

  async function equipSelected() {
    if (!selectedItem) return;
    const equip = selectedItem.equipment;
    let slot = equip.slot;

    if (slot === 'ring') {
      if (!ringSlotChoice) { ringSlotChoice = 'ring1'; return; } // first click: ask which ring
      slot = ringSlotChoice;
    }

    try {
      await equipmentAPI.equip(selectedItem.user_equipment_id, slot);
      await loadAll();
      closePopup();
    } catch (err) {
      errorMsg = err?.response?.data?.error || 'Failed to equip item.';
    }
  }

  async function unequipSlot(slotName, event) {
    event.stopPropagation();
    try {
      await equipmentAPI.unequip(slotName);
      await loadAll();
    } catch (err) {
      errorMsg = 'Failed to unequip.';
    }
  }

  function modLabel(mod) {
    const labels = {
      str: 'STR', int: 'INT', dex: 'DEX',
      resist_fire: 'Fire Res', resist_lightning: 'Lightning Res',
      resist_ice: 'Ice Res', resist_poison: 'Poison Res', resist_chaos: 'Chaos Res',
    };
    const sign = mod.value >= 0 ? '+' : '';
    return `${sign}${mod.value} ${labels[mod.type] || mod.type}`;
  }
</script>

<!-- ── Equipment page ───────────────────────────────────────────────── -->
<div class="view-equipment">
  <div class="page-header">
    <h1 class="page-title">&#x2694;&#xFE0F; Equipment</h1>
    <p class="page-subtitle">Loadout & Armory</p>
  </div>

  {#if errorMsg}
    <div class="error-banner">{errorMsg}</div>
  {/if}

  {#if loading}
    <div class="loading">Loading equipment…</div>
  {:else}

    <!-- ── Section 1: Body layout ──────────────────────────────────── -->
    <div class="card">
      <div class="card-header">
        <span class="card-icon">&#x1FA96;</span>
        <h2 class="card-title">Loadout</h2>
      </div>

      <div class="body-grid">
        {#each bodyLayout as row}
          {#each row as cellSlot}
            {#if cellSlot === 'empty'}
              <div class="body-cell empty-cell"></div>
            {:else}
              {@const slotData = getSlot(cellSlot)}
              {@const meta = slotMeta[cellSlot]}
              <div
                class="body-cell slot-cell"
                class:slot-filled={!!slotData.equipment}
              >
                <span class="slot-label">{meta.label}</span>
                {#if slotData.equipment}
                  <span class="slot-equip-icon">{slotData.equipment.icon}</span>
                  <span class="slot-equip-name" style="color:{rarityColor[slotData.equipment.rarity] || 'inherit'}">{slotData.equipment.name}</span>
                  <button class="unequip-btn" on:click={(e) => unequipSlot(cellSlot, e)} title="Unequip">✕</button>
                {:else}
                  <span class="slot-empty-icon">{meta.icon}</span>
                  <span class="slot-empty-text">—</span>
                {/if}
              </div>
            {/if}
          {/each}
        {/each}
      </div>
    </div>

    <!-- ── Section 2: Equipment bag ──────────────────────────────────── -->
    <div class="card">
      <div class="card-header">
        <span class="card-icon">&#x1F392;</span>
        <h2 class="card-title">Armory Bag</h2>
        <div class="sort-btns">
          <button class="sort-btn" class:sort-active={bagSort === 'latest'} on:click={() => bagSort = 'latest'}>Latest</button>
          <button class="sort-btn" class:sort-active={bagSort === 'alpha'}  on:click={() => bagSort = 'alpha'}>A–Z</button>
        </div>
      </div>

      {#if sortedBag.length === 0}
        <p class="empty-bag">No equipment obtained yet. Earn gear through combat and scavenging.</p>
      {:else}
        <div class="bag-list">
          {#each sortedBag as item}
            <button class="bag-item" on:click={() => openPopup(item)}>
              <span class="bag-icon">{item.equipment.icon}</span>
              <div class="bag-info">
                <span class="bag-name" style="color:{rarityColor[item.equipment.rarity] || 'inherit'}">{item.equipment.name}</span>
                <span class="bag-slot">{item.equipment.slot.toUpperCase()} · {item.equipment.rarity}</span>
              </div>
              {#if item.equipment.base_attack > 0}
                <span class="bag-stat atk">ATK {item.equipment.base_attack}</span>
              {:else if item.equipment.base_defence > 0}
                <span class="bag-stat def">DEF {item.equipment.base_defence}</span>
              {/if}
              <span class="bag-chevron">›</span>
            </button>
          {/each}
        </div>
      {/if}
    </div>

  {/if}
</div>

<!-- ── Equipment detail popup ─────────────────────────────────────────── -->
{#if selectedItem}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="popup-backdrop" on:click={closePopup}>
    <div class="popup" on:click|stopPropagation>
      <div class="popup-header" style="border-color:{rarityColor[selectedItem.equipment.rarity]}">
        <span class="popup-icon">{selectedItem.equipment.icon}</span>
        <div>
          <div class="popup-name" style="color:{rarityColor[selectedItem.equipment.rarity]}">{selectedItem.equipment.name}</div>
          <div class="popup-meta">{selectedItem.equipment.slot.toUpperCase()} · {selectedItem.equipment.rarity.toUpperCase()}</div>
        </div>
      </div>

      <p class="popup-desc">{selectedItem.equipment.description}</p>

      <div class="popup-stats">
        {#if selectedItem.equipment.base_attack > 0}
          <div class="popup-stat">
            <span class="stat-key">⚔️ Attack</span>
            <span class="stat-val">{selectedItem.equipment.base_attack} ({selectedItem.equipment.attack_type})</span>
          </div>
        {/if}
        {#if selectedItem.equipment.base_defence > 0}
          <div class="popup-stat">
            <span class="stat-key">🛡️ Defence</span>
            <span class="stat-val">{selectedItem.equipment.base_defence}</span>
          </div>
        {/if}
        {#each selectedItem.equipment.modifiers as mod}
          <div class="popup-stat">
            <span class="stat-key">✦ Bonus</span>
            <span class="stat-val mod-val">{modLabel(mod)}</span>
          </div>
        {/each}
        {#if selectedItem.equipment.level_required > 1}
          <div class="popup-stat">
            <span class="stat-key">📊 Level req.</span>
            <span class="stat-val">{selectedItem.equipment.level_required}</span>
          </div>
        {/if}
      </div>

      {#if selectedItem.equipment.slot === 'ring' && ringSlotChoice === ''}
        <!-- Ring slot picker shown on first Equip click -->
      {/if}

      {#if selectedItem.equipment.slot === 'ring' && ringSlotChoice !== ''}
        <div class="ring-picker">
          <span class="ring-picker-label">Equip to which ring slot?</span>
          <div class="ring-picker-btns">
            <button class="ring-btn" class:ring-selected={ringSlotChoice === 'ring1'} on:click={() => ringSlotChoice = 'ring1'}>Ring L</button>
            <button class="ring-btn" class:ring-selected={ringSlotChoice === 'ring2'} on:click={() => ringSlotChoice = 'ring2'}>Ring R</button>
          </div>
        </div>
      {/if}

      <div class="popup-actions">
        <button class="popup-btn equip-btn" on:click={equipSelected}>
          {selectedItem.equipment.slot === 'ring' && ringSlotChoice === '' ? 'Equip…' : 'Equip'}
        </button>
        <button class="popup-btn cancel-btn" on:click={closePopup}>Cancel</button>
      </div>
    </div>
  </div>
{/if}

<style>
  /* ── Page shell ─────────────────────────────────────────────────────── */
  .view-equipment {
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

  .error-banner {
    background: rgba(200,50,50,0.15);
    border: 1px solid var(--color-danger);
    color: var(--color-danger-bright);
    padding: 10px 14px;
    border-radius: 8px;
    font-size: 13px;
  }

  .loading {
    color: var(--color-text-muted);
    font-size: 14px;
    padding: 20px 0;
    text-align: center;
  }

  /* ── Card shell ─────────────────────────────────────────────────────── */
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
    margin-bottom: 16px;
  }

  .card-icon { font-size: 18px; }

  .card-title {
    font-family: var(--font-heading);
    font-size: 15px;
    color: var(--color-text-heading);
    margin: 0;
    font-weight: 600;
    flex: 1;
  }

  /* ── Body layout grid ───────────────────────────────────────────────── */
  .body-grid {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    gap: 8px;
  }

  .body-cell {
    min-height: 80px;
    border-radius: 8px;
  }

  .empty-cell {
    background: transparent;
  }

  .slot-cell {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    padding: 10px 6px;
    background-color: var(--color-bg-elevated);
    border: 1px dashed var(--color-border-subtle);
    position: relative;
    transition: border-color var(--transition-fast);
  }

  .slot-cell.slot-filled {
    border-style: solid;
    border-color: var(--color-border);
  }

  .slot-label {
    font-size: 10px;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  .slot-equip-icon, .slot-empty-icon { font-size: 22px; }

  .slot-equip-name {
    font-size: 11px;
    font-weight: 600;
    text-align: center;
    line-height: 1.2;
  }

  .slot-empty-text {
    font-size: 12px;
    color: var(--color-text-muted);
  }

  .unequip-btn {
    position: absolute;
    top: 4px;
    right: 4px;
    background: none;
    border: none;
    color: var(--color-text-muted);
    cursor: pointer;
    font-size: 11px;
    line-height: 1;
    padding: 2px;
    border-radius: 3px;
    transition: color var(--transition-fast);
  }

  .unequip-btn:hover { color: var(--color-danger-bright); }

  /* ── Bag ────────────────────────────────────────────────────────────── */
  .sort-btns {
    display: flex;
    gap: 6px;
  }

  .sort-btn {
    font-size: 12px;
    padding: 3px 10px;
    border-radius: 10px;
    border: 1px solid var(--color-border);
    background: var(--color-bg-elevated);
    color: var(--color-text-muted);
    cursor: pointer;
    font-family: var(--font-body);
    transition: border-color var(--transition-fast), color var(--transition-fast);
  }

  .sort-btn.sort-active {
    border-color: var(--color-gold-dim);
    color: var(--color-gold);
  }

  .empty-bag {
    color: var(--color-text-muted);
    font-size: 14px;
    font-style: italic;
    text-align: center;
    padding: 20px 0;
    margin: 0;
  }

  .bag-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .bag-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 12px;
    background-color: var(--color-bg-elevated);
    border: 1px solid var(--color-border-subtle);
    border-radius: 8px;
    cursor: pointer;
    text-align: left;
    width: 100%;
    font-family: var(--font-body);
    transition: background-color var(--transition-fast), border-color var(--transition-fast);
  }

  .bag-item:hover {
    background-color: var(--color-bg-hover);
    border-color: var(--color-border);
  }

  .bag-icon { font-size: 22px; width: 28px; text-align: center; flex-shrink: 0; }

  .bag-info { flex: 1; display: flex; flex-direction: column; gap: 2px; }

  .bag-name { font-size: 14px; font-weight: 600; }

  .bag-slot {
    font-size: 11px;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .bag-stat {
    font-size: 12px;
    font-weight: 700;
    padding: 2px 8px;
    border-radius: 6px;
    flex-shrink: 0;
  }

  .bag-stat.atk { color: var(--color-danger-bright); background: rgba(200,60,60,0.12); }
  .bag-stat.def { color: var(--color-magic-bright);  background: rgba(60,120,200,0.12); }

  .bag-chevron { color: var(--color-text-muted); font-size: 18px; }

  /* ── Popup ──────────────────────────────────────────────────────────── */
  .popup-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0,0,0,0.7);
    z-index: 500;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
  }

  .popup {
    background: var(--color-bg-panel);
    border: 1px solid var(--color-border);
    border-radius: 14px;
    padding: 20px;
    width: 100%;
    max-width: 380px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .popup-header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding-bottom: 12px;
    border-bottom: 2px solid var(--color-border);
  }

  .popup-icon { font-size: 32px; }

  .popup-name {
    font-family: var(--font-heading);
    font-size: 18px;
    font-weight: 700;
  }

  .popup-meta {
    font-size: 11px;
    color: var(--color-text-muted);
    letter-spacing: 1px;
    text-transform: uppercase;
    margin-top: 2px;
  }

  .popup-desc {
    font-size: 13px;
    color: var(--color-text-muted);
    line-height: 1.5;
    margin: 0;
  }

  .popup-stats {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .popup-stat {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 13px;
    padding: 5px 10px;
    background: var(--color-bg-elevated);
    border-radius: 6px;
  }

  .stat-key { color: var(--color-text-muted); }
  .stat-val { color: var(--color-text-heading); font-weight: 600; }
  .mod-val  { color: var(--color-gold); }

  /* Ring slot picker */
  .ring-picker {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .ring-picker-label {
    font-size: 12px;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .ring-picker-btns {
    display: flex;
    gap: 8px;
  }

  .ring-btn {
    flex: 1;
    padding: 8px;
    border-radius: 8px;
    border: 1px solid var(--color-border);
    background: var(--color-bg-elevated);
    color: var(--color-text);
    cursor: pointer;
    font-family: var(--font-body);
    font-size: 13px;
    transition: border-color var(--transition-fast), color var(--transition-fast);
  }

  .ring-btn.ring-selected {
    border-color: var(--color-gold-dim);
    color: var(--color-gold);
  }

  /* Popup action buttons */
  .popup-actions {
    display: flex;
    gap: 10px;
  }

  .popup-btn {
    flex: 1;
    padding: 11px;
    border-radius: 8px;
    border: 1px solid var(--color-border);
    cursor: pointer;
    font-family: var(--font-body);
    font-size: 14px;
    font-weight: 600;
    transition: background-color var(--transition-fast), border-color var(--transition-fast), color var(--transition-fast);
  }

  .equip-btn {
    background: var(--color-gold-dim);
    border-color: var(--color-gold);
    color: var(--color-bg);
  }

  .equip-btn:hover { background: var(--color-gold); }

  .cancel-btn {
    background: var(--color-bg-elevated);
    color: var(--color-text-muted);
  }

  .cancel-btn:hover { color: var(--color-text); border-color: var(--color-text-muted); }
</style>
