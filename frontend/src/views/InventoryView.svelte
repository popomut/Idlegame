<script>
  const equippedSlots = [
    { slot: 'Head',     icon: '&#x1F3A9;', item: null },
    { slot: 'Chest',    icon: '&#x1F9E5;', item: 'Leather Tunic' },
    { slot: 'Legs',     icon: '&#x1FA72;', item: null },
    { slot: 'Weapon',   icon: '&#x1F5E1;&#xFE0F;', item: 'Rusty Sword' },
    { slot: 'Shield',   icon: '&#x1F6E1;&#xFE0F;', item: null },
    { slot: 'Ring',     icon: '&#x1F48D;', item: null },
  ];

  const inventoryItems = [
    { id: 1, name: 'Health Potion',   icon: '&#x1F9EA;', qty: 3, rarity: 'common' },
    { id: 2, name: 'Mana Crystal',    icon: '&#x1F52E;', qty: 2, rarity: 'uncommon' },
    { id: 3, name: 'Iron Ore',        icon: '&#x26AB;',  qty: 10, rarity: 'common' },
    { id: 4, name: 'Dragon Scale',    icon: '&#x1F409;', qty: 1, rarity: 'rare' },
    { id: 5, name: 'Old Map',         icon: '&#x1F5FA;&#xFE0F;', qty: 1, rarity: 'uncommon' },
    { id: 6, name: 'Gold Coin',       icon: '&#x1FA99;', qty: 50, rarity: 'common' },
  ];
</script>

<div class="view-inventory">
  <div class="page-header">
    <h1 class="page-title">&#x1F392; Inventory</h1>
    <p class="page-subtitle">Your items and equipment</p>
  </div>

  <!-- Equipment -->
  <div class="card">
    <div class="card-header">
      <span class="card-icon">&#x1F6E1;&#xFE0F;</span>
      <h2 class="card-title">Equipment</h2>
    </div>
    <div class="equipment-grid">
      {#each equippedSlots as slot}
        <div class="equipment-slot" class:empty={!slot.item} class:filled={!!slot.item}>
          <span class="slot-icon">{@html slot.icon}</span>
          <div class="slot-info">
            <div class="slot-name">{slot.slot}</div>
            <div class="slot-item">{slot.item || 'Empty'}</div>
          </div>
        </div>
      {/each}
    </div>
  </div>

  <!-- Inventory bag -->
  <div class="card">
    <div class="card-header">
      <span class="card-icon">&#x1F392;</span>
      <h2 class="card-title">Bag</h2>
      <span class="bag-count">{inventoryItems.length} / 24</span>
    </div>
    <div class="items-list">
      {#each inventoryItems as item}
        <div class="item-row rarity-{item.rarity}">
          <span class="item-icon">{@html item.icon}</span>
          <span class="item-name">{item.name}</span>
          <span class="item-qty">x{item.qty}</span>
        </div>
      {/each}
    </div>
  </div>
</div>

<style>
  .view-inventory {
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

  .bag-count {
    font-size: 12px;
    color: var(--color-text-muted);
    background-color: var(--color-bg-elevated);
    padding: 2px 8px;
    border-radius: 10px;
    border: 1px solid var(--color-border);
  }

  /* Equipment */
  .equipment-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
  }

  .equipment-slot {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px;
    border-radius: 8px;
    border: 1px solid var(--color-border-subtle);
    background-color: var(--color-bg-elevated);
  }

  .equipment-slot.filled {
    border-color: var(--color-border);
  }

  .slot-icon {
    font-size: 22px;
    width: 30px;
    text-align: center;
    flex-shrink: 0;
  }

  .slot-name {
    font-size: 11px;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .slot-item {
    font-size: 13px;
    color: var(--color-text-heading);
  }

  .empty .slot-item {
    color: var(--color-text-muted);
    font-style: italic;
  }

  /* Items list */
  .items-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .item-row {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 12px;
    border-radius: 6px;
    background-color: var(--color-bg-elevated);
    border-left: 3px solid transparent;
    transition: background-color var(--transition-fast);
  }

  .item-row:hover {
    background-color: var(--color-bg-hover);
  }

  .rarity-common {
    border-left-color: var(--color-text-muted);
  }

  .rarity-uncommon {
    border-left-color: #22c55e;
  }

  .rarity-rare {
    border-left-color: var(--color-magic-bright);
  }

  .rarity-epic {
    border-left-color: var(--color-danger-bright);
  }

  .item-icon {
    font-size: 20px;
    width: 28px;
    text-align: center;
    flex-shrink: 0;
  }

  .item-name {
    flex: 1;
    font-size: 14px;
    color: var(--color-text);
  }

  .item-qty {
    font-size: 13px;
    color: var(--color-gold-dim);
    font-weight: 500;
  }
</style>
