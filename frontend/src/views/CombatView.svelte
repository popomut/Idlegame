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
    attack: 8,
  };

  let combatLog = ['You enter the battlefield...', 'A Goblin Scout appears!'];
  let isBattling = false;
  let isDead = false;

  function getHpPercent(hp, maxHp) {
    return Math.round((hp / maxHp) * 100);
  }

  function startBattle() {
    isBattling = true;
    isDead = false;
    enemy.hp = enemy.maxHp;
    combatLog = ['Battle started!', 'A ' + enemy.name + ' appears!'];
    addLogEntry('You engage in combat with ' + enemy.name + '!');
  }

  function strike() {
    if (!isBattling || isDead) return;

    // Player attacks enemy
    const playerStr = $player.str ?? 5;
    const playerDmg = Math.max(1, playerStr + Math.floor(Math.random() * 6));
    enemy.hp = Math.max(0, enemy.hp - playerDmg);
    combatLog = [`You strike ${enemy.name} for ${playerDmg} damage. (${enemy.hp}/${enemy.maxHp} HP)`, ...combatLog];

    if (enemy.hp <= 0) {
      combatLog = [`${enemy.name} is defeated!`, ...combatLog];
      addLogEntry(`You defeated ${enemy.name}!`);
      isBattling = false;
      return;
    }

    // Enemy counter-attacks
    const enemyDmg = Math.max(1, enemy.attack + Math.floor(Math.random() * 4));
    player.update(p => {
      const newHp = Math.max(0, p.hp - enemyDmg);
      combatLog = [`${enemy.name} hits you for ${enemyDmg} damage. (${newHp}/${p.maxHp} HP)`, ...combatLog];

      if (newHp <= 0) {
        isDead = true;
        isBattling = false;
        combatLog = ['⚠️ You have been defeated!', ...combatLog];
        addLogEntry('You were defeated by ' + enemy.name + '!');
      }

      return { ...p, hp: newHp };
    });
  }

  function fleeBattle() {
    isBattling = false;
    combatLog = ['You fell back from battle.', ...combatLog];
    addLogEntry('You fled from ' + enemy.name + '.');
  }

  function returnToBase() {
    isDead = false;
    isBattling = false;
    enemy.hp = enemy.maxHp;
    combatLog = ['You return to Base Camp. HP fully restored.', 'You enter the battlefield...', 'A Goblin Scout appears!'];
    player.update(p => ({ ...p, hp: p.maxHp }));
    addLogEntry('You returned to Base Camp and recovered.');
  }
</script>

<div class="view-combat">
  <div class="page-header">
    <h1 class="page-title">&#x2694;&#xFE0F; Engagement</h1>
    <p class="page-subtitle">Neutralize hostiles, secure objectives</p>
  </div>

  <!-- Death overlay -->
  {#if isDead}
    <div class="death-overlay card">
      <div class="death-icon">💀</div>
      <h2 class="death-title">Operative Down</h2>
      <p class="death-msg">You were defeated in battle. Return to Base Camp to recover.</p>
      <button class="action-btn return-btn" on:click={returnToBase}>
        <span>🏕️</span>
        <span>Return to Base Camp</span>
      </button>
    </div>
  {/if}

  <!-- Arena -->
  <div class="card arena-card">
    <div class="combatants">
      <!-- Player side -->
      <div class="combatant player-side">
        <div class="combatant-icon">&#x1FA96;</div>
        <div class="combatant-name">{$player.name}</div>
        <div class="combatant-level">Rank {$player.level}</div>
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
        <div class="combatant-level">Threat Lv. {enemy.level}</div>
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
      <span class="card-icon">&#x1F3AF;</span>
      <h2 class="card-title">Actions</h2>
    </div>
    <div class="combat-actions">
      {#if isDead}
        <p class="dead-notice">⚠️ Operative incapacitated — return to Base Camp.</p>
      {:else if !isBattling}
        <button class="action-btn start-btn" on:click={startBattle}>
          <span>&#x2694;&#xFE0F;</span>
          <span>Engage Target</span>
        </button>
      {:else}
        <button class="action-btn attack-btn" on:click={strike}>
          <span>&#x1F5E1;&#xFE0F;</span>
          <span>Strike</span>
        </button>
        <button class="action-btn magic-btn">
          <span>&#x2622;&#xFE0F;</span>
          <span>Deploy Agent</span>
        </button>
        <button class="action-btn flee-btn" on:click={fleeBattle}>
          <span>&#x1F4A8;</span>
          <span>Fall Back</span>
        </button>
      {/if}
    </div>
  </div>

  <!-- Combat log -->
  <div class="card">
    <div class="card-header">
      <span class="card-icon">&#x1F4E1;</span>
      <h2 class="card-title">Engagement Log</h2>
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
    background: linear-gradient(90deg, #1a5a1a, #2a9e2a);
  }

  .enemy-hp {
    background: linear-gradient(90deg, #992200, var(--color-danger-bright));
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
    font-size: 15px;
    border-color: var(--color-danger);
    color: var(--color-danger-bright);
    letter-spacing: 1px;
    text-transform: uppercase;
  }

  .start-btn:hover {
    background-color: rgba(204, 74, 0, 0.15);
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

  .death-overlay {
    border-color: var(--color-danger);
    background-color: rgba(150, 20, 20, 0.12);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 28px 20px;
    text-align: center;
  }

  .death-icon {
    font-size: 52px;
    line-height: 1;
  }

  .death-title {
    font-family: var(--font-heading);
    font-size: 20px;
    font-weight: 700;
    color: var(--color-danger-bright);
    margin: 0;
  }

  .death-msg {
    font-size: 13px;
    color: var(--color-text-muted);
    margin: 0;
  }

  .return-btn {
    margin-top: 4px;
    padding: 12px 28px;
    border-color: var(--color-magic);
    color: var(--color-magic-bright);
    font-size: 15px;
    font-weight: 600;
    letter-spacing: 0.5px;
  }

  .return-btn:hover {
    background-color: rgba(80, 80, 200, 0.15);
  }

  .dead-notice {
    font-size: 13px;
    color: var(--color-danger-bright);
    margin: 0;
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
