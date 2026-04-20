<script>
  import { offlineGains, syncOreInventory } from '../stores/mining.js';
  import { fade, scale } from 'svelte/transition';

  async function handleAwesomeClick() {
    await syncOreInventory();
    offlineGains.set(null);
  }
</script>

{#if $offlineGains?.wasOffline}
  <div class="offline-popup" in:fade out:fade>
    <div class="popup-content" in:scale out:scale>
      <div class="popup-icon">⚡</div>
      <h2 class="popup-title">Welcome Back!</h2>
      <p class="popup-text">
        You earned <span class="ore-count">{$offlineGains.oresGained}</span>
        <span class="ore-name">{$offlineGains.oreName}</span>
        while away!
      </p>
      <p class="popup-time">
        ({Math.round($offlineGains.timeMs / 1000 / 60)} minutes offline)
      </p>
      <button class="popup-btn" on:click={handleAwesomeClick}>
        Awesome!
      </button>
    </div>
  </div>
{/if}

<style>
  .offline-popup {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .popup-content {
    background: var(--color-bg-panel);
    border: 2px solid var(--color-gold-dim);
    border-radius: 12px;
    padding: 24px;
    text-align: center;
    max-width: 400px;
  }

  .popup-icon {
    font-size: 48px;
    margin-bottom: 12px;
    animation: bounce 0.6s ease-out;
  }

  .popup-title {
    font-family: var(--font-heading);
    font-size: 24px;
    margin-bottom: 8px;
    color: var(--color-text-heading);
  }

  .popup-text {
    font-size: 16px;
    color: var(--color-text);
    margin-bottom: 4px;
  }

  .ore-count {
    font-weight: 700;
    color: var(--color-gold-bright);
  }

  .ore-name {
    color: var(--color-magic-bright);
  }

  .popup-time {
    font-size: 13px;
    color: var(--color-text-muted);
    margin-bottom: 16px;
  }

  .popup-btn {
    padding: 10px 24px;
    background: var(--color-gold-dim);
    color: #000;
    border: none;
    border-radius: 6px;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s;
  }

  .popup-btn:hover {
    background: var(--color-gold-bright);
  }

  @keyframes bounce {
    0% {
      transform: scale(0) rotate(0);
    }
    50% {
      transform: scale(1.2);
    }
    100% {
      transform: scale(1);
    }
  }
</style>
