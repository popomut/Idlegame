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
      const res = await adminAPI.getAllMiningLevels();
      levels = res.data;
      error = null;
    } catch (e) {
      error = 'Failed to load mining levels';
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
        await adminAPI.updateMiningLevel(editingId, {
          xp_required: formData.xp_required,
        });
      } else {
        await adminAPI.createMiningLevel(formData);
      }
      await loadData();
      closeModal();
    } catch (e) {
      error = e?.response?.data?.error || 'Failed to save mining level';
      console.error(error, e);
    }
  }

  async function deleteLevel(level) {
    if (!confirm(`Delete level ${level}?`)) return;

    try {
      await adminAPI.deleteMiningLevel(level);
      await loadData();
    } catch (e) {
      error = 'Failed to delete mining level';
      console.error(e);
    }
  }
</script>

<div class="view-admin-mining">
  {#if loading}
    <div class="loading">Loading mining levels...</div>
  {:else}
    <div class="header">
      <h3 class="title">Mining Level Configuration</h3>
      <button class="btn btn-primary" on:click={() => openModal()}>
        ➕ Add Level
      </button>
    </div>

    {#if error}
      <div class="error-message">{error}</div>
    {/if}

    <div class="table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th>Level</th>
            <th>XP Required</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each levels as level}
            <tr>
              <td>{level.Level}</td>
              <td>{level.XPRequired}</td>
              <td>
                <button class="btn btn-small btn-edit" on:click={() => openModal(level)}>
                  Edit
                </button>
                <button class="btn btn-small btn-danger" on:click={() => deleteLevel(level.Level)}>
                  Delete
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    {#if showModal}
      <div class="modal-overlay" on:click={closeModal} role="presentation">
        <div class="modal" on:click|stopPropagation>
          <div class="modal-header">
            <h4>{editingId ? `Edit Level ${editingId}` : 'Add Mining Level'}</h4>
            <button class="modal-close" on:click={closeModal}>✕</button>
          </div>

          <div class="modal-body">
            {#if error}
              <div class="error-message">{error}</div>
            {/if}

            <div class="form-group">
              <label for="level">Level *</label>
              <input
                id="level"
                type="number"
                bind:value={formData.level}
                disabled={editingId !== null}
                min="1"
              />
            </div>

            <div class="form-group">
              <label for="xp">XP Required *</label>
              <input
                id="xp"
                type="number"
                bind:value={formData.xp_required}
                min="0"
              />
            </div>
          </div>

          <div class="modal-footer">
            <button class="btn btn-secondary" on:click={closeModal}>Cancel</button>
            <button class="btn btn-primary" on:click={saveLevel}>Save</button>
          </div>
        </div>
      </div>
    {/if}
  {/if}
</div>

<style>
  .view-admin-mining {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .loading {
    padding: 20px;
    text-align: center;
    color: var(--color-text-muted);
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .title {
    font-size: 16px;
    font-weight: 600;
    color: var(--color-text-heading);
    margin: 0;
  }

  .error-message {
    background: rgba(220, 38, 38, 0.1);
    border-left: 3px solid var(--color-danger-bright);
    padding: 12px;
    border-radius: 4px;
    color: var(--color-danger-bright);
    font-size: 13px;
    margin-bottom: 12px;
  }

  .table-container {
    background: var(--color-bg-panel);
    border: 1px solid var(--color-border);
    border-radius: 8px;
    overflow: hidden;
  }

  .data-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
  }

  .data-table thead {
    background-color: var(--color-bg-elevated);
    border-bottom: 1px solid var(--color-border);
  }

  .data-table th {
    padding: 12px;
    text-align: left;
    font-weight: 600;
    color: var(--color-text-heading);
  }

  .data-table td {
    padding: 12px;
    border-bottom: 1px solid var(--color-border-subtle);
    color: var(--color-text);
  }

  .data-table tbody tr:hover {
    background-color: var(--color-bg-elevated);
  }

  .btn {
    padding: 8px 12px;
    border: 1px solid var(--color-border);
    border-radius: 4px;
    background: var(--color-bg-elevated);
    color: var(--color-text);
    cursor: pointer;
    font-size: 12px;
    font-weight: 500;
    transition: all 0.2s;
  }

  .btn:hover {
    background: var(--color-bg-hover);
  }

  .btn-primary {
    background: var(--color-magic);
    color: #000;
    border-color: var(--color-magic);
  }

  .btn-primary:hover {
    background: var(--color-magic-bright);
  }

  .btn-small {
    padding: 4px 8px;
    font-size: 11px;
    margin-right: 4px;
  }

  .btn-edit {
    color: var(--color-magic);
  }

  .btn-danger {
    color: var(--color-danger-bright);
  }

  .btn-secondary {
    background: var(--color-bg-panel);
  }

  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background: var(--color-bg-panel);
    border: 1px solid var(--color-border);
    border-radius: 8px;
    max-width: 400px;
    width: 90%;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
    border-bottom: 1px solid var(--color-border);
  }

  .modal-header h4 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
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
  }

  .modal-close:hover {
    color: var(--color-text);
  }

  .modal-body {
    padding: 16px;
  }

  .form-group {
    margin-bottom: 16px;
  }

  .form-group label {
    display: block;
    font-size: 13px;
    font-weight: 600;
    color: var(--color-text-heading);
    margin-bottom: 6px;
  }

  .form-group input {
    width: 100%;
    padding: 8px;
    background: var(--color-bg-elevated);
    border: 1px solid var(--color-border);
    border-radius: 4px;
    color: var(--color-text);
    font-family: var(--font-body);
    font-size: 13px;
  }

  .form-group input:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .modal-footer {
    display: flex;
    gap: 8px;
    padding: 16px;
    border-top: 1px solid var(--color-border);
    justify-content: flex-end;
  }
</style>
