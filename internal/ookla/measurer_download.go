package ookla

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/bejaneps/speedtest/internal/measurement"
	"golang.org/x/sync/errgroup"
)

const (
	downloadServerURLSuffix = "/upload.php"

	downloadLength = 1000
	downloadWidth  = 1000
	downloadSize   = float64(downloadLength) * downloadWidth * 2
)

// DefaultDownloadFunc is a variable to wrap download function
// for deterministic results
var defaultDownloadFunc = download

// MeasureDownload measures download speed per second using Ookla's speedtest.net API
func (c *Client) MeasureDownload(ctx context.Context) (
	downloadRate measurement.BitRate,
	err error,
) {
	servers, err := c.getServersDetails(ctx)
	if err != nil {
		return 0, err
	}

	// for each server calculate download speeds
	// and take average number
	avgDownloadRate := 0.0
	for _, server := range servers {
		url := strings.TrimSuffix(server.URL, downloadServerURLSuffix)

		downloadRate, err := c.measureDownload(ctx, url)
		if err != nil {
			return 0, err
		}

		avgDownloadRate += downloadRate
	}
	avgDownloadRate = avgDownloadRate / float64(len(servers))

	return measurement.BitRate(avgDownloadRate), nil
}

// measureDownload measures download speed by requesting provided url,
// it sends n iterations of requests to server
// TODO: change n iterations to be based on latency, coordinates of server/user
// and initial warm up
func (c *Client) measureDownload(ctx context.Context, url string) (float64, error) {
	eg := errgroup.Group{}

	start := time.Now()
	for i := 0; i < workload; i++ {
		eg.Go(func() error {
			return defaultDownloadFunc(ctx, c.doer, url)
		})
	}
	if err := eg.Wait(); err != nil {
		return 0, err
	}
	end := time.Now()

	return downloadSize * bitsInByte * workload / end.Sub(start).Seconds(), nil
}

// download downloads random content from provided url
func download(ctx context.Context, doer HTTPDoer, url string) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/random%dx%d.jpg", url, downloadLength, downloadWidth),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := doer.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}()
	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		// just log this error
		// as it's not very important
		fmt.Printf("failed to copy response body: %v\n", err)
	}

	return nil
}
