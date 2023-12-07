package main

import (
	"main/database"
	"main/router"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)



func hello (c *fiber.Ctx) error {
	return c.SendString("Hello, World!")
}

func main(){
	app:=fiber.New()
	app.Use(cors.New())

	database.ConnectDB()

	router.SetupRoutes(app)
	app.Listen(":3000")
}