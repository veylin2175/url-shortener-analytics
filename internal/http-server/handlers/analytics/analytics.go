package analytics

import (
	"analiticsURLShortener/internal/lib/api/response"
	"analiticsURLShortener/internal/lib/logger/sl"
	"analiticsURLShortener/internal/storage"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type AnalyticsResponse struct {
	response.Response
	TotalClicks int64            `json:"total_clicks"`
	UserAgents  map[string]int64 `json:"user_agents"`
	Daily       map[string]int64 `json:"daily_clicks"`
	Monthly     map[string]int64 `json:"monthly_clicks"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.51.1 --name=URLAnalyticsGetter
type URLAnalyticsGetter interface {
	GetAnalytics(alias string) (storage.AnalyticsData, error)
}

func New(log *slog.Logger, analyticsGetter URLAnalyticsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.analytics.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "short_url")
		if alias == "" {
			log.Info("alias is empty")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		analyticsData, err := analyticsGetter.GetAnalytics(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, response.Error("not found"))
			return
		}
		if err != nil {
			log.Error("failed to get analytics", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		responseOK(w, r, analyticsData)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, data storage.AnalyticsData) {
	render.JSON(w, r, AnalyticsResponse{
		Response:    response.OK(),
		TotalClicks: data.TotalClicks,
		UserAgents:  data.UserAgents,
		Daily:       data.Daily,
		Monthly:     data.Monthly,
	})
}
