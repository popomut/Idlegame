<script>
  import { onMount, onDestroy } from 'svelte';
  import { combatState, fetchCombatStatus, fleeCombatSession, startPolling, stopPolling, dismissOfflineGains } from '../stores/combat.js';
  import { navigateTo } from '../stores/navigation.js';
  import { syncCharacter } from '../stores/game.js';

  let fleeLoading = false;

  function hpPercent(current, max) {
    if (!max || max <= 0) return 0;
    return Math.max(0, Math.min(100, Math.round((current / max) * 100)));
  }

  function formatDuration(startMs) {
    if (!startMs) return '0s';
    const secs = Math.floor((Date.now() - startMs) / 1000);
    if (secs < 60) return `${secs}s`;
    const m = Math.floor(secs / 60);
    const s = secs % 60;
    return `${m}m ${s}s`;
  }

  function formatOfflineTime(ms) {
    if (!ms) return '';
    const secs = Math.floor(ms / 1000);
    if (secs < 60) return `${secs} seconds`;
    const m = Math.floor(secs / 60);
    const s = secs % 60;
    return s > 0 ? `${m}m ${s}s` : `${m} minutes`;
  }

  function logTypeClass(type) {
    return {
      strike: 'log-strike',
      hit: 'log-hit',
      defeat: 'log-defeat',
      spawn: 'log-spawn',
      death: 'log-death',
      info: 'log-info',
    }[type] || 'log-info';
  }

  async function flee() {
    fleeLoading = true;
    try {
      await fleeCombatSession();
      await syncCharacter();
    } finally {
      fleeLoading = false;
    }
  }

  function goToMap() {
    stopPolling();
    navigateTo('map');
  }

  onMount(async () => {
    await fetchCombatStatus();
    if ($combatState.isActive) {
      startPolling();
    }
  });

  onDestroy(() => {
    // Keep polling alive in background — combat continues while on other pages
    // Only stop if status is terminal
    if (!$combatState.isActive) {
      stopPolling();
    }
  });
</script>

