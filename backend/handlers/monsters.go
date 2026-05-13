package handlers

import (
	"github.com/gofiber/fiber/v2"
	"idlegame-backend/database"
)

// MonsterResponse is the public-facing shape for a monster.
type MonsterResponse struct {
	ID          uint    `json:"id"`
	MonsterKey  string  `json:"monster_key"`
	Name        string  `json:"name"`
	Icon        string  `json:"icon"`
	Description string  `json:"description"`
	HP          int     `json:"hp"`
	DEX         int     `json:"dex"`
	AttackType  string  `json:"attack_type"`
	AttackValue int     `json:"attack_value"`
	PhysDef     int     `json:"phys_def"`

	ResistFire      int `json:"resist_fire"`
	ResistLightning int `json:"resist_lightning"`
	ResistIce       int `json:"resist_ice"`
	ResistPoison    int `json:"resist_poison"`
	ResistChaos     int `json:"resist_chaos"`

	MoneyDropMin int `json:"money_drop_min"`
	MoneyDropMax int `json:"money_drop_max"`
	XPDrop       int `json:"xp_drop"`

	SortOrder int              `json:"sort_order"`
	Drops     []MonsterDropResponse `json:"drops"`
}

// MonsterDropResponse is the public shape for a loot table entry.
type MonsterDropResponse struct {
	DropType string  `json:"drop_type"`
	DropKey  string  `json:"drop_key"`
	DropRate float64 `json:"drop_rate"`
	DropMin  int     `json:"drop_min"`
	DropMax  int     `json:"drop_max"`
}

// GetMonsters returns all monsters with their drop tables.
// GET /api/monsters — public endpoint (no auth required)
func GetMonsters(c *fiber.Ctx) error {
	var monsters []database.Monster
	if err := database.DB.Order("sort_order ASC").Find(&monsters).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load monsters"})
	}

	// Load drops for each monster
	var allDrops []database.MonsterDrop
	database.DB.Find(&allDrops)

	// Group drops by monster ID
	dropsByMonster := make(map[uint][]MonsterDropResponse)
	for _, d := range allDrops {
		dropsByMonster[d.MonsterID] = append(dropsByMonster[d.MonsterID], MonsterDropResponse{
			DropType: d.DropType,
			DropKey:  d.DropKey,
			DropRate: d.DropRate,
			DropMin:  d.DropMin,
			DropMax:  d.DropMax,
		})
	}

	result := make([]MonsterResponse, 0, len(monsters))
	for _, m := range monsters {
		drops := dropsByMonster[m.ID]
		if drops == nil {
			drops = []MonsterDropResponse{}
		}
		result = append(result, MonsterResponse{
			ID:          m.ID,
			MonsterKey:  m.MonsterKey,
			Name:        m.Name,
			Icon:        m.Icon,
			Description: m.Description,
			HP:          m.HP,
			DEX:         m.DEX,
			AttackType:  m.AttackType,
			AttackValue: m.AttackValue,
			PhysDef:     m.PhysDef,

			ResistFire:      m.ResistFire,
			ResistLightning: m.ResistLightning,
			ResistIce:       m.ResistIce,
			ResistPoison:    m.ResistPoison,
			ResistChaos:     m.ResistChaos,

			MoneyDropMin: m.MoneyDropMin,
			MoneyDropMax: m.MoneyDropMax,
			XPDrop:       m.XPDrop,

			SortOrder: m.SortOrder,
			Drops:     drops,
		})
	}

	return c.JSON(result)
}

// ── Admin endpoints (for development only — delete before production) ──────────

// AdminGetAllMonsters returns all monsters from the master table.
// GET /api/admin/monsters
func AdminGetAllMonsters(c *fiber.Ctx) error {
	var monsters []database.Monster
	if err := database.DB.Order("sort_order ASC, id ASC").Find(&monsters).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load monsters"})
	}

	result := make([]MonsterResponse, 0, len(monsters))
	for _, m := range monsters {
		result = append(result, MonsterResponse{
			ID:          m.ID,
			MonsterKey:  m.MonsterKey,
			Name:        m.Name,
			Icon:        m.Icon,
			Description: m.Description,
			HP:          m.HP,
			DEX:         m.DEX,
			AttackType:  m.AttackType,
			AttackValue: m.AttackValue,
			PhysDef:     m.PhysDef,

			ResistFire:      m.ResistFire,
			ResistLightning: m.ResistLightning,
			ResistIce:       m.ResistIce,
			ResistPoison:    m.ResistPoison,
			ResistChaos:     m.ResistChaos,

			MoneyDropMin: m.MoneyDropMin,
			MoneyDropMax: m.MoneyDropMax,
			XPDrop:       m.XPDrop,

			SortOrder: m.SortOrder,
			Drops:     []MonsterDropResponse{},
		})
	}

	return c.JSON(result)
}

