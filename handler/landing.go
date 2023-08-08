package handler

import (
	"github.com/CRTOsp3ck/mims-app/helper"
	"github.com/gofiber/fiber/v2"
)

func Landing(c *fiber.Ctx) error {
	// Render index
	return c.Render("index", fiber.Map{
		"Title": "Hello, I am MIMS!",
	})
}

func Dashboard(c *fiber.Ctx) error {
	if !helper.CheckAuthState(c) {
		return c.Render("login", fiber.Map{
			"Title": "Login",
		})
	}

	// Render dashboard within layouts/main
	return c.Render("dashboard", fiber.Map{
		"Title": "Dashboard",
	}, "layouts/main")
}
