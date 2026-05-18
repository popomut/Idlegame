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

// MiningLevel defines XP progression for mining skill — completely separate from combat
type MiningLevel struct {
	Level       int `gorm:"primaryKey"` // 1, 2, 3, ...
	XPRequired  int `gorm:"not null"` // total XP needed to reach this level
}

// UserMiningSkill tracks player's mining progression — one row per user
type UserMiningSkill struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"uniqueIndex;not null"`
	User      User      `gorm:"foreignKey:UserID"`
	Level     int       `gorm:"default:1"`
	XP        int       `gorm:"default:0"`
	UpdatedAt time.Time
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
	Status      string `gorm:"default:'active';index;not null"`
	
	CreatedAt   time.Time
}

// Continent defines a top-level map region. Add rows to extend — no code changes needed.
type Continent struct {
	ID           uint   `gorm:"primaryKey"`
	ContinentKey string `gorm:"uniqueIndex;not null"` // e.g. "scorched_wastes"
	Name         string `gorm:"not null"`
	Icon         string
	Description  string
	Difficulty   string `gorm:"default:'easy'"` // easy, medium, hard, extreme
	SortOrder    int    `gorm:"default:0"`
	CreatedAt    time.Time
}

// Area is a zone within a Continent. Monsters spawn here; a boss appears after FightsBeforeBoss kills.
type Area struct {
	ID               uint      `gorm:"primaryKey"`
	AreaKey          string    `gorm:"uniqueIndex;not null"` // e.g. "dusty_outpost"
	ContinentID      uint      `gorm:"not null;index"`
	Continent        Continent `gorm:"foreignKey:ContinentID"`
	Name             string    `gorm:"not null"`
	Icon             string
	Description      string
	Difficulty       string `gorm:"default:'easy'"`
	FightsBeforeBoss int    `gorm:"default:5"`
	BossMonsterKey   string `gorm:"not null"` // references Monster.MonsterKey
	SortOrder        int    `gorm:"default:0"`
	CreatedAt        time.Time
}

// AreaMonster links monsters to areas with a spawn weight.
// Higher Weight = appears more often. Add rows to extend — no code changes needed.
type AreaMonster struct {
	ID         uint   `gorm:"primaryKey"`
	AreaID     uint   `gorm:"not null;uniqueIndex:idx_area_monster"`
	MonsterKey string `gorm:"not null;uniqueIndex:idx_area_monster"`
	Weight     int    `gorm:"default:1"` // relative spawn probability
}

// CombatSession tracks an active fight sequence for a player.
// One row per user (unique on UserID). Entering a new area replaces the current session.
type CombatSession struct {
	ID                uint   `gorm:"primaryKey"`
	UserID            uint   `gorm:"not null;uniqueIndex"`
	AreaKey           string `gorm:"not null"`
	FightCount        int    `gorm:"default:0"`
	FightsBeforeBoss  int    `gorm:"default:5"`
	Status            string `gorm:"default:'fighting'"` // fighting, boss, complete
	CurrentMonsterKey string

	StartedAt time.Time
	UpdatedAt time.Time
}

// Equipment defines a piece of gear in the master table.
// Slot values: head, chest, legs, weapon, shield, ring, amulet
// Rarity values: common, uncommon, rare, epic, legendary
// AttackType: physical, fire, lightning, ice, poison, chaos
// ModifiersJSON: JSON array of {type, value} e.g. [{"type":"str","value":5}]
//   Modifier types: str, int, dex, resist_fire, resist_lightning, resist_ice, resist_poison, resist_chaos
// Adding a new equipment: INSERT a row here — no frontend/backend code changes needed.
type Equipment struct {
	ID           uint   `gorm:"primaryKey"`
	EquipmentKey string `gorm:"uniqueIndex;not null"` // e.g. "rusty_blade"
	Name         string `gorm:"not null"`
	Icon         string
	Description  string
	Slot         string `gorm:"not null"` // head, chest, legs, weapon, shield, ring, amulet
	Rarity       string `gorm:"default:'common'"`

	// Primary combat stat (weapon: use BaseAttack+AttackType; armor: use BaseDefence)
	BaseAttack  int    `gorm:"default:0"`
	AttackType  string `gorm:"default:'physical'"`
	BaseDefence int    `gorm:"default:0"`

	// Secondary modifiers as JSON string
	ModifiersJSON string `gorm:"default:'[]'"`

	LevelRequired int `gorm:"default:1"`
	SortOrder     int `gorm:"default:0"`

	CreatedAt time.Time
}

// UserEquipment is the equipment bag — one row per piece of gear a user has obtained.
type UserEquipment struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"not null;index"`
	EquipmentID uint      `gorm:"not null"`
	Equipment   Equipment `gorm:"foreignKey:EquipmentID"`
	ObtainedAt  time.Time
}

// UserEquippedSlot stores what the user is currently wearing in each slot.
// Slot values: head, chest, legs, weapon, shield, ring1, ring2, amulet
// UserEquipmentID = 0 means the slot is empty.
type UserEquippedSlot struct {
	ID             uint   `gorm:"primaryKey"`
	UserID         uint   `gorm:"uniqueIndex:idx_user_slot;not null"`
	Slot           string `gorm:"uniqueIndex:idx_user_slot;not null"`
	UserEquipmentID uint  `gorm:"default:0"` // references UserEquipment.ID; 0 = empty
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

// ActiveCombat tracks the current real-time combat session for a player.
// Server is authoritative — all damage and HP are calculated here.
// One row per user (unique). Previous sessions are overwritten on new combat start.
type ActiveCombat struct {
	ID     uint   `gorm:"primaryKey"`
	UserID uint   `gorm:"uniqueIndex"`
	Status string `gorm:"default:'active'"` // active, fled, dead

	// Which zone the player entered (used to pick zone-specific monsters)
	ZoneKey string

	// Current enemy state — persisted so fight continues correctly on resume
	CurrentEnemyKey   string
	CurrentEnemyHP    int
	CurrentEnemyMaxHP int

	// Player HP during this combat session (separate from User.HP which is out-of-combat)
	PlayerHPCurrent int

	// Session totals (for display and reward awarding)
	EnemiesDefeated  int   `gorm:"default:0"`
	TotalXPGained    int64 `gorm:"default:0"`
	TotalMoneyGained int64 `gorm:"default:0"`

	// Tracks how much has already been awarded to user (prevents double-awarding)
	XPAwarded    int64 `gorm:"default:0"`
	MoneyAwarded int64 `gorm:"default:0"`

	// Combat log — rolling window of last 50 entries (JSON array)
	CombatLogsJSON string `gorm:"type:text;default:'[]'"`

	// Timing — LastTickAt is the key field for calculating offline progress
	StartedAt  time.Time
	LastTickAt time.Time
}
