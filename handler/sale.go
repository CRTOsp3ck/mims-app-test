package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/CRTOsp3ck/mims-app/config"
	"github.com/CRTOsp3ck/mims-app/helper"
	"github.com/CRTOsp3ck/mims-app/model"
	"github.com/gofiber/fiber/v2"
)

func NewSale(c *fiber.Ctx) error {
	if !helper.CheckAuthState(c) {
		return c.Render("login", fiber.Map{
			"Title": "Login",
		})
	}

	return c.Render("new-sale", fiber.Map{
		"Title": "New Sale",
	}, "layouts/main")
}

func NewSaleRequest(c *fiber.Ctx) error {
	if !helper.CheckAuthState(c) {
		return c.Render("login", fiber.Map{
			"Title": "Login",
		})
	}

	ns := new(model.FormNewSale)
	if err := c.BodyParser(ns); err != nil {
		return err
	}

	var amt int
	var qty int
	var url string
	var itemId int
	var groupSaleId int
	paymentType, _ := helper.ParsePaymentMethodToInt(ns.PaymentMethod)
	operationId := 1
	groupSaleId = 0

	//i need to change DB structure to accommodate ever growing product list.
	//this is hardcoded now, since we only selling 1 product. Its ok for now...
	if ns.Qty_FreshJuice > 0 {
		amt = ns.Qty_FreshJuice * 8
		qty = ns.Qty_FreshJuice
		itemId = 1
		url = config.Config("API_SERVER_ADDR") + "/sa/new/" +
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

}

func SalesHistory(c *fiber.Ctx) error {
	if !helper.CheckAuthState(c) {
		return c.Render("login", fiber.Map{
			"Title": "Login",
		})
	}

	url := config.Config("API_SERVER_ADDR") + "/sa/find/"

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
	var sales []model.JsonSale

	//convert that string (body) to json
	if err := json.Unmarshal(body, &sales); err != nil {
		log.Println("Error unmarshalling body into JSON -", err)
		//redirect back to /main/new-sale w/ toast saying error occured
		return c.Redirect("/main/sales-history")
	}

	//create an array of view sales with that json information..
	viewSales := []*model.ViewSale{}

	for index := range sales {
		//parsing some stuff before hand
		paymentType, _ := helper.ParsePaymentMethodToString(sales[index].PaymentType)
		operation, _ := helper.ParseOperationToString(sales[index].OperationID)
		item, _ := helper.ParseItemToString(sales[index].ItemID)
		time := strings.Split(strings.Split(sales[index].CreatedAt.String(), " ")[1], ".")[0]
		date := strings.Split(sales[index].CreatedAt.String(), " ")[0]

		viewSale := model.ViewSale{
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
	viewSales = helper.ReverseViewSales(viewSales)

	//pass it to the renderer
	return c.Render("sales-history", fiber.Map{
		"Title": "Sales History",
		"Sales": viewSales,
	}, "layouts/main")
}

func SalesReport(c *fiber.Ctx) error {
	if !helper.CheckAuthState(c) {
		return c.Render("login", fiber.Map{
			"Title": "Login",
		})
	}

	url := config.Config("API_SERVER_ADDR") + "/sa/find/"

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
		Sales []model.JsonSale `json:"sales"`
	}

	//convert that string (body) to json
	if err := json.Unmarshal(body, &jsonSales.Sales); err != nil {
		log.Println("Error unmarshalling body to json -", err)
		return c.Redirect("/main/sales-report")
	}

	// Lifetime VSR
	lifetimeVsr := model.ViewSalesReport{}

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
	periodicVsr := model.ViewSalesReport{}
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
}

func SalesReportUpdatePeriodic(c *fiber.Ctx) error {
	if !helper.CheckAuthState(c) {
		return c.Render("login", fiber.Map{
			"Title": "Login",
		})
	}

	d := new(model.Dates)
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
	url := config.Config("API_SERVER_ADDR") + "/sa/find/" + d.StartDate + "-" + d.EndDate

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
		Sales []model.JsonSale `json:"sales"`
	}

	//convert that string (body) to json
	if err := json.Unmarshal(body, &jsonSales.Sales); err != nil {
		log.Println("Error unmarshalling body into JSON (lifetime) -", err)
		//redirect back to /main/new-sale w/ toast saying error occured
		return c.Redirect("/main/sales-report")
	}

	// Periodic VSR
	periodicVsr := model.ViewSalesReport{}

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
	url = config.Config("API_SERVER_ADDR") + "/sa/find/"

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
	lifetimeVsr := model.ViewSalesReport{}

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
}
