package hysa

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"
)

type Bank struct {
	Name, ApyDataLink, ApySelector string
	IsStatic                       bool
}

// scrapers need to account for multiple elements found
func scrapeStaticAPY(bank *Bank) (string, error) {
	var apy string
	var found bool

	c := colly.NewCollector(
		colly.UserAgent("ScopeBot/1.0"),
		colly.MaxDepth(1),
	)

	// Capture HTTP / request errors
	c.OnError(func(r *colly.Response, err error) {
		found = false
	})

	// Capture APY if selector matches
	c.OnHTML(bank.ApySelector, func(e *colly.HTMLElement) {
		apy = strings.TrimSpace(e.Text)
		found = true
	})

	// Visit page
	err := c.Visit(bank.ApyDataLink)
	if err != nil {
		return "", err
	}

	// Ensure collector finishes
	c.Wait()

	// Selector never matched
	if !found {
		return "", fmt.Errorf("APY selector not found for %s", bank.Name)
	}

	if apy == "" {
		return "", fmt.Errorf("empty APY extracted for %s", bank.Name)
	}

	return apy, nil
}

func scrapeDynamicApy(bank *Bank) (string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var apy string

	err := chromedp.Run(ctx,
		chromedp.Navigate(bank.ApyDataLink),
		chromedp.WaitVisible(bank.ApySelector, chromedp.ByQuery),
		chromedp.Text(bank.ApySelector, &apy, chromedp.ByQuery),
	)

	if err != nil {
		return "", fmt.Errorf("%s dynamic page scrape error:%w", bank.Name, err)
	}
	return apy, nil
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
