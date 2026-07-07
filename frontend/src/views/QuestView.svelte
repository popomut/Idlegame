<script>
  import { onMount } from 'svelte';
  import { quests, syncQuests, completeQuest, isLoadingQuests } from '../stores/quest.js';
  import { player } from '../stores/game.js';

  let selectedQuest = null;
  let completing = false;
  let completionResult = null; // { success, data } shown in popup
  let errorMsg = null;

  onMount(function () {
    syncQuests();
  });

  // Group quests by chapter
  function groupByChapter(list) {
    const chapters = {};
    (list || []).forEach(function (q) {
      const ch = q.chapter || 1;
      if (!chapters[ch]) chapters[ch] = [];
      chapters[ch].push(q);
    });
    return Object.entries(chapters).sort(function (a, b) { return a[0] - b[0]; });
  }

  function selectQuest(q) {
    if (q.status === 'locked') return;
    selectedQuest = q;
    errorMsg = null;
  }

  async function handleComplete() {
    if (!selectedQuest || completing) return;
    completing = true;
    errorMsg = null;
    const result = await completeQuest(selectedQuest.quest_key);
    completing = false;
    if (result.ok) {
      completionResult = result.data;
      selectedQuest = null;
    } else {
      errorMsg = result.unmet
        ? result.unmet.join('\n')
        : (result.error || 'Failed to complete quest.');
    }
  }

  function dismissCompletion() {
    completionResult = null;
  }

  function formatReward(r) {
    if (r.reward_type === 'xp') return `+${r.amount} XP`;
    if (r.reward_type === 'money') return `+${r.amount} 💰`;
    if (r.reward_type === 'equipment') return `Equipment: ${r.reward_key.replace(/_/g, ' ')}`;
    return r.reward_type;
  }

  function objectiveLabel(obj) {
    if (obj.display_text) return obj.display_text;
    const count = obj.target_count;
    const key = obj.target_key ? obj.target_key.replace(/_/g, ' ') : '';
    switch (obj.objective_type) {
      case 'kill': return key ? `Kill ${count} ${key}` : `Defeat ${count} enemies`;
      case 'mine': return `Mine ${count} ${key}`;
      case 'gather': return `Gather ${count} ${key}`;
      case 'craft': return `Craft ${count} ${key}`;
      case 'deliver': return `Deliver ${count} ${key}`;
      case 'reach_char_level': return `Reach character level ${count}`;
      case 'reach_mining_level': return `Reach mining level ${count}`;
      case 'reach_blacksmith_level': return `Reach blacksmith level ${count}`;
      default: return obj.objective_type;
    }
  }

  $: chapterGroups = groupByChapter($quests);
</script>

