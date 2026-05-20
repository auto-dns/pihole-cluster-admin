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

// Info

type versionWireResponse struct {
	Version struct {
		Core struct {
			Local struct {
				Version string `json:"version"`
			} `json:"local"`
		} `json:"core"`
		FTL struct {
			Local struct {
				Version string `json:"version"`
			} `json:"local"`
		} `json:"ftl"`
	} `json:"version"`
}

type gravityInfoWireResponse struct {
	Gravity struct {
		DomainsBeingBlocked int64 `json:"domains_being_blocked"`
		LastUpdate          int64 `json:"last_update"`
	} `json:"gravity"`
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

type setPasswordWireRequest struct {
	Webserver struct {
		API struct {
			Password string `json:"password"`
		} `json:"api"`
	} `json:"webserver"`
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

type regexTestWireRequest struct {
	Domain string `json:"domain"`
}

// Config

type piholeConfigWireResponse struct {
	Config piholeConfigWire `json:"config"`
}

type piholeConfigWire struct {
	DNS       piholeConfigDNSWire       `json:"dns"`
	Misc      piholeConfigMiscWire      `json:"misc"`
	FTL       piholeConfigFTLWire       `json:"ftl"`
	Webserver piholeConfigWebserverWire `json:"webserver"`
	Resolver  piholeConfigResolverWire  `json:"resolver"`
}

type piholeConfigDNSWire struct {
	Upstreams    []string `json:"upstreams"`
	Interface    string   `json:"interface"`
	Port         int      `json:"port"`
	DNSSEC       bool     `json:"dnssec"`
	DomainNeeded bool     `json:"domainNeeded"`
	ExpandHosts  bool     `json:"expandHosts"`
	LocalTTL     int      `json:"localTTL"`
	BlockingMode string   `json:"blockingmode"`
	BlockingIPv4 string   `json:"blockingipv4"`
	BlockingIPv6 string   `json:"blockingipv6"`
	RateLimit    struct {
		Count    int `json:"count"`
		Interval int `json:"interval"`
	} `json:"ratelimit"`
	RevServer struct {
		Active bool   `json:"active"`
		CIDR   string `json:"cidr"`
		Target string `json:"target"`
		Domain string `json:"domain"`
	} `json:"revServer"`
	PiholePTR string `json:"piholePTR"`
	QueryLog  struct {
		Enabled bool `json:"enabled"`
	} `json:"querylog"`
}

type piholeConfigMiscWire struct {
	PrivacyLevel int `json:"privacylevel"`
	DelayStartup int `json:"delay_startup"`
	Nice         int `json:"nice"`
}

type piholeConfigFTLWire struct {
	QueryDisplay string `json:"query_display"`
	DelayStartup int    `json:"delay_startup"`
	Database     struct {
		DBInterval float64 `json:"DBinterval"`
		MaxDBDays  int     `json:"maxDBdays"`
	} `json:"database"`
}

type piholeConfigWebserverWire struct {
	API struct {
		ExcludeClients []string `json:"excludeClients"`
		ExcludeDomains []string `json:"excludeDomains"`
		MaxHistory     int      `json:"maxHistory"`
	} `json:"api"`
}

type piholeConfigResolverWire struct {
	ResolveIPv4  bool `json:"resolveIPv4"`
	ResolveIPv6  bool `json:"resolveIPv6"`
	NetworkNames bool `json:"networkNames"`
}

// piholeConfigPatchWire mirrors the nested config tree but with all fields
// omitempty so only explicitly set fields are sent to Pi-hole.
type piholeConfigPatchWire struct {
	DNS       *piholeConfigDNSPatchWire       `json:"dns,omitempty"`
	Misc      *piholeConfigMiscPatchWire      `json:"misc,omitempty"`
	FTL       *piholeConfigFTLPatchWire       `json:"ftl,omitempty"`
	Webserver *piholeConfigWebserverPatchWire `json:"webserver,omitempty"`
	Resolver  *piholeConfigResolverPatchWire  `json:"resolver,omitempty"`
}

type piholeConfigDNSPatchWire struct {
	Upstreams    []string                          `json:"upstreams,omitempty"`
	Interface    *string                           `json:"interface,omitempty"`
	Port         *int                              `json:"port,omitempty"`
	DNSSEC       *bool                             `json:"dnssec,omitempty"`
	DomainNeeded *bool                             `json:"domainNeeded,omitempty"`
	ExpandHosts  *bool                             `json:"expandHosts,omitempty"`
	LocalTTL     *int                              `json:"localTTL,omitempty"`
	BlockingMode *string                           `json:"blockingmode,omitempty"`
	BlockingIPv4 *string                           `json:"blockingipv4,omitempty"`
	BlockingIPv6 *string                           `json:"blockingipv6,omitempty"`
	RateLimit    *piholeConfigRateLimitPatchWire   `json:"ratelimit,omitempty"`
	RevServer    *piholeConfigRevServerPatchWire   `json:"revServer,omitempty"`
	PiholePTR    *string                           `json:"piholePTR,omitempty"`
	QueryLog     *piholeConfigQueryLogPatchWire    `json:"querylog,omitempty"`
}

type piholeConfigRateLimitPatchWire struct {
	Count    *int `json:"count,omitempty"`
	Interval *int `json:"interval,omitempty"`
}

type piholeConfigRevServerPatchWire struct {
	Active *bool   `json:"active,omitempty"`
	CIDR   *string `json:"cidr,omitempty"`
	Target *string `json:"target,omitempty"`
	Domain *string `json:"domain,omitempty"`
}

type piholeConfigQueryLogPatchWire struct {
	Enabled *bool `json:"enabled,omitempty"`
}

type piholeConfigMiscPatchWire struct {
	PrivacyLevel *int `json:"privacylevel,omitempty"`
	DelayStartup *int `json:"delay_startup,omitempty"`
	Nice         *int `json:"nice,omitempty"`
}

type piholeConfigFTLPatchWire struct {
	QueryDisplay *string                        `json:"query_display,omitempty"`
	DelayStartup *int                           `json:"delay_startup,omitempty"`
	Database     *piholeConfigFTLDBPatchWire    `json:"database,omitempty"`
}

type piholeConfigFTLDBPatchWire struct {
	DBInterval *float64 `json:"DBinterval,omitempty"`
	MaxDBDays  *int     `json:"maxDBdays,omitempty"`
}

type piholeConfigWebserverPatchWire struct {
	API *piholeConfigWebserverAPIPatchWire `json:"api,omitempty"`
}

type piholeConfigWebserverAPIPatchWire struct {
	ExcludeClients []string `json:"excludeClients,omitempty"`
	ExcludeDomains []string `json:"excludeDomains,omitempty"`
	MaxHistory     *int     `json:"maxHistory,omitempty"`
}

type piholeConfigResolverPatchWire struct {
	ResolveIPv4  *bool `json:"resolveIPv4,omitempty"`
	ResolveIPv6  *bool `json:"resolveIPv6,omitempty"`
	NetworkNames *bool `json:"networkNames,omitempty"`
}

func configFromWire(w piholeConfigWire) domain.PiholeConfig {
	cfg := domain.PiholeConfig{}
	d := w.DNS
	cfg.DNS = domain.PiholeConfigDNS{
		Upstreams:    d.Upstreams,
		Interface:    d.Interface,
		Port:         d.Port,
		DNSSEC:       d.DNSSEC,
		DomainNeeded: d.DomainNeeded,
		ExpandHosts:  d.ExpandHosts,
		LocalTTL:     d.LocalTTL,
		BlockingMode: d.BlockingMode,
		BlockingIPv4: d.BlockingIPv4,
		BlockingIPv6: d.BlockingIPv6,
		RateLimit: domain.PiholeConfigRateLimit{
			Count:    d.RateLimit.Count,
			Interval: d.RateLimit.Interval,
		},
		RevServer: domain.PiholeConfigRevServer{
			Active: d.RevServer.Active,
			CIDR:   d.RevServer.CIDR,
			Target: d.RevServer.Target,
			Domain: d.RevServer.Domain,
		},
		PiholePTR: d.PiholePTR,
		QueryLog:  domain.PiholeConfigQueryLog{Enabled: d.QueryLog.Enabled},
	}
	cfg.Misc = domain.PiholeConfigMisc{
		PrivacyLevel: w.Misc.PrivacyLevel,
		DelayStartup: w.Misc.DelayStartup,
		Nice:         w.Misc.Nice,
	}
	cfg.FTL = domain.PiholeConfigFTL{
		QueryDisplay: w.FTL.QueryDisplay,
		DelayStartup: w.FTL.DelayStartup,
		Database: domain.PiholeConfigFTLDatabase{
			DBInterval: w.FTL.Database.DBInterval,
			MaxDBDays:  w.FTL.Database.MaxDBDays,
		},
	}
	cfg.Webserver = domain.PiholeConfigWebserver{
		API: domain.PiholeConfigWebserverAPI{
			ExcludeClients: w.Webserver.API.ExcludeClients,
			ExcludeDomains: w.Webserver.API.ExcludeDomains,
			MaxHistory:     w.Webserver.API.MaxHistory,
		},
	}
	cfg.Resolver = domain.PiholeConfigResolver{
		ResolveIPv4:  w.Resolver.ResolveIPv4,
		ResolveIPv6:  w.Resolver.ResolveIPv6,
		NetworkNames: w.Resolver.NetworkNames,
	}
	return cfg
}

func patchToWire(p domain.PiholeConfigPatch) piholeConfigPatchWire {
	out := piholeConfigPatchWire{}
	if p.DNS != nil {
		d := p.DNS
		dw := &piholeConfigDNSPatchWire{
			Upstreams:    fromPtrSlice(d.Upstreams),
			Interface:    d.Interface,
			Port:         d.Port,
			DNSSEC:       d.DNSSEC,
			DomainNeeded: d.DomainNeeded,
			ExpandHosts:  d.ExpandHosts,
			LocalTTL:     d.LocalTTL,
			BlockingMode: d.BlockingMode,
			BlockingIPv4: d.BlockingIPv4,
			BlockingIPv6: d.BlockingIPv6,
			PiholePTR:    d.PiholePTR,
		}
		if d.RateLimit != nil {
			dw.RateLimit = &piholeConfigRateLimitPatchWire{
				Count:    &d.RateLimit.Count,
				Interval: &d.RateLimit.Interval,
			}
		}
		if d.RevServer != nil {
			dw.RevServer = &piholeConfigRevServerPatchWire{
				Active: &d.RevServer.Active,
				CIDR:   &d.RevServer.CIDR,
				Target: &d.RevServer.Target,
				Domain: &d.RevServer.Domain,
			}
		}
		if d.QueryLog != nil {
			dw.QueryLog = &piholeConfigQueryLogPatchWire{Enabled: d.QueryLog.Enabled}
		}
		out.DNS = dw
	}
	if p.Misc != nil {
		out.Misc = &piholeConfigMiscPatchWire{
			PrivacyLevel: p.Misc.PrivacyLevel,
			DelayStartup: p.Misc.DelayStartup,
			Nice:         p.Misc.Nice,
		}
	}
	if p.FTL != nil {
		fw := &piholeConfigFTLPatchWire{
			QueryDisplay: p.FTL.QueryDisplay,
			DelayStartup: p.FTL.DelayStartup,
		}
		if p.FTL.Database != nil {
			fw.Database = &piholeConfigFTLDBPatchWire{
				DBInterval: p.FTL.Database.DBInterval,
				MaxDBDays:  p.FTL.Database.MaxDBDays,
			}
		}
		out.FTL = fw
	}
	if p.Webserver != nil && p.Webserver.API != nil {
		aw := p.Webserver.API
		out.Webserver = &piholeConfigWebserverPatchWire{
			API: &piholeConfigWebserverAPIPatchWire{
				ExcludeClients: fromPtrSlice(aw.ExcludeClients),
				ExcludeDomains: fromPtrSlice(aw.ExcludeDomains),
				MaxHistory:     aw.MaxHistory,
			},
		}
	}
	if p.Resolver != nil {
		out.Resolver = &piholeConfigResolverPatchWire{
			ResolveIPv4:  p.Resolver.ResolveIPv4,
			ResolveIPv6:  p.Resolver.ResolveIPv6,
			NetworkNames: p.Resolver.NetworkNames,
		}
	}
	return out
}

func fromPtrSlice[T any](p *[]T) []T {
	if p == nil {
		return nil
	}
	return *p
}
