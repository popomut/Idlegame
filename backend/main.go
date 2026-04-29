package main

import (
        "fmt"
        "log"

        "github.com/gofiber/fiber/v2"
        "github.com/gofiber/fiber/v2/middleware/cors"

        "idlegame-backend/database"
        "idlegame-backend/handlers"
        "idlegame-backend/middleware"
)

func main() {
        // Initialize database
        err := database.Init()
        if err != nil {
                log.Fatalf("Failed to initialize database: %v", err)
        }
        defer database.Close()

        // Create Fiber app
        app := fiber.New(fiber.Config{
                Prefork: false,
        })

        // Middleware
        app.Use(cors.New(cors.Config{
                AllowOriginsFunc: func(origin string) bool { return true },
                AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
                AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
                AllowCredentials: true,
        }))

        // Public routes (no auth required)
        app.Post("/api/auth/register", handlers.Register)
        app.Post("/api/auth/login", handlers.Login)
        app.Post("/api/auth/guest", handlers.GuestLogin)
        app.Post("/api/auth/logout", handlers.Logout)
        app.Get("/api/ore-types", handlers.GetOreTypes)          // Master table — public
        app.Get("/api/monsters", handlers.GetMonsters)            // Monster master table — public
        app.Get("/api/equipment/types", handlers.GetEquipmentTypes) // Equipment master table — public
        app.Get("/api/map/continents", handlers.GetContinents)       // Map master table — public

        // Protected routes (require JWT token)
        api := app.Group("/api", middleware.AuthMiddleware())

        // User routes
        api.Get("/user", handlers.GetUser)
        api.Post("/user/update", handlers.UpdateUser)

        // Character routes
        api.Get("/character", handlers.GetCharacter)
        api.Post("/character/heal", handlers.HealHP)

        // Mining routes
        api.Post("/mining/start", handlers.StartMining)
        api.Post("/mining/stop", handlers.StopMining)
        api.Get("/mining/status", handlers.GetMiningStatus)

        // Inventory routes
        api.Get("/inventory/ores", handlers.GetOreInventory)

        // Equipment routes
        api.Get("/equipment/bag", handlers.GetEquipmentBag)
        api.Get("/equipment/slots", handlers.GetEquippedSlots)
        api.Post("/equipment/equip", handlers.EquipItem)
        api.Post("/equipment/unequip", handlers.UnequipSlot)
        api.Post("/equipment/give", handlers.GiveEquipment)

        // Map / combat routes
        api.Post("/map/enter", handlers.EnterArea)
        api.Get("/map/session", handlers.GetCombatSession)
        api.Post("/map/resume", handlers.ResumeCombat)
        api.Post("/map/advance", handlers.AdvanceFight)
        api.Post("/map/flee", handlers.FleeCombat)

        // Start server
        port := 5000
        fmt.Printf("🚀 Server running on http://0.0.0.0:%d\n", port)
        log.Fatal(app.Listen(fmt.Sprintf("0.0.0.0:%d", port)))
}

