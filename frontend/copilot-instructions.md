# Realm of Eternity — Copilot Instructions

## Stack
- **Frontend**: Svelte 5 + Vite (SPA, no SSR)
- **Backend** *(future)*: Go + Fiber, WebSocket transport
- **Database** *(future)*: PostgreSQL
- **Cache** *(future)*: Redis

## Dev commands
```bash
npm run dev      # Start Vite dev server at http://localhost:5173
npm run build    # Production build → dist/
npm run preview  # Preview production build locally
```

## Architecture

The app is a fixed-layout SPA with three always-present chrome components and a swappable view area:

```
┌─ TopBar (fixed, z:100) ──────────────────────────┐
│ Sidebar (fixed, z:200) │ main.main-content        │
│                        │  ← current view renders  │
│                        │    here, scrolls alone   │
└─ BottomBar (fixed, z:100) ───────────────────────┘
```

**Navigation is store-driven** — there is no URL router. `src/stores/navigation.js` holds a history stack (`historyStack`). `navigateTo(page)` pushes, `goBack()` pops, `goHome()` resets. `App.svelte` uses `{#if $currentPage === 'x'}` blocks to swap views without any page reload.

**Sidebar has two modes controlled by one store** (`src/stores/sidebar.js → sidebarOpen`):
- Desktop: `false` = icon-only (56 px wide), `true` = expanded (220 px wide). Width transitions via CSS.
- Mobile (`≤767 px`): sidebar is `position: fixed` overlay. `false` = translated off-screen, `true` = slides in. A semi-transparent backdrop renders behind it and closes it on click.

The `main.main-content` element uses `left:` (not `margin-left:`) and transitions in sync with the sidebar width on desktop; on mobile `left: 0` always.

## Key conventions

### Stores (`src/stores/`)
| File | What it holds |
|---|---|
| `navigation.js` | `currentPage`, `canGoBack`, `navigateTo`, `goBack`, `goHome` |
| `sidebar.js` | `sidebarOpen`, `toggleSidebar`, `closeSidebar` |
| `game.js` | `player` (writable), `activityLog`, `addLogEntry` |

All stores use Svelte's `writable` / `derived` (not runes). Prefer this pattern for new stores.

### Components (`src/components/`)
Small, single-responsibility chrome components. They import from stores directly — no props passed from App.svelte. Keep chrome components prop-free where possible.

### Views (`src/views/`)
Each view is a full-page content component rendered inside `main.main-content`. Views get their data from stores, not from parent props. Add a new view by:
1. Creating `src/views/NewView.svelte`
2. Importing it in `App.svelte`
3. Adding an `{:else if $currentPage === 'newpage'}` branch
4. Adding a menu entry object to the `menuItems` array in `Sidebar.svelte`

### Styling
- All design tokens are CSS custom properties in `src/app.css` `:root`. Never hard-code colors/sizes in component `<style>` blocks — always reference a variable.
- Component `<style>` blocks are scoped automatically by Svelte.
- The theme is **always dark** — no light-mode variant.

### Theme palette (quick reference)
| Token | Use |
|---|---|
| `--color-gold` | Primary accent — rewards, headings, active states |
| `--color-magic` / `--color-magic-bright` | Magic, spells, mana |
| `--color-danger` / `--color-danger-bright` | Combat, devil, fire, errors |
| `--color-bg-panel` | Cards and sidebars |
| `--color-bg-elevated` | Items inside cards |
| `--font-heading` | Cinzel serif — section titles only |
| `--font-body` | Inter — all body text |

### Code style
- Do not use tricks to shorten code at the cost of readability.
- Use plain functions (`function foo() {}`) rather than arrow functions for named store updaters.
- HTML entities (e.g. `&#x2694;&#xFE0F;`) are used for icons via `{@html item.icon}` — this is intentional to avoid icon library dependencies.
