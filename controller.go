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
func (controller Controller) GetRecords(key string) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		searchQuery := escapeRedisTag(c.Query("search"))

		queryString := searchQuery + "* "

		if c.Query("ingredients") != "" && searchQuery == "" {
			queryString = ""
		}

		if key == "cocktails" && c.Query("ingredients") != "" {
			tags := strings.Split(c.Query("ingredients"), ",")
			queryString += FormatRedisTagsQuery("ingredients", tags)
		}

		queryArgs := []interface{}{
			"FT.SEARCH",
			"idx:" + key,
			queryString,
			"LIMIT",
			c.Locals("offset"),
			c.Locals("limit"),
		}

		results, err := controller.Redis.Do(*controller.Ctx, queryArgs...).Slice()

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		count, results := FormatRedisOutput(results)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"count": count, key: results,
		})
	}
}

// FormatRedisOutput formats the output of a Redis FT.SEARCH query
func FormatRedisOutput(output []interface{}) (int64, []interface{}) {

	results := make([]interface{}, 0)

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
		"*", "\\*",
	)
	return replacer.Replace(tag)
}
