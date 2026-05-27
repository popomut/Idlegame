package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"idlegame-backend/database"
	"time"
)

// StartMiningRequest contains ore selection
type StartMiningRequest struct {
	OreID uint `json:"ore_id"`
}

// MiningStatusResponse returns mining progress and offline gains
type MiningStatusResponse struct {
	IsActive     bool             `json:"is_active"`
	CurrentOre   *OreTypeResponse `json:"current_ore,omitempty"`
	StartedAt    time.Time        `json:"started_at,omitempty"`
	OfflineGains OfflineGainsInfo `json:"offline_gains,omitempty"`
	CurrentOres  map[string]int   `json:"current_ores"`
	CurrentHerbs map[string]int   `json:"current_herbs"`
}

type OreTypeResponse struct {
	ID              uint   `json:"id"`
	OreKey          string `json:"ore_key"`
	OreName         string `json:"ore_name"`
	Icon            string `json:"icon"`
	Difficulty      string `json:"difficulty"`
	MiningTimeMS    int    `json:"mining_time_ms"`
	PickaxeRequired string `json:"pickaxe_required"`
	MaxQuantity     int    `json:"max_quantity"`
	ResourceType    string `json:"resource_type"` // "ore" or "herb"
}

type OfflineGainsInfo struct {
	WasOffline     bool      `json:"was_offline"`
	OfflineTime    int64     `json:"offline_time_ms"`
	OresGained     int       `json:"ores_gained"`
	OreName        string    `json:"ore_name"`
	LastActiveTime time.Time `json:"last_active_time"`
}

// StartMining begins mining/gathering session (ore or herb)
func StartMining(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var req struct {
		OreID        uint   `json:"ore_id"`
		ResourceType string `json:"resource_type"` // "ore" or "herb"
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	isHerb := req.ResourceType == "herb"

	var ore database.OreType
	var herb database.HerbType

	if isHerb {
		if err := database.DB.First(&herb, req.OreID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "herb not found"})
		}
	} else {
		if err := database.DB.First(&ore, req.OreID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "ore not found"})
		}
	}

	// Get resource name & level requirement
	var resourceName string
	var levelRequired int
	if isHerb {
		resourceName = herb.HerbName
		levelRequired = herb.LevelRequired
	} else {
		resourceName = ore.OreName
		levelRequired = ore.LevelRequired
	}

	// Check extraction level requirement
	var skill database.UserMiningSkill
	database.DB.Where("user_id = ?", userID).First(&skill)
	skillLevel := skill.Level
	if skillLevel == 0 {
		skillLevel = 1
	}
	if skillLevel < levelRequired {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": fmt.Sprintf("extraction level %d required", levelRequired),
		})
	}

	// Stop any existing active session first
	var existingSession database.MiningSession
	database.DB.Where("user_id = ? AND status = ?", userID, "active").First(&existingSession)
	if existingSession.ID != 0 {
		CalculateAndSaveResourceGains(userID, existingSession)
	}

	// Create new session
	session := database.MiningSession{
		UserID:    userID,
		StartedAt: time.Now().UTC(),
		Status:    "active",
	}

	if isHerb {
		session.HerbID = herb.ID
		session.OreID = 0
	} else {
		session.OreID = ore.ID
		session.HerbID = 0
	}

	if err := database.DB.Create(&session).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to start extraction"})
	}

	database.LogActivity(userID, "Started extracting "+resourceName)

	return c.JSON(fiber.Map{
		"status":     "extraction started",
		"session_id": session.ID,
		"ore_name":   resourceName,
		"started_at": session.StartedAt,
	})
}

// StopMining stops the current extraction session (ore or herb)
func StopMining(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var session database.MiningSession
	result := database.DB.Where("user_id = ? AND status = ?", userID, "active").First(&session)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no active extraction session"})
	}

	resourceGained := CalculateAndSaveResourceGains(userID, session)

	// Get resource name + type for log
	var resourceName string
	resourceType := "ore"
	if session.HerbID != 0 {
		var herb database.HerbType
		database.DB.First(&herb, session.HerbID)
		resourceName = herb.HerbName
		resourceType = "herb"
	} else {
		var ore database.OreType
		database.DB.First(&ore, session.OreID)
		resourceName = ore.OreName
	}

	database.LogActivity(userID, fmt.Sprintf("Stopped extracting %s. Gained %d units.", resourceName, resourceGained))

	return c.JSON(fiber.Map{
		"status":        "extraction stopped",
		"ores_gained":   resourceGained,
		"resource_type": resourceType,
	})
}

