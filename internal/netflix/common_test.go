package netflix

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/bejaneps/speedtest/internal/config"
	"github.com/bejaneps/speedtest/internal/netflix/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetServersDetails(t *testing.T) {
	tableTests := map[string]struct {
		setup            func() *mocks.HTTPDoer
		expectedErr      error
		expectedResponse []serverDetails
	}{
		"success": {
			setup: func() *mocks.HTTPDoer {
				buf := &bytes.Buffer{}
				servers := []serverDetails{
					{
						URL: "https://example.com",
					},
				}
				err := json.NewEncoder(buf).Encode(&servers)
				assert.NoError(t, err)

				mockDoer := new(mocks.HTTPDoer)
				mockDoer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.URL.String() == "https://api.fast.com/netflix/speedtest?https=true&token=abc&urlCount=1"
				})).Return(&http.Response{
					Body: io.NopCloser(buf),
				}, nil)

				return mockDoer
			},
			expectedErr: nil,
			expectedResponse: []serverDetails{
				{
					URL: "https://example.com",
				},
			},
		},
		"error-from-doer-fail": {
			setup: func() *mocks.HTTPDoer {
				mockDoer := new(mocks.HTTPDoer)
				mockDoer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.URL.String() == "https://api.fast.com/netflix/speedtest?https=true&token=abc&urlCount=1"
				})).Return(nil, errors.New("random error"))

				return mockDoer
			},
			expectedErr:      errors.New("failed to send request: random error"),
			expectedResponse: nil,
		},
	}

	for testName, testCase := range tableTests {
		testCase := testCase
		t.Run(testName, func(t *testing.T) {
			doer := testCase.setup()

			cli, err := NewClient(
				&config.Config{
					ServerCount: 1,
					Token:       "abc",
				},
				doer,
			)
			assert.NoError(t, err)

			details, err := cli.getServersDetails(context.Background())
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
				assert.Empty(t, testCase.expectedResponse)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResponse, details)
			}
		})
	}
}
