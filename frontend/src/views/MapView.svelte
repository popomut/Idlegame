<script>
  import { onMount, onDestroy } from 'svelte';
  import { mapAPI } from '../services/api.js';
  import { backOverride, navigateTo } from '../stores/navigation.js';
  import { startCombatSession } from '../stores/combat.js';

  // ── State ──────────────────────────────────────────────────────────────
  let view = 'continents';    // 'continents' | 'areas'
  let continents = [];
  let selectedContinent = null;
  let pendingArea = null;
  let loading = true;
  let actionLoading = false;
  let errorMsg = '';

  // ── Load ───────────────────────────────────────────────────────────────
  onMount(async function () {
    try {
      const res = await mapAPI.getContinents();
      continents = res.data;
    } catch (e) {
      errorMsg = 'Failed to load map data.';
    } finally {
      loading = false;
    }
  });

  // ── Navigation ─────────────────────────────────────────────────────────
  function selectContinent(cont) { selectedContinent = cont; view = 'areas'; }
  function backToContinents()    { selectedContinent = null; view = 'continents'; }
  function openConfirm(area)     { pendingArea = area; }
  function closeConfirm()        { pendingArea = null; }

  // ── Combat entry ────────────────────────────────────────────────────────
  async function enterArea() {
    if (!pendingArea) return;
    actionLoading = true;
    errorMsg = '';
    try {
      await startCombatSession(pendingArea.area_key);
      closeConfirm();
      navigateTo('combat');
    } catch (e) {
      errorMsg = e?.response?.data?.error || 'Failed to start combat.';
    } finally {
      actionLoading = false;
    }
  }

  // ── Back override ──────────────────────────────────────────────────────
  $: {
    if (view === 'continents') {
      backOverride.set(null);
    } else if (view === 'areas') {
      backOverride.set({ fn: backToContinents, canGoBack: true });
    }
  }

  onDestroy(function () {
    backOverride.set(null);
  });
</script>

