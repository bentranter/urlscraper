package scraper

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bentranter/urlscraper/app/models"

	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

// Scraper starts a scraper that runs forver.
//
// After each interval, the scraper does three things:
//
//   - Batch-scrapes the links to check for changes in a fixed-size goroutine
//     pool.
//   - Batch screenshots the changed URLs in a goroutine pool that is the size
//     of the number of $GOMAXPROCS.
//   - Updates the daily scheduled email to be included in the next interval.
func Scrape(interval time.Duration) {
	for {
		time.Sleep(interval)
		lastScrapeStartAt := time.Now()

		links, err := models.AllLinks(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("failed to select all links")
			continue
		}

		// Batch scrpae the links.
		changedLinks, err := batchScrapeLinks(links)
		if err != nil {
			log.Error().Err(err).Msg("failed to batch scrape links")
			continue
		}

		changedLinks, err = batchScreenshotURLs(changedLinks)
		if err != nil {
			log.Error().Err(err).Msg("failed to batch scrape links")
			continue
		}

		// TODO Batch update the changed links.
		for _, changedLink := range changedLinks {
			if _, err := models.UpdateLink(context.Background(), changedLink.ID, changedLink); err != nil {
				log.Error().Err(err).Msg("failed to update link")
			}
		}

		scrapeDuration := time.Since(lastScrapeStartAt)
		log.Info().Str("scrape_duration", scrapeDuration.String()).Msg("completed scrape")
	}
}

// A scrapedLink is used to check if a link was changed or not.
type scrapedLink struct {
	link    *models.Link
	changed bool
}

// batchScrapeLinks scrapes all links with limited concurrency, with a timeout
// of 10 minutes.
//
// This implementation follows the example at
// https://pkg.go.dev/golang.org/x/sync/errgroup#example-Group-Pipeline.
func batchScrapeLinks(links []*models.Link) ([]*models.Link, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	scrapedLinkCh := make(chan *scrapedLink)

	// Send all the links over the channel.
	g.Go(func() error {
		defer close(scrapedLinkCh)

		for _, link := range links {
			sl := &scrapedLink{
				link:    link,
				changed: false,
			}

			select {
			case scrapedLinkCh <- sl:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return nil
	})

	// Start a fixed number of goroutines to send the HTTP request to scrape
	// the links.
	updatedLinkCh := make(chan *scrapedLink)
	const numDigesters = 8

	for i := 0; i < numDigesters; i++ {
		g.Go(func() error {
			for sl := range scrapedLinkCh {
				hash, changed, err := scrapeURL(sl.link.URL, sl.link.Hash)
				if err != nil {
					return err
				}

				// Save the returned hash. In the recv on the updatedLinkCh,
				// we'll check to make sure the link was changed before
				// appending to the resulting slice of links, which ensures
				// we only return changed links.
				sl.link.Hash = hash
				sl.link.ScrapeCount = sl.link.ScrapeCount + 1
				sl.link.UpdatedAt = time.Now()

				select {
				case updatedLinkCh <- &scrapedLink{changed: changed, link: sl.link}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}

	go func() {
		g.Wait()
		close(updatedLinkCh)
	}()

	// Append any changed links to the result set.
	updatedLinks := make([]*models.Link, 0)
	for updatedLink := range updatedLinkCh {
		log.Debug().Str("url", updatedLink.link.URL).Msg("finished scraping link")

		if updatedLink.changed {
			log.Debug().Str("url", updatedLink.link.URL).Msg("page changed")
			updatedLink.link.LastChangeAt = time.Now()
			updatedLinks = append(updatedLinks, updatedLink.link)
		}
	}

	// Check whether any of the goroutines failed. Since g is accumulating the
	// errors, we don't need to send them (or check for them) in the individual
	// results sent on the channel.
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return updatedLinks, nil
}

// scrapeURL fetches a website and computes a hash of its document body.
//
// If the document body has changed, the returned boolean will be true.
func scrapeURL(url, lasthash string) (string, bool, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("failed to get url")
		return "", false, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("failed to read body")
		return "", false, err
	}
	defer resp.Body.Close()

	h := sha256.New()
	h.Write(data)
	hash := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return hash, hash != lasthash, nil
}

// batchScreenshotURLs concurrently computes the screenshot of the given URLs
// with a max concurrency of 8.
func batchScreenshotURLs(links []*models.Link) ([]*models.Link, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	linkCh := make(chan *models.Link)

	// Send all the links over the channel.
	g.Go(func() error {
		defer close(linkCh)

		for _, link := range links {
			select {
			case linkCh <- link:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return nil
	})

	// Start a fixed number of goroutines to take screenshots of the links.
	screenshotCh := make(chan *models.Link)
	const numDigesters = 8

	for i := 0; i < numDigesters; i++ {
		g.Go(func() error {
			for link := range linkCh {
				screenshot, err := screenshotURL(link.Name, link.URL)
				if err == nil {
					link.Screenshot = screenshot
				} else {
					log.Error().Err(err).Str("url", link.URL).Msg("failed to screenshot url")
				}

				select {
				case screenshotCh <- link:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}

	go func() {
		g.Wait()
		close(screenshotCh)
	}()

	// Append any changed links to the result set.
	changedLinks := make([]*models.Link, 0, len(links))
	for link := range screenshotCh {
		log.Debug().Str("url", link.URL).Msg("finished screenshotting link")
		changedLinks = append(changedLinks, link)
	}

	// Check whether any of the goroutines failed. Since g is accumulating the
	// errors, we don't need to send them (or check for them) in the individual
	// results sent on the channel.
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return changedLinks, nil
}

// screenshotURL takes a screenshot of the given URL. It returns the relative
// path to the screenshot on success, and returns an error otherwise.
//
// It shells out to `capture-website-cli` to do this. You must have installed
// Node.js deps before this will work.
//
// Spinning up a process that has to cold-start the Node.js runtime and
// interpreter is probably a bad idea, but this is easier for now than setting
// up a JSON or GRPC service in Node.js.
func screenshotURL(name, urlstr string) (string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var buf []byte

	if err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.FullScreenshot(&buf, 90),
	}); err != nil {
		return "", err
	}

	path := filepath.Join("public", "images", strings.ToLower(name)+".png")
	if err := os.WriteFile(path, buf, 0o644); err != nil {
		return "", err
	}

	return path, nil
}
