package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Item struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Qty  int    `json:"qty"`
}

var items = []Item{
	{
		Id:   1,
		Name: "Shampoo",
		Qty:  15,
	},
	{
		Id:   2,
		Name: "Snack",
		Qty:  150,
	},
	{
		Id:   3,
		Name: "Cigaret",
		Qty:  90,
	},
}

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello world!")
	})

	app.Post("/items", func(c *fiber.Ctx) error {
		item := Item{}
		err := c.BodyParser(&item)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
		}

		items = append(items, item)

		return c.JSON(fiber.Map{
			"success": true,
			"message": "",
			"data":    item,
		})
	})

	app.Get("/items", func(c *fiber.Ctx) error {
		name := c.Query("name")

		if name != "" {
			for _, item := range items {
				if item.Name == name {
					return c.JSON(fiber.Map{
						"success": true,
						"message": "",
						"data":    []Item{item},
					})
				}
			}

			return c.JSON(fiber.Map{
				"success": true,
				"message": fmt.Sprintf("Items with name %s not found", name),
				"data":    []Item{},
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "",
			"data":    items,
		})
	})

	app.Get("/items/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
		}

		for _, item := range items {
			if item.Id == id {
				return c.JSON(fiber.Map{
					"success": true,
					"message": "",
					"data":    item,
				})
			}
		}

		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Item with ID %d not found", id),
			"data":    nil,
		})
	})

	app.Put("/items/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
		}

		item := Item{}
		err = c.BodyParser(&item)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
		}

		if item.Name == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Name is required",
				"data":    nil,
			})
		}

		for i := 0; i < len(items); i++ {
			if items[i].Id == id {
				items[i].Name = item.Name

				return c.JSON(fiber.Map{
					"success": true,
					"message": "",
					"data":    items[i],
				})
			}
		}

		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Item with ID %d not found", id),
			"data":    nil,
		})
	})

	app.Delete("/items/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
		}

		for i, item := range items {
			if item.Id == id {
				items = append(items[:i], items[i+1:]...)
				return c.JSON(fiber.Map{
					"success": true,
					"message": "",
					"data":    items,
				})
			}
		}

		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Item with ID %d not found", id),
			"data":    nil,
		})
	})

	app.Listen(":8080")
}
