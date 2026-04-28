package database

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
)

var DB *gorm.DB

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
	err = seedOreTypes()
	if err != nil {
		return err
	}

	// Seed character levels
	err = seedCharacterLevels()
	if err != nil {
		return err
	}

	// Migrate legacy flat OreInventory rows to new pivot table
	err = migrateOreInventoryToItems()
	if err != nil {
		return err
	}

	// Seed monsters
	err = seedMonsters()
	if err != nil {
		return err
	}

	return nil
}

func migrate() error {
	return DB.AutoMigrate(
		&User{},
		&CharacterLevel{},   // level master table
		&OreInventory{},     // kept for backward-compat / migration source
		&OreInventoryItem{}, // new pivot table
		&OreType{},
		&MiningSession{},
		&Monster{},
		&MonsterDrop{},
		&ActivityLog{},
	)
}

func seedOreTypes() error {
	ores := []OreType{
		{
			OreKey:          "copper_ore",
			OreName:         "Copper Ore",
			Icon:            "🪨",
			Color:           "#b87333",
			Difficulty:      "Common",
			MiningTimeMS:    3000,
			XPPerOre:        10,
			LevelRequired:   1,
			PickaxeRequired: "none",
			MaxQuantity:     500,
			SortOrder:       1,
		},
		{
			OreKey:          "iron_ore",
			OreName:         "Iron Ore",
			Icon:            "⚫",
			Color:           "#5a5a5a",
			Difficulty:      "Uncommon",
			MiningTimeMS:    6000,
			XPPerOre:        20,
			LevelRequired:   5,
			PickaxeRequired: "none",
			MaxQuantity:     300,
			SortOrder:       2,
		},
		{
			OreKey:          "gold_ore",
			OreName:         "Gold Ore",
			Icon:            "✨",
			Color:           "#ffd700",
			Difficulty:      "Rare",
			MiningTimeMS:    12000,
			XPPerOre:        40,
			LevelRequired:   15,
			PickaxeRequired: "iron_pickaxe",
			MaxQuantity:     100,
			SortOrder:       3,
		},
		{
			OreKey:          "mithril_ore",
			OreName:         "Mithril Ore",
			Icon:            "💎",
			Color:           "#00bfff",
			Difficulty:      "Epic",
			MiningTimeMS:    25000,
			XPPerOre:        75,
			LevelRequired:   30,
			PickaxeRequired: "gold_pickaxe",
			MaxQuantity:     50,
			SortOrder:       4,
		},
		{
			OreKey:          "diamond_ore",
			OreName:         "Diamond Ore",
			Icon:            "💠",
			Color:           "#00ffff",
			Difficulty:      "Legendary",
			MiningTimeMS:    60000,
			XPPerOre:        150,
			LevelRequired:   50,
			PickaxeRequired: "mithril_pickaxe",
			MaxQuantity:     25,
			SortOrder:       5,
		},
	}

	// Upsert: update existing rows so new fields (PickaxeRequired, MaxQuantity, SortOrder) are applied
	for _, ore := range ores {
		var existing OreType
		result := DB.Where("ore_key = ?", ore.OreKey).First(&existing)
		if result.Error != nil {
			// Not found — create
			DB.Create(&ore)
		} else {
			// Found — update with latest master data
			ore.ID = existing.ID
			ore.CreatedAt = existing.CreatedAt
			DB.Save(&ore)
		}
	}

	return nil
}

// migrateOreInventoryToItems converts the old flat OreInventory rows into
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
