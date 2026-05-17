package v1

import (
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/http/transport"
	"github.com/go-chi/chi"
)

func registerStats(r chi.Router, d Deps) {
	r.Get("/stats/summary", statsGetSummary(d))
	r.Get("/stats/history", statsGetHistory(d))
	r.Get("/stats/top_domains", statsGetTopDomains(d))
	r.Get("/stats/top_clients", statsGetTopClients(d))
}

func statsGetSummary(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := d.StatsService.GetSummary(r.Context())
		if err != nil {
			transport.WriteErr(w, err)
			return
		}
		transport.WriteJSON(w, http.StatusOK, statsSummaryResponseFromDomain(res))
	}
}

func statsGetHistory(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := parseStatsTimeRangeParams(r.URL.Query())
		if err != nil {
			transport.WriteErr(w, err)
			return
		}
		res, err := d.StatsService.GetHistory(r.Context(), params.From, params.Until)
		if err != nil {
			transport.WriteErr(w, err)
			return
		}
		transport.WriteJSON(w, http.StatusOK, statsHistoryResponseFromDomain(res))
	}
}

func statsGetTopDomains(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := parseStatsTimeRangeParams(r.URL.Query())
		if err != nil {
			transport.WriteErr(w, err)
			return
		}
		res, err := d.StatsService.GetTopDomains(r.Context(), params.From, params.Until, params.Count)
		if err != nil {
			transport.WriteErr(w, err)
			return
		}
		transport.WriteJSON(w, http.StatusOK, statsTopDomainsResponseFromDomain(res))
	}
}

func statsGetTopClients(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := parseStatsTimeRangeParams(r.URL.Query())
		if err != nil {
			transport.WriteErr(w, err)
			return
		}
		res, err := d.StatsService.GetTopClients(r.Context(), params.From, params.Until, params.Count)
		if err != nil {
			transport.WriteErr(w, err)
			return
		}
		transport.WriteJSON(w, http.StatusOK, statsTopClientsResponseFromDomain(res))
	}
}
