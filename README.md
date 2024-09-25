# Facebook Marketplace Scraper
> [WIP] A library for scraping listings off facebook marketplace

## Synopsis
A library (written in golang) for pulling listings from facebook marketplace.
It uses chromedriver and selenium to find listings and pull the relevant data from each, routing them into a nice channel that can be pulled infinitely (in theory).

## Usage
Make sure you already have chrome driver running
```go
service, err := selenium.NewChromeDriverService("chromedriver", 4444)
if err != nil {
	log.Fatal("Error:", err)
}
// Technically this shouldn't be deferred because Stop() can return an error
defer service.Stop()
```

Create capabilities of the driver you are about to pass to the web driver
```go
caps := selenium.Capabilities{}
caps.AddChrome(chrome.Capabilities{Args: []string{
	"--headless",
}})
```

Make a new instance of the scraper and start it providing a context
```go
scraper, err := NewInstance(&caps)
if err != nil {
	panic(err)
}

ctx, cancel := context.WithCancel(context.Background())

scraper.Start(ctx)
```

Now you can pull products from the channel in the scraper struct
```go
// This will give you one of the listings
listing := <-scraper.Products
fmt.Println(listing.Name)
```

## Future
The list of features that I want to add
- [ ] Specify the town / city
- [ ] Specify the search criteria
- [ ] Specify category
- [ ] Specify min price and max price
- [ ] Navigate to /marketplace/item/{id} and get extra information
