package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/CRTOsp3ck/mims-app/helper"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

// for production
const apiServerAddr string = "http://127.0.0.1:3001"

// for development
// const apiServerAddr string = "http://104.248.98.237:3001"

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

	// Auth - Login
	app.Post("/auth/login", func(c *fiber.Ctx) error {
		auth := new(FormAuth)
		if err := c.BodyParser(auth); err != nil {
			return err
		}

		url := apiServerAddr + "/auth/login"

		bytesObj := []byte(fmt.Sprintf(`{
			"identity": %q,
			"password": %q
		}`, auth.Identity, auth.Password))
		body := bytes.NewBuffer(bytesObj)
		// log.Println("BODY - ", body)

		req, err := http.NewRequest(http.MethodPost, url, body)
		if err != nil {
			log.Println("Post request not completed -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/login")
		}

		req.Header.Set("User-Agent", "mims-app")
		req.Header.Add("Content-Type", "application/json")
		client := http.Client{
			Timeout: time.Second * 2, // Timeout after 2 seconds
		}
		res, err := client.Do(req)
		if err != nil {
			log.Println("Error occured while awaiting response -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/login")
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		b, err := io.ReadAll(res.Body)
		if err != nil {
			log.Println("Error reading response body -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/login")
		}

		var respBody ResponseBody
		err = json.Unmarshal(b, &respBody)
		if err != nil {
			log.Println("Error unmarshalling response body -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/login")
		}

		// Create cookie
		cookie := new(fiber.Cookie)
		cookie.Name = "token"
		cookie.Value = respBody.Data
		cookie.Expires = time.Now().Add(24 * time.Hour)

		// Set cookie
		c.Cookie(cookie)

		return c.Redirect("/main")
	})

	app.Get("/main/login", func(c *fiber.Ctx) error {
		return c.Render("login", fiber.Map{
			"Title": "Login",
		})
	})

	// Auth - Logout
	app.Post("/auth/logout", func(c *fiber.Ctx) error {
		// Clear cookie
		c.ClearCookie("token")

		// Set cookie expiration to past
		c.Cookie(&fiber.Cookie{
			Name: "token",
			// Set expiry date to the past
			Expires:  time.Now().Add(-(time.Hour * 2)),
			HTTPOnly: true,
			SameSite: "lax",
		})

		return c.Redirect("/main/login")
	})

	// Dashboard
	app.Get("/main", func(c *fiber.Ctx) error {
		if !checkAuthState(c) {
			return c.Render("login", fiber.Map{
				"Title": "Login",
			})
		}

		// Render dashboard within layouts/main
		return c.Render("dashboard", fiber.Map{
			"Title": "Dashboard",
		}, "layouts/main")
	})

	// New Sale
	app.Get("/main/new-sale", func(c *fiber.Ctx) error {
		if !checkAuthState(c) {
			return c.Render("login", fiber.Map{
				"Title": "Login",
			})
		}

		return c.Render("new-sale", fiber.Map{
			"Title": "New Sale",
		}, "layouts/main")
	})

	// POST New Sale
	app.Post("/main/new-sale/", func(c *fiber.Ctx) error {
		if !checkAuthState(c) {
			return c.Render("login", fiber.Map{
				"Title": "Login",
			})
		}

		ns := new(FormNewSale)
		if err := c.BodyParser(ns); err != nil {
			return err
		}

		var amt int
		var qty int
		var url string
		var itemId int
		var groupSaleId int
		paymentType, _ := parsePaymentMethodToInt(ns.PaymentMethod)
		operationId := 1
		groupSaleId = 0

		//i need to change DB structure to accommodate ever growing product list.
		//this is hardcoded now, since we only selling 1 product. Its ok for now...
		if ns.Qty_FreshJuice > 0 {
			amt = ns.Qty_FreshJuice * 8
			qty = ns.Qty_FreshJuice
			itemId = 1
			url = apiServerAddr + "/sa/new/" +
				strconv.Itoa(amt) + "-" + strconv.Itoa(qty) + "-" + strconv.Itoa(paymentType) + "-" + strconv.Itoa(operationId) + "-" + strconv.Itoa(itemId) + "-" + strconv.Itoa(groupSaleId)
		}

		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			log.Println("Post request not completed -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/new-sale")
		}

		// add authorization header to the req
		bearer := "Bearer " + c.Cookies("token")
		req.Header.Add("Authorization", bearer)
		req.Header.Set("User-Agent", "mims-app")

		client := http.Client{
			Timeout: time.Second * 2, // Timeout after 2 seconds
		}
		res, err := client.Do(req)
		if err != nil {
			log.Println("Error occured while awaiting response -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/new-sale")
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		body, err := io.ReadAll(res.Body)
		_ = body // i should return the body
		if err != nil {
			log.Println("Error reading response body -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/new-sale")
		}

		//redirect to /main/sales-history w/ toast saying sale successfully registered
		//after i create the toast, i can always redirect to "/main/new-sale" if need be instead of "/main/sales-history"
		return c.Redirect("/main/sales-history")

	})

	// Sales history
	app.Get("/main/sales-history", func(c *fiber.Ctx) error {
		if !checkAuthState(c) {
			return c.Render("login", fiber.Map{
				"Title": "Login",
			})
		}

		url := apiServerAddr + "/sa/find/"

		client := http.Client{
			Timeout: time.Second * 2, // Timeout after 2 seconds
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Println("Post request not completed -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/sales-history")
		}

		// add authorization header to the req
		bearer := "Bearer " + c.Cookies("token")
		req.Header.Add("Authorization", bearer)
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

		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Println("Error reading response body -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/sales-history")
		}

		// log.Println("Body -", string(body))
		var sales []JsonSale

		//convert that string (body) to json
		if err := json.Unmarshal(body, &sales); err != nil {
			log.Println("Error unmarshalling body into JSON -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/sales-history")
		}

		//create an array of view sales with that json information..
		viewSales := []*ViewSale{}

		for index := range sales {
			//parsing some stuff before hand
			paymentType, _ := parsePaymentMethodToString(sales[index].PaymentType)
			operation, _ := parseOperationToString(sales[index].OperationID)
			item, _ := parseItemToString(sales[index].ItemID)
			time := strings.Split(strings.Split(sales[index].CreatedAt.String(), " ")[1], ".")[0]
			date := strings.Split(sales[index].CreatedAt.String(), " ")[0]

			viewSale := ViewSale{
				ID:          sales[index].ID,
				Amount:      "RM" + strconv.FormatFloat(float64(sales[index].Amount), 'f', -1, 64),
				Qty:         strconv.FormatFloat(float64(sales[index].Qty), 'f', -1, 64) + " unit(s)",
				PaymentType: paymentType,
				Operation:   operation,
				Item:        item,
				Time:        time,
				Date:        date,
			}
			viewSales = append(viewSales, &viewSale)
		}

		//reversing to sort by descending order
		viewSales = reverseViewSales(viewSales)

		//pass it to the renderer
		return c.Render("sales-history", fiber.Map{
			"Title": "Sales History",
			"Sales": viewSales,
		}, "layouts/main")
	})

	// Sales report
	app.Get("/main/sales-report", func(c *fiber.Ctx) error {
		if !checkAuthState(c) {
			return c.Render("login", fiber.Map{
				"Title": "Login",
			})
		}

		url := apiServerAddr + "/sa/find/"

		client := http.Client{
			Timeout: time.Second * 2, // Timeout after 2 seconds
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Println("Post request not completed -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/sales-report")
		}

		// add authorization header to the req
		bearer := "Bearer " + c.Cookies("token")
		req.Header.Add("Authorization", bearer)
		req.Header.Set("User-Agent", "mims-app")

		res, err := client.Do(req)
		if err != nil {
			log.Println("Error occured while awaiting response -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/sales-report")
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		body, err := io.ReadAll(res.Body)
		_ = body
		if err != nil {
			log.Println("Error reading response body -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/sales-report")
		}

		var jsonSales struct {
			Sales []JsonSale `json:"sales"`
		}

		//convert that string (body) to json
		if err := json.Unmarshal(body, &jsonSales.Sales); err != nil {
			log.Println("Error unmarshalling body to json -", err)
			return c.Redirect("/main/sales-report")
		}

		// Lifetime VSR
		lifetimeVsr := ViewSalesReport{}

		// calcuating all the revenue of every sale ever made...
		// i shouldnt be iterating as below
		// not efficient. lets start thinking of this when shit hits the fan
		for index := range jsonSales.Sales {
			lifetimeVsr.TotalGrossRevenue += float64(jsonSales.Sales[index].Amount)
		}

		lifetimeVsr.TotalExpenses = helper.RoundTo(0.00, 2)
		lifetimeVsr.TotalNetRevenue = helper.RoundTo(lifetimeVsr.TotalGrossRevenue-lifetimeVsr.TotalExpenses, 2)
		lifetimeVsr.IncomeTax = helper.RoundTo(lifetimeVsr.TotalNetRevenue*0.12, 2)
		lifetimeVsr.GrantLoan = helper.RoundTo(0.00, 2)
		lifetimeVsr.ProfitLoss = helper.RoundTo(lifetimeVsr.TotalGrossRevenue+lifetimeVsr.GrantLoan-lifetimeVsr.TotalExpenses-lifetimeVsr.IncomeTax, 2)

		// Periodic VSR
		// It will be the same as lifetime when page loads
		// maybe i should set the default range as the start of that current month until the last day of operation in that month
		periodicVsr := ViewSalesReport{}
		periodicVsr.TotalGrossRevenue = lifetimeVsr.TotalGrossRevenue
		periodicVsr.TotalExpenses = lifetimeVsr.TotalExpenses
		periodicVsr.TotalNetRevenue = lifetimeVsr.TotalNetRevenue
		periodicVsr.IncomeTax = lifetimeVsr.IncomeTax
		periodicVsr.GrantLoan = lifetimeVsr.GrantLoan
		periodicVsr.ProfitLoss = lifetimeVsr.ProfitLoss

		//pass it to the renderer
		return c.Render("sales-report", fiber.Map{
			"Title":       "Sales Analysis",
			"LifetimeVsr": lifetimeVsr,
			"PeriodicVsr": periodicVsr,
		}, "layouts/main")
	})

	type Dates struct {
		StartDate string `json:"periodic_sd" xml:"periodic_sd" form:"periodic_sd"`
		EndDate   string `json:"periodic_ed" xml:"periodic_ed" form:"periodic_ed"`
	}

	// POST Update periodic sales report
	app.Post("/main/sales-report/update-periodic", func(c *fiber.Ctx) error {
		if !checkAuthState(c) {
			return c.Render("login", fiber.Map{
				"Title": "Login",
			})
		}

		d := new(Dates)
		// parse body into struct
		if err := c.BodyParser(d); err != nil {
			log.Println("Error parsing dates into struct -", err)
			return err
		}

		//extract only the dates
		d.StartDate = strings.Split(d.StartDate, "T")[0]
		d.EndDate = strings.Split(d.EndDate, "T")[0]

		//Fetch from API Server for Periodic VSR
		//follow the api specification from mims-datastore
		url := apiServerAddr + "/sa/find/" + d.StartDate + "-" + d.EndDate

		client := http.Client{
			Timeout: time.Second * 2, // Timeout after 2 seconds
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Println("Post request not completed -", err)
			//redirect back to /main/sales-report w/ toast saying error occured
			return c.Redirect("/main/sales-report")
		}

		// add authorization header to the req
		bearer := "Bearer " + c.Cookies("token")
		req.Header.Add("Authorization", bearer)
		req.Header.Set("User-Agent", "mims-app")

		res, err := client.Do(req)
		if err != nil {
			log.Println("Error occured while awaiting response -", err)
			//redirect back to /main/sales-report w/ toast saying error occured
			return c.Redirect("/main/sales-report")
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		body, err := io.ReadAll(res.Body)
		_ = body
		if err != nil {
			log.Println("Error reading response body -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/sales-report")
		}

		var jsonSales struct {
			Sales []JsonSale `json:"sales"`
		}

		//convert that string (body) to json
		if err := json.Unmarshal(body, &jsonSales.Sales); err != nil {
			log.Println("Error unmarshalling body into JSON (lifetime) -", err)
			//redirect back to /main/new-sale w/ toast saying error occured
			return c.Redirect("/main/sales-report")
		}

		// Periodic VSR
		periodicVsr := ViewSalesReport{}

		// calcuating all the revenue of every sale ever made...
		// i shouldnt be iterating as below
		// not efficient. lets start thinking of this when shit hits the fan
		for index := range jsonSales.Sales {
			periodicVsr.TotalGrossRevenue += float64(jsonSales.Sales[index].Amount)
		}

		periodicVsr.TotalExpenses = helper.RoundTo(0.00, 2)
		periodicVsr.TotalNetRevenue = helper.RoundTo(periodicVsr.TotalGrossRevenue-periodicVsr.TotalExpenses, 2)
		periodicVsr.IncomeTax = helper.RoundTo(periodicVsr.TotalNetRevenue*0.12, 2)
		periodicVsr.GrantLoan = helper.RoundTo(0.00, 2)
		periodicVsr.ProfitLoss = helper.RoundTo(periodicVsr.TotalGrossRevenue+periodicVsr.GrantLoan-periodicVsr.TotalExpenses-periodicVsr.IncomeTax, 2)

		// Fetch from API Server for Lifetime VSR
		url = apiServerAddr + "/sa/find/"

		client = http.Client{
			Timeout: time.Second * 2, // Timeout after 2 seconds
		}

		req, err = http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Println("Post request not completed -", err)
			//redirect back to /main/sales-report w/ toast saying error occured
			return c.Redirect("/main/sales-report")
		}

		// add authorization header to the req
		bearer = "Bearer " + c.Cookies("token")
		req.Header.Add("Authorization", bearer)
		req.Header.Set("User-Agent", "mims-app")

		res, err = client.Do(req)
		if err != nil {
			log.Println("Error occured while awaiting response -", err)
			//redirect back to /main/sales-report w/ toast saying error occured
			return c.Redirect("/main/sales-report")
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		body, err = io.ReadAll(res.Body)
		_ = body
		if err != nil {
			log.Println("Error reading response body -", err)
			//redirect back to /main/sales-report w/ toast saying error occured
			return c.Redirect("/main/sales-report")
		}

		//convert that string (body) to json
		if err := json.Unmarshal(body, &jsonSales.Sales); err != nil {
			log.Println("Error unmarshalling body into JSON (periodic) -", err)
			//redirect back to /main/sales-report w/ toast saying error occured
			return c.Redirect("/main/sales-report")
		}

		// Lifetime VSR
		lifetimeVsr := ViewSalesReport{}

		// calcuating all the revenue of every sale ever made...
		// i shouldnt be iterating as below
		// not efficient. lets start thinking of this when shit hits the fan
		for index := range jsonSales.Sales {
			lifetimeVsr.TotalGrossRevenue += float64(jsonSales.Sales[index].Amount)
		}

		lifetimeVsr.TotalExpenses = helper.RoundTo(0.00, 2)
		lifetimeVsr.TotalNetRevenue = helper.RoundTo(lifetimeVsr.TotalGrossRevenue-lifetimeVsr.TotalExpenses, 2)
		lifetimeVsr.IncomeTax = helper.RoundTo(lifetimeVsr.TotalNetRevenue*0.12, 2)
		lifetimeVsr.GrantLoan = helper.RoundTo(0.00, 2)
		lifetimeVsr.ProfitLoss = helper.RoundTo(lifetimeVsr.TotalGrossRevenue+lifetimeVsr.GrantLoan-lifetimeVsr.TotalExpenses-lifetimeVsr.IncomeTax, 2)

		//pass it to the renderer
		return c.Render("sales-report", fiber.Map{
			"Title":       "Sales Analysis",
			"PeriodicVsr": periodicVsr,
			"LifetimeVsr": lifetimeVsr,
		}, "layouts/main")
	})

	// Add purchase
	app.Get("/main/add-purchase", func(c *fiber.Ctx) error {
		if !checkAuthState(c) {
			return c.Render("login", fiber.Map{
				"Title": "Login",
			})
		}

		//pass it to the renderer
		return c.Render("add-purchase", fiber.Map{
			"Title": "Add Purchase",
		}, "layouts/main")
	})

	// List purchase
	app.Get("/main/purchase-history", func(c *fiber.Ctx) error {
		if !checkAuthState(c) {
			return c.Render("login", fiber.Map{
				"Title": "Login",
			})
		}

		//pass it to the renderer
		return c.Render("purchase-history", fiber.Map{
			"Title": "List Purchase",
		}, "layouts/main")
	})

	// Static file server
	app.Static("/static", "./static")

	// Http server
	log.Fatal(app.Listen(":3000"))
}

func reverseViewSales(input []*ViewSale) []*ViewSale {
	if len(input) == 0 {
		return input
	}
	return append(reverseViewSales(input[1:]), input[0])
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

func checkAuthState(c *fiber.Ctx) bool {
	// send request to /auth
	url := apiServerAddr + "/auth/sta"

	client := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println("Post request not completed -", err)
		return false
	}

	// create a Bearer string by appending string access token
	bearer := "Bearer " + c.Cookies("token")
	// add authorization header to the req
	req.Header.Add("Authorization", bearer)
	req.Header.Set("User-Agent", "mims-app")

	res, err := client.Do(req)
	if err != nil {
		log.Println("Error occured while awaiting response -", err)
		return false
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Error reading response body -", err)
		return false
	}

	var respBody ResponseBody

	if err := json.Unmarshal(body, &respBody); err != nil {
		log.Println("Error unmarshalling response body -", err)
		return false
	}

	if respBody.Message == "Invalid or expired JWT" {
		return false
	} else if respBody.Message == "authenticated" {
		return true
	}

	// hmm?
	return false
}

// I should use this format when returning json from server (nest the data json inside this json?)
type ResponseBody struct {
	Data    string `json:"data"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type FormAuth struct {
	Identity   string `json:"identity" xml:"identity" form:"identity"`
	Password   string `json:"password" xml:"password" form:"password"`
	RememberMe bool   `json:"remember_me" xml:"remember_me" form:"remember_me"`
}

type FormNewSale struct {
	Sale_TimeDate  string `json:"sale_time_date" xml:"sale_time_date" form:"sale_time_date"`
	PaymentMethod  string `json:"payment_type" xml:"payment_type" form:"payment_type"`
	Qty_FreshJuice int    `json:"fresh_juice_qty" xml:"fresh_juice_qty" form:"fresh_juice_qty"`
	Qty_CutFruit   int    `json:"cut_fruit_qty" xml:"cut_fruit_qty" form:"cut_fruit_qty"`
	Qty_RawFruit   int    `json:"raw_fruit_qty" xml:"raw_fruit_qty" form:"raw_fruit_qty"`
}

type JsonSale struct {
	ID          int       `json:"ID"`
	Amount      float32   `json:"amount"`
	Qty         float32   `json:"qty"` //this is float and not int bcos in case we plan to sell by weight, then it wouldnt make sense to use int
	PaymentType int       `json:"payment_type"`
	OperationID int       `json:"operation_id"`
	ItemID      int       `json:"item_id"`
	GroupSaleID int       `json:"group_sale_id"`
	CreatedAt   time.Time `json:"CreatedAt"`
	UpdatedAt   time.Time `json:"UpdatedAt"`
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

type ViewSalesReport struct {
	TotalGrossRevenue float64 `json:"total_gross_revenue"`
	TotalExpenses     float64 `json:"total_expenses"`
	TotalNetRevenue   float64 `json:"total_net_revenue"`
	IncomeTax         float64 `json:"income_tax"`
	GrantLoan         float64 `json:"grant_loan"`
	ProfitLoss        float64 `json:"profit_loss"`
}
