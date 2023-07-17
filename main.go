package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func main() {
	// Create a new engine
	engine := html.New("./views", ".html")

	// Pass the engine to the Views
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		// Render index
		return c.Render("index", fiber.Map{
			"Title": "Hello, I am MIMS!",
		})
	})

	app.Get("/main", func(c *fiber.Ctx) error {
		// Render dashboard within layouts/main
		return c.Render("dashboard", fiber.Map{
			"Title": "Dashboard",
		}, "layouts/main")
	})

	app.Get("/main/new-sale", func(c *fiber.Ctx) error {
		return c.Render("new-sale", fiber.Map{
			"Title": "New Sale",
		}, "layouts/main")
	})

	app.Get("/main/sales-history", func(c *fiber.Ctx) error {
		return c.Render("sales-history", fiber.Map{
			"Title": "Sales History",
		}, "layouts/main")
	})

	// Static file server
	app.Static("/static", "./static")

	// Start listening
	log.Fatal(app.Listen(":3000"))
}
