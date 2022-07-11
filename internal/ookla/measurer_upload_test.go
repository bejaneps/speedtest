package ookla

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/bejaneps/speedtest/internal/config"
	"github.com/bejaneps/speedtest/internal/measurement"
	"github.com/bejaneps/speedtest/internal/ookla/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMeasureUpload(t *testing.T) {
	t.Cleanup(func() {
		defaultUploadFunc = upload
	})

	tableTests := map[string]struct {
		setup             func() *mocks.HTTPDoer
		serverCount       int
		expectedErr       error
		expectedRateRange []measurement.BitRate
	}{
		"success-30-mbit": {
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
						req.URL.String() == "https://example.com/upload.php"
				})).Return(&http.Response{
					Body: io.NopCloser(buf),
				}, nil)

				defaultUploadFunc = func(ctx context.Context, doer HTTPDoer, url string) error {
					time.Sleep(time.Second / 10)
					return nil
				}

				return mockDoer
			},
			serverCount:       1,
			expectedErr:       nil,
			expectedRateRange: []measurement.BitRate{30000000, 39000000},
		},
		"success-3-mbit": {
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
						req.URL.String() == "https://example.com/upload.php"
				})).Return(&http.Response{
					Body: io.NopCloser(buf),
				}, nil)

				defaultUploadFunc = func(ctx context.Context, doer HTTPDoer, url string) error {
					time.Sleep(1 * time.Second)
					return nil
				}

				return mockDoer
			},
			serverCount:       1,
			expectedErr:       nil,
			expectedRateRange: []measurement.BitRate{3000000, 3900000},
		},
		"error-from-servers-details-fail": {
			setup: func() *mocks.HTTPDoer {
				mockDoer := mocks.NewHTTPDoer(t)
				mockDoer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.URL.String() == "https://www.speedtest.net/api/js/servers?engine=js&limit=1"
				})).Return(nil, errors.New("random error"))

				return mockDoer
			},
			serverCount:       1,
			expectedErr:       errors.New("failed to send request: random error"),
			expectedRateRange: []measurement.BitRate{0, 0},
		},
		"error-from-upload-fail": {
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
						req.URL.String() == "https://example.com/upload.php"
				})).Return(&http.Response{
					Body: io.NopCloser(buf),
				}, nil)

				defaultUploadFunc = func(ctx context.Context, doer HTTPDoer, url string) error {
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

			rate, err := cli.MeasureUpload(context.Background())
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
				assert.Equal(t, 0, int(rate))
			} else {
				assert.NoError(t, err)

				assert.True(t, rate > testCase.expectedRateRange[0] &&
					rate < testCase.expectedRateRange[1], rate)
			}
		})
	}
}

func TestUpload(t *testing.T) {
	tableTests := map[string]struct {
		setup       func() *mocks.HTTPDoer
		expectedErr error
	}{
		"success": {
			setup: func() *mocks.HTTPDoer {
				mockDoer := new(mocks.HTTPDoer)
				mockDoer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					b, err := io.ReadAll(req.Body)
					if err != nil {
						return false
					}
					if !strings.HasPrefix(string(b), "content=") {
						return false
					}
					return req.URL.String() == "https://example.com/upload.php"
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
					b, err := io.ReadAll(req.Body)
					if err != nil {
						return false
					}
					if !strings.HasPrefix(string(b), "content=") {
						return false
					}
					return req.URL.String() == "https://example.com/upload.php"
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

			err := upload(context.Background(), doer, "https://example.com/upload.php")
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
