package ookla

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/bejaneps/speedtest/internal/config"
	"github.com/bejaneps/speedtest/internal/measurement"
	"github.com/bejaneps/speedtest/internal/ookla/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMeasureDownload(t *testing.T) {
	t.Cleanup(func() {
		defaultDownloadFunc = download
	})

	tableTests := map[string]struct {
		setup             func() *mocks.HTTPDoer
		serverCount       int
		expectedErr       error
		expectedRateRange []measurement.BitRate
	}{
		"success-600-mbit": {
			setup: func() *mocks.HTTPDoer {
				buf := &bytes.Buffer{}
				servers := []serverDetails{
					{
						URL: "https://example.com/upload.php",
					},
				}
				err := json.NewEncoder(buf).Encode(&servers)
				assert.NoError(t, err)

				mockDoer := mocks.NewHTTPDoer(t)
				mockDoer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.URL.String() == "https://www.speedtest.net/api/js/servers?engine=js&limit=1" ||
						req.URL.String() == "https://example.com/random1000x1000.jpg"
				})).Return(&http.Response{
					Body: io.NopCloser(buf),
				}, nil)

				defaultDownloadFunc = func(ctx context.Context, doer HTTPDoer, url string) error {
					time.Sleep(time.Second / 10)
					return nil
				}

				return mockDoer
			},
			serverCount:       1,
			expectedErr:       nil,
			expectedRateRange: []measurement.BitRate{600000000, 690000000},
		},
		"success-60-mbit": {
			setup: func() *mocks.HTTPDoer {
				buf := &bytes.Buffer{}
				servers := []serverDetails{
					{
						URL: "https://example.com/upload.php",
					},
				}
				err := json.NewEncoder(buf).Encode(&servers)
				assert.NoError(t, err)

				mockDoer := mocks.NewHTTPDoer(t)
				mockDoer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.URL.String() == "https://www.speedtest.net/api/js/servers?engine=js&limit=1" ||
						req.URL.String() == "https://example.com/random1000x1000.jpg"
				})).Return(&http.Response{
					Body: io.NopCloser(buf),
				}, nil)

				defaultDownloadFunc = func(ctx context.Context, doer HTTPDoer, url string) error {
					time.Sleep(1 * time.Second)
					return nil
				}

				return mockDoer
			},
			serverCount:       1,
			expectedErr:       nil,
			expectedRateRange: []measurement.BitRate{60000000, 69000000},
		},
		"error-from-download-fail": {
			setup: func() *mocks.HTTPDoer {
				buf := &bytes.Buffer{}
				servers := []serverDetails{
					{
						URL: "https://example.com/upload.php",
					},
				}
				err := json.NewEncoder(buf).Encode(&servers)
				assert.NoError(t, err)

				mockDoer := mocks.NewHTTPDoer(t)
				mockDoer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.URL.String() == "https://www.speedtest.net/api/js/servers?engine=js&limit=1" ||
						req.URL.String() == "https://example.com/random1000x1000.jpg"
				})).Return(&http.Response{
					Body: io.NopCloser(buf),
				}, nil)

				defaultDownloadFunc = func(ctx context.Context, doer HTTPDoer, url string) error {
					return errors.New("random error")
				}

				return mockDoer
			},
			serverCount:       1,
			expectedErr:       errors.New("random error"),
			expectedRateRange: []measurement.BitRate{0, 0},
		},
	}

	for testName, testCase := range tableTests {
		testCase := testCase
		t.Run(testName, func(t *testing.T) {
			doer := testCase.setup()

			cli := NewClient(
				&config.Config{
					ServerCount: testCase.serverCount,
				},
				doer,
			)

			rate, err := cli.MeasureDownload(context.Background())
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)

				assert.True(t, rate > testCase.expectedRateRange[0] &&
					rate < testCase.expectedRateRange[1], rate)
			}
		})
	}
}

func TestDownload(t *testing.T) {
	tableTests := map[string]struct {
		setup       func() *mocks.HTTPDoer
		expectedErr error
	}{
		"success": {
			setup: func() *mocks.HTTPDoer {
				mockDoer := new(mocks.HTTPDoer)
				mockDoer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.URL.String() == "https://example.com/random1000x1000.jpg"
				})).Return(&http.Response{
					Body: io.NopCloser(bytes.NewBufferString("blob")),
				}, nil)

				return mockDoer
			},
			expectedErr: nil,
		},
		"error-from-doer-fail": {
			setup: func() *mocks.HTTPDoer {
				mockDoer := new(mocks.HTTPDoer)
				mockDoer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.URL.String() == "https://example.com/random1000x1000.jpg"
				})).Return(nil, errors.New("random error"))

				return mockDoer
			},
			expectedErr: errors.New("failed to send request: random error"),
		},
	}

	for testName, testCase := range tableTests {
		testCase := testCase
		t.Run(testName, func(t *testing.T) {
			doer := testCase.setup()

			err := download(context.Background(), doer, "https://example.com")
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
