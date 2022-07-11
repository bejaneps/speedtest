package netflix

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bejaneps/speedtest/internal/measurement"
	"golang.org/x/sync/errgroup"
)

// DefaultDownloadFunc is a variable to wrap download function
// for deterministic results
var defaultDownloadFunc = download

// MeasureDownload measures download speed per second using Netflix's fast.com API
func (c *Client) MeasureDownload(ctx context.Context) (
	downloadRate measurement.BitRate,
	err error,
) {
	servers, err := c.getServersDetails(ctx)
	if err != nil {
		return 0, err
	}

	// run each calculation function in separate
	// goroutine so it finishes faster
	eg := errgroup.Group{}
	downloadRateChan := make(chan float64, len(servers))

	avgDownloadRate := 0.0
	for _, server := range servers {
		url := server.URL

		eg.Go(func() error {
			downloadRate, err := defaultDownloadFunc(ctx, c.doer, url)
			if err != nil {
				return err
			}
			downloadRateChan <- downloadRate
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		close(downloadRateChan)

		return 0, err
	}
	close(downloadRateChan)

	// for each server calculate download speeds
	// and take average number
	for downloadRate := range downloadRateChan {
		avgDownloadRate += downloadRate
	}
	avgDownloadRate = avgDownloadRate / float64(len(servers))

	return measurement.BitRate(avgDownloadRate), nil
}

// download downloads content from provided url and calculates
// amount of bits downloaded
func download(ctx context.Context, doer HTTPDoer, url string) (float64, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	start := time.Now()
	resp, err := doer.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	end := time.Now()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}()

	buf := &bytes.Buffer{}
	b, err := io.Copy(buf, resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to copy response body: %v\n", err)
	}

	return float64(b*bitsInByte) / end.Sub(start).Seconds(), nil
}
