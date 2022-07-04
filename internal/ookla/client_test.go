package ookla_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/bejaneps/speedtest/internal/config"
	"github.com/bejaneps/speedtest/internal/ookla"
	"github.com/bejaneps/speedtest/internal/ookla/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient(t *testing.T) {
	t.Parallel()

	tableTests := map[string]struct {
		serverCount int
		expectedURL string
	}{
		"correct-client-5-server-count": {
			serverCount: 5,
			expectedURL: "https://www.speedtest.net/api/js/servers?engine=js&limit=5",
		},
		"correct-client-0-server-count": {
			serverCount: 0,
			expectedURL: "https://www.speedtest.net/api/js/servers?engine=js&limit=1",
		},
	}

	for testName, testCase := range tableTests {
		testCase := testCase
		t.Run(testName, func(t *testing.T) {
			mockDoer := mocks.NewHTTPDoer(t)
			mockDoer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return req.URL.String() == testCase.expectedURL
			})).Return(nil, errors.New("random error"))

			cli := ookla.NewClient(
				&config.Config{
					ServerCount: testCase.serverCount,
				},
				mockDoer,
			)

			_, err := cli.MeasureDownload(context.Background())
			assert.Error(t, err)
		})
	}
}
