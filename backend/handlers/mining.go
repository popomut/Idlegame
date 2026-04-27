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
}

type OfflineGainsInfo struct {
	WasOffline     bool      `json:"was_offline"`
	OfflineTime    int64     `json:"offline_time_ms"`
	OresGained     int       `json:"ores_gained"`
	OreName        string    `json:"ore_name"`
	LastActiveTime time.Time `json:"last_active_time"`
}

// StartMining begins a mining session
func StartMining(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	req := new(StartMiningRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	var ore database.OreType
	if err := database.DB.First(&ore, req.OreID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "ore not found"})
	}

	// Stop any existing active session first
	var existingSession database.MiningSession
	database.DB.Where("user_id = ? AND status = ?", userID, "active").First(&existingSession)
	if existingSession.ID != 0 {
		CalculateAndSaveOreGains(userID, existingSession)
		database.DB.Model(&existingSession).Update("status", "completed")
	}

	session := database.MiningSession{
		UserID:    userID,
		OreID:     ore.ID,
		StartedAt: time.Now().UTC(),
		Status:    "active",
	}
	if err := database.DB.Create(&session).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to start mining"})
	}

	database.LogActivity(userID, "Started extracting "+ore.OreName)

	return c.JSON(fiber.Map{
		"status":     "mining started",
		"session_id": session.ID,
		"ore_name":   ore.OreName,
		"started_at": session.StartedAt,
	})
}

// StopMining stops the current mining session
func StopMining(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var session database.MiningSession
	result := database.DB.Where("user_id = ? AND status = ?", userID, "active").First(&session)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no active mining session"})
	}

	oresGained := CalculateAndSaveOreGains(userID, session)

	now := time.Now().UTC()
	database.DB.Model(&session).Updates(map[string]interface{}{
		"status":   "completed",
		"ended_at": now,
	})

	var ore database.OreType
	database.DB.First(&ore, session.OreID)

	database.LogActivity(userID, fmt.Sprintf("Stopped extracting %s. Gained %d units.", ore.OreName, oresGained))

	return c.JSON(fiber.Map{
		"status":     "mining stopped",
		"ores_gained": oresGained,
	})
}

// GetMiningStatus returns current mining status and offline gains
func GetMiningStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var session database.MiningSession
	isActive := false
	var ore database.OreType

	result := database.DB.Where("user_id = ? AND status = ?", userID, "active").Preload("OreType").First(&session)
	if result.Error == nil {
		isActive = true
		ore = session.OreType
	}

	// Build current ore quantities dynamically from pivot table
	var items []database.OreInventoryItem
	database.DB.Where("user_id = ?", userID).Preload("OreType").Find(&items)

	currentOres := make(map[string]int)
	for _, item := range items {
		currentOres[item.OreType.OreKey] = item.Quantity
	}

	// Add pending (unsaved) ores for the active session
	if isActive {
		now := time.Now().UTC()
		elapsed := now.Sub(session.StartedAt)
		pendingOres := int(elapsed.Milliseconds()) / ore.MiningTimeMS

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
		currentOres[ore.OreKey] += pendingOres
	}

	response := MiningStatusResponse{
		IsActive:    isActive,
		CurrentOres: currentOres,
	}

	if isActive {
		response.CurrentOre = &OreTypeResponse{
			ID:              ore.ID,
			OreKey:          ore.OreKey,
			OreName:         ore.OreName,
			Icon:            ore.Icon,
			Difficulty:      ore.Difficulty,
			MiningTimeMS:    ore.MiningTimeMS,
			PickaxeRequired: ore.PickaxeRequired,
			MaxQuantity:     ore.MaxQuantity,
		}
		response.StartedAt = session.StartedAt
		response.OfflineGains = CalculateOfflineGains(userID, session)
	}

	return c.JSON(response)
}

// CalculateAndSaveOreGains calculates earned ores using server timestamps and saves to pivot table.
// Returns the number of ores earned.
func CalculateAndSaveOreGains(userID uint, session database.MiningSession) int {
	now := time.Now().UTC()
	elapsed := now.Sub(session.StartedAt)

	var ore database.OreType
	database.DB.First(&ore, session.OreID)

	oresEarned := int(elapsed.Milliseconds()) / ore.MiningTimeMS
	if oresEarned == 0 {
		return 0
	}

	// Find or create the pivot row
	var item database.OreInventoryItem
	database.DB.FirstOrCreate(&item, database.OreInventoryItem{UserID: userID, OreTypeID: ore.ID})

	// Apply max quantity cap
	if ore.MaxQuantity > 0 {
		remaining := ore.MaxQuantity - item.Quantity
		if remaining <= 0 {
			return 0
		}
		if oresEarned > remaining {
			oresEarned = remaining
		}
	}

	// Atomic increment
	database.DB.Exec(
		"UPDATE ore_inventory_items SET quantity = quantity + ?, updated_at = ? WHERE id = ?",
		oresEarned, time.Now().UTC(), item.ID,
	)

	database.LogActivity(userID, fmt.Sprintf("Extracted %d %s", oresEarned, ore.OreName))

	return oresEarned
}

// CalculateOfflineGains determines what the player earned while offline
func CalculateOfflineGains(userID uint, session database.MiningSession) OfflineGainsInfo {
	gains := OfflineGainsInfo{WasOffline: false}

	var ore database.OreType
	database.DB.First(&ore, session.OreID)
	gains.OreName = ore.OreName

	now := time.Now().UTC()
	elapsed := now.Sub(session.StartedAt)
	miningTimePerOre := time.Duration(ore.MiningTimeMS) * time.Millisecond

	if elapsed > miningTimePerOre {
		gains.WasOffline = true
		gains.OfflineTime = elapsed.Milliseconds()
		gains.OresGained = int(elapsed.Milliseconds()) / ore.MiningTimeMS
		gains.LastActiveTime = session.StartedAt
	}

	return gains
}
