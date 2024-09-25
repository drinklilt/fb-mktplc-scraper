package fbmktplcscraper

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Product struct {
	Price    uint64
	Currency string
	Name     string
	Location string
	Link     url.URL
}

func extractFirstCurrency(s string) (string, uint64, error) {
	if s == "FREE" {
		return "", 0, nil
	}

	// Regular expression to match the first currency symbol and numeric value
	re := regexp.MustCompile(`^([^\d]+)([\d,]+)`)

	// Find the submatches for the first price
	matches := re.FindStringSubmatch(s)

	// If we get at least 2 matches (currency and number)
	if len(matches) > 2 {
		currency := matches[1]
		number := matches[2]

		// Remove any commas from the number part
		number = strings.ReplaceAll(number, ",", "")

		// Convert the number part to uint64
		value, err := strconv.ParseUint(number, 10, 64)
		if err != nil {
			return "", 0, err
		}

		return currency, value, nil
	}

	// Return error if format doesn't match
	return "", 0, fmt.Errorf("invalid format or no valid price found %s", s)
}
