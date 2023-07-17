package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

const apiServerAddr string = "http://104.248.98.237:3000/"

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

	// New Sale
	app.Get("/main/new-sale", func(c *fiber.Ctx) error {
		return c.Render("new-sale", fiber.Map{
			"Title": "New Sale",
		}, "layouts/main")
	})

	// POST New Sale
	app.Post("/main/new-sale/", func(c *fiber.Ctx) error {
		ns := new(NewSaleFormData)
		if err := c.BodyParser(ns); err != nil {
			return err
		}

		var amt int
		var qty int
		var url string
		var itemId int
		paymentType, _ := parsePaymentMethod(ns.PaymentMethod)
		operationId := 1

		//i need to change DB structure to accommodate ever growing product list.
		//this is hardcoded now, since we only selling 1 product. Its ok for now...
		if ns.Qty_FreshJuice > 0 {
			amt = ns.Qty_FreshJuice * 8
			qty = ns.Qty_FreshJuice
			itemId = 1
			url = apiServerAddr + "sales/new/" +
				strconv.Itoa(amt) + "-" + strconv.Itoa(qty) + "-" + strconv.Itoa(paymentType) + "-" + strconv.Itoa(operationId) + "-" + strconv.Itoa(itemId)
		}

		client := http.Client{
			Timeout: time.Second * 2, // Timeout after 2 seconds
		}

		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			log.Println("Post request not completed -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/new-sale")
		}

		req.Header.Set("User-Agent", "mims-app")
		res, err := client.Do(req)
		if err != nil {
			log.Println("Error occured while awaiting response -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/new-sale")
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		body, err := ioutil.ReadAll(res.Body)
		_ = body
		if err != nil {
			log.Println("Error reading response body -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/new-sale")
		}

		//redirect to /main/sales-history w/ toast saying sale successfully registered
		//after i create the toast, i can always redirect to "/main/new-sale" if need be instead of "/main/sales-history"
		return c.Redirect("/main/sales-history")

	})

	app.Get("/main/sales-history", func(c *fiber.Ctx) error {
		return c.Render("sales-history", fiber.Map{
			"Title": "Sales History",
		}, "layouts/main")
	})

	// Static file server
	app.Static("/static", "./static")

	// Http server
	log.Fatal(app.Listen(":3000"))
}

func parsePaymentMethod(paymentType string) (int, error) {
	switch {
	case paymentType == "Cash":
		return 1, nil
	case paymentType == "QR - Maybank":
		return 2, nil
	case paymentType == "QR - Touch & Go":
		return 3, nil
	case paymentType == "Free":
		return 99, nil
	default:
		return 0, errors.New("Unable to parse payment method")
	}
}

type NewSaleFormData struct {
	Sale_TimeDate  string `json:"sale_time_date" xml:"sale_time_date" form:"sale_time_date"`
	PaymentMethod  string `json:"payment_type" xml:"payment_type" form:"payment_type"`
	Qty_FreshJuice int    `json:"fresh_juice_qty" xml:"fresh_juice_qty" form:"fresh_juice_qty"`
	Qty_CutFruit   int    `json:"cut_fruit_qty" xml:"cut_fruit_qty" form:"cut_fruit_qty"`
	Qty_RawFruit   int    `json:"raw_fruit_qty" xml:"raw_fruit_qty" form:"raw_fruit_qty"`
}
