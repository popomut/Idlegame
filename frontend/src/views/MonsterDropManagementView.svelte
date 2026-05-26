<script>
  import { onMount } from 'svelte';
  import { adminAPI } from '../services/admin.js';

  let monsterDrops = [];
  let equipment = [];
  let monsters = [];
  let loading = false;
  let error = null;
  let showModal = false;
  let editingId = null;

  let formData = {
    monster_id: null,
    equipment_key: null,
    drop_rate: 0.5,
    drop_min: 1,
    drop_max: 1,
  };

  onMount(async () => {
    await loadData();
  });

  async function loadData() {
    loading = true;
    error = null;
    try {
      const [dropsRes, equipRes, monstersRes] = await Promise.all([
        adminAPI.getAllMonsterDrops(),
        adminAPI.getAllEquipment(),
        adminAPI.getAllMonsters(),
      ]);
      monsterDrops = dropsRes.data || [];
      equipment = equipRes.data || [];
      monsters = monstersRes.data || [];
    } catch (e) {
      error = e?.response?.data?.error || 'Failed to load data';
      console.error(error, e);
    } finally {
      loading = false;
    }
  }

  function openModal(item = null) {
    if (item) {
      editingId = item.id;
      formData = {
        monster_id: item.monster_id,
        equipment_key: item.equipment_key,
        drop_rate: item.drop_rate,
        drop_min: item.drop_min,
        drop_max: item.drop_max,
      };
    } else {
      editingId = null;
      formData = {
        monster_id: null,
        equipment_key: null,
        drop_rate: 0.5,
        drop_min: 1,
        drop_max: 1,
      };
    }
    showModal = true;
  }

  function closeModal() {
    showModal = false;
  }

  async function saveDrop() {
    error = null;
    if (!formData.monster_id || !formData.equipment_key) {
      error = 'Please select both Monster and Equipment';
      return;
    }
    try {
      if (editingId) {
        await adminAPI.updateMonsterDrop(editingId, formData);
      } else {
        await adminAPI.createMonsterDrop(formData);
      }
      await loadData();
      closeModal();
    } catch (e) {
      error = e?.response?.data?.error || 'Failed to save monster drop';
      console.error(error, e);
    }
  }

  async function deleteDrop(id) {
    if (!confirm('Are you sure you want to delete this drop?')) return;
    error = null;
    try {
      await adminAPI.deleteMonsterDrop(id);
      await loadData();
    } catch (e) {
      error = e?.response?.data?.error || 'Failed to delete monster drop';
      console.error(error, e);
    }
  }

  function getMonsterName(monsterId) {
    return monsters.find(m => m.id === monsterId)?.name || 'Unknown';
  }

  function getEquipmentName(equipId) {
    return equipment.find(e => e.id === equipId)?.name || 'Unknown';
  }
</script>

