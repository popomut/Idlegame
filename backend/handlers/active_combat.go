package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"idlegame-backend/database"
)

// ── Constants ──────────────────────────────────────────────────────────────

const (
	combatRoundMS       = 1000  // 1 second per combat round
	combatMaxRounds     = 14400 // cap offline progress at 4 hours
	combatMaxLogs       = 50    // rolling log window size
	combatOnlineThresh  = 10    // seconds — below this = "online" (detailed logs)
)

// ── Response types ─────────────────────────────────────────────────────────

type CombatLogEntry struct {
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"` // strike, hit, defeat, spawn, death, info
	Message   string `json:"message"`
}

type CombatEnemyState struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	HPCurrent   int    `json:"hp_current"`
	HPMax       int    `json:"hp_max"`
	AttackValue int    `json:"attack_value"`
	AttackType  string `json:"attack_type"`
}

type CombatStatusResponse struct {
	Status           string            `json:"status"`    // active, fled, dead, none
	IsActive         bool              `json:"is_active"`
	ZoneKey          string            `json:"zone_key"`
	PlayerHPCurrent  int               `json:"player_hp_current"`
	PlayerMaxHP      int               `json:"player_max_hp"`
	CurrentEnemy     *CombatEnemyState `json:"current_enemy"`
	EnemiesDefeated  int               `json:"enemies_defeated"`
	TotalXPGained    int64             `json:"total_xp_gained"`
	TotalMoneyGained int64             `json:"total_money_gained"`
	RecentLogs       []CombatLogEntry  `json:"recent_logs"`
	SessionStartedAt int64             `json:"session_started_at"` // unix ms
	WasOffline       bool              `json:"was_offline"`
	OfflineTimeMS    int64             `json:"offline_time_ms"`
	OfflineEnemies   int               `json:"offline_enemies"`
}

// ── Handlers ───────────────────────────────────────────────────────────────

// StartCombat - POST /api/combat/start
// Starts a new combat session or resumes the existing active one.
func StartCombat(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req struct {
		ZoneKey string `json:"zone_key"`
	}
	c.BodyParser(&req)

	var user database.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	// Resume existing active session
	var existing database.ActiveCombat
	if err := database.DB.Where("user_id = ? AND status = ?", userID, "active").First(&existing).Error; err == nil {
		return c.Status(200).JSON(fiber.Map{"session_id": existing.ID, "resumed": true})
	}

	// Clear any old ended session for this user
	database.DB.Where("user_id = ?", userID).Delete(&database.ActiveCombat{})

	// Pick first enemy from zone
	monsters := combatGetZoneMonsters(req.ZoneKey)
	if len(monsters) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "no monsters available"})
	}
	firstEnemy := combatPickRandom(monsters)

	now := time.Now().UTC()
	startLog := CombatLogEntry{
		Timestamp: now.UnixMilli(),
		Type:      "info",
		Message:   fmt.Sprintf("⚔️ Combat started! %s %s appears!", firstEnemy.Icon, firstEnemy.Name),
	}
	logsJSON, _ := json.Marshal([]CombatLogEntry{startLog})

	session := database.ActiveCombat{
		UserID:            userID,
		Status:            "active",
		ZoneKey:           req.ZoneKey,
		CurrentEnemyKey:   firstEnemy.MonsterKey,
		CurrentEnemyHP:    firstEnemy.HP,
		CurrentEnemyMaxHP: firstEnemy.HP,
		PlayerHPCurrent:   user.HP,
		CombatLogsJSON:    string(logsJSON),
		StartedAt:         now,
		LastTickAt:        now,
	}

	if err := database.DB.Create(&session).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to start combat"})
	}

	return c.Status(200).JSON(fiber.Map{"session_id": session.ID, "resumed": false})
}

