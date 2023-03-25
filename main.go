package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Item struct {
	gorm.Model
	Name   string `json:"name"`
	Qty    int    `json:"qty"`
	UserID int    `json:"user_id"`
	User   User   `json:"user"`
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
	DB.AutoMigrate(&User{}, &Item{})

	app := fiber.New()

	app.Use(cors.New())

	authMiddleware := func(c *fiber.Ctx) error {
		tokenHeaders := c.Get("Authorization")
		if tokenHeaders == "" || !strings.HasPrefix(tokenHeaders, "Bearer ") {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "",
				"data":    nil,
			})
		}

		token, err := jwt.Parse(strings.Split(tokenHeaders, " ")[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte("JWT_SECRET"), nil
		})
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "",
				"data":    nil,
			})
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Locals("user_id", claims["user_id"])
		} else {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "",
				"data":    nil,
			})
		}

		return c.Next()
	}

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello world 2!")
	})

	app.Post("/register", func(c *fiber.Ctx) error {
		user := User{}
		err := c.BodyParser(&user)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
		}

		err = DB.Create(&user).Error
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "",
			"data":    user,
		})
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		userRequest := User{}
		err := c.BodyParser(&userRequest)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
		}

		user := User{}
		result := DB.Where("email = ?", userRequest.Email).First(&user)
		if result.RowsAffected == 0 {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Email or password invalid",
				"data":    nil,
			})
		}

		if userRequest.Password != user.Password {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Email or password invalid",
				"data":    nil,
			})
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user.ID,
		})

		tokenString, err := token.SignedString([]byte("JWT_SECRET"))
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "",
			"data":    tokenString,
		})
	})

	app.Post("/items", authMiddleware, func(c *fiber.Ctx) error {
		item := Item{}
		err := c.BodyParser(&item)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
		}

		item.UserID = int(c.Locals("user_id").(float64))

		err = DB.Create(&item).Error
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "",
			"data":    item,
		})
	})

	app.Get("/items", authMiddleware, func(c *fiber.Ctx) error {
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

		err := DB.Preload("User").Where("user_id = ?", c.Locals("user_id")).Find(&items).Error
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "",
			"data":    items,
		})
	})

	app.Get("/items/:id", authMiddleware, func(c *fiber.Ctx) error {
		item := Item{}
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
		}

		result := DB.Preload("User").Where("user_id = ?", c.Locals("user_id")).First(&item, id)
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

	app.Put("/items/:id", authMiddleware, func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
		}

		itemRequest := Item{}
		err = c.BodyParser(&itemRequest)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
		}

		if itemRequest.Name == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Name is required",
				"data":    nil,
			})
		}

		item := Item{}
		result := DB.Preload("User").Where("user_id = ?", c.Locals("user_id")).First(&item, id)
		if result.RowsAffected == 0 {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": fmt.Sprintf("Item with ID %d not found", id),
				"data":    nil,
			})
		}

		itemRequest.ID = item.ID
		itemRequest.UserID = item.UserID

		result = DB.Save(&itemRequest)
		if result.RowsAffected != 0 {
			return c.JSON(fiber.Map{
				"success": true,
				"message": "",
				"data":    itemRequest,
			})
		}

		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Item with ID %d not found", id),
			"data":    nil,
		})
	})

	app.Delete("/items/:id", authMiddleware, func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
		}

		result := DB.Where("user_id = ?", c.Locals("user_id")).Delete(&Item{}, id)
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
