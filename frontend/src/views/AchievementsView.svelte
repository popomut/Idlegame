<script>
  const achievements = [
    { id: 1, name: 'First Steps',       icon: '&#x1F463;', desc: 'Begin your adventure',          unlocked: true  },
    { id: 2, name: 'Blood and Steel',   icon: '&#x2694;&#xFE0F;', desc: 'Win your first battle',   unlocked: true  },
    { id: 3, name: 'Scholar',           icon: '&#x1F4D6;', desc: 'Learn your first spell',         unlocked: false },
    { id: 4, name: 'Merchant',          icon: '&#x1F3EA;', desc: 'Buy an item from the shop',      unlocked: false },
    { id: 5, name: 'Dragon Slayer',     icon: '&#x1F409;', desc: 'Defeat a dragon',                unlocked: false },
    { id: 6, name: 'Devil\'s Bargain',  icon: '&#x1F608;', desc: 'Survive an encounter with evil', unlocked: false },
    { id: 7, name: 'Knight\'s Honor',   icon: '&#x1F6E1;&#xFE0F;', desc: 'Reach Level 10',         unlocked: false },
    { id: 8, name: 'Archmage',          icon: '&#x2728;',  desc: 'Reach Magic skill level 50',     unlocked: false },
  ];

  $: unlockedCount = achievements.filter(function (a) { return a.unlocked; }).length;
</script>

<div class="view-achievements">
  <div class="page-header">
    <h1 class="page-title">&#x1F3C6; Achievements</h1>
    <p class="page-subtitle">{unlockedCount} / {achievements.length} unlocked</p>
  </div>

  <div class="card">
    <div class="achievements-list">
      {#each achievements as achievement}
        <div class="achievement-row" class:unlocked={achievement.unlocked} class:locked={!achievement.unlocked}>
          <span class="ach-icon">{@html achievement.icon}</span>
          <div class="ach-info">
            <div class="ach-name">{achievement.name}</div>
            <div class="ach-desc">{achievement.desc}</div>
          </div>
          <span class="ach-status">
            {#if achievement.unlocked}
              &#x2713;
            {:else}
              &#x1F512;
            {/if}
          </span>
        </div>
      {/each}
    </div>
  </div>
</div>

<style>
  .view-achievements {
    padding: 20px 16px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    max-width: 760px;
  }

  .page-header {
    padding: 8px 0 4px;
  }

  .page-title {
    font-family: var(--font-heading);
    font-size: 22px;
    color: var(--color-text-heading);
    margin: 0 0 4px;
    font-weight: 600;
  }

  .page-subtitle {
    font-size: 13px;
    color: var(--color-gold);
    margin: 0;
    letter-spacing: 0.5px;
    text-transform: uppercase;
  }

  .card {
    background-color: var(--color-bg-panel);
    border: 1px solid var(--color-border);
    border-radius: 10px;
    padding: 16px;
  }

  .achievements-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .achievement-row {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    border-radius: 8px;
    border: 1px solid var(--color-border-subtle);
    transition: border-color var(--transition-fast);
  }

  .achievement-row.unlocked {
    background-color: rgba(200, 168, 75, 0.06);
    border-color: var(--color-gold-dim);
  }

  .achievement-row.locked {
    background-color: var(--color-bg-elevated);
    opacity: 0.6;
  }

  .ach-icon {
    font-size: 26px;
    width: 36px;
    text-align: center;
    flex-shrink: 0;
  }

  .locked .ach-icon {
    filter: grayscale(1);
  }

  .ach-info {
    flex: 1;
    min-width: 0;
  }

  .ach-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--color-text-heading);
    margin-bottom: 2px;
  }

  .locked .ach-name {
    color: var(--color-text-muted);
  }

  .ach-desc {
    font-size: 12px;
    color: var(--color-text-muted);
  }

  .ach-status {
    font-size: 18px;
    flex-shrink: 0;
  }

  .unlocked .ach-status {
    color: var(--color-gold);
  }

  .locked .ach-status {
    color: var(--color-text-muted);
    font-size: 16px;
  }
</style>