// GetCombatStatus - GET /api/combat/status
// Runs the combat tick and returns full current state.
func GetCombatStatus(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var user database.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	var session database.ActiveCombat
	if err := database.DB.Where("user_id = ?", userID).First(&session).Error; err != nil {
		return c.Status(200).JSON(CombatStatusResponse{Status: "none", IsActive: false})
	}

	// Load existing logs
	var logs []CombatLogEntry
	json.Unmarshal([]byte(session.CombatLogsJSON), &logs)

	if session.Status != "active" {
		// Ended session — just return stored state
		return c.Status(200).JSON(buildStatusResponse(&session, &user, nil, logs, false, 0, 0))
	}

	// Run tick
	monsters := combatGetZoneMonsters(session.ZoneKey)
	equipment := combatGetEquipment(userID)
	elapsed := time.Since(session.LastTickAt)
	wasOffline := elapsed.Seconds() > float64(combatOnlineThresh)
	enemiesBefore := session.EnemiesDefeated

	newLogs := processCombatRounds(&session, &user, equipment, monsters)

	// Merge logs, keep rolling window
	logs = append(logs, newLogs...)
	if len(logs) > combatMaxLogs {
		logs = logs[len(logs)-combatMaxLogs:]
	}
	logsJSON, _ := json.Marshal(logs)
	session.CombatLogsJSON = string(logsJSON)

	// Award any newly earned XP/money to user
	newXP := session.TotalXPGained - session.XPAwarded
	newMoney := session.TotalMoneyGained - session.MoneyAwarded
	if newXP > 0 || newMoney > 0 {
		database.DB.Model(&user).Updates(map[string]interface{}{
			"xp":    gorm.Expr("xp + ?", newXP),
			"money": gorm.Expr("money + ?", newMoney),
		})
		session.XPAwarded = session.TotalXPGained
		session.MoneyAwarded = session.TotalMoneyGained
	}

	database.DB.Save(&session)

	offlineEnemies := 0
	if wasOffline {
		offlineEnemies = session.EnemiesDefeated - enemiesBefore
	}

	// Build enemy state for response
	var currentEnemy *CombatEnemyState
	for _, m := range monsters {
		if m.MonsterKey == session.CurrentEnemyKey {
			currentEnemy = &CombatEnemyState{
				Key:         m.MonsterKey,
				Name:        m.Name,
				Icon:        m.Icon,
				HPCurrent:   session.CurrentEnemyHP,
				HPMax:       session.CurrentEnemyMaxHP,
				AttackValue: m.AttackValue,
				AttackType:  m.AttackType,
			}
			break
		}
	}

	return c.Status(200).JSON(buildStatusResponse(&session, &user, currentEnemy, logs,
		wasOffline && offlineEnemies > 0,
		elapsed.Milliseconds(),
		offlineEnemies,
	))
}

// FleeCombat - POST /api/combat/flee
// Ends the current combat session, awards any pending rewards.
func FleeCombat(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var user database.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	var session database.ActiveCombat
	if err := database.DB.Where("user_id = ? AND status = ?", userID, "active").First(&session).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "no active combat session"})
	}

	// Final tick before flee
	monsters := combatGetZoneMonsters(session.ZoneKey)
	equipment := combatGetEquipment(userID)
	newLogs := processCombatRounds(&session, &user, equipment, monsters)

	// Add flee log
	var logs []CombatLogEntry
	json.Unmarshal([]byte(session.CombatLogsJSON), &logs)
	logs = append(logs, newLogs...)
	logs = append(logs, CombatLogEntry{
		Timestamp: time.Now().UnixMilli(),
		Type:      "info",
		Message:   "🏃 You fled from combat!",
	})
	if len(logs) > combatMaxLogs {
		logs = logs[len(logs)-combatMaxLogs:]
	}
	logsJSON, _ := json.Marshal(logs)
	session.CombatLogsJSON = string(logsJSON)
	session.Status = "fled"

	// Award all pending rewards
	newXP := session.TotalXPGained - session.XPAwarded
	newMoney := session.TotalMoneyGained - session.MoneyAwarded
	if newXP > 0 || newMoney > 0 {
		database.DB.Model(&user).Updates(map[string]interface{}{
			"xp":    gorm.Expr("xp + ?", newXP),
			"money": gorm.Expr("money + ?", newMoney),
		})
		session.XPAwarded = session.TotalXPGained
		session.MoneyAwarded = session.TotalMoneyGained
	}

	database.DB.Save(&session)

	return c.Status(200).JSON(fiber.Map{
		"reason":           "fled",
		"enemies_defeated": session.EnemiesDefeated,
		"total_xp":         session.TotalXPGained,
		"total_money":      session.TotalMoneyGained,
	})
}

