package download

func Example() {
	downloader, err := download.New("https://jonesboroughfarmersmkt.shopkeepapp.com", "chad@snapstudent.com", "password")
	if err != nil {
		log.Fatalln(err)
	}

	err = downloader.GetSoldItemsReport("files/sold_items.csv", "2014-02-28", "2014-03-29")
	if err != nil {
		log.Fatalln(err)
	} 
}

func ExampleDownloader_GetSoldItemsReport() {
	err := downloader.GetSoldItemsReport("files/sold_items.csv", "2014-02-28", "2014-03-29")
	if err != nil {
		log.Fatalln(err)
	} 
}

func ExampleNew() {
	downloader, err := download.New("https://jonesboroughfarmersmkt.shopkeepapp.com", "chad@snapstudent.com", "password")
	if err != nil {
		log.Fatalln(err)
	}
}