<script>
  import { onMount } from 'svelte';
  import { adminAPI } from '../services/admin.js';

  let ores = [];
  let loading = true;
  let errorMsg = '';
  let showModal = false;
  let editingId = null;

  let formData = {
    ore_key: '',
    ore_name: '',
    icon: '',
    color: '',
    difficulty: '',
    mining_time_ms: 3000,
    xp_per_ore: 10,
    level_required: 1,
    pickaxe_required: 'none',
    max_quantity: 0,
    sort_order: 0,
  };

  const difficultyOptions = ['Common', 'Uncommon', 'Rare', 'Epic', 'Legendary'];
  const pickaxeOptions = ['none', 'iron_pickaxe', 'gold_pickaxe', 'mithril_pickaxe'];

  onMount(async () => {
    await loadOres();
  });

  async function loadOres() {
    loading = true;
    errorMsg = '';
    try {
      const res = await adminAPI.getAllOres();
      ores = res.data;
    } catch (e) {
      errorMsg = e.response?.data?.error || 'Failed to load ores';
    } finally {
      loading = false;
    }
  }

  function openModal(ore = null) {
    if (ore) {
      editingId = ore.id;
      formData = {
        ore_key: ore.ore_key,
        ore_name: ore.ore_name,
        icon: ore.icon,
        color: ore.color,
        difficulty: ore.difficulty,
        mining_time_ms: ore.mining_time_ms,
        xp_per_ore: ore.xp_per_ore,
        level_required: ore.level_required,
        pickaxe_required: ore.pickaxe_required,
        max_quantity: ore.max_quantity,
        sort_order: ore.sort_order,
      };
    } else {
      editingId = null;
      formData = {
        ore_key: '',
        ore_name: '',
        icon: '',
        color: '',
        difficulty: 'Common',
        mining_time_ms: 3000,
        xp_per_ore: 10,
        level_required: 1,
        pickaxe_required: 'none',
        max_quantity: 0,
        sort_order: 0,
      };
    }
    showModal = true;
  }

  function closeModal() {
    showModal = false;
    editingId = null;
  }

  async function saveOre() {
    errorMsg = '';
    try {
      if (editingId) {
        await adminAPI.updateOre(editingId, formData);
      } else {
        await adminAPI.createOre(formData);
      }
      await loadOres();
      closeModal();
    } catch (e) {
      errorMsg = e.response?.data?.error || 'Failed to save ore';
    }
  }

  async function deleteOre(id) {
    if (!confirm('Delete this ore?')) return;
    errorMsg = '';
    try {
      await adminAPI.deleteOre(id);
      await loadOres();
    } catch (e) {
      errorMsg = e.response?.data?.error || 'Failed to delete ore';
    }
  }
</script>

