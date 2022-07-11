package netflix

import (
	"context"

	"github.com/bejaneps/speedtest/internal/measurement"
)

// MeasureUpload measures upload speed per second using Netflix's fast.com API
func (c *Client) MeasureUpload(ctx context.Context) (
	uploadRate measurement.BitRate,
	err error,
) {
	return 0, nil
}
