package fbmktplcscraper

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/tebeka/selenium"
)

// An instace of everything you need to scrape things from facebook marketplace
// Driver: the selenium web Driver
// Processed: a map of all the processed products
// Products: a channel of products that will infinitely buffer new products from fb marketplace (unless Pause is true)
type Scraper struct {
	Driver    selenium.WebDriver
	Processed map[string]bool
	Products  chan Product
}

// NewInstance creates a new instance. This initialises the WebDrivers and makes all the data structures required later
func NewInstance(c *selenium.Capabilities) (*Scraper, error) {
	// Make a new driver
	driver, err := selenium.NewRemote(*c, "")
	if err != nil {
		return nil, err
	}

	// maximize the current window to avoid responsive rendering
	err = driver.MaximizeWindow("")
	if err != nil {
		return nil, err
	}

	// navigate to fb marketplace
	err = driver.Get("https://www.facebook.com/marketplace/")
	if err != nil {
		return nil, err
	}

	err = handleCookiesAndDialogue(driver)
	if err != nil {
		return nil, err
	}

	return &Scraper{
		Driver:    driver,
		Processed: make(map[string]bool),
		Products:  make(chan Product),
	}, nil
}

func (i *Scraper) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				// TODO: Gracefully shutdown everything
				close(i.Products)
			default:
				if len(i.Products) < 32 {
					moreProducts, err := i.fetchMoreProducts()
					if err != nil {
						log.Printf("Failed to fetch more products: %v\n", err)
						return
					}

					if len(moreProducts) == 0 {
						log.Println("No more products.")
						continue
					}

					// This probably could be done in a cooler way to do this but oh well
					for _, p := range moreProducts {
						i.Products <- p
					}
				}
			}
		}
	}()
}

func (i *Scraper) fetchMoreProducts() ([]Product, error) {
	var products []Product

	// Find Marketplace Listings using XPATH
	mktplcListings, err := i.Driver.FindElements(selenium.ByXPATH, `//div[@aria-label="Collection of Marketplace items" and @role="main"]//a[starts-with(@href, '/marketplace/item')]`)
	if err != nil {
		return nil, err
	}

	for _, v := range mktplcListings {
		var product Product

		href, err := v.GetAttribute("href")
		if err != nil {
			log.Printf("Failed to get the 'href' attribute: %v\n", err)
			return nil, err
		}

		u, err := url.Parse(href)
		if err != nil {
			log.Printf("Failed to parse the URL of the listing: %v\n", err)
		}

		// Check if this href is already in i.Processed
		if i.Processed[u.Path] != false {
			// If its not false, this product has aleady been processed, to continue
			log.Printf("Skipping alreay processed item: %s", u.Path)
			continue
		}

		// Set the link of the new product
		product.Link = *u

		itemContainer, err := v.FindElements(selenium.ByXPATH, `.//div//div[2]//span[@dir="auto"]`)
		if err != nil {
			log.Printf("Failed to find the listing item container: %v\n", err)
			return products, err
		}

		for i, v := range itemContainer {
			text, err := v.GetAttribute("innerText")
			if err != nil {
				log.Printf("Failed to get attribute 'innerText': %v\n", err)
				return products, err
			}
			switch i {
			case 0:
				currency, price, err := extractFirstCurrency(text)
				if err != nil {
					log.Printf("Failed to get the currency: %v\n", err)
					return products, err
				}
				product.Currency = currency
				product.Price = price
			case 1:
				product.Name = text
			case 2:
				product.Location = text
			}
		}

		// Set the processed flag for this patah to be true
		i.Processed[u.Path] = true

		products = append(products, product)
	}

	return products, nil
}

// handleCookiesAndDialogue is used in the initialisation process to remove the initial cookies, and sign in prompts
func handleCookiesAndDialogue(driver selenium.WebDriver) error {

	// decline optional cookies
	declineOptionalCookiesButton, err := driver.FindElement(selenium.ByXPATH, `//div[@aria-label="Decline optional cookies" and @role="button" and @tabindex="0"]`)
	if err != nil {
		return err
	}

	// Just because there is no error, doesn't mean there is no button
	if declineOptionalCookiesButton != nil {
		// click the decline optional cookies
		err = declineOptionalCookiesButton.Click()
		if err != nil {
			return err
		}
	}

	// close the sign in prompt, we're doing this anonymous :sunglasses:
	closeSignInPromptButton, err := driver.FindElement(selenium.ByXPATH, `//div[@role='dialog']//div[@aria-label='Close' and @role='button']`)
	if err != nil {
		return err
	}

	if closeSignInPromptButton != nil {
		// click the button
		err = closeSignInPromptButton.Click()
		if err != nil {
			return err
		}
	}

	// Wait 5 seconds for the address entry / shop local popup that appears sometimes
	shopItemsNearYouCloseButton, err := waitForElement(driver, `//div[@role='dialog' and .//div[@aria-label='Shop local']]//div[@aria-label='Close']`, 10*time.Second)
	if err != nil {
		return err
	}

	if shopItemsNearYouCloseButton != nil {
		err = shopItemsNearYouCloseButton.Click()
		if err != nil {
			return err
		}
	}

	return nil
}

// waitForElement is used in conjunction with handleCookiesAndDialogue because some of the modals are delayed (looking at you shop local address input prompt)
func waitForElement(wd selenium.WebDriver, xpath string, timeout time.Duration) (selenium.WebElement, error) {
	end := time.Now().Add(timeout)
	for time.Now().Before(end) {
		// Try to find the element
		elem, err := wd.FindElement(selenium.ByXPATH, xpath)
		if err == nil {
			// Check if the element is visible
			visible, err := elem.IsDisplayed()
			if err == nil && visible {
				return elem, nil
			}
		}
		// Wait a little before trying again
		time.Sleep(500 * time.Millisecond)
	}
	return nil, fmt.Errorf("element not found or not visible after %v", timeout)
}
