<script>
  import { player } from '../stores/game.js';

  const shopItems = [
    { id: 1, name: 'Health Potion',    icon: '&#x1F9EA;', cost: 25,  desc: 'Restores 50 HP',         category: 'consumable' },
    { id: 2, name: 'Mana Elixir',      icon: '&#x1F4A7;', cost: 30,  desc: 'Restores 25 Mana',       category: 'consumable' },
    { id: 3, name: 'Iron Sword',       icon: '&#x1F5E1;&#xFE0F;', cost: 120, desc: '+5 Attack',       category: 'weapon' },
    { id: 4, name: 'Leather Armor',    icon: '&#x1F9E5;', cost: 80,  desc: '+8 Defense',             category: 'armor' },
    { id: 5, name: 'Mage Staff',       icon: '&#x1FA84;', cost: 150, desc: '+10 Magic Attack',       category: 'weapon' },
    { id: 6, name: 'Tome of Flames',   icon: '&#x1F4D5;', cost: 200, desc: 'Learn Fireball spell',   category: 'book' },
  ];

  let selectedCategory = 'all';

  const categories = [
    { id: 'all',        label: 'All' },
    { id: 'consumable', label: 'Potions' },
    { id: 'weapon',     label: 'Weapons' },
    { id: 'armor',      label: 'Armor' },
    { id: 'book',       label: 'Books' },
  ];

  $: filteredItems = selectedCategory === 'all'
    ? shopItems
    : shopItems.filter(function (item) { return item.category === selectedCategory; });

  function canAfford(cost) {
    return $player.gold >= cost;
  }
</script>

<div class="view-shop">
  <div class="page-header">
    <h1 class="page-title">&#x1F3EA; Shop</h1>
    <p class="page-subtitle">Spend your hard-earned gold</p>
  </div>

  <div class="gold-display card">
    <span class="gold-icon">&#x1F4B0;</span>
    <span class="gold-label">Your Gold:</span>
    <span class="gold-amount">{$player.gold}</span>
  </div>

  <!-- Category filter -->
  <div class="category-bar">
    {#each categories as cat}
      <button
        class="cat-btn"
        class:active={selectedCategory === cat.id}
        on:click={() => { selectedCategory = cat.id; }}
      >
        {cat.label}
      </button>
    {/each}
  </div>

  <!-- Items -->
  <div class="card">
    <div class="items-list">
      {#each filteredItems as item}
        <div class="shop-item">
          <span class="item-icon">{@html item.icon}</span>
          <div class="item-info">
            <div class="item-name">{item.name}</div>
            <div class="item-desc">{item.desc}</div>
          </div>
          <div class="item-buy">
            <span class="item-cost" class:unaffordable={!canAfford(item.cost)}>
              &#x1F4B0; {item.cost}
            </span>
            <button
              class="buy-btn"
              disabled={!canAfford(item.cost)}
              title={canAfford(item.cost) ? 'Buy' : 'Not enough gold'}
            >
              Buy
            </button>
          </div>
        </div>
      {/each}
    </div>
  </div>
</div>

<style>
  .view-shop {
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

  .gold-display {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 16px;
    border-color: var(--color-gold-dim);
  }

  .gold-icon {
    font-size: 20px;
  }

  .gold-label {
    font-size: 14px;
    color: var(--color-text-muted);
    flex: 1;
  }

  .gold-amount {
    font-size: 18px;
    font-weight: 600;
    color: var(--color-gold);
    font-family: var(--font-heading);
  }

  /* Category tabs */
  .category-bar {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .cat-btn {
    padding: 6px 14px;
    background-color: var(--color-bg-panel);
    border: 1px solid var(--color-border);
    border-radius: 20px;
    color: var(--color-text-muted);
    cursor: pointer;
    font-size: 13px;
    font-family: var(--font-body);
    transition: background-color var(--transition-fast), color var(--transition-fast), border-color var(--transition-fast);
  }

  .cat-btn:hover {
    background-color: var(--color-bg-hover);
    color: var(--color-text);
  }

  .cat-btn.active {
    background-color: var(--color-bg-elevated);
    border-color: var(--color-gold-dim);
    color: var(--color-gold);
  }

  /* Shop items */
  .items-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .shop-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 12px;
    background-color: var(--color-bg-elevated);
    border: 1px solid var(--color-border-subtle);
    border-radius: 8px;
    transition: border-color var(--transition-fast);
  }

  .shop-item:hover {
    border-color: var(--color-border);
  }

  .item-icon {
    font-size: 24px;
    width: 34px;
    text-align: center;
    flex-shrink: 0;
  }

  .item-info {
    flex: 1;
    min-width: 0;
  }

  .item-name {
    font-size: 14px;
    font-weight: 500;
    color: var(--color-text-heading);
    margin-bottom: 2px;
  }

  .item-desc {
    font-size: 12px;
    color: var(--color-text-muted);
  }

  .item-buy {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 6px;
    flex-shrink: 0;
  }

  .item-cost {
    font-size: 13px;
    color: var(--color-gold);
    font-weight: 500;
  }

  .item-cost.unaffordable {
    color: var(--color-danger-bright);
  }

  .buy-btn {
    padding: 5px 14px;
    background-color: var(--color-bg-panel);
    border: 1px solid var(--color-gold-dim);
    border-radius: 6px;
    color: var(--color-gold);
    cursor: pointer;
    font-size: 13px;
    font-family: var(--font-body);
    font-weight: 500;
    transition: background-color var(--transition-fast), color var(--transition-fast);
  }

  .buy-btn:hover:not(:disabled) {
    background-color: rgba(200, 168, 75, 0.15);
  }

  .buy-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
    border-color: var(--color-border);
    color: var(--color-text-muted);
  }
</style>
