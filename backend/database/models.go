package database

import (
	"time"
)

// User represents a player account
type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"uniqueIndex;not null"`
	Email    string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	IsGuest  bool   `gorm:"default:false"`

	// Identity
	PlayerName  string `gorm:"default:'Operative'"`
	PlayerClass string `gorm:"default:'Recruit'"`

	// Progression
	Level int   `gorm:"default:1"`
	XP    int64 `gorm:"default:0"`

	// Base stats — updated on level-up from CharacterLevel table
	HP         int `gorm:"default:100"`
	MaxHP      int `gorm:"default:100"`
	Stamina    int `gorm:"default:50"`
	MaxStamina int `gorm:"default:50"`
	Str        int `gorm:"default:5"`
	Int        int `gorm:"default:5"`
	Dex        int `gorm:"default:5"`

	// Currency
	Money int64 `gorm:"default:0"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// CharacterLevel defines XP required and stat bonuses per level.
// Add rows to this table to extend the level cap — no code changes needed.
type CharacterLevel struct {
	Level       int `gorm:"primaryKey"` // 1, 2, 3, ...
	XPRequired  int64 `gorm:"not null"` // total XP needed to reach this level
	MaxHP       int `gorm:"default:100"`
	MaxStamina  int `gorm:"default:50"`
	Str         int `gorm:"default:5"`
	Int         int `gorm:"default:5"`
	Dex         int `gorm:"default:5"`
}

// OreInventory stores player's ore counts
type OreInventory struct {
	ID       uint   `gorm:"primaryKey"`
	UserID   uint   `gorm:"uniqueIndex;not null"`
	User     User   `gorm:"foreignKey:UserID"`
	
	CopperOre   int `gorm:"default:5"`
	IronOre     int `gorm:"default:2"`
	GoldOre     int `gorm:"default:0"`
	MithrilOre  int `gorm:"default:0"`
	DiamondOre  int `gorm:"default:0"`
	
	UpdatedAt time.Time
}

// OreType defines ore properties — add rows to this table to add new ore types with no code changes
type OreType struct {
	ID               uint   `gorm:"primaryKey"`
	OreKey           string `gorm:"uniqueIndex;not null"` // e.g. "copper_ore"
	OreName          string `gorm:"not null"`
	Icon             string
	Color            string
	Difficulty       string
	MiningTimeMS     int    `gorm:"default:3000"` // milliseconds per ore
	XPPerOre         int    `gorm:"default:10"`
	LevelRequired    int    `gorm:"default:1"`
	PickaxeRequired  string `gorm:"default:'none'"` // "none", "iron_pickaxe", "gold_pickaxe", "mithril_pickaxe"
	MaxQuantity      int    `gorm:"default:0"`      // 0 = unlimited
	SortOrder        int    `gorm:"default:0"`      // display order in UI

	CreatedAt time.Time
}

// OreInventoryItem is a pivot table: one row per (user, ore type) pair.
// Adding a new ore to OreType automatically works — no code changes needed.
type OreInventoryItem struct {
	ID        uint    `gorm:"primaryKey"`
	UserID    uint    `gorm:"not null;uniqueIndex:idx_inv_user_ore"`
	OreTypeID uint    `gorm:"not null;uniqueIndex:idx_inv_user_ore"`
	OreType   OreType `gorm:"foreignKey:OreTypeID"`
	Quantity  int     `gorm:"default:0"`

	UpdatedAt time.Time
}

// MiningSession tracks mining progress
type MiningSession struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index:idx_user_active,unique,where:status='active'"`
	User      User      `gorm:"foreignKey:UserID"`
	OreID     uint      `gorm:"not null"`
	OreType   OreType   `gorm:"foreignKey:OreID"`
	
	// Server-side timestamps (cannot be hacked)
	StartedAt   time.Time `gorm:"not null"`
	EndedAt     *time.Time
	
	// How many ores earned in this session
	OresMined   int    `gorm:"default:0"`
	
	// Session status: 'active', 'paused', 'completed'
	Status      string `gorm:"default:'active';index"`
	
	CreatedAt   time.Time
}

// Monster defines enemy stats — add rows to this table to add new monsters with no code changes.
// Attack types: physical, fire, lightning, ice, poison, chaos
// Resistances: 0 = no resistance, 100 = full immunity
type Monster struct {
	ID          uint   `gorm:"primaryKey"`
	MonsterKey  string `gorm:"uniqueIndex;not null"` // e.g. "wasteland_scavenger"
	Name        string `gorm:"not null"`
	Icon        string
	Description string

	// Combat stats
	HP          int    `gorm:"default:50"`
	DEX         int    `gorm:"default:1"`
	AttackType  string `gorm:"default:'physical'"` // physical, fire, lightning, ice, poison, chaos
	AttackValue int    `gorm:"default:5"`
	PhysDef     int    `gorm:"default:0"` // flat physical damage reduction

	// Elemental resistances (0–100 percent)
	ResistFire      int `gorm:"default:0"`
	ResistLightning int `gorm:"default:0"`
	ResistIce       int `gorm:"default:0"`
	ResistPoison    int `gorm:"default:0"`
	ResistChaos     int `gorm:"default:0"`

	// Rewards
	MoneyDropMin int `gorm:"default:0"`
	MoneyDropMax int `gorm:"default:0"`
	XPDrop       int `gorm:"default:5"`

	SortOrder int `gorm:"default:0"` // display/spawn order

	CreatedAt time.Time
}

// MonsterDrop defines loot table entries for a monster.
// DropType: "item" or "equipment". DropKey references the item/equipment master key.
// DropRate: 0.0 (never) to 1.0 (always).
type MonsterDrop struct {
	ID         uint    `gorm:"primaryKey"`
	MonsterID  uint    `gorm:"not null;index"`
	Monster    Monster `gorm:"foreignKey:MonsterID"`
	DropType   string  `gorm:"not null"` // "item" or "equipment"
	DropKey    string  `gorm:"not null"` // references item_key or equipment_key
	DropRate   float64 `gorm:"default:0.1"` // probability 0.0–1.0
	DropMin    int     `gorm:"default:1"`   // min quantity (for stackable items)
	DropMax    int     `gorm:"default:1"`   // max quantity

	CreatedAt time.Time
}

// ActivityLog stores player actions
type ActivityLog struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null;index"`
	User      User   `gorm:"foreignKey:UserID"`
	Message   string
	
	CreatedAt time.Time
}
