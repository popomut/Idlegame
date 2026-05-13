package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"idlegame-backend/database"
)

// MonsterDropManageResponse is the shape for admin CRUD operations.
type MonsterDropManageResponse struct {
	ID            uint    `json:"id"`
	MonsterID     uint    `json:"monster_id"`
	MonsterKey    string  `json:"monster_key"`
	MonsterName   string  `json:"monster_name"`
	EquipmentKey  string  `json:"equipment_key"`
	EquipmentName string  `json:"equipment_name"`
	DropRate      float64 `json:"drop_rate"`
	DropMin       int     `json:"drop_min"`
	DropMax       int     `json:"drop_max"`
}

// AdminGetAllMonsterDrops returns all monster drops with full details.
// GET /api/admin/monster-drops
func AdminGetAllMonsterDrops(c *fiber.Ctx) error {
	var drops []database.MonsterDrop
	if err := database.DB.
		Preload("Monster").
		Order("id ASC").
		Find(&drops).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load monster drops"})
	}

	result := make([]MonsterDropManageResponse, 0, len(drops))
	for _, drop := range drops {
		// Get equipment details by equipment_key
		var equip database.Equipment
		database.DB.Where("equipment_key = ?", drop.DropKey).First(&equip)

		result = append(result, MonsterDropManageResponse{
			ID:            drop.ID,
			MonsterID:     drop.MonsterID,
			MonsterKey:    drop.Monster.MonsterKey,
			MonsterName:   drop.Monster.Name,
			EquipmentKey:  drop.DropKey,
			EquipmentName: equip.Name,
			DropRate:      drop.DropRate,
			DropMin:       drop.DropMin,
			DropMax:       drop.DropMax,
		})
	}
	return c.JSON(result)
}

// AdminCreateMonsterDrop creates a new monster drop entry.
// POST /api/admin/monster-drops
func AdminCreateMonsterDrop(c *fiber.Ctx) error {
	var body struct {
		MonsterID    uint    `json:"monster_id"`
		EquipmentKey string  `json:"equipment_key"`
		DropRate     float64 `json:"drop_rate"`
		DropMin      int     `json:"drop_min"`
		DropMax      int     `json:"drop_max"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if body.MonsterID == 0 || body.EquipmentKey == "" {
		return c.Status(400).JSON(fiber.Map{"error": "monster_id and equipment_key are required"})
	}

	// Validate equipment exists
	var equip database.Equipment
	if err := database.DB.Where("equipment_key = ?", body.EquipmentKey).First(&equip).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Equipment not found"})
	}

	// Validate drop rate
	if body.DropRate < 0 || body.DropRate > 1 {
		return c.Status(400).JSON(fiber.Map{"error": "drop_rate must be between 0 and 1"})
	}

	drop := database.MonsterDrop{
		MonsterID: body.MonsterID,
		DropType:  "equipment",
		DropKey:   body.EquipmentKey,
		DropRate:  body.DropRate,
		DropMin:   body.DropMin,
		DropMax:   body.DropMax,
		CreatedAt: time.Now(),
	}

	if err := database.DB.Create(&drop).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create monster drop"})
	}

	// Fetch full details for response
	database.DB.Preload("Monster").First(&drop)

	return c.Status(201).JSON(MonsterDropManageResponse{
		ID:            drop.ID,
		MonsterID:     drop.MonsterID,
		MonsterKey:    drop.Monster.MonsterKey,
		MonsterName:   drop.Monster.Name,
		EquipmentKey:  drop.DropKey,
		EquipmentName: equip.Name,
		DropRate:      drop.DropRate,
		DropMin:       drop.DropMin,
		DropMax:       drop.DropMax,
	})
}

// AdminUpdateMonsterDrop updates an existing monster drop.
// PUT /api/admin/monster-drops/:id
func AdminUpdateMonsterDrop(c *fiber.Ctx) error {
	id := c.Params("id")

	var drop database.MonsterDrop
	if err := database.DB.First(&drop, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Monster drop not found"})
	}

	var body struct {
		MonsterID    uint    `json:"monster_id"`
		EquipmentKey string  `json:"equipment_key"`
		DropRate     float64 `json:"drop_rate"`
		DropMin      int     `json:"drop_min"`
		DropMax      int     `json:"drop_max"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate drop rate
	if body.DropRate < 0 || body.DropRate > 1 {
		return c.Status(400).JSON(fiber.Map{"error": "drop_rate must be between 0 and 1"})
	}

	// Update fields
	if body.MonsterID > 0 {
		drop.MonsterID = body.MonsterID
	}
	if body.EquipmentKey != "" {
		// Validate equipment exists
		var equip database.Equipment
		if err := database.DB.Where("equipment_key = ?", body.EquipmentKey).First(&equip).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Equipment not found"})
		}
		drop.DropKey = body.EquipmentKey
	}
	drop.DropRate = body.DropRate
	drop.DropMin = body.DropMin
	drop.DropMax = body.DropMax

	if err := database.DB.Save(&drop).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update monster drop"})
	}

	// Fetch full details for response
	database.DB.Preload("Monster").First(&drop)
	var equip database.Equipment
	database.DB.Where("equipment_key = ?", drop.DropKey).First(&equip)

	return c.JSON(MonsterDropManageResponse{
		ID:            drop.ID,
		MonsterID:     drop.MonsterID,
		MonsterKey:    drop.Monster.MonsterKey,
		MonsterName:   drop.Monster.Name,
		EquipmentKey:  drop.DropKey,
		EquipmentName: equip.Name,
		DropRate:      drop.DropRate,
		DropMin:       drop.DropMin,
		DropMax:       drop.DropMax,
	})
}

// AdminDeleteMonsterDrop deletes a monster drop entry.
// DELETE /api/admin/monster-drops/:id
func AdminDeleteMonsterDrop(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := database.DB.Delete(&database.MonsterDrop{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete monster drop"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Monster drop deleted"})
}

