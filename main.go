package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
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
		ns := new(FormDataNewSale)
		if err := c.BodyParser(ns); err != nil {
			return err
		}

		var amt int
		var qty int
		var url string
		var itemId int
		paymentType, _ := parsePaymentMethodToInt(ns.PaymentMethod)
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
		url := apiServerAddr + "sales/find/"

		client := http.Client{
			Timeout: time.Second * 2, // Timeout after 2 seconds
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Println("Post request not completed -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/sales-history")
		}

		req.Header.Set("User-Agent", "mims-app")
		res, err := client.Do(req)
		if err != nil {
			log.Println("Error occured while awaiting response -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/sales-history")
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		body, err := ioutil.ReadAll(res.Body)
		_ = body
		if err != nil {
			log.Println("Error reading response body -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/sales-history")
		}

		var jsonSales struct {
			Sales []JsonSale `json:"sales"`
		}

		//convert that string (body) to json
		if err := json.Unmarshal(body, &jsonSales); err != nil {
			log.Println("Error unmarshalling body into JSON -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/sales-history")
		}

		//create an array of view sales with that json information..
		viewSales := []*ViewSale{}

		for index := range jsonSales.Sales {
			//parsing some stuff before hand
			paymentType, _ := parsePaymentMethodToString(jsonSales.Sales[index].PaymentType)
			operation, _ := parseOperationToString(jsonSales.Sales[index].OperationID)
			item, _ := parseItemToString(jsonSales.Sales[index].ItemID)
			time := strings.Split(strings.Split(jsonSales.Sales[index].CreatedAt.String(), " ")[1], ".")[0]
			date := strings.Split(jsonSales.Sales[index].CreatedAt.String(), " ")[0]

			viewSale := ViewSale{
				ID:          jsonSales.Sales[index].ID,
				Amount:      "RM" + strconv.FormatFloat(float64(jsonSales.Sales[index].Amount), 'f', -1, 64),
				Qty:         strconv.FormatFloat(float64(jsonSales.Sales[index].Amount), 'f', -1, 64) + " unit(s)",
				PaymentType: paymentType,
				Operation:   operation,
				Item:        item,
				Time:        time,
				Date:        date,
			}
			viewSales = append(viewSales, &viewSale)
		}

		viewSales = reverseSales(viewSales)

		//**Test sales data to see if it renders properly via the html template**
		// sales := []Sale{
		// 	{
		// 		ID:          1,
		// 		Amount:      16.00,
		// 		Qty:         2,
		// 		PaymentType: 1,
		// 		OperationID: 1,
		// 		ItemID:      1,
		// 		CreatedAt:   time.Now().Truncate(time.Duration(time.Now().Day())),
		// 		UpdatedAt:   time.Now().Truncate(time.Duration(time.Now().Day())),
		// 	},
		// 	{
		// 		ID:          2,
		// 		Amount:      24.00,
		// 		Qty:         3,
		// 		PaymentType: 2,
		// 		OperationID: 1,
		// 		ItemID:      1,
		// 		CreatedAt:   time.Now().Truncate(time.Duration(time.Now().Day())),
		// 		UpdatedAt:   time.Now().Truncate(time.Duration(time.Now().Day())),
		// 	},
		// }

		//pass it to the renderer
		return c.Render("sales-history", fiber.Map{
			"Title": "Sales History",
			"Sales": viewSales,
		}, "layouts/main")
	})

	// Static file server
	app.Static("/static", "./static")

	// Http server
	log.Fatal(app.Listen(":3000"))
}

func reverseSales(input []*ViewSale) []*ViewSale {
	if len(input) == 0 {
		return input
	}
	return append(reverseSales(input[1:]), input[0])
}

func parsePaymentMethodToInt(paymentType string) (int, error) {
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

func parsePaymentMethodToString(paymentType int) (string, error) {
	switch {
	case paymentType == 1:
		return "Cash", nil
	case paymentType == 2:
		return "QR - Maybank", nil
	case paymentType == 3:
		return "QR - Touch & Go", nil
	case paymentType == 99:
		return "Free", nil
	default:
		return "", errors.New("Unable to parse payment method")
	}
}

func parseOperationToString(operationId int) (string, error) {
	switch {
	case operationId == 1:
		return "Kebun Che Mah, Kemensah", nil
	default:
		return "", errors.New("Unable to parse operation id")
	}
}

func parseItemToString(itemId int) (string, error) {
	switch {
	case itemId == 1:
		return "MD2 Cold Pressed", nil
	case itemId == 2:
		return "MD2 Fresh Cut Fruit", nil
	case itemId == 3:
		return "MD2 Raw Fruit", nil
	default:
		return "", errors.New("Unable to parse item id")
	}
}

type FormDataNewSale struct {
	Sale_TimeDate  string `json:"sale_time_date" xml:"sale_time_date" form:"sale_time_date"`
	PaymentMethod  string `json:"payment_type" xml:"payment_type" form:"payment_type"`
	Qty_FreshJuice int    `json:"fresh_juice_qty" xml:"fresh_juice_qty" form:"fresh_juice_qty"`
	Qty_CutFruit   int    `json:"cut_fruit_qty" xml:"cut_fruit_qty" form:"cut_fruit_qty"`
	Qty_RawFruit   int    `json:"raw_fruit_qty" xml:"raw_fruit_qty" form:"raw_fruit_qty"`
}

type JsonSale struct {
	ID          int       `json:"id"`
	Amount      float32   `json:"amount"`
	Qty         float32   `json:"quantity"` //this is float and not int bcos in case we plan to sell by weight, then it wouldnt make sense to use int
	PaymentType int       `json:"payment_type"`
	OperationID int       `json:"operation_id"`
	ItemID      int       `json:"item_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ViewSale struct {
	ID          int    `json:"id"`
	Amount      string `json:"amount"`
	Qty         string `json:"quantity"` //this is float and not int bcos in case we plan to sell by weight, then it wouldnt make sense to use int
	PaymentType string `json:"payment_type"`
	Operation   string `json:"operation"`
	Item        string `json:"item"`
	Time        string `json:"time"`
	Date        string `json:"date"`
}
