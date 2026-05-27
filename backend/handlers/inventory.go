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

// GetOreTypes returns ore types from master table, optionally filtered by extraction type
func GetOreTypes(c *fiber.Ctx) error {
	var oreTypes []database.OreType
	query := database.DB.Order("sort_order ASC, id ASC")
	
	// Optional filter by extraction_type_id
	if extractionTypeID := c.Query("extraction_type_id"); extractionTypeID != "" {
		query = query.Where("extraction_type_id = ?", extractionTypeID)
	}
	
	query.Find(&oreTypes)
	return c.JSON(oreTypes)
}

// GetExtractableTypes returns all extraction types (e.g., Ore, Herb, Fish)
func GetExtractableTypes(c *fiber.Ctx) error {
	var types []database.ExtractableType
	database.DB.Order("sort_order ASC, id ASC").Find(&types)
	return c.JSON(types)
}

// GetHerbTypes returns herb types from master table, optionally filtered by extraction type
func GetHerbTypes(c *fiber.Ctx) error {
	var herbTypes []database.HerbType
	query := database.DB.Order("sort_order ASC, id ASC")
	
	// Optional filter by extraction_type_id
	if extractionTypeID := c.Query("extraction_type_id"); extractionTypeID != "" {
		query = query.Where("extraction_type_id = ?", extractionTypeID)
	}
	
	query.Find(&herbTypes)
	return c.JSON(herbTypes)
}

// HerbInventoryItemResponse for response
type HerbInventoryItemResponse struct {
	HerbKey       string `json:"herb_key"`
	HerbName      string `json:"herb_name"`
	Icon          string `json:"icon"`
	Color         string `json:"color"`
	Difficulty    string `json:"difficulty"`
	GatherTimeMS  int    `json:"gather_time_ms"`
	XPPerHerb     int    `json:"xp_per_herb"`
	LevelRequired int    `json:"level_required"`
	MaxQuantity   int    `json:"max_quantity"`
	SortOrder     int    `json:"sort_order"`
	Quantity      int    `json:"quantity"`
}

// GetHerbInventory returns all herb types with player's current quantity
func GetHerbInventory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var herbTypes []database.HerbType
	database.DB.Order("sort_order ASC, id ASC").Find(&herbTypes)

	var items []database.HerbInventoryItem
	database.DB.Where("user_id = ?", userID).Find(&items)

	// Build quantity map by HerbTypeID
	quantityMap := make(map[uint]int)
	for _, item := range items {
		quantityMap[item.HerbTypeID] = item.Quantity
	}

	response := make([]HerbInventoryItemResponse, 0, len(herbTypes))
	for _, ht := range herbTypes {
		response = append(response, HerbInventoryItemResponse{
			HerbKey:       ht.HerbKey,
			HerbName:      ht.HerbName,
			Icon:          ht.Icon,
			Color:         ht.Color,
			Difficulty:    ht.Difficulty,
			GatherTimeMS:  ht.GatherTimeMS,
			XPPerHerb:     ht.XPPerHerb,
			LevelRequired: ht.LevelRequired,
			MaxQuantity:   ht.MaxQuantity,
			SortOrder:     ht.SortOrder,
			Quantity:      quantityMap[ht.ID],
		})
	}
	return c.JSON(response)
}
