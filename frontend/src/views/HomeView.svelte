<script>
  import { onMount, onDestroy } from 'svelte';
  import { get } from 'svelte/store';
  import { player, activityLog } from '../stores/game.js';
  import { characterAPI } from '../services/api.js';
  import { navigateTo } from '../stores/navigation.js';

  function pct(val, max) {
    if (!max) return 0;
    return Math.min(100, Math.round((val / max) * 100));
  }

  // ── HP regen: +1 HP/sec while on Base Camp, persist to DB every 30s ───
  let regenInterval = null;
  let ticksSinceSave = 0;
  const PERSIST_EVERY = 30; // seconds between DB saves

  onMount(() => {
    regenInterval = setInterval(() => {
      player.update(p => {
        if (p.hp >= p.maxHp) return p;
        const newHp = Math.min(p.hp + 1, p.maxHp);
        ticksSinceSave++;
        if (ticksSinceSave >= PERSIST_EVERY || newHp === p.maxHp) {
          ticksSinceSave = 0;
          characterAPI.heal(newHp).catch(() => {});
        }
        return { ...p, hp: newHp };
      });
    }, 1000);
  });

  onDestroy(() => {
    if (regenInterval) {
      clearInterval(regenInterval);
      // Persist whatever HP the player is at when they leave
      const currentHp = get(player).hp;
      characterAPI.heal(currentHp).catch(() => {});
    }
  });
</script>

<div class="view-home">
  <div class="page-header">
    <h1 class="page-title">&#x1FA96; OPERATIVE: {$player.name}</h1>
    <p class="page-subtitle">{$player.class}</p>
  </div>

  <!-- Operator Profile card -->
  <div class="card character-card">
    <div class="card-header">
      <span class="card-icon">&#x1FA96;</span>
      <h2 class="card-title">Operator Profile</h2>
      <span class="level-tag">RANK {$player.level}</span>
    </div>

    <div class="stat-bars">
      <!-- XP -->
      <div class="stat-row">
        <span class="stat-label">&#x1F3AF; EXP</span>
        <div class="stat-bar-track">
          <div class="stat-bar-fill xp-fill" style="width: {pct($player.xp, $player.xpRequired)}%"></div>
        </div>
        <span class="stat-value">{$player.xp} / {$player.xpRequired}</span>
      </div>

      <!-- HP -->
      <div class="stat-row">
        <span class="stat-label">&#x2764; HP</span>
        <div class="stat-bar-track">
          <div class="stat-bar-fill hp-fill" style="width: {pct($player.hp, $player.maxHp)}%"></div>
        </div>
        <span class="stat-value">{$player.hp} / {$player.maxHp}</span>
      </div>

      <!-- Stamina -->
      <div class="stat-row">
        <span class="stat-label">&#x26A1; STAMINA</span>
        <div class="stat-bar-track">
          <div class="stat-bar-fill stamina-fill" style="width: {pct($player.stamina, $player.maxStamina)}%"></div>
        </div>
        <span class="stat-value">{$player.stamina} / {$player.maxStamina}</span>
      </div>
    </div>
  </div>

  <!-- Combat stats card -->
  <div class="card stats-card">
    <div class="card-header">
      <span class="card-icon">&#x2694;&#xFE0F;</span>
      <h2 class="card-title">Operator Stats</h2>
      <span class="money-tag">&#x1F4B0; {$player.money}</span>
    </div>
    <div class="stat-grid">
      <div class="stat-box">
        <span class="stat-box-icon">&#x1F4AA;</span>
        <span class="stat-box-label">STR</span>
        <span class="stat-box-value">{$player.str}</span>
      </div>
      <div class="stat-box">
        <span class="stat-box-icon">&#x1F9E0;</span>
        <span class="stat-box-label">INT</span>
        <span class="stat-box-value">{$player.int}</span>
      </div>
      <div class="stat-box">
        <span class="stat-box-icon">&#x1F3AF;</span>
        <span class="stat-box-label">DEX</span>
        <span class="stat-box-value">{$player.dex}</span>
      </div>
    </div>
  </div>

  <!-- Quick actions -->
  <div class="card">
    <div class="card-header">
      <span class="card-icon">&#x1F3AF;</span>
      <h2 class="card-title">Tactical Options</h2>
    </div>
    <div class="quick-actions">
      <button class="action-btn danger-btn" on:click={() => navigateTo('combat')}>
        <span class="action-icon">&#x2694;&#xFE0F;</span>
        <span>Engage</span>
      </button>
      <button class="action-btn magic-btn" on:click={() => navigateTo('skills')}>
        <span class="action-icon">&#x1F4AA;</span>
        <span>Train</span>
      </button>
      <button class="action-btn gold-btn" on:click={() => navigateTo('shop')}>
        <span class="action-icon">&#x1F4B0;</span>
        <span>Black Market</span>
      </button>
      <button class="action-btn" on:click={() => navigateTo('inventory')}>
        <span class="action-icon">&#x1F392;</span>
        <span>Field Kit</span>
      </button>
    </div>
  </div>

  <!-- Activity log -->
  <div class="card activity-card">
    <div class="card-header">
      <span class="card-icon">&#x1F4E1;</span>
      <h2 class="card-title">Comms Log</h2>
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
    color: var(--color-magic-bright);
    margin: 0;
    letter-spacing: 1.5px;
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
    font-weight: 700;
    letter-spacing: 1px;
  }

  .money-tag {
    font-size: 13px;
    color: var(--color-gold);
    font-weight: 600;
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
    font-size: 12px;
    color: var(--color-text-muted);
    width: 72px;
    flex-shrink: 0;
    text-transform: uppercase;
    letter-spacing: 0.5px;
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

  .xp-fill { background: linear-gradient(90deg, var(--color-gold-dim), var(--color-gold)); }
  .hp-fill  { background: linear-gradient(90deg, #1a5a1a, #2a9e2a); }
  .stamina-fill { background: linear-gradient(90deg, var(--color-magic), var(--color-magic-bright)); }

  .stat-value {
    font-size: 12px;
    color: var(--color-text-muted);
    width: 78px;
    text-align: right;
    flex-shrink: 0;
  }

  /* Combat stats grid */
  .stats-card { border-color: var(--color-border); }

  .stat-grid {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    gap: 10px;
  }

  .stat-box {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 12px 8px;
    background-color: var(--color-bg-elevated);
    border: 1px solid var(--color-border-subtle);
    border-radius: 8px;
  }

  .stat-box-icon { font-size: 20px; }

  .stat-box-label {
    font-size: 10px;
    color: var(--color-text-muted);
    letter-spacing: 1.5px;
    text-transform: uppercase;
  }

  .stat-box-value {
    font-size: 20px;
    font-weight: 700;
    color: var(--color-text-heading);
    font-family: var(--font-heading);
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

  .action-btn:hover { background-color: var(--color-bg-hover); border-color: var(--color-text-muted); }
  .danger-btn:hover { border-color: var(--color-danger); color: var(--color-danger-bright); }
  .magic-btn:hover  { border-color: var(--color-magic); color: var(--color-magic-bright); }
  .gold-btn:hover   { border-color: var(--color-gold-dim); color: var(--color-gold); }

  .action-icon { font-size: 18px; }

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
