package handlers

import (
	"github.com/gofiber/fiber/v2"
	"idlegame-backend/database"
)

// OreInventoryItemResponse is one row in the player's ore inventory
type OreInventoryItemResponse struct {
	OreKey          string `json:"ore_key"`
	OreName         string `json:"ore_name"`
	Icon            string `json:"icon"`
	Color           string `json:"color"`
	Difficulty      string `json:"difficulty"`
	MiningTimeMS    int    `json:"mining_time_ms"`
	XPPerOre        int    `json:"xp_per_ore"`
	LevelRequired   int    `json:"level_required"`
	PickaxeRequired string `json:"pickaxe_required"`
	MaxQuantity     int    `json:"max_quantity"`
	SortOrder       int    `json:"sort_order"`
	Quantity        int    `json:"quantity"`
}

// GetOreInventory returns all ore types with the player's current quantity for each
func GetOreInventory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var oreTypes []database.OreType
	database.DB.Order("sort_order ASC, id ASC").Find(&oreTypes)

	var items []database.OreInventoryItem
	database.DB.Where("user_id = ?", userID).Find(&items)

	// Build quantity map by OreTypeID
	quantityMap := make(map[uint]int)
	for _, item := range items {
		quantityMap[item.OreTypeID] = item.Quantity
	}

	response := make([]OreInventoryItemResponse, 0, len(oreTypes))
	for _, ot := range oreTypes {
		response = append(response, OreInventoryItemResponse{
			OreKey:          ot.OreKey,
			OreName:         ot.OreName,
			Icon:            ot.Icon,
			Color:           ot.Color,
			Difficulty:      ot.Difficulty,
			MiningTimeMS:    ot.MiningTimeMS,
			XPPerOre:        ot.XPPerOre,
			LevelRequired:   ot.LevelRequired,
			PickaxeRequired: ot.PickaxeRequired,
			MaxQuantity:     ot.MaxQuantity,
			SortOrder:       ot.SortOrder,
			Quantity:        quantityMap[ot.ID],
		})
	}
	return c.JSON(response)
}

// GetOreTypes returns all ore types from the master table (for frontend dynamic rendering)
func GetOreTypes(c *fiber.Ctx) error {
	var oreTypes []database.OreType
	database.DB.Order("sort_order ASC, id ASC").Find(&oreTypes)
	return c.JSON(oreTypes)
}