// ── Core tick algorithm ────────────────────────────────────────────────────

// processCombatRounds calculates all combat rounds since LastTickAt.
// Uses batch O(enemy_deaths) algorithm — efficient for long offline periods.
func processCombatRounds(session *database.ActiveCombat, user *database.User, equipment []database.UserEquipment, monsters []database.Monster) []CombatLogEntry {
	now := time.Now().UTC()
	elapsed := now.Sub(session.LastTickAt)
	rounds := int64(elapsed.Milliseconds() / combatRoundMS)
	if rounds > combatMaxRounds {
		rounds = combatMaxRounds
	}
	if rounds <= 0 {
		session.LastTickAt = now
		return nil
	}

	isOnline := elapsed.Seconds() <= float64(combatOnlineThresh)
	playerDmg := combatPlayerDamage(user, equipment)
	if playerDmg < 1 {
		playerDmg = 1
	}

	// Find current enemy
	currentEnemy := combatFindMonster(monsters, session.CurrentEnemyKey)
	if currentEnemy == nil {
		if len(monsters) > 0 {
			currentEnemy = combatPickRandom(monsters)
			session.CurrentEnemyKey = currentEnemy.MonsterKey
			session.CurrentEnemyHP = currentEnemy.HP
			session.CurrentEnemyMaxHP = currentEnemy.HP
		} else {
			session.LastTickAt = now
			return nil
		}
	}

	var logs []CombatLogEntry

	for rounds > 0 && session.Status == "active" {
		enemyDmg := combatEnemyDamage(currentEnemy, user, equipment)
		if enemyDmg < 1 {
			enemyDmg = 1
		}

		// Rounds needed to kill enemy (player attacks first each round)
		roundsToKillEnemy := combatCeilDiv(int64(session.CurrentEnemyHP), int64(playerDmg))
		// Rounds needed for enemy to kill player
		roundsToKillPlayer := combatCeilDiv(int64(session.PlayerHPCurrent), int64(enemyDmg))

		if roundsToKillEnemy <= roundsToKillPlayer {
			// Player kills enemy first (ties go to player — they attack first)
			if roundsToKillEnemy <= rounds {
				// Enemy dies within budget
				rounds -= roundsToKillEnemy
				// Player takes enemy hits for N-1 rounds (enemy can't counter on dying round)
				if roundsToKillEnemy > 1 {
					session.PlayerHPCurrent -= enemyDmg * int(roundsToKillEnemy-1)
					if session.PlayerHPCurrent < 1 {
						session.PlayerHPCurrent = 1 // survived by a hair
					}
				}
				session.CurrentEnemyHP = 0
				session.EnemiesDefeated++
				xp := int64(currentEnemy.XPDrop)
				money := int64(combatRandRange(currentEnemy.MoneyDropMin, currentEnemy.MoneyDropMax))
				session.TotalXPGained += xp
				session.TotalMoneyGained += money

				if len(logs) < combatMaxLogs {
					logs = append(logs, CombatLogEntry{
						Timestamp: now.UnixMilli(),
						Type:      "defeat",
						Message:   fmt.Sprintf("☠️ %s %s defeated! +%d XP, +%d 💰", currentEnemy.Icon, currentEnemy.Name, xp, money),
					})
				}

				// Spawn next enemy
				currentEnemy = combatPickRandom(monsters)
				session.CurrentEnemyKey = currentEnemy.MonsterKey
				session.CurrentEnemyHP = currentEnemy.HP
				session.CurrentEnemyMaxHP = currentEnemy.HP

				if len(logs) < combatMaxLogs {
					logs = append(logs, CombatLogEntry{
						Timestamp: now.UnixMilli(),
						Type:      "spawn",
						Message:   fmt.Sprintf("👹 %s %s appears! (HP: %d)", currentEnemy.Icon, currentEnemy.Name, currentEnemy.HP),
					})
				}
			} else {
				// Not enough rounds to finish enemy — apply partial damage
				session.CurrentEnemyHP -= playerDmg * int(rounds)
				session.PlayerHPCurrent -= enemyDmg * int(rounds)
				if session.CurrentEnemyHP < 0 {
					session.CurrentEnemyHP = 0
				}
				if session.PlayerHPCurrent < 0 {
					session.PlayerHPCurrent = 0
				}
				if isOnline && len(logs) < combatMaxLogs {
					logs = append(logs, CombatLogEntry{
						Timestamp: now.UnixMilli(),
						Type:      "strike",
						Message:   fmt.Sprintf("⚔️ You hit %s %s for %d! (HP: %d/%d)", currentEnemy.Icon, currentEnemy.Name, playerDmg*int(rounds), session.CurrentEnemyHP, session.CurrentEnemyMaxHP),
					})
					logs = append(logs, CombatLogEntry{
						Timestamp: now.UnixMilli(),
						Type:      "hit",
						Message:   fmt.Sprintf("💢 %s %s hits you for %d! Your HP: %d/%d", currentEnemy.Icon, currentEnemy.Name, enemyDmg*int(rounds), session.PlayerHPCurrent, user.MaxHP),
					})
				}
				rounds = 0
			}
		} else {
			// Enemy kills player first
			if roundsToKillPlayer <= rounds {
				// Player dies within budget
				session.CurrentEnemyHP -= playerDmg * int(roundsToKillPlayer)
				if session.CurrentEnemyHP < 0 {
					session.CurrentEnemyHP = 0
				}
				session.PlayerHPCurrent = 0
				session.Status = "dead"

				if len(logs) < combatMaxLogs {
					logs = append(logs, CombatLogEntry{
						Timestamp: now.UnixMilli(),
						Type:      "death",
						Message:   fmt.Sprintf("💀 You were defeated by %s %s!", currentEnemy.Icon, currentEnemy.Name),
					})
				}
				rounds = 0
			} else {
				// Partial fight — no death yet
				session.CurrentEnemyHP -= playerDmg * int(rounds)
				session.PlayerHPCurrent -= enemyDmg * int(rounds)
				if session.CurrentEnemyHP < 0 {
					session.CurrentEnemyHP = 0
				}
				if session.PlayerHPCurrent < 1 {
					session.PlayerHPCurrent = 1
				}
				if isOnline && len(logs) < combatMaxLogs {
					logs = append(logs, CombatLogEntry{
						Timestamp: now.UnixMilli(),
						Type:      "strike",
						Message:   fmt.Sprintf("⚔️ You hit %s %s for %d! (HP: %d/%d)", currentEnemy.Icon, currentEnemy.Name, playerDmg*int(rounds), session.CurrentEnemyHP, session.CurrentEnemyMaxHP),
					})
					logs = append(logs, CombatLogEntry{
						Timestamp: now.UnixMilli(),
						Type:      "hit",
						Message:   fmt.Sprintf("💢 %s %s hits you for %d! Your HP: %d/%d", currentEnemy.Icon, currentEnemy.Name, enemyDmg*int(rounds), session.PlayerHPCurrent, user.MaxHP),
					})
				}
				rounds = 0
			}
		}
	}

	session.LastTickAt = now
	return logs
}

