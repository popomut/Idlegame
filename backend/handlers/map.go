package handlers

import (
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
	"idlegame-backend/database"
)

// ── Response shapes ────────────────────────────────────────────────────────

type AreaMonsterEntry struct {
	MonsterKey string           `json:"monster_key"`
	Weight     int              `json:"weight"`
	Monster    *MonsterResponse `json:"monster"`
}

type AreaResponse struct {
	AreaKey          string             `json:"area_key"`
	Name             string             `json:"name"`
	Icon             string             `json:"icon"`
	Description      string             `json:"description"`
	Difficulty       string             `json:"difficulty"`
	FightsBeforeBoss int                `json:"fights_before_boss"`
	BossMonsterKey   string             `json:"boss_monster_key"`
	BossMonster      *MonsterResponse   `json:"boss_monster"`
	Monsters         []AreaMonsterEntry `json:"monsters"`
}

type ContinentResponse struct {
	ContinentKey string         `json:"continent_key"`
	Name         string         `json:"name"`
	Icon         string         `json:"icon"`
	Description  string         `json:"description"`
	Difficulty   string         `json:"difficulty"`
	SortOrder    int            `json:"sort_order"`
	Areas        []AreaResponse `json:"areas"`
}

type CombatSessionResponse struct {
	AreaKey          string           `json:"area_key"`
	AreaName         string           `json:"area_name"`
	AreaIcon         string           `json:"area_icon"`
	FightCount       int              `json:"fight_count"`
	FightsBeforeBoss int              `json:"fights_before_boss"`
	Status           string           `json:"status"` // fighting, boss, complete
	CurrentMonster   *MonsterResponse `json:"current_monster"`
}

// ── Helpers ────────────────────────────────────────────────────────────────

func monsterByKey(key string) *MonsterResponse {
	var m database.Monster
	if err := database.DB.Where("monster_key = ?", key).First(&m).Error; err != nil {
		return nil
	}
	r := toMonsterResponse(m)
	return &r
}

func toMonsterResponse(m database.Monster) MonsterResponse {
	return MonsterResponse{
		ID: m.ID, MonsterKey: m.MonsterKey, Name: m.Name, Icon: m.Icon,
		Description: m.Description, HP: m.HP, DEX: m.DEX,
		AttackType: m.AttackType, AttackValue: m.AttackValue, PhysDef: m.PhysDef,
		ResistFire: m.ResistFire, ResistLightning: m.ResistLightning,
		ResistIce: m.ResistIce, ResistPoison: m.ResistPoison, ResistChaos: m.ResistChaos,
		MoneyDropMin: m.MoneyDropMin, MoneyDropMax: m.MoneyDropMax, XPDrop: m.XPDrop,
		SortOrder: m.SortOrder, Drops: []MonsterDropResponse{},
	}
}

// pickRandomMonster selects a weighted-random monster key for the area.
func pickRandomMonster(areaID uint) string {
	var entries []database.AreaMonster
	database.DB.Where("area_id = ?", areaID).Find(&entries)
	if len(entries) == 0 {
		return ""
	}
	total := 0
	for _, e := range entries {
		total += e.Weight
	}
	r := rand.Intn(total)
	cumulative := 0
	for _, e := range entries {
		cumulative += e.Weight
		if r < cumulative {
			return e.MonsterKey
		}
	}
	return entries[0].MonsterKey
}

