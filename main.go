package main

import (
	"context"
	"log"

	"github.com/Loukay/tipsee-api/pagination"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/joho/godotenv"
)

func main() {

	ctx := context.Background()

	if godotenv.Load() != nil {
		log.Print("No .env file found. Using process envrionment variables...")
	}

	redis, err := RedisClient(&ctx)

	if err != nil {
		panic("There was a problem connecting to the Redis server.")
	}

	buildIndexes(&ctx, redis)

	var app *fiber.App = fiber.New(fiber.Config{
		Prefork: false,
	})

	controller := Controller{
		Redis: redis,
		Ctx:   &ctx,
	}

	app.Use(cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Path() == "/monitor"
		},
		KeyGenerator: func(c *fiber.Ctx) string {
			log.Print(c.OriginalURL())
			return utils.CopyString(c.OriginalURL())
		},
	}))

	app.Use(cors.New())
	app.Use(pagination.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON("The CocktailDB Cache")
	})

	app.Get("/monitor", monitor.New())

	app.Get("/ingredients", controller.GetRecords("ingredients"))
	app.Get("/alcohols", controller.GetRecords("alcohols"))
	app.Get("/cocktails", controller.GetRecords("cocktails"))

	err = app.Listen(":3000")

	if err != nil {
		log.Fatal("Failed to listen to web server")
		panic(err)
	}

}

func buildIndexes(ctx *context.Context, redis *redis.Client) {
	_, err :=
		redis.Do(*ctx, "FT.CREATE", "idx:ingredients",
			"ON", "hash",
			"PREFIX", "1", "ingredient:",
			"SCHEMA",
			"name", "TEXT",
			"type", "TEXT").Result()

	if err != nil {
		log.Printf("Couldn't create ingredients index %v", err)
	}

	_, err =
		redis.Do(*ctx, "FT.CREATE", "idx:alcohols",
			"ON", "hash",
			"PREFIX", "1", "alcohol:",
			"SCHEMA",
			"name", "TEXT",
			"type", "TEXT").Result()

	if err != nil {
		log.Printf("Couldn't create alcohols index %v", err)
	}

	_, err =
		redis.Do(*ctx, "FT.CREATE", "idx:cocktails",
			"ON", "hash",
			"PREFIX", "1", "cocktail:",
			"SCHEMA",
			"name", "TEXT",
			"category", "TEXT",
			"ingredients", "TAG").Result()
	if err != nil {
		log.Printf("Couldn't create coocktails index %v", err)
	}
}
