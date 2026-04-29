package handlers

import (
	"encoding/json"
	"fmt"
	"math"

	"idlegame-backend/database"
)

// ── Types ──────────────────────────────────────────────────────────────────

// PlayerCombatStats holds a player's effective combat stats after applying all equipment bonuses.
type PlayerCombatStats struct {
	HP    int
	MaxHP int
	Dex   int

	AttackValue int
	AttackType  string // physical, fire, lightning, ice, poison, chaos

	PhysDef         int // flat physical damage reduction
	ResistFire      int // % reduction 0–100
	ResistLightning int
	ResistIce       int
	ResistPoison    int
	ResistChaos     int
}

// CombatEvent records one action in the fight log.
type CombatEvent struct {
	Actor   string `json:"actor"`   // player | monster | system
	Damage  int    `json:"damage"`
	Message string `json:"message"`
}

// FightResultResponse is returned from POST /api/map/advance after full fight resolution.
type FightResultResponse struct {
	Outcome        string                 `json:"outcome"`         // player_wins | player_dies
	Log            []CombatEvent          `json:"log"`
	PlayerHPBefore int                    `json:"player_hp_before"`
	PlayerHPAfter  int                    `json:"player_hp_after"`
	PlayerMaxHP    int                    `json:"player_max_hp"`
	XPGained       int64                  `json:"xp_gained"`
	MoneyGained    int64                  `json:"money_gained"`
	Session        *CombatSessionResponse `json:"session"` // nil when player dies and session is cleared
}

// ── GetPlayerCombatStats ───────────────────────────────────────────────────

// GetPlayerCombatStats builds effective combat stats from the player's base stats
// plus all bonuses from currently equipped gear. All calculation is server-side.
func GetPlayerCombatStats(userID uint) PlayerCombatStats {
	var user database.User
	database.DB.First(&user, userID)

	// Defaults: unarmed = Str-based physical attack
	stats := PlayerCombatStats{
		HP:          user.HP,
		MaxHP:       user.MaxHP,
		Dex:         user.Dex,
		AttackValue: user.Str,
		AttackType:  "physical",
	}

	// Load all filled equipment slots
	var slots []database.UserEquippedSlot
	database.DB.Where("user_id = ? AND user_equipment_id != 0", userID).Find(&slots)

	for _, slot := range slots {
		var ue database.UserEquipment
		if err := database.DB.Preload("Equipment").First(&ue, slot.UserEquipmentID).Error; err != nil {
			continue
		}
		e := ue.Equipment

		// Primary stat: weapon sets attack, armour/shield/rings add defence
		if e.Slot == "weapon" {
			stats.AttackValue = e.BaseAttack
			stats.AttackType = e.AttackType
		} else {
			stats.PhysDef += e.BaseDefence
		}

		// Modifier bonuses (dex, elemental resistances)
		var mods []EquipmentModifier
		_ = json.Unmarshal([]byte(e.ModifiersJSON), &mods)
		for _, mod := range mods {
			switch mod.Type {
			case "dex":
				stats.Dex += mod.Value
			case "resist_fire":
				stats.ResistFire += mod.Value
			case "resist_lightning":
				stats.ResistLightning += mod.Value
			case "resist_ice":
				stats.ResistIce += mod.Value
			case "resist_poison":
				stats.ResistPoison += mod.Value
			case "resist_chaos":
				stats.ResistChaos += mod.Value
			// str and int affect base stats — applied to User directly via equipment grants (future)
			}
		}
	}

	return stats
}

// ── Damage calculation ─────────────────────────────────────────────────────

