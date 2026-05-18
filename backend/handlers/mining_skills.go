package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"idlegame-backend/database"
)

// MiningSkillResponse represents user's mining progression
type MiningSkillResponse struct {
	Level      int `json:"level"`
	XP         int `json:"xp"`
	XPRequired int `json:"xp_required"`
	XPProgress int `json:"xp_progress"` // XP earned towards next level
}

// AwardMiningXP grants XP to player's mining skill and handles level-ups
// Similar to AwardXP in character.go but for mining skill
func AwardMiningXP(userID uint, xpAmount int) error {
	var skill database.UserMiningSkill
	if err := database.DB.First(&skill, "user_id = ?", userID).Error; err != nil {
		return err
	}

	// Add XP
	skill.XP += xpAmount
	skill.UpdatedAt = time.Now()

	// Check for level-ups (could level up multiple times in one big spike)
	for {
		var nextLevel database.MiningLevel
		if err := database.DB.First(&nextLevel, skill.Level+1).Error; err != nil {
			// Max level reached
			break
		}

		if skill.XP >= nextLevel.XPRequired {
			// Level up!
			skill.Level++
		} else {
			// Not enough XP for next level
			break
		}
	}

	// Save updated skill
	return database.DB.Save(&skill).Error
}

// GetMiningSkill returns the player's mining skill details
// GET /api/mining/skill
func GetMiningSkill(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var skill database.UserMiningSkill
	if err := database.DB.First(&skill, "user_id = ?", userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "mining skill not found"})
	}

	// Get current level's XP requirement (cumulative total)
	var currentLevel database.MiningLevel
	database.DB.First(&currentLevel, skill.Level)

	// Get next level's XP requirement
	var nextLevel database.MiningLevel
	var nextLevelXPRequired int
	if err := database.DB.First(&nextLevel, skill.Level+1).Error; err == nil {
		nextLevelXPRequired = nextLevel.XPRequired
	} else {
		// Max level, no next level
		nextLevelXPRequired = currentLevel.XPRequired
	}

	// Calculate progress towards next level
	progressTowards := skill.XP - currentLevel.XPRequired
	if progressTowards < 0 {
		progressTowards = 0
	}

	return c.JSON(MiningSkillResponse{
		Level:      skill.Level,
		XP:         skill.XP,
		XPRequired: nextLevelXPRequired - currentLevel.XPRequired, // XP needed for next level
		XPProgress: progressTowards,                               // XP earned towards next level
	})
}

// ── Admin endpoints (for development only) ──

// AdminGetMiningLevels returns all mining level configuration
// GET /api/admin/mining-levels
func AdminGetMiningLevels(c *fiber.Ctx) error {
	var levels []database.MiningLevel
	if err := database.DB.Order("level ASC").Find(&levels).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load mining levels"})
	}
	return c.JSON(levels)
}

// AdminCreateMiningLevel creates a new mining level entry
// POST /api/admin/mining-levels
func AdminCreateMiningLevel(c *fiber.Ctx) error {
	var body struct {
		Level      int `json:"level"`
		XPRequired int `json:"xp_required"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if body.Level < 1 {
		return c.Status(400).JSON(fiber.Map{"error": "Level must be >= 1"})
	}

	// Check if already exists
	var existing database.MiningLevel
	if err := database.DB.First(&existing, body.Level).Error; err == nil {
		return c.Status(409).JSON(fiber.Map{"error": "Level already exists"})
	}

	ml := database.MiningLevel{
		Level:      body.Level,
		XPRequired: body.XPRequired,
	}
	if err := database.DB.Create(&ml).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create mining level"})
	}

	return c.Status(201).JSON(ml)
}

// AdminUpdateMiningLevel updates a mining level entry
// PUT /api/admin/mining-levels/:level
func AdminUpdateMiningLevel(c *fiber.Ctx) error {
	levelStr := c.Params("level")
	var level int
	if _, err := fmt.Sscanf(levelStr, "%d", &level); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid level parameter"})
	}

	var body struct {
		XPRequired int `json:"xp_required"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var ml database.MiningLevel
	if err := database.DB.First(&ml, level).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Mining level not found"})
	}

	ml.XPRequired = body.XPRequired
	if err := database.DB.Save(&ml).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update mining level"})
	}

	return c.JSON(ml)
}

// AdminDeleteMiningLevel deletes a mining level entry
// DELETE /api/admin/mining-levels/:level
func AdminDeleteMiningLevel(c *fiber.Ctx) error {
	levelStr := c.Params("level")
	var level int
	if _, err := fmt.Sscanf(levelStr, "%d", &level); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid level parameter"})
	}

	var ml database.MiningLevel
	if err := database.DB.First(&ml, level).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Mining level not found"})
	}

	if err := database.DB.Delete(&ml).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete mining level"})
	}

	return c.JSON(fiber.Map{"success": true, "level": level})
}
