package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/CRTOsp3ck/mims-app/config"
	"github.com/CRTOsp3ck/mims-app/model"
	"github.com/gofiber/fiber/v2"
)

// Auth - Login action
func LoginRequest(c *fiber.Ctx) error {
	auth := new(model.FormAuth)
	if err := c.BodyParser(auth); err != nil {
		return err
	}

	url := config.Config("API_SERVER_ADDR") + "/auth/login"

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

	var respBody model.ResponseBody
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
}

// Auth - Login page
func LoginPage(c *fiber.Ctx) error {
	return c.Render("login", fiber.Map{
		"Title": "Login",
	})
}

// Auth - Logout
func LogoutRequest(c *fiber.Ctx) error {
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
}
