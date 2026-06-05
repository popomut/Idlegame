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
                AllowCredentials: false,
        }))

        // Public routes (no auth required)
        app.Post("/api/auth/register", handlers.Register)
        app.Post("/api/auth/login", handlers.Login)
        app.Post("/api/auth/guest", handlers.GuestLogin)
        app.Post("/api/auth/logout", handlers.Logout)
        app.Get("/api/extractable-types", handlers.GetExtractableTypes) // Extraction types master table — public
        app.Get("/api/ore-types", handlers.GetOreTypes)          // Ore master table — public
        app.Get("/api/herb-types", handlers.GetHerbTypes)        // Herb master table — public
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
        api.Get("/mining/skill", handlers.GetMiningSkill)

        // Blacksmith routes
        api.Post("/blacksmith/start", handlers.StartCrafting)
        api.Post("/blacksmith/stop", handlers.StopCrafting)
        api.Get("/blacksmith/status", handlers.GetCraftingStatus)
        api.Get("/blacksmith/skill", handlers.GetBlacksmithSkill)
        api.Get("/blacksmith/recipes", handlers.GetCraftableItems)
        api.Get("/blacksmith/inventory", handlers.GetIngotInventory)
        api.Get("/blacksmith/inventory/potions", handlers.GetPotionInventory)

        // Inventory routes
        api.Get("/inventory/ores", handlers.GetOreInventory)
        api.Get("/inventory/herbs", handlers.GetHerbInventory)

        // Equipment routes
        api.Get("/equipment/bag", handlers.GetEquipmentBag)
        api.Get("/equipment/slots", handlers.GetEquippedSlots)
        api.Post("/equipment/equip", handlers.EquipItem)
        api.Post("/equipment/unequip", handlers.UnequipSlot)
        api.Post("/equipment/sell", handlers.SellEquipment)
        api.Post("/equipment/give", handlers.GiveEquipment)

        // Map routes (legacy turn-based combat)
        api.Post("/map/enter", handlers.EnterArea)
        api.Get("/map/session", handlers.GetCombatSession)
        api.Post("/map/resume", handlers.ResumeCombat)
        api.Post("/map/advance", handlers.AdvanceFight)
        api.Post("/map/flee", handlers.FleeMapCombat)

        // Active combat (background, server-authoritative)
        api.Post("/combat/start", handlers.StartCombat)
        api.Get("/combat/status", handlers.GetCombatStatus)
        api.Post("/combat/flee", handlers.FleeCombat)

        // Admin equipment management (development only — delete before production)
        api.Get("/admin/equipment", handlers.AdminGetAllEquipment)
        api.Post("/admin/equipment", handlers.AdminCreateEquipment)
        api.Put("/admin/equipment/:id", handlers.AdminUpdateEquipment)
        api.Delete("/admin/equipment/:id", handlers.AdminDeleteEquipment)

        // Admin monster management (development only — delete before production)
        api.Get("/admin/monsters", handlers.AdminGetAllMonsters)
        api.Post("/admin/monsters", handlers.AdminCreateMonster)
        api.Put("/admin/monsters/:id", handlers.AdminUpdateMonster)
        api.Delete("/admin/monsters/:id", handlers.AdminDeleteMonster)

        // Admin ores management (development only — delete before production)
        api.Get("/admin/ores", handlers.AdminGetAllOres)
        api.Post("/admin/ores", handlers.AdminCreateOre)
        api.Put("/admin/ores/:id", handlers.AdminUpdateOre)
        api.Delete("/admin/ores/:id", handlers.AdminDeleteOre)

        // Admin monster drops management (development only — delete before production)
        api.Get("/admin/monster-drops", handlers.AdminGetAllMonsterDrops)
        api.Post("/admin/monster-drops", handlers.AdminCreateMonsterDrop)
        api.Put("/admin/monster-drops/:id", handlers.AdminUpdateMonsterDrop)
        api.Delete("/admin/monster-drops/:id", handlers.AdminDeleteMonsterDrop)

        // Admin mining levels management (development only — delete before production)
        api.Get("/admin/mining-levels", handlers.AdminGetMiningLevels)
        api.Post("/admin/mining-levels", handlers.AdminCreateMiningLevel)
        api.Put("/admin/mining-levels/:level", handlers.AdminUpdateMiningLevel)
        api.Delete("/admin/mining-levels/:level", handlers.AdminDeleteMiningLevel)

        // Admin blacksmith management (development only — delete before production)
        api.Get("/admin/craftable-items", handlers.AdminGetCraftableItems)
        api.Post("/admin/craftable-items", handlers.AdminCreateCraftableItem)
        api.Put("/admin/craftable-items/:id", handlers.AdminUpdateCraftableItem)
        api.Delete("/admin/craftable-items/:id", handlers.AdminDeleteCraftableItem)

        // Admin blacksmith levels management (development only — delete before production)
        api.Get("/admin/blacksmith-levels", handlers.AdminGetBlacksmithLevels)
        api.Post("/admin/blacksmith-levels", handlers.AdminCreateBlacksmithLevel)
        api.Put("/admin/blacksmith-levels/:level", handlers.AdminUpdateBlacksmithLevel)
        api.Delete("/admin/blacksmith-levels/:level", handlers.AdminDeleteBlacksmithLevel)

        // Start server
        port := 5000
        fmt.Printf("🚀 Server running on http://0.0.0.0:%d\n", port)
        log.Fatal(app.Listen(fmt.Sprintf("0.0.0.0:%d", port)))
}

