<script>
  import { onMount } from 'svelte';
  import { sidebarOpen } from './stores/sidebar.js';
  import { currentPage } from './stores/navigation.js';
  import { isAuthenticated, loadAuthFromStorage } from './stores/auth.js';
  import { initMiningStatus, activeMining, tabSwitchGains } from './stores/mining.js';
  import { initCraftingStatus, activeCrafting, tabSwitchCraftingGains } from './stores/blacksmith.js';
  import { loadConfigFromStorage } from './stores/config.js';
  import { theme } from './stores/theme.js';
  import { syncCharacter } from './stores/game.js';
  import { combatState, fetchCombatStatus } from './stores/combat.js';

  import TopBar from './components/TopBar.svelte';
  import Sidebar from './components/Sidebar.svelte';
  import BottomBar from './components/BottomBar.svelte';
  import OfflineGainsPopup from './components/OfflineGainsPopup.svelte';

  import LoginView from './views/LoginView.svelte';
  import HomeView from './views/HomeView.svelte';
  import CombatView from './views/CombatView.svelte';
  import MiningView from './views/MiningView.svelte';
  import BlacksmithView from './views/BlacksmithView.svelte';
  import Crafting2View from './views/Crafting2View.svelte';
  import SkillsView from './views/SkillsView.svelte';
  import InventoryView from './views/InventoryView.svelte';
  import EquipmentView from './views/EquipmentView.svelte';
  import MapView from './views/MapView.svelte';
  import ShopView from './views/ShopView.svelte';
  import AchievementsView from './views/AchievementsView.svelte';
  import SettingsView from './views/SettingsView.svelte';
  import AdminView from './views/AdminView.svelte';

  let _touchStartX = 0;
  let _touchStartY = 0;

  function handleTouchStart(e) {
    _touchStartX = e.touches[0].clientX;
    _touchStartY = e.touches[0].clientY;
  }

  function handleTouchEnd(e) {
    const dx = e.changedTouches[0].clientX - _touchStartX;
    const dy = e.changedTouches[0].clientY - _touchStartY;
    // Only trigger on clearly horizontal swipes (more X than Y, at least 50px)
    if (Math.abs(dx) < 50 || Math.abs(dx) < Math.abs(dy)) return;
    if (dx > 0 && _touchStartX < 50) {
      // Right swipe from left edge → open
      sidebarOpen.set(true);
    } else if (dx < 0) {
      // Left swipe anywhere → close
      sidebarOpen.set(false);
    }
  }

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
      // Load character stats and mining status on app load
      await syncCharacter();
      await initMiningStatus();
      await initCraftingStatus();
    }

    // Global page visibility handler: when user returns from background,
    // sync combat status immediately to catch up with offline progress.
    // Also show tab-switch summary popup for mining and crafting.
    if (typeof document !== 'undefined') {
      let hiddenAt = null;
      const MIN_TAB_SWITCH_MS = 5000; // only show popup if hidden >= 5 seconds

      const handlePageReturn = async () => {
        if (document.hidden) {
          // Tab became hidden — record time
          hiddenAt = Date.now();
        } else {
          // Tab became visible
          if ($combatState.isActive) {
            await fetchCombatStatus();
          }

          if (hiddenAt !== null) {
            const elapsed = Date.now() - hiddenAt;
            hiddenAt = null;

            if (elapsed >= MIN_TAB_SWITCH_MS) {
              // Check mining tab-switch gains
              const mining = $activeMining;
              if (mining && mining.startedAt && mining.extractionTimeMS) {
                const gained = Math.floor(elapsed / mining.extractionTimeMS);
                if (gained > 0) {
                  tabSwitchGains.set({
                    timeMs: elapsed,
                    gained,
                    resourceName: mining.oreName,
                    resourceType: mining.resourceType,
                  });
                }
              }

              // Check crafting tab-switch gains
              const crafting = $activeCrafting;
              if (crafting && crafting.startedAt && crafting.craftingTimeMS) {
                const gained = Math.floor(elapsed / crafting.craftingTimeMS);
                if (gained > 0) {
                  tabSwitchCraftingGains.set({
                    timeMs: elapsed,
                    gained,
                    recipeName: crafting.recipeName,
                  });
                }
              }
            }
          }
        }
      };
      document.addEventListener('visibilitychange', handlePageReturn);

      // Cleanup listener on unmount
      return () => {
        document.removeEventListener('visibilitychange', handlePageReturn);
      };
    }
  });
</script>

<svelte:window on:touchstart={handleTouchStart} on:touchend={handleTouchEnd} />

{#if $isAuthenticated}
  <TopBar />
  <Sidebar />

  <main class="main-content" class:sidebar-expanded={$sidebarOpen}>
    {#if $currentPage === 'home'}
      <HomeView />
    {:else if $currentPage === 'combat'}
      <CombatView />
    {:else if $currentPage === 'map'}
      <MapView />
    {:else if $currentPage === 'mining'}
      <MiningView />
    {:else if $currentPage === 'blacksmith'}
      <BlacksmithView />
    {:else if $currentPage === 'crafting2'}
      <Crafting2View />
    {:else if $currentPage === 'equipment'}
      <EquipmentView />
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
    {:else if $currentPage === 'admin'}
      <AdminView />
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

