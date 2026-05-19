package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"idlegame-backend/database"
)

// BlacksmithSkillResponse represents user's blacksmith progression
type BlacksmithSkillResponse struct {
	Level      int `json:"level"`
	XP         int `json:"xp"`
	XPRequired int `json:"xp_required"`
	XPProgress int `json:"xp_progress"` // XP earned towards next level
}

// OfflineCraftingGainsInfo represents ingots crafted while offline
type OfflineCraftingGainsInfo struct {
	WasOffline     bool      `json:"was_offline"`
	OfflineTime    int64     `json:"offline_time_ms"`
	IngotsGained   int       `json:"ingots_gained"`
	RecipeName     string    `json:"recipe_name"`
	LastActiveTime time.Time `json:"last_active_time"`
}

// AwardBlacksmithXP adds XP to a user's blacksmith skill with auto level-up
func AwardBlacksmithXP(userID uint, xpAmount int) error {
	var skill database.UserBlacksmithSkill
	if err := database.DB.First(&skill, "user_id = ?", userID).Error; err != nil {
		return fmt.Errorf("blacksmith skill not found: %w", err)
	}

	// Add XP
	skill.XP += xpAmount
	skill.UpdatedAt = time.Now()

	// Check for level-ups (could level up multiple times in one big spike)
	for {
		var nextLevel database.BlacksmithLevel
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

// GetBlacksmithSkill returns the player's blacksmith skill details
// GET /api/blacksmith/skill
func GetBlacksmithSkill(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var skill database.UserBlacksmithSkill
	if err := database.DB.First(&skill, "user_id = ?", userID).Error; err != nil {
		// If skill doesn't exist, create it with level 1, xp 0
		skill = database.UserBlacksmithSkill{
			UserID:    userID,
			Level:     1,
			XP:        0,
			UpdatedAt: time.Now(),
		}
		if err := database.DB.Create(&skill).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to initialize blacksmith skill"})
		}
	}

	// Get current level's XP requirement (cumulative total)
	var currentLevel database.BlacksmithLevel
	database.DB.First(&currentLevel, skill.Level)

	// Get next level's XP requirement
	var nextLevel database.BlacksmithLevel
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
	progressAmount := nextLevelXPRequired - currentLevel.XPRequired
	if progressAmount <= 0 {
		progressAmount = 1 // Avoid divide by zero
	}
	xpProgress := (progressTowards * 100) / progressAmount

	return c.JSON(BlacksmithSkillResponse{
		Level:      skill.Level,
		XP:         skill.XP,
		XPRequired: nextLevelXPRequired,
		XPProgress: xpProgress,
	})
}

// GetCraftableItems returns all recipes player can craft (filtered by level)
// GET /api/blacksmith/recipes
func GetCraftableItems(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	// Get player's blacksmith level (lazy initialize if not found)
	var skill database.UserBlacksmithSkill
	if err := database.DB.First(&skill, "user_id = ?", userID).Error; err != nil {
		// If skill doesn't exist, create it with level 1, xp 0
		skill = database.UserBlacksmithSkill{
			UserID:    userID,
			Level:     1,
			XP:        0,
			UpdatedAt: time.Now(),
		}
		if err := database.DB.Create(&skill).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to initialize blacksmith skill"})
		}
	}

	// Get all craftable items ordered by sort order
	var recipes []database.CraftableItem
	database.DB.Order("sort_order").Find(&recipes)

	// Build response with ingredients for each recipe
	type RecipeResponse struct {
		ID               uint                                    `json:"id"`
		Name             string                                  `json:"name"`
		Icon             string                                  `json:"icon"`
		ItemKey          string                                  `json:"item_key"`
		CraftingTimeMS   int                                     `json:"crafting_time_ms"`
		XPPerCraft       int                                     `json:"xp_per_craft"`
		LevelRequired    int                                     `json:"level_required"`
		IsUnlocked       bool                                    `json:"is_unlocked"`
		Ingredients      []database.CraftRecipeIngredient        `json:"ingredients"`
	}

	var response []RecipeResponse
	for _, recipe := range recipes {
		var ingredients []database.CraftRecipeIngredient
		database.DB.Where("craftable_item_id = ?", recipe.ID).Find(&ingredients)

		response = append(response, RecipeResponse{
			ID:             recipe.ID,
			Name:           recipe.Name,
			Icon:           recipe.Icon,
			ItemKey:        recipe.ItemKey,
			CraftingTimeMS: recipe.CraftingTimeMS,
			XPPerCraft:     recipe.XPPerCraft,
			LevelRequired:  recipe.LevelRequired,
			IsUnlocked:     skill.Level >= recipe.LevelRequired,
			Ingredients:    ingredients,
		})
	}

	return c.JSON(response)
}

