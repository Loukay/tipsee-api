package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

// Controller has methods for fetching ingredients and cocktails from Redis
type Controller struct {
	Ctx   *context.Context
	Redis *redis.Client
}

// GetRecords fetches all records (ingredients, alcohols or cocktails) from Redis
func (controller Controller) GetRecords(c *fiber.Ctx) error {
	path := c.Path()
	var key string

	switch path {
	case "/ingredients":
		key = "ingredients"
	case "/alcohols":
		key = "alcohols"
	}

	results, err := controller.Redis.Do(*controller.Ctx, "FT.SEARCH", "idx:"+key, "*", "LIMIT", c.Locals("offset"), c.Locals("limit")).Slice()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	count, results := FormatRedisOutput(results)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"count": count, key: results,
	})
}

// GetCocktails fetches all cocktails from Redis
func (controller Controller) GetCocktails(c *fiber.Ctx) error {

	tagsQuery := c.Query("ingredients")

	var redisQuery string

	if tagsQuery != "" {
		tags := strings.Split(tagsQuery, ",")
		redisQuery = FormatRedisTagsQuery("ingredients", tags)
	} else {
		redisQuery = "*"
	}

	results, err := controller.Redis.Do(*controller.Ctx, "FT.SEARCH", "idx:cocktails", redisQuery, "LIMIT", c.Locals("offset"), c.Locals("limit")).Slice()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	count, results := FormatRedisOutput(results)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"count": count, "cocktails": results,
	})

}

// FormatRedisOutput formats the output of a Redis FT.SEARCH query
func FormatRedisOutput(output []interface{}) (int64, []interface{}) {

	var results []interface{}

	length := len(output)

	for i := 2; i < length; i += 2 {
		current := output[i].([]interface{})
		size := len(current)
		result := map[string]string{}
		for j := 0; j < size; j += 2 {
			key := current[j].(string)
			value := current[j+1].(string)
			result[key] = value
		}
		results = append(results, result)
	}

	return output[0].(int64), results

}

// FormatRedisTagsQuery format a slice of tag values to a usable FT.SEARCH query
func FormatRedisTagsQuery(tag string, values []string) string {
	output := ""
	for _, value := range values {
		output += fmt.Sprintf("@%s:{ %s } ", tag, escapeRedisTag(value))
	}
	return output
}

// espaceRedisTag escapes some characters that break Redis queries
func escapeRedisTag(tag string) string {
	replacer := strings.NewReplacer(
		"-", "\\-",
		"(", "\\(",
		")", "\\)",
		"{", "\\{",
		"}", "\\}",
		"[", "\\[",
		"]", "\\]",
	)
	return replacer.Replace(tag)
}
