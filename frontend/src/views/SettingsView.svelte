<script>
  import { player } from '../stores/game.js';
  import { theme } from '../stores/theme.js';

  let playerName = $player.name;
  let notificationsEnabled = true;
  let autoSaveEnabled = true;

  function saveName() {
    player.update(function (p) {
      return { ...p, name: playerName };
    });
  }

  function handleThemeToggle(e) {
    // e.target.checked is already the NEW state after toggle
    const newTheme = e.target.checked ? 'dark' : 'light';
    theme.setTheme(newTheme);
  }
</script>

<div class="view-settings">
  <div class="page-header">
    <h1 class="page-title">&#x2699;&#xFE0F; Operations</h1>
    <p class="page-subtitle">Configure field parameters</p>
  </div>

  <!-- Player settings -->
  <div class="card">
    <div class="card-header">
      <span class="card-icon">&#x1FA96;</span>
      <h2 class="card-title">Soldier Profile</h2>
    </div>

    <div class="setting-row">
      <label class="setting-label" for="player-name">Call Sign</label>
      <div class="name-input-group">
        <input
          id="player-name"
          class="text-input"
          type="text"
          bind:value={playerName}
          maxlength="20"
          placeholder="Enter call sign..."
        />
        <button class="save-btn" on:click={saveName}>Confirm</button>
      </div>
    </div>
  </div>

  <!-- Game settings -->
  <div class="card">
    <div class="card-header">
      <span class="card-icon">&#x1F3AE;</span>
      <h2 class="card-title">Field Settings</h2>
    </div>

    <div class="setting-row">
      <div class="setting-info">
        <span class="setting-label">Alert System</span>
        <span class="setting-desc">Receive alerts for completed objectives</span>
      </div>
      <label class="toggle">
        <input type="checkbox" bind:checked={notificationsEnabled} />
        <span class="toggle-track"></span>
      </label>
    </div>

    <div class="setting-row">
      <div class="setting-info">
        <span class="setting-label">Auto-Sync</span>
        <span class="setting-desc">Automatically sync mission progress</span>
      </div>
      <label class="toggle">
        <input type="checkbox" bind:checked={autoSaveEnabled} />
        <span class="toggle-track"></span>
      </label>
    </div>

    <div class="setting-row">
      <div class="setting-info">
        <span class="setting-label">Night Vision</span>
        <span class="setting-desc">Toggle display mode</span>
      </div>
      <label class="toggle">
        <input type="checkbox" checked={$theme === 'dark'} on:change={handleThemeToggle} />
        <span class="toggle-track"></span>
      </label>
    </div>
  </div>

  <!-- Data settings -->
  <div class="card danger-card">
    <div class="card-header">
      <span class="card-icon">&#x26A0;&#xFE0F;</span>
      <h2 class="card-title">Data</h2>
    </div>
    <div class="setting-row">
      <div class="setting-info">
        <span class="setting-label">Wipe Mission Data</span>
        <span class="setting-desc">Delete all field data and restart from scratch</span>
      </div>
      <button class="danger-btn">Wipe</button>
    </div>
  </div>
</div>

<style>
  .view-settings {
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
    color: var(--color-text-muted);
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

  .danger-card {
    border-color: rgba(220, 38, 38, 0.4);
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

  .setting-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 10px 0;
    border-bottom: 1px solid var(--color-border-subtle);
  }

  .setting-row:last-child {
    border-bottom: none;
    padding-bottom: 0;
  }

  .setting-label {
    font-size: 14px;
    color: var(--color-text);
    font-weight: 500;
    display: block;
  }

  .setting-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .setting-desc {
    font-size: 12px;
    color: var(--color-text-muted);
  }

  /* Name input */
  .name-input-group {
    display: flex;
    gap: 8px;
  }

  .text-input {
    background-color: var(--color-bg-elevated);
    border: 1px solid var(--color-border);
    border-radius: 6px;
    color: var(--color-text);
    font-size: 14px;
    font-family: var(--font-body);
    padding: 6px 10px;
    width: 140px;
    outline: none;
    transition: border-color var(--transition-fast);
  }

  .text-input:focus {
    border-color: var(--color-magic);
  }

  .save-btn {
    background-color: var(--color-bg-elevated);
    border: 1px solid var(--color-magic);
    border-radius: 3px;
    color: var(--color-magic-bright);
    cursor: pointer;
    font-size: 13px;
    font-family: var(--font-body);
    padding: 6px 14px;
    transition: background-color var(--transition-fast);
    letter-spacing: 1px;
    text-transform: uppercase;
  }

  .save-btn:hover {
    background-color: rgba(42, 158, 42, 0.15);
  }

  /* Toggle switch */
  .toggle {
    position: relative;
    display: inline-flex;
    cursor: pointer;
    flex-shrink: 0;
  }

  .toggle input {
    position: absolute;
    opacity: 0;
    width: 0;
    height: 0;
  }

  .toggle-track {
    width: 44px;
    height: 24px;
    background-color: var(--color-bg-elevated);
    border: 1px solid var(--color-border);
    border-radius: 12px;
    transition: background-color var(--transition-fast), border-color var(--transition-fast);
  }

  .toggle-track::after {
    content: '';
    position: absolute;
    top: 3px;
    left: 3px;
    width: 18px;
    height: 18px;
    background-color: var(--color-text-muted);
    border-radius: 50%;
    transition: transform var(--transition-fast), background-color var(--transition-fast);
  }

  .toggle input:checked + .toggle-track {
    background-color: rgba(42, 158, 42, 0.25);
    border-color: var(--color-magic);
  }

  .toggle input:checked + .toggle-track::after {
    transform: translateX(20px);
    background-color: var(--color-magic-bright);
  }

  /* Danger button */
  .danger-btn {
    padding: 7px 16px;
    background: none;
    border: 1px solid var(--color-danger);
    border-radius: 6px;
    color: var(--color-danger-bright);
    cursor: pointer;
    font-size: 13px;
    font-family: var(--font-body);
    font-weight: 500;
    transition: background-color var(--transition-fast);
    flex-shrink: 0;
  }

  .danger-btn:hover {
    background-color: rgba(220, 38, 38, 0.15);
  }
</style>
