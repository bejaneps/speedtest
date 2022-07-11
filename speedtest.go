package speedtest

import (
	"context"
	"net/http"
	"time"

	"github.com/bejaneps/speedtest/internal/config"
	"github.com/bejaneps/speedtest/internal/measurement"
	"github.com/bejaneps/speedtest/internal/netflix"
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
func New(tool measurementTool, opts ...config.Option) (Measurer, error) {
	conf := &config.Config{}

	for _, opt := range opts {
		opt(conf)
	}

	var (
		measurer Measurer
		err      error
	)

	switch tool {
	case OoklaSpeedtest:
		measurer = ookla.NewClient(
			conf,
			&http.Client{
				Timeout: reqTimeoutDuration,
			},
		)
	case NetflixFast:
		measurer, err = netflix.NewClient(
			conf,
			&http.Client{
				Timeout: reqTimeoutDuration,
			},
		)
	}

	return measurer, err
}

// WithServerCount sets limit on how many servers
// should be used for measuring speed
func WithServerCount(serverCount int) config.Option {
	return func(c *config.Config) {
		c.ServerCount = serverCount
	}
}

// WithToken sets authentication token for Netflix's
// fast.com api.
//
// TODO: replace this option with a function
// that will fetch token from response
func WithToken(token string) config.Option {
	return func(c *config.Config) {
		c.Token = token
	}
}
