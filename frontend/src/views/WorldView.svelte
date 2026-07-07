<script>
  import { onMount, onDestroy } from 'svelte';
  import { worldActivities, worldLoading, worldError, worldLastSync, fetchWorldActivity } from '../stores/world.js';
  import { player } from '../stores/game.js';

  const POLL_MS = 7000;
  let pollTimer = null;

  function startPolling() {
    if (pollTimer) return;
    fetchWorldActivity();
    pollTimer = setInterval(fetchWorldActivity, POLL_MS);
  }
  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
  }
  function onVisibilityChange() {
    if (document.hidden) stopPolling();
    else startPolling();
  }

  onMount(() => {
    startPolling();
    document.addEventListener('visibilitychange', onVisibilityChange);
  });
  onDestroy(() => {
    stopPolling();
    document.removeEventListener('visibilitychange', onVisibilityChange);
  });

  // ── Visual mapping per action ──────────────────────────────────────────
  // Each action has a base actor emoji (the person), a tool emoji (the
  // animated bit), and an accent color used for the nameplate.
  const ACTION_VISUALS = {
    mining:    { actor: '⛏️', tool: '🪨', cls: 'act-mining',    label: 'Mining'    },
    gathering: { actor: '🌿', tool: '🌱', cls: 'act-gathering', label: 'Gathering' },
    crafting:  { actor: '🔨', tool: '⚙️', cls: 'act-crafting',  label: 'Crafting'  },
    combat:    { actor: '⚔️', tool: '💥', cls: 'act-combat',    label: 'Fighting'  },
  };

  function visuals(action) {
    return ACTION_VISUALS[action] || ACTION_VISUALS.mining;
  }

  // Deterministic placement: same user_id always lands at roughly the same
  // spot so the page doesn't jitter between polls. Returns a {leftPct, topPct}.
  function spotForUser(userId) {
    const id = Number(userId) || 0;
    // Two cheap mixers — keeps the spread visually irregular.
    const x = ((id * 73 + 17) % 91);   // 0..90
    const y = ((id * 47 + 31) % 70);   // 0..69
    return { leftPct: 4 + x * 0.9, topPct: 8 + y * 0.85 };
  }

  function elapsedLabel(startedAtMs, nowMs) {
    if (!startedAtMs || !nowMs) return '';
    const s = Math.max(0, Math.floor((nowMs - startedAtMs) / 1000));
    if (s < 60) return `${s}s`;
    const m = Math.floor(s / 60);
    if (m < 60) return `${m}m`;
    const h = Math.floor(m / 60);
    return `${h}h`;
  }

  $: myUsername = $player.username;
  $: actorCount = $worldActivities.length;
</script>

