package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"idlegame-backend/database"
)

// ── Response types ─────────────────────────────────────────────────────────

type QuestObjectiveResponse struct {
	ID            uint   `json:"id"`
	ObjectiveType string `json:"objective_type"`
	TargetKey     string `json:"target_key"`
	TargetCount   int    `json:"target_count"`
	DisplayText   string `json:"display_text"`
}

type QuestRewardResponse struct {
	ID         uint   `json:"id"`
	RewardType string `json:"reward_type"`
	RewardKey  string `json:"reward_key"`
	Amount     int64  `json:"amount"`
}

type QuestResponse struct {
	ID             uint                     `json:"id"`
	QuestKey       string                   `json:"quest_key"`
	Title          string                   `json:"title"`
	Chapter        int                      `json:"chapter"`
	SortOrder      int                      `json:"sort_order"`
	IntroText      string                   `json:"intro_text"`
	CompletionText string                   `json:"completion_text"`
	Status         string                   `json:"status"`   // locked | available | completed
	CompletedAt    *time.Time               `json:"completed_at"`
	Objectives     []QuestObjectiveResponse `json:"objectives,omitempty"`
	Rewards        []QuestRewardResponse    `json:"rewards,omitempty"`
}

// ── GetQuests ──────────────────────────────────────────────────────────────

// GetQuests returns all quests with the requesting user's status.
// Objectives and rewards are included for available and completed quests.
func GetQuests(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var allQuests []database.Quest
	if err := database.DB.Order("chapter ASC, sort_order ASC").Find(&allQuests).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to load quests"})
	}

	// Load all UserQuest rows for this user
	var userQuests []database.UserQuest
	database.DB.Where("user_id = ?", userID).Find(&userQuests)
	statusMap := map[uint]database.UserQuest{}
	for _, uq := range userQuests {
		statusMap[uq.QuestID] = uq
	}

	var result []QuestResponse
	for _, q := range allQuests {
		uq := statusMap[q.ID]
		status := uq.Status
		if status == "" {
			status = "locked"
		}

		qr := QuestResponse{
			ID:          q.ID,
			QuestKey:    q.QuestKey,
			Title:       q.Title,
			Chapter:     q.Chapter,
			SortOrder:   q.SortOrder,
			Status:      status,
			CompletedAt: uq.CompletedAt,
		}

		// Include full detail for non-locked quests
		if status != "locked" {
			qr.IntroText = q.IntroText
			qr.CompletionText = q.CompletionText
			qr.Objectives = loadObjectives(q.ID)
			qr.Rewards = loadRewards(q.ID)
		}

		result = append(result, qr)
	}

	return c.JSON(result)
}

// ── GetQuest ───────────────────────────────────────────────────────────────

// GetQuest returns full detail for a single quest.
func GetQuest(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	questKey := c.Params("key")

	var q database.Quest
	if err := database.DB.Where("quest_key = ?", questKey).First(&q).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "quest not found"})
	}

	var uq database.UserQuest
	database.DB.Where("user_id = ? AND quest_id = ?", userID, q.ID).First(&uq)
	status := uq.Status
	if status == "" {
		status = "locked"
	}

	return c.JSON(QuestResponse{
		ID:             q.ID,
		QuestKey:       q.QuestKey,
		Title:          q.Title,
		Chapter:        q.Chapter,
		SortOrder:      q.SortOrder,
		IntroText:      q.IntroText,
		CompletionText: q.CompletionText,
		Status:         status,
		CompletedAt:    uq.CompletedAt,
		Objectives:     loadObjectives(q.ID),
		Rewards:        loadRewards(q.ID),
	})
}

// ── CompleteQuest ──────────────────────────────────────────────────────────

