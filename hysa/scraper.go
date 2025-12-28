package hysa

import (
	"context"
	"fmt"
	"strings"

	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"
)

type Scraper interface {
	Scrape(bank *Bank)
}

type staticScraper struct{}
type dynamicScraper struct{}

func (staticScraper) Scrape(b *Bank) (string, error) {
	return scrapeStaticAPY(b)
}

func (dynamicScraper) Scrape(b *Bank) (string, error) {
	return scrapeDynamicApy(b)
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
