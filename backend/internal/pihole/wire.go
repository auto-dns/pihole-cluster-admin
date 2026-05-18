package pihole

import (
	"math"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

// General

// Auth

type authWireRequest struct {
	Password string `json:"password"`
}

type authWireResponse struct {
	Session struct {
		Valid    bool   `json:"valid"`
		SID      string `json:"sid"`
		CSRF     string `json:"csrf"`
		Validity int    `json:"validity"`
	} `json:"session"`
	Took float64 `json:"took"`
}

// Blocking

type setBlockingWireRequest struct {
	Blocking bool `json:"blocking"`
	Timer    *int `json:"timer"`
}

type blockingWireResponse struct {
	Blocking string   `json:"blocking"` // "enabled"|"disabled"|"failed"|"unknown"
	Timer    *float64 `json:"timer"`
	Took     float64  `json:"took"`
}

func blockingWireResponseToDomain(w blockingWireResponse) domain.BlockingState {
	out := domain.BlockingState{
		Status: domain.BlockingStatus(w.Blocking),
		Took:   time.Duration(math.Round(max(w.Took, 0) * float64(time.Second))),
	}
	if w.Timer != nil && *w.Timer > 0 {
		d := time.Duration(*w.Timer * float64(time.Second))
		out.TimerLeft = &d
	}
	return out
}

// Query log
// -- Request

type queriesWireRequest struct {
	Filters queriesWireFilters
	Cursor  *int
	Length  *int
	Start   *int
}

type queriesWireFilters struct {
	From, Until *int64
	Domain      *string
	ClientIP    *string
	ClientName  *string
	Upstream    *string
	Type        *string
	Status      *string
	Reply       *string
	DNSSEC      *string
	Disk        *bool
}

// -- Response

type queriesWireResponse struct {
	Queries         []queryWireEntry `json:"queries"`
	Cursor          int              `json:"cursor"`
	RecordsTotal    int64            `json:"recordsTotal"`
	RecordsFiltered int64            `json:"recordsFiltered"`
	Draw            int64            `json:"draw"`
	Took            float64          `json:"took"`
}

type queryWireEntry struct {
	Id       int64   `json:"id"`
	Time     float64 `json:"time"`
	Type     string  `json:"type"`
	Status   string  `json:"status"`
	DNSSEC   string  `json:"dnssec"`
	Domain   string  `json:"domain"`
	Upstream *string `json:"upstream"`
	Reply    struct {
		Type string  `json:"type"`
		Time float64 `json:"time"`
	} `json:"reply"`
	Client struct {
		IP   string  `json:"ip"`
		Name *string `json:"name"`
	} `json:"client"`
	ListID *int64 `json:"list_id"`
	EDE    struct {
		Code int64   `json:"code"`
		Text *string `json:"text"`
	} `json:"ede"`
	CNAME *string `json:"cname"`
}

// Domain Rules

// -- Request

type addDomainsWireRequest struct {
	Domain  []string `json:"domain"`
	Comment *string  `json:"comment,omitempty"`
	Groups  []int    `json:"groups,omitempty"`
	Enabled *bool    `json:"enabled,omitempty"`
}

// -- Response

type domainsWireResponse struct {
	Domains []domainWireInfo `json:"domains"`
	Took    float64          `json:"took"`
}

type domainWireInfo struct {
	Domain       string  `json:"domain"`
	Unicode      string  `json:"unicode"`
	Type         string  `json:"type"` // "allow" | "deny"
	Kind         string  `json:"kind"` // "exact" | "regex"
	Comment      *string `json:"comment,omitempty"`
	Groups       []int   `json:"groups"`
	Enabled      bool    `json:"enabled"`
	Id           int     `json:"id"`
	DateAdded    int64   `json:"date_added"`
	DateModified int64   `json:"date_modified"`
}

// Stats

type statsSummaryWireResponse struct {
	Queries struct {
		Total          int64   `json:"total"`
		Blocked        int64   `json:"blocked"`
		PercentBlocked float64 `json:"percent_blocked"`
		UniqueDomains  int64   `json:"unique_domains"`
	} `json:"queries"`
	Clients struct {
		Active int64 `json:"active"`
		Total  int64 `json:"total"`
	} `json:"clients"`
	Gravity struct {
		DomainsBeingBlocked int64 `json:"domains_being_blocked"`
	} `json:"gravity"`
	Took float64 `json:"took"`
}

type statsHistoryEntryWire struct {
	Timestamp int64 `json:"timestamp"`
	Total     int64 `json:"total"`
	Blocked   int64 `json:"blocked"`
}

type statsHistoryWireResponse struct {
	History []statsHistoryEntryWire `json:"history"`
	Took    float64                 `json:"took"`
}

type topDomainEntryWire struct {
	Domain string `json:"domain"`
	Count  int64  `json:"count"`
}

type statsTopDomainsWireResponse struct {
	Domains []topDomainEntryWire `json:"domains"`
	Took    float64              `json:"took"`
}

type topClientEntryWire struct {
	IP    string `json:"ip"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type statsTopClientsWireResponse struct {
	Clients []topClientEntryWire `json:"clients"`
	Took    float64              `json:"took"`
}

// Adlists

type adlistWireEntry struct {
	Id             int64   `json:"id"`
	Address        string  `json:"address"`
	Type           string  `json:"type"` // "block" | "allow"
	Comment        *string `json:"comment"`
	Groups         []int   `json:"groups"`
	Enabled        bool    `json:"enabled"`
	DateAdded      int64   `json:"date_added"`
	DateModified   int64   `json:"date_modified"`
	DateUpdated    int64   `json:"date_updated"`
	Number         int64   `json:"number"`
	InvalidDomains int64   `json:"invalid_domains"`
	Status         int     `json:"status"`
}

type listsWireResponse struct {
	Lists []adlistWireEntry `json:"lists"`
	Took  float64           `json:"took"`
}

type addAdlistWireRequest struct {
	Address string  `json:"address"`
	Type    string  `json:"type"`
	Comment *string `json:"comment,omitempty"`
	Groups  []int   `json:"groups,omitempty"`
	Enabled bool    `json:"enabled"`
}

type updateAdlistWireRequest struct {
	Enabled *bool   `json:"enabled,omitempty"`
	Comment *string `json:"comment,omitempty"`
	Groups  *[]int  `json:"groups,omitempty"`
}

// Groups

type groupWireEntry struct {
	Id           int     `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	Enabled      bool    `json:"enabled"`
	DateAdded    int64   `json:"date_added"`
	DateModified int64   `json:"date_modified"`
}

type groupsWireResponse struct {
	Groups []groupWireEntry `json:"groups"`
	Took   float64          `json:"took"`
}

type addGroupWireRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Enabled     bool    `json:"enabled"`
}

type updateGroupWireRequest struct {
	Description *string `json:"description,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
}

// Clients

type clientWireEntry struct {
	Id           int64   `json:"id"`
	IP           string  `json:"ip"`
	Name         string  `json:"name"`
	Comment      *string `json:"comment"`
	Groups       []int   `json:"groups"`
	DateAdded    int64   `json:"date_added"`
	DateModified int64   `json:"date_modified"`
}

type clientsWireResponse struct {
	Clients []clientWireEntry `json:"clients"`
	Took    float64           `json:"took"`
}

type updateClientWireRequest struct {
	Groups  []int   `json:"groups"`
	Comment *string `json:"comment,omitempty"`
}

type addDomainsWireResponse struct {
	Domains   []domainWireInfo `json:"domains"`
	Processed *struct {
		Success []struct {
			Item string `json:"item"`
		} `json:"success"`
		Errors []struct {
			Item  string `json:"item"`
			Error string `json:"error"`
		} `json:"errors"`
	} `json:"processed,omitempty"`
	Took float64 `json:"took"`
}
