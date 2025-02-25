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
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("remote-debugging-port", "9222"),
		chromedp.Flag("remote-debugging-address", "0.0.0.0"),
		chromedp.Flag("disable-crash-reporter", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.Flag("memory-pressure-off", true),
		chromedp.Flag("disable-software-rasterizer", true),
	)

	// Create allocator context
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// Create browser context with longer timeout
	ctx, _ := chromedp.NewContext(
		allocCtx,
		chromedp.WithLogf(nil),
		chromedp.WithBrowserOption(
			chromedp.WithDialTimeout(10*time.Second),
		),
	)

	// Add timeout to context
	ctx, cancel = context.WithTimeout(ctx, time.Duration(s.timeout)*time.Second)

	s.ctx = ctx
	s.cancel = cancel

	var err error
	for i := 0; i < 3; i++ {
		err = chromedp.Run(ctx)
		if err == nil {
			return nil
		}
		time.Sleep(time.Second * 2)
	}

	return err
}

func (s *Scrape) Cleanup() {
	if s.cancel != nil {
		s.cancel()
	}
}
