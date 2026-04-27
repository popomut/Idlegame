<script>
  const skillGroups = [
    {
      name: 'Combat Ops',
      icon: '&#x2694;&#xFE0F;',
      skills: [
        { id: 'attack',    name: 'Strike',       icon: '&#x1F5E1;&#xFE0F;', level: 3, xp: 45, maxXp: 100 },
        { id: 'defense',   name: 'Fortification',icon: '&#x1F6E1;&#xFE0F;', level: 2, xp: 20, maxXp: 100 },
        { id: 'strength',  name: 'Endurance',    icon: '&#x1F4AA;',          level: 2, xp: 60, maxXp: 100 },
      ],
    },
    {
      name: 'Chemical Warfare',
      icon: '&#x2622;&#xFE0F;',
      skills: [
        { id: 'fire',      name: 'Incendiary Ops', icon: '&#x1F525;', level: 1, xp: 10, maxXp: 100 },
        { id: 'frost',     name: 'Cryo Agents',    icon: '&#x2744;&#xFE0F;', level: 1, xp: 0,  maxXp: 100 },
        { id: 'lightning', name: 'EMP Tactics',    icon: '&#x26A1;',  level: 1, xp: 5,  maxXp: 100 },
      ],
    },
    {
      name: 'Field Ops',
      icon: '&#x1F333;',
      skills: [
        { id: 'mining',    name: 'Extraction',    icon: '&#x26CF;&#xFE0F;', level: 1, xp: 0, maxXp: 100 },
        { id: 'herbalism', name: 'Bio Sampling',  icon: '&#x1F9EC;',        level: 1, xp: 0, maxXp: 100 },
        { id: 'fishing',   name: 'Water Sourcing',icon: '&#x1F3A3;',        level: 1, xp: 0, maxXp: 100 },
      ],
    },
    {
      name: 'Engineering',
      icon: '&#x1F528;',
      skills: [
        { id: 'smithing',   name: 'Fabrication',         icon: '&#x1F528;', level: 1, xp: 0, maxXp: 100 },
        { id: 'alchemy',    name: 'Chemical Synthesis',  icon: '&#x2697;&#xFE0F;', level: 1, xp: 0, maxXp: 100 },
        { id: 'enchanting', name: 'Weapon Modification', icon: '&#x1FA84;', level: 1, xp: 0, maxXp: 100 },
      ],
    },
  ];

  function getXpPercent(xp, maxXp) {
    return Math.round((xp / maxXp) * 100);
  }
</script>

<div class="view-skills">
  <div class="page-header">
    <h1 class="page-title">&#x1F4AA; Training</h1>
    <p class="page-subtitle">Develop field proficiencies</p>
  </div>

  {#each skillGroups as group}
    <div class="card">
      <div class="card-header">
        <span class="card-icon">{@html group.icon}</span>
        <h2 class="card-title">{group.name}</h2>
      </div>

      <div class="skills-grid">
        {#each group.skills as skill}
          <div class="skill-card">
            <div class="skill-icon">{@html skill.icon}</div>
            <div class="skill-info">
              <div class="skill-name">{skill.name}</div>
              <div class="skill-level">Level {skill.level}</div>
              <div class="xp-bar-track">
                <div
                  class="xp-bar-fill"
                  style="width: {getXpPercent(skill.xp, skill.maxXp)}%"
                ></div>
              </div>
              <div class="xp-text">{skill.xp} / {skill.maxXp} XP</div>
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/each}
</div>

<style>
  .view-skills {
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
    color: var(--color-magic-bright);
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

  .card-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 14px;
  }

  .card-icon {
    font-size: 18px;
  }

  .card-title {
    font-family: var(--font-heading);
    font-size: 15px;
    color: var(--color-text-heading);
    margin: 0;
    font-weight: 600;
  }

  .skills-grid {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .skill-card {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 12px;
    background-color: var(--color-bg-elevated);
    border: 1px solid var(--color-border-subtle);
    border-radius: 8px;
    transition: border-color var(--transition-fast);
  }

  .skill-card:hover {
    border-color: var(--color-border);
  }

  .skill-icon {
    font-size: 24px;
    width: 36px;
    height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .skill-info {
    flex: 1;
    min-width: 0;
  }

  .skill-name {
    font-size: 14px;
    font-weight: 500;
    color: var(--color-text-heading);
    margin-bottom: 2px;
  }

  .skill-level {
    font-size: 12px;
    color: var(--color-magic-bright);
    margin-bottom: 5px;
  }

  .xp-bar-track {
    height: 6px;
    background-color: var(--color-bg-panel);
    border-radius: 3px;
    overflow: hidden;
    margin-bottom: 3px;
  }

  .xp-bar-fill {
    height: 100%;
    background: linear-gradient(90deg, var(--color-gold-dim), var(--color-gold));
    border-radius: 3px;
    transition: width 0.4s ease;
  }

  .xp-text {
    font-size: 11px;
    color: var(--color-text-muted);
  }
</style>
