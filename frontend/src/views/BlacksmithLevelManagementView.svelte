<script>
  import { onMount } from 'svelte';
  import { adminAPI } from '../services/admin.js';

  let levels = [];
  let loading = false;
  let error = null;
  let showModal = false;
  let editingId = null;

  let formData = {
    level: null,
    xp_required: 0,
  };

  onMount(loadData);

  async function loadData() {
    loading = true;
    try {
      const res = await adminAPI.getAllBlacksmithLevels();
      levels = res.data;
      error = null;
    } catch (e) {
      error = 'Failed to load blacksmith levels';
      console.error(e);
    } finally {
      loading = false;
    }
  }

  function openModal(item = null) {
    if (item) {
      editingId = item.Level;
      formData = {
        level: item.Level,
        xp_required: item.XPRequired,
      };
    } else {
      editingId = null;
      formData = {
        level: null,
        xp_required: 0,
      };
    }
    showModal = true;
  }

  function closeModal() {
    showModal = false;
    editingId = null;
    formData = {
      level: null,
      xp_required: 0,
    };
  }

  async function saveLevel() {
    error = null;
    if (formData.level === null || formData.level < 1) {
      error = 'Level must be >= 1';
      return;
    }

    try {
      if (editingId) {
        await adminAPI.updateBlacksmithLevel(editingId, {
          xp_required: formData.xp_required,
        });
      } else {
        await adminAPI.createBlacksmithLevel(formData);
      }
      await loadData();
      closeModal();
    } catch (e) {
      error = e?.response?.data?.error || 'Failed to save blacksmith level';
      console.error(error, e);
    }
  }

  async function deleteLevel(level) {
    if (!confirm(`Delete level ${level}?`)) return;

    try {
      await adminAPI.deleteBlacksmithLevel(level);
      await loadData();
    } catch (e) {
      error = 'Failed to delete blacksmith level';
      console.error(e);
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
      + Add New Level
    </button>
  </div>

  {#if loading}
    <div class="loading-state">Loading blacksmith levels...</div>
  {:else if levels.length === 0}
    <div class="empty-state">
      <div class="empty-icon">⚒️</div>
      <p>No levels found</p>
    </div>
  {:else}
    <div class="card table-card">
      <div class="table-responsive">
        <table class="levels-table">
          <thead>
            <tr>
              <th>Level</th>
              <th>XP Required</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {#each levels as level (level.Level)}
              <tr>
                <td><strong>{level.Level}</strong></td>
                <td>{level.XPRequired.toLocaleString()}</td>
                <td class="actions-cell">
                  <button class="btn btn-sm btn-edit" on:click={() => openModal(level)}>Edit</button>
                  <button class="btn btn-sm btn-delete" on:click={() => deleteLevel(level.Level)}>Delete</button>
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
          <h2>{editingId ? `Edit Level ${editingId}` : 'Add New Blacksmith Level'}</h2>
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
            <label for="bs_level">Level *</label>
            <input
              id="bs_level"
              type="number"
              bind:value={formData.level}
              disabled={editingId !== null}
              min="1"
              placeholder="Enter level number"
            />
          </div>

          <div class="form-group">
            <label for="bs_xp">XP Required *</label>
            <input
              id="bs_xp"
              type="number"
              bind:value={formData.xp_required}
              min="0"
              placeholder="Enter XP required for this level"
            />
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn btn-secondary" on:click={closeModal}>Cancel</button>
          <button class="btn btn-primary" on:click={saveLevel}>
            {editingId ? 'Update' : 'Create'} Level
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

  .levels-table {
    width: 100%;
    border-collapse: collapse;
    background-color: var(--color-bg-panel);
  }

  .levels-table thead {
    background-color: var(--color-bg-secondary);
    border-bottom: 2px solid var(--color-border);
  }

  .levels-table th {
    padding: 12px 16px;
    text-align: left;
    font-weight: 600;
    color: var(--color-text-heading);
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .levels-table td {
    padding: 12px 16px;
    border-bottom: 1px solid var(--color-border);
    color: var(--color-text);
    font-size: 13px;
  }

  .levels-table tbody tr {
    transition: background-color 0.15s;
  }

  .levels-table tbody tr:hover {
    background-color: var(--color-bg-secondary);
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
    max-width: 500px;
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

  .form-group label {
    font-weight: 500;
    font-size: 13px;
    color: var(--color-text-heading);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .form-group input {
    padding: 10px 12px;
    border: 1px solid var(--color-border);
    border-radius: 6px;
    background-color: var(--color-bg);
    color: var(--color-text);
    font-size: 13px;
    font-family: var(--font-body);
    transition: border-color 0.15s;
  }

  .form-group input:focus {
    outline: none;
    border-color: var(--color-magic);
    box-shadow: 0 0 0 3px rgba(80, 80, 200, 0.1);
  }

  .form-group input:disabled {
    background-color: var(--color-bg-secondary);
    color: var(--color-text-muted);
    cursor: not-allowed;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding: 16px 24px;
    border-top: 1px solid var(--color-border);
  }
</style>
