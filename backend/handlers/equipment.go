package handlers

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"idlegame-backend/database"
)

// ── Shared response shapes ─────────────────────────────────────────────────

type EquipmentModifier struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type EquipmentResponse struct {
	ID           uint                `json:"id"`
	EquipmentKey string              `json:"equipment_key"`
	Name         string              `json:"name"`
	Icon         string              `json:"icon"`
	Description  string              `json:"description"`
	Slot         string              `json:"slot"`
	Rarity       string              `json:"rarity"`
	BaseAttack   int                 `json:"base_attack"`
	AttackType   string              `json:"attack_type"`
	BaseDefence  int                 `json:"base_defence"`
	Modifiers    []EquipmentModifier `json:"modifiers"`
	LevelRequired int                `json:"level_required"`
}

type UserEquipmentResponse struct {
	UserEquipmentID uint              `json:"user_equipment_id"`
	Equipment       EquipmentResponse `json:"equipment"`
	ObtainedAt      time.Time         `json:"obtained_at"`
}

type EquippedSlotResponse struct {
	Slot            string             `json:"slot"`
	UserEquipmentID uint               `json:"user_equipment_id"` // 0 = empty
	Equipment       *EquipmentResponse `json:"equipment"`         // nil = empty
}

// parseModifiers safely unmarshals ModifiersJSON; returns empty slice on error.
func parseModifiers(raw string) []EquipmentModifier {
	var mods []EquipmentModifier
	if raw == "" || raw == "[]" {
		return mods
	}
	_ = json.Unmarshal([]byte(raw), &mods)
	return mods
}

func toEquipmentResponse(e database.Equipment) EquipmentResponse {
	return EquipmentResponse{
		ID:            e.ID,
		EquipmentKey:  e.EquipmentKey,
		Name:          e.Name,
		Icon:          e.Icon,
		Description:   e.Description,
		Slot:          e.Slot,
		Rarity:        e.Rarity,
		BaseAttack:    e.BaseAttack,
		AttackType:    e.AttackType,
		BaseDefence:   e.BaseDefence,
		Modifiers:     parseModifiers(e.ModifiersJSON),
		LevelRequired: e.LevelRequired,
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────

// GetEquipmentTypes returns the full equipment master table.
// GET /api/equipment/types — public (no auth required)
func GetEquipmentTypes(c *fiber.Ctx) error {
	var items []database.Equipment
	if err := database.DB.Order("sort_order ASC").Find(&items).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load equipment types"})
	}

	result := make([]EquipmentResponse, 0, len(items))
	for _, e := range items {
		result = append(result, toEquipmentResponse(e))
	}
	return c.JSON(result)
}

// GetEquipmentBag returns the current user's equipment bag (all obtained items).
// GET /api/equipment/bag
func GetEquipmentBag(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var rows []database.UserEquipment
	if err := database.DB.
		Preload("Equipment").
		Where("user_id = ?", userID).
		Order("obtained_at DESC").
		Find(&rows).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load equipment bag"})
	}

	result := make([]UserEquipmentResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, UserEquipmentResponse{
			UserEquipmentID: row.ID,
			Equipment:       toEquipmentResponse(row.Equipment),
			ObtainedAt:      row.ObtainedAt,
		})
	}
	return c.JSON(result)
}

// GetEquippedSlots returns the user's currently equipped items for all 8 slots.
// GET /api/equipment/slots
func GetEquippedSlots(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	allSlots := []string{"head", "chest", "legs", "weapon", "shield", "ring1", "ring2", "amulet"}

	// Fetch all slot rows for this user
	var slotRows []database.UserEquippedSlot
	database.DB.Where("user_id = ?", userID).Find(&slotRows)

	// Index by slot name
	slotMap := make(map[string]database.UserEquippedSlot)
	for _, s := range slotRows {
		slotMap[s.Slot] = s
	}

	// Build response — one entry per slot, filling in equipment if equipped
	result := make([]EquippedSlotResponse, 0, len(allSlots))
	for _, slotName := range allSlots {
		entry := EquippedSlotResponse{Slot: slotName}

		if row, ok := slotMap[slotName]; ok && row.UserEquipmentID != 0 {
			entry.UserEquipmentID = row.UserEquipmentID

			// Load the UserEquipment + Equipment
			var ue database.UserEquipment
			if err := database.DB.Preload("Equipment").First(&ue, row.UserEquipmentID).Error; err == nil {
				er := toEquipmentResponse(ue.Equipment)
				entry.Equipment = &er
			}
		}

		result = append(result, entry)
	}
	return c.JSON(result)
}

