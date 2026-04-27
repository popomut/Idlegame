package handlers

import (
	"github.com/gofiber/fiber/v2"
	"idlegame-backend/database"
	"idlegame-backend/utils"
	"time"
)

// RegisterRequest represents user registration data
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents user login data
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse returns user info (token sent as httpOnly cookie)
type AuthResponse struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// createInitialOreInventory seeds copper and iron pivot rows for a new user
func createInitialOreInventory(userID uint) {
	initialOres := map[string]int{
		"copper_ore": 5,
		"iron_ore":   2,
	}
	for oreKey, qty := range initialOres {
		var oreType database.OreType
		if err := database.DB.Where("ore_key = ?", oreKey).First(&oreType).Error; err == nil {
			database.DB.Create(&database.OreInventoryItem{
				UserID:    userID,
				OreTypeID: oreType.ID,
				Quantity:  qty,
			})
		}
	}
}

// Register creates a new user account
func Register(c *fiber.Ctx) error {
	req := new(RegisterRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	// Validate input
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing required fields"})
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash password"})
	}

	// Create user
	user := database.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	result := database.DB.Create(&user)
	if result.Error != nil {
		// Check if user already exists
		if result.Error.Error() == "UNIQUE constraint failed: users.username" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "username already taken"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create user"})
	}

	// Create ore inventory for new user
	createInitialOreInventory(user.ID)

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate token"})
	}

	// Set httpOnly cookie (cannot be accessed by JavaScript)
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * 7 * time.Hour), // 7 days
		HTTPOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: "Lax",
	})

	return c.Status(fiber.StatusCreated).JSON(AuthResponse{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}

// Login authenticates a user
func Login(c *fiber.Ctx) error {
	req := new(LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	// Find user by username
	var user database.User
	result := database.DB.Where("username = ?", req.Username).First(&user)
	if result.Error != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid username or password"})
	}

	// Verify password
	if !utils.VerifyPassword(user.Password, req.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid username or password"})
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate token"})
	}

	// Set httpOnly cookie
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * 7 * time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})

	return c.JSON(AuthResponse{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}

// GuestLogin creates an anonymous guest account
func GuestLogin(c *fiber.Ctx) error {
	// Generate unique username for guest
	guestUsername := "Guest_" + utils.GenerateRandomID(8)
	
	// Create guest user (no password needed)
	user := database.User{
		Username: guestUsername,
		Email:    guestUsername + "@guest.local",
		Password: "", // No password for guests
		IsGuest:  true,
	}
	
	result := database.DB.Create(&user)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create guest account"})
	}
	
	// Create initial ore inventory
	createInitialOreInventory(user.ID)
	
	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate token"})
	}
	
	// Set httpOnly cookie
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * 7 * time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})
	
	return c.Status(fiber.StatusCreated).JSON(AuthResponse{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}

// Logout clears the httpOnly auth cookie
func Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})
	
	return c.JSON(fiber.Map{"message": "logged out successfully"})
}
