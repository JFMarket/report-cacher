// This package handles everything required to download reports from ShopKeep.
// It is not responsible for downloading these reports on a schedule.
package download

import (
	// "code.google.com/p/go.net/html"
	// "io/ioutil"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type Downloader struct {
	// This client is used throughout this package to interact with ShopKeep.
	client *http.Client
}

var client *http.Client

// This should be some form of argument.
var site := "https://jonesboroughfarmersmkt.shopkeepapp.com"

// init() initializes the http.Client with a cookiejar.
func init() {
	cj, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalln("Could not initialize cookiejar.")
	}

	client = &http.Client{
		Jar: cj,
	}
}

// Login() authenticates with ShopKeep.
// Returns a non-nil error value if login fails.
func Login() error {
	// Get the login page
	lp, err := client.Get("https://jonesboroughfarmersmkt.shopkeepapp.com/")
	if err != nil {
		return errors.New("Could not get: https://jonesboroughfarmersmkt.shopkeepapp.com/")
	}
	defer lp.Body.Close()

	// Pull the login page into a goquery.Document
	loginPage, err := goquery.NewDocumentFromReader(lp.Body)
	if err != nil {
		return errors.New("Failed to login: Could not read response body.")
	}

	at := authToken(loginPage)
	log.Println("Found authenticity_token: " + at)

	// Get the homepage by posting login credentials
	hp, err := client.PostForm("https://jonesboroughfarmersmkt.shopkeepapp.com/session",
		url.Values{
			"authenticity_token": {at},
			"utf8":               {"✓"},
			"login":              {"chad@snapstudent.com"},
			"password":           {"password"},
			"commit":             {"Sign in"},
		})
	if err != nil {
		return errors.New("Failed POSTing login form: " + err.Error())
	}
	defer hp.Body.Close()

	// Pull the homepage response into a goquery.Document
	homePage, err := goquery.NewDocumentFromReader(hp.Body)
	if err != nil {
		return errors.New("Failed to access homepage: " + err.Error())
	}

	// Check the login status.
	// Can't simply check response status (ShopKeep returns 200 whether login was successful or not).
	// Can't check location header as it is not included in the response.
	if LoginStatus(homePage) == false {
		return errors.New("Login failed. Invalid username or password")
	}

	log.Println("Login successful!")

	return nil
}

// Downloads the Sold Items report.
func GetSoldItemsReport() {
	err := Login()
	if err != nil {
		log.Fatalln("Could not login. " + err.Error())
	}
}

// Gets the authenticity token from a form in a goquery.Document.
func authToken(doc *goquery.Document) string {
	at, _ := doc.Find(`input[name="authenticity_token"]`).Attr("value")
	return at
}

// Determines whether or not the client is currently logged in based on a goquery.Document.
func LoginStatus(doc *goquery.Document) bool {
	if doc.Find(`#user-controls`).Length() > 0 {
		return true
	}

	return false
}
