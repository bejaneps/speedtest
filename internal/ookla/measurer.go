package ookla

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bejaneps/speedtest/internal/measurement"
)

const apiURL = "https://www.speedtest.net/api/js/servers?engine=js&limit=%d"

const (
	bitsInByte = 8
	workload   = 4
)

type serverDetails struct {
	URL string `json:"url"`
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
