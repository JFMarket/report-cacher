package main

import (
	"github.com/jfmarket/report-cacher/download"
	"log"
)

func main() {
	log.Println("Starting...")
	downloader, err := download.New("https://jonesboroughfarmersmkt.shopkeepapp.com", "chad@snapstudent.com", "password")
	if err != nil {
		log.Fatalln(err)
	}

	downloader.GetSoldItemsReport()
}
