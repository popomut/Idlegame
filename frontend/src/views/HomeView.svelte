<script>
  import { player, activityLog } from '../stores/game.js';
  import { navigateTo } from '../stores/navigation.js';

  function getHpPercent(hp, maxHp) {
    return Math.round((hp / maxHp) * 100);
  }

  function getManaPercent(mana, maxMana) {
    return Math.round((mana / maxMana) * 100);
  }

  function getXpPercent(xp, xpToNextLevel) {
    return Math.round((xp / xpToNextLevel) * 100);
  }
</script>

<div class="view-home">
  <!-- Page header -->
  <div class="page-header">
    <h1 class="page-title">&#x2694;&#xFE0F; Welcome, {$player.name}</h1>
    <p class="page-subtitle">{$player.class}</p>
  </div>

  <!-- Character card -->
  <div class="card character-card">
    <div class="card-header">
      <span class="card-icon">&#x1F9D9;</span>
      <h2 class="card-title">Character</h2>
      <span class="level-tag">Level {$player.level}</span>
    </div>

    <div class="stat-bars">
      <!-- XP bar -->
      <div class="stat-row">
        <span class="stat-label">&#x2B50; XP</span>
        <div class="stat-bar-track">
          <div
            class="stat-bar-fill xp-fill"
            style="width: {getXpPercent($player.xp, $player.xpToNextLevel)}%"
          ></div>
        </div>
        <span class="stat-value">{$player.xp} / {$player.xpToNextLevel}</span>
      </div>

      <!-- HP bar -->
      <div class="stat-row">
        <span class="stat-label">&#x2764;&#xFE0F; HP</span>
        <div class="stat-bar-track">
          <div
            class="stat-bar-fill hp-fill"
            style="width: {getHpPercent($player.hp, $player.maxHp)}%"
          ></div>
        </div>
        <span class="stat-value">{$player.hp} / {$player.maxHp}</span>
      </div>

      <!-- Mana bar -->
      <div class="stat-row">
        <span class="stat-label">&#x1F52E; Mana</span>
        <div class="stat-bar-track">
          <div
            class="stat-bar-fill mana-fill"
            style="width: {getManaPercent($player.mana, $player.maxMana)}%"
          ></div>
        </div>
        <span class="stat-value">{$player.mana} / {$player.maxMana}</span>
      </div>
    </div>
  </div>

  <!-- Quick actions -->
  <div class="card">
    <div class="card-header">
      <span class="card-icon">&#x26A1;</span>
      <h2 class="card-title">Quick Actions</h2>
    </div>
    <div class="quick-actions">
      <button class="action-btn danger-btn" on:click={() => navigateTo('combat')}>
        <span class="action-icon">&#x2694;&#xFE0F;</span>
        <span>Combat</span>
      </button>
      <button class="action-btn magic-btn" on:click={() => navigateTo('skills')}>
        <span class="action-icon">&#x1F4D6;</span>
        <span>Skills</span>
      </button>
      <button class="action-btn gold-btn" on:click={() => navigateTo('shop')}>
        <span class="action-icon">&#x1F3EA;</span>
        <span>Shop</span>
      </button>
      <button class="action-btn" on:click={() => navigateTo('inventory')}>
        <span class="action-icon">&#x1F392;</span>
        <span>Inventory</span>
      </button>
    </div>
  </div>

  <!-- Activity log -->
  <div class="card activity-card">
    <div class="card-header">
      <span class="card-icon">&#x1F4DC;</span>
      <h2 class="card-title">Activity Log</h2>
    </div>
    <ul class="activity-log">
      {#each $activityLog as entry}
        <li class="log-entry">
          <span class="log-bullet">&#x25B8;</span>
          {entry}
        </li>
      {/each}
    </ul>
  </div>
</div>

<style>
  .view-home {
    padding: 20px 16px 20px;
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
    color: var(--color-magic-bright);
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
    letter-spacing: 0.3px;
    flex: 1;
  }

  .level-tag {
    font-size: 12px;
    color: var(--color-gold);
    background-color: var(--color-bg-elevated);
    border: 1px solid var(--color-gold-dim);
    padding: 2px 8px;
    border-radius: 10px;
  }

  /* Stat bars */
  .stat-bars {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .stat-row {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .stat-label {
    font-size: 13px;
    color: var(--color-text-muted);
    width: 64px;
    flex-shrink: 0;
  }

  .stat-bar-track {
    flex: 1;
    height: 10px;
    background-color: var(--color-bg-elevated);
    border-radius: 5px;
    overflow: hidden;
    border: 1px solid var(--color-border-subtle);
  }

  .stat-bar-fill {
    height: 100%;
    border-radius: 5px;
    transition: width 0.4s ease;
  }

  .xp-fill {
    background: linear-gradient(90deg, var(--color-gold-dim), var(--color-gold));
  }

  .hp-fill {
    background: linear-gradient(90deg, #15803d, #22c55e);
  }

  .mana-fill {
    background: linear-gradient(90deg, var(--color-magic), var(--color-magic-bright));
  }

  .stat-value {
    font-size: 12px;
    color: var(--color-text-muted);
    width: 70px;
    text-align: right;
    flex-shrink: 0;
  }

  /* Quick actions */
  .quick-actions {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }

  .action-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 12px 10px;
    background-color: var(--color-bg-elevated);
    border: 1px solid var(--color-border);
    border-radius: 8px;
    color: var(--color-text);
    cursor: pointer;
    font-size: 14px;
    font-family: var(--font-body);
    font-weight: 500;
    transition: background-color var(--transition-fast), border-color var(--transition-fast), color var(--transition-fast);
  }

  .action-btn:hover {
    background-color: var(--color-bg-hover);
    border-color: var(--color-text-muted);
  }

  .danger-btn:hover {
    border-color: var(--color-danger);
    color: var(--color-danger-bright);
  }

  .magic-btn:hover {
    border-color: var(--color-magic);
    color: var(--color-magic-bright);
  }

  .gold-btn:hover {
    border-color: var(--color-gold-dim);
    color: var(--color-gold);
  }

  .action-icon {
    font-size: 18px;
  }

  /* Activity log */
  .activity-log {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
    max-height: 180px;
    overflow-y: auto;
  }

  .log-entry {
    display: flex;
    gap: 8px;
    font-size: 13px;
    color: var(--color-text-muted);
    line-height: 1.5;
  }

  .log-bullet {
    color: var(--color-gold-dim);
    flex-shrink: 0;
    margin-top: 1px;
  }
</style>
