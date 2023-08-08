package helper

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/CRTOsp3ck/mims-app/config"
	"github.com/CRTOsp3ck/mims-app/model"
	"github.com/gofiber/fiber/v2"
)

func ReverseViewSales(input []*model.ViewSale) []*model.ViewSale {
	if len(input) == 0 {
		return input
	}
	return append(ReverseViewSales(input[1:]), input[0])
}

func ParsePaymentMethodToInt(paymentType string) (int, error) {
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

func ParsePaymentMethodToString(paymentType int) (string, error) {
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

func ParseOperationToString(operationId int) (string, error) {
	switch {
	case operationId == 1:
		return "Kebun Che Mah, Kemensah", nil
	default:
		return "", errors.New("Unable to parse operation id")
	}
}

func ParseItemToString(itemId int) (string, error) {
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

func CheckAuthState(c *fiber.Ctx) bool {
	// send request to /auth
	url := config.Config("API_SERVER_ADDR") + "/auth/sta"

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

	var respBody model.ResponseBody

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
