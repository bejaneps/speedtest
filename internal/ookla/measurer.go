package ookla

import (
	"context"
	"encoding/json"
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
	apiURL          = "https://www.speedtest.net/api/js/servers?engine=js&limit=%d"
	serverURLSuffix = "/upload.php"
)

const (
	bitsInByte = 8

	downloadLength = 1000
	downloadWidth  = 1000
	downloadMb     = float64(downloadLength) * downloadWidth * 2

	requestDownloadCount = 4
)

type serverDetails struct {
	URL string `json:"url"`
}

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
		url := strings.TrimSuffix(server.URL, serverURLSuffix)

		downloadRate, err := c.measureDownload(ctx, url)
		if err != nil {
			return 0, err
		}

		avgDownloadRate += downloadRate
	}
	avgDownloadRate = avgDownloadRate / float64(len(servers))

	return measurement.BitRate(avgDownloadRate), nil
}

// MeasureUpload measures upload speed per second using Ookla's speedtest.net API
func (c *Client) MeasureUpload(ctx context.Context) (
	uploadRate measurement.BitRate,
	err error,
) {
	return 0, nil
}

// Measure measures download and upload speeds per second using Ookla's speedtest.net API
func (c *Client) Measure(ctx context.Context) (
	download measurement.BitRate,
	upload measurement.BitRate,
	err error,
) {
	return 0, 0, nil
}

// getServersDetails requests from speedtest.net list of servers for
// running download and upload tests
func (c *Client) getServersDetails(ctx context.Context) ([]serverDetails, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(apiURL, c.conf.ServerCount),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.doer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("failed to close response body: %v", err)
		}
	}()

	servers := make([]serverDetails, 0, c.conf.ServerCount)
	err = json.NewDecoder(resp.Body).Decode(&servers)
	if err != nil {
		return nil, fmt.Errorf("failed to json unmarshal response body: %w", err)
	}

	return servers, nil
}

// measureDownload measures download speed by requesting provided url,
// it sends n iterations of requests to server
// TODO: change n iterations to be based on latency, coordinates of server/user
// and initial warm up
func (c *Client) measureDownload(ctx context.Context, url string) (float64, error) {
	eg := errgroup.Group{}

	start := time.Now()
	for i := 0; i < requestDownloadCount; i++ {
		eg.Go(func() error {
			return defaultDownloadFunc(ctx, c.doer, url)
		})
	}
	if err := eg.Wait(); err != nil {
		return 0, err
	}
	end := time.Now()

	return downloadMb * bitsInByte * requestDownloadCount / end.Sub(start).Seconds(), nil
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