// CompleteQuest validates all objectives and, if met, grants rewards and unlocks the next quest.
func CompleteQuest(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	questKey := c.Params("key")

	// Load quest
	var q database.Quest
	if err := database.DB.Where("quest_key = ?", questKey).First(&q).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "quest not found"})
	}

	// Load user quest row — must be available
	var uq database.UserQuest
	if err := database.DB.Where("user_id = ? AND quest_id = ?", userID, q.ID).First(&uq).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "quest not found for user"})
	}
	if uq.Status == "completed" {
		return c.Status(400).JSON(fiber.Map{"error": "quest already completed"})
	}
	if uq.Status == "locked" {
		return c.Status(400).JSON(fiber.Map{"error": "quest is locked"})
	}

	// Validate all objectives
	var objectives []database.QuestObjective
	database.DB.Where("quest_id = ?", q.ID).Find(&objectives)

	var unmet []string
	for _, obj := range objectives {
		if err := checkObjective(userID, obj); err != nil {
			unmet = append(unmet, err.Error())
		}
	}
	if len(unmet) > 0 {
		return c.Status(400).JSON(fiber.Map{
			"error":    "objectives not met",
			"unmet":    unmet,
		})
	}

	// Consume delivered items (deliver objectives)
	for _, obj := range objectives {
		if obj.ObjectiveType == "deliver" {
			consumeDeliveryItems(userID, obj)
		}
	}

	// Mark quest completed
	now := time.Now().UTC()
	database.DB.Model(&uq).Updates(map[string]interface{}{
		"status":       "completed",
		"completed_at": now,
	})

	// Grant rewards
	var rewards []database.QuestReward
	database.DB.Where("quest_id = ?", q.ID).Find(&rewards)
	grantedSummary := grantRewards(userID, rewards)

	// Unlock any quests that require this one
	var nextQuests []database.Quest
	database.DB.Where("requires_quest_id = ?", q.ID).Find(&nextQuests)
	for _, next := range nextQuests {
		var nextUQ database.UserQuest
		if database.DB.Where("user_id = ? AND quest_id = ?", userID, next.ID).First(&nextUQ).Error == nil {
			database.DB.Model(&nextUQ).Update("status", "available")
		} else {
			database.DB.Create(&database.UserQuest{
				UserID:  userID,
				QuestID: next.ID,
				Status:  "available",
			})
		}
	}

	database.LogActivity(userID, fmt.Sprintf("Completed quest: %s", q.Title))

	return c.JSON(fiber.Map{
		"status":         "completed",
		"quest":          q.Title,
		"completion_text": q.CompletionText,
		"rewards":        grantedSummary,
	})
}

// ── Objective validation ───────────────────────────────────────────────────

