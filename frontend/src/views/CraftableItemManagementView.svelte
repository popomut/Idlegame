<script>
  import { onMount } from 'svelte';
  import axios from 'axios';
  import { API_BASE_URL } from '../services/api.js';

  let items = [];
  let showForm = false;
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
    ingredients: []
  };

  let ingredientForm = {
    ingredient_type: 'ore',
    ingredient_key: '',
    quantity_required: 1
  };

  onMount(loadItems);

  async function loadItems() {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/admin/craftable-items`, {
        withCredentials: true,
      });
      items = response.data || [];
    } catch (error) {
      console.error('Failed to load items:', error);
    }
  }

  function openForm(item = null) {
    if (item) {
      editingId = item.id;
      formData = { ...item };
      formData.ingredients = formData.ingredients || [];
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
        ingredients: []
      };
    }
    showForm = true;
  }

  function closeForm() {
    showForm = false;
    editingId = null;
  }

  function addIngredient() {
    formData.ingredients = [
      ...formData.ingredients,
      { ...ingredientForm }
    ];
    ingredientForm = {
      ingredient_type: 'ore',
      ingredient_key: '',
      quantity_required: 1
    };
  }

  function removeIngredient(index) {
    formData.ingredients = formData.ingredients.filter((_, i) => i !== index);
  }

  async function saveItem() {
    if (!formData.name || !formData.item_key) {
      alert('Please fill in all required fields');
      return;
    }

    try {
      if (editingId) {
        await axios.put(
          `${API_BASE_URL}/api/admin/craftable-items/${editingId}`,
          formData,
          { withCredentials: true }
        );
      } else {
        await axios.post(
          `${API_BASE_URL}/api/admin/craftable-items`,
          formData,
          { withCredentials: true }
        );
      }
      await loadItems();
      closeForm();
    } catch (error) {
      console.error('Failed to save item:', error);
      alert('Failed to save item');
    }
  }

  async function deleteItem(id) {
    if (!confirm('Delete this recipe?')) return;

    try {
      await axios.delete(
        `${API_BASE_URL}/api/admin/craftable-items/${id}`,
        { withCredentials: true }
      );
      await loadItems();
    } catch (error) {
      console.error('Failed to delete item:', error);
      alert('Failed to delete item');
    }
  }
</script>

<div class="admin-craftable-items">
  <div class="header">
    <h2>Craftable Items Management</h2>
    <button class="add-btn" on:click={() => openForm()}>+ Add Recipe</button>
  </div>

  {#if !showForm}
    <div class="items-list">
      {#if items.length === 0}
        <div class="empty-state">No recipes yet</div>
      {:else}
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Type</th>
              <th>Level Req</th>
              <th>Crafting Time</th>
              <th>XP</th>
              <th>Ingredients</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {#each items as item (item.id)}
              <tr>
                <td>{item.icon} {item.name}</td>
                <td>{item.output_type}</td>
                <td>{item.level_required}</td>
                <td>{item.crafting_time_ms / 1000}s</td>
                <td>{item.xp_per_craft}</td>
                <td>{item.ingredients?.length || 0} ingredients</td>
                <td>
                  <button class="edit-btn" on:click={() => openForm(item)}>Edit</button>
                  <button class="delete-btn" on:click={() => deleteItem(item.id)}>Delete</button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      {/if}
    </div>
  {:else}
    <div class="form-container">
      <h3>{editingId ? 'Edit Recipe' : 'New Recipe'}</h3>
      
      <div class="form-group">
        <label>Name *</label>
        <input bind:value={formData.name} type="text" />
      </div>

      <div class="form-group">
        <label>Icon</label>
        <input bind:value={formData.icon} type="text" placeholder="e.g., ⚒️" />
      </div>

      <div class="form-group">
        <label>Item Key *</label>
        <input bind:value={formData.item_key} type="text" placeholder="e.g., copper_ingot" />
      </div>

      <div class="form-group">
        <label>Description</label>
        <textarea bind:value={formData.description}></textarea>
      </div>

      <div class="form-row">
        <div class="form-group">
          <label>Output Type</label>
          <select bind:value={formData.output_type}>
            <option value="ingot">Ingot</option>
            <option value="equipment">Equipment</option>
          </select>
        </div>

        <div class="form-group">
          <label>Crafting Time (ms)</label>
          <input bind:value={formData.crafting_time_ms} type="number" />
        </div>

        <div class="form-group">
          <label>XP Per Craft</label>
          <input bind:value={formData.xp_per_craft} type="number" />
        </div>

        <div class="form-group">
          <label>Level Required</label>
          <input bind:value={formData.level_required} type="number" min="1" />
        </div>

        <div class="form-group">
          <label>Max Quantity (0 = unlimited)</label>
          <input bind:value={formData.max_quantity} type="number" min="0" />
        </div>

        <div class="form-group">
          <label>Sort Order</label>
          <input bind:value={formData.sort_order} type="number" />
        </div>
      </div>

      <div class="ingredients-section">
        <h4>Ingredients</h4>
        <div class="ingredients-list">
          {#each formData.ingredients as ing, index (index)}
            <div class="ingredient-row">
              <span>{ing.quantity_required}x {ing.ingredient_key}</span>
              <button class="remove-btn" on:click={() => removeIngredient(index)}>Remove</button>
            </div>
          {/each}
        </div>

        <div class="add-ingredient-form">
          <select bind:value={ingredientForm.ingredient_type}>
            <option value="ore">Ore</option>
            <option value="ingot">Ingot</option>
          </select>
          <input bind:value={ingredientForm.ingredient_key} type="text" placeholder="Ingredient key" />
          <input bind:value={ingredientForm.quantity_required} type="number" min="1" />
          <button class="add-ingredient-btn" on:click={addIngredient}>Add</button>
        </div>
      </div>

      <div class="form-actions">
        <button class="save-btn" on:click={saveItem}>Save</button>
        <button class="cancel-btn" on:click={closeForm}>Cancel</button>
      </div>
    </div>
  {/if}
</div>

<style>
  .admin-craftable-items {
    padding: 20px;
    background: #1a1a2e;
    color: #e0e0e0;
    border-radius: 8px;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  .header h2 {
    margin: 0;
    color: #ffa500;
  }

  .add-btn {
    padding: 10px 20px;
    background: #00cc00;
    color: #000;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-weight: bold;
    transition: all 0.2s ease;
  }

  .add-btn:hover {
    background: #00ff00;
    transform: scale(1.05);
  }

  .items-list {
    background: rgba(20, 20, 40, 0.8);
    border-radius: 8px;
    padding: 15px;
    overflow-x: auto;
  }

  .empty-state {
    text-align: center;
    color: #666;
    padding: 40px;
    font-style: italic;
  }

  table {
    width: 100%;
    border-collapse: collapse;
  }

  thead {
    background: rgba(255, 165, 0, 0.2);
  }

  th {
    padding: 12px;
    text-align: left;
    color: #ffa500;
    font-weight: bold;
    border-bottom: 1px solid rgba(255, 165, 0, 0.3);
  }

  td {
    padding: 12px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  tr:hover {
    background: rgba(255, 165, 0, 0.1);
  }

  .edit-btn,
  .delete-btn {
    padding: 6px 12px;
    margin-right: 8px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.85rem;
    transition: all 0.2s ease;
  }

  .edit-btn {
    background: #0066cc;
    color: white;
  }

  .edit-btn:hover {
    background: #0088ff;
  }

  .delete-btn {
    background: #cc0000;
    color: white;
  }

  .delete-btn:hover {
    background: #ff0000;
  }

  .form-container {
    background: rgba(20, 20, 40, 0.8);
    border-radius: 8px;
    padding: 20px;
  }

  .form-container h3 {
    color: #ffa500;
    margin-top: 0;
  }

  .form-group {
    margin-bottom: 15px;
  }

  .form-row {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 15px;
    margin-bottom: 15px;
  }

  label {
    display: block;
    margin-bottom: 5px;
    color: #ffa500;
    font-weight: bold;
    font-size: 0.9rem;
  }

  input,
  select,
  textarea {
    width: 100%;
    padding: 8px;
    background: rgba(0, 0, 0, 0.5);
    border: 1px solid rgba(255, 165, 0, 0.3);
    border-radius: 4px;
    color: #e0e0e0;
    font-family: inherit;
  }

  input:focus,
  select:focus,
  textarea:focus {
    outline: none;
    border-color: #ffa500;
    box-shadow: 0 0 10px rgba(255, 165, 0, 0.3);
  }

  textarea {
    resize: vertical;
    min-height: 80px;
  }

  .ingredients-section {
    background: rgba(0, 0, 0, 0.3);
    border: 1px solid rgba(255, 165, 0, 0.2);
    border-radius: 6px;
    padding: 15px;
    margin-bottom: 15px;
  }

  .ingredients-section h4 {
    margin-top: 0;
    color: #ffa500;
  }

  .ingredients-list {
    margin-bottom: 15px;
  }

  .ingredient-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px;
    background: rgba(255, 165, 0, 0.1);
    border-radius: 4px;
    margin-bottom: 8px;
  }

  .remove-btn {
    padding: 4px 10px;
    background: #cc0000;
    color: white;
    border: none;
    border-radius: 3px;
    cursor: pointer;
    font-size: 0.8rem;
  }

  .remove-btn:hover {
    background: #ff0000;
  }

  .add-ingredient-form {
    display: grid;
    grid-template-columns: 1fr 1fr 100px 100px;
    gap: 10px;
  }

  .add-ingredient-form select,
  .add-ingredient-form input {
    padding: 8px;
  }

  .add-ingredient-btn {
    padding: 8px;
    background: #0066cc;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-weight: bold;
  }

  .add-ingredient-btn:hover {
    background: #0088ff;
  }

  .form-actions {
    display: flex;
    gap: 10px;
    justify-content: flex-end;
  }

  .save-btn,
  .cancel-btn {
    padding: 10px 20px;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-weight: bold;
    transition: all 0.2s ease;
  }

  .save-btn {
    background: #00cc00;
    color: #000;
  }

  .save-btn:hover {
    background: #00ff00;
  }

  .cancel-btn {
    background: #666;
    color: white;
  }

  .cancel-btn:hover {
    background: #888;
  }
</style>