// GetIngotInventory returns player's ingot inventory
// GET /api/blacksmith/inventory
func GetIngotInventory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var items []database.UserIngotInventory
	database.DB.Where("user_id = ?", userID).Find(&items)

	response := make(map[string]int)
	for _, item := range items {
		response[item.IngotKey] = item.Quantity
	}

	return c.JSON(response)
}

// StartCrafting begins a crafting session
// POST /api/blacksmith/start
func StartCrafting(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	req := new(struct {
		CraftableItemID uint `json:"craftable_item_id"`
	})
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	// Verify recipe exists
	var recipe database.CraftableItem
	if err := database.DB.First(&recipe, req.CraftableItemID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "recipe not found"})
	}

	// Check if user has required level
	var skill database.UserBlacksmithSkill
	if err := database.DB.First(&skill, "user_id = ?", userID).Error; err != nil {
		// If skill doesn't exist, create it with level 1, xp 0
		skill = database.UserBlacksmithSkill{
			UserID:    userID,
			Level:     1,
			XP:        0,
			UpdatedAt: time.Now(),
		}
		if err := database.DB.Create(&skill).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to initialize blacksmith skill"})
		}
	}

	if skill.Level < recipe.LevelRequired {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "level required: " + fmt.Sprintf("%d", recipe.LevelRequired)})
	}

	// Check if user has required ingredients
	var ingredients []database.CraftRecipeIngredient
	database.DB.Where("craftable_item_id = ?", req.CraftableItemID).Find(&ingredients)

	for _, ing := range ingredients {
		var hasIngredient int64
		if ing.IngredientType == "ore" {
			var oreType database.OreType
			database.DB.Where("ore_key = ?", ing.IngredientKey).First(&oreType)
			var item database.OreInventoryItem
			database.DB.Where("user_id = ? AND ore_type_id = ?", userID, oreType.ID).First(&item)
			if item.Quantity < ing.QuantityRequired {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "not enough " + ing.IngredientKey})
			}
		} else if ing.IngredientType == "ingot" {
			database.DB.Model(&database.UserIngotInventory{}).
				Where("user_id = ? AND ingot_key = ?", userID, ing.IngredientKey).
				Count(&hasIngredient)
			// Create if doesn't exist
			var ingotItem database.UserIngotInventory
			database.DB.FirstOrCreate(&ingotItem, database.UserIngotInventory{UserID: userID, IngotKey: ing.IngredientKey})
			if ingotItem.Quantity < ing.QuantityRequired {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "not enough " + ing.IngredientKey})
			}
		}
	}

	// Stop any existing active session first
	var existingSession database.BlacksmithSession
	database.DB.Where("user_id = ? AND status = ?", userID, "active").First(&existingSession)
	if existingSession.ID != 0 {
		database.DB.Model(&existingSession).Update("status", "completed")
	}

	// Create new crafting session
	session := database.BlacksmithSession{
		UserID:          userID,
		CraftableItemID: req.CraftableItemID,
		StartedAt:       time.Now().UTC(),
		Status:          "active",
	}
	if err := database.DB.Create(&session).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to start crafting"})
	}

	database.LogActivity(userID, "Started crafting "+recipe.Name)

	return c.JSON(fiber.Map{
		"status":       "crafting started",
		"session_id":   session.ID,
		"recipe_name":  recipe.Name,
		"started_at":   session.StartedAt,
	})
}

// CalculateOfflineCraftingGains determines what the player crafted while offline
func CalculateOfflineCraftingGains(userID uint, session database.BlacksmithSession) OfflineCraftingGainsInfo {
	gains := OfflineCraftingGainsInfo{WasOffline: false}

	recipe := session.CraftableItem
	gains.RecipeName = recipe.Name

	now := time.Now().UTC()
	elapsed := now.Sub(session.StartedAt)
	craftingTimePerItem := time.Duration(recipe.CraftingTimeMS) * time.Millisecond

	if elapsed > craftingTimePerItem {
		gains.WasOffline = true
		gains.OfflineTime = elapsed.Milliseconds()
		gains.IngotsGained = int(elapsed.Milliseconds()) / recipe.CraftingTimeMS
		gains.LastActiveTime = session.StartedAt
	}

	return gains
}

