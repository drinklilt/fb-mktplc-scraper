package fbmktplcscraper

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func TestNewInstance(t *testing.T) {
	// initialize a Chrome browser instance on port 4444
	service, err := selenium.NewChromeDriverService("chromedriver", 4444)
	if err != nil {
		log.Fatal("Error:", err)
	}
	log.Println("Started selenium chrome driver service.")

	// configure the browser options
	caps := selenium.Capabilities{}
	caps.AddChrome(chrome.Capabilities{Args: []string{
		//"--headless", // comment out this line for testing
	}})

	// Make a new fancy new scraper
	scraper, err := NewInstance(&caps)
	if err != nil {
		panic(err)
	}

	log.Println("Created new scraper instance")

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	scraper.Start(ctx)

	// Get a bunch of products from the instance
	listing := <-scraper.Products

	fmt.Println(listing)

	cancel()

	err = service.Stop()
	if err != nil {
		t.Fatal(err)
	}
}