<div class="management-view">
  <div class="card-header">
    <h2>Ore Management</h2>
    <button class="btn btn-primary" on:click={() => openModal()}>+ Add Ore</button>
  </div>

  {#if errorMsg}
    <div class="error-banner">{errorMsg}</div>
  {/if}

  {#if loading}
    <div class="loading-state">Loading ores…</div>
  {:else if ores.length === 0}
    <div class="empty-state">
      <div class="empty-icon">⛏️</div>
      <p>No ores defined yet.</p>
    </div>
  {:else}
    <div class="table-responsive">
      <table class="equipment-table">
        <thead>
          <tr>
            <th>Icon</th>
            <th>Key</th>
            <th>Name</th>
            <th>Difficulty</th>
            <th>Mining (ms)</th>
            <th>XP/Ore</th>
            <th>Level Req</th>
            <th>Max Qty</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each ores as ore (ore.id)}
            <tr>
              <td class="icon-cell">{ore.icon}</td>
              <td><span class="code">{ore.ore_key}</span></td>
              <td>{ore.ore_name}</td>
              <td>{ore.difficulty}</td>
              <td>{ore.mining_time_ms}</td>
              <td>{ore.xp_per_ore}</td>
              <td>{ore.level_required}</td>
              <td>{ore.max_quantity === 0 ? '∞' : ore.max_quantity}</td>
              <td>
                <div class="actions-cell">
                  <button class="btn btn-sm btn-edit" on:click={() => openModal(ore)}>Edit</button>
                  <button class="btn btn-sm btn-delete" on:click={() => deleteOre(ore.id)}>Delete</button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}

  <!-- Modal -->
  {#if showModal}
    <div class="modal-overlay" role="button" tabindex="0" on:click={closeModal} on:keydown={(e) => e.key === 'Escape' && closeModal()}>
      <div class="modal-content" on:click|stopPropagation={() => {}}>
        <div class="modal-header">
          <h2>{editingId ? 'Edit Ore' : 'Add New Ore'}</h2>
          <button class="modal-close" on:click={closeModal}>✕</button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label for="ore_key">Ore Key *</label>
            <input id="ore_key" type="text" bind:value={formData.ore_key} placeholder="e.g. copper_ore" />
          </div>

          <div class="form-group">
            <label for="ore_name">Name *</label>
            <input id="ore_name" type="text" bind:value={formData.ore_name} placeholder="e.g. Copper Ore" />
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="ore_icon">Icon</label>
              <input id="ore_icon" type="text" bind:value={formData.icon} placeholder="e.g. 🪨" />
            </div>
            <div class="form-group">
              <label for="ore_color">Color</label>
              <input id="ore_color" type="text" bind:value={formData.color} placeholder="e.g. #b87333" />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="ore_difficulty">Difficulty</label>
              <select id="ore_difficulty" bind:value={formData.difficulty}>
                {#each difficultyOptions as diff}
                  <option value={diff}>{diff}</option>
                {/each}
              </select>
            </div>
            <div class="form-group">
              <label for="ore_pickaxe">Pickaxe Required</label>
              <select id="ore_pickaxe" bind:value={formData.pickaxe_required}>
                {#each pickaxeOptions as pickaxe}
                  <option value={pickaxe}>{pickaxe}</option>
                {/each}
              </select>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="ore_mining">Mining Time (ms)</label>
              <input id="ore_mining" type="number" bind:value={formData.mining_time_ms} min="1000" />
            </div>
            <div class="form-group">
              <label for="ore_xp">XP Per Ore</label>
              <input id="ore_xp" type="number" bind:value={formData.xp_per_ore} min="0" />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="ore_level">Level Required</label>
              <input id="ore_level" type="number" bind:value={formData.level_required} min="1" />
            </div>
            <div class="form-group">
              <label for="ore_max">Max Quantity (0 = unlimited)</label>
              <input id="ore_max" type="number" bind:value={formData.max_quantity} min="0" />
            </div>
          </div>

          <div class="form-group">
            <label for="ore_sort">Sort Order</label>
            <input id="ore_sort" type="number" bind:value={formData.sort_order} min="0" />
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn btn-secondary" on:click={closeModal}>Cancel</button>
          <button class="btn btn-primary" on:click={saveOre}>
            {editingId ? 'Update' : 'Create'} Ore
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .management-view {
    padding: 20px 16px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    max-width: 1200px;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .card-header h2 {
    margin: 0;
    font-size: 16px;
    color: var(--color-text-heading);
  }

  .error-banner {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 16px;
    background-color: rgba(204, 74, 0, 0.1);
    border: 1px solid var(--color-danger);
    border-radius: 8px;
    color: var(--color-danger-bright);
    font-size: 13px;
  }

  .btn {
    padding: 10px 16px;
    border-radius: 6px;
    border: 1px solid var(--color-border);
    background-color: var(--color-bg-elevated);
    color: var(--color-text);
    cursor: pointer;
    font-size: 13px;
    font-family: var(--font-body);
    font-weight: 600;
    transition: all 0.15s;
  }

  .btn-primary {
    background-color: var(--color-magic);
    border-color: var(--color-magic);
    color: white;
  }

  .btn-primary:hover {
    background-color: var(--color-magic-bright);
  }

  .btn-secondary {
    border-color: var(--color-border);
  }

  .btn-secondary:hover {
    background-color: var(--color-bg-panel);
  }

  .btn-sm {
    padding: 6px 10px;
    font-size: 12px;
  }

  .btn-edit {
    color: var(--color-magic-bright);
  }

  .btn-delete {
    color: var(--color-danger-bright);
  }

  .loading-state,
  .empty-state {
    padding: 40px 20px;
    text-align: center;
    color: var(--color-text-muted);
  }

  .empty-icon {
    font-size: 48px;
    margin-bottom: 12px;
  }

  .table-responsive {
    overflow-x: auto;
  }

  .equipment-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
  }

  .equipment-table thead {
    background-color: var(--color-bg-elevated);
    border-bottom: 1px solid var(--color-border);
  }

  .equipment-table th {
    padding: 10px 8px;
    text-align: left;
    font-weight: 600;
    color: var(--color-text-heading);
  }

  .equipment-table td {
    padding: 10px 8px;
    border-bottom: 1px solid var(--color-border);
  }

  .equipment-table tbody tr:hover {
    background-color: rgba(80, 80, 200, 0.05);
  }

  .icon-cell {
    font-size: 18px;
  }

  .code {
    font-family: monospace;
    color: var(--color-text-muted);
    font-size: 12px;
  }

  .actions-cell {
    display: flex;
    gap: 4px;
  }

  /* Modal styles */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal-content {
    background-color: var(--color-bg-panel);
    border: 1px solid var(--color-border);
    border-radius: 12px;
    max-width: 600px;
    width: 90%;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
    border-bottom: 1px solid var(--color-border);
  }

  .modal-header h2 {
    margin: 0;
    font-size: 16px;
    color: var(--color-text-heading);
  }

  .modal-close {
    background: none;
    border: none;
    color: var(--color-text-muted);
    font-size: 20px;
    cursor: pointer;
    padding: 0;
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .modal-close:hover {
    color: var(--color-text);
  }

  .modal-body {
    padding: 16px;
    overflow-y: auto;
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .modal-footer {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
    padding: 16px;
    border-top: 1px solid var(--color-border);
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .form-group label {
    font-size: 12px;
    font-weight: 600;
    color: var(--color-text-heading);
  }

  .form-group input,
  .form-group select {
    padding: 8px 10px;
    border: 1px solid var(--color-border);
    border-radius: 6px;
    background-color: var(--color-bg-elevated);
    color: var(--color-text);
    font-family: var(--font-body);
    font-size: 13px;
  }

  .form-group input:focus,
  .form-group select:focus {
    outline: none;
    border-color: var(--color-magic);
    box-shadow: 0 0 0 2px rgba(80, 80, 200, 0.2);
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
  }

  @media (max-width: 600px) {
    .form-row {
      grid-template-columns: 1fr;
    }

    .equipment-table {
      font-size: 12px;
    }

    .equipment-table th,
    .equipment-table td {
      padding: 6px 4px;
    }

    .actions-cell {
      flex-direction: column;
    }
  }
</style>
