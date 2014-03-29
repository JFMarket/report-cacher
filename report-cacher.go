package main

import (
	"errors"
	"github.com/jfmarket/report-cacher/download"
	"log"
	"sync"
	"time"
)

func main() {
	log.Println("Starting...")

	done := make(chan bool)
	// Update once a minute.
	// This should be configurable on the command line and
	// definitely should not remain so low.
	go downloadManager(1*time.Minute, done)

	// Limit the downloadManager to 3 minutes to avoid
	// bugging ShopKeep
	time.Sleep(3 * time.Minute)
	done <- true

	// Wait one minute for the downloadManager to finish.
	time.Sleep(1 * time.Minute)
}

// downloadManager() is responsible for refreshing reports at the given interval.
// It can be stopped by passing true through the done channel.
func downloadManager(updateInterval time.Duration, done <-chan bool) {
	for {
		select {
		case <-time.Tick(updateInterval):
			log.Println("Updating...")
			err := downloadAll()
			if err != nil {
				log.Println(err)
			}
			log.Println("Reports updated.")
		case <-done:
			log.Println("Stopping...")
			return
		}
	}
}

// downloadAll() orchestrates downloading all known reports concurrently.
// It returns an error if there is a problem logging in.
func downloadAll() error {
	downloader, err := download.New("https://jonesboroughfarmersmkt.shopkeepapp.com", "chad@snapstudent.com", "password")
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

// downloadSoldItemsReport() downloads the Sold Items report for the past week.
// This may need to be adjusted for more configurability.
func downloadSoldItemsReport(d *download.Downloader) {
	log.Println("Inside downloadSoldItemsReport")

	// Calculate and format the date a week ago and today.
	const timeLayout = "2006-01-02"
	t := time.Now()
	today := t.Format(timeLayout)
	aWeekAgo := t.AddDate(0, 0, -7).Format(timeLayout)

	// files/sold_items.csv may not be cross platform.
	err := d.GetSoldItemsReport("files/sold_items.csv", aWeekAgo, today)
	if err != nil {
		log.Println("Failed to download sold items report. Error: " + err.Error())
	}
}

// This function is temporary to demonstrate concurrency.
func fakeDownload(d *download.Downloader) {
	log.Println("Inside fakeDownload")
}
