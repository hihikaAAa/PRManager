package statshandler

import (
    "log/slog"
    "net/http"
	"context"

    httpresp "github.com/hihikaAAa/PRManager/internal/lib/api/response"
	statsservice "github.com/hihikaAAa/PRManager/internal/services/statsservice"
)

type StatsGetter interface {
    GetStats(ctx context.Context) (statsservice.Stats, error)
}

func New(log *slog.Logger, s StatsGetter) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        const op = "internal.http-server.handlers.stats.get"
        logger := log.With(slog.String("op", op))

        st, err := s.GetStats(r.Context())
        if err != nil {
            logger.Error("failed to get stats", slog.Any("err", err))
            httpresp.WriteError(w, r, http.StatusInternalServerError, httpresp.CodeNotFound, "internal error")
            return
        }

		logger.Info("stats captured")
        httpresp.WriteOK(w, r, st)
    }
}