func checkObjective(userID uint, obj database.QuestObjective) error {
	switch obj.ObjectiveType {

	case "kill":
		if obj.TargetKey == "" {
			// Any monster — sum all kills
			var total int64
			database.DB.Model(&database.UserMonsterKills{}).
				Where("user_id = ?", userID).
				Select("COALESCE(SUM(total_kills), 0)").
				Scan(&total)
			if int(total) < obj.TargetCount {
				label := obj.DisplayText
				if label == "" {
					label = fmt.Sprintf("Defeat %d enemies", obj.TargetCount)
				}
				return fmt.Errorf("%s (%d/%d)", label, total, obj.TargetCount)
			}
		} else {
			var kills database.UserMonsterKills
			database.DB.Where("user_id = ? AND monster_key = ?", userID, obj.TargetKey).First(&kills)
			if kills.TotalKills < obj.TargetCount {
				label := obj.DisplayText
				if label == "" {
					label = fmt.Sprintf("Kill %d %s", obj.TargetCount, obj.TargetKey)
				}
				return fmt.Errorf("%s (%d/%d)", label, kills.TotalKills, obj.TargetCount)
			}
		}

	case "mine":
		var item database.OreInventoryItem
		database.DB.Joins("JOIN ore_types ON ore_types.id = ore_inventory_items.ore_type_id").
			Where("ore_inventory_items.user_id = ? AND ore_types.ore_key = ?", userID, obj.TargetKey).
			First(&item)
		if item.Quantity < obj.TargetCount {
			label := obj.DisplayText
			if label == "" {
				label = fmt.Sprintf("Mine %d %s", obj.TargetCount, obj.TargetKey)
			}
			return fmt.Errorf("%s (%d/%d)", label, item.Quantity, obj.TargetCount)
		}

	case "gather":
		var item database.HerbInventoryItem
		database.DB.Joins("JOIN herb_types ON herb_types.id = herb_inventory_items.herb_type_id").
			Where("herb_inventory_items.user_id = ? AND herb_types.herb_key = ?", userID, obj.TargetKey).
			First(&item)
		if item.Quantity < obj.TargetCount {
			label := obj.DisplayText
			if label == "" {
				label = fmt.Sprintf("Gather %d %s", obj.TargetCount, obj.TargetKey)
			}
			return fmt.Errorf("%s (%d/%d)", label, item.Quantity, obj.TargetCount)
		}

	case "craft":
		var inv database.UserIngotInventory
		database.DB.Where("user_id = ? AND ingot_key = ?", userID, obj.TargetKey).First(&inv)
		if inv.Quantity < obj.TargetCount {
			label := obj.DisplayText
			if label == "" {
				label = fmt.Sprintf("Craft %d %s", obj.TargetCount, obj.TargetKey)
			}
			return fmt.Errorf("%s (%d/%d)", label, inv.Quantity, obj.TargetCount)
		}

	case "reach_char_level":
		var user database.User
		database.DB.First(&user, userID)
		if user.Level < obj.TargetCount {
			label := obj.DisplayText
			if label == "" {
				label = fmt.Sprintf("Reach character level %d", obj.TargetCount)
			}
			return fmt.Errorf("%s (current: %d)", label, user.Level)
		}

	case "reach_mining_level":
		var skill database.UserMiningSkill
		database.DB.Where("user_id = ?", userID).First(&skill)
		lvl := skill.Level
		if lvl == 0 {
			lvl = 1
		}
		if lvl < obj.TargetCount {
			label := obj.DisplayText
			if label == "" {
				label = fmt.Sprintf("Reach mining level %d", obj.TargetCount)
			}
			return fmt.Errorf("%s (current: %d)", label, lvl)
		}

	case "reach_blacksmith_level":
		var skill database.UserBlacksmithSkill
		database.DB.Where("user_id = ?", userID).First(&skill)
		lvl := skill.Level
		if lvl == 0 {
			lvl = 1
		}
		if lvl < obj.TargetCount {
			label := obj.DisplayText
			if label == "" {
				label = fmt.Sprintf("Reach blacksmith level %d", obj.TargetCount)
			}
			return fmt.Errorf("%s (current: %d)", label, lvl)
		}

	case "deliver":
		// deliver = same as mine check — validated here, consumed after all pass
		var item database.OreInventoryItem
		database.DB.Joins("JOIN ore_types ON ore_types.id = ore_inventory_items.ore_type_id").
			Where("ore_inventory_items.user_id = ? AND ore_types.ore_key = ?", userID, obj.TargetKey).
			First(&item)
		if item.Quantity < obj.TargetCount {
			label := obj.DisplayText
			if label == "" {
				label = fmt.Sprintf("Deliver %d %s", obj.TargetCount, obj.TargetKey)
			}
			return fmt.Errorf("%s (%d/%d)", label, item.Quantity, obj.TargetCount)
		}
	}

	return nil
}

// consumeDeliveryItems deducts items for deliver-type objectives.
func consumeDeliveryItems(userID uint, obj database.QuestObjective) {
	// Ore delivery
	var item database.OreInventoryItem
	if database.DB.Joins("JOIN ore_types ON ore_types.id = ore_inventory_items.ore_type_id").
		Where("ore_inventory_items.user_id = ? AND ore_types.ore_key = ?", userID, obj.TargetKey).
		First(&item).Error == nil {
		database.DB.Exec(
			"UPDATE ore_inventory_items SET quantity = quantity - ?, updated_at = ? WHERE id = ?",
			obj.TargetCount, time.Now().UTC(), item.ID,
		)
	}
}

// ── Reward granting ────────────────────────────────────────────────────────

type rewardSummaryItem struct {
	Type   string `json:"type"`
	Key    string `json:"key,omitempty"`
	Amount int64  `json:"amount,omitempty"`
}

func grantRewards(userID uint, rewards []database.QuestReward) []rewardSummaryItem {
	var summary []rewardSummaryItem

	for _, r := range rewards {
		switch r.RewardType {
		case "xp":
			AwardXP(userID, r.Amount)
			summary = append(summary, rewardSummaryItem{Type: "xp", Amount: r.Amount})

		case "money":
			database.DB.Exec(
				"UPDATE users SET money = money + ? WHERE id = ?",
				r.Amount, userID,
			)
			summary = append(summary, rewardSummaryItem{Type: "money", Amount: r.Amount})

		case "equipment":
			var eq database.Equipment
			if database.DB.Where("equipment_key = ?", r.RewardKey).First(&eq).Error == nil {
				database.DB.Create(&database.UserEquipment{
					UserID:      userID,
					EquipmentID: eq.ID,
					ObtainedAt:  time.Now().UTC(),
				})
				summary = append(summary, rewardSummaryItem{Type: "equipment", Key: r.RewardKey, Amount: 1})
			}
		}
	}

	return summary
}

