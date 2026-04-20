<script>
  import { onMount } from 'svelte';
  import { sidebarOpen } from './stores/sidebar.js';
  import { currentPage } from './stores/navigation.js';
  import { isAuthenticated, loadAuthFromStorage } from './stores/auth.js';
  import { initMiningStatus } from './stores/mining.js';
  import { loadConfigFromStorage } from './stores/config.js';
  import { theme } from './stores/theme.js';

  import TopBar from './components/TopBar.svelte';
  import Sidebar from './components/Sidebar.svelte';
  import BottomBar from './components/BottomBar.svelte';
  import OfflineGainsPopup from './components/OfflineGainsPopup.svelte';

  import LoginView from './views/LoginView.svelte';
  import HomeView from './views/HomeView.svelte';
  import CombatView from './views/CombatView.svelte';
  import MiningView from './views/MiningView.svelte';
  import SkillsView from './views/SkillsView.svelte';
  import InventoryView from './views/InventoryView.svelte';
  import ShopView from './views/ShopView.svelte';
  import AchievementsView from './views/AchievementsView.svelte';
  import SettingsView from './views/SettingsView.svelte';

  onMount(async function () {
    // Subscribe to theme changes and apply to DOM
    const unsubscribe = theme.subscribe(value => {
      if (typeof document !== 'undefined') {
        document.documentElement.setAttribute('data-theme', value);
      }
    });
    
    // Load config from localStorage on app startup
    loadConfigFromStorage();
    
    // Load auth from localStorage on app startup
    const hasAuth = await loadAuthFromStorage();
    
    if (hasAuth) {
      // Check mining status on app load
      await initMiningStatus();
    }
  });
</script>

{#if $isAuthenticated}
  <TopBar />
  <Sidebar />

  <main class="main-content" class:sidebar-expanded={$sidebarOpen}>
    {#if $currentPage === 'home'}
      <HomeView />
    {:else if $currentPage === 'combat'}
      <CombatView />
    {:else if $currentPage === 'mining'}
      <MiningView />
    {:else if $currentPage === 'skills'}
      <SkillsView />
    {:else if $currentPage === 'inventory'}
      <InventoryView />
    {:else if $currentPage === 'shop'}
      <ShopView />
    {:else if $currentPage === 'achievements'}
      <AchievementsView />
    {:else if $currentPage === 'settings'}
      <SettingsView />
    {/if}
  </main>

  <BottomBar />
  <OfflineGainsPopup />
{:else}
  <LoginView />
{/if}

<style>
  .main-content {
    position: fixed;
    top: var(--topbar-height);
    bottom: var(--bottombar-height);
    left: var(--sidebar-width-collapsed);
    right: 0;
    overflow-y: auto;
    transition: left var(--transition-normal);
  }

  .main-content.sidebar-expanded {
    left: var(--sidebar-width-expanded);
  }

  @media (max-width: 767px) {
    /* On mobile the sidebar is an overlay, so main content ignores it */
    .main-content,
    .main-content.sidebar-expanded {
      left: 0;
    }
  }
</style>