// ── Helpers ────────────────────────────────────────────────────────────────

func combatPlayerDamage(user *database.User, equipment []database.UserEquipment) int {
	base := user.Str * 2
	for _, ue := range equipment {
		if ue.Equipment.Slot == "weapon" {
			base += ue.Equipment.BaseAttack
		}
	}
	if base < 1 {
		return 1
	}
	return base
}

func combatEnemyDamage(enemy *database.Monster, user *database.User, equipment []database.UserEquipment) int {
	defense := 0
	for _, ue := range equipment {
		slot := ue.Equipment.Slot
		if slot == "head" || slot == "chest" || slot == "legs" || slot == "shield" {
			defense += ue.Equipment.BaseDefence
		}
	}
	net := enemy.AttackValue - defense
	if net < 1 {
		return 1
	}
	return net
}

func combatGetZoneMonsters(zoneKey string) []database.Monster {
	if zoneKey != "" {
		var area database.Area
		if err := database.DB.Where("area_key = ?", zoneKey).First(&area).Error; err == nil {
			var areaMonsters []database.AreaMonster
			database.DB.Where("area_id = ?", area.ID).Find(&areaMonsters)
			if len(areaMonsters) > 0 {
				var monsters []database.Monster
				for _, am := range areaMonsters {
					var m database.Monster
					if database.DB.Where("monster_key = ?", am.MonsterKey).First(&m).Error == nil {
						// Add the monster Weight times for weighted selection
						weight := am.Weight
						if weight < 1 {
							weight = 1
						}
						for w := 0; w < weight; w++ {
							monsters = append(monsters, m)
						}
					}
				}
				if len(monsters) > 0 {
					return monsters
				}
			}
		}
	}
	// Fallback: all monsters
	var monsters []database.Monster
	database.DB.Find(&monsters)
	return monsters
}

