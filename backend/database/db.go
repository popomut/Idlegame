package database

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"math"
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
		&OreInventoryItem{}, // new pivot table
		&OreType{},
		&MiningSession{},
		&CraftableItem{},    // blacksmith recipes
		&CraftRecipeIngredient{}, // recipe ingredients
		&UserIngotInventory{}, // ingot inventory pivot
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
			OreKey:          "silver_ore",
			OreName:         "Silver Ore",
			Icon:            "⚪",
			Color:           "#c0c0c0",
			Difficulty:      "Uncommon",
			MiningTimeMS:    4000,
			XPPerOre:        15,
			LevelRequired:   3,
			PickaxeRequired: "none",
			MaxQuantity:     400,
			SortOrder:       2,
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
			SortOrder:       3,
		},
		{
			OreKey:          "bronze_ore",
			OreName:         "Bronze Ore",
			Icon:            "🟤",
			Color:           "#cd7f32",
			Difficulty:      "Uncommon",
			MiningTimeMS:    5000,
			XPPerOre:        18,
			LevelRequired:   7,
			PickaxeRequired: "none",
			MaxQuantity:     350,
			SortOrder:       4,
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
			SortOrder:       5,
		},
		{
			OreKey:          "platinum_ore",
			OreName:         "Platinum Ore",
			Icon:            "⭐",
			Color:           "#e5e4e2",
			Difficulty:      "Rare",
			MiningTimeMS:    15000,
			XPPerOre:        50,
			LevelRequired:   20,
			PickaxeRequired: "iron_pickaxe",
			MaxQuantity:     80,
			SortOrder:       6,
		},
		{
			OreKey:          "emerald_ore",
			OreName:         "Emerald Ore",
			Icon:            "💚",
			Color:           "#50c878",
			Difficulty:      "Rare",
			MiningTimeMS:    18000,
			XPPerOre:        55,
			LevelRequired:   25,
			PickaxeRequired: "iron_pickaxe",
			MaxQuantity:     70,
			SortOrder:       7,
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
			SortOrder:       8,
		},
		{
			OreKey:          "sapphire_ore",
			OreName:         "Sapphire Ore",
			Icon:            "💙",
			Color:           "#0f52ba",
			Difficulty:      "Epic",
			MiningTimeMS:    30000,
			XPPerOre:        85,
			LevelRequired:   35,
			PickaxeRequired: "gold_pickaxe",
			MaxQuantity:     40,
			SortOrder:       9,
		},
		{
			OreKey:          "ruby_ore",
			OreName:         "Ruby Ore",
			Icon:            "❤️",
			Color:           "#e0115f",
			Difficulty:      "Epic",
			MiningTimeMS:    35000,
			XPPerOre:        100,
			LevelRequired:   40,
			PickaxeRequired: "gold_pickaxe",
			MaxQuantity:     35,
			SortOrder:       10,
		},
		{
			OreKey:          "titanium_ore",
			OreName:         "Titanium Ore",
			Icon:            "🔷",
			Color:           "#878681",
			Difficulty:      "Epic",
			MiningTimeMS:    45000,
			XPPerOre:        120,
			LevelRequired:   45,
			PickaxeRequired: "mithril_pickaxe",
			MaxQuantity:     30,
			SortOrder:       11,
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
			SortOrder:       12,
		},
		{
			OreKey:          "obsidian_ore",
			OreName:         "Obsidian Ore",
			Icon:            "🖤",
			Color:           "#0b1107",
			Difficulty:      "Legendary",
			MiningTimeMS:    70000,
			XPPerOre:        160,
			LevelRequired:   52,
			PickaxeRequired: "mithril_pickaxe",
			MaxQuantity:     20,
			SortOrder:       13,
		},
		{
			OreKey:          "orichalcum_ore",
			OreName:         "Orichalcum Ore",
			Icon:            "🟡",
			Color:           "#b76e00",
			Difficulty:      "Legendary",
			MiningTimeMS:    80000,
			XPPerOre:        180,
			LevelRequired:   55,
			PickaxeRequired: "mithril_pickaxe",
			MaxQuantity:     15,
			SortOrder:       14,
		},
		{
			OreKey:          "celestial_ore",
			OreName:         "Celestial Ore",
			Icon:            "✨🌙",
			Color:           "#9932cc",
			Difficulty:      "Mythic",
			MiningTimeMS:    100000,
			XPPerOre:        200,
			LevelRequired:   60,
			PickaxeRequired: "mithril_pickaxe",
			MaxQuantity:     10,
			SortOrder:       15,
		},
	}

	// Only create if doesn't exist — preserve user edits on restart
	for _, ore := range ores {
		var existing OreType
		result := DB.Where("ore_key = ?", ore.OreKey).First(&existing)
		if result.Error != nil {
			// Not found — create
			DB.Create(&ore)
		}
		// If found, do nothing — preserve user edits from admin panel
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

// seedCraftableItems seeds the recipe master table
func seedCraftableItems() error {
	recipes := []CraftableItem{
		{
			Name:           "Copper Ingot",
			Description:    "A basic ingot for crafting",
			Icon:           "🪨",
			ItemKey:        "copper_ingot",
			OutputType:     "ingot",
			CraftingTimeMS: 5000,
			XPPerCraft:     20,
			LevelRequired:  1,
			MaxQuantity:    500,
			SortOrder:      1,
		},
		{
			Name:           "Iron Ingot",
			Description:    "A strong ingot for better tools",
			Icon:           "⚫",
			ItemKey:        "iron_ingot",
			OutputType:     "ingot",
			CraftingTimeMS: 8000,
			XPPerCraft:     40,
			LevelRequired:  5,
			MaxQuantity:    300,
			SortOrder:      2,
		},
		{
			Name:           "Gold Ingot",
			Description:    "A precious ingot",
			Icon:           "✨",
			ItemKey:        "gold_ingot",
			OutputType:     "ingot",
			CraftingTimeMS: 12000,
			XPPerCraft:     60,
			LevelRequired:  15,
			MaxQuantity:    100,
			SortOrder:      3,
		},
		{
			Name:           "Mithril Ingot",
			Description:    "A legendary ingot",
			Icon:           "💎",
			ItemKey:        "mithril_ingot",
			OutputType:     "ingot",
			CraftingTimeMS: 20000,
			XPPerCraft:     100,
			LevelRequired:  30,
			MaxQuantity:    50,
			SortOrder:      4,
		},
		{
			Name:           "Diamond Ingot",
			Description:    "The rarest ingot",
			Icon:           "💠",
			ItemKey:        "diamond_ingot",
			OutputType:     "ingot",
			CraftingTimeMS: 30000,
			XPPerCraft:     150,
			LevelRequired:  50,
			MaxQuantity:    25,
			SortOrder:      5,
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
			}
		}
		// If found, do nothing — preserve user edits from admin panel
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
			ModifiersJSON: "[]", LevelRequired: 1, SortOrder: 1,
		},
		{
			EquipmentKey: "combat_pistol",
			Name: "Combat Pistol", Icon: "🔫",
			Description: "Standard-issue sidearm. Reliable even after years of neglect.",
			Slot: "weapon", Rarity: "uncommon",
			BaseAttack: 12, AttackType: "physical",
			ModifiersJSON: `[{"type":"dex","value":2}]`, LevelRequired: 5, SortOrder: 2,
		},
		{
			EquipmentKey: "incendiary_launcher",
			Name: "Incendiary Launcher", Icon: "🔥",
			Description: "Fires canisters of burning chemical gel. Leaves nothing but ash.",
			Slot: "weapon", Rarity: "rare",
			BaseAttack: 20, AttackType: "fire",
			ModifiersJSON: `[{"type":"resist_fire","value":10}]`, LevelRequired: 15, SortOrder: 3,
		},
		{
			EquipmentKey: "tesla_coil_gun",
			Name: "Tesla Coil Gun", Icon: "⚡",
			Description: "Repurposed power-grid tech. Arcs through multiple targets.",
			Slot: "weapon", Rarity: "epic",
			BaseAttack: 30, AttackType: "lightning",
			ModifiersJSON: `[{"type":"dex","value":3},{"type":"int","value":3}]`, LevelRequired: 30, SortOrder: 4,
		},
		{
			EquipmentKey: "venom_blade",
			Name: "Venom Blade", Icon: "🐍",
			Description: "Coated in synthesised toxin. Each cut festers.",
			Slot: "weapon", Rarity: "rare",
			BaseAttack: 18, AttackType: "poison",
			ModifiersJSON: `[{"type":"dex","value":5}]`, LevelRequired: 20, SortOrder: 5,
		},
		// ── HEAD ─────────────────────────────────────────────────────────────
		{
			EquipmentKey: "scrap_helmet",
			Name: "Scrap Helmet", Icon: "⛑️",
			Description: "Welded together from vehicle panels. Crude but effective.",
			Slot: "head", Rarity: "common",
			BaseDefence: 3,
			ModifiersJSON: "[]", LevelRequired: 1, SortOrder: 10,
		},
		{
			EquipmentKey: "military_helmet",
			Name: "Military Helmet", Icon: "🪖",
			Description: "Pre-war composite helmet. Still rated for combat.",
			Slot: "head", Rarity: "uncommon",
			BaseDefence: 6,
			ModifiersJSON: `[{"type":"resist_fire","value":10}]`, LevelRequired: 5, SortOrder: 11,
		},
		{
			EquipmentKey: "hazmat_hood",
			Name: "Hazmat Hood", Icon: "😷",
			Description: "Full-face chemical protection. Filters airborne toxins.",
			Slot: "head", Rarity: "rare",
			BaseDefence: 4,
			ModifiersJSON: `[{"type":"resist_poison","value":30}]`, LevelRequired: 15, SortOrder: 12,
		},
		// ── CHEST ────────────────────────────────────────────────────────────
		{
			EquipmentKey: "tattered_vest",
			Name: "Tattered Vest", Icon: "🧥",
			Description: "Strips of leather sewn over a damaged flak jacket.",
			Slot: "chest", Rarity: "common",
			BaseDefence: 4,
			ModifiersJSON: "[]", LevelRequired: 1, SortOrder: 20,
		},
		{
			EquipmentKey: "kevlar_vest",
			Name: "Kevlar Vest", Icon: "🦺",
			Description: "Multi-layer ballistic weave. Stops fragments and pistol rounds.",
			Slot: "chest", Rarity: "uncommon",
			BaseDefence: 10,
			ModifiersJSON: `[{"type":"resist_fire","value":15}]`, LevelRequired: 8, SortOrder: 21,
		},
		{
			EquipmentKey: "nbc_suit",
			Name: "NBC Suit", Icon: "☢️",
			Description: "Nuclear-Biological-Chemical rated full-body suit. Invaluable in contaminated zones.",
			Slot: "chest", Rarity: "rare",
			BaseDefence: 8,
			ModifiersJSON: `[{"type":"resist_poison","value":40},{"type":"resist_chaos","value":20}]`, LevelRequired: 20, SortOrder: 22,
		},
		// ── LEGS ─────────────────────────────────────────────────────────────
		{
			EquipmentKey: "scrap_leggings",
			Name: "Scrap Leggings", Icon: "🩲",
			Description: "Sheet metal strapped to canvas. Uncomfortable but protective.",
			Slot: "legs", Rarity: "common",
			BaseDefence: 3,
			ModifiersJSON: "[]", LevelRequired: 1, SortOrder: 30,
		},
		{
			EquipmentKey: "combat_pants",
			Name: "Combat Pants", Icon: "👖",
			Description: "Reinforced tactical trousers with knee guards.",
			Slot: "legs", Rarity: "uncommon",
			BaseDefence: 7,
			ModifiersJSON: `[{"type":"dex","value":2}]`, LevelRequired: 5, SortOrder: 31,
		},
		// ── SHIELD ───────────────────────────────────────────────────────────
		{
			EquipmentKey: "scrap_shield",
			Name: "Scrap Shield", Icon: "🛡️",
			Description: "A car door repurposed as a shield. Heavy but solid.",
			Slot: "shield", Rarity: "common",
			BaseDefence: 5,
			ModifiersJSON: "[]", LevelRequired: 1, SortOrder: 40,
		},
		{
			EquipmentKey: "ballistic_shield",
			Name: "Ballistic Shield", Icon: "🔰",
			Description: "Police-grade riot shield. Rated for high-velocity impacts.",
			Slot: "shield", Rarity: "uncommon",
			BaseDefence: 12,
			ModifiersJSON: `[{"type":"resist_fire","value":10}]`, LevelRequired: 10, SortOrder: 41,
		},
		// ── RINGS ────────────────────────────────────────────────────────────
		{
			EquipmentKey: "strength_band",
			Name: "Strength Band", Icon: "💪",
			Description: "A weighted training band that permanently enhances muscle output.",
			Slot: "ring", Rarity: "common",
			ModifiersJSON: `[{"type":"str","value":3}]`, LevelRequired: 1, SortOrder: 50,
		},
		{
			EquipmentKey: "toxin_ring",
			Name: "Toxin Ring", Icon: "💍",
			Description: "Contains a slow-release antitoxin compound. Grants poison resistance.",
			Slot: "ring", Rarity: "uncommon",
			ModifiersJSON: `[{"type":"resist_poison","value":20}]`, LevelRequired: 5, SortOrder: 51,
		},
		{
			EquipmentKey: "lightning_ward",
			Name: "Lightning Ward", Icon: "⚡",
			Description: "A Faraday-cage ring that dissipates electrical energy.",
			Slot: "ring", Rarity: "rare",
			ModifiersJSON: `[{"type":"resist_lightning","value":25}]`, LevelRequired: 15, SortOrder: 52,
		},
		// ── AMULET ───────────────────────────────────────────────────────────
		{
			EquipmentKey: "dog_tag_amulet",
			Name: "Dog Tag Amulet", Icon: "🪪",
			Description: "The tags of a fallen ally. Wearing them sharpens your edge.",
			Slot: "amulet", Rarity: "common",
			ModifiersJSON: `[{"type":"str","value":2},{"type":"dex","value":2}]`, LevelRequired: 1, SortOrder: 60,
		},
		{
			EquipmentKey: "commanders_amulet",
			Name: "Commander's Amulet", Icon: "🎖️",
			Description: "Recovered from a high-ranking officer. Radiates authority and power.",
			Slot: "amulet", Rarity: "epic",
			ModifiersJSON: `[{"type":"str","value":5},{"type":"int","value":5},{"type":"dex","value":5}]`, LevelRequired: 40, SortOrder: 61,
		},
	}

	for _, item := range items {
		var existing Equipment
		result := DB.Where("equipment_key = ?", item.EquipmentKey).First(&existing)
		if result.Error != nil {
			// Equipment doesn't exist, create it
			DB.Create(&item)
		}
		// If it exists, skip it (preserve any admin modifications)
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