// AdminCreateMonster creates a new monster.
// POST /api/admin/monsters
func AdminCreateMonster(c *fiber.Ctx) error {
	var body struct {
		MonsterKey      string `json:"monster_key"`
		Name            string `json:"name"`
		Icon            string `json:"icon"`
		Description     string `json:"description"`
		HP              int    `json:"hp"`
		DEX             int    `json:"dex"`
		AttackType      string `json:"attack_type"`
		AttackValue     int    `json:"attack_value"`
		PhysDef         int    `json:"phys_def"`
		ResistFire      int    `json:"resist_fire"`
		ResistLightning int    `json:"resist_lightning"`
		ResistIce       int    `json:"resist_ice"`
		ResistPoison    int    `json:"resist_poison"`
		ResistChaos     int    `json:"resist_chaos"`
		MoneyDropMin    int    `json:"money_drop_min"`
		MoneyDropMax    int    `json:"money_drop_max"`
		XPDrop          int    `json:"xp_drop"`
		SortOrder       int    `json:"sort_order"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if body.MonsterKey == "" || body.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "monster_key and name are required"})
	}

	monster := database.Monster{
		MonsterKey:      body.MonsterKey,
		Name:            body.Name,
		Icon:            body.Icon,
		Description:     body.Description,
		HP:              body.HP,
		DEX:             body.DEX,
		AttackType:      body.AttackType,
		AttackValue:     body.AttackValue,
		PhysDef:         body.PhysDef,
		ResistFire:      body.ResistFire,
		ResistLightning: body.ResistLightning,
		ResistIce:       body.ResistIce,
		ResistPoison:    body.ResistPoison,
		ResistChaos:     body.ResistChaos,
		MoneyDropMin:    body.MoneyDropMin,
		MoneyDropMax:    body.MoneyDropMax,
		XPDrop:          body.XPDrop,
		SortOrder:       body.SortOrder,
	}

	if err := database.DB.Create(&monster).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create monster"})
	}

	return c.Status(201).JSON(MonsterResponse{
		ID:          monster.ID,
		MonsterKey:  monster.MonsterKey,
		Name:        monster.Name,
		Icon:        monster.Icon,
		Description: monster.Description,
		HP:          monster.HP,
		DEX:         monster.DEX,
		AttackType:  monster.AttackType,
		AttackValue: monster.AttackValue,
		PhysDef:     monster.PhysDef,

		ResistFire:      monster.ResistFire,
		ResistLightning: monster.ResistLightning,
		ResistIce:       monster.ResistIce,
		ResistPoison:    monster.ResistPoison,
		ResistChaos:     monster.ResistChaos,

		MoneyDropMin: monster.MoneyDropMin,
		MoneyDropMax: monster.MoneyDropMax,
		XPDrop:       monster.XPDrop,

		SortOrder: monster.SortOrder,
		Drops:     []MonsterDropResponse{},
	})
}

// AdminUpdateMonster updates an existing monster.
// PUT /api/admin/monsters/:id
func AdminUpdateMonster(c *fiber.Ctx) error {
	id := c.Params("id")

	var monster database.Monster
	if err := database.DB.First(&monster, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Monster not found"})
	}

	var body struct {
		MonsterKey      string `json:"monster_key"`
		Name            string `json:"name"`
		Icon            string `json:"icon"`
		Description     string `json:"description"`
		HP              int    `json:"hp"`
		DEX             int    `json:"dex"`
		AttackType      string `json:"attack_type"`
		AttackValue     int    `json:"attack_value"`
		PhysDef         int    `json:"phys_def"`
		ResistFire      int    `json:"resist_fire"`
		ResistLightning int    `json:"resist_lightning"`
		ResistIce       int    `json:"resist_ice"`
		ResistPoison    int    `json:"resist_poison"`
		ResistChaos     int    `json:"resist_chaos"`
		MoneyDropMin    int    `json:"money_drop_min"`
		MoneyDropMax    int    `json:"money_drop_max"`
		XPDrop          int    `json:"xp_drop"`
		SortOrder       int    `json:"sort_order"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Update fields
	if body.MonsterKey != "" {
		monster.MonsterKey = body.MonsterKey
	}
	if body.Name != "" {
		monster.Name = body.Name
	}
	monster.Icon = body.Icon
	monster.Description = body.Description
	monster.HP = body.HP
	monster.DEX = body.DEX
	monster.AttackType = body.AttackType
	monster.AttackValue = body.AttackValue
	monster.PhysDef = body.PhysDef
	monster.ResistFire = body.ResistFire
	monster.ResistLightning = body.ResistLightning
	monster.ResistIce = body.ResistIce
	monster.ResistPoison = body.ResistPoison
	monster.ResistChaos = body.ResistChaos
	monster.MoneyDropMin = body.MoneyDropMin
	monster.MoneyDropMax = body.MoneyDropMax
	monster.XPDrop = body.XPDrop
	monster.SortOrder = body.SortOrder

	if err := database.DB.Save(&monster).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update monster"})
	}

	return c.JSON(MonsterResponse{
		ID:          monster.ID,
		MonsterKey:  monster.MonsterKey,
		Name:        monster.Name,
		Icon:        monster.Icon,
		Description: monster.Description,
		HP:          monster.HP,
		DEX:         monster.DEX,
		AttackType:  monster.AttackType,
		AttackValue: monster.AttackValue,
		PhysDef:     monster.PhysDef,

		ResistFire:      monster.ResistFire,
		ResistLightning: monster.ResistLightning,
		ResistIce:       monster.ResistIce,
		ResistPoison:    monster.ResistPoison,
		ResistChaos:     monster.ResistChaos,

		MoneyDropMin: monster.MoneyDropMin,
		MoneyDropMax: monster.MoneyDropMax,
		XPDrop:       monster.XPDrop,

		SortOrder: monster.SortOrder,
		Drops:     []MonsterDropResponse{},
	})
}

// AdminDeleteMonster deletes a monster.
// DELETE /api/admin/monsters/:id
func AdminDeleteMonster(c *fiber.Ctx) error {
	id := c.Params("id")

	var monster database.Monster
	if err := database.DB.First(&monster, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Monster not found"})
	}

	if err := database.DB.Delete(&monster).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete monster"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Monster deleted"})
}

