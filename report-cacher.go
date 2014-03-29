package main

import (
	"github.com/jfmarket/report-cacher/download"
	"log"
)

func main() {
	log.Println("Starting...")
	download.GetSoldItemsReport()
}
