package main

import (
	"log"

	"github.com/CRTOsp3ck/mims-app/handler"
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

	// --> Landing
	// Home
	app.Get("/", handler.Landing)
	// Dashboard
	app.Get("/main", handler.Dashboard)

	// --> Auth
	// Auth - Login page
	app.Get("/main/login", handler.LoginPage)
	// Auth - Login request
	app.Post("/auth/login", handler.LoginRequest)
	// Auth - Logout
	app.Post("/auth/logout", handler.LogoutRequest)

	// --> Sales
	// New Sale
	app.Get("/main/new-sale", handler.NewSale)
	// POST New Sale
	app.Post("/main/new-sale/", handler.NewSaleRequest)
	// Sales history
	app.Get("/main/sales-history", handler.SalesHistory)
	// Sales report
	app.Get("/main/sales-report", handler.SalesReport)
	// POST Update periodic sales report
	app.Post("/main/sales-report/update-periodic", handler.SalesReportUpdatePeriodic)

	// --> Purchases
	// Add purchase
	app.Get("/main/add-purchase", handler.AddPurchase)
	// List purchase
	app.Get("/main/purchase-history", handler.ListPurchase)

	// Static file server
	app.Static("/static", "./static")

	// Http server
	log.Fatal(app.Listen(":3000"))
}
