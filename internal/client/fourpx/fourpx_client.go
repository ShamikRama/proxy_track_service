package fourpx

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"github.com/shamil/proxy_track_service-1/internal/client"
	"github.com/shamil/proxy_track_service-1/pkg/models"
)

type FourPXClient struct {
	baseURL     string
	httpClient  *http.Client
	hashPattern string
}

func NewFourPXClient(baseURL string, hashpattern string, timeout time.Duration) client.ExternalAPIClient {
	return &FourPXClient{
		baseURL:     strings.TrimSuffix(baseURL, "/"),
		hashPattern: hashpattern,
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				DisableKeepAlives:   false,
				MaxIdleConnsPerHost: 10,
			},
		},
	}
}

func (c *FourPXClient) TrackPackage(ctx context.Context, trackCode string) (*models.TrackData, error) {
	results, err := c.TrackPackagesBatch(ctx, []string{trackCode})
	if err != nil {
		return nil, err
	}

	if data, exists := results[trackCode]; exists {
		return data, nil
	}

	return nil, fmt.Errorf("tracking data not found for code: %s", trackCode)
}

func (c *FourPXClient) TrackPackagesBatch(ctx context.Context, trackCodes []string) (map[string]*models.TrackData, error) {
	if len(trackCodes) == 0 {
		return nil, fmt.Errorf("no track codes provided")
	}

	htmlContent, err := c.scrapeWithChromedp(ctx, trackCodes)
	if err != nil {
		return nil, fmt.Errorf("chromedp scraping failed: %w", err)
	}

	return ParseHTML(htmlContent, trackCodes)
}

func (c *FourPXClient) scrapeWithChromedp(ctx context.Context, trackCodes []string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	browserPath, err := findBrowserPath()
	if err != nil {
		return "", fmt.Errorf("browser not found: %w", err)
	}

	allocatorOpts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(browserPath),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		chromedp.WindowSize(1920, 1080),
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("ignore-certificate-errors", true),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, allocatorOpts...)
	defer cancelAlloc()

	tabCtx, cancelTab := chromedp.NewContext(allocCtx)
	defer cancelTab()

	trackCodeParam := strings.Join(trackCodes, ",")
	url := fmt.Sprintf("%s%s%s", c.baseURL, c.hashPattern, trackCodeParam)

	var htmlContent string

	err = chromedp.Run(tabCtx,
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {

			start := time.Now()
			for time.Since(start) < 30*time.Second {
				var title string
				if err := chromedp.Title(&title).Do(ctx); err == nil && title != "" {
					return nil
				}
				time.Sleep(1 * time.Second)
			}
			return fmt.Errorf("page failed to load within 30 seconds")
		}),

		chromedp.Sleep(5*time.Second),

		chromedp.ActionFunc(func(ctx context.Context) error {
			var content string
			err := chromedp.Evaluate(`document.body.textContent || document.body.innerText || ''`, &content).Do(ctx)
			if err != nil {
				return err
			}
			if strings.TrimSpace(content) == "" {
				return fmt.Errorf("page content is empty after loading")
			}
			return nil
		}),

		chromedp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			htmlContent, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			return err
		}),
	)

	if err != nil {
		return "", fmt.Errorf("chromedp run failed: %w", err)
	}

	return htmlContent, nil
}

func findBrowserPath() (string, error) {
	possiblePaths := []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	if path, err := exec.LookPath("google-chrome"); err == nil {
		return path, nil
	}
	if path, err := exec.LookPath("google-chrome-stable"); err == nil {
		return path, nil
	}
	if path, err := exec.LookPath("chromium"); err == nil {
		return path, nil
	}
	if path, err := exec.LookPath("chromium-browser"); err == nil {
		return path, nil
	}
	if path, err := exec.LookPath("chrome"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("no browser found. Please install Chrome or Chromium")
}

func (c *FourPXClient) Health(ctx context.Context) error {
	_, err := c.TrackPackage(ctx, "LK520419617CN")
	if err != nil {
		if strings.Contains(err.Error(), "failed to create request") ||
			strings.Contains(err.Error(), "HTTP request failed") ||
			strings.Contains(err.Error(), "chromedp run failed") {
			return fmt.Errorf("external API unavailable: %w", err)
		}
	}
	return nil
}
