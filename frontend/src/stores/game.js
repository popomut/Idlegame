import { writable } from 'svelte/store';
import { characterAPI } from '../services/api.js';

export const player = writable({
  name: 'Operative',
  username: '',
  class: 'Recruit',
  level: 1,
  xp: 0,
  xpRequired: 100,
  hp: 100,
  maxHp: 100,
  stamina: 50,
  maxStamina: 50,
  str: 5,
  int: 5,
  dex: 5,
  money: 0,
});

export async function syncCharacter() {
  try {
    const response = await characterAPI.getCharacter();
    const d = response.data;
    player.set({
      name: d.player_name || d.username,
      username: d.username,
      class: d.player_class,
      level: d.level,
      xp: d.xp,
      xpRequired: d.xp_required,
      hp: d.hp,
      maxHp: d.max_hp,
      stamina: d.stamina,
      maxStamina: d.max_stamina,
      str: d.str,
      int: d.int,
      dex: d.dex,
      money: d.money,
    });
  } catch (err) {
    // Not logged in yet — leave defaults
  }
}

// Ores inventory — dynamic map keyed by ore_key from the DB (e.g. "copper_ore")
// Populated by syncOreInventory() — never hardcode ore keys here
export const ores = writable({});

// Herbs inventory — dynamic map keyed by herb_key from the DB (e.g. "lavender_herb")
// Populated by syncHerbInventory()
export const herbs = writable({});

export const activityLog = writable([
  'TOXIC PROTOCOL initialised.',
  'Awaiting operator authentication.',
]);

export function addLogEntry(message) {
  activityLog.update(function (log) {
    const updated = [message, ...log];
    if (updated.length > 50) {
      return updated.slice(0, 50);
    }
    return updated;
  });
}