// GetMiningStatus returns current mining status and offline gains
func GetMiningStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var session database.MiningSession
	isActive := false
	var ore database.OreType
	var herb database.HerbType
	isHerbSession := false

	result := database.DB.Where("user_id = ? AND status = ?", userID, "active").First(&session)
	if result.Error == nil {
		isActive = true
		if session.HerbID != 0 {
			isHerbSession = true
			database.DB.First(&herb, session.HerbID)
		} else {
			database.DB.Preload("OreType").First(&session, session.ID)
			ore = session.OreType
		}
	}

	// Build current ore quantities dynamically from pivot table
	var items []database.OreInventoryItem
	database.DB.Where("user_id = ?", userID).Preload("OreType").Find(&items)

	currentOres := make(map[string]int)
	for _, item := range items {
		currentOres[item.OreType.OreKey] = item.Quantity
	}

	// Build current herb quantities from herb pivot table
	var herbItems []database.HerbInventoryItem
	database.DB.Where("user_id = ?", userID).Preload("HerbType").Find(&herbItems)

	currentHerbs := make(map[string]int)
	for _, item := range herbItems {
		currentHerbs[item.HerbType.HerbKey] = item.Quantity
	}

	// Add pending (unsaved) resources for OFFLINE GAINS ONLY
	// currentOres/currentHerbs should return ACTUAL saved values only
	// Frontend tracks live pending locally via client-side counter
	if isActive {
		now := time.Now().UTC()
		elapsed := now.Sub(session.StartedAt)

		// Calculate pending for OFFLINE GAINS display, but DON'T add to currentOres
		if isHerbSession {
			pendingHerbs := 0
			if herb.GatherTimeMS > 0 {
				pendingHerbs = int(elapsed.Milliseconds()) / herb.GatherTimeMS
			}
			if herb.MaxQuantity > 0 {
				existing := currentHerbs[herb.HerbKey]
				remaining := herb.MaxQuantity - existing
				if pendingHerbs > remaining {
					pendingHerbs = remaining
				}
				if pendingHerbs < 0 {
					pendingHerbs = 0
				}
			}
			// NOTE: Don't add to currentHerbs — frontend tracks locally
		} else {
			pendingOres := 0
			if ore.MiningTimeMS > 0 {
				pendingOres = int(elapsed.Milliseconds()) / ore.MiningTimeMS
			}
			if ore.MaxQuantity > 0 {
				existing := currentOres[ore.OreKey]
				remaining := ore.MaxQuantity - existing
				if pendingOres > remaining {
					pendingOres = remaining
				}
				if pendingOres < 0 {
					pendingOres = 0
				}
			}
			// NOTE: Don't add to currentOres — frontend tracks locally
		}
	}

	response := MiningStatusResponse{
		IsActive:     isActive,
		CurrentOres:  currentOres,
		CurrentHerbs: currentHerbs,
	}

	if isActive {
		if isHerbSession {
			response.CurrentOre = &OreTypeResponse{
				ID:           herb.ID,
				OreKey:       herb.HerbKey,
				OreName:      herb.HerbName,
				Icon:         herb.Icon,
				Difficulty:   herb.Difficulty,
				MiningTimeMS: herb.GatherTimeMS,
				MaxQuantity:  herb.MaxQuantity,
				ResourceType: "herb",
			}
		} else {
			response.CurrentOre = &OreTypeResponse{
				ID:              ore.ID,
				OreKey:          ore.OreKey,
				OreName:         ore.OreName,
				Icon:            ore.Icon,
				Difficulty:      ore.Difficulty,
				MiningTimeMS:    ore.MiningTimeMS,
				PickaxeRequired: ore.PickaxeRequired,
				MaxQuantity:     ore.MaxQuantity,
				ResourceType:    "ore",
			}
		}
		response.StartedAt = session.StartedAt
		response.OfflineGains = CalculateOfflineGains(userID, session)
	}

	return c.JSON(response)
}

