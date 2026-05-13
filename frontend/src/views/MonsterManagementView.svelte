<script>
  import { onMount } from 'svelte';
  import { adminAPI } from '../services/admin.js';

  let monsters = [];
  let loading = false;
  let error = null;
  let showModal = false;
  let editingId = null;

  let formData = {
    monster_key: '',
    name: '',
    icon: '',
    description: '',
    hp: 50,
    dex: 1,
    attack_type: 'physical',
    attack_value: 5,
    phys_def: 0,
    resist_fire: 0,
    resist_lightning: 0,
    resist_ice: 0,
    resist_poison: 0,
    resist_chaos: 0,
    money_drop_min: 0,
    money_drop_max: 0,
    xp_drop: 5,
    sort_order: 0,
  };

  const attackTypeOptions = ['physical', 'fire', 'lightning', 'ice', 'poison', 'chaos'];

  onMount(async () => {
    await loadMonsters();
  });

  async function loadMonsters() {
    loading = true;
    error = null;
    try {
      const res = await adminAPI.getAllMonsters();
      monsters = res.data || [];
    } catch (e) {
      error = e?.response?.data?.error || 'Failed to load monsters';
      console.error(error, e);
    } finally {
      loading = false;
    }
  }

  function openModal(item = null) {
    if (item) {
      editingId = item.id;
      formData = {
        monster_key: item.monster_key,
        name: item.name,
        icon: item.icon,
        description: item.description,
        hp: item.hp,
        dex: item.dex,
        attack_type: item.attack_type,
        attack_value: item.attack_value,
        phys_def: item.phys_def,
        resist_fire: item.resist_fire,
        resist_lightning: item.resist_lightning,
        resist_ice: item.resist_ice,
        resist_poison: item.resist_poison,
        resist_chaos: item.resist_chaos,
        money_drop_min: item.money_drop_min,
        money_drop_max: item.money_drop_max,
        xp_drop: item.xp_drop,
        sort_order: item.sort_order,
      };
    } else {
      editingId = null;
      formData = {
        monster_key: '',
        name: '',
        icon: '',
        description: '',
        hp: 50,
        dex: 1,
        attack_type: 'physical',
        attack_value: 5,
        phys_def: 0,
        resist_fire: 0,
        resist_lightning: 0,
        resist_ice: 0,
        resist_poison: 0,
        resist_chaos: 0,
        money_drop_min: 0,
        money_drop_max: 0,
        xp_drop: 5,
        sort_order: 0,
      };
    }
    showModal = true;
  }

  function closeModal() {
    showModal = false;
  }

  async function saveMonster() {
    error = null;
    try {
      if (editingId) {
        await adminAPI.updateMonster(editingId, formData);
      } else {
        await adminAPI.createMonster(formData);
      }
      await loadMonsters();
      closeModal();
    } catch (e) {
      error = e?.response?.data?.error || 'Failed to save monster';
      console.error(error, e);
    }
  }

  async function deleteMonster(id) {
    if (!confirm('Are you sure you want to delete this monster?')) {
      return;
    }
    error = null;
    try {
      await adminAPI.deleteMonster(id);
      await loadMonsters();
    } catch (e) {
      error = e?.response?.data?.error || 'Failed to delete monster';
      console.error(error, e);
    }
  }
</script>