func combatGetEquipment(userID uint) []database.UserEquipment {
	var slots []database.UserEquippedSlot
	database.DB.Where("user_id = ? AND user_equipment_id > 0", userID).Find(&slots)
	var equipment []database.UserEquipment
	for _, slot := range slots {
		var ue database.UserEquipment
		if database.DB.Preload("Equipment").First(&ue, slot.UserEquipmentID).Error == nil {
			equipment = append(equipment, ue)
		}
	}
	return equipment
}

func combatFindMonster(monsters []database.Monster, key string) *database.Monster {
	for i := range monsters {
		if monsters[i].MonsterKey == key {
			return &monsters[i]
		}
	}
	return nil
}

func combatPickRandom(monsters []database.Monster) *database.Monster {
	if len(monsters) == 0 {
		return nil
	}
	return &monsters[rand.Intn(len(monsters))]
}

func combatRandRange(min, max int) int {
	if min >= max {
		return min
	}
	return min + rand.Intn(max-min+1)
}

func combatCeilDiv(a, b int64) int64 {
	if b == 0 {
		return 0
	}
	return int64(math.Ceil(float64(a) / float64(b)))
}

func buildStatusResponse(
	session *database.ActiveCombat,
	user *database.User,
	enemy *CombatEnemyState,
	logs []CombatLogEntry,
	wasOffline bool,
	offlineMS int64,
	offlineEnemies int,
) CombatStatusResponse {
	return CombatStatusResponse{
		Status:           session.Status,
		IsActive:         session.Status == "active",
		ZoneKey:          session.ZoneKey,
		PlayerHPCurrent:  session.PlayerHPCurrent,
		PlayerMaxHP:      user.MaxHP,
		CurrentEnemy:     enemy,
		EnemiesDefeated:  session.EnemiesDefeated,
		TotalXPGained:    session.TotalXPGained,
		TotalMoneyGained: session.TotalMoneyGained,
		RecentLogs:       logs,
		SessionStartedAt: session.StartedAt.UnixMilli(),
		WasOffline:       wasOffline,
		OfflineTimeMS:    offlineMS,
		OfflineEnemies:   offlineEnemies,
	}
}
