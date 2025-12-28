package hysa

import (
	"fmt"
	"regexp"
	"strconv"
)

type Bank struct {
	Name, ApyDataLink, ApySelector string
	IsStatic                       bool
}

func extractRate(s string) (float64, error) {
	re := regexp.MustCompile(`\d+(\.\d+)?`)
	match := re.FindString(s)
	if match == "" {
		return 0, fmt.Errorf("no number found")
	}
	return strconv.ParseFloat(match, 64)
}

func (bank *Bank) GetSavingsApy() (float64, error) {
	scrape := scrapeDynamicApy
	if bank.IsStatic {
		scrape = scrapeStaticAPY
	}

	apy, err := scrape(bank)
	if err != nil {
		return 0, err
	}

	sanitizedApy, err := extractRate(apy)
	if err != nil {
		return 0, fmt.Errorf("Error extacting APY for %s: %w\n", bank.Name, err)
	}

	return sanitizedApy, nil
}
