package analytics

import (
	"analiticsURLShortener/internal/http-server/handlers/analytics/mocks"
	"analiticsURLShortener/internal/storage"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name          string
	alias         string
	mockError     error
	mockAnalytics storage.AnalyticsData
	expectedCode  int
	expectedBody  string
}

func TestNew(t *testing.T) {
	tests := []testCase{
		{
			name:  "Success",
			alias: "test-alias",
			mockAnalytics: storage.AnalyticsData{
				TotalClicks: 10,
				UserAgents: map[string]int64{
					"Mozilla/5.0": 7,
					"Googlebot":   3,
				},
				Daily: map[string]int64{
					"2023-10-26": 5,
					"2023-10-27": 5,
				},
				Monthly: map[string]int64{
					"2023-10": 10,
				},
			},
			mockError:    nil,
			expectedCode: http.StatusOK,
			expectedBody: `{"status":"OK","total_clicks":10,"user_agents":{"Googlebot":3,"Mozilla/5.0":7},"daily_clicks":{"2023-10-26":5,"2023-10-27":5},"monthly_clicks":{"2023-10":10}}`,
		},
		{
			name:         "URL Not Found",
			alias:        "not-found-alias",
			mockError:    storage.ErrURLNotFound,
			expectedCode: http.StatusNotFound,
			expectedBody: `{"status":"Error","error":"not found"}`,
		},
		{
			name:         "Internal Error",
			alias:        "internal-error",
			mockError:    errors.New("db error"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"status":"Error","error":"internal error"}`,
		},
		{
			name:         "Empty Alias",
			alias:        "",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"status":"Error","error":"invalid request"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAnalyticsGetter := mocks.NewURLAnalyticsGetter(t)

			if tt.alias != "" {
				mockAnalyticsGetter.On("GetAnalytics", tt.alias).
					Return(tt.mockAnalytics, tt.mockError).
					Once()
			}

			recorder := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/analytics/"+tt.alias, nil)

			rctx := chi.NewRouteContext()
			if tt.alias != "" {
				rctx.URLParams.Add("short_url", tt.alias)
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler := New(slog.Default(), mockAnalyticsGetter)
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedCode, recorder.Code)

			body, _ := io.ReadAll(recorder.Body)
			var expected, actual map[string]interface{}
			json.Unmarshal([]byte(tt.expectedBody), &expected)
			json.Unmarshal(body, &actual)
			assert.Equal(t, expected, actual)

			mockAnalyticsGetter.AssertExpectations(t)
		})
	}
}
