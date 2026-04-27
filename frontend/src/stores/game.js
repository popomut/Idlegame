import { writable } from 'svelte/store';

export const player = writable({
  name: 'Hero',
  level: 1,
  xp: 24,
  xpToNextLevel: 100,
  hp: 85,
  maxHp: 100,
  mana: 42,
  maxMana: 50,
  gold: 150,
  class: 'Apprentice Knight',
});

// Ores inventory — dynamic map keyed by ore_key from the DB (e.g. "copper_ore")
// Populated by syncOreInventory() — never hardcode ore keys here
export const ores = writable({});

export const activityLog = writable([
  'You have entered the Realm of Eternity...',
  'The ancient gates creak open before you.',
  'Your adventure begins.',
]);

export function addLogEntry(message) {
  activityLog.update(function (log) {
    // Keep only the last 50 entries
    const updated = [message, ...log];
    if (updated.length > 50) {
      return updated.slice(0, 50);
    }
    return updated;
  });
}