<div class="view-admin">
  <div class="admin-header">
    <h1 class="admin-title">👹 Monster Management</h1>
    <div class="admin-subtitle">Edit monsters in the master table</div>
  </div>

  {#if error}
    <div class="error-banner">
      <span class="error-icon">⚠️</span>
      <span>{error}</span>
    </div>
  {/if}

  <div class="card controls-card">
    <button class="btn btn-primary" on:click={() => openModal()}>
      + Add New Monster
    </button>
  </div>

  {#if loading}
    <div class="loading-state">Loading monsters...</div>
  {:else if monsters.length === 0}
    <div class="empty-state">
      <div class="empty-icon">👹</div>
      <p>No monsters found</p>
    </div>
  {:else}
    <div class="card table-card">
      <div class="table-responsive">
        <table class="monsters-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Icon</th>
              <th>Key</th>
              <th>Name</th>
              <th>HP</th>
              <th>ATK</th>
              <th>Type</th>
              <th>DEF</th>
              <th>XP</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {#each monsters as item (item.id)}
              <tr>
                <td>{item.id}</td>
                <td class="icon-cell">{item.icon}</td>
                <td class="code">{item.monster_key}</td>
                <td><strong>{item.name}</strong></td>
                <td>{item.hp}</td>
                <td>{item.attack_value}</td>
                <td>{item.attack_type}</td>
                <td>{item.phys_def}</td>
                <td>{item.xp_drop}</td>
                <td class="actions-cell">
                  <button class="btn btn-sm btn-edit" on:click={() => openModal(item)}>Edit</button>
                  <button class="btn btn-sm btn-delete" on:click={() => deleteMonster(item.id)}>Delete</button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}

  <!-- Modal -->
  {#if showModal}
    <div class="modal-overlay" role="button" tabindex="0" on:click={closeModal} on:keydown={(e) => e.key === 'Escape' && closeModal()}>
      <div class="modal-content" on:click|stopPropagation={() => {}}>
        <div class="modal-header">
          <h2>{editingId ? 'Edit Monster' : 'Add New Monster'}</h2>
          <button class="modal-close" on:click={closeModal}>✕</button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label for="monster_key">Monster Key *</label>
            <input id="monster_key" type="text" bind:value={formData.monster_key} placeholder="e.g. wasteland_scavenger" />
          </div>

          <div class="form-group">
            <label for="monster_name">Name *</label>
            <input id="monster_name" type="text" bind:value={formData.name} placeholder="e.g. Wasteland Scavenger" />
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="monster_icon">Icon</label>
              <input id="monster_icon" type="text" bind:value={formData.icon} placeholder="e.g. 👹" />
            </div>
            <div class="form-group">
              <label for="monster_attacktype">Attack Type</label>
              <select id="monster_attacktype" bind:value={formData.attack_type}>
                {#each attackTypeOptions as type}
                  <option value={type}>{type}</option>
                {/each}
              </select>
            </div>
          </div>

          <div class="form-group">
            <label for="monster_description">Description</label>
            <textarea id="monster_description" bind:value={formData.description} placeholder="Monster description"></textarea>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="monster_hp">HP</label>
              <input id="monster_hp" type="number" bind:value={formData.hp} min="1" />
            </div>
            <div class="form-group">
              <label for="monster_dex">DEX</label>
              <input id="monster_dex" type="number" bind:value={formData.dex} min="0" />
            </div>
            <div class="form-group">
              <label for="monster_attackvalue">Attack Value</label>
              <input id="monster_attackvalue" type="number" bind:value={formData.attack_value} min="0" />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="monster_physdef">Physical Defence</label>
              <input id="monster_physdef" type="number" bind:value={formData.phys_def} min="0" />
            </div>
            <div class="form-group">
              <label for="monster_xpdrop">XP Drop</label>
              <input id="monster_xpdrop" type="number" bind:value={formData.xp_drop} min="0" />
            </div>
          </div>

          <fieldset class="resistances-group">
            <legend>Resistances (0-100)</legend>
            <div class="form-row">
              <div class="form-group">
                <label for="monster_fire">Fire</label>
                <input id="monster_fire" type="number" bind:value={formData.resist_fire} min="0" max="100" />
              </div>
              <div class="form-group">
                <label for="monster_lightning">Lightning</label>
                <input id="monster_lightning" type="number" bind:value={formData.resist_lightning} min="0" max="100" />
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label for="monster_ice">Ice</label>
                <input id="monster_ice" type="number" bind:value={formData.resist_ice} min="0" max="100" />
              </div>
              <div class="form-group">
                <label for="monster_poison">Poison</label>
                <input id="monster_poison" type="number" bind:value={formData.resist_poison} min="0" max="100" />
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label for="monster_chaos">Chaos</label>
                <input id="monster_chaos" type="number" bind:value={formData.resist_chaos} min="0" max="100" />
              </div>
            </div>
          </fieldset>

          <fieldset class="drops-group">
            <legend>Money Drops</legend>
            <div class="form-row">
              <div class="form-group">
                <label for="monster_moneydropmin">Min</label>
                <input id="monster_moneydropmin" type="number" bind:value={formData.money_drop_min} min="0" />
              </div>
              <div class="form-group">
                <label for="monster_moneydropmax">Max</label>
                <input id="monster_moneydropmax" type="number" bind:value={formData.money_drop_max} min="0" />
              </div>
              <div class="form-group">
                <label for="monster_sort">Sort Order</label>
                <input id="monster_sort" type="number" bind:value={formData.sort_order} min="0" />
              </div>
            </div>
          </fieldset>
        </div>

        <div class="modal-footer">
          <button class="btn btn-secondary" on:click={closeModal}>Cancel</button>
          <button class="btn btn-primary" on:click={saveMonster}>
            {editingId ? 'Update' : 'Create'} Monster
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .view-admin {
    padding: 20px 16px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    max-width: 1200px;
  }

  .admin-header {
    display: flex;
    flex-direction: column;
    gap: 4px;
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

  .error-icon {
    font-size: 16px;
  }

  .card {
    background-color: var(--color-bg-panel);
    border: 1px solid var(--color-border);
    border-radius: 10px;
    padding: 16px;
  }

  .controls-card {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
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

  .monsters-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
  }

  .monsters-table thead {
    background-color: var(--color-bg-elevated);
    border-bottom: 1px solid var(--color-border);
  }

  .monsters-table th {
    padding: 10px 8px;
    text-align: left;
    font-weight: 600;
    color: var(--color-text-heading);
  }

  .monsters-table td {
    padding: 10px 8px;
    border-bottom: 1px solid var(--color-border);
  }

  .monsters-table tbody tr:hover {
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
  .form-group select,
  .form-group textarea {
    padding: 8px 10px;
    border: 1px solid var(--color-border);
    border-radius: 6px;
    background-color: var(--color-bg-elevated);
    color: var(--color-text);
    font-family: var(--font-body);
    font-size: 13px;
  }

  .form-group input:focus,
  .form-group select:focus,
  .form-group textarea:focus {
    outline: none;
    border-color: var(--color-magic);
    box-shadow: 0 0 0 2px rgba(80, 80, 200, 0.2);
  }

  .form-group textarea {
    resize: vertical;
    min-height: 60px;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
  }

  fieldset {
    border: 1px solid var(--color-border);
    border-radius: 6px;
    padding: 12px;
  }

  legend {
    font-size: 12px;
    font-weight: 600;
    color: var(--color-text-heading);
    padding: 0 4px;
  }

  .resistances-group,
  .drops-group {
    margin-top: 12px;
  }

  @media (max-width: 600px) {
    .form-row {
      grid-template-columns: 1fr;
    }

    .monsters-table {
      font-size: 12px;
    }

    .monsters-table th,
    .monsters-table td {
      padding: 6px 4px;
    }

    .actions-cell {
      flex-direction: column;
    }
  }
</style>
