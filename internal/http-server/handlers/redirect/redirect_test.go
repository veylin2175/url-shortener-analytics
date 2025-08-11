package redirect

import (
	"analiticsURLShortener/internal/http-server/handlers/redirect/mocks"
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
	"github.com/stretchr/testify/mock"
)

type testCase struct {
	name           string
	alias          string
	mockGetError   error
	mockGetURL     string
	mockSaveError  error
	expectedCode   int
	expectedBody   string
	expectedHeader string
}

func TestNew(t *testing.T) {
	tests := []testCase{
		{
			name:           "Success",
			alias:          "test-alias",
			mockGetURL:     "https://google.com",
			mockGetError:   nil,
			mockSaveError:  nil,
			expectedCode:   http.StatusFound, // 302
			expectedBody:   "",
			expectedHeader: "https://google.com",
		},
		{
			name:           "URL Not Found",
			alias:          "not-found-alias",
			mockGetURL:     "",
			mockGetError:   storage.ErrURLNotFound,
			mockSaveError:  nil,
			expectedCode:   http.StatusNotFound, // 404
			expectedBody:   `{"status":"Error","error":"not found"}`,
			expectedHeader: "",
		},
		{
			name:           "Internal Error",
			alias:          "internal-error",
			mockGetURL:     "",
			mockGetError:   errors.New("db error"),
			mockSaveError:  nil,
			expectedCode:   http.StatusInternalServerError, // 500
			expectedBody:   `{"status":"Error","error":"internal error"}`,
			expectedHeader: "",
		},
		{
			name:           "Analytics Save Failed (Redirect still works)",
			alias:          "analytics-fail-alias",
			mockGetURL:     "https://go.dev",
			mockGetError:   nil,
			mockSaveError:  errors.New("analytics save error"),
			expectedCode:   http.StatusFound, // 302
			expectedBody:   "",
			expectedHeader: "https://go.dev",
		},
		{
			name:           "Empty Alias",
			alias:          "",
			mockGetURL:     "",
			mockGetError:   nil,
			mockSaveError:  nil,
			expectedCode:   http.StatusBadRequest, // 400
			expectedBody:   `{"status":"Error","error":"invalid request"}`,
			expectedHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRedirector := mocks.NewURLRedirector(t)

			// Мокируем вызовы GetURL и SaveAnalytics только для тех кейсов, где они ожидаются
			if tt.alias != "" {
				if tt.name != "URL Not Found" && tt.name != "Internal Error" {
					mockRedirector.On("GetURL", tt.alias).Return(tt.mockGetURL, tt.mockGetError).Once()
					mockRedirector.On("SaveAnalytics", tt.alias, mock.AnythingOfType("string")).Return(tt.mockSaveError).Once()
				} else {
					mockRedirector.On("GetURL", tt.alias).Return(tt.mockGetURL, tt.mockGetError).Once()
				}
			}

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/s/"+tt.alias, nil)

			// Подготавливаем контекст с URL-параметрами для Chi
			// Этот блок кода необходим, чтобы хендлер мог получить alias
			rctx := chi.NewRouteContext()
			if tt.alias != "" {
				rctx.URLParams.Add("short_url", tt.alias)
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler := New(slog.Default(), mockRedirector)
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedCode, recorder.Code)

			if tt.expectedBody != "" {
				body, _ := io.ReadAll(recorder.Body)
				var expected, actual map[string]interface{}
				json.Unmarshal([]byte(tt.expectedBody), &expected)
				json.Unmarshal(body, &actual)
				assert.Equal(t, expected, actual)
			}

			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, recorder.Header().Get("Location"))
			}

			mockRedirector.AssertExpectations(t)
		})
	}
}
