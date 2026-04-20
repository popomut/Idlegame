<script>
  import { player } from '../stores/game.js';
  import { addLogEntry } from '../stores/game.js';

  // Placeholder enemy data
  let enemy = {
    name: 'Goblin Scout',
    level: 1,
    hp: 30,
    maxHp: 30,
    icon: '&#x1F47A;',
  };

  let combatLog = ['You enter the battlefield...', 'A Goblin Scout appears!'];
  let isBattling = false;

  function getHpPercent(hp, maxHp) {
    return Math.round((hp / maxHp) * 100);
  }

  function startBattle() {
    isBattling = true;
    combatLog = ['Battle started!', ...combatLog];
    addLogEntry('You engage in combat with ' + enemy.name + '!');
  }

  function fleeBattle() {
    isBattling = false;
    combatLog = ['You fled from battle.', ...combatLog];
    addLogEntry('You fled from ' + enemy.name + '.');
  }
</script>

<div class="view-combat">
  <div class="page-header">
    <h1 class="page-title">&#x2694;&#xFE0F; Combat</h1>
    <p class="page-subtitle">Fight monsters, earn rewards</p>
  </div>

  <!-- Arena -->
  <div class="card arena-card">
    <div class="combatants">
      <!-- Player side -->
      <div class="combatant player-side">
        <div class="combatant-icon">&#x1F9DD;</div>
        <div class="combatant-name">{$player.name}</div>
        <div class="combatant-level">Lv. {$player.level}</div>
        <div class="hp-bar-track">
          <div
            class="hp-bar-fill player-hp"
            style="width: {getHpPercent($player.hp, $player.maxHp)}%"
          ></div>
        </div>
        <div class="hp-text">{$player.hp} / {$player.maxHp}</div>
      </div>

      <div class="vs-divider">VS</div>

      <!-- Enemy side -->
      <div class="combatant enemy-side">
        <div class="combatant-icon">{@html enemy.icon}</div>
        <div class="combatant-name">{enemy.name}</div>
        <div class="combatant-level">Lv. {enemy.level}</div>
        <div class="hp-bar-track">
          <div
            class="hp-bar-fill enemy-hp"
            style="width: {getHpPercent(enemy.hp, enemy.maxHp)}%"
          ></div>
        </div>
        <div class="hp-text">{enemy.hp} / {enemy.maxHp}</div>
      </div>
    </div>
  </div>

  <!-- Combat actions -->
  <div class="card">
    <div class="card-header">
      <span class="card-icon">&#x26A1;</span>
      <h2 class="card-title">Actions</h2>
    </div>
    <div class="combat-actions">
      {#if !isBattling}
        <button class="action-btn start-btn" on:click={startBattle}>
          <span>&#x2694;&#xFE0F;</span>
          <span>Start Battle</span>
        </button>
      {:else}
        <button class="action-btn attack-btn">
          <span>&#x1F5E1;&#xFE0F;</span>
          <span>Attack</span>
        </button>
        <button class="action-btn magic-btn">
          <span>&#x1F52E;</span>
          <span>Cast Spell</span>
        </button>
        <button class="action-btn flee-btn" on:click={fleeBattle}>
          <span>&#x1F4A8;</span>
          <span>Flee</span>
        </button>
      {/if}
    </div>
  </div>

  <!-- Combat log -->
  <div class="card">
    <div class="card-header">
      <span class="card-icon">&#x1F4DC;</span>
      <h2 class="card-title">Battle Log</h2>
    </div>
    <ul class="combat-log">
      {#each combatLog as entry}
        <li class="log-entry">
          <span class="log-bullet">&#x25B8;</span>
          {entry}
        </li>
      {/each}
    </ul>
  </div>
</div>

<style>
  .view-combat {
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
    color: var(--color-danger-bright);
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
  }

  .arena-card {
    border-color: var(--color-danger);
  }

  .combatants {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }

  .combatant {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
  }

  .combatant-icon {
    font-size: 48px;
    line-height: 1;
    padding: 12px;
    background-color: var(--color-bg-elevated);
    border-radius: 50%;
    border: 2px solid var(--color-border);
  }

  .player-side .combatant-icon {
    border-color: var(--color-magic);
  }

  .enemy-side .combatant-icon {
    border-color: var(--color-danger);
  }

  .combatant-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--color-text-heading);
  }

  .combatant-level {
    font-size: 12px;
    color: var(--color-text-muted);
  }

  .hp-bar-track {
    width: 100%;
    height: 8px;
    background-color: var(--color-bg-elevated);
    border-radius: 4px;
    overflow: hidden;
  }

  .hp-bar-fill {
    height: 100%;
    border-radius: 4px;
    transition: width 0.4s ease;
  }

  .player-hp {
    background: linear-gradient(90deg, #15803d, #22c55e);
  }

  .enemy-hp {
    background: linear-gradient(90deg, #b91c1c, var(--color-danger-bright));
  }

  .hp-text {
    font-size: 12px;
    color: var(--color-text-muted);
  }

  .vs-divider {
    font-family: var(--font-heading);
    font-size: 20px;
    font-weight: 700;
    color: var(--color-gold-dim);
    flex-shrink: 0;
  }

  .combat-actions {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }

  .action-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 12px 20px;
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

  .start-btn {
    width: 100%;
    padding: 14px;
    font-size: 16px;
    border-color: var(--color-danger);
    color: var(--color-danger-bright);
  }

  .start-btn:hover {
    background-color: rgba(220, 38, 38, 0.15);
  }

  .attack-btn:hover {
    border-color: var(--color-gold-dim);
    color: var(--color-gold);
  }

  .magic-btn:hover {
    border-color: var(--color-magic);
    color: var(--color-magic-bright);
  }

  .flee-btn:hover {
    border-color: var(--color-text-muted);
    color: var(--color-text);
  }

  .combat-log {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
    max-height: 150px;
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
    color: var(--color-danger-bright);
    flex-shrink: 0;
  }
</style>
