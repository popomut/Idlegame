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

	// Migrate legacy flat OreInventory rows to new pivot table
	err = migrateOreInventoryToItems()
	if err != nil {
		return err
	}

	return nil
}

func migrate() error {
	return DB.AutoMigrate(
		&User{},
		&OreInventory{},     // kept for backward-compat / migration source
		&OreInventoryItem{}, // new pivot table
		&OreType{},
		&MiningSession{},
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
