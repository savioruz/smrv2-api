package scrape

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

type Scrape struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewScrape creates a new scraper with initialized ChromeDP context
func NewScrape(timeout int) *Scrape {
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0"),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// Create context with timeout
	ctx, _ := chromedp.NewContext(allocCtx)
	ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)

	return &Scrape{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Scrape) Cleanup() {
	if s.cancel != nil {
		s.cancel()
	}
}