<div class="view-monster-drops">
  <div class="header">
    <h2>Monster Drop Table Management</h2>
    <button class="add-btn" on:click={() => openModal()}>+ Add Drop</button>
  </div>

  {#if error}
    <div class="error-msg">{error}</div>
  {/if}

  {#if loading}
    <p class="loading-msg">Loading...</p>
  {:else}
    <table class="drops-table">
      <thead>
        <tr>
          <th>Monster</th>
          <th>Equipment</th>
          <th>Drop Rate</th>
          <th>Quantity</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each monsterDrops as drop (drop.id)}
          <tr>
            <td>{drop.monster_name}</td>
            <td>{drop.equipment_name}</td>
            <td>{(drop.drop_rate * 100).toFixed(1)}%</td>
            <td>{drop.drop_min}–{drop.drop_max}</td>
            <td class="actions">
              <button class="edit-btn" on:click={() => openModal(drop)}>Edit</button>
              <button class="delete-btn" on:click={() => deleteDrop(drop.id)}>Delete</button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>

    {#if monsterDrops.length === 0}
      <p class="empty-msg">No monster drops configured yet. Click "Add Drop" to get started.</p>
    {/if}
  {/if}

  <!-- Modal -->
  {#if showModal}
    <div class="modal-overlay" on:click={closeModal}>
      <div class="modal-content" on:click|stopPropagation>
        <h3>{editingId ? 'Edit Drop' : 'Add New Drop'}</h3>

        {#if error}
          <div class="error-msg">{error}</div>
        {/if}

        <div class="modal-body">
          <div class="form-group">
            <label for="monster_select">Monster *</label>
            <select id="monster_select" bind:value={formData.monster_id}>
              <option value={null}>Select a monster...</option>
              {#each monsters as m}
                <option value={m.id}>{m.name}</option>
              {/each}
            </select>
          </div>

          <div class="form-group">
            <label for="equipment_select">Equipment *</label>
            <select id="equipment_select" bind:value={formData.equipment_key}>
              <option value={null}>Select equipment...</option>
              {#each equipment as e}
                <option value={e.equipment_key}>{e.name}</option>
              {/each}
            </select>
          </div>

          <div class="form-group">
            <label for="drop_rate">Drop Rate (0–1)</label>
            <input id="drop_rate" type="number" bind:value={formData.drop_rate} min="0" max="1" step="0.1" />
            <small>{(formData.drop_rate * 100).toFixed(1)}% chance</small>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="drop_min">Min Quantity</label>
              <input id="drop_min" type="number" bind:value={formData.drop_min} min="1" />
            </div>
            <div class="form-group">
              <label for="drop_max">Max Quantity</label>
              <input id="drop_max" type="number" bind:value={formData.drop_max} min="1" />
            </div>
          </div>
        </div>

        <div class="modal-actions">
          <button class="save-btn" on:click={saveDrop}>Save</button>
          <button class="cancel-btn" on:click={closeModal}>Cancel</button>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .view-monster-drops {
    padding: 20px;
    max-width: 1200px;
    margin: 0 auto;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  .header h2 {
    margin: 0;
    font-size: 20px;
    color: var(--color-text-heading);
  }

  .add-btn {
    padding: 10px 16px;
    background: var(--color-gold-dim);
    color: #000;
    border: none;
    border-radius: 6px;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s;
  }

  .add-btn:hover {
    background: var(--color-gold-bright);
  }

  .error-msg {
    padding: 12px;
    background: rgba(220, 53, 69, 0.1);
    border: 1px solid #dc3545;
    border-radius: 6px;
    color: #dc3545;
    margin-bottom: 16px;
    font-size: 14px;
  }

  .loading-msg {
    text-align: center;
    color: var(--color-text-muted);
    padding: 40px;
  }

  .drops-table {
    width: 100%;
    border-collapse: collapse;
    background: var(--color-bg-panel);
    border: 1px solid var(--color-border);
    border-radius: 8px;
    overflow: hidden;
  }

  .drops-table thead {
    background: var(--color-bg-elevated);
    border-bottom: 1px solid var(--color-border);
  }

  .drops-table th {
    padding: 12px;
    text-align: left;
    font-weight: 600;
    color: var(--color-text-heading);
    font-size: 13px;
    letter-spacing: 0.5px;
    text-transform: uppercase;
  }

  .drops-table td {
    padding: 12px;
    border-bottom: 1px solid var(--color-border-subtle);
    color: var(--color-text);
    font-size: 14px;
  }

  .drops-table tbody tr:last-child td {
    border-bottom: none;
  }

  .drops-table tbody tr:hover {
    background: var(--color-bg-hover);
  }

  .actions {
    display: flex;
    gap: 6px;
  }

  .edit-btn, .delete-btn {
    padding: 6px 10px;
    border: none;
    border-radius: 4px;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s;
  }

  .edit-btn {
    background: var(--color-magic);
    color: white;
  }

  .edit-btn:hover {
    background: var(--color-magic-bright);
  }

  .delete-btn {
    background: var(--color-danger);
    color: white;
  }

  .delete-btn:hover {
    background: var(--color-danger-bright);
  }

  .empty-msg {
    text-align: center;
    color: var(--color-text-muted);
    padding: 40px;
    font-style: italic;
  }

  /* Modal */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal-content {
    background: var(--color-bg-panel);
    border: 1px solid var(--color-border);
    border-radius: 10px;
    padding: 24px;
    max-width: 500px;
    width: 90%;
    max-height: 80vh;
    overflow-y: auto;
  }

  .modal-content h3 {
    margin: 0 0 16px;
    font-size: 18px;
    color: var(--color-text-heading);
  }

  .modal-body {
    margin-bottom: 20px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
  }

  .form-group label {
    font-size: 13px;
    font-weight: 600;
    color: var(--color-text-heading);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .form-group input,
  .form-group select {
    padding: 10px;
    background: var(--color-bg-elevated);
    border: 1px solid var(--color-border);
    border-radius: 6px;
    color: var(--color-text);
    font-size: 14px;
  }

  .form-group input:focus,
  .form-group select:focus {
    outline: none;
    border-color: var(--color-magic);
    box-shadow: 0 0 0 2px rgba(138, 43, 226, 0.1);
  }

  .form-group small {
    font-size: 12px;
    color: var(--color-text-muted);
  }

  .modal-actions {
    display: flex;
    gap: 10px;
    justify-content: flex-end;
  }

  .save-btn, .cancel-btn {
    padding: 10px 16px;
    border: none;
    border-radius: 6px;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s;
  }

  .save-btn {
    background: var(--color-gold-dim);
    color: #000;
  }

  .save-btn:hover {
    background: var(--color-gold-bright);
  }

  .cancel-btn {
    background: var(--color-bg-elevated);
    color: var(--color-text);
    border: 1px solid var(--color-border);
  }

  .cancel-btn:hover {
    background: var(--color-bg-hover);
  }

  @media (max-width: 600px) {
    .view-admin {
      padding: 12px;
    }

    .header {
      flex-direction: column;
      gap: 12px;
      align-items: flex-start;
    }

    .header button {
      width: 100%;
    }

    .drops-table {
      font-size: 12px;
    }

    .drops-table th,
    .drops-table td {
      padding: 8px 6px;
    }

    .actions {
      display: flex;
      flex-direction: column;
      gap: 4px;
    }

    .action-btn {
      width: 100%;
      padding: 6px;
      font-size: 11px;
    }

    .modal-content {
      width: 95%;
      max-width: none;
      padding: 16px;
      max-height: 90vh;
    }

    .modal-content h3 {
      font-size: 16px;
      margin-bottom: 12px;
    }

    .modal-body {
      gap: 12px;
      margin-bottom: 16px;
    }

    .form-row {
      grid-template-columns: 1fr;
    }

    .form-group input,
    .form-group select {
      padding: 8px;
      font-size: 13px;
    }

    .modal-actions {
      flex-wrap: wrap;
      gap: 6px;
    }

    .save-btn,
    .cancel-btn {
      padding: 8px 12px;
      font-size: 12px;
    }
  }
</style>
