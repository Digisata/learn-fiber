package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Item struct {
	gorm.Model
	Name string `json:"name"`
	Qty  int    `json:"qty"`
}

var DB *gorm.DB

func init() {
	dsn := "host=localhost user=postgres password=nopalgemay32 dbname=learn_gorm port=5431 sslmode=disable TimeZone=Asia/Jakarta"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	DB.AutoMigrate(&Item{})

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello world 2!")
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

		DB.Create(&item)

		return c.JSON(fiber.Map{
			"success": true,
			"message": "",
			"data":    item,
		})
	})

	app.Get("/items", func(c *fiber.Ctx) error {
		name := c.Query("name")
		items := []Item{}

		if name != "" {
			DB.Where("name LIKE ?", fmt.Sprintf("%%%s%%", name)).Find(&items)
			if len(items) > 0 {
				return c.JSON(fiber.Map{
					"success": true,
					"message": "",
					"data":    items,
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"message": fmt.Sprintf("Items with name %s not found", name),
				"data":    items,
			})
		}

		DB.Find(&items)

		return c.JSON(fiber.Map{
			"success": true,
			"message": "",
			"data":    items,
		})
	})

	app.Get("/items/:id", func(c *fiber.Ctx) error {
		item := Item{}
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
		}

		result := DB.First(&item, id)
		if result.RowsAffected != 0 {
			return c.JSON(fiber.Map{
				"success": true,
				"message": "",
				"data":    item,
			})
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

		item.ID = uint(id)

		if item.Name == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Name is required",
				"data":    nil,
			})
		}

		result := DB.Save(&item)
		if result.RowsAffected != 0 {
			return c.JSON(fiber.Map{
				"success": true,
				"message": "",
				"data":    item,
			})
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

		result := DB.Delete(&Item{}, id)
		if result.RowsAffected != 0 {
			return c.JSON(fiber.Map{
				"success": true,
				"message": "",
				"data":    nil,
			})
		}

		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Item with ID %d not found", id),
			"data":    nil,
		})
	})

	app.Listen(":8080")
}
