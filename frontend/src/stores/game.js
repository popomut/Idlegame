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

// Ores inventory
export const ores = writable({
  copperOre: 5,
  ironOre: 2,
  goldOre: 0,
  mithrilOre: 0,
  diamondOre: 0,
});

export function addOre(oreType) {
  ores.update(function (inv) {
    return { ...inv, [oreType]: (inv[oreType] || 0) + 1 };
  });
}

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
