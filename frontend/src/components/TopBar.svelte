<script>
  import { toggleSidebar } from '../stores/sidebar.js';
  import { player } from '../stores/game.js';
</script>

<header class="topbar">
  <button class="topbar-menu-btn" on:click={toggleSidebar} aria-label="Toggle menu">
    <span class="hamburger-icon">&#9776;</span>
  </button>

  <div class="topbar-title">
    <span class="title-icon">&#x2623;&#xFE0F;</span>
    <span class="title-text">TOXIC PROTOCOL</span>
  </div>

  <div class="topbar-right">
    <span class="gold-badge">
      <span class="gold-icon">&#x1FA99;</span>
      <span class="gold-amount">{$player.gold}</span>
    </span>
    <span class="level-badge">RANK {$player.level}</span>
  </div>
</header>

<style>
  .topbar {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    height: var(--topbar-height);
    background-color: var(--color-bg-panel);
    border-bottom: 1px solid var(--color-border);
    display: flex;
    align-items: center;
    padding: 0 12px;
    gap: 10px;
    z-index: 100;
  }

  .topbar-menu-btn {
    background: none;
    border: none;
    color: var(--color-text);
    cursor: pointer;
    padding: 8px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color var(--transition-fast);
    flex-shrink: 0;
  }

  .topbar-menu-btn:hover {
    background-color: var(--color-bg-hover);
    color: var(--color-gold);
  }

  .hamburger-icon {
    font-size: 20px;
    line-height: 1;
  }

  .topbar-title {
    display: flex;
    align-items: center;
    gap: 8px;
    flex: 1;
    min-width: 0;
  }

  .title-icon {
    font-size: 20px;
    flex-shrink: 0;
  }

  .title-text {
    font-family: var(--font-heading);
    font-size: 16px;
    color: var(--color-magic-bright);
    font-weight: 700;
    letter-spacing: 2px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    text-transform: uppercase;
  }

  .topbar-right {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  .gold-badge {
    display: flex;
    align-items: center;
    gap: 4px;
    background-color: var(--color-bg-elevated);
    border: 1px solid var(--color-border);
    border-radius: 12px;
    padding: 3px 10px;
  }

  .gold-icon {
    font-size: 13px;
  }

  .gold-amount {
    font-size: 13px;
    color: var(--color-gold);
    font-weight: 500;
  }

  .level-badge {
    font-size: 13px;
    color: var(--color-magic-bright);
    background-color: var(--color-bg-elevated);
    border: 1px solid var(--color-border);
    padding: 3px 10px;
    border-radius: 12px;
    font-weight: 500;
    white-space: nowrap;
  }

  @media (max-width: 400px) {
    .title-text {
      font-size: 14px;
    }
    .gold-badge {
      display: none;
    }
  }
</style>
