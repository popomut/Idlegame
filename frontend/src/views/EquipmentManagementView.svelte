<script>
  import { onMount } from 'svelte';
  import { adminAPI } from '../services/admin.js';

  let equipment = [];
  let loading = false;
  let error = null;
  let showModal = false;
  let editingId = null;

  let formData = {
    equipment_key: '',
    name: '',
    icon: '',
    description: '',
    slot: 'weapon',
    rarity: 'common',
    base_attack: 0,
    attack_type: 'physical',
    base_defence: 0,
    level_required: 1,
    sort_order: 0,
    modifiers: [],
  };

  const slotOptions = ['head', 'chest', 'legs', 'weapon', 'shield', 'ring', 'amulet'];
  const rarityOptions = ['common', 'uncommon', 'rare', 'epic', 'legendary'];
  const attackTypeOptions = ['physical', 'fire', 'lightning', 'ice', 'poison', 'chaos'];

  onMount(async () => {
    await loadEquipment();
  });

  async function loadEquipment() {
    loading = true;
    error = null;
    try {
      const res = await adminAPI.getAllEquipment();
      equipment = res.data || [];
    } catch (e) {
      error = e?.response?.data?.error || 'Failed to load equipment';
      console.error(error, e);
    } finally {
      loading = false;
    }
  }

  function openModal(item = null) {
    if (item) {
      editingId = item.id;
      formData = {
        equipment_key: item.equipment_key,
        name: item.name,
        icon: item.icon,
        description: item.description,
        slot: item.slot,
        rarity: item.rarity,
        base_attack: item.base_attack,
        attack_type: item.attack_type,
        base_defence: item.base_defence,
        level_required: item.level_required,
        sort_order: item.sort_order,
        modifiers: item.modifiers || [],
      };
    } else {
      editingId = null;
      formData = {
        equipment_key: '',
        name: '',
        icon: '',
        description: '',
        slot: 'weapon',
        rarity: 'common',
        base_attack: 0,
        attack_type: 'physical',
        base_defence: 0,
        level_required: 1,
        sort_order: 0,
        modifiers: [],
      };
    }
    showModal = true;
  }

  function closeModal() {
    showModal = false;
  }

  async function saveEquipment() {
    error = null;
    try {
      if (editingId) {
        await adminAPI.updateEquipment(editingId, formData);
      } else {
        await adminAPI.createEquipment(formData);
      }
      await loadEquipment();
      closeModal();
    } catch (e) {
      error = e?.response?.data?.error || 'Failed to save equipment';
      console.error(error, e);
    }
  }

  async function deleteEquipment(id) {
    if (!confirm('Are you sure you want to delete this equipment?')) {
      return;
    }
    error = null;
    try {
      await adminAPI.deleteEquipment(id);
      await loadEquipment();
    } catch (e) {
      error = e?.response?.data?.error || 'Failed to delete equipment';
      console.error(error, e);
    }
  }
</script>

<div class="management-view">
  {#if error}
    <div class="error-banner">
      <span class="error-icon">⚠️</span>
      <span>{error}</span>
    </div>
  {/if}

  <div class="card controls-card">
    <button class="btn btn-primary" on:click={() => openModal()}>
      + Add New Equipment
    </button>
  </div>

  {#if loading}
    <div class="loading-state">Loading equipment...</div>
  {:else if equipment.length === 0}
    <div class="empty-state">
      <div class="empty-icon">📦</div>
      <p>No equipment found</p>
    </div>
  {:else}
    <div class="card table-card">
      <div class="table-responsive">
        <table class="equipment-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Icon</th>
              <th>Key</th>
              <th>Name</th>
              <th>Slot</th>
              <th>Rarity</th>
              <th>Attack</th>
              <th>Defence</th>
              <th>Level</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {#each equipment as item (item.id)}
              <tr>
                <td>{item.id}</td>
                <td class="icon-cell">{item.icon}</td>
                <td class="code">{item.equipment_key}</td>
                <td><strong>{item.name}</strong></td>
                <td>{item.slot}</td>
                <td><span class="rarity-badge {item.rarity}">{item.rarity}</span></td>
                <td>{item.base_attack}</td>
                <td>{item.base_defence}</td>
                <td>{item.level_required}</td>
                <td class="actions-cell">
                  <button class="btn btn-sm btn-edit" on:click={() => openModal(item)}>Edit</button>
                  <button class="btn btn-sm btn-delete" on:click={() => deleteEquipment(item.id)}>Delete</button>
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
          <h2>{editingId ? 'Edit Equipment' : 'Add New Equipment'}</h2>
          <button class="modal-close" on:click={closeModal}>✕</button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label for="equipment_key">Equipment Key *</label>
            <input id="equipment_key" type="text" bind:value={formData.equipment_key} placeholder="e.g. rusty_blade" />
          </div>

          <div class="form-group">
            <label for="equipment_name">Name *</label>
            <input id="equipment_name" type="text" bind:value={formData.name} placeholder="e.g. Rusty Blade" />
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="equipment_icon">Icon</label>
              <input id="equipment_icon" type="text" bind:value={formData.icon} placeholder="e.g. ⚔️" />
            </div>
            <div class="form-group">
              <label for="equipment_slot">Slot *</label>
              <select id="equipment_slot" bind:value={formData.slot}>
                {#each slotOptions as slot}
                  <option value={slot}>{slot}</option>
                {/each}
              </select>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="equipment_rarity">Rarity</label>
              <select id="equipment_rarity" bind:value={formData.rarity}>
                {#each rarityOptions as rarity}
                  <option value={rarity}>{rarity}</option>
                {/each}
              </select>
            </div>
            <div class="form-group">
              <label for="equipment_level">Level Required</label>
              <input id="equipment_level" type="number" bind:value={formData.level_required} min="1" />
            </div>
          </div>

          <div class="form-group">
            <label for="equipment_description">Description</label>
            <textarea id="equipment_description" bind:value={formData.description} placeholder="Item description"></textarea>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="equipment_attack">Base Attack</label>
              <input id="equipment_attack" type="number" bind:value={formData.base_attack} min="0" />
            </div>
            <div class="form-group">
              <label for="equipment_attacktype">Attack Type</label>
              <select id="equipment_attacktype" bind:value={formData.attack_type}>
                {#each attackTypeOptions as type}
                  <option value={type}>{type}</option>
                {/each}
              </select>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="equipment_defence">Base Defence</label>
              <input id="equipment_defence" type="number" bind:value={formData.base_defence} min="0" />
            </div>
            <div class="form-group">
              <label for="equipment_sort">Sort Order</label>
              <input id="equipment_sort" type="number" bind:value={formData.sort_order} min="0" />
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn btn-secondary" on:click={closeModal}>Cancel</button>
          <button class="btn btn-primary" on:click={saveEquipment}>
            {editingId ? 'Update' : 'Create'} Equipment
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

  .rarity-badge {
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
  }

  .rarity-badge.common {
    background-color: rgba(128, 128, 128, 0.2);
    color: #888;
  }

  .rarity-badge.uncommon {
    background-color: rgba(0, 128, 0, 0.2);
    color: #00aa00;
  }

  .rarity-badge.rare {
    background-color: rgba(0, 0, 255, 0.2);
    color: #0066ff;
  }

  .rarity-badge.epic {
    background-color: rgba(128, 0, 128, 0.2);
    color: #9933ff;
  }

  .rarity-badge.legendary {
    background-color: rgba(255, 165, 0, 0.2);
    color: #ff9900;
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
