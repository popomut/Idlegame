<script>
  import { canGoBack, goBack, goHome } from '../stores/navigation.js';
  import { activeMining, miningProgress } from '../stores/mining.js';
</script>

<nav class="bottom-bar" aria-label="Bottom navigation">
  <button
    class="bottom-btn back-btn"
    on:click={goBack}
    disabled={!$canGoBack}
    aria-label="Go back"
  >
    <span class="btn-icon">&#x276E;</span>
    <span class="btn-label">Back</span>
  </button>

  <div class="bottom-center">
    {#if $activeMining}
      <div class="activity-status">
        <span class="status-icon">⛏️</span>
        <div class="status-content">
          <span class="status-text">Mining {$activeMining.oreName}...</span>
          <div class="progress-bar-small">
            <div class="progress-fill" style="width: {$miningProgress}%"></div>
          </div>
        </div>
      </div>
    {:else}
      <span class="realm-icon">&#x2605;</span>
    {/if}
  </div>

  <button
    class="bottom-btn home-btn"
    on:click={goHome}
    aria-label="Go to home"
  >
    <span class="btn-label">Home</span>
    <span class="btn-icon">&#x2302;</span>
  </button>
</nav>

<style>
  .bottom-bar {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    height: var(--bottombar-height);
    background-color: var(--color-bg-panel);
    border-top: 1px solid var(--color-border);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 16px;
    z-index: 100;
  }

  .bottom-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    background: none;
    border: none;
    color: var(--color-text-muted);
    cursor: pointer;
    font-size: 15px;
    font-family: var(--font-body);
    padding: 8px 14px;
    border-radius: 8px;
    transition: background-color var(--transition-fast), color var(--transition-fast);
    min-width: 80px;
  }

  .bottom-btn:hover:not(:disabled) {
    background-color: var(--color-bg-hover);
    color: var(--color-text);
  }

  .bottom-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  .back-btn {
    justify-content: flex-start;
  }

  .home-btn {
    justify-content: flex-end;
    color: var(--color-gold-dim);
  }

  .home-btn:hover:not(:disabled) {
    color: var(--color-gold);
  }

  .btn-icon {
    font-size: 18px;
    line-height: 1;
  }

  .btn-label {
    font-weight: 500;
  }

  .bottom-center {
    display: flex;
    align-items: center;
    justify-content: center;
    flex: 1;
  }

  .realm-icon {
    font-size: 22px;
    color: var(--color-border);
  }

  .activity-status {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 12px;
    background-color: var(--color-bg-elevated);
    border-radius: 8px;
    border: 1px solid var(--color-border-subtle);
    max-width: 300px;
  }

  .status-icon {
    font-size: 18px;
    animation: pulse 1s ease-in-out infinite;
  }

  .status-content {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .status-text {
    font-size: 14px;
    color: var(--color-magic-bright);
    font-weight: 500;
  }

  .progress-bar-small {
    width: 100px;
    height: 6px;
    background-color: var(--color-bg-panel);
    border-radius: 3px;
    overflow: hidden;
    border: 1px solid var(--color-border-subtle);
  }

  .progress-fill {
    height: 100%;
    background: linear-gradient(90deg, var(--color-magic-bright), var(--color-gold-bright));
    transition: width 0.05s linear;
  }

  @keyframes pulse {
    0%, 100% {
      opacity: 1;
    }
    50% {
      opacity: 0.6;
    }
  }
</style>
