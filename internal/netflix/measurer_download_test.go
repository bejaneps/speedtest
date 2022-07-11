package netflix

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/bejaneps/speedtest/internal/config"
	"github.com/bejaneps/speedtest/internal/ookla/mocks"
	"github.com/bejaneps/speedtest/internal/pkg/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMeasureDownload(t *testing.T) {
	t.Cleanup(func() {
		defaultDownloadFunc = download
	})

	tableTests := map[string]struct {
		setup       func() *mocks.HTTPDoer
		serverCount int
		token       string
		expectedErr error
	}{
		"success-100-mbit": {
			setup: func() *mocks.HTTPDoer {
				buf := &bytes.Buffer{}
				servers := []serverDetails{
					{
						URL: "https://example.com",
					},
				}
				err := json.NewEncoder(buf).Encode(&servers)
				assert.NoError(t, err)

				mockDoer := mocks.NewHTTPDoer(t)
				mockDoer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.URL.String() == "https://api.fast.com/netflix/speedtest?https=true&token=abc&urlCount=1" ||
						req.URL.String() == "https://example.com"
				})).Return(&http.Response{
					Body: io.NopCloser(buf),
				}, nil)

				defaultDownloadFunc = func(ctx context.Context, doer HTTPDoer, url string) (float64, error) {
					return 100, nil
				}

				return mockDoer
			},
			serverCount: 1,
			token:       "abc",
			expectedErr: nil,
		},
		"error-from-download-fail": {
			setup: func() *mocks.HTTPDoer {
				buf := &bytes.Buffer{}
				servers := []serverDetails{
					{
						URL: "https://example.com",
					},
				}
				err := json.NewEncoder(buf).Encode(&servers)
				assert.NoError(t, err)

				mockDoer := mocks.NewHTTPDoer(t)
				mockDoer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.URL.String() == "https://api.fast.com/netflix/speedtest?https=true&token=abc&urlCount=1" ||
						req.URL.String() == "https://example.com"
				})).Return(&http.Response{
					Body: io.NopCloser(buf),
				}, nil)

				defaultDownloadFunc = func(ctx context.Context, doer HTTPDoer, url string) (float64, error) {
					return 0, errors.New("random error")
				}

				return mockDoer
			},
			serverCount: 1,
			token:       "abc",
			expectedErr: errors.New("random error"),
		},
	}

	for testName, testCase := range tableTests {
		testCase := testCase
		t.Run(testName, func(t *testing.T) {
			doer := testCase.setup()

			cli, err := NewClient(
				&config.Config{
					ServerCount: testCase.serverCount,
					Token:       testCase.token,
				},
				doer,
			)
			assert.NoError(t, err)

			rate, err := cli.MeasureDownload(context.Background())
			if testCase.expectedErr != nil {
				fmt.Printf(err.Error())
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)

				assert.True(t, rate == 100, rate)
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
				mockDoer.
					On("Do", mock.MatchedBy(func(req *http.Request) bool {
						return req.URL.String() == "https://example.com"
					})).
					WaitUntil(time.After(1*time.Second)).
					Return(&http.Response{
						Body: io.NopCloser(bytes.NewBufferString(random.String(100_000))), // download 100_000 bytes in 1 second
					}, nil)

				return mockDoer
			},
			expectedErr: nil,
		},
		"error-from-doer-fail": {
			setup: func() *mocks.HTTPDoer {
				mockDoer := new(mocks.HTTPDoer)
				mockDoer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.URL.String() == "https://example.com"
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

			rate, err := download(context.Background(), doer, "https://example.com")
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)

				assert.True(t, rate > 799_000 && rate < 810_000, rate)
			}
		})
	}
}