func buildCombatResponse(session database.CombatSession) CombatSessionResponse {
	var area database.Area
	database.DB.Where("area_key = ?", session.AreaKey).First(&area)

	return CombatSessionResponse{
		AreaKey:          session.AreaKey,
		AreaName:         area.Name,
		AreaIcon:         area.Icon,
		FightCount:       session.FightCount,
		FightsBeforeBoss: session.FightsBeforeBoss,
		Status:           session.Status,
		CurrentMonster:   monsterByKey(session.CurrentMonsterKey),
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────

// GetContinents returns the full map: continents with nested areas and monster lists.
// GET /api/map/continents — public
func GetContinents(c *fiber.Ctx) error {
	var continents []database.Continent
	database.DB.Order("sort_order ASC").Find(&continents)

	var areas []database.Area
	database.DB.Preload("Continent").Order("sort_order ASC").Find(&areas)

	var areaMonsters []database.AreaMonster
	database.DB.Find(&areaMonsters)

	// Group area monsters by area ID
	amByArea := make(map[uint][]database.AreaMonster)
	for _, am := range areaMonsters {
		amByArea[am.AreaID] = append(amByArea[am.AreaID], am)
	}

	// Group areas by continent ID
	areasByCont := make(map[uint][]database.Area)
	for _, a := range areas {
		areasByCont[a.ContinentID] = append(areasByCont[a.ContinentID], a)
	}

	result := make([]ContinentResponse, 0, len(continents))
	for _, cont := range continents {
		contResp := ContinentResponse{
			ContinentKey: cont.ContinentKey,
			Name:         cont.Name,
			Icon:         cont.Icon,
			Description:  cont.Description,
			Difficulty:   cont.Difficulty,
			SortOrder:    cont.SortOrder,
		}

		for _, area := range areasByCont[cont.ID] {
			areaResp := AreaResponse{
				AreaKey:          area.AreaKey,
				Name:             area.Name,
				Icon:             area.Icon,
				Description:      area.Description,
				Difficulty:       area.Difficulty,
				FightsBeforeBoss: area.FightsBeforeBoss,
				BossMonsterKey:   area.BossMonsterKey,
				BossMonster:      monsterByKey(area.BossMonsterKey),
			}
			for _, am := range amByArea[area.ID] {
				entry := AreaMonsterEntry{
					MonsterKey: am.MonsterKey,
					Weight:     am.Weight,
					Monster:    monsterByKey(am.MonsterKey),
				}
				areaResp.Monsters = append(areaResp.Monsters, entry)
			}
			contResp.Areas = append(contResp.Areas, areaResp)
		}

		result = append(result, contResp)
	}

	return c.JSON(result)
}

// EnterArea starts (or restarts) a combat session in the given area.
// POST /api/map/enter  body: { "area_key": "dusty_outpost" }
func EnterArea(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var body struct {
		AreaKey string `json:"area_key"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var area database.Area
	if err := database.DB.Where("area_key = ?", body.AreaKey).First(&area).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Area not found"})
	}

	monsterKey := pickRandomMonster(area.ID)

	session := database.CombatSession{
		UserID:            userID,
		AreaKey:           area.AreaKey,
		FightCount:        0,
		FightsBeforeBoss:  area.FightsBeforeBoss,
		Status:            "fighting",
		CurrentMonsterKey: monsterKey,
		StartedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Upsert: replace any existing session
	var existing database.CombatSession
	if database.DB.Where("user_id = ?", userID).First(&existing).Error == nil {
		session.ID = existing.ID
		database.DB.Save(&session)
	} else {
		database.DB.Create(&session)
	}

	return c.JSON(buildCombatResponse(session))
}

// GetCombatSession returns the player's current session, or null if none.
// GET /api/map/session
func GetCombatSession(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var session database.CombatSession
	if err := database.DB.Where("user_id = ?", userID).First(&session).Error; err != nil {
		return c.JSON(fiber.Map{"session": nil})
	}

	return c.JSON(buildCombatResponse(session))
}

// AdvanceFight simulates a full player-vs-monster fight (server-side, cannot be cheated).
// Returns a FightResultResponse with the combat log, outcome, HP changes, and rewards.
// POST /api/map/advance
func AdvanceFight(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var session database.CombatSession
	if err := database.DB.Where("user_id = ?", userID).First(&session).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "No active combat session"})
	}
	if session.Status == "complete" {
		resp := buildCombatResponse(session)
		return c.JSON(FightResultResponse{Outcome: "already_complete", Session: &resp})
	}

	// Load the monster being fought (server-side — client cannot pick the monster)
	var monster database.Monster
	if err := database.DB.Where("monster_key = ?", session.CurrentMonsterKey).First(&monster).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Monster data not found"})
	}

	// Build player's effective stats (base + all equipped gear bonuses)
	player := GetPlayerCombatStats(userID)
	hpBefore := player.HP

	// ── Simulate the full fight ────────────────────────────────────────────
	outcome, combatLog, playerHPAfter := SimulateFight(player, monster)

	// ── Persist player HP changes ──────────────────────────────────────────
	// On death: restore to full HP so next session starts fresh.
	// On win:   save remaining HP.
	hpToSave := playerHPAfter
	if outcome == "player_dies" {
		hpToSave = player.MaxHP
	}
	database.DB.Model(&database.User{}).Where("id = ?", userID).Update("hp", hpToSave)

	result := FightResultResponse{
		Outcome:        outcome,
		Log:            combatLog,
		PlayerHPBefore: hpBefore,
		PlayerHPAfter:  playerHPAfter,
		PlayerMaxHP:    player.MaxHP,
	}

	if outcome == "player_wins" {
		// ── Award XP and money ───────────────────────────────────────────
		moneyGained := int64(monster.MoneyDropMin)
		if monster.MoneyDropMax > monster.MoneyDropMin {
			moneyGained += int64(rand.Intn(monster.MoneyDropMax - monster.MoneyDropMin + 1))
		}
		xpGained := int64(monster.XPDrop)

		// Load user, add money, save (AwardXP then loads fresh again for XP/level)
		var u database.User
		database.DB.First(&u, userID)
		u.Money += moneyGained
		database.DB.Save(&u)

		// Award XP — loads fresh user from DB (with updated money), handles level-up
		AwardXP(userID, xpGained)

		result.XPGained = xpGained
		result.MoneyGained = moneyGained

		// ── Advance combat session ───────────────────────────────────────
		session.FightCount++
		var area database.Area
		database.DB.Where("area_key = ?", session.AreaKey).First(&area)

		if session.Status == "boss" {
			session.Status = "complete"
		} else if session.FightCount >= session.FightsBeforeBoss {
			session.Status = "boss"
			session.CurrentMonsterKey = area.BossMonsterKey
		} else {
			session.CurrentMonsterKey = pickRandomMonster(area.ID)
		}
		session.UpdatedAt = time.Now()
		database.DB.Save(&session)

		sessionResp := buildCombatResponse(session)
		result.Session = &sessionResp

	} else {
		// Player died — clear the combat session; HP restored to full in DB above
		database.DB.Where("user_id = ?", userID).Delete(&database.CombatSession{})
		result.Session = nil
	}

	return c.JSON(result)
}

// FleeCombat abandons the current combat session.
// POST /api/map/flee
func FleeCombat(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	database.DB.Where("user_id = ?", userID).Delete(&database.CombatSession{})
	return c.JSON(fiber.Map{"success": true})
}
