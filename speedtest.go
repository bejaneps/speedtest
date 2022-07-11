package speedtest

import (
	"context"
	"net/http"
	"time"

	"github.com/bejaneps/speedtest/internal/config"
	"github.com/bejaneps/speedtest/internal/measurement"
	"github.com/bejaneps/speedtest/internal/ookla"
)

const reqTimeoutDuration = 60 * time.Second

type measurementTool int

const (
	// OoklaSpeedtest is Ookla's speedtest.net tool
	OoklaSpeedtest measurementTool = iota

	// NetflixFast is Netflix's fast.com tool
	NetflixFast
)

// Measurer is an interface for measuring download/upload speeds
type Measurer interface {
	// MeasureDownload measures download speed per second
	MeasureDownload(ctx context.Context) (
		downloadRate measurement.BitRate,
		err error,
	)

	// MeasureUpload measures upload speed per second
	MeasureUpload(ctx context.Context) (
		uploadRate measurement.BitRate,
		err error,
	)
}

// New is a constructor for speedtest measure api
func New(tool measurementTool, opts ...config.Option) Measurer {
	conf := &config.Config{}

	for _, opt := range opts {
		opt(conf)
	}

	var measurer Measurer

	switch tool {
	case OoklaSpeedtest:
		measurer = ookla.NewClient(
			conf,
			&http.Client{
				Timeout: reqTimeoutDuration,
			},
		)
	}

	return measurer
}

// WithServerCount sets limit on how many servers
// should be used for measuring speed
func WithServerCount(serverCount int) config.Option {
	return func(c *config.Config) {
		c.ServerCount = serverCount
	}
}
