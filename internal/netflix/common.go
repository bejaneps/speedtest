package netflix

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const apiURL = "https://api.fast.com/netflix/speedtest?https=true&token=%s&urlCount=%d"

const bitsInByte = 8

type serverDetails struct {
	URL string `json:"url"`
}

// getServersDetails requests from fast.com list of servers for
// running download and upload tests
func (c *Client) getServersDetails(ctx context.Context) ([]serverDetails, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(apiURL, c.conf.Token, c.conf.ServerCount),
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