// ── Admin handlers ─────────────────────────────────────────────────────────

func AdminGetAllQuests(c *fiber.Ctx) error {
	var quests []database.Quest
	database.DB.Order("chapter ASC, sort_order ASC").Find(&quests)
	return c.JSON(quests)
}

func AdminCreateQuest(c *fiber.Ctx) error {
	var q database.Quest
	if err := c.BodyParser(&q); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	database.DB.Create(&q)
	// Make quest available for all existing users if no prerequisite
	if q.RequiresQuestID == 0 {
		var users []database.User
		database.DB.Find(&users)
		for _, u := range users {
			database.DB.Create(&database.UserQuest{UserID: u.ID, QuestID: q.ID, Status: "available"})
		}
	}
	return c.JSON(q)
}

func AdminUpdateQuest(c *fiber.Ctx) error {
	id := c.Params("id")
	var q database.Quest
	if database.DB.First(&q, id).Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	c.BodyParser(&q)
	database.DB.Save(&q)
	return c.JSON(q)
}

func AdminDeleteQuest(c *fiber.Ctx) error {
	id := c.Params("id")
	database.DB.Delete(&database.Quest{}, id)
	return c.JSON(fiber.Map{"status": "deleted"})
}

func AdminGetQuestObjectives(c *fiber.Ctx) error {
	questID := c.Params("id")
	var objs []database.QuestObjective
	database.DB.Where("quest_id = ?", questID).Find(&objs)
	return c.JSON(objs)
}

func AdminCreateQuestObjective(c *fiber.Ctx) error {
	var obj database.QuestObjective
	if err := c.BodyParser(&obj); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	database.DB.Create(&obj)
	return c.JSON(obj)
}

func AdminUpdateQuestObjective(c *fiber.Ctx) error {
	id := c.Params("id")
	var obj database.QuestObjective
	if database.DB.First(&obj, id).Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	c.BodyParser(&obj)
	database.DB.Save(&obj)
	return c.JSON(obj)
}

func AdminDeleteQuestObjective(c *fiber.Ctx) error {
	id := c.Params("id")
	database.DB.Delete(&database.QuestObjective{}, id)
	return c.JSON(fiber.Map{"status": "deleted"})
}

func AdminGetQuestRewards(c *fiber.Ctx) error {
	questID := c.Params("id")
	var rewards []database.QuestReward
	database.DB.Where("quest_id = ?", questID).Find(&rewards)
	return c.JSON(rewards)
}

func AdminCreateQuestReward(c *fiber.Ctx) error {
	var r database.QuestReward
	if err := c.BodyParser(&r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	database.DB.Create(&r)
	return c.JSON(r)
}

func AdminUpdateQuestReward(c *fiber.Ctx) error {
	id := c.Params("id")
	var r database.QuestReward
	if database.DB.First(&r, id).Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	c.BodyParser(&r)
	database.DB.Save(&r)
	return c.JSON(r)
}

func AdminDeleteQuestReward(c *fiber.Ctx) error {
	id := c.Params("id")
	database.DB.Delete(&database.QuestReward{}, id)
	return c.JSON(fiber.Map{"status": "deleted"})
}

// ── Helpers ────────────────────────────────────────────────────────────────

func loadObjectives(questID uint) []QuestObjectiveResponse {
	var objs []database.QuestObjective
	database.DB.Where("quest_id = ?", questID).Find(&objs)
	result := make([]QuestObjectiveResponse, len(objs))
	for i, o := range objs {
		result[i] = QuestObjectiveResponse{
			ID:            o.ID,
			ObjectiveType: o.ObjectiveType,
			TargetKey:     o.TargetKey,
			TargetCount:   o.TargetCount,
			DisplayText:   o.DisplayText,
		}
	}
	return result
}

func loadRewards(questID uint) []QuestRewardResponse {
	var rewards []database.QuestReward
	database.DB.Where("quest_id = ?", questID).Find(&rewards)
	result := make([]QuestRewardResponse, len(rewards))
	for i, r := range rewards {
		result[i] = QuestRewardResponse{
			ID:         r.ID,
			RewardType: r.RewardType,
			RewardKey:  r.RewardKey,
			Amount:     r.Amount,
		}
	}
	return result
}