<div class="quest-view">
  <h1 class="view-title">&#x1F4DC; Quests</h1>

  {#if $isLoadingQuests}
    <p class="loading-text">Loading quests…</p>
  {:else if chapterGroups.length === 0}
    <p class="empty-text">No quests available.</p>
  {:else}
    <div class="quest-layout">
      <!-- Quest list panel -->
      <div class="quest-list">
        {#each chapterGroups as [chapter, questList]}
          <div class="chapter-group">
            <h2 class="chapter-heading">Chapter {chapter}</h2>
            {#each questList as q}
              <button
                class="quest-item"
                class:available={q.status === 'available'}
                class:completed={q.status === 'completed'}
                class:locked={q.status === 'locked'}
                class:selected={selectedQuest && selectedQuest.quest_key === q.quest_key}
                on:click={function () { selectQuest(q); }}
                disabled={q.status === 'locked'}
              >
                <span class="quest-status-icon">
                  {#if q.status === 'completed'}✅
                  {:else if q.status === 'available'}🟡
                  {:else}🔒{/if}
                </span>
                <span class="quest-title">{q.title}</span>
              </button>
            {/each}
          </div>
        {/each}
      </div>

      <!-- Quest detail panel -->
      <div class="quest-detail">
        {#if selectedQuest}
          <div class="detail-card">
            <h2 class="detail-title">{selectedQuest.title}</h2>

            {#if selectedQuest.status === 'completed'}
              <p class="completed-badge">✅ Completed</p>
            {/if}

            {#if selectedQuest.intro_text}
              <p class="intro-text">{selectedQuest.intro_text}</p>
            {/if}

            {#if selectedQuest.objectives && selectedQuest.objectives.length > 0}
              <div class="section">
                <h3 class="section-heading">Objectives</h3>
                <ul class="objectives-list">
                  {#each selectedQuest.objectives as obj}
                    <li class="objective-item">
                      <span class="obj-icon">◆</span>
                      {objectiveLabel(obj)}
                    </li>
                  {/each}
                </ul>
              </div>
            {/if}

            {#if selectedQuest.rewards && selectedQuest.rewards.length > 0}
              <div class="section">
                <h3 class="section-heading">Rewards</h3>
                <ul class="rewards-list">
                  {#each selectedQuest.rewards as r}
                    <li class="reward-item">{formatReward(r)}</li>
                  {/each}
                </ul>
              </div>
            {/if}

            {#if errorMsg}
              <div class="error-box">
                <strong>Not ready:</strong>
                <pre>{errorMsg}</pre>
              </div>
            {/if}

            {#if selectedQuest.status === 'available'}
              <button
                class="complete-btn"
                on:click={handleComplete}
                disabled={completing}
              >
                {completing ? 'Checking…' : 'Complete Quest'}
              </button>
            {/if}
          </div>
        {:else}
          <div class="detail-placeholder">
            <p>Select an available quest to view details.</p>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<!-- Completion popup -->
{#if completionResult}
  <div class="overlay" on:click={dismissCompletion}>
    <div class="completion-popup" on:click|stopPropagation>
      <h2 class="popup-title">&#x2728; Quest Complete!</h2>
      <p class="popup-quest-name">{completionResult.quest}</p>

      {#if completionResult.completion_text}
        <p class="popup-flavour">{completionResult.completion_text}</p>
      {/if}

      {#if completionResult.rewards && completionResult.rewards.length > 0}
        <div class="popup-rewards">
          <h3>Rewards</h3>
          <ul>
            {#each completionResult.rewards as r}
              <li>{formatReward(r)}</li>
            {/each}
          </ul>
        </div>
      {/if}

      <button class="dismiss-btn" on:click={dismissCompletion}>Continue</button>
    </div>
  </div>
{/if}

<style>
  .quest-view {
    padding: 1.5rem;
    max-width: 900px;
    margin: 0 auto;
  }

  .view-title {
    font-family: var(--font-heading);
    color: var(--color-gold);
    font-size: 1.6rem;
    margin-bottom: 1.5rem;
  }

  .loading-text,
  .empty-text {
    color: var(--color-text-muted, #888);
    text-align: center;
    margin-top: 3rem;
  }

  .quest-layout {
    display: grid;
    grid-template-columns: 260px 1fr;
    gap: 1.5rem;
    align-items: start;
  }

  @media (max-width: 640px) {
    .quest-layout {
      grid-template-columns: 1fr;
    }
  }

  /* ── Quest list ────────────────── */
  .quest-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .chapter-group {
    background: var(--color-bg-panel);
    border-radius: 8px;
    padding: 0.75rem;
  }

  .chapter-heading {
    font-family: var(--font-heading);
    color: var(--color-gold);
    font-size: 0.85rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    margin: 0 0 0.5rem 0;
    padding-bottom: 0.4rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  .quest-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    width: 100%;
    text-align: left;
    background: var(--color-bg-elevated);
    border: 1px solid transparent;
    border-radius: 6px;
    padding: 0.5rem 0.75rem;
    cursor: pointer;
    transition: border-color 0.15s, background 0.15s;
    margin-bottom: 0.35rem;
    color: var(--color-text, #ccc);
  }

  .quest-item:last-child {
    margin-bottom: 0;
  }

  .quest-item.available {
    border-color: var(--color-gold);
    color: var(--color-gold);
  }

  .quest-item.completed {
    opacity: 0.6;
    color: var(--color-text-muted, #888);
  }

  .quest-item.locked {
    opacity: 0.35;
    cursor: not-allowed;
  }

  .quest-item.selected {
    background: rgba(255, 200, 0, 0.1);
    border-color: var(--color-gold);
  }

  .quest-status-icon {
    font-size: 0.85rem;
    flex-shrink: 0;
  }

  .quest-title {
    font-size: 0.9rem;
  }

  /* ── Detail panel ──────────────── */
  .detail-card {
    background: var(--color-bg-panel);
    border-radius: 8px;
    padding: 1.5rem;
  }

  .detail-placeholder {
    background: var(--color-bg-panel);
    border-radius: 8px;
    padding: 2rem;
    text-align: center;
    color: var(--color-text-muted, #888);
  }

  .detail-title {
    font-family: var(--font-heading);
    color: var(--color-gold);
    font-size: 1.2rem;
    margin: 0 0 0.75rem 0;
  }

  .completed-badge {
    color: #4caf50;
    font-size: 0.9rem;
    margin-bottom: 0.75rem;
  }

  .intro-text {
    color: var(--color-text, #ccc);
    font-style: italic;
    line-height: 1.6;
    margin-bottom: 1.25rem;
    font-size: 0.95rem;
  }

  .section {
    margin-bottom: 1.25rem;
  }

  .section-heading {
    font-family: var(--font-heading);
    color: var(--color-text, #ccc);
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    margin: 0 0 0.5rem 0;
  }

  .objectives-list,
  .rewards-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .objective-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    color: var(--color-text, #ccc);
    font-size: 0.9rem;
  }

  .obj-icon {
    color: var(--color-gold);
    font-size: 0.6rem;
  }

  .reward-item {
    color: var(--color-gold);
    font-size: 0.9rem;
  }

  .error-box {
    background: rgba(255, 60, 60, 0.12);
    border: 1px solid var(--color-danger, #f44);
    border-radius: 6px;
    padding: 0.75rem 1rem;
    margin-bottom: 1rem;
    color: var(--color-danger-bright, #f77);
    font-size: 0.85rem;
  }

  .error-box pre {
    margin: 0.35rem 0 0;
    white-space: pre-wrap;
    font-family: var(--font-body);
  }

  .complete-btn {
    display: block;
    width: 100%;
    padding: 0.7rem 1rem;
    background: var(--color-gold);
    color: #000;
    font-weight: 700;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.95rem;
    transition: opacity 0.15s;
  }

  .complete-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .complete-btn:not(:disabled):hover {
    opacity: 0.85;
  }

  /* ── Completion popup ──────────── */
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 500;
  }

  .completion-popup {
    background: var(--color-bg-panel);
    border: 2px solid var(--color-gold);
    border-radius: 12px;
    padding: 2rem;
    max-width: 420px;
    width: 90%;
    text-align: center;
  }

  .popup-title {
    font-family: var(--font-heading);
    color: var(--color-gold);
    font-size: 1.4rem;
    margin: 0 0 0.5rem 0;
  }

  .popup-quest-name {
    color: var(--color-text, #ccc);
    font-size: 1rem;
    font-weight: 600;
    margin-bottom: 1rem;
  }

  .popup-flavour {
    font-style: italic;
    color: var(--color-text, #ccc);
    font-size: 0.9rem;
    line-height: 1.6;
    margin-bottom: 1.25rem;
  }

  .popup-rewards {
    background: var(--color-bg-elevated);
    border-radius: 6px;
    padding: 0.75rem 1rem;
    margin-bottom: 1.25rem;
    text-align: left;
  }

  .popup-rewards h3 {
    font-family: var(--font-heading);
    color: var(--color-text, #ccc);
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    margin: 0 0 0.5rem 0;
  }

  .popup-rewards ul {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
  }

  .popup-rewards li {
    color: var(--color-gold);
    font-size: 0.9rem;
  }

  .dismiss-btn {
    padding: 0.6rem 2rem;
    background: var(--color-gold);
    color: #000;
    font-weight: 700;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.9rem;
    transition: opacity 0.15s;
  }

  .dismiss-btn:hover {
    opacity: 0.85;
  }
</style>