// calcDamage computes one hit's damage.
//   Physical: max(1, attackValue − physDef)
//   Elemental: max(1, attackValue × (1 − resist% / 100))
func calcDamage(attackValue int, attackType string, physDef int, resists map[string]int) int {
	var dmg int
	if attackType == "physical" {
		dmg = attackValue - physDef
	} else {
		resist := clampInt(resists["resist_"+attackType], 0, 100)
		dmg = int(math.Round(float64(attackValue) * (1.0 - float64(resist)/100.0)))
	}
	if dmg < 1 {
		dmg = 1 // always deal at least 1
	}
	return dmg
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ── SimulateFight ──────────────────────────────────────────────────────────

// SimulateFight runs a full player-vs-monster fight and returns the outcome and log.
//
// Turn order rules (all server-side, cannot be spoofed):
//   • Player attacks first when player.Dex >= monster.Dex
//   • Player attacks floor(player.Dex / monster.Dex) times per turn (minimum 1)
//   • Monster always attacks once per turn
//   • 200-round safety cap prevents infinite loops
func SimulateFight(player PlayerCombatStats, monster database.Monster) (outcome string, log []CombatEvent, playerHPAfter int) {
	playerHP := player.HP
	monsterHP := monster.HP

	playerFirst := player.Dex >= monster.DEX
	attacksPerTurn := player.Dex / monster.DEX
	if attacksPerTurn < 1 {
		attacksPerTurn = 1
	}

	playerResists := map[string]int{
		"resist_fire":      player.ResistFire,
		"resist_lightning": player.ResistLightning,
		"resist_ice":       player.ResistIce,
		"resist_poison":    player.ResistPoison,
		"resist_chaos":     player.ResistChaos,
	}
	monsterResists := map[string]int{
		"resist_fire":      monster.ResistFire,
		"resist_lightning": monster.ResistLightning,
		"resist_ice":       monster.ResistIce,
		"resist_poison":    monster.ResistPoison,
		"resist_chaos":     monster.ResistChaos,
	}

	for round := 0; round < 200 && playerHP > 0 && monsterHP > 0; round++ {
		if playerFirst {
			// Player attacks N times
			for i := 0; i < attacksPerTurn && monsterHP > 0; i++ {
				dmg := calcDamage(player.AttackValue, player.AttackType, monster.PhysDef, monsterResists)
				monsterHP -= dmg
				log = append(log, CombatEvent{
					Actor:   "player",
					Damage:  dmg,
					Message: fmt.Sprintf("You strike %s for %d %s damage.", monster.Name, dmg, player.AttackType),
				})
			}
			if monsterHP <= 0 {
				break
			}
			// Monster attacks once
			dmg := calcDamage(monster.AttackValue, monster.AttackType, player.PhysDef, playerResists)
			playerHP -= dmg
			log = append(log, CombatEvent{
				Actor:   "monster",
				Damage:  dmg,
				Message: fmt.Sprintf("%s attacks you for %d %s damage.", monster.Name, dmg, monster.AttackType),
			})
		} else {
			// Monster attacks first
			dmg := calcDamage(monster.AttackValue, monster.AttackType, player.PhysDef, playerResists)
			playerHP -= dmg
			log = append(log, CombatEvent{
				Actor:   "monster",
				Damage:  dmg,
				Message: fmt.Sprintf("%s attacks you for %d %s damage.", monster.Name, dmg, monster.AttackType),
			})
			if playerHP <= 0 {
				break
			}
			// Player attacks N times
			for i := 0; i < attacksPerTurn && monsterHP > 0; i++ {
				dmg := calcDamage(player.AttackValue, player.AttackType, monster.PhysDef, monsterResists)
				monsterHP -= dmg
				log = append(log, CombatEvent{
					Actor:   "player",
					Damage:  dmg,
					Message: fmt.Sprintf("You strike %s for %d %s damage.", monster.Name, dmg, player.AttackType),
				})
			}
		}
	}

	if monsterHP <= 0 {
		outcome = "player_wins"
		playerHPAfter = clampInt(playerHP, 1, player.MaxHP)
		log = append(log, CombatEvent{
			Actor:   "system",
			Message: fmt.Sprintf("☠️ %s is defeated!", monster.Name),
		})
	} else {
		outcome = "player_dies"
		playerHPAfter = 1
		log = append(log, CombatEvent{
			Actor:   "system",
			Message: "💀 You have been defeated!",
		})
	}

	return
}