// CalculateAndSaveCraftingGains calculates ingots produced and saves them
// Similar to CalculateAndSaveOreGains in mining
// Checks ingredients before each ingot production to prevent negative quantities
func CalculateAndSaveCraftingGains(userID uint, session database.BlacksmithSession) int {
	// Validate recipe still exists
	if session.CraftableItem.ID == 0 {
		// Recipe was deleted - produce nothing
		return 0
	}

	now := time.Now().UTC()
	elapsed := now.Sub(session.StartedAt)
	maxIngotsProduced := int(elapsed.Milliseconds()) / session.CraftableItem.CraftingTimeMS

	recipe := session.CraftableItem
	var ingredients []database.CraftRecipeIngredient
	database.DB.Where("craftable_item_id = ?", recipe.ID).Find(&ingredients)

	// Produce ingots one at a time, checking ingredients for each
	ingotsProduced := 0
	for i := 0; i < maxIngotsProduced; i++ {
		// Check if we have all required ingredients BEFORE producing this ingot
		canProduce := true
		for _, ing := range ingredients {
			if ing.IngredientType == "ore" {
				var oreType database.OreType
				database.DB.Where("ore_key = ?", ing.IngredientKey).First(&oreType)
				var item database.OreInventoryItem
				database.DB.Where("user_id = ? AND ore_type_id = ?", userID, oreType.ID).First(&item)
				if item.Quantity < ing.QuantityRequired {
					canProduce = false
					break
				}
			} else if ing.IngredientType == "ingot" {
				var item database.UserIngotInventory
				database.DB.Where("user_id = ? AND ingot_key = ?", userID, ing.IngredientKey).First(&item)
				if item.Quantity < ing.QuantityRequired {
					canProduce = false
					break
				}
			}
		}

		// If we can't produce, stop here
		if !canProduce {
			break
		}

		// Deduct ingredients for this ingot
		for _, ing := range ingredients {
			if ing.IngredientType == "ore" {
				var oreType database.OreType
				database.DB.Where("ore_key = ?", ing.IngredientKey).First(&oreType)
				database.DB.Exec(
					"UPDATE ore_inventory_items SET quantity = quantity - ? WHERE user_id = ? AND ore_type_id = ?",
					ing.QuantityRequired, userID, oreType.ID,
				)
			} else if ing.IngredientType == "ingot" {
				database.DB.Exec(
					"UPDATE user_ingot_inventories SET quantity = quantity - ? WHERE user_id = ? AND ingot_key = ?",
					ing.QuantityRequired, userID, ing.IngredientKey,
				)
			}
		}

		ingotsProduced++
	}

	// Add crafted ingots to inventory
	if recipe.OutputType == "ingot" && ingotsProduced > 0 {
		var ingotItem database.UserIngotInventory
		database.DB.FirstOrCreate(&ingotItem, database.UserIngotInventory{UserID: userID, IngotKey: recipe.ItemKey})

		// Apply max quantity cap if set
		quantityToAdd := ingotsProduced
		if recipe.MaxQuantity > 0 {
			remaining := recipe.MaxQuantity - ingotItem.Quantity
			if quantityToAdd > remaining {
				quantityToAdd = remaining
			}
			if quantityToAdd < 0 {
				quantityToAdd = 0
			}
		}

		if quantityToAdd > 0 {
			database.DB.Exec(
				"UPDATE user_ingot_inventories SET quantity = quantity + ? WHERE user_id = ? AND ingot_key = ?",
				quantityToAdd, userID, recipe.ItemKey,
			)
		}
	}

	// Award blacksmith XP for all ingots produced
	if ingotsProduced > 0 {
		totalXP := ingotsProduced * recipe.XPPerCraft
		AwardBlacksmithXP(userID, totalXP)
	}

	return ingotsProduced
}

// StopCrafting completes the crafting session
// POST /api/blacksmith/stop
func StopCrafting(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var session database.BlacksmithSession
	result := database.DB.Where("user_id = ? AND status = ?", userID, "active").
		Preload("CraftableItem").
		First(&session)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no active crafting session"})
	}

	ingotsProduced := CalculateAndSaveCraftingGains(userID, session)

	now := time.Now().UTC()
	database.DB.Model(&session).Updates(map[string]interface{}{
		"status":   "completed",
		"ended_at": now,
	})

	recipe := session.CraftableItem
	database.LogActivity(userID, fmt.Sprintf("Stopped crafting %s. Produced %d ingots.", recipe.Name, ingotsProduced))

	return c.JSON(fiber.Map{
		"status":           "crafting stopped",
		"recipe_name":      recipe.Name,
		"ingots_produced":  ingotsProduced,
		"xp_earned":        ingotsProduced * recipe.XPPerCraft,
	})
}

