package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
)

//====================================
// 1. Check Login
// 2. If Login pass then generate token
// 3. Check token from Login that is the same as ENV
//
//
//====================================

func main() {

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app := fiber.New()

	// Create Login
	app.Post("/login", login)

	// JWT Middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
	}))

	app.Use(checkMiddleware)

	// Setup routes
	app.Get("/book", getBooks)
	app.Get("/book/:id", getBook)
	app.Post("/book", createBook)
	app.Put("/book/:id", updateBook)
	app.Delete("/book/:id", deleteBook)

	app.Post("/upload", uploadFile)

	// Setup routes
	app.Get("/api/config", getEnv)

	app.Listen(":8080")
}

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Dummy user for example
var memberUser = User{
	Email:    "user@example.com",
	Password: "password123",
}

func login(c *fiber.Ctx) error {
	user := new(User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if user.Email != memberUser.Email || user.Password != memberUser.Password {
		return fiber.ErrUnauthorized
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims) // Get token and encypt
	claims["email"] = user.Email
	claims["role"] = "admin"
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"message": "Login succes",
		"token":   t,
	})
}

func checkMiddleware(c *fiber.Ctx) error {
	// Extract the token from the Fiber context (inserted by the JWT middleware)
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)

	if claims["role"] != "admin" {
		return fiber.ErrUnauthorized
	}

	return c.Next()
}

// loggingMiddleware logs the processing time for each request
// func checkMiddleware(c *fiber.Ctx) error {
// 	// Start timer
// 	start := time.Now()

// 	// Process request
// 	err := c.Next()

// 	// Calculate processing time
// 	duration := time.Since(start)

// 	// Log the information
// 	fmt.Printf("Request URL: %s - Method: %s - Duration: %s\n", c.OriginalURL(), c.Method(), duration)

// 	return err
// }

func uploadFile(c *fiber.Ctx) error {
	file, err := c.FormFile("image")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	err = c.SaveFile(file, "./uploads/"+file.Filename)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	return c.SendString("File upload complete!")
}

// --> Using ENV FILE BY USING GODOTENV
func getEnv(c *fiber.Ctx) error {
	secret := os.Getenv("SECRET")

	if secret == "" {
		secret = "defaultsecret"
	}

	return c.JSON(fiber.Map{
		"SECRET": secret,
	})
}

// --> Using EXPORT KEY
// func getEnv(c *fiber.Ctx) error {
// 	if value, exists := os.LookupEnv("SECRET"); exists {
// 		return c.JSON(fiber.Map{
// 			"SECRET": value,
// 		})
// 	}
// 	return c.JSON(fiber.Map{
// 		"SECRET": "defaultsecret",
// 	})
// }

// func getConfig(c *fiber.Ctx) error {
// 	// Example: Return a configuration value from environment variable
// 	secretKey := getEnv("SECRET_KEY", "defaultSecret")

// 	return c.JSON(fiber.Map{
// 		"secret_key": secretKey,
// 	})
// }
