package handlers

import (
	"github.com/gofiber/fiber/v2"
	"idlegame-backend/database"
)

// ── Responses ──────────────────────────────────────────────────────────────

type OreResponse struct {
	ID              uint   `json:"id"`
	OreKey          string `json:"ore_key"`
	OreName         string `json:"ore_name"`
	Icon            string `json:"icon"`
	Color           string `json:"color"`
	SVG             string `json:"svg"`
	Difficulty      string `json:"difficulty"`
	MiningTimeMS    int    `json:"mining_time_ms"`
	XPPerOre        int    `json:"xp_per_ore"`
	LevelRequired   int    `json:"level_required"`
	PickaxeRequired string `json:"pickaxe_required"`
	MaxQuantity     int    `json:"max_quantity"`
	SortOrder       int    `json:"sort_order"`
}

func toOreResponse(ore database.OreType) OreResponse {
	return OreResponse{
		ID:              ore.ID,
		OreKey:          ore.OreKey,
		OreName:         ore.OreName,
		Icon:            ore.Icon,
		Color:           ore.Color,
		SVG:             ore.SVG,
		Difficulty:      ore.Difficulty,
		MiningTimeMS:    ore.MiningTimeMS,
		XPPerOre:        ore.XPPerOre,
		LevelRequired:   ore.LevelRequired,
		PickaxeRequired: ore.PickaxeRequired,
		MaxQuantity:     ore.MaxQuantity,
		SortOrder:       ore.SortOrder,
	}
}

// ── Admin Ores CRUD ────────────────────────────────────────────────────────

// AdminGetAllOres returns all ore types
func AdminGetAllOres(c *fiber.Ctx) error {
	var ores []database.OreType
	if err := database.DB.Find(&ores).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load ores"})
	}

	responses := make([]OreResponse, len(ores))
	for i, ore := range ores {
		responses[i] = toOreResponse(ore)
	}
	return c.JSON(responses)
}

// AdminCreateOre creates a new ore type
func AdminCreateOre(c *fiber.Ctx) error {
	type CreateOreRequest struct {
		OreKey          string `json:"ore_key"`
		OreName         string `json:"ore_name"`
		Icon            string `json:"icon"`
		Color           string `json:"color"`
		SVG             string `json:"svg"`
		Difficulty      string `json:"difficulty"`
		MiningTimeMS    int    `json:"mining_time_ms"`
		XPPerOre        int    `json:"xp_per_ore"`
		LevelRequired   int    `json:"level_required"`
		PickaxeRequired string `json:"pickaxe_required"`
		MaxQuantity     int    `json:"max_quantity"`
		SortOrder       int    `json:"sort_order"`
		BasePrice       int    `json:"base_price"`
	}

	var req CreateOreRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if req.OreKey == "" || req.OreName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ore_key and ore_name are required"})
	}

	ore := database.OreType{
		OreKey:          req.OreKey,
		OreName:         req.OreName,
		Icon:            req.Icon,
		Color:           req.Color,
		SVG:             req.SVG,
		Difficulty:      req.Difficulty,
		MiningTimeMS:    req.MiningTimeMS,
		XPPerOre:        req.XPPerOre,
		LevelRequired:   req.LevelRequired,
		PickaxeRequired: req.PickaxeRequired,
		MaxQuantity:     req.MaxQuantity,
		SortOrder:       req.SortOrder,
		BasePrice:       req.BasePrice,
	}

	if err := database.DB.Create(&ore).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create ore"})
	}

	return c.Status(201).JSON(toOreResponse(ore))
}

// AdminUpdateOre updates an existing ore type
func AdminUpdateOre(c *fiber.Ctx) error {
	id := c.Params("id")

	type UpdateOreRequest struct {
		OreName         string `json:"ore_name"`
		Icon            string `json:"icon"`
		Color           string `json:"color"`
		SVG             string `json:"svg"`
		Difficulty      string `json:"difficulty"`
		MiningTimeMS    int    `json:"mining_time_ms"`
		XPPerOre        int    `json:"xp_per_ore"`
		LevelRequired   int    `json:"level_required"`
		PickaxeRequired string `json:"pickaxe_required"`
		MaxQuantity     int    `json:"max_quantity"`
		SortOrder       int    `json:"sort_order"`
		BasePrice       int    `json:"base_price"`
	}

	var req UpdateOreRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var ore database.OreType
	if err := database.DB.First(&ore, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Ore not found"})
	}

	updates := database.OreType{
		OreName:         req.OreName,
		Icon:            req.Icon,
		Color:           req.Color,
		SVG:             req.SVG,
		Difficulty:      req.Difficulty,
		MiningTimeMS:    req.MiningTimeMS,
		XPPerOre:        req.XPPerOre,
		LevelRequired:   req.LevelRequired,
		PickaxeRequired: req.PickaxeRequired,
		MaxQuantity:     req.MaxQuantity,
		SortOrder:       req.SortOrder,
		BasePrice:       req.BasePrice,
	}

	if err := database.DB.Model(&ore).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update ore"})
	}

	return c.JSON(toOreResponse(ore))
}

// AdminDeleteOre deletes an ore type
func AdminDeleteOre(c *fiber.Ctx) error {
	id := c.Params("id")

	var ore database.OreType
	if err := database.DB.First(&ore, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Ore not found"})
	}

	if err := database.DB.Delete(&ore).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete ore"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Ore deleted successfully"})
}