// GetCraftingStatus returns current crafting status and all ingot counts with pending
// GET /api/blacksmith/status
func GetCraftingStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var session database.BlacksmithSession
	result := database.DB.Where("user_id = ? AND status = ?", userID, "active").
		Preload("CraftableItem").
		First(&session)

	isActive := result.Error == nil

	// Safety: if recipe was deleted, mark session as inactive
	if isActive && session.CraftableItem.ID == 0 {
		database.DB.Model(&session).Update("status", "completed")
		isActive = false
	}

	// Build current ingot quantities from database
	var items []database.UserIngotInventory
	database.DB.Where("user_id = ?", userID).Find(&items)

	currentIngots := make(map[string]int)
	for _, item := range items {
		currentIngots[item.IngotKey] = item.Quantity
	}

	// Add pending (unsaved) ingots for the active session
	if isActive {
		recipe := session.CraftableItem
		now := time.Now().UTC()
		elapsed := now.Sub(session.StartedAt)
		pendingIngots := int(elapsed.Milliseconds()) / recipe.CraftingTimeMS

		// Apply max quantity cap
		if recipe.MaxQuantity > 0 {
			existing := currentIngots[recipe.ItemKey]
			remaining := recipe.MaxQuantity - existing
			if pendingIngots > remaining {
				pendingIngots = remaining
			}
			if pendingIngots < 0 {
				pendingIngots = 0
			}
		}

		currentIngots[recipe.ItemKey] += pendingIngots
	}

	response := fiber.Map{
		"is_active":       isActive,
		"current_ingots":  currentIngots,
	}

	if isActive {
		recipe := session.CraftableItem
		response["craftable_item_id"] = recipe.ID
		response["recipe_name"] = recipe.Name
		response["crafting_time_ms"] = recipe.CraftingTimeMS
		response["started_at"] = session.StartedAt
		response["offline_gains"] = CalculateOfflineCraftingGains(userID, session)
	}

	return c.JSON(response)
}

// ===== ADMIN CRUD ENDPOINTS =====

// AdminGetCraftableItems returns all craftable items
// GET /api/admin/craftable-items
func AdminGetCraftableItems(c *fiber.Ctx) error {
	var items []database.CraftableItem
	database.DB.Order("sort_order").Find(&items)
	return c.JSON(items)
}

