<script>
  import { onMount } from 'svelte';
  import EquipmentManagementView from './EquipmentManagementView.svelte';
  import MonsterManagementView from './MonsterManagementView.svelte';
  import OreManagementView from './OreManagementView.svelte';
  import MonsterDropManagementView from './MonsterDropManagementView.svelte';
  import MiningLevelManagementView from './MiningLevelManagementView.svelte';

  let currentTab = 'equipment'; // 'equipment', 'monsters', 'ores', 'monster-drops', 'mining-levels'

  function switchTab(tab) {
    currentTab = tab;
  }
</script>

<div class="view-admin">
  <div class="admin-header">
    <h1 class="admin-title">⚙️ Admin Panel</h1>
    <div class="admin-subtitle">Development Only - Delete Before Production</div>
  </div>

  <div class="card tabs-card">
    <div class="tabs-container">
      <button
        class="tab-btn"
        class:active={currentTab === 'equipment'}
        on:click={() => switchTab('equipment')}
      >
        🛡️ Equipment Management
      </button>
      <button
        class="tab-btn"
        class:active={currentTab === 'monsters'}
        on:click={() => switchTab('monsters')}
      >
        👹 Monster Management
      </button>
      <button
        class="tab-btn"
        class:active={currentTab === 'ores'}
        on:click={() => switchTab('ores')}
      >
        ⛏️ Ores Management
      </button>
      <button
        class="tab-btn"
        class:active={currentTab === 'monster-drops'}
        on:click={() => switchTab('monster-drops')}
      >
        🎁 Monster Drops
      </button>
      <button
        class="tab-btn"
        class:active={currentTab === 'mining-levels'}
        on:click={() => switchTab('mining-levels')}
      >
        ⛏️ Mining Levels
      </button>
    </div>
  </div>

  {#if currentTab === 'equipment'}
    <EquipmentManagementView />
  {:else if currentTab === 'monsters'}
    <MonsterManagementView />
  {:else if currentTab === 'ores'}
    <OreManagementView />
  {:else if currentTab === 'monster-drops'}
    <MonsterDropManagementView />
  {:else if currentTab === 'mining-levels'}
    <MiningLevelManagementView />
  {/if}
</div>

<style>
  .view-admin {
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0;
  }

  .admin-header {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 20px 16px 0;
    margin-bottom: 12px;
  }

  .admin-title {
    font-size: 24px;
    font-weight: 700;
    color: var(--color-text-heading);
    margin: 0;
  }

  .admin-subtitle {
    font-size: 13px;
    color: var(--color-text-muted);
  }

  .card {
    background-color: var(--color-bg-panel);
    border: 1px solid var(--color-border);
    border-radius: 0;
    margin: 0 16px;
    padding: 0;
    border-bottom: 1px solid var(--color-border);
  }

  .tabs-container {
    display: flex;
    gap: 0;
  }

  .tab-btn {
    flex: 1;
    padding: 12px 16px;
    border: none;
    background: none;
    color: var(--color-text-muted);
    cursor: pointer;
    font-size: 14px;
    font-family: var(--font-body);
    font-weight: 600;
    border-bottom: 2px solid transparent;
    transition: all 0.15s;
  }

  .tab-btn:hover {
    background-color: rgba(80, 80, 200, 0.05);
    color: var(--color-text);
  }

  .tab-btn.active {
    color: var(--color-magic);
    border-bottom-color: var(--color-magic);
  }
</style>
