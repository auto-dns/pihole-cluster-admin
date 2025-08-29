package v1

import (
	"encoding/json"
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/transport/httpx"
	"github.com/go-chi/chi"
)

func registerQueryLog(r chi.Router, d Deps) {
	// Read
	r.Get("/", queryLogsGet(d))
}

func queryLogsGet(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqDto, httpErr := parseQueryLogParams(r.URL.Query())
		if httpErr != nil {
			httpx.WriteJSONErrorFromErr(w, httpErr)
			return
		}

		query := reqDto.QueryLogReqDTOToDomain()

		d.Logger.Debug().Msg("fetching query logs")
		res, err := d.QueryLogService.Fetch(r.Context(), query)
		if err != nil {
			httpx.WriteJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, nr := range res.Results {
			if nr.Error != nil {
				d.Logger.Warn().Err(nr.Error).Msg("partial failure fetching logs")
			}
		}

		resDto := queryLogResponseFromDomain(res)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resDto)
	}
}
