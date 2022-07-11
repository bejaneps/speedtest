package ookla

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	stdURL "net/url"
	"strings"
	"time"

	"github.com/bejaneps/speedtest/internal/measurement"
	"github.com/bejaneps/speedtest/internal/pkg/random"
	"golang.org/x/sync/errgroup"
)

const uploadSize = 100_000

// defaultUploadFunc is a variable to wrap upload function
// for deterministic results
var defaultUploadFunc = upload

// MeasureUpload measures upload speed per second using Ookla's speedtest.net API
func (c *Client) MeasureUpload(ctx context.Context) (
	uploadRate measurement.BitRate,
	err error,
) {
	servers, err := c.getServersDetails(ctx)
	if err != nil {
		return 0, err
	}

	// for each server calculate upload speeds
	// and take average number
	avgUploadRate := 0.0
	for _, server := range servers {
		url := server.URL

		downloadRate, err := c.measureUpload(ctx, url)
		if err != nil {
			return 0, err
		}

		avgUploadRate += downloadRate
	}
	avgUploadRate = avgUploadRate / float64(len(servers))

	return measurement.BitRate(avgUploadRate), nil
}

func (c *Client) measureUpload(ctx context.Context, url string) (float64, error) {
	eg := errgroup.Group{}

	start := time.Now()
	for i := 0; i < workload; i++ {
		eg.Go(func() error {
			return defaultUploadFunc(ctx, c.doer, url)
		})
	}
	if err := eg.Wait(); err != nil {
		return 0, err
	}
	end := time.Now()

	return uploadSize * bitsInByte * workload / end.Sub(start).Seconds(), nil
}

// upload uploads random content to provided url
func upload(ctx context.Context, doer HTTPDoer, url string) error {
	values := stdURL.Values{}
	randString := random.String(uploadSize)

	values.Add("content", randString)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
