package database

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"math"
)

var DB *gorm.DB

// generateOreSVG creates SVG for ore with copper ore structure but custom colors
func generateOreSVG(baseColor string) string {
	// baseColor = main ore color (e.g. "#b87333" for copper)
	// Create lighter and darker variants
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 500" width="100%%" height="100%%">
  <defs>
    <filter id="drop-shadow" x="-10%%" y="-10%%" width="130%%" height="130%%">
      <feDropShadow dx="0" dy="12" stdDeviation="8" flood-color="#0c0705" flood-opacity="0.4"/>
    </filter>
    <linearGradient id="ore-base" x1="0%%" y1="0%%" x2="0%%" y2="100%%">
      <stop offset="0%%" stop-color="%s"/>
      <stop offset="100%%" stop-color="%s"/>
    </linearGradient>
    <linearGradient id="ore-accent" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
      <stop offset="0%%" stop-color="%s"/>
      <stop offset="100%%" stop-color="%s"/>
    </linearGradient>
  </defs>
  <g filter="url(#drop-shadow)">
    <path d="M 315 110 L 420 110 L 460 150 L 475 220 L 550 260 L 580 340 L 590 410 L 560 450 L 490 470 L 400 445 L 375 445 L 290 460 L 245 425 L 240 345 L 280 295 L 270 240 L 315 110 Z" fill="#26120c" stroke="#26120c" stroke-width="16" stroke-linejoin="round"/>
    <path d="M 320 115 L 415 115 L 455 155 L 470 225 L 410 295 L 340 265 L 320 200 Z" fill="url(#ore-base)"/>
    <path d="M 320 115 L 415 115 L 390 190 L 320 200 Z" fill="%s" opacity="0.6"/>
    <path d="M 450 230 L 545 265 L 575 345 L 550 425 L 460 395 L 420 300 Z" fill="url(#ore-base)"/>
    <path d="M 450 230 L 545 265 L 510 330 L 420 300 Z" fill="%s" opacity="0.5"/>
    <path d="M 275 245 L 345 270 L 415 300 L 380 390 L 300 370 L 275 310 Z" fill="url(#ore-base)"/>
    <path d="M 275 245 L 345 270 L 330 330 L 275 310 Z" fill="%s" opacity="0.7"/>
    <path d="M 245 350 L 325 350 L 395 385 L 375 445 L 290 455 L 245 420 Z" fill="url(#ore-base)"/>
    <path d="M 245 350 L 325 350 L 340 410 L 245 420 Z" fill="%s" opacity="0.4"/>
    <path d="M 410 375 L 485 355 L 565 390 L 555 445 L 485 465 L 410 425 Z" fill="url(#ore-base)"/>
    <path d="M 410 375 L 485 355 L 500 420 L 410 425 Z" fill="%s" opacity="0.6"/>
    <g fill="#26120c" opacity="0.7">
      <circle cx="350" cy="140" r="3"/><circle cx="365" cy="155" r="2"/>
      <circle cx="430" cy="180" r="4"/><circle cx="445" cy="210" r="2.5"/>
      <circle cx="530" cy="310" r="3.5"/><circle cx="550" cy="350" r="2"/>
      <circle cx="490" cy="410" r="4"/><circle cx="530" cy="420" r="3"/>
      <circle cx="330" cy="410" r="3"/><circle cx="280" cy="390" r="2.5"/>
      <circle cx="370" cy="330" r="3.5"/>
    </g>
    <g fill="%s" opacity="0.4">
      <circle cx="340" cy="130" r="2"/><circle cx="415" cy="245" r="3"/>
      <circle cx="475" cy="265" r="2.5"/><circle cx="435" cy="395" r="2"/>
    </g>
    <path d="M 365 240 Q 372 240 372 233 Q 372 240 379 240 Q 372 240 372 247 Q 372 240 365 240 Z" fill="url(#ore-accent)"/>
    <g stroke="#3b190e" stroke-width="8" stroke-linecap="round" stroke-linejoin="round" fill="none" opacity="0.8">
      <path d="M 210 330 L 235 355"/><path d="M 195 345 Q 215 325 230 320"/>
      <path d="M 210 355 Q 220 340 240 335"/><path d="M 550 190 L 525 215"/>
      <path d="M 535 180 Q 555 195 560 215"/><path d="M 520 195 Q 540 205 545 225"/>
    </g>
  </g>
</svg>`,
		lightenColor(baseColor, 0.3),  // light variant
		darkenColor(baseColor, 0.4),   // dark variant
		lightenColor(baseColor, 0.5),  // accent light
		darkenColor(baseColor, 0.2),   // accent dark
		lightenColor(baseColor, 0.25), // highlights 1
		lightenColor(baseColor, 0.2),  // highlights 2
		lightenColor(baseColor, 0.15), // highlights 3
		lightenColor(baseColor, 0.1),  // highlights 4
		lightenColor(baseColor, 0.1),  // highlights 5
		lightenColor(baseColor, 0.4),  // sparkle
	)
}

// lightenColor makes a hex color lighter by factor (0-1)
func lightenColor(hexColor string, factor float64) string {
	if len(hexColor) != 7 {
		return hexColor
	}
	var r, g, b int
	fmt.Sscanf(hexColor, "#%02x%02x%02x", &r, &g, &b)
	r = int(float64(r) + (255-float64(r))*factor)
	g = int(float64(g) + (255-float64(g))*factor)
	b = int(float64(b) + (255-float64(b))*factor)
	if r > 255 {
		r = 255
	}
	if g > 255 {
		g = 255
	}
	if b > 255 {
		b = 255
	}
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// darkenColor makes a hex color darker by factor (0-1)
func darkenColor(hexColor string, factor float64) string {
	if len(hexColor) != 7 {
		return hexColor
	}
	var r, g, b int
	fmt.Sscanf(hexColor, "#%02x%02x%02x", &r, &g, &b)
	r = int(float64(r) * (1 - factor))
	g = int(float64(g) * (1 - factor))
	b = int(float64(b) * (1 - factor))
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func Init() error {
	// Open SQLite database using pure Go driver (no CGO needed)
	db, err := gorm.Open(sqlite.Open("idlegame.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	DB = db

	// Run migrations
	err = migrate()
	if err != nil {
		return err
	}

	// Seed ore types if they don't exist
	err = seedExtractableTypes()
	if err != nil {
		return err
	}

	err = seedOreTypes()
	if err != nil {
		return err
	}

	err = seedHerbTypes()
	if err != nil {
		return err
	}

	// Seed character levels
	err = seedCharacterLevels()
	if err != nil {
		return err
	}

	// Seed mining levels
	err = seedMiningLevels()
	if err != nil {
		return err
	}

	// Seed blacksmith levels
	err = seedBlacksmithLevels()
	if err != nil {
		return err
	}

	// Migrate legacy flat OreInventory rows to new pivot table
	err = migrateOreInventoryToItems()
	if err != nil {
		return err
	}

	// Seed craftable items (recipes)
	err = seedCraftableItems()
	if err != nil {
		return err
	}

	// Seed monsters
	err = seedMonsters()
	if err != nil {
		return err
	}

	// Seed map (continents, areas, area monsters)
	err = seedMap()
	if err != nil {
		return err
	}

	// Seed equipment master table
	err = seedEquipment()
	if err != nil {
		return err
	}

	// Back-fill starter equipment for existing users who have none
	err = grantStarterEquipmentToExistingUsers()
	if err != nil {
		return err
	}

	// Seed Chapter 1 quest chain
	err = seedQuests()
	if err != nil {
		return err
	}

	// Back-fill quest rows for existing users
	err = backfillUserQuests()
	if err != nil {
		return err
	}

	return nil
}

func migrate() error {
	return DB.AutoMigrate(
		&User{},
		&CharacterLevel{},   // level master table
		&MiningLevel{},      // mining level master table
		&UserMiningSkill{},  // user mining skill tracker
		&BlacksmithLevel{},  // blacksmith level master table
		&UserBlacksmithSkill{}, // user blacksmith skill tracker
		&OreInventory{},     // kept for backward-compat / migration source
		&ExtractableType{},  // extraction type master table (must be before OreType/HerbType)
		&OreInventoryItem{}, // new pivot table
		&OreType{},
		&HerbType{},
		&HerbInventoryItem{}, // herb inventory pivot table
		&MiningSession{},
		&CraftableItem{},    // blacksmith recipes
		&CraftRecipeIngredient{}, // recipe ingredients
		&UserIngotInventory{}, // ingot inventory pivot
		&UserPotionInventoryItem{}, // potion inventory pivot
		&BlacksmithSession{}, // active crafting session
		&Monster{},
		&MonsterDrop{},
		&Continent{},
		&Area{},
		&AreaMonster{},
		&CombatSession{},
		&ActiveCombat{},
		&Equipment{},
		&UserEquipment{},
		&UserEquippedSlot{},
		&Quest{},
		&QuestObjective{},
		&QuestReward{},
		&UserQuest{},
		&UserMonsterKills{},
		&ActivityLog{},
	)
}

func seedExtractableTypes() error {
	types := []ExtractableType{
		{
			TypeKey:   "ore",
			TypeName:  "Ore",
			Icon:      "⚙️",
			SortOrder: 1,
		},
		{
			TypeKey:   "herb",
			TypeName:  "Herb",
			Icon:      "🌿",
			SortOrder: 2,
		},
	}

	// Check if each type exists before creating
	for _, t := range types {
		var existing ExtractableType
		result := DB.Where("type_key = ?", t.TypeKey).First(&existing)
		if result.Error != nil {
			// Not found — create
			DB.Create(&t)
		}
	}

	return nil
}

func seedOreTypes() error {
	ores := []OreType{
		{
			OreKey:          "copper_ore",
			OreName:         "Copper Ore",
			Icon:            "🪨",
			Color:           "#b87333",
			SVG:             generateOreSVG("#b87333"),
			Difficulty:      "Common",
			MiningTimeMS:    3000,
			XPPerOre:        10,
			LevelRequired:   1,
			PickaxeRequired: "none",
			MaxQuantity:     99999,
			SortOrder:       1,
			BasePrice:       2,
		},
		{
			OreKey:          "silver_ore",
			OreName:         "Silver Ore",
			Icon:            "⚪",
			Color:           "#c0c0c0",
			SVG:             generateOreSVG("#c0c0c0"),
			Difficulty:      "Uncommon",
			MiningTimeMS:    4000,
			XPPerOre:        15,
			LevelRequired:   3,
			PickaxeRequired: "none",
			MaxQuantity:     99999,
			SortOrder:       2,
			BasePrice:       4,
		},
		{
			OreKey:          "iron_ore",
			OreName:         "Iron Ore",
			Icon:            "⚫",
			Color:           "#5a5a5a",
			SVG:             generateOreSVG("#5a5a5a"),
			Difficulty:      "Uncommon",
			MiningTimeMS:    6000,
			XPPerOre:        20,
			LevelRequired:   5,
			PickaxeRequired: "none",
			MaxQuantity:     99999,
			SortOrder:       3,
			BasePrice:       5,
		},
		{
			OreKey:          "bronze_ore",
			OreName:         "Bronze Ore",
			Icon:            "🟤",
			Color:           "#cd7f32",
			SVG:             generateOreSVG("#cd7f32"),
			Difficulty:      "Uncommon",
			MiningTimeMS:    5000,
			XPPerOre:        18,
			LevelRequired:   7,
			PickaxeRequired: "none",
			MaxQuantity:     99999,
			SortOrder:       4,
			BasePrice:       5,
		},
		{
			OreKey:          "gold_ore",
			OreName:         "Gold Ore",
			Icon:            "✨",
			Color:           "#ffd700",
			SVG:             generateOreSVG("#ffd700"),
			Difficulty:      "Rare",
			MiningTimeMS:    12000,
			XPPerOre:        40,
			LevelRequired:   15,
			PickaxeRequired: "iron_pickaxe",
			MaxQuantity:     99999,
			SortOrder:       5,
			BasePrice:       15,
		},
		{
			OreKey:          "platinum_ore",
			OreName:         "Platinum Ore",
			Icon:            "⭐",
			Color:           "#e5e4e2",
			SVG:             generateOreSVG("#e5e4e2"),
			Difficulty:      "Rare",
			MiningTimeMS:    15000,
			XPPerOre:        50,
			LevelRequired:   20,
			PickaxeRequired: "iron_pickaxe",
			MaxQuantity:     99999,
			SortOrder:       6,
			BasePrice:       20,
		},
		{
			OreKey:          "emerald_ore",
			OreName:         "Emerald Ore",
			Icon:            "💚",
			Color:           "#50c878",
			SVG:             generateOreSVG("#50c878"),
			Difficulty:      "Rare",
			MiningTimeMS:    18000,
			XPPerOre:        55,
			LevelRequired:   25,
			PickaxeRequired: "iron_pickaxe",
			MaxQuantity:     99999,
			SortOrder:       7,
			BasePrice:       25,
		},
		{
			OreKey:          "mithril_ore",
			OreName:         "Mithril Ore",
			Icon:            "💎",
			Color:           "#00bfff",
			SVG:             generateOreSVG("#00bfff"),
			Difficulty:      "Epic",
			MiningTimeMS:    25000,
			XPPerOre:        75,
			LevelRequired:   30,
			PickaxeRequired: "gold_pickaxe",
			MaxQuantity:     99999,
			SortOrder:       8,
			BasePrice:       40,
		},
		{
			OreKey:          "sapphire_ore",
			OreName:         "Sapphire Ore",
			Icon:            "💙",
			Color:           "#0f52ba",
			SVG:             generateOreSVG("#0f52ba"),
			Difficulty:      "Epic",
			MiningTimeMS:    30000,
			XPPerOre:        85,
			LevelRequired:   35,
			PickaxeRequired: "gold_pickaxe",
			MaxQuantity:     99999,
			SortOrder:       9,
			BasePrice:       50,
		},
		{
			OreKey:          "ruby_ore",
			OreName:         "Ruby Ore",
			Icon:            "❤️",
			Color:           "#e0115f",
			SVG:             generateOreSVG("#e0115f"),
			Difficulty:      "Epic",
			MiningTimeMS:    35000,
			XPPerOre:        100,
			LevelRequired:   40,
			PickaxeRequired: "gold_pickaxe",
			MaxQuantity:     99999,
			SortOrder:       10,
			BasePrice:       60,
		},
		{
			OreKey:          "titanium_ore",
			OreName:         "Titanium Ore",
			Icon:            "🔷",
			Color:           "#878681",
			SVG:             generateOreSVG("#878681"),
			Difficulty:      "Epic",
			MiningTimeMS:    45000,
			XPPerOre:        120,
			LevelRequired:   45,
			PickaxeRequired: "mithril_pickaxe",
			MaxQuantity:     99999,
			SortOrder:       11,
			BasePrice:       80,
		},
		{
			OreKey:          "diamond_ore",
			OreName:         "Diamond Ore",
			Icon:            "💠",
			Color:           "#00ffff",
			SVG:             generateOreSVG("#00ffff"),
			Difficulty:      "Legendary",
			MiningTimeMS:    60000,
			XPPerOre:        150,
			LevelRequired:   50,
			PickaxeRequired: "mithril_pickaxe",
			MaxQuantity:     99999,
			SortOrder:       12,
			BasePrice:       100,
		},
		{
			OreKey:          "obsidian_ore",
			OreName:         "Obsidian Ore",
			Icon:            "🖤",
			Color:           "#0b1107",
			SVG:             generateOreSVG("#0b1107"),
			Difficulty:      "Legendary",
			MiningTimeMS:    70000,
			XPPerOre:        160,
			LevelRequired:   52,
			PickaxeRequired: "mithril_pickaxe",
			MaxQuantity:     99999,
			SortOrder:       13,
			BasePrice:       110,
		},
		{
			OreKey:          "orichalcum_ore",
			OreName:         "Orichalcum Ore",
			Icon:            "🟡",
			Color:           "#b76e00",
			SVG:             generateOreSVG("#b76e00"),
			Difficulty:      "Legendary",
			MiningTimeMS:    80000,
			XPPerOre:        180,
			LevelRequired:   55,
			PickaxeRequired: "mithril_pickaxe",
			MaxQuantity:     99999,
			SortOrder:       14,
			BasePrice:       140,
		},
		{
			OreKey:          "celestial_ore",
			OreName:         "Celestial Ore",
			Icon:            "✨🌙",
			Color:           "#9932cc",
			SVG:             generateOreSVG("#9932cc"),
			Difficulty:      "Mythic",
			MiningTimeMS:    100000,
			XPPerOre:        200,
			LevelRequired:   60,
			PickaxeRequired: "mithril_pickaxe",
			MaxQuantity:     99999,
			SortOrder:       15,
			BasePrice:       200,
		},
	}

	// Get the "ore" extraction type ID
	var oreType ExtractableType
	DB.Where("type_key = ?", "ore").First(&oreType)
	if oreType.ID == 0 {
		// Should not happen if seedExtractableTypes ran first
		return fmt.Errorf("ore extraction type not found")
	}

	// Only create if doesn't exist — preserve user edits on restart
	for _, ore := range ores {
		ore.ExtractionTypeID = oreType.ID
		var existing OreType
		result := DB.Where("ore_key = ?", ore.OreKey).First(&existing)
		if result.Error != nil {
			// Not found — create
			DB.Create(&ore)
		} else {
			// Backfill extraction_type_id for old rows (fixes missing ores in dropdown)
			if existing.ExtractionTypeID == 0 {
				DB.Model(&existing).Update("extraction_type_id", oreType.ID)
			}
			// Backfill base_price for existing rows that never had it
			if existing.BasePrice == 0 && ore.BasePrice > 0 {
				DB.Model(&existing).Update("base_price", ore.BasePrice)
			}
			// Backfill SVG for existing rows that never had it
			if existing.SVG == "" && ore.SVG != "" {
				DB.Model(&existing).Update("svg", ore.SVG)
			}
		}
	}

	return nil
}

func seedHerbTypes() error {
	herbs := []HerbType{
		{
			HerbKey:       "lavender_herb",
			HerbName:      "Lavender",
			Icon:          "🟪",
			Color:         "#967bb6",
			Difficulty:    "Common",
			GatherTimeMS:  3000,
			XPPerHerb:     10,
			LevelRequired: 1,
			MaxQuantity:   99999,
			SortOrder:     1,
			BasePrice:     2,
		},
		{
			HerbKey:       "mint_herb",
			HerbName:      "Mint",
			Icon:          "🟩",
			Color:         "#00b359",
			Difficulty:    "Common",
			GatherTimeMS:  3500,
			XPPerHerb:     12,
			LevelRequired: 1,
			MaxQuantity:   99999,
			SortOrder:     2,
			BasePrice:     2,
		},
		{
			HerbKey:       "sage_herb",
			HerbName:      "Sage",
			Icon:          "🟫",
			Color:         "#8b6f47",
			Difficulty:    "Uncommon",
			GatherTimeMS:  4000,
			XPPerHerb:     15,
			LevelRequired: 3,
			MaxQuantity:   99999,
			SortOrder:     3,
			BasePrice:     3,
		},
		{
			HerbKey:       "rosemary_herb",
			HerbName:      "Rosemary",
			Icon:          "🌲",
			Color:         "#2d5016",
			Difficulty:    "Uncommon",
			GatherTimeMS:  4500,
			XPPerHerb:     18,
			LevelRequired: 5,
			MaxQuantity:   99999,
			SortOrder:     4,
			BasePrice:     3,
		},
		{
			HerbKey:       "thyme_herb",
			HerbName:      "Thyme",
			Icon:          "🟪",
			Color:         "#996633",
			Difficulty:    "Uncommon",
			GatherTimeMS:  4000,
			XPPerHerb:     15,
			LevelRequired: 4,
			MaxQuantity:   99999,
			SortOrder:     5,
			BasePrice:     3,
		},
		{
			HerbKey:       "moonflower_herb",
			HerbName:      "Moonflower",
			Icon:          "🌙",
			Color:         "#e6e6fa",
			Difficulty:    "Rare",
			GatherTimeMS:  8000,
			XPPerHerb:     30,
			LevelRequired: 10,
			MaxQuantity:   99999,
			SortOrder:     6,
			BasePrice:     8,
		},
		{
			HerbKey:       "bloodleaf_herb",
			HerbName:      "Bloodleaf",
			Icon:          "🩸",
			Color:         "#cc0000",
			Difficulty:    "Rare",
			GatherTimeMS:  10000,
			XPPerHerb:     35,
			LevelRequired: 12,
			MaxQuantity:   99999,
			SortOrder:     7,
			BasePrice:     10,
		},
		{
			HerbKey:       "starflower_herb",
			HerbName:      "Starflower",
			Icon:          "⭐",
			Color:         "#ffdd00",
			Difficulty:    "Rare",
			GatherTimeMS:  12000,
			XPPerHerb:     40,
			LevelRequired: 15,
			MaxQuantity:   99999,
			SortOrder:     8,
			BasePrice:     12,
		},
		{
			HerbKey:       "twilight_herb",
			HerbName:      "Twilight Herb",
			Icon:          "🌌",
			Color:         "#1a0033",
			Difficulty:    "Epic",
			GatherTimeMS:  20000,
			XPPerHerb:     60,
			LevelRequired: 25,
			MaxQuantity:   99999,
			SortOrder:     9,
			BasePrice:     25,
		},
		{
			HerbKey:       "ethereal_herb",
			HerbName:      "Ethereal Essence",
			Icon:          "✨",
			Color:         "#b3d9ff",
			Difficulty:    "Legendary",
			GatherTimeMS:  30000,
			XPPerHerb:     100,
			LevelRequired: 40,
			MaxQuantity:   99999,
			SortOrder:     10,
			BasePrice:     50,
		},
		{
			HerbKey:       "healgrass",
			HerbName:      "Healgrass",
			Icon:          "💚",
			Color:         "#00cc66",
			Difficulty:    "Common",
			GatherTimeMS:  3000,
			XPPerHerb:     10,
			LevelRequired: 1,
			MaxQuantity:   99999,
			SortOrder:     11,
			BasePrice:     2,
		},
		{
			HerbKey:       "silverleaf",
			HerbName:      "Silverleaf",
			Icon:          "🌟",
			Color:         "#c0c0c0",
			Difficulty:    "Uncommon",
			GatherTimeMS:  5000,
			XPPerHerb:     20,
			LevelRequired: 5,
			MaxQuantity:   99999,
			SortOrder:     12,
			BasePrice:     5,
		},
		{
			HerbKey:       "ironroot",
			HerbName:      "Ironroot",
			Icon:          "🪨",
			Color:         "#808080",
			Difficulty:    "Uncommon",
			GatherTimeMS:  4500,
			XPPerHerb:     18,
			LevelRequired: 8,
			MaxQuantity:   99999,
			SortOrder:     13,
			BasePrice:     4,
		},
	}

	// Get herb extraction type ID
	var herbType ExtractableType
	DB.Where("type_key = ?", "herb").First(&herbType)
	if herbType.ID == 0 {
		return fmt.Errorf("herb extraction type not found")
	}

	// Create herbs — skip if exists
	for _, herb := range herbs {
		herb.ExtractionTypeID = herbType.ID
		var existing HerbType
		result := DB.Where("herb_key = ?", herb.HerbKey).First(&existing)
		if result.Error != nil {
			// Not found — create
			DB.Create(&herb)
		}
	}

	return nil
}
// the new OreInventoryItem pivot rows. Safe to run multiple times.
func migrateOreInventoryToItems() error {
	// Nothing to do if new table already has data
	newCount := int64(0)
	DB.Model(&OreInventoryItem{}).Count(&newCount)
	if newCount > 0 {
		return nil
	}

	// Nothing to do if old table is empty
	oldCount := int64(0)
	DB.Model(&OreInventory{}).Count(&oldCount)
	if oldCount == 0 {
		return nil
	}

	// Build ore_key → ID lookup
	var oreTypes []OreType
	DB.Find(&oreTypes)
	oreTypeMap := make(map[string]uint)
	for _, ot := range oreTypes {
		oreTypeMap[ot.OreKey] = ot.ID
	}

	// Read all old inventory rows and create pivot records
	var oldInventories []OreInventory
	DB.Find(&oldInventories)

	for _, old := range oldInventories {
		legacyOres := map[string]int{
			"copper_ore":  old.CopperOre,
			"iron_ore":    old.IronOre,
			"gold_ore":    old.GoldOre,
			"mithril_ore": old.MithrilOre,
			"diamond_ore": old.DiamondOre,
		}
		for oreKey, qty := range legacyOres {
			if oreTypeID, ok := oreTypeMap[oreKey]; ok {
				item := OreInventoryItem{
					UserID:    old.UserID,
					OreTypeID: oreTypeID,
					Quantity:  qty,
				}
				DB.Create(&item)
			}
		}
	}
	return nil
}

// seedCharacterLevels upserts the level progression table.
// Formula: XP(n) = 100 * n^1.6  (exponential curve)
// Stats grow steadily: HP +10/level, Stamina +5/level, Str/Int/Dex +1 every 5 levels
func seedCharacterLevels() error {
	for lvl := 1; lvl <= 100; lvl++ {
		xpRequired := int64(float64(100) * float64(lvl*lvl) * 0.8)
		if lvl == 1 {
			xpRequired = 0 // Level 1 is the starting level
		}

		maxHP := 100 + (lvl-1)*10
		maxStamina := 50 + (lvl-1)*5
		str := 5 + (lvl-1)/5
		intStat := 5 + (lvl-1)/5
		dex := 5 + (lvl-1)/5

		cl := CharacterLevel{
			Level:      lvl,
			XPRequired: xpRequired,
			MaxHP:      maxHP,
			MaxStamina: maxStamina,
			Str:        str,
			Int:        intStat,
			Dex:        dex,
		}

		var existing CharacterLevel
		if err := DB.First(&existing, lvl).Error; err != nil {
			DB.Create(&cl)
		} else {
			DB.Save(&cl)
		}
	}
	return nil
}

// seedMiningLevels seeds the mining level progression table (separate from combat levels).
// Mining progression is steeper to give achievable short-term goals.
// Formula: XP(n) = 50 * n^1.3  (slower than combat)
// Max level: 50 (configurable in admin panel)
func seedMiningLevels() error {
	for lvl := 1; lvl <= 50; lvl++ {
		xpRequired := int(float64(50) * math.Pow(float64(lvl), 1.3))
		if lvl == 1 {
			xpRequired = 0 // Level 1 is the starting level
		}

		ml := MiningLevel{
			Level:      lvl,
			XPRequired: xpRequired,
		}

		var existing MiningLevel
		if err := DB.First(&existing, lvl).Error; err != nil {
			DB.Create(&ml)
		} else {
			DB.Save(&ml)
		}
	}
	return nil
}

// seedBlacksmithLevels seeds the blacksmith level progression table (separate from mining/combat).
// Formula: XP(n) = 60 * n^1.2  (similar pace to mining)
// Max level: 50 (configurable in admin panel)
func seedBlacksmithLevels() error {
	for lvl := 1; lvl <= 50; lvl++ {
		xpRequired := int(float64(60) * math.Pow(float64(lvl), 1.2))
		if lvl == 1 {
			xpRequired = 0 // Level 1 is the starting level
		}

		bl := BlacksmithLevel{
			Level:      lvl,
			XPRequired: xpRequired,
		}

		var existing BlacksmithLevel
		if err := DB.First(&existing, lvl).Error; err != nil {
			DB.Create(&bl)
		}
		// If found, do nothing — preserve user edits from admin panel
	}
	return nil
}

func seedCraftableItems() error {
	recipes := []CraftableItem{
		{
			Name:           "Copper Ingot",
			Description:    "A basic ingot for crafting",
			Icon:           "🟠",
			ItemKey:        "copper_ingot",
			OutputType:     "ingot",
			CraftingTimeMS: 5000,
			XPPerCraft:     20,
			LevelRequired:  1,
			MaxQuantity:    500,
			SortOrder:      1,
			BasePrice:      15,
		},
		{
			Name:           "Iron Ingot",
			Description:    "A strong ingot for better tools",
			Icon:           "⬛",
			ItemKey:        "iron_ingot",
			OutputType:     "ingot",
			CraftingTimeMS: 8000,
			XPPerCraft:     40,
			LevelRequired:  5,
			MaxQuantity:    300,
			SortOrder:      2,
			BasePrice:      30,
		},
		{
			Name:           "Gold Ingot",
			Description:    "A precious ingot",
			Icon:           "🟨",
			ItemKey:        "gold_ingot",
			OutputType:     "ingot",
			CraftingTimeMS: 12000,
			XPPerCraft:     60,
			LevelRequired:  15,
			MaxQuantity:    100,
			SortOrder:      3,
			BasePrice:      80,
		},
		{
			Name:           "Mithril Ingot",
			Description:    "A legendary ingot",
			Icon:           "🟪",
			ItemKey:        "mithril_ingot",
			OutputType:     "ingot",
			CraftingTimeMS: 20000,
			XPPerCraft:     100,
			LevelRequired:  30,
			MaxQuantity:    50,
			SortOrder:      4,
			BasePrice:      200,
		},
		{
			Name:           "Diamond Ingot",
			Description:    "The rarest ingot",
			Icon:           "🔹",
			ItemKey:        "diamond_ingot",
			OutputType:     "ingot",
			CraftingTimeMS: 30000,
			XPPerCraft:     150,
			LevelRequired:  50,
			MaxQuantity:    25,
			SortOrder:      5,
			BasePrice:      500,
		},
		{
			Name:           "Health Potion",
			Description:    "Restores vitality when needed",
			Icon:           "🔴",
			ItemKey:        "health_potion",
			OutputType:     "potion",
			CraftingTimeMS: 4000,
			XPPerCraft:     15,
			LevelRequired:  1,
			MaxQuantity:    300,
			SortOrder:      6,
			BasePrice:      20,
		},
		{
			Name:           "STR Potion",
			Description:    "Boosts physical strength temporarily",
			Icon:           "💪",
			ItemKey:        "str_potion",
			OutputType:     "potion",
			CraftingTimeMS: 7000,
			XPPerCraft:     35,
			LevelRequired:  10,
			MaxQuantity:    150,
			SortOrder:      7,
			BasePrice:      50,
		},
	}

	// Only create if doesn't exist — preserve user edits on restart
	for _, recipe := range recipes {
		var existing CraftableItem
		result := DB.Where("item_key = ?", recipe.ItemKey).First(&existing)
		if result.Error != nil {
			// Not found — create and add ingredients
			DB.Create(&recipe)
			
			// Add recipe ingredients based on item_key
			switch recipe.ItemKey {
			case "copper_ingot":
				DB.Create(&CraftRecipeIngredient{
					CraftableItemID: recipe.ID,
					IngredientType:  "ore",
					IngredientKey:   "copper_ore",
					QuantityRequired: 3,
				})
			case "iron_ingot":
				DB.Create(&CraftRecipeIngredient{
					CraftableItemID: recipe.ID,
					IngredientType:  "ore",
					IngredientKey:   "iron_ore",
					QuantityRequired: 4,
				})
				DB.Create(&CraftRecipeIngredient{
					CraftableItemID: recipe.ID,
					IngredientType:  "ore",
					IngredientKey:   "copper_ore",
					QuantityRequired: 2,
				})
			case "gold_ingot":
				DB.Create(&CraftRecipeIngredient{
					CraftableItemID: recipe.ID,
					IngredientType:  "ore",
					IngredientKey:   "gold_ore",
					QuantityRequired: 3,
				})
				DB.Create(&CraftRecipeIngredient{
					CraftableItemID: recipe.ID,
					IngredientType:  "ingot",
					IngredientKey:   "iron_ingot",
					QuantityRequired: 2,
				})
			case "mithril_ingot":
				DB.Create(&CraftRecipeIngredient{
					CraftableItemID: recipe.ID,
					IngredientType:  "ore",
					IngredientKey:   "mithril_ore",
					QuantityRequired: 3,
				})
				DB.Create(&CraftRecipeIngredient{
					CraftableItemID: recipe.ID,
					IngredientType:  "ingot",
					IngredientKey:   "gold_ingot",
					QuantityRequired: 2,
				})
			case "diamond_ingot":
				DB.Create(&CraftRecipeIngredient{
					CraftableItemID: recipe.ID,
					IngredientType:  "ore",
					IngredientKey:   "diamond_ore",
					QuantityRequired: 3,
				})
				DB.Create(&CraftRecipeIngredient{
					CraftableItemID: recipe.ID,
					IngredientType:  "ingot",
					IngredientKey:   "mithril_ingot",
					QuantityRequired: 3,
				})
			case "health_potion":
				DB.Create(&CraftRecipeIngredient{
					CraftableItemID: recipe.ID,
					IngredientType:  "herb",
					IngredientKey:   "healgrass",
					QuantityRequired: 2,
				})
				DB.Create(&CraftRecipeIngredient{
					CraftableItemID: recipe.ID,
					IngredientType:  "herb",
					IngredientKey:   "silverleaf",
					QuantityRequired: 1,
				})
			case "str_potion":
				DB.Create(&CraftRecipeIngredient{
					CraftableItemID: recipe.ID,
					IngredientType:  "herb",
					IngredientKey:   "ironroot",
					QuantityRequired: 3,
				})
				DB.Create(&CraftRecipeIngredient{
					CraftableItemID: recipe.ID,
					IngredientType:  "herb",
					IngredientKey:   "silverleaf",
					QuantityRequired: 2,
				})
			}
		}  else if existing.BasePrice == 0 && recipe.BasePrice > 0 {
			// Backfill base_price for existing rows
			DB.Model(&existing).Update("base_price", recipe.BasePrice)
		}
		// If found (and price set), do nothing — preserve user edits from admin panel
	}
	return nil
}

// seedMonsters upserts the 10 base monsters into the master table.
// To add a new monster: INSERT a row here — no code changes on frontend or backend needed.
func seedMonsters() error {
	monsters := []Monster{
		{
			MonsterKey:  "wasteland_scavenger",
			Name:        "Wasteland Scavenger",
			Icon:        "🧟",
			Description: "A desperate survivor picking through ruins. Weak but unpredictable.",
			HP: 40, DEX: 1, AttackType: "physical", AttackValue: 4, PhysDef: 0,
			MoneyDropMin: 1, MoneyDropMax: 5, XPDrop: 8, SortOrder: 1,
		},
		{
			MonsterKey:  "rad_rat",
			Name:        "Rad Rat",
			Icon:        "🐀",
			Description: "A rodent bloated by radiation. Its bite carries a toxic sting.",
			HP: 25, DEX: 3, AttackType: "poison", AttackValue: 3, PhysDef: 0,
			ResistPoison: 50,
			MoneyDropMin: 0, MoneyDropMax: 2, XPDrop: 6, SortOrder: 2,
		},
		{
			MonsterKey:  "toxic_crawler",
			Name:        "Toxic Crawler",
			Icon:        "🦂",
			Description: "An irradiated insect that secretes corrosive venom.",
			HP: 35, DEX: 2, AttackType: "poison", AttackValue: 6, PhysDef: 2,
			ResistPoison: 75,
			MoneyDropMin: 0, MoneyDropMax: 3, XPDrop: 10, SortOrder: 3,
		},
		{
			MonsterKey:  "rust_golem",
			Name:        "Rust Golem",
			Icon:        "🤖",
			Description: "A war machine left to corrode. Slow but nearly impenetrable.",
			HP: 120, DEX: 1, AttackType: "physical", AttackValue: 10, PhysDef: 8,
			ResistFire: 20, ResistLightning: 30,
			MoneyDropMin: 5, MoneyDropMax: 15, XPDrop: 20, SortOrder: 4,
		},
		{
			MonsterKey:  "chem_soldier",
			Name:        "Chem Soldier",
			Icon:        "☣️",
			Description: "A surviving soldier who weaponised chemical agents against all living things.",
			HP: 70, DEX: 2, AttackType: "fire", AttackValue: 8, PhysDef: 3,
			ResistFire: 40, ResistPoison: 20,
			MoneyDropMin: 8, MoneyDropMax: 20, XPDrop: 18, SortOrder: 5,
		},
		{
			MonsterKey:  "irradiated_hound",
			Name:        "Irradiated Hound",
			Icon:        "🐺",
			Description: "A wolf twisted by fallout. Faster than anything in the wastes.",
			HP: 55, DEX: 4, AttackType: "physical", AttackValue: 7, PhysDef: 1,
			ResistPoison: 25,
			MoneyDropMin: 2, MoneyDropMax: 8, XPDrop: 14, SortOrder: 6,
		},
		{
			MonsterKey:  "gas_mask_raider",
			Name:        "Gas Mask Raider",
			Icon:        "💀",
			Description: "An armed looter hiding behind salvaged gear. Carries crude explosives.",
			HP: 80, DEX: 2, AttackType: "fire", AttackValue: 9, PhysDef: 4,
			ResistFire: 15,
			MoneyDropMin: 10, MoneyDropMax: 30, XPDrop: 22, SortOrder: 7,
		},
		{
			MonsterKey:  "biohazard_slime",
			Name:        "Biohazard Slime",
			Icon:        "🟢",
			Description: "A gelatinous mass of mutated organic matter. Cold and corrosive to the touch.",
			HP: 60, DEX: 1, AttackType: "ice", AttackValue: 7, PhysDef: 0,
			ResistIce: 80, ResistPoison: 50, ResistFire: -20,
			MoneyDropMin: 0, MoneyDropMax: 5, XPDrop: 15, SortOrder: 8,
		},
		{
			MonsterKey:  "war_drone",
			Name:        "War Drone",
			Icon:        "🚁",
			Description: "An autonomous combat UAV still running its kill protocol.",
			HP: 90, DEX: 3, AttackType: "lightning", AttackValue: 11, PhysDef: 5,
			ResistLightning: 50, ResistFire: 10,
			MoneyDropMin: 15, MoneyDropMax: 40, XPDrop: 28, SortOrder: 9,
		},
		{
			MonsterKey:  "mutant_brute",
			Name:        "Mutant Brute",
			Icon:        "👹",
			Description: "A hulking mass of mutated muscle. The apex predator of the wastelands.",
			HP: 200, DEX: 1, AttackType: "physical", AttackValue: 18, PhysDef: 10,
			ResistPoison: 30, ResistChaos: 20,
			MoneyDropMin: 20, MoneyDropMax: 60, XPDrop: 45, SortOrder: 10,
		},
	}

	for _, m := range monsters {
		var existing Monster
		result := DB.Where("monster_key = ?", m.MonsterKey).First(&existing)
		if result.Error != nil {
			DB.Create(&m)
		} else {
			m.ID = existing.ID
			m.CreatedAt = existing.CreatedAt
			DB.Save(&m)
		}
	}
	return nil
}

// seedEquipment upserts the equipment master table.
// To add new equipment: INSERT a row here — no frontend or backend code changes needed.
func seedEquipment() error {
	items := []Equipment{
		// ── WEAPONS ──────────────────────────────────────────────────────────
		{
			EquipmentKey: "rusty_blade",
			Name: "Rusty Blade", Icon: "🗡️",
			Description: "A corroded knife salvaged from a fallen soldier. Better than bare hands.",
			Slot: "weapon", Rarity: "common",
			BaseAttack: 5, AttackType: "physical",
			ModifiersJSON: "[]", LevelRequired: 1, SortOrder: 1, BasePrice: 50,
		},
		{
			EquipmentKey: "combat_pistol",
			Name: "Combat Pistol", Icon: "🔫",
			Description: "Standard-issue sidearm. Reliable even after years of neglect.",
			Slot: "weapon", Rarity: "uncommon",
			BaseAttack: 12, AttackType: "physical",
			ModifiersJSON: `[{"type":"dex","value":2}]`, LevelRequired: 5, SortOrder: 2, BasePrice: 200,
		},
		{
			EquipmentKey: "incendiary_launcher",
			Name: "Incendiary Launcher", Icon: "🔥",
			Description: "Fires canisters of burning chemical gel. Leaves nothing but ash.",
			Slot: "weapon", Rarity: "rare",
			BaseAttack: 20, AttackType: "fire",
			ModifiersJSON: `[{"type":"resist_fire","value":10}]`, LevelRequired: 15, SortOrder: 3, BasePrice: 500,
		},
		{
			EquipmentKey: "tesla_coil_gun",
			Name: "Tesla Coil Gun", Icon: "⚡",
			Description: "Repurposed power-grid tech. Arcs through multiple targets.",
			Slot: "weapon", Rarity: "epic",
			BaseAttack: 30, AttackType: "lightning",
			ModifiersJSON: `[{"type":"dex","value":3},{"type":"int","value":3}]`, LevelRequired: 30, SortOrder: 4, BasePrice: 1200,
		},
		{
			EquipmentKey: "venom_blade",
			Name: "Venom Blade", Icon: "🐍",
			Description: "Coated in synthesised toxin. Each cut festers.",
			Slot: "weapon", Rarity: "rare",
			BaseAttack: 18, AttackType: "poison",
			ModifiersJSON: `[{"type":"dex","value":5}]`, LevelRequired: 20, SortOrder: 5, BasePrice: 600,
		},
		// ── HEAD ─────────────────────────────────────────────────────────────
		{
			EquipmentKey: "scrap_helmet",
			Name: "Scrap Helmet", Icon: "⛑️",
			Description: "Welded together from vehicle panels. Crude but effective.",
			Slot: "head", Rarity: "common",
			BaseDefence: 3,
			ModifiersJSON: "[]", LevelRequired: 1, SortOrder: 10, BasePrice: 40,
		},
		{
			EquipmentKey: "military_helmet",
			Name: "Military Helmet", Icon: "🪖",
			Description: "Pre-war composite helmet. Still rated for combat.",
			Slot: "head", Rarity: "uncommon",
			BaseDefence: 6,
			ModifiersJSON: `[{"type":"resist_fire","value":10}]`, LevelRequired: 5, SortOrder: 11, BasePrice: 180,
		},
		{
			EquipmentKey: "hazmat_hood",
			Name: "Hazmat Hood", Icon: "😷",
			Description: "Full-face chemical protection. Filters airborne toxins.",
			Slot: "head", Rarity: "rare",
			BaseDefence: 4,
			ModifiersJSON: `[{"type":"resist_poison","value":30}]`, LevelRequired: 15, SortOrder: 12, BasePrice: 450,
		},
		// ── CHEST ────────────────────────────────────────────────────────────
		{
			EquipmentKey: "tattered_vest",
			Name: "Tattered Vest", Icon: "🧥",
			Description: "Strips of leather sewn over a damaged flak jacket.",
			Slot: "chest", Rarity: "common",
			BaseDefence: 4,
			ModifiersJSON: "[]", LevelRequired: 1, SortOrder: 20, BasePrice: 60,
		},
		{
			EquipmentKey: "kevlar_vest",
			Name: "Kevlar Vest", Icon: "🦺",
			Description: "Multi-layer ballistic weave. Stops fragments and pistol rounds.",
			Slot: "chest", Rarity: "uncommon",
			BaseDefence: 10,
			ModifiersJSON: `[{"type":"resist_fire","value":15}]`, LevelRequired: 8, SortOrder: 21, BasePrice: 250,
		},
		{
			EquipmentKey: "nbc_suit",
			Name: "NBC Suit", Icon: "☢️",
			Description: "Nuclear-Biological-Chemical rated full-body suit. Invaluable in contaminated zones.",
			Slot: "chest", Rarity: "rare",
			BaseDefence: 8,
			ModifiersJSON: `[{"type":"resist_poison","value":40},{"type":"resist_chaos","value":20}]`, LevelRequired: 20, SortOrder: 22, BasePrice: 700,
		},
		// ── LEGS ─────────────────────────────────────────────────────────────
		{
			EquipmentKey: "scrap_leggings",
			Name: "Scrap Leggings", Icon: "🩲",
			Description: "Sheet metal strapped to canvas. Uncomfortable but protective.",
			Slot: "legs", Rarity: "common",
			BaseDefence: 3,
			ModifiersJSON: "[]", LevelRequired: 1, SortOrder: 30, BasePrice: 40,
		},
		{
			EquipmentKey: "combat_pants",
			Name: "Combat Pants", Icon: "👖",
			Description: "Reinforced tactical trousers with knee guards.",
			Slot: "legs", Rarity: "uncommon",
			BaseDefence: 7,
			ModifiersJSON: `[{"type":"dex","value":2}]`, LevelRequired: 5, SortOrder: 31, BasePrice: 180,
		},
		// ── SHIELD ───────────────────────────────────────────────────────────
		{
			EquipmentKey: "scrap_shield",
			Name: "Scrap Shield", Icon: "🛡️",
			Description: "A car door repurposed as a shield. Heavy but solid.",
			Slot: "shield", Rarity: "common",
			BaseDefence: 5,
			ModifiersJSON: "[]", LevelRequired: 1, SortOrder: 40, BasePrice: 50,
		},
		{
			EquipmentKey: "ballistic_shield",
			Name: "Ballistic Shield", Icon: "🔰",
			Description: "Police-grade riot shield. Rated for high-velocity impacts.",
			Slot: "shield", Rarity: "uncommon",
			BaseDefence: 12,
			ModifiersJSON: `[{"type":"resist_fire","value":10}]`, LevelRequired: 10, SortOrder: 41, BasePrice: 220,
		},
		// ── RINGS ────────────────────────────────────────────────────────────
		{
			EquipmentKey: "strength_band",
			Name: "Strength Band", Icon: "💪",
			Description: "A weighted training band that permanently enhances muscle output.",
			Slot: "ring", Rarity: "common",
			ModifiersJSON: `[{"type":"str","value":3}]`, LevelRequired: 1, SortOrder: 50, BasePrice: 60,
		},
		{
			EquipmentKey: "toxin_ring",
			Name: "Toxin Ring", Icon: "💍",
			Description: "Contains a slow-release antitoxin compound. Grants poison resistance.",
			Slot: "ring", Rarity: "uncommon",
			ModifiersJSON: `[{"type":"resist_poison","value":20}]`, LevelRequired: 5, SortOrder: 51, BasePrice: 200,
		},
		{
			EquipmentKey: "lightning_ward",
			Name: "Lightning Ward", Icon: "⚡",
			Description: "A Faraday-cage ring that dissipates electrical energy.",
			Slot: "ring", Rarity: "rare",
			ModifiersJSON: `[{"type":"resist_lightning","value":25}]`, LevelRequired: 15, SortOrder: 52, BasePrice: 450,
		},
		// ── AMULET ───────────────────────────────────────────────────────────
		{
			EquipmentKey: "dog_tag_amulet",
			Name: "Dog Tag Amulet", Icon: "🪪",
			Description: "The tags of a fallen ally. Wearing them sharpens your edge.",
			Slot: "amulet", Rarity: "common",
			ModifiersJSON: `[{"type":"str","value":2},{"type":"dex","value":2}]`, LevelRequired: 1, SortOrder: 60, BasePrice: 70,
		},
		{
			EquipmentKey: "commanders_amulet",
			Name: "Commander's Amulet", Icon: "🎖️",
			Description: "Recovered from a high-ranking officer. Radiates authority and power.",
			Slot: "amulet", Rarity: "epic",
			ModifiersJSON: `[{"type":"str","value":5},{"type":"int","value":5},{"type":"dex","value":5}]`, LevelRequired: 40, SortOrder: 61, BasePrice: 1500,
		},
	}

	for _, item := range items {
		var existing Equipment
		result := DB.Where("equipment_key = ?", item.EquipmentKey).First(&existing)
		if result.Error != nil {
			// Equipment doesn't exist, create it
			DB.Create(&item)
		} else if existing.BasePrice == 0 && item.BasePrice > 0 {
			// Backfill base_price for existing rows
			DB.Model(&existing).Update("base_price", item.BasePrice)
		}
		// If it exists (and price set), skip (preserve any admin modifications)
	}
	return nil
}

// seedMap upserts continents, areas, and area-monster assignments.
// To add a new area or continent: INSERT rows here — no code changes needed.
func seedMap() error {
	// ── Continents ────────────────────────────────────────────────────────
	continents := []Continent{
		{ContinentKey: "scorched_wastes",    Name: "Scorched Wastes",    Icon: "🏜️", Difficulty: "easy",    Description: "Sun-scorched badlands stripped bare by the final war. Weak survivors and irradiated vermin roam here.", SortOrder: 1},
		{ContinentKey: "industrial_ruins",   Name: "Industrial Ruins",   Icon: "🏭", Difficulty: "medium",  Description: "Collapsed factories leaking chemical waste. Automated machines still follow their kill protocols.", SortOrder: 2},
		{ContinentKey: "irradiated_badlands",Name: "Irradiated Badlands",Icon: "☢️", Difficulty: "hard",    Description: "Saturated with fallout from the dirty bombs. Mutations run rampant.", SortOrder: 3},
		{ContinentKey: "the_dead_zone",      Name: "The Dead Zone",      Icon: "💀", Difficulty: "extreme", Description: "Ground zero of the final strike. Nothing should be alive here — but something is.", SortOrder: 4},
	}
	continentIDs := make(map[string]uint)
	for _, c := range continents {
		var existing Continent
		if DB.Where("continent_key = ?", c.ContinentKey).First(&existing).Error != nil {
			DB.Create(&c)
			continentIDs[c.ContinentKey] = c.ID
		} else {
			c.ID = existing.ID
			c.CreatedAt = existing.CreatedAt
			DB.Save(&c)
			continentIDs[c.ContinentKey] = existing.ID
		}
	}

	// ── Areas ─────────────────────────────────────────────────────────────
	type areaSeed struct {
		ContinentKey     string
		AreaKey          string
		Name             string
		Icon             string
		Description      string
		Difficulty       string
		FightsBeforeBoss int
		BossMonsterKey   string
		SortOrder        int
	}
	areas := []areaSeed{
		// Scorched Wastes
		{"scorched_wastes", "dusty_outpost",     "Dusty Outpost",     "🏕️", "A crumbling waystation. Scavengers pick through the rubble.",          "easy",    5,  "irradiated_hound",  1},
		{"scorched_wastes", "crumbled_highway",  "Crumbled Highway",  "🛣️", "A raised road buckled by the blast. Rodents nest in the wreckage.",    "easy",    5,  "chem_soldier",      2},
		{"scorched_wastes", "abandoned_refinery","Abandoned Refinery","🏗️", "A fuel depot gone silent. Toxic fumes linger near the tanks.",         "medium",  7,  "gas_mask_raider",   3},
		// Industrial Ruins
		{"industrial_ruins", "factory_floor",   "Factory Floor",     "⚙️",  "Assembly lines frozen mid-cycle. War machines patrol the aisles.",     "medium",  7,  "rust_golem",        1},
		{"industrial_ruins", "sewer_network",   "Sewer Network",     "🕳️", "Flooded tunnels beneath the plant. Slimes and vermin breed here.",      "medium",  7,  "biohazard_slime",   2},
		{"industrial_ruins", "collapsed_bridge","Collapsed Bridge",  "🌉", "A suspension bridge sagging into the river. Hounds hunt in packs.",     "hard",    10, "war_drone",         3},
		// Irradiated Badlands
		{"irradiated_badlands","toxic_swamp",   "Toxic Swamp",       "🌿", "A bog of chemical runoff. Slimes and crawlers thrive in the filth.",    "hard",    10, "biohazard_slime",   1},
		{"irradiated_badlands","radiation_fields","Radiation Fields", "☢️", "Open plains pulsing with gamma radiation. Fast predators here.",        "hard",    10, "war_drone",         2},
		{"irradiated_badlands","chemical_plant", "Chemical Plant",   "🧪", "An active chem lab still producing weapons. Soldiers guard it.",         "extreme", 12, "chem_soldier",      3},
		// The Dead Zone
		{"the_dead_zone","mass_grave",     "Mass Grave",       "💀", "Thousands of bodies. Something has reanimated them.",                    "extreme", 12, "mutant_brute",      1},
		{"the_dead_zone","command_bunker", "Command Bunker",   "🏰", "A reinforced military complex. Drones still respond to old orders.",     "extreme", 12, "war_drone",         2},
		{"the_dead_zone","ground_zero",    "Ground Zero",      "☠️", "The crater of the final detonation. The apex predator rules here.",      "extreme", 15, "mutant_brute",      3},
	}
	areaIDs := make(map[string]uint)
	for _, a := range areas {
		cid := continentIDs[a.ContinentKey]
		area := Area{
			ContinentID: cid, AreaKey: a.AreaKey, Name: a.Name, Icon: a.Icon,
			Description: a.Description, Difficulty: a.Difficulty,
			FightsBeforeBoss: a.FightsBeforeBoss, BossMonsterKey: a.BossMonsterKey, SortOrder: a.SortOrder,
		}
		var existing Area
		if DB.Where("area_key = ?", a.AreaKey).First(&existing).Error != nil {
			DB.Create(&area)
			areaIDs[a.AreaKey] = area.ID
		} else {
			area.ID = existing.ID
			area.CreatedAt = existing.CreatedAt
			DB.Save(&area)
			areaIDs[a.AreaKey] = existing.ID
		}
	}

	// ── Area monsters ─────────────────────────────────────────────────────
	type amSeed struct{ AreaKey, MonsterKey string; Weight int }
	areaMonsters := []amSeed{
		{"dusty_outpost",     "wasteland_scavenger", 3},
		{"dusty_outpost",     "rad_rat",             2},
		{"crumbled_highway",  "rad_rat",             3},
		{"crumbled_highway",  "toxic_crawler",       2},
		{"crumbled_highway",  "wasteland_scavenger", 1},
		{"abandoned_refinery","toxic_crawler",       3},
		{"abandoned_refinery","gas_mask_raider",     2},
		{"factory_floor",     "rust_golem",          2},
		{"factory_floor",     "gas_mask_raider",     3},
		{"sewer_network",     "rad_rat",             2},
		{"sewer_network",     "biohazard_slime",     3},
		{"sewer_network",     "toxic_crawler",       2},
		{"collapsed_bridge",  "irradiated_hound",    3},
		{"collapsed_bridge",  "gas_mask_raider",     2},
		{"collapsed_bridge",  "rust_golem",          1},
		{"toxic_swamp",       "biohazard_slime",     3},
		{"toxic_swamp",       "toxic_crawler",       2},
		{"radiation_fields",  "irradiated_hound",    2},
		{"radiation_fields",  "chem_soldier",        3},
		{"chemical_plant",    "chem_soldier",        3},
		{"chemical_plant",    "biohazard_slime",     2},
		{"mass_grave",        "mutant_brute",        2},
		{"mass_grave",        "gas_mask_raider",     3},
		{"command_bunker",    "war_drone",           3},
		{"command_bunker",    "chem_soldier",        2},
		{"ground_zero",       "mutant_brute",        3},
		{"ground_zero",       "war_drone",           2},
	}
	for _, am := range areaMonsters {
		areaID := areaIDs[am.AreaKey]
		var existing AreaMonster
		if DB.Where("area_id = ? AND monster_key = ?", areaID, am.MonsterKey).First(&existing).Error != nil {
			DB.Create(&AreaMonster{AreaID: areaID, MonsterKey: am.MonsterKey, Weight: am.Weight})
		} else {
			existing.Weight = am.Weight
			DB.Save(&existing)
		}
	}
	return nil
}

// grantStarterEquipmentToExistingUsers back-fills starter gear for any user
// who currently has zero items in their equipment bag (idempotent).
func grantStarterEquipmentToExistingUsers() error {
	starterKeys := []string{
		"rusty_blade", "scrap_helmet", "tattered_vest",
		"scrap_leggings", "scrap_shield", "strength_band", "dog_tag_amulet",
	}

	// Build equipment_key → ID lookup
	equipMap := make(map[string]uint)
	for _, key := range starterKeys {
		var e Equipment
		if err := DB.Where("equipment_key = ?", key).First(&e).Error; err == nil {
			equipMap[key] = e.ID
		}
	}

	// Find all users
	var users []User
	DB.Find(&users)

	for _, u := range users {
		var count int64
		DB.Model(&UserEquipment{}).Where("user_id = ?", u.ID).Count(&count)
		if count > 0 {
			continue // already has gear
		}
		// Grant starter items
		for _, key := range starterKeys {
			if id, ok := equipMap[key]; ok {
				DB.Create(&UserEquipment{
					UserID:      u.ID,
					EquipmentID: id,
					ObtainedAt:  u.CreatedAt,
				})
			}
		}
	}
	return nil
}

func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Log helper for debugging
func LogError(msg string, err error) {
	if err != nil {
		log.Printf("[ERROR] %s: %v\n", msg, err)
	}
}

// ── Quest seeding ──────────────────────────────────────────────────────────

// seedQuests creates the Chapter 1 quest chain if it doesn't already exist.
func seedQuests() error {
	quests := []Quest{
		{
			QuestKey:       "ch1_first_blood",
			Title:          "First Blood",
			Chapter:        1,
			SortOrder:      1,
			RequiresQuestID: 0,
			IntroText:      "The wasteland is unforgiving. Your first lesson: survive. Head into the ruins and bring something down. Prove you belong here.",
			CompletionText: "You did it. First kill. It won't be your last — but it matters. The road ahead is long.",
		},
		{
			QuestKey:       "ch1_ore_collector",
			Title:          "Ore Collector",
			Chapter:        1,
			SortOrder:      2,
			IntroText:      "Fighting keeps you alive, but resources build your future. Get your hands dirty at the mines — we need copper, and lots of it.",
			CompletionText: "A solid haul. Copper isn't glamorous but it's the foundation of everything. Hold onto it.",
		},
		{
			QuestKey:       "ch1_forged_in_fire",
			Title:          "Forged in Fire",
			Chapter:        1,
			SortOrder:      3,
			IntroText:      "Raw ore is just rock until a smith shapes it. Get to the Blacksmith and forge your first ingot. That's where real value is created.",
			CompletionText: "That ingot came from nothing — rock, heat, and skill. You're starting to understand how this world works.",
		},
		{
			QuestKey:       "ch1_deep_delver",
			Title:          "Deep Delver",
			Chapter:        1,
			SortOrder:      4,
			IntroText:      "Copper will only take you so far. Push your mining deeper. The tougher seams are where real rewards lie.",
			CompletionText: "Level 3 in the mines. You've earned access to iron deposits. The real grind begins now.",
		},
		{
			QuestKey:       "ch1_proving_grounds",
			Title:          "Proving Grounds",
			Chapter:        1,
			SortOrder:      5,
			IntroText:      "Word spreads fast out here. You've shown you can mine and fight, but can you do both? This is your final trial for Chapter 1.",
			CompletionText: "Chapter 1 complete. You've survived, gathered, crafted, and fought. You're no longer a recruit. What comes next will test you harder.",
		},
	}

	// Insert quests, capture IDs for objectives/rewards and prerequisite chaining
	questIDs := map[string]uint{}
	for i := range quests {
		var existing Quest
		if DB.Where("quest_key = ?", quests[i].QuestKey).First(&existing).Error != nil {
			// Set prerequisite based on previous quest (except first)
			if i > 0 {
				quests[i].RequiresQuestID = questIDs[quests[i-1].QuestKey]
			}
			DB.Create(&quests[i])
			questIDs[quests[i].QuestKey] = quests[i].ID
		} else {
			questIDs[existing.QuestKey] = existing.ID
		}
	}

	// Seed objectives
	type objSeed struct {
		questKey      string
		objectiveType string
		targetKey     string
		targetCount   int
		displayText   string
	}
	objectives := []objSeed{
		{"ch1_first_blood",      "kill",               "",            1,  "Defeat 1 enemy"},
		{"ch1_ore_collector",    "mine",               "copper_ore",  50, "Mine 50 Copper Ore"},
		{"ch1_forged_in_fire",   "craft",              "copper_ingot", 1, "Craft 1 Copper Ingot"},
		{"ch1_deep_delver",      "reach_mining_level", "",            3,  "Reach Mining Level 3"},
		{"ch1_proving_grounds",  "kill",               "",            10, "Defeat 10 enemies"},
		{"ch1_proving_grounds",  "deliver",            "copper_ore",  20, "Deliver 20 Copper Ore"},
	}
	for _, o := range objectives {
		qID, ok := questIDs[o.questKey]
		if !ok {
			continue
		}
		var existing QuestObjective
		if DB.Where("quest_id = ? AND objective_type = ? AND target_key = ?", qID, o.objectiveType, o.targetKey).First(&existing).Error != nil {
			DB.Create(&QuestObjective{
				QuestID:       qID,
				ObjectiveType: o.objectiveType,
				TargetKey:     o.targetKey,
				TargetCount:   o.targetCount,
				DisplayText:   o.displayText,
			})
		}
	}

	// Seed rewards
	type rewSeed struct {
		questKey   string
		rewardType string
		rewardKey  string
		amount     int64
	}
	rewards := []rewSeed{
		{"ch1_first_blood",     "xp",        "",        50},
		{"ch1_first_blood",     "money",      "",        10},
		{"ch1_ore_collector",   "xp",        "",        100},
		{"ch1_ore_collector",   "money",      "",        25},
		{"ch1_forged_in_fire",  "xp",        "",        150},
		{"ch1_forged_in_fire",  "equipment",  "rusty_blade", 1},
		{"ch1_deep_delver",     "xp",        "",        200},
		{"ch1_deep_delver",     "money",      "",        50},
		{"ch1_proving_grounds", "xp",        "",        300},
		{"ch1_proving_grounds", "money",      "",        100},
	}
	for _, r := range rewards {
		qID, ok := questIDs[r.questKey]
		if !ok {
			continue
		}
		var existing QuestReward
		if DB.Where("quest_id = ? AND reward_type = ? AND reward_key = ?", qID, r.rewardType, r.rewardKey).First(&existing).Error != nil {
			DB.Create(&QuestReward{
				QuestID:    qID,
				RewardType: r.rewardType,
				RewardKey:  r.rewardKey,
				Amount:     r.amount,
			})
		}
	}

	return nil
}

// InitUserQuests creates UserQuest rows for a newly registered or guest user.
// The first quest (no prerequisite) is set to available; others start locked.
func InitUserQuests(userID uint) {
	var allQuests []Quest
	DB.Order("sort_order ASC").Find(&allQuests)

	for _, q := range allQuests {
		var existing UserQuest
		if DB.Where("user_id = ? AND quest_id = ?", userID, q.ID).First(&existing).Error != nil {
			status := "locked"
			if q.RequiresQuestID == 0 {
				status = "available"
			}
			DB.Create(&UserQuest{
				UserID:  userID,
				QuestID: q.ID,
				Status:  status,
			})
		}
	}
}

// backfillUserQuests ensures existing users have UserQuest rows for any quests added after their registration.
func backfillUserQuests() error {
	var users []User
	DB.Find(&users)
	for _, u := range users {
		InitUserQuests(u.ID)
	}
	return nil
}
