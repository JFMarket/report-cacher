// Caches reports and makes them available to other applications.
package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/jfmarket/report-cacher/download"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sync"
	"time"
)

// Define program flags.
var (
	interval  = flag.Duration("interval", 6*time.Hour, "The interval at which reports will be retrieved. 30 minutes would be 30m or 0.5h. (Required)")
	site      = flag.String("site", "https://jonesboroughfarmersmkt.shopkeepapp.com", "The address of the ShopKeep site reports will be retrieved from.")
	email     = flag.String("email", "", "The email used to login. (Required)")
	password  = flag.String("password", "", "The password used to login. (Required)")
	directory = flag.String("directory", "files", "The directory where reports will be placed.")
	port      = flag.Int("port", 8085, "The port the webserver will listen on to serve reports.")
	noweb     = flag.Bool("noweb", false, "When true, the webserver is disabled.")
)

func main() {
	// Parse and verify required options are set.
	flag.Parse()

	if *email == "" {
		log.Fatalln("An email is required. -email='x@yz.com'")
	}

	if *password == "" {
		log.Fatalln("A password is required. -password=mypassword")
	}

	ensureDirectoryExists(*directory)

	log.Println("Starting...")
	log.Println("Reports will be stored in: " + *directory)

	done := make(chan bool)

	// Update on the interval specified on the command line.
	// close()ing the done channel stops the download manager.
	go downloadManager(*interval, done)

	// Gracefully handle Ctrl-C
	catchCtrlC(done)

	if !*noweb {
		// launch webserver. goroutine for now.
		go func() {
			log.Printf("Listenting on port %[1]d. Visit http://localhost:%[1]d in your browser.", *port)
			err := http.ListenAndServe(fmt.Sprintf(":%d", *port), http.FileServer(http.Dir(*directory)))
			if err != nil {
				log.Fatalln("ListenAndServe: ", err)
			}
		}()
	}

	// Limit the downloadManager to 3 minutes to avoid
	// bugging ShopKeep
	time.Sleep(3 * time.Minute)
	close(done)
}

// downloadManager() is responsible for refreshing reports at the given interval.
// It can be stopped by close()ing the done channel.
//     go downloadManager(1*time.Hour, done)
func downloadManager(updateInterval time.Duration, done <-chan bool) {
	log.Println("Update interval is: " + updateInterval.String())

	// Perform initial download when downloadManager starts.
	update()

	// Perform updates at the given interval
	for {
		select {
		case <-time.Tick(updateInterval):
			update()
		case <-done:
			log.Println("Stopping...")
			return
		}
	}
}

// downloadAll() orchestrates downloading all known reports concurrently.
// It returns an error if there is a problem logging in.
func downloadAll() error {
	downloader, err := download.New(*site, *email, *password)
	if err != nil {
		return errors.New("Failed to initialize downloader: " + err.Error())
	}

	var wg sync.WaitGroup

	// Store download functions in a slice to simplify concurrent downloading.
	downloadFunctions := []func(*download.Downloader){
		downloadSoldItemsReport,
		fakeDownload,
	}

	// Call each download function concurrently.
	// A sync.WaitGroup is used to make sure the function does not return
	// until all downloads are finished.
	for _, df := range downloadFunctions {
		wg.Add(1)
		go func(f func(*download.Downloader)) {
			defer wg.Done()
			f(downloader)
		}(df)
	}

	wg.Wait()

	return nil
}

// Run downloadAll() and handle error
func update() {
	log.Println("Updating...")
	err := downloadAll()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Reports updated.")
}

// downloadSoldItemsReport() downloads the Sold Items report for the past week.
// This may need to be adjusted for more configurability.
func downloadSoldItemsReport(d *download.Downloader) {
	log.Println("Inside downloadSoldItemsReport")

	// Calculate and format the date a week ago and today.
	const timeLayout = "2006-01-02"
	t := time.Now()
	today := t.Format(timeLayout)
	aWeekAgo := t.AddDate(0, 0, -7).Format(timeLayout)

	err := d.GetSoldItemsReport(path.Join(*directory, "sold_items.csv"), aWeekAgo, today)
	if err != nil {
		log.Println("Failed to download sold items report. Error: " + err.Error())
	}
}

// If the given directory structure does not exist,
// create it.
func ensureDirectoryExists(d string) {
	if _, err := os.Stat(*directory); err != nil {
		if os.IsNotExist(err) {
			log.Println(*directory + " does not exist. Creating it...")
			if error := os.MkdirAll(*directory, 0755); error != nil {
				log.Fatalln("Something went wrong. " + error.Error())
			} else {
				log.Println("Successfully created " + *directory)
			}
		} else {
			log.Fatalln("Something went wrong creating the desired directory. " + err.Error())
		}
	}
}

// Catches Ctrl-C and cleans up
func catchCtrlC(done chan bool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		close(done)
		time.Sleep(8 * time.Second)
		os.Exit(1)
	}()
}

// This function is temporary to demonstrate concurrency.
func fakeDownload(d *download.Downloader) {
	log.Println("Inside fakeDownload")
}
