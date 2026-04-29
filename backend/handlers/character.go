package handlers

import (
	"github.com/gofiber/fiber/v2"
	"idlegame-backend/database"
)

// CharacterResponse is the full character data returned to the frontend
type CharacterResponse struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	PlayerName  string `json:"player_name"`
	PlayerClass string `json:"player_class"`
	Level       int    `json:"level"`
	XP          int64  `json:"xp"`
	XPRequired  int64  `json:"xp_required"` // XP needed to reach NEXT level
	HP          int    `json:"hp"`
	MaxHP       int    `json:"max_hp"`
	Stamina     int    `json:"stamina"`
	MaxStamina  int    `json:"max_stamina"`
	Str         int    `json:"str"`
	Int         int    `json:"int"`
	Dex         int    `json:"dex"`
	Money       int64  `json:"money"`
}

// GetCharacter returns full character stats for the logged-in player
func GetCharacter(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var user database.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	xpRequired := xpForNextLevel(user.Level)

	return c.JSON(CharacterResponse{
		UserID:      user.ID,
		Username:    user.Username,
		PlayerName:  user.PlayerName,
		PlayerClass: user.PlayerClass,
		Level:       user.Level,
		XP:          user.XP,
		XPRequired:  xpRequired,
		HP:          user.HP,
		MaxHP:       user.MaxHP,
		Stamina:     user.Stamina,
		MaxStamina:  user.MaxStamina,
		Str:         user.Str,
		Int:         user.Int,
		Dex:         user.Dex,
		Money:       user.Money,
	})
}

// HealHP persists the player's current HP during regen (called periodically by the client).
// POST /api/character/heal  body: { hp: int }
func HealHP(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var body struct {
		HP int `json:"hp"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	var user database.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	// Clamp to valid range
	hp := body.HP
	if hp < 0 {
		hp = 0
	}
	if hp > user.MaxHP {
		hp = user.MaxHP
	}

	database.DB.Model(&user).Update("hp", hp)
	return c.JSON(fiber.Map{"hp": hp})
}

func xpForNextLevel(currentLevel int) int64 {
	nextLevel := currentLevel + 1
	var cl database.CharacterLevel
	if err := database.DB.First(&cl, nextLevel).Error; err != nil {
		// At max level — return a very large number
		return 999999999
	}
	return cl.XPRequired
}

// AwardXP adds XP to a user and handles level-up if threshold is crossed.
// Called internally by mining, combat, etc.
func AwardXP(userID uint, amount int64) {
	if amount <= 0 {
		return
	}

	var user database.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return
	}

	user.XP += amount

	// Check for level-up (loop in case of multiple levels gained at once)
	for {
		nextLevel := user.Level + 1
		var nextCL database.CharacterLevel
		if err := database.DB.First(&nextCL, nextLevel).Error; err != nil {
			break // Already at max level
		}
		if user.XP < nextCL.XPRequired {
			break
		}

		// Level up!
		user.Level = nextLevel
		user.MaxHP = nextCL.MaxHP
		user.HP = nextCL.MaxHP // restore HP on level-up
		user.MaxStamina = nextCL.MaxStamina
		user.Stamina = nextCL.MaxStamina
		user.Str = nextCL.Str
		user.Int = nextCL.Int
		user.Dex = nextCL.Dex

		database.LogActivity(userID, "Level up! Now Rank "+itoa(nextLevel))
	}

	database.DB.Save(&user)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	digits := "0123456789"
	for n > 0 {
		result = string(digits[n%10]) + result
		n /= 10
	}
	if neg {
		result = "-" + result
	}
	return result
}