// AdminCreateCraftableItem creates a new craftable item
// POST /api/admin/craftable-items
func AdminCreateCraftableItem(c *fiber.Ctx) error {
	var body struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		Icon           string `json:"icon"`
		ItemKey        string `json:"item_key"`
		OutputType     string `json:"output_type"`
		CraftingTimeMS int    `json:"crafting_time_ms"`
		XPPerCraft     int    `json:"xp_per_craft"`
		LevelRequired  int    `json:"level_required"`
		MaxQuantity    int    `json:"max_quantity"`
		SortOrder      int    `json:"sort_order"`
		Ingredients    []struct {
			IngredientType   string `json:"ingredient_type"`
			IngredientKey    string `json:"ingredient_key"`
			QuantityRequired int    `json:"quantity_required"`
		} `json:"ingredients"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	item := database.CraftableItem{
		Name:           body.Name,
		Description:    body.Description,
		Icon:           body.Icon,
		ItemKey:        body.ItemKey,
		OutputType:     body.OutputType,
		CraftingTimeMS: body.CraftingTimeMS,
		XPPerCraft:     body.XPPerCraft,
		LevelRequired:  body.LevelRequired,
		MaxQuantity:    body.MaxQuantity,
		SortOrder:      body.SortOrder,
	}

	if err := database.DB.Create(&item).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create item"})
	}

	// Add ingredients
	for _, ing := range body.Ingredients {
		ingredient := database.CraftRecipeIngredient{
			CraftableItemID:  item.ID,
			IngredientType:   ing.IngredientType,
			IngredientKey:    ing.IngredientKey,
			QuantityRequired: ing.QuantityRequired,
		}
		database.DB.Create(&ingredient)
	}

	return c.Status(201).JSON(item)
}

// AdminUpdateCraftableItem updates a craftable item
// PUT /api/admin/craftable-items/:id
func AdminUpdateCraftableItem(c *fiber.Ctx) error {
	idStr := c.Params("id")
	var id uint
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID parameter"})
	}

	var body struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		Icon           string `json:"icon"`
		CraftingTimeMS int    `json:"crafting_time_ms"`
		XPPerCraft     int    `json:"xp_per_craft"`
		LevelRequired  int    `json:"level_required"`
		MaxQuantity    int    `json:"max_quantity"`
		SortOrder      int    `json:"sort_order"`
		Ingredients    []struct {
			IngredientType   string `json:"ingredient_type"`
			IngredientKey    string `json:"ingredient_key"`
			QuantityRequired int    `json:"quantity_required"`
		} `json:"ingredients"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var item database.CraftableItem
	if err := database.DB.First(&item, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Item not found"})
	}

	item.Name = body.Name
	item.Description = body.Description
	item.Icon = body.Icon
	item.CraftingTimeMS = body.CraftingTimeMS
	item.XPPerCraft = body.XPPerCraft
	item.LevelRequired = body.LevelRequired
	item.MaxQuantity = body.MaxQuantity
	item.SortOrder = body.SortOrder

	if err := database.DB.Save(&item).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update item"})
	}

	// Update ingredients (delete old, add new)
	database.DB.Where("craftable_item_id = ?", id).Delete(&database.CraftRecipeIngredient{})
	for _, ing := range body.Ingredients {
		ingredient := database.CraftRecipeIngredient{
			CraftableItemID:  item.ID,
			IngredientType:   ing.IngredientType,
			IngredientKey:    ing.IngredientKey,
			QuantityRequired: ing.QuantityRequired,
		}
		database.DB.Create(&ingredient)
	}

	return c.JSON(item)
}

// AdminDeleteCraftableItem deletes a craftable item
// DELETE /api/admin/craftable-items/:id
func AdminDeleteCraftableItem(c *fiber.Ctx) error {
	idStr := c.Params("id")
	var id uint
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID parameter"})
	}

	var item database.CraftableItem
	if err := database.DB.First(&item, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Item not found"})
	}

	// Delete ingredients first
	database.DB.Where("craftable_item_id = ?", id).Delete(&database.CraftRecipeIngredient{})

	if err := database.DB.Delete(&item).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete item"})
	}

	return c.JSON(fiber.Map{"success": true})
}

// AdminGetBlacksmithLevels returns all blacksmith levels
// GET /api/admin/blacksmith-levels
func AdminGetBlacksmithLevels(c *fiber.Ctx) error {
	var levels []database.BlacksmithLevel
	database.DB.Order("level").Find(&levels)
	return c.JSON(levels)
}

// AdminCreateBlacksmithLevel creates a new blacksmith level
// POST /api/admin/blacksmith-levels
func AdminCreateBlacksmithLevel(c *fiber.Ctx) error {
	var body struct {
		Level      int `json:"level"`
		XPRequired int `json:"xp_required"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	bl := database.BlacksmithLevel{
		Level:      body.Level,
		XPRequired: body.XPRequired,
	}

	if err := database.DB.Create(&bl).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create blacksmith level"})
	}

	return c.Status(201).JSON(bl)
}

// AdminUpdateBlacksmithLevel updates a blacksmith level
// PUT /api/admin/blacksmith-levels/:level
func AdminUpdateBlacksmithLevel(c *fiber.Ctx) error {
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

	var bl database.BlacksmithLevel
	if err := database.DB.First(&bl, level).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Blacksmith level not found"})
	}

	bl.XPRequired = body.XPRequired
	if err := database.DB.Save(&bl).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update blacksmith level"})
	}

	return c.JSON(bl)
}

// AdminDeleteBlacksmithLevel deletes a blacksmith level
// DELETE /api/admin/blacksmith-levels/:level
func AdminDeleteBlacksmithLevel(c *fiber.Ctx) error {
	levelStr := c.Params("level")
	var level int
	if _, err := fmt.Sscanf(levelStr, "%d", &level); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid level parameter"})
	}

	var bl database.BlacksmithLevel
	if err := database.DB.First(&bl, level).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Blacksmith level not found"})
	}

	if err := database.DB.Delete(&bl).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete blacksmith level"})
	}

	return c.JSON(fiber.Map{"success": true, "level": level})
}