// EquipItem equips a user_equipment_id into the given slot.
// POST /api/equipment/equip  body: { "user_equipment_id": 5, "slot": "weapon" }
func EquipItem(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var body struct {
		UserEquipmentID uint   `json:"user_equipment_id"`
		Slot            string `json:"slot"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	validSlots := map[string]bool{
		"head": true, "chest": true, "legs": true,
		"weapon": true, "shield": true,
		"ring1": true, "ring2": true, "amulet": true,
	}
	if !validSlots[body.Slot] {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid slot name"})
	}

	// Verify the UserEquipment belongs to this user
	var ue database.UserEquipment
	if err := database.DB.Preload("Equipment").
		Where("id = ? AND user_id = ?", body.UserEquipmentID, userID).
		First(&ue).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Equipment not found in your bag"})
	}

	// Ring slots accept any ring equipment; other slots must match exactly
	expectedSlot := ue.Equipment.Slot
	if expectedSlot == "ring" {
		if body.Slot != "ring1" && body.Slot != "ring2" {
			return c.Status(400).JSON(fiber.Map{"error": "Rings can only be equipped in ring1 or ring2 slots"})
		}
	} else if expectedSlot != body.Slot {
		return c.Status(400).JSON(fiber.Map{"error": "Equipment does not fit this slot"})
	}

	// Upsert the slot row
	var slotRow database.UserEquippedSlot
	result := database.DB.Where("user_id = ? AND slot = ?", userID, body.Slot).First(&slotRow)
	if result.Error != nil {
		// Create new slot row
		slotRow = database.UserEquippedSlot{
			UserID:          userID,
			Slot:            body.Slot,
			UserEquipmentID: body.UserEquipmentID,
		}
		database.DB.Create(&slotRow)
	} else {
		// Update existing
		slotRow.UserEquipmentID = body.UserEquipmentID
		database.DB.Save(&slotRow)
	}

	return c.JSON(fiber.Map{"success": true, "slot": body.Slot})
}

// UnequipSlot clears a slot.
// POST /api/equipment/unequip  body: { "slot": "weapon" }
func UnequipSlot(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var body struct {
		Slot string `json:"slot"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	database.DB.
		Where("user_id = ? AND slot = ?", userID, body.Slot).
		Delete(&database.UserEquippedSlot{})

	return c.JSON(fiber.Map{"success": true, "slot": body.Slot})
}

// GiveEquipment grants an equipment item to the player (for testing / future drop system).
// POST /api/equipment/give  body: { "equipment_key": "rusty_blade" }
func GiveEquipment(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var body struct {
		EquipmentKey string `json:"equipment_key"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var equip database.Equipment
	if err := database.DB.Where("equipment_key = ?", body.EquipmentKey).First(&equip).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Equipment not found"})
	}

	ue := database.UserEquipment{
		UserID:      userID,
		EquipmentID: equip.ID,
		ObtainedAt:  time.Now(),
	}
	database.DB.Create(&ue)

	return c.JSON(fiber.Map{"success": true, "user_equipment_id": ue.ID})
}

// ── Admin endpoints (for development only — delete before production) ──────────

// AdminGetAllEquipment returns all equipment from the master table.
// GET /api/admin/equipment
func AdminGetAllEquipment(c *fiber.Ctx) error {
	var items []database.Equipment
	if err := database.DB.Order("sort_order ASC, id ASC").Find(&items).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load equipment"})
	}

	result := make([]EquipmentResponse, 0, len(items))
	for _, e := range items {
		result = append(result, toEquipmentResponse(e))
	}
	return c.JSON(result)
}

// AdminCreateEquipment creates a new equipment item.
// POST /api/admin/equipment
func AdminCreateEquipment(c *fiber.Ctx) error {
	var body struct {
		EquipmentKey string `json:"equipment_key"`
		Name         string `json:"name"`
		Icon         string `json:"icon"`
		Description  string `json:"description"`
		Slot         string `json:"slot"`
		Rarity       string `json:"rarity"`
		BaseAttack   int    `json:"base_attack"`
		AttackType   string `json:"attack_type"`
		BaseDefence  int    `json:"base_defence"`
		Modifiers    []EquipmentModifier `json:"modifiers"`
		LevelRequired int    `json:"level_required"`
		SortOrder    int    `json:"sort_order"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if body.EquipmentKey == "" || body.Name == "" || body.Slot == "" {
		return c.Status(400).JSON(fiber.Map{"error": "equipment_key, name, and slot are required"})
	}

	// Convert modifiers to JSON
	modifiersJSON := "[]"
	if len(body.Modifiers) > 0 {
		if data, err := json.Marshal(body.Modifiers); err == nil {
			modifiersJSON = string(data)
		}
	}

	equip := database.Equipment{
		EquipmentKey:  body.EquipmentKey,
		Name:          body.Name,
		Icon:          body.Icon,
		Description:   body.Description,
		Slot:          body.Slot,
		Rarity:        body.Rarity,
		BaseAttack:    body.BaseAttack,
		AttackType:    body.AttackType,
		BaseDefence:   body.BaseDefence,
		ModifiersJSON: modifiersJSON,
		LevelRequired: body.LevelRequired,
		SortOrder:     body.SortOrder,
		CreatedAt:     time.Now(),
	}

	if err := database.DB.Create(&equip).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create equipment"})
	}

	return c.Status(201).JSON(toEquipmentResponse(equip))
}

// AdminUpdateEquipment updates an existing equipment item.
// PUT /api/admin/equipment/:id
func AdminUpdateEquipment(c *fiber.Ctx) error {
	id := c.Params("id")

	var equip database.Equipment
	if err := database.DB.First(&equip, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Equipment not found"})
	}

	var body struct {
		EquipmentKey string `json:"equipment_key"`
		Name         string `json:"name"`
		Icon         string `json:"icon"`
		Description  string `json:"description"`
		Slot         string `json:"slot"`
		Rarity       string `json:"rarity"`
		BaseAttack   int    `json:"base_attack"`
		AttackType   string `json:"attack_type"`
		BaseDefence  int    `json:"base_defence"`
		Modifiers    []EquipmentModifier `json:"modifiers"`
		LevelRequired int    `json:"level_required"`
		SortOrder    int    `json:"sort_order"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Update fields
	if body.EquipmentKey != "" {
		equip.EquipmentKey = body.EquipmentKey
	}
	if body.Name != "" {
		equip.Name = body.Name
	}
	equip.Icon = body.Icon
	equip.Description = body.Description
	if body.Slot != "" {
		equip.Slot = body.Slot
	}
	equip.Rarity = body.Rarity
	equip.BaseAttack = body.BaseAttack
	equip.AttackType = body.AttackType
	equip.BaseDefence = body.BaseDefence
	equip.LevelRequired = body.LevelRequired
	equip.SortOrder = body.SortOrder

	// Convert modifiers to JSON
	if len(body.Modifiers) > 0 {
		if data, err := json.Marshal(body.Modifiers); err == nil {
			equip.ModifiersJSON = string(data)
		}
	} else {
		equip.ModifiersJSON = "[]"
	}

	if err := database.DB.Save(&equip).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update equipment"})
	}

	return c.JSON(toEquipmentResponse(equip))
}

// AdminDeleteEquipment deletes an equipment item.
// DELETE /api/admin/equipment/:id
func AdminDeleteEquipment(c *fiber.Ctx) error {
	id := c.Params("id")

	var equip database.Equipment
	if err := database.DB.First(&equip, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Equipment not found"})
	}

	if err := database.DB.Delete(&equip).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete equipment"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Equipment deleted"})
}

