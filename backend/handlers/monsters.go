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
