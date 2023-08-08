package handler

import (
	"github.com/CRTOsp3ck/mims-app/helper"
	"github.com/gofiber/fiber/v2"
)

func AddPurchase(c *fiber.Ctx) error {
	if !helper.CheckAuthState(c) {
		return c.Render("login", fiber.Map{
			"Title": "Login",
		})
	}

	//pass it to the renderer
	return c.Render("add-purchase", fiber.Map{
		"Title": "Add Purchase",
	}, "layouts/main")
}

func ListPurchase(c *fiber.Ctx) error {
	if !helper.CheckAuthState(c) {
		return c.Render("login", fiber.Map{
			"Title": "Login",
		})
	}

	//pass it to the renderer
	return c.Render("purchase-history", fiber.Map{
		"Title": "List Purchase",
	}, "layouts/main")
}
