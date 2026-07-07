package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"idlegame-backend/database"
)

// WorldActivity is one row in the live-world feed: a player and what they're doing right now.
type WorldActivity struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Action    string `json:"action"` // mining, gathering, crafting, combat
	Target    string `json:"target"` // resource / recipe / monster display name
	Icon      string `json:"icon"`   // emoji for the target (used as a small overlay)
	StartedAt int64  `json:"started_at"` // unix ms — lets the client show "for 30s"
}

type WorldActivityResponse struct {
	Now        int64           `json:"now"`
	Activities []WorldActivity `json:"activities"`
}

// GetWorldActivity returns every currently-active player across the three idle systems:
// mining, crafting, and combat. There is no separate presence table — an "active" session
// row IS the presence signal. When a player stops, their row's status flips and they
// vanish from the feed on the next poll.
func GetWorldActivity(c *fiber.Ctx) error {
	activities := make([]WorldActivity, 0, 32)

	// ── Mining + Herb gathering ──────────────────────────────────────────────
	var mining []database.MiningSession
	database.DB.
		Preload("User").
		Preload("OreType").
		Preload("HerbType").
		Where("status = ?", "active").
		Find(&mining)
	for _, s := range mining {
		action := "mining"
		target := s.OreType.OreName
		icon := s.OreType.Icon
		if s.HerbID != 0 {
			action = "gathering"
			target = s.HerbType.HerbName
			icon = s.HerbType.Icon
		}
		activities = append(activities, WorldActivity{
			UserID:    s.UserID,
			Username:  s.User.Username,
			Action:    action,
			Target:    target,
			Icon:      icon,
			StartedAt: s.StartedAt.UnixMilli(),
		})
	}

	// ── Crafting ─────────────────────────────────────────────────────────────
	var crafting []database.BlacksmithSession
	database.DB.
		Preload("User").
		Preload("CraftableItem").
		Where("status = ?", "active").
		Find(&crafting)
	for _, s := range crafting {
		activities = append(activities, WorldActivity{
			UserID:    s.UserID,
			Username:  s.User.Username,
			Action:    "crafting",
			Target:    s.CraftableItem.Name,
			Icon:      s.CraftableItem.Icon,
			StartedAt: s.StartedAt.UnixMilli(),
		})
	}

	// ── Combat ───────────────────────────────────────────────────────────────
	// ActiveCombat has no GORM relation to User, so we batch-fetch usernames
	// and monsters separately.
	var combat []database.ActiveCombat
	database.DB.Where("status = ?", "active").Find(&combat)

	if len(combat) > 0 {
		userIDs := make([]uint, 0, len(combat))
		monsterKeys := make([]string, 0, len(combat))
		for _, s := range combat {
			userIDs = append(userIDs, s.UserID)
			if s.CurrentEnemyKey != "" {
				monsterKeys = append(monsterKeys, s.CurrentEnemyKey)
			}
		}

		userMap := map[uint]string{}
		var users []database.User
		database.DB.Where("id IN ?", userIDs).Find(&users)
		for _, u := range users {
			userMap[u.ID] = u.Username
		}

		monsterMap := map[string]database.Monster{}
		if len(monsterKeys) > 0 {
			var monsters []database.Monster
			database.DB.Where("monster_key IN ?", monsterKeys).Find(&monsters)
			for _, m := range monsters {
				monsterMap[m.MonsterKey] = m
			}
		}

		for _, s := range combat {
			m := monsterMap[s.CurrentEnemyKey]
			target := m.Name
			if target == "" {
				target = "an enemy"
			}
			activities = append(activities, WorldActivity{
				UserID:    s.UserID,
				Username:  userMap[s.UserID],
				Action:    "combat",
				Target:    target,
				Icon:      m.Icon,
				StartedAt: s.StartedAt.UnixMilli(),
			})
		}
	}

	return c.JSON(WorldActivityResponse{
		Now:        time.Now().UTC().UnixMilli(),
		Activities: activities,
	})
}
