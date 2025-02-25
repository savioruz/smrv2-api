package scrape

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

type Scrape struct {
	timeout int
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewScrape creates a new scraper with initialized ChromeDP context
func NewScrape(timeout int) *Scrape {
	return &Scrape{
		timeout: timeout,
	}
}

// Initialize sets up the ChromeDP context
func (s *Scrape) Initialize() error {
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0"),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("single-process", true),
		chromedp.Flag("no-zygote", true),
		chromedp.Flag("memory-pressure-off", true),
		chromedp.Flag("disable-software-rasterizer", true),
		chromedp.Flag("disable-gpu-compositing", true),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// Create context with timeout
	ctx, _ := chromedp.NewContext(allocCtx)
	ctx, cancel = context.WithTimeout(ctx, time.Duration(s.timeout)*time.Second)

	s.ctx = ctx
	s.cancel = cancel
	return nil
}

func (s *Scrape) Cleanup() {
	if s.cancel != nil {
		s.cancel()
	}
}