// CalculateAndSaveResourceGains handles both ore mining + herb gathering
// Returns the number of resources earned
func CalculateAndSaveResourceGains(userID uint, session database.MiningSession) int {
	// Atomically claim session — prevents double-award on concurrent stop calls
	claimed := database.DB.Exec(
		"UPDATE mining_sessions SET status = 'completed', ended_at = ? WHERE id = ? AND status = 'active'",
		time.Now().UTC(), session.ID,
	)
	if claimed.RowsAffected == 0 {
		return 0
	}

	now := time.Now().UTC()
	elapsed := now.Sub(session.StartedAt)
	resourceEarned := 0
	
	// Check if ore or herb
	if session.HerbID != 0 {
		// Herb gathering
		var herb database.HerbType
		database.DB.First(&herb, session.HerbID)
		
		if herb.GatherTimeMS > 0 {
			resourceEarned = int(elapsed.Milliseconds()) / herb.GatherTimeMS
		}
		if resourceEarned == 0 {
			return 0
		}

		// Find or create pivot row
		var item database.HerbInventoryItem
		database.DB.FirstOrCreate(&item, database.HerbInventoryItem{UserID: userID, HerbTypeID: herb.ID})

		// Apply max quantity cap
		if herb.MaxQuantity > 0 {
			remaining := herb.MaxQuantity - item.Quantity
			if remaining <= 0 {
				return 0
			}
			if resourceEarned > remaining {
				resourceEarned = remaining
			}
		}

		// Atomic increment
		database.DB.Exec(
			"UPDATE herb_inventory_items SET quantity = quantity + ?, updated_at = ? WHERE id = ?",
			resourceEarned, time.Now().UTC(), item.ID,
		)

		// Award gathering XP
		xpGained := herb.XPPerHerb * resourceEarned
		AwardMiningXP(userID, xpGained)

		database.LogActivity(userID, fmt.Sprintf("Gathered %d %s", resourceEarned, herb.HerbName))
		
	} else if session.OreID != 0 {
		// Ore mining
		var ore database.OreType
		database.DB.First(&ore, session.OreID)

		if ore.MiningTimeMS > 0 {
			resourceEarned = int(elapsed.Milliseconds()) / ore.MiningTimeMS
		}
		if resourceEarned == 0 {
			return 0
		}

		// Find or create pivot row
		var item database.OreInventoryItem
		database.DB.FirstOrCreate(&item, database.OreInventoryItem{UserID: userID, OreTypeID: ore.ID})

		// Apply max quantity cap
		if ore.MaxQuantity > 0 {
			remaining := ore.MaxQuantity - item.Quantity
			if remaining <= 0 {
				return 0
			}
			if resourceEarned > remaining {
				resourceEarned = remaining
			}
		}

		// Atomic increment
		database.DB.Exec(
			"UPDATE ore_inventory_items SET quantity = quantity + ?, updated_at = ? WHERE id = ?",
			resourceEarned, time.Now().UTC(), item.ID,
		)

		// Award mining XP
		xpGained := ore.XPPerOre * resourceEarned
		AwardMiningXP(userID, xpGained)

		database.LogActivity(userID, fmt.Sprintf("Extracted %d %s", resourceEarned, ore.OreName))
	}

	return resourceEarned
}

// CalculateAndSaveOreGains kept for backward compat — calls generic version
func CalculateAndSaveOreGains(userID uint, session database.MiningSession) int {
	return CalculateAndSaveResourceGains(userID, session)
}

// CalculateOfflineGains determines what the player earned while offline (ore or herb)
func CalculateOfflineGains(userID uint, session database.MiningSession) OfflineGainsInfo {
	gains := OfflineGainsInfo{WasOffline: false}

	now := time.Now().UTC()
	elapsed := now.Sub(session.StartedAt)

	if session.HerbID != 0 {
		var herb database.HerbType
		database.DB.First(&herb, session.HerbID)
		gains.OreName = herb.HerbName
		gatherTime := time.Duration(herb.GatherTimeMS) * time.Millisecond
		if elapsed > gatherTime && herb.GatherTimeMS > 0 {
			gains.WasOffline = true
			gains.OfflineTime = elapsed.Milliseconds()
			gains.OresGained = int(elapsed.Milliseconds()) / herb.GatherTimeMS
			gains.LastActiveTime = session.StartedAt
		}
	} else {
		var ore database.OreType
		database.DB.First(&ore, session.OreID)
		gains.OreName = ore.OreName
		miningTimePerOre := time.Duration(ore.MiningTimeMS) * time.Millisecond
		if elapsed > miningTimePerOre && ore.MiningTimeMS > 0 {
			gains.WasOffline = true
			gains.OfflineTime = elapsed.Milliseconds()
			gains.OresGained = int(elapsed.Milliseconds()) / ore.MiningTimeMS
			gains.LastActiveTime = session.StartedAt
		}
	}

	return gains
}
