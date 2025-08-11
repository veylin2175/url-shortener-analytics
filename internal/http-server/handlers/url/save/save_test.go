package save

import (
	"analiticsURLShortener/internal/http-server/handlers/url/save/mocks"
	"analiticsURLShortener/internal/storage"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testCase struct {
	name          string
	url           string
	alias         string
	requestBody   string
	mockError     error
	expectedCode  int
	expectedBody  string
	expectedAlias string
}

func TestNew(t *testing.T) {
	tests := []testCase{
		{
			name:          "Success with alias",
			url:           "https://example.com",
			alias:         "test_alias",
			requestBody:   `{"url": "https://example.com", "alias": "test_alias"}`,
			mockError:     nil,
			expectedCode:  http.StatusOK,
			expectedBody:  `{"status":"OK","alias":"test_alias"}`,
			expectedAlias: "test_alias",
		},
		{
			name:          "Success without alias",
			url:           "https://google.com",
			alias:         "",
			requestBody:   `{"url": "https://google.com"}`,
			mockError:     nil,
			expectedCode:  http.StatusOK,
			expectedBody:  `{"status":"OK","alias":"%s"}`,
			expectedAlias: "placeholder",
		},
		{
			name:         "URL already exists",
			url:          "https://existing-url.com",
			alias:        "existing",
			requestBody:  `{"url": "https://existing-url.com", "alias": "existing"}`,
			mockError:    storage.ErrURLExists,
			expectedCode: http.StatusConflict,
			expectedBody: `{"status":"Error","error":"url already exists"}`,
		},
		{
			name:         "Invalid URL",
			url:          "invalid-url",
			alias:        "",
			requestBody:  `{"url": "invalid-url"}`,
			mockError:    nil,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"status":"Error","error":"field URL is not a valid URL"}`,
		},
		{
			name:         "Invalid JSON",
			url:          "",
			alias:        "",
			requestBody:  `{"url": "invalid-url",}`,
			mockError:    nil,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"status":"Error","error":"failed to decode request"}`,
		},
		{
			name:         "Internal server error",
			url:          "https://internal-error.com",
			alias:        "error",
			requestBody:  `{"url": "https://internal-error.com", "alias": "error"}`,
			mockError:    errors.New("db error"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"status":"Error","error":"failed to add url"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockURLSaver := mocks.NewURLSaver(t)

			if tt.expectedCode == http.StatusOK || tt.expectedCode == http.StatusConflict || tt.expectedCode == http.StatusInternalServerError {
				mockURLSaver.On("SaveURL", tt.url, mock.AnythingOfType("string")).Return(int64(1), tt.mockError).Once()
			}

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewReader([]byte(tt.requestBody)))
			req.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "test-request-id")
			req = req.WithContext(ctx)

			handler := New(slog.Default(), mockURLSaver)
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedCode, recorder.Code)

			body, _ := io.ReadAll(recorder.Body)
			var expected, actual map[string]interface{}
			json.Unmarshal([]byte(tt.expectedBody), &expected)
			json.Unmarshal(body, &actual)

			if tt.name == "Success without alias" {
				assert.NotEmpty(t, actual["alias"])
				expected["alias"] = actual["alias"]
			}

			assert.Equal(t, expected, actual)

			mockURLSaver.AssertExpectations(t)
		})
	}
}