<!-- ── Map page ─────────────────────────────────────────────────────────── -->
<div class="view-map">

  {#if loading}
    <div class="loading">Loading map data…</div>

  {:else if view === 'continents'}
    <!-- ── Continent selection ───────────────────────────────────────── -->
    <div class="page-header">
      <h1 class="page-title">🗺️ World Map</h1>
      <p class="page-subtitle">Select a region to deploy</p>
    </div>

    {#if errorMsg}<div class="error-banner">{errorMsg}</div>{/if}

    <div class="continent-grid">
      {#each continents as cont}
        <button class="continent-card" on:click={() => selectContinent(cont)}>
          <span class="cont-icon">{cont.icon}</span>
          <div class="cont-info">
            <span class="cont-name">{cont.name}</span>
            <span class="cont-desc">{cont.description}</span>
          </div>
          <div class="cont-footer">
            <span class="diff-badge">{cont.difficulty.toUpperCase()}</span>
            <span class="area-count">{cont.areas?.length ?? 0} zones</span>
          </div>
        </button>
      {/each}
    </div>

  {:else if view === 'areas' && selectedContinent}
    <!-- ── Area list ─────────────────────────────────────────────────── -->
    <div class="page-header">
      <button class="back-btn" on:click={backToContinents}>← Back</button>
      <div>
        <h1 class="page-title">{selectedContinent.icon} {selectedContinent.name}</h1>
        <p class="page-subtitle">Choose a zone to enter</p>
      </div>
    </div>

    {#if errorMsg}<div class="error-banner">{errorMsg}</div>{/if}

    <div class="area-list">
      {#each selectedContinent.areas as area}
        <button class="area-card" on:click={() => openConfirm(area)}>
          <div class="area-top">
            <span class="area-icon">{area.icon}</span>
            <div class="area-info">
              <span class="area-name">{area.name}</span>
              <span class="area-diff">{area.difficulty.toUpperCase()}</span>
            </div>
            <span class="area-enter">Enter ›</span>
          </div>
          <p class="area-desc">{area.description}</p>
          <div class="area-meta">
            <span class="meta-pill">⚔️ Endless combat</span>
          </div>
        </button>
      {/each}
    </div>
  {/if}
</div>

<!-- ── Area confirm popup ─────────────────────────────────────────────────── -->
{#if pendingArea}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="popup-backdrop" on:click={closeConfirm}>
    <div class="popup" on:click|stopPropagation>
      <div class="popup-header">
        <span class="popup-area-icon">{pendingArea.icon}</span>
        <div>
          <div class="popup-area-name">{pendingArea.name}</div>
          <div class="popup-area-diff">{pendingArea.difficulty.toUpperCase()}</div>
        </div>
      </div>

      <p class="popup-desc">{pendingArea.description}</p>

      <div class="popup-section">
        <div class="popup-section-label">Zone Monsters</div>
        <div class="popup-monsters">
          {#each (pendingArea.monsters ?? []) as entry}
            {#if entry.monster}
              <span class="popup-monster-pill">{entry.monster.icon} {entry.monster.name}</span>
            {/if}
          {/each}
        </div>
      </div>

      {#if errorMsg}<div class="error-banner" style="margin-top:4px">{errorMsg}</div>{/if}

      <div class="popup-actions">
        <button class="popup-btn enter-btn" on:click={enterArea} disabled={actionLoading}>
          {actionLoading ? 'Starting…' : '⚔️ Enter Zone'}
        </button>
        <button class="popup-btn cancel-btn" on:click={closeConfirm}>Cancel</button>
      </div>
    </div>
  </div>
{/if}

<style>
  /* ── Shell ──────────────────────────────────────────────────────────── */
  .view-map {
    padding: 20px 16px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    max-width: 760px;
  }

  .loading {
    color: var(--color-text-muted);
    padding: 40px 0;
    text-align: center;
    font-size: 14px;
  }

  .error-banner {
    background: rgba(200,50,50,0.15);
    border: 1px solid var(--color-danger);
    color: var(--color-danger-bright);
    padding: 10px 14px;
    border-radius: 8px;
    font-size: 13px;
  }



  .page-header {
    display: flex;
    align-items: flex-start;
    gap: 12px;
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

  .back-btn {
    flex-shrink: 0;
    background: var(--color-bg-elevated);
    border: 1px solid var(--color-border);
    color: var(--color-text-muted);
    padding: 6px 12px;
    border-radius: 8px;
    cursor: pointer;
    font-family: var(--font-body);
    font-size: 13px;
    margin-top: 4px;
    transition: color var(--transition-fast), border-color var(--transition-fast);
  }

  .back-btn:hover { color: var(--color-text); border-color: var(--color-text-muted); }
  .back-btn:disabled { opacity: 0.5; cursor: not-allowed; }

  /* ── Continent grid ─────────────────────────────────────────────────── */
  .continent-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }

  @media (max-width: 500px) {
    .continent-grid { grid-template-columns: 1fr; }
  }

  .continent-card {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 16px;
    background: var(--color-bg-panel);
    border: 1px solid var(--color-border);
    border-radius: 12px;
    cursor: pointer;
    text-align: left;
    font-family: var(--font-body);
    transition: background-color var(--transition-fast), border-color var(--transition-fast);
  }

  .continent-card:hover {
    background: var(--color-bg-hover);
    border-color: var(--color-text-muted);
  }

  .cont-icon { font-size: 36px; }

  .cont-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
    flex: 1;
  }

  .cont-name {
    font-family: var(--font-heading);
    font-size: 16px;
    color: var(--color-text-heading);
    font-weight: 600;
  }

  .cont-desc {
    font-size: 12px;
    color: var(--color-text-muted);
    line-height: 1.4;
  }

  .cont-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .diff-badge {
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 1px;
  }

  .area-count {
    font-size: 11px;
    color: var(--color-text-muted);
  }

  /* ── Area list ──────────────────────────────────────────────────────── */
  .area-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .area-card {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 14px 16px;
    background: var(--color-bg-panel);
    border: 1px solid var(--color-border);
    border-radius: 10px;
    cursor: pointer;
    text-align: left;
    font-family: var(--font-body);
    width: 100%;
    transition: background-color var(--transition-fast), border-color var(--transition-fast);
  }

  .area-card:hover {
    background: var(--color-bg-hover);
    border-color: var(--color-text-muted);
  }

  .area-top {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .area-icon { font-size: 28px; flex-shrink: 0; }

  .area-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .area-name {
    font-family: var(--font-heading);
    font-size: 15px;
    color: var(--color-text-heading);
    font-weight: 600;
  }

  .area-diff { font-size: 11px; font-weight: 700; letter-spacing: 1px; }

  .area-enter { font-size: 18px; color: var(--color-text-muted); }

  .area-desc {
    font-size: 12px;
    color: var(--color-text-muted);
    margin: 0;
    line-height: 1.4;
  }

  .area-meta {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .meta-pill {
    font-size: 11px;
    padding: 2px 8px;
    background: var(--color-bg-elevated);
    border: 1px solid var(--color-border-subtle);
    border-radius: 10px;
    color: var(--color-text-muted);
  }

  .boss-pill { color: var(--color-gold); border-color: var(--color-gold-dim); }

  /* ── Combat arena ───────────────────────────────────────────────────── */
  .card {
    background: var(--color-bg-panel);
    border: 1px solid var(--color-border);
    border-radius: 10px;
    padding: 16px;
  }

  .progress-track {
    position: relative;
    height: 12px;
    background: var(--color-bg-elevated);
    border-radius: 6px;
    overflow: hidden;
    border: 1px solid var(--color-border-subtle);
  }

  .progress-fill {
    height: 100%;
    background: linear-gradient(90deg, #1a4a1a, #3a9a3a);
    border-radius: 6px;
    transition: width 0.4s ease;
  }

  .progress-fill.progress-boss {
    background: linear-gradient(90deg, var(--color-danger), var(--color-danger-bright));
  }

  .progress-label {
    position: absolute;
    right: 8px;
    top: 50%;
    transform: translateY(-50%);
    font-size: 10px;
    color: var(--color-text-muted);
    font-weight: 600;
  }

  /* Monster card */
  .monster-card { display: flex; flex-direction: column; gap: 14px; }

  .boss-card {
    border-color: var(--color-danger);
    box-shadow: 0 0 16px rgba(200,60,60,0.2);
  }

  .boss-banner {
    text-align: center;
    font-family: var(--font-heading);
    font-size: 13px;
    font-weight: 700;
    letter-spacing: 2px;
    color: var(--color-danger-bright);
    padding: 6px;
    background: rgba(200,60,60,0.1);
    border-radius: 6px;
  }

  .monster-header {
    display: flex;
    align-items: flex-start;
    gap: 14px;
  }

  .monster-icon { font-size: 48px; flex-shrink: 0; }

  .monster-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .monster-name {
    font-family: var(--font-heading);
    font-size: 20px;
    font-weight: 700;
    color: var(--color-text-heading);
  }

  .monster-desc {
    font-size: 13px;
    color: var(--color-text-muted);
    line-height: 1.4;
  }

  .monster-stats {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 8px;
  }

  .mstat {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 3px;
    padding: 8px 4px;
    background: var(--color-bg-elevated);
    border-radius: 6px;
    border: 1px solid var(--color-border-subtle);
  }

  .mstat-label { font-size: 10px; color: var(--color-text-muted); text-transform: uppercase; letter-spacing: 1px; }
  .mstat-val   { font-size: 14px; font-weight: 700; color: var(--color-text-heading); font-family: var(--font-heading); }

  .resist-row {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }

  .resist-pill {
    font-size: 11px;
    padding: 2px 8px;
    background: var(--color-bg-elevated);
    border: 1px solid var(--color-border-subtle);
    border-radius: 10px;
    font-weight: 600;
  }

  .combat-actions {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: 10px;
  }

  .action-btn {
    padding: 12px;
    border-radius: 8px;
    border: 1px solid var(--color-border);
    background: var(--color-bg-elevated);
    color: var(--color-text);
    cursor: pointer;
    font-family: var(--font-body);
    font-size: 15px;
    font-weight: 600;
    transition: background-color var(--transition-fast), border-color var(--transition-fast), color var(--transition-fast);
  }

  .action-btn:disabled { opacity: 0.5; cursor: not-allowed; }
  .action-btn:hover:not(:disabled) { background: var(--color-bg-hover); border-color: var(--color-text-muted); }
  .danger-btn:hover:not(:disabled) { border-color: var(--color-danger); color: var(--color-danger-bright); }
  .gold-btn { border-color: var(--color-gold-dim); color: var(--color-gold); }
  .gold-btn:hover:not(:disabled) { background: var(--color-gold-dim); color: var(--color-bg); }

  /* Victory card */
  .victory-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 32px 16px;
    text-align: center;
  }

  .victory-icon { font-size: 56px; }

  .victory-title {
    font-family: var(--font-heading);
    font-size: 24px;
    color: var(--color-gold);
    margin: 0;
  }

  .victory-sub {
    font-size: 14px;
    color: var(--color-text-muted);
    margin: 0;
  }

  /* ── Popup ──────────────────────────────────────────────────────────── */
  .popup-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0,0,0,0.75);
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
    max-width: 400px;
    display: flex;
    flex-direction: column;
    gap: 14px;
    max-height: 90vh;
    overflow-y: auto;
  }

  .popup-header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding-bottom: 12px;
    border-bottom: 1px solid var(--color-border);
  }

  .popup-area-icon { font-size: 36px; }

  .popup-area-name {
    font-family: var(--font-heading);
    font-size: 18px;
    font-weight: 700;
    color: var(--color-text-heading);
  }

  .popup-area-diff { font-size: 11px; font-weight: 700; letter-spacing: 1px; margin-top: 2px; }

  .popup-desc {
    font-size: 13px;
    color: var(--color-text-muted);
    margin: 0;
    line-height: 1.5;
  }

  .popup-section { display: flex; flex-direction: column; gap: 6px; }

  .popup-section-label {
    font-size: 11px;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 1px;
    font-weight: 600;
  }

  .popup-monsters { display: flex; flex-wrap: wrap; gap: 6px; }

  .popup-monster-pill {
    font-size: 12px;
    padding: 3px 10px;
    background: var(--color-bg-elevated);
    border: 1px solid var(--color-border-subtle);
    border-radius: 10px;
    color: var(--color-text-muted);
  }

  .popup-boss-pill {
    display: inline-block;
    font-size: 13px;
    padding: 4px 12px;
    background: rgba(200,60,60,0.1);
    border: 1px solid var(--color-danger);
    border-radius: 10px;
    color: var(--color-danger-bright);
    font-weight: 600;
  }

  .popup-fight-count { font-size: 14px; color: var(--color-text-heading); font-weight: 600; }

  .popup-actions { display: flex; gap: 10px; }

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

  .popup-btn:disabled { opacity: 0.5; cursor: not-allowed; }

  .enter-btn {
    background: var(--color-danger);
    border-color: var(--color-danger-bright);
    color: #fff;
  }

  .enter-btn:hover:not(:disabled) { background: var(--color-danger-bright); }

  .cancel-btn {
    background: var(--color-bg-elevated);
    color: var(--color-text-muted);
  }

  .cancel-btn:hover:not(:disabled) { color: var(--color-text); border-color: var(--color-text-muted); }

  /* ── Live combat card (HP bars side-by-side) ───────────────────────── */
  .live-combat-card {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 16px;
  }

  .combatant-block {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 5px;
  }

  .combatant-label {
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 1px;
    text-transform: uppercase;
    margin-bottom: 2px;
  }

  .you-label   { color: var(--color-gold); }
  .enemy-label { color: var(--color-danger-bright); }

  .vs-badge {
    font-family: var(--font-heading);
    font-size: 13px;
    font-weight: 700;
    color: var(--color-text-muted);
    padding: 20px 4px 0;
    flex-shrink: 0;
  }

  /* HP bars */
  .hp-label-row {
    display: flex;
    justify-content: space-between;
    font-size: 11px;
    color: var(--color-text-muted);
    font-weight: 600;
  }

  .hp-track {
    height: 10px;
    background: var(--color-bg-elevated);
    border-radius: 5px;
    overflow: hidden;
  }

  .hp-fill {
    height: 100%;
    background: #3a9a3a;
    border-radius: 5px;
    transition: width 0.5s ease;
  }

  .enemy-track { }
  .enemy-fill  { background: var(--color-danger); }

  .hp-fill.hp-low        { background: var(--color-danger-bright); }
  .enemy-fill.hp-low     { background: #ff2200; }

  /* ── Live log ────────────────────────────────────────────────────────── */
  .live-log {
    max-height: 280px;
    min-height: 100px;
  }

  .log-entry-in {
    animation: fadeSlideIn 0.25s ease both;
  }

  @keyframes fadeSlideIn {
    from { opacity: 0; transform: translateX(-8px); }
    to   { opacity: 1; transform: translateX(0); }
  }

  .log-typing {
    color: var(--color-text-muted);
    font-size: 14px;
    padding: 4px 0;
    animation: blink 0.8s step-end infinite;
  }

  @keyframes blink {
    50% { opacity: 0; }
  }

  /* Damage badge */
  .dmg-badge {
    margin-left: auto;
    flex-shrink: 0;
    font-size: 11px;
    font-weight: 700;
    padding: 1px 7px;
    border-radius: 8px;
  }

  .dmg-player  { color: #ff8844; background: rgba(255,136,68,0.15); }
  .dmg-monster { color: var(--color-danger-bright); background: rgba(200,60,60,0.15); }

  /* ── Combat result card ─────────────────────────────────────────────── */
  .result-card {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .result-win  { border-color: var(--color-gold-dim); }
  .result-die  { border-color: var(--color-danger); box-shadow: 0 0 16px rgba(200,60,60,0.2); }

  .result-banner {
    text-align: center;
    font-family: var(--font-heading);
    font-size: 16px;
    font-weight: 700;
    letter-spacing: 3px;
    padding: 8px;
    border-radius: 6px;
  }

  .result-win .result-banner { color: var(--color-gold); background: rgba(180,140,0,0.1); }
  .result-die .result-banner { color: var(--color-danger-bright); background: rgba(200,60,60,0.1); }

  .reward-row { display: flex; gap: 8px; flex-wrap: wrap; }

  .reward-pill {
    font-size: 13px;
    font-weight: 700;
    padding: 4px 12px;
    border-radius: 10px;
    border: 1px solid;
  }

  .xp-pill    { color: var(--color-gold); border-color: var(--color-gold-dim); background: rgba(180,140,0,0.1); }
  .money-pill { color: #88ff44; border-color: #44882244; background: rgba(68,136,34,0.1); }

  .result-actions { display: flex; gap: 10px; }

  .auto-advance-msg {
    font-size: 12px;
    color: var(--color-text-muted);
    margin: 4px 0 0;
    letter-spacing: 0.5px;
  }

  /* ── Combat log (shared) ─────────────────────────────────────────────── */
  .combat-log {
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 3px;
    background: var(--color-bg-elevated);
    border-radius: 8px;
    padding: 10px;
    font-size: 12px;
    border: 1px solid var(--color-border-subtle);
  }

  .log-row {
    display: flex;
    gap: 8px;
    align-items: baseline;
    padding: 2px 0;
    border-bottom: 1px solid var(--color-border-subtle);
  }

  .log-row:last-child { border-bottom: none; }

  .log-actor {
    font-size: 9px;
    font-weight: 700;
    letter-spacing: 1px;
    flex-shrink: 0;
    width: 42px;
    text-align: right;
    padding: 1px 4px;
    border-radius: 3px;
  }

  .log-player  .log-actor { color: var(--color-gold); background: rgba(180,140,0,0.15); }
  .log-monster .log-actor { color: var(--color-danger-bright); background: rgba(200,60,60,0.15); }
  .log-system  .log-actor { color: var(--color-text-muted); background: transparent; }

  .log-msg { color: var(--color-text-muted); line-height: 1.4; flex: 1; }
  .log-player .log-msg  { color: var(--color-text); }
  .log-system .log-msg  { color: var(--color-text-muted); font-style: italic; }
</style>
