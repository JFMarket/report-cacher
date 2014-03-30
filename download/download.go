// This package handles everything required to download reports from ShopKeep.
// It is not responsible for downloading these reports on a schedule.
package download

import (
	// "code.google.com/p/go.net/html"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

// This struct is used to interface with ShopKeep and download reports.
// Generally, it should be created with New()
type Downloader struct {
	client             *http.Client // This client is used throughout this package to interact with ShopKeep.
	site               string       // The url of the shopkeep site: https://jonesboroughfarmersmkt.shopkeepapp.com
	username           string
	password           string
	authenticity_token string // The authenticity token used by ShopKeep for form submissions. Obtained at login.
}

// Returns a reference to a Downloader that is logged in and ready to begin
// downloading reports.
// Takes the site url, a username and password.
func New(s string, u string, p string) (*Downloader, error) {
	cj, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	// Initialize the object
	d := &Downloader{
		client: &http.Client{
			Jar: cj,
		},
		site:     s,
		username: u,
		password: p,
	}

	// Go ahead and login
	err = d.Login()
	if err != nil {
		return nil, errors.New("Login Failed. " + err.Error())
	}

	return d, nil
}

// Login() authenticates with ShopKeep.
// Returns a non-nil error value if login fails.
func (d *Downloader) Login() error {
	// Get the login page
	lp, err := d.client.Get(d.site)
	if err != nil {
		return errors.New("Could not get: " + d.site)
	}
	defer lp.Body.Close()

	// Pull the login page into a goquery.Document
	loginPage, err := goquery.NewDocumentFromReader(lp.Body)
	if err != nil {
		return errors.New("Failed to login: Could not read response body.")
	}

	// Determine what the authenticity token is.
	at := authToken(loginPage)
	if at == "" {
		return errors.New("Failed to find authenticity_token.")
	}
	d.authenticity_token = at
	log.Println("Found authenticity_token: " + d.authenticity_token)

	// Get the homepage by posting login credentials
	hp, err := d.client.PostForm(d.site+"/session",
		url.Values{
			"authenticity_token": {d.authenticity_token},
			"utf8":               {"✓"},
			"login":              {d.username},
			"password":           {d.password},
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
	if loginStatus(homePage) == false {
		return errors.New("Invalid username or password")
	}

	log.Println("Login successful!")

	return nil
}

// Downloads the Sold Items report from startDate to endDate to path p.
// Dates must be in the form YYYY-MM-DD.
func (d *Downloader) GetSoldItemsReport(p string, startDate string, endDate string) error {
	if d.LoggedIn() == false {
		return errors.New("Not logged in. Perhaps call Login()?")
	}

	// Get the Sold Items download page by POSTing relevant information.
	sip, err := d.client.PostForm(d.site+"/sold_items/create_export",
		url.Values{
			"authenticity_token": {d.authenticity_token},
			"utf8":               {"✓"},
			"start_date":         {startDate},
			"end_date":           {endDate},
			"chart_requested":    {},
			"grouped_by":         {},
			"commit":             {"Retrieve"},
		})
	if err != nil {
		return errors.New("Failed POSTing sold_items/create_export form. " + err.Error())
	}
	defer sip.Body.Close()

	// Return an error if the status code is not success.
	// This is useful when parameters are POSTed incorrectly.
	if sip.StatusCode != 200 {
		return errors.New("sold_items/create_export responded with " + sip.Status)
	}

	// Pull the export respones into a goquery.Document
	soldItemsPage, err := goquery.NewDocumentFromReader(sip.Body)
	if err != nil {
		return errors.New("Failed to access sold_items/create_export results. " + err.Error())
	}

	// Find the URL of the export
	reportURL, exists := soldItemsPage.Find(`#download_button input.button[type="submit"]`).Attr("data_reportfile")
	if !exists {
		return errors.New("Failed to find a download link for the Sold Items export")
	}

	// Get the CSV file
	reportRes, err := d.client.Get(reportURL)
	if err != nil {
		return errors.New("Failed to download the report from " + reportURL + " " + err.Error())
	}
	defer reportRes.Body.Close()

	// Read the CSV
	report, err := ioutil.ReadAll(reportRes.Body)
	if err != nil {
		return errors.New("Failed to read report. " + err.Error())
	}

	// Write the CSV to the given file
	err = ioutil.WriteFile(p, report, 0644)
	if err != nil {
		return errors.New("Failed to write file to " + p + " Error: " + err.Error())
	}

	return nil
}

// Checks to see if the Downloader is currently logged in.
func (d *Downloader) LoggedIn() bool {
	hp, err := d.client.Get(d.site)
	if err != nil {
		return false
	}
	defer hp.Body.Close()

	homePage, err := goquery.NewDocumentFromReader(hp.Body)
	if err != nil {
		return false
	}

	return loginStatus(homePage)
}

// Gets the authenticity token from a form in a goquery.Document.
func authToken(doc *goquery.Document) string {
	at, _ := doc.Find(`input[name="authenticity_token"]`).Attr("value")
	return at
}

// Determines whether or not the client is currently logged in based on a goquery.Document.
func loginStatus(doc *goquery.Document) bool {
	if doc.Find(`#user-controls`).Length() > 0 {
		return true
	}

	return false
}
