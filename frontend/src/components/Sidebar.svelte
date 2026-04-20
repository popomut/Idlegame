<script>
  import { sidebarOpen, closeSidebar } from '../stores/sidebar.js';
  import { currentPage, navigateTo } from '../stores/navigation.js';

  const menuItems = [
    { id: 'home',         label: 'Home',          icon: '&#x1F3E0;' },
    { id: 'combat',       label: 'Combat',         icon: '&#x2694;&#xFE0F;' },
    { id: 'mining',       label: 'Mining',         icon: '&#x26CF;&#xFE0F;' },
    { id: 'skills',       label: 'Skills',          icon: '&#x1F4D6;' },
    { id: 'inventory',    label: 'Inventory',      icon: '&#x1F392;' },
    { id: 'shop',         label: 'Shop',           icon: '&#x1F3EA;' },
    { id: 'achievements', label: 'Achievements',   icon: '&#x1F3C6;' },
    { id: 'settings',     label: 'Settings',       icon: '&#x2699;&#xFE0F;' },
  ];

  function handleMenuClick(itemId) {
    navigateTo(itemId);
    // On mobile close the sidebar after navigating
    if (window.innerWidth < 768) {
      closeSidebar();
    }
  }
</script>

<!-- Mobile semi-transparent backdrop; hidden on desktop via CSS -->
{#if $sidebarOpen}
  <div
    class="sidebar-backdrop"
    on:click={closeSidebar}
    role="presentation"
    aria-hidden="true"
  ></div>
{/if}

<nav class="sidebar" class:expanded={$sidebarOpen} aria-label="Main menu">
  <ul class="menu-list">
    {#each menuItems as item}
      <li>
        <button
          class="menu-item"
          class:active={$currentPage === item.id}
          on:click={() => handleMenuClick(item.id)}
          title={item.label}
        >
          <span class="menu-icon">{@html item.icon}</span>
          <span class="menu-label">{item.label}</span>
        </button>
      </li>
    {/each}
  </ul>
</nav>

<style>
  /* ── Backdrop (only visible on mobile) ────────────────── */
  .sidebar-backdrop {
    display: none;
    position: fixed;
    inset: 0;
    background-color: rgba(0, 0, 0, 0.65);
    z-index: 150;
  }

  /* ── Sidebar shell ─────────────────────────────────────── */
  .sidebar {
    position: fixed;
    top: var(--topbar-height);
    left: 0;
    bottom: var(--bottombar-height);
    width: var(--sidebar-width-collapsed);
    background-color: var(--color-bg-panel);
    border-right: 1px solid var(--color-border);
    overflow-x: hidden;
    overflow-y: auto;
    scrollbar-width: none;
    transition: width var(--transition-normal);
    z-index: 200;
  }

  .sidebar::-webkit-scrollbar {
    display: none;
  }

  .sidebar.expanded {
    width: var(--sidebar-width-expanded);
  }

  /* ── Menu list ─────────────────────────────────────────── */
  .menu-list {
    list-style: none;
    margin: 0;
    padding: 8px 0;
  }

  .menu-item {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 13px 16px;
    background: none;
    border: none;
    border-right: 2px solid transparent;
    color: var(--color-text-muted);
    cursor: pointer;
    text-align: left;
    font-size: 15px;
    font-family: var(--font-body);
    white-space: nowrap;
    transition: background-color var(--transition-fast), color var(--transition-fast), border-color var(--transition-fast);
  }

  .menu-item:hover {
    background-color: var(--color-bg-hover);
    color: var(--color-text);
  }

  .menu-item.active {
    background-color: var(--color-bg-elevated);
    color: var(--color-gold);
    border-right-color: var(--color-gold);
  }

  .menu-icon {
    font-size: 20px;
    min-width: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  /* Label fades in when sidebar is expanded */
  .menu-label {
    opacity: 0;
    transition: opacity var(--transition-normal);
    pointer-events: none;
  }

  .expanded .menu-label {
    opacity: 1;
  }

  /* ── Mobile overrides ──────────────────────────────────── */
  @media (max-width: 767px) {
    .sidebar-backdrop {
      display: block;
    }

    /* On mobile the sidebar starts off-screen and slides in */
    .sidebar {
      width: var(--sidebar-width-expanded);
      transform: translateX(-100%);
      transition: transform var(--transition-normal);
    }

    .sidebar.expanded {
      transform: translateX(0);
      /* Keep width at full expanded on mobile */
      width: var(--sidebar-width-expanded);
    }

    /* Labels always visible on mobile (sidebar is always full-width when open) */
    .menu-label {
      opacity: 1;
    }
  }
</style>