<div class="view-combat">
  <!-- Offline Gains Banner -->
  {#if $combatState.wasOffline && $combatState.offlineEnemies > 0}
    <div class="offline-banner card">
      <div class="offline-header">
        <span class="offline-icon">⏰</span>
        <h3 class="offline-title">Welcome Back!</h3>
      </div>
      <p class="offline-body">
        While you were away for <strong>{formatOfflineTime($combatState.offlineTimeMS)}</strong>,
        your character defeated <strong>{$combatState.offlineEnemies} enemies</strong>.
      </p>
      <button class="dismiss-btn" on:click={dismissOfflineGains}>Got it!</button>
    </div>
  {/if}

  {#if $combatState.status === 'none'}
    <!-- No active combat -->
    <div class="card no-combat-card">
      <div class="no-combat-icon">🗺️</div>
      <p class="no-combat-msg">No active combat. Enter a zone from the World Map.</p>
      <button class="action-btn map-btn" on:click={goToMap}>Go to World Map</button>
    </div>

  {:else}
    <!-- ── Arena ──────────────────────────────────────────────────────── -->
    <div class="card arena-card">
      <!-- Player side -->
      <div class="combatants">
        <div class="combatant player-side">
          <div class="combatant-icon">🪖</div>
          <div class="combatant-name">You</div>
          <div class="hp-bar-track">
            <div
              class="hp-bar-fill player-hp"
              style="width: {hpPercent($combatState.playerHPCurrent, $combatState.playerMaxHP)}%"
            ></div>
          </div>
          <div class="hp-text">{$combatState.playerHPCurrent} / {$combatState.playerMaxHP}</div>
        </div>

        <div class="vs-divider">VS</div>

        <!-- Enemy side -->
        <div class="combatant enemy-side">
          {#if $combatState.currentEnemy}
            <div class="combatant-icon">{$combatState.currentEnemy.icon}</div>
            <div class="combatant-name">{$combatState.currentEnemy.name}</div>
            <div class="hp-bar-track">
              <div
                class="hp-bar-fill enemy-hp"
                style="width: {hpPercent($combatState.currentEnemy.hp_current, $combatState.currentEnemy.hp_max)}%"
              ></div>
            </div>
            <div class="hp-text">{$combatState.currentEnemy.hp_current} / {$combatState.currentEnemy.hp_max}</div>
          {:else}
            <div class="combatant-icon">❓</div>
            <div class="combatant-name">No enemy</div>
          {/if}
        </div>
      </div>
    </div>

    <!-- ── Stats row ─────────────────────────────────────────────────── -->
    <div class="stats-row card">
      <div class="stat-item">
        <div class="stat-value">{$combatState.enemiesDefeated}</div>
        <div class="stat-label">Enemies Killed</div>
      </div>
      <div class="stat-item">
        <div class="stat-value">+{$combatState.totalXPGained}</div>
        <div class="stat-label">XP Gained</div>
      </div>
      <div class="stat-item">
        <div class="stat-value">+{$combatState.totalMoneyGained}💰</div>
        <div class="stat-label">Gold Gained</div>
      </div>
      <div class="stat-item">
        <div class="stat-value">{formatDuration($combatState.sessionStartedAt)}</div>
        <div class="stat-label">Duration</div>
      </div>
    </div>

    <!-- ── Action buttons ─────────────────────────────────────────────── -->
    <div class="card actions-card">
      {#if $combatState.status === 'active'}
        <button class="action-btn flee-btn" on:click={flee} disabled={fleeLoading}>
          {fleeLoading ? '⏳ Fleeing…' : '🏃 Flee Combat'}
        </button>
      {:else if $combatState.status === 'dead'}
        <div class="terminal-msg death-msg">
          <span class="terminal-icon">💀</span>
          <span>You were defeated!</span>
        </div>
        <button class="action-btn map-btn" on:click={goToMap}>Back to World Map</button>
      {:else if $combatState.status === 'fled'}
        <div class="terminal-msg fled-msg">
          <span class="terminal-icon">🏃</span>
          <span>You fled from combat.</span>
        </div>
        <button class="action-btn map-btn" on:click={goToMap}>Back to World Map</button>
      {/if}
    </div>

    <!-- ── Combat log ─────────────────────────────────────────────────── -->
    <div class="card log-card">
      <div class="card-header">
        <span class="card-icon">📋</span>
        <h2 class="card-title">Combat Log</h2>
      </div>
      <ul class="combat-log">
        {#each $combatState.recentLogs as entry (entry.timestamp + entry.message)}
          <li class="log-entry {logTypeClass(entry.type)}">{entry.message}</li>
        {/each}
        {#if $combatState.recentLogs.length === 0}
          <li class="log-entry log-info">No combat events yet…</li>
        {/if}
      </ul>
    </div>
  {/if}
</div>

<style>
  .view-combat {
    padding: 20px 16px;
    display: flex;
    flex-direction: column;
    gap: 14px;
    max-width: 760px;
  }

  .card {
    background-color: var(--color-bg-panel);
    border: 1px solid var(--color-border);
    border-radius: 10px;
    padding: 16px;
  }

  /* ── Offline banner ── */
  .offline-banner {
    border-color: var(--color-gold-dim);
    background-color: rgba(180, 130, 0, 0.08);
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .offline-header {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .offline-icon { font-size: 20px; }

  .offline-title {
    font-family: var(--font-heading);
    font-size: 16px;
    font-weight: 700;
    color: var(--color-gold);
    margin: 0;
  }

  .offline-body {
    font-size: 13px;
    color: var(--color-text-muted);
    margin: 0;
  }

  .dismiss-btn {
    align-self: flex-end;
    padding: 8px 18px;
    background: var(--color-bg-elevated);
    border: 1px solid var(--color-gold-dim);
    border-radius: 6px;
    color: var(--color-gold);
    cursor: pointer;
    font-size: 13px;
    font-family: var(--font-body);
  }

  .dismiss-btn:hover { background: rgba(180, 130, 0, 0.15); }

  /* ── No combat ── */
  .no-combat-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 32px 20px;
    text-align: center;
  }

  .no-combat-icon { font-size: 48px; }
  .no-combat-msg { font-size: 14px; color: var(--color-text-muted); margin: 0; }

  /* ── Arena ── */
  .arena-card { border-color: var(--color-danger); }

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
    font-size: 44px;
    line-height: 1;
    padding: 10px;
    background-color: var(--color-bg-elevated);
    border-radius: 50%;
    border: 2px solid var(--color-border);
  }

  .player-side .combatant-icon { border-color: var(--color-magic); }
  .enemy-side .combatant-icon  { border-color: var(--color-danger); }

  .combatant-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--color-text-heading);
  }

  .hp-bar-track {
    width: 100%;
    height: 10px;
    background-color: var(--color-bg-elevated);
    border-radius: 5px;
    overflow: hidden;
  }

  .hp-bar-fill {
    height: 100%;
    border-radius: 5px;
    transition: width 0.6s ease;
  }

  .player-hp { background: linear-gradient(90deg, #1a5a1a, #2a9e2a); }
  .enemy-hp  { background: linear-gradient(90deg, #992200, var(--color-danger-bright)); }

  .hp-text {
    font-size: 12px;
    color: var(--color-text-muted);
  }

  .vs-divider {
    font-family: var(--font-heading);
    font-size: 18px;
    font-weight: 700;
    color: var(--color-gold-dim);
    flex-shrink: 0;
  }

  /* ── Stats row ── */
  .stats-row {
    display: flex;
    justify-content: space-around;
    gap: 8px;
    flex-wrap: wrap;
  }

  .stat-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
  }

  .stat-value {
    font-family: var(--font-heading);
    font-size: 18px;
    font-weight: 700;
    color: var(--color-gold);
  }

  .stat-label {
    font-size: 11px;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  /* ── Actions ── */
  .actions-card {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .action-btn {
    padding: 12px 24px;
    border-radius: 8px;
    border: 1px solid var(--color-border);
    background-color: var(--color-bg-elevated);
    color: var(--color-text);
    cursor: pointer;
    font-size: 14px;
    font-family: var(--font-body);
    font-weight: 600;
    transition: background-color 0.15s, border-color 0.15s, color 0.15s;
  }

  .action-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .flee-btn {
    border-color: var(--color-danger);
    color: var(--color-danger-bright);
  }

  .flee-btn:hover:not(:disabled) { background-color: rgba(204, 74, 0, 0.15); }

  .map-btn {
    border-color: var(--color-magic);
    color: var(--color-magic-bright);
  }

  .map-btn:hover { background-color: rgba(80, 80, 200, 0.15); }

  .terminal-msg {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 15px;
    font-weight: 600;
  }

  .death-msg { color: var(--color-danger-bright); }
  .fled-msg  { color: var(--color-text-muted); }
  .terminal-icon { font-size: 20px; }

  /* ── Combat log ── */
  .card-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
  }

  .card-icon { font-size: 16px; }

  .card-title {
    font-family: var(--font-heading);
    font-size: 14px;
    font-weight: 600;
    color: var(--color-text-heading);
    margin: 0;
  }

  .combat-log {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 5px;
    max-height: 220px;
    overflow-y: auto;
  }

  .log-entry {
    font-size: 13px;
    line-height: 1.5;
    color: var(--color-text-muted);
    padding: 2px 4px;
    border-radius: 3px;
  }

  .log-strike { color: var(--color-gold); }
  .log-hit    { color: var(--color-danger-bright); }
  .log-defeat { color: #55ee55; }
  .log-spawn  { color: var(--color-text); }
  .log-death  { color: var(--color-danger-bright); font-weight: 600; }
  .log-info   { color: var(--color-text-muted); }
</style>
