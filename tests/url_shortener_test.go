package tests

import (
	"net/http"
	"net/url"
	"testing"

	"analiticsURLShortener/internal/http-server/handlers/url/save"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
)

const (
	host = "localhost:8082"
)

func TestURLShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	e.POST("/shorten").
		WithJSON(save.Request{
			URL:   gofakeit.URL(),
			Alias: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		ContainsKey("alias")
}

func TestURLShortener_SaveRedirect(t *testing.T) {
	testCases := []struct {
		name          string
		url           string
		alias         string
		expectedCode  int
		expectedError string
	}{
		{
			name:         "Valid URL",
			url:          gofakeit.URL(),
			alias:        gofakeit.Word(),
			expectedCode: http.StatusOK,
		},
		{
			name:          "Invalid URL",
			url:           "invalid_url",
			alias:         gofakeit.Word(),
			expectedCode:  http.StatusBadRequest,
			expectedError: "field URL is not a valid URL",
		},
		{
			name:         "Empty Alias",
			url:          gofakeit.URL(),
			alias:        "",
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}
			e := httpexpect.Default(t, u.String())

			resp := e.POST("/shorten").
				WithJSON(save.Request{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				Expect().Status(tc.expectedCode).
				JSON().Object()

			if tc.expectedError != "" {
				resp.NotContainsKey("alias")
				resp.Value("error").String().IsEqual(tc.expectedError)
				return
			}

			alias := tc.alias
			if tc.alias != "" {
				resp.Value("alias").String().IsEqual(tc.alias)
			} else {
				resp.Value("alias").String().NotEmpty()
				alias = resp.Value("alias").String().Raw()
			}

			redirectClient := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
			eRedirect := httpexpect.WithConfig(httpexpect.Config{
				BaseURL:  u.String(),
				Client:   redirectClient,
				Reporter: httpexpect.NewAssertReporter(t),
			})

			eRedirect.GET("/s/" + alias).
				Expect().
				Status(http.StatusFound).
				Header("Location").IsEqual(tc.url)
		})
	}
}

func TestURLShortener_Analytics(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	alias := gofakeit.Word()
	originalURL := gofakeit.URL()

	e.POST("/shorten").
		WithJSON(save.Request{
			URL:   originalURL,
			Alias: alias,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		Value("alias").String().IsEqual(alias)

	redirectClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	eRedirect := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  u.String(),
		Client:   redirectClient,
		Reporter: httpexpect.NewAssertReporter(t),
	})

	for i := 0; i < 5; i++ {
		eRedirect.GET("/s/" + alias).
			Expect().
			Status(http.StatusFound).
			Header("Location").IsEqual(originalURL)
	}

	resp := e.GET("/analytics/" + alias).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	resp.Value("total_clicks").Number().IsEqual(5)
}
