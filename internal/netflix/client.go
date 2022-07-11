package netflix

import (
	"errors"
	"net/http"

	"github.com/bejaneps/speedtest/internal/config"
)

// Client is Netflix's speedtest client
type Client struct {
	conf *config.Config
	doer HTTPDoer
}

// HTTPDoer is used for mocking purposes
//go:generate mockery --name HTTPDoer
type HTTPDoer interface {
	// Do sends an HTTP request and returns an HTTP response
	Do(req *http.Request) (*http.Response, error)
}

// NewClient is a constructor for Netflix's speedtest client
func NewClient(conf *config.Config, doer HTTPDoer) (*Client, error) {
	// in case if count is set to 0,
	// better set it to 1, because 0 limit
	// will return max limit
	if conf.ServerCount == 0 {
		conf.ServerCount = 1
	}

	// TODO: remove this check when functionality
	// for retrieving token is done
	if conf.Token == "" {
		return nil, errors.New("token is required for fast.com API")
	}

	cli := &Client{
		conf: conf,
		doer: doer,
	}

	return cli, nil
}