<div class="world-page">
  <header class="hdr">
    <h1>Live World</h1>
    <p class="sub">
      {#if $worldError}
        <span class="err">⚠ {$worldError}</span>
      {:else if actorCount === 0}
        The wasteland is quiet right now.
      {:else}
        {actorCount} operative{actorCount === 1 ? '' : 's'} active
      {/if}
    </p>
  </header>

  <div class="scene" class:loading={$worldLoading && actorCount === 0}>
    <!-- Pixel-ish ground: layered radial + linear gradients tuned to the dark
         green theme. No external image needed. -->
    <div class="ground"></div>

    {#each $worldActivities as a (a.user_id)}
      {@const v = visuals(a.action)}
      {@const pos = spotForUser(a.user_id)}
      {@const isMe = a.username === myUsername}
      <div
        class="actor {v.cls}"
        class:me={isMe}
        style="left: {pos.leftPct}%; top: {pos.topPct}%;"
        title="{a.username} — {v.label} {a.target}"
      >
        <div class="sprite">
          <span class="body">{v.actor}</span>
          <span class="tool">{v.tool}</span>
        </div>
        <div class="nameplate">
          <span class="name">{a.username}{isMe ? ' (you)' : ''}</span>
          <span class="target">{a.target}</span>
          <span class="elapsed">{elapsedLabel(a.started_at, $worldLastSync)}</span>
        </div>
      </div>
    {/each}
  </div>

  <footer class="legend">
    {#each Object.values(ACTION_VISUALS) as v}
      <span class="legend-item {v.cls}">
        <span class="legend-dot">{v.actor}</span>
        {v.label}
      </span>
    {/each}
  </footer>
</div>

<style>
  .world-page {
    padding: 16px;
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 12px;
    color: var(--color-text);
  }

  .hdr h1 {
    font-family: var(--font-heading);
    color: var(--color-text-heading);
    margin: 0;
    letter-spacing: 0.06em;
  }
  .hdr .sub {
    margin: 2px 0 0;
    color: var(--color-text-muted);
    font-size: 13px;
  }
  .hdr .err { color: var(--color-danger-bright); }

  .scene {
    position: relative;
    flex: 1;
    min-height: 360px;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-card);
    overflow: hidden;
    image-rendering: pixelated;
  }
  .scene.loading::after {
    content: 'syncing…';
    position: absolute;
    inset: 0;
    display: grid;
    place-items: center;
    color: var(--color-text-muted);
    font-size: 13px;
  }

  /* Pixel-art-feel ground built from layered gradients. */
  .ground {
    position: absolute;
    inset: 0;
    background-color: #0c180c;
    background-image:
      /* lighter speckles */
      radial-gradient(circle at 20% 30%, #1a2e1a 1px, transparent 1.5px),
      radial-gradient(circle at 70% 60%, #1a2e1a 1px, transparent 1.5px),
      radial-gradient(circle at 50% 80%, #15291a 1px, transparent 1.5px),
      /* tile-like grid */
      linear-gradient(to right, rgba(20, 40, 20, 0.5) 1px, transparent 1px),
      linear-gradient(to bottom, rgba(20, 40, 20, 0.5) 1px, transparent 1px);
    background-size: 32px 32px, 28px 28px, 36px 36px, 32px 32px, 32px 32px;
  }

  /* ── Actor ─────────────────────────────────────────────────────────────── */
  .actor {
    position: absolute;
    width: 92px;
    transform: translate(-50%, -50%);
    display: flex;
    flex-direction: column;
    align-items: center;
    pointer-events: auto;
  }
  .actor.me {
    z-index: 2;
    filter: drop-shadow(0 0 6px var(--color-gold-bright));
  }

  .sprite {
    position: relative;
    width: 48px;
    height: 48px;
    display: grid;
    place-items: center;
  }
  .sprite .body {
    font-size: 32px;
    line-height: 1;
    animation: bob 1.6s ease-in-out infinite;
    image-rendering: pixelated;
  }
  .sprite .tool {
    position: absolute;
    right: 0;
    bottom: 4px;
    font-size: 18px;
    line-height: 1;
    transform-origin: 50% 80%;
  }

  /* Per-action tool animation — this is what visually distinguishes
     mining vs crafting vs combat. */
  .act-mining .tool    { animation: swing-down 0.55s ease-in-out infinite; }
  .act-gathering .tool { animation: pluck      1.2s ease-in-out infinite; }
  .act-crafting .tool  { animation: hammer     0.45s ease-in-out infinite; }
  .act-combat .tool    { animation: slash      0.50s ease-in-out infinite; }

  @keyframes bob {
    0%, 100% { transform: translateY(0); }
    50%      { transform: translateY(-3px); }
  }
  @keyframes swing-down {
    0%, 100% { transform: rotate(-20deg) translateY(0); }
    50%      { transform: rotate(40deg) translateY(2px); }
  }
  @keyframes pluck {
    0%, 100% { transform: translateY(0)   rotate(0); }
    50%      { transform: translateY(-3px) rotate(-15deg); }
  }
  @keyframes hammer {
    0%, 100% { transform: rotate(-30deg); }
    50%      { transform: rotate(25deg); }
  }
  @keyframes slash {
    0%, 100% { transform: rotate(-35deg) scale(1); }
    50%      { transform: rotate(45deg) scale(1.15); }
  }

  /* ── Nameplate ─────────────────────────────────────────────────────────── */
  .nameplate {
    margin-top: 2px;
    background: rgba(8, 14, 8, 0.85);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    padding: 2px 6px;
    font-size: 11px;
    line-height: 1.25;
    text-align: center;
    max-width: 110px;
  }
  .actor.me .nameplate { border-color: var(--color-gold); }
  .nameplate .name {
    display: block;
    color: var(--color-text-heading);
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .nameplate .target {
    display: block;
    color: var(--color-text-muted);
    font-size: 10px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .nameplate .elapsed {
    display: block;
    color: var(--color-gold-dim);
    font-size: 10px;
  }

  /* Action-tinted nameplate underline */
  .act-mining    .nameplate { box-shadow: inset 0 -2px 0 var(--color-gold); }
  .act-gathering .nameplate { box-shadow: inset 0 -2px 0 var(--color-magic); }
  .act-crafting  .nameplate { box-shadow: inset 0 -2px 0 var(--color-danger); }
  .act-combat    .nameplate { box-shadow: inset 0 -2px 0 var(--color-danger-bright); }

  /* ── Legend ────────────────────────────────────────────────────────────── */
  .legend {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    padding: 8px 10px;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-card);
    background: var(--color-bg-panel);
    font-size: 12px;
    color: var(--color-text-muted);
  }
  .legend-item { display: inline-flex; align-items: center; gap: 4px; }
  .legend-dot  { font-size: 14px; }
</style>
