<script>
  const inventoryItems = [
    { id: 1, name: 'Medkit',          icon: '&#x1F9EA;', qty: 3,  rarity: 'common' },
    { id: 2, name: 'Stim Pack',       icon: '&#x1F489;', qty: 2,  rarity: 'uncommon' },
    { id: 3, name: 'Iron Ore',        icon: '&#x26AB;',  qty: 10, rarity: 'common' },
    { id: 4, name: 'Mutant Sample',   icon: '&#x1F9EC;', qty: 1,  rarity: 'rare' },
    { id: 5, name: 'Tactical Map',    icon: '&#x1F5FA;&#xFE0F;', qty: 1, rarity: 'uncommon' },
    { id: 6, name: 'Dog Tag',         icon: '&#x1FA96;', qty: 50, rarity: 'common' },
  ];
</script>

<div class="view-inventory">
  <div class="page-header">
    <h1 class="page-title">&#x1F392; Field Kit</h1>
    <p class="page-subtitle">Supplies and consumables</p>
  </div>

  <!-- Inventory bag (consumables/general items only — equipment is in the Equipment screen) -->
  <div class="card">
    <div class="card-header">
      <span class="card-icon">&#x1F392;</span>
      <h2 class="card-title">Field Pack</h2>
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

  .card-icon { font-size: 18px; }

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

  .item-row:hover { background-color: var(--color-bg-hover); }

  .rarity-common    { border-left-color: var(--color-text-muted); }
  .rarity-uncommon  { border-left-color: #3a8a3a; }
  .rarity-rare      { border-left-color: var(--color-magic-bright); }
  .rarity-epic      { border-left-color: var(--color-danger-bright); }

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
