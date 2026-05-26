<script>
  import { onMount } from 'svelte';
  import { adminAPI } from '../services/admin.js';

  let items = [];
  let loading = false;
  let error = null;
  let showModal = false;
  let editingId = null;

  let formData = {
    name: '',
    description: '',
    icon: '',
    item_key: '',
    output_type: 'ingot',
    crafting_time_ms: 5000,
    xp_per_craft: 25,
    level_required: 1,
    max_quantity: 0,
    sort_order: 0,
  };

  const outputTypeOptions = ['ingot', 'equipment'];

  onMount(loadItems);

  async function loadItems() {
    loading = true;
    error = null;
    try {
      const res = await adminAPI.getAllCraftableItems();
      items = res.data || [];
    } catch (e) {
      error = e?.response?.data?.error || 'Failed to load recipes';
      console.error(error, e);
    } finally {
      loading = false;
    }
  }

  function openModal(item = null) {
    if (item) {
      editingId = item.id;
      formData = {
        name: item.name,
        description: item.description,
        icon: item.icon,
        item_key: item.item_key,
        output_type: item.output_type,
        crafting_time_ms: item.crafting_time_ms,
        xp_per_craft: item.xp_per_craft,
        level_required: item.level_required,
        max_quantity: item.max_quantity,
        sort_order: item.sort_order,
      };
    } else {
      editingId = null;
      formData = {
        name: '',
        description: '',
        icon: '',
        item_key: '',
        output_type: 'ingot',
        crafting_time_ms: 5000,
        xp_per_craft: 25,
        level_required: 1,
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

  async function saveItem() {
    error = null;
    if (!formData.name || !formData.item_key) {
      error = 'Name and Item Key are required';
      return;
    }

    try {
      if (editingId) {
        await adminAPI.updateCraftableItem(editingId, formData);
      } else {
        await adminAPI.createCraftableItem(formData);
      }
      await loadItems();
      closeModal();
    } catch (e) {
      error = e?.response?.data?.error || 'Failed to save recipe';
      console.error(error, e);
    }
  }

  async function deleteItem(id) {
    if (!confirm('Are you sure you want to delete this recipe?')) {
      return;
    }
    error = null;
    try {
      await adminAPI.deleteCraftableItem(id);
      await loadItems();
    } catch (e) {
      error = e?.response?.data?.error || 'Failed to delete recipe';
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
      + Add New Recipe
    </button>
  </div>

  {#if loading}
    <div class="loading-state">Loading recipes...</div>
  {:else if items.length === 0}
    <div class="empty-state">
      <div class="empty-icon">📜</div>
      <p>No recipes found</p>
    </div>
  {:else}
    <div class="card table-card">
      <div class="table-responsive">
        <table class="items-table">
          <thead>
            <tr>
              <th>Icon</th>
              <th>Name</th>
              <th>Key</th>
              <th>Type</th>
              <th>Time</th>
              <th>XP</th>
              <th>Level</th>
              <th>Max</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {#each items as item (item.id)}
              <tr>
                <td class="icon-cell">{item.icon}</td>
                <td><strong>{item.name}</strong></td>
                <td class="code">{item.item_key}</td>
                <td>{item.output_type}</td>
                <td>{item.crafting_time_ms / 1000}s</td>
                <td>{item.xp_per_craft}</td>
                <td>{item.level_required}</td>
                <td>{item.max_quantity === 0 ? '∞' : item.max_quantity}</td>
                <td class="actions-cell">
                  <button class="btn btn-sm btn-edit" on:click={() => openModal(item)}>Edit</button>
                  <button class="btn btn-sm btn-delete" on:click={() => deleteItem(item.id)}>Delete</button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}

  {#if showModal}
    <div class="modal-overlay" role="button" tabindex="0" on:click={closeModal} on:keydown={(e) => e.key === 'Escape' && closeModal()}>
      <div class="modal-content" on:click|stopPropagation={() => {}}>
        <div class="modal-header">
          <h2>{editingId ? 'Edit Recipe' : 'Add New Recipe'}</h2>
          <button class="modal-close" on:click={closeModal}>✕</button>
        </div>

        <div class="modal-body">
          {#if error}
            <div class="error-banner">
              <span class="error-icon">⚠️</span>
              <span>{error}</span>
            </div>
          {/if}

          <div class="form-group">
            <label for="recipe_name">Name *</label>
            <input id="recipe_name" type="text" bind:value={formData.name} placeholder="e.g. Copper Ingot" />
          </div>

          <div class="form-group">
            <label for="recipe_key">Item Key *</label>
            <input id="recipe_key" type="text" bind:value={formData.item_key} placeholder="e.g. copper_ingot" />
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="recipe_icon">Icon</label>
              <input id="recipe_icon" type="text" bind:value={formData.icon} placeholder="e.g. 🟠" />
            </div>
            <div class="form-group">
              <label for="recipe_type">Output Type</label>
              <select id="recipe_type" bind:value={formData.output_type}>
                {#each outputTypeOptions as type}
                  <option value={type}>{type}</option>
                {/each}
              </select>
            </div>
          </div>

          <div class="form-group">
            <label for="recipe_desc">Description</label>
            <textarea id="recipe_desc" bind:value={formData.description} placeholder="Item description"></textarea>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="recipe_time">Crafting Time (ms)</label>
              <input id="recipe_time" type="number" bind:value={formData.crafting_time_ms} min="100" />
            </div>
            <div class="form-group">
              <label for="recipe_xp">XP Per Craft</label>
              <input id="recipe_xp" type="number" bind:value={formData.xp_per_craft} min="0" />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="recipe_level">Level Required</label>
              <input id="recipe_level" type="number" bind:value={formData.level_required} min="1" />
            </div>
            <div class="form-group">
              <label for="recipe_max">Max Quantity (0 = unlimited)</label>
              <input id="recipe_max" type="number" bind:value={formData.max_quantity} min="0" />
            </div>
            <div class="form-group">
              <label for="recipe_sort">Sort Order</label>
              <input id="recipe_sort" type="number" bind:value={formData.sort_order} min="0" />
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn btn-secondary" on:click={closeModal}>Cancel</button>
          <button class="btn btn-primary" on:click={saveItem}>
            {editingId ? 'Update' : 'Create'} Recipe
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
    margin: 0 auto;
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

  .loading-state {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 40px 16px;
    color: var(--color-text-muted);
    font-size: 14px;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px 16px;
    color: var(--color-text-muted);
    gap: 8px;
  }

  .empty-icon {
    font-size: 48px;
    opacity: 0.5;
  }

  .empty-state p {
    margin: 0;
    font-size: 14px;
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
    background-color: var(--color-magic-dark);
    border-color: var(--color-magic-dark);
  }

  .btn-secondary {
    background-color: var(--color-bg-elevated);
    border-color: var(--color-border);
  }

  .btn-secondary:hover {
    background-color: var(--color-bg-secondary);
  }

  .btn-sm {
    padding: 6px 10px;
    font-size: 12px;
  }

  .btn-edit {
    background-color: var(--color-info);
    border-color: var(--color-info);
    color: white;
  }

  .btn-edit:hover {
    background-color: var(--color-info-dark);
    border-color: var(--color-info-dark);
  }

  .btn-delete {
    background-color: var(--color-danger);
    border-color: var(--color-danger);
    color: white;
  }

  .btn-delete:hover {
    background-color: var(--color-danger-dark);
    border-color: var(--color-danger-dark);
  }

  .table-card {
    padding: 0;
    overflow: hidden;
  }

  .table-responsive {
    overflow-x: auto;
  }

  .items-table {
    width: 100%;
    border-collapse: collapse;
    background-color: var(--color-bg-panel);
  }

  .items-table thead {
    background-color: var(--color-bg-secondary);
    border-bottom: 2px solid var(--color-border);
  }

  .items-table th {
    padding: 12px 16px;
    text-align: left;
    font-weight: 600;
    color: var(--color-text-heading);
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .items-table td {
    padding: 12px 16px;
    border-bottom: 1px solid var(--color-border);
    color: var(--color-text);
    font-size: 13px;
  }

  .items-table tbody tr {
    transition: background-color 0.15s;
  }

  .items-table tbody tr:hover {
    background-color: var(--color-bg-secondary);
  }

  .icon-cell {
    font-size: 18px;
    text-align: center;
  }

  .code {
    font-family: monospace;
    font-size: 12px;
    color: var(--color-text-muted);
  }

  .actions-cell {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }

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
    border-radius: 10px;
    width: 90%;
    max-width: 600px;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid var(--color-border);
  }

  .modal-header h2 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
    color: var(--color-text-heading);
  }

  .modal-close {
    background: none;
    border: none;
    color: var(--color-text-muted);
    font-size: 24px;
    cursor: pointer;
    padding: 0;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color 0.15s;
  }

  .modal-close:hover {
    color: var(--color-text);
  }

  .modal-body {
    padding: 20px 24px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .form-row {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 12px;
  }

  .form-group label {
    font-weight: 500;
    font-size: 13px;
    color: var(--color-text-heading);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .form-group input,
  .form-group select,
  .form-group textarea {
    padding: 10px 12px;
    border: 1px solid var(--color-border);
    border-radius: 6px;
    background-color: var(--color-bg);
    color: var(--color-text);
    font-size: 13px;
    font-family: var(--font-body);
    transition: border-color 0.15s;
  }

  .form-group input:focus,
  .form-group select:focus,
  .form-group textarea:focus {
    outline: none;
    border-color: var(--color-magic);
    box-shadow: 0 0 0 3px rgba(80, 80, 200, 0.1);
  }

  .form-group textarea {
    resize: vertical;
    min-height: 80px;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding: 16px 24px;
    border-top: 1px solid var(--color-border);
  }

  @media (max-width: 600px) {
    .management-view {
      padding: 12px;
    }

    .card-header {
      flex-direction: column;
      gap: 12px;
      align-items: flex-start;
    }

    .card-header .btn {
      width: 100%;
    }

    .controls-card {
      flex-direction: column;
    }

    .controls-card .btn {
      width: 100%;
    }

    .items-table {
      font-size: 12px;
    }

    .items-table th,
    .items-table td {
      padding: 8px 12px;
    }

    .actions-cell {
      display: flex;
      flex-direction: column;
      gap: 4px;
    }

    .btn-edit,
    .btn-delete {
      width: 100%;
      justify-content: center;
    }

    .modal-content {
      width: 95%;
      max-width: none;
    }

    .modal-header {
      padding: 16px;
    }

    .modal-header h2 {
      font-size: 16px;
    }

    .modal-body {
      padding: 16px;
      gap: 16px;
    }

    .form-row {
      grid-template-columns: 1fr;
    }

    .form-group input,
    .form-group select,
    .form-group textarea {
      padding: 8px 10px;
      font-size: 13px;
    }

    .modal-footer {
      padding: 12px 16px;
      flex-wrap: wrap;
      gap: 6px;
    }

    .btn {
      padding: 8px 12px;
      font-size: 12px;
    }

    .btn-sm {
      padding: 5px 8px;
      font-size: 11px;
    }
  }
</style>
