package redirect

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"

	resp "analiticsURLShortener/internal/lib/api/response"
	"analiticsURLShortener/internal/lib/logger/sl"
	"analiticsURLShortener/internal/storage"
)

//go:generate go run github.com/vektra/mockery/v2@v2.51.1 --name=URLRedirector
type URLRedirector interface {
	GetURL(alias string) (string, error)
	SaveAnalytics(alias string, userAgent string) error
}

func New(log *slog.Logger, urlRedirector URLRedirector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "short_url")
		if alias == "" {
			log.Info("alias is empty")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		resURL, err := urlRedirector.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, resp.Error("not found"))

			return
		}
		if err != nil {
			log.Error("failed to get url", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		userAgent := r.UserAgent()
		err = urlRedirector.SaveAnalytics(alias, userAgent)
		if err != nil {
			log.Error("failed to save analytics", sl.Err(err))
		}

		log.Info("got url", slog.String("url", resURL))

		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
